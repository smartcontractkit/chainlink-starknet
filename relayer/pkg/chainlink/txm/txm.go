package txm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	starknetaccount "github.com/NethermindEth/starknet.go/account"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"

	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const (
	MaxQueueLen = 1000
)

type TxManager interface {
	Enqueue(accountAddress *felt.Felt, publicKey *felt.Felt, txFn starknetrpc.FunctionCall) error
	InflightCount() (int, int)
}

type Tx struct {
	publicKey      *felt.Felt
	accountAddress *felt.Felt
	call           starknetrpc.FunctionCall
}

type StarkTXM interface {
	services.Service
	TxManager
}

type starktxm struct {
	starter utils.StartStopOnce
	lggr    logger.Logger
	done    sync.WaitGroup
	stop    chan struct{}
	queue   chan Tx
	ks      KeystoreAdapter
	cfg     Config
	nonce   NonceManager

	client       *utils.LazyLoad[*starknet.Client]
	feederClient *utils.LazyLoad[*starknet.FeederClient]
	accountStore *AccountStore
}

func New(lggr logger.Logger, keystore loop.Keystore, cfg Config, getClient func() (*starknet.Client, error),
	getFeederClient func() (*starknet.FeederClient, error)) (StarkTXM, error) {
	txm := &starktxm{
		lggr:         logger.Named(lggr, "StarknetTxm"),
		queue:        make(chan Tx, MaxQueueLen),
		stop:         make(chan struct{}),
		client:       utils.NewLazyLoad(getClient),
		feederClient: utils.NewLazyLoad(getFeederClient),
		ks:           NewKeystoreAdapter(keystore),
		cfg:          cfg,
		accountStore: NewAccountStore(),
	}
	txm.nonce = NewNonceManager(txm.lggr)

	return txm, nil
}

func (txm *starktxm) Name() string {
	return txm.lggr.Name()
}

func (txm *starktxm) Start(ctx context.Context) error {
	return txm.starter.StartOnce("starktxm", func() error {
		if err := txm.nonce.Start(ctx); err != nil {
			return err
		}

		txm.done.Add(2) // waitgroup: tx sender + confirmer
		go txm.broadcastLoop()
		go txm.confirmLoop()

		return nil
	})
}

func (txm *starktxm) broadcastLoop() {
	defer txm.done.Done()

	ctx, cancel := utils.ContextFromChan(txm.stop)
	defer cancel()

	txm.lggr.Debugw("broadcastLoop: started")
	for {
		select {
		case <-txm.stop:
			txm.lggr.Debugw("broadcastLoop: stopped")
			return
		case tx := <-txm.queue:
			if _, err := txm.client.Get(); err != nil {
				txm.lggr.Errorw("failed to fetch client: skipping processing tx", "error", err)
				continue
			}

			// broadcast tx serially - wait until accepted by mempool before processing next
			hash, err := txm.broadcast(ctx, tx.publicKey, tx.accountAddress, tx.call)
			if err != nil {
				txm.lggr.Errorw("transaction failed to broadcast", "error", err, "tx", tx.call)
			} else {
				txm.lggr.Infow("transaction broadcast", "txhash", hash)
			}
		}
	}
}

const FEE_MARGIN uint32 = 115
const RPC_NONCE_ERR = "Invalid transaction nonce"

// TODO: change Errors to Debugs after testing
func (txm *starktxm) estimateFriFee(ctx context.Context, client *starknet.Client, accountAddress *felt.Felt, tx starknetrpc.InvokeTxnV3) (*starknetrpc.FeeEstimate, error) {
	// skip prevalidation, which is known to overestimate amount of gas needed and error with L1GasBoundsExceedsBalance
	simFlags := []starknetrpc.SimulationFlag{starknetrpc.SKIP_VALIDATE}

	for i := 1; i <= 5; i++ {
		txm.lggr.Infow("attempt to estimate fee", "attempt", i)

		estimateNonce, err := client.AccountNonce(ctx, accountAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to check account nonce: %+w", err)
		}
		tx.Nonce = estimateNonce

		feeEstimate, err := client.Provider.EstimateFee(ctx, []starknetrpc.BroadcastTxn{tx}, simFlags, starknetrpc.BlockID{Tag: "pending"})
		if err != nil {
			var dataErr ethrpc.DataError
			if !errors.As(err, &dataErr) {
				return nil, fmt.Errorf("failed to read EstimateFee error: %T %+v", err, err)
			}
			data := dataErr.ErrorData()
			dataStr := fmt.Sprintf("%+v", data)

			txm.lggr.Errorw("failed to estimate fee", "attempt", i, "error", err, "data", dataStr)

			if strings.Contains(dataStr, RPC_NONCE_ERR) {
				continue
			}

			return nil, fmt.Errorf("failed to estimate fee: %T %+v", err, err)
		}

		// track the FRI estimate, but keep looping so we print out all estimates
		var friEstimate *starknetrpc.FeeEstimate
		for j, f := range feeEstimate {
			txm.lggr.Infow("Estimated fee", "attempt", i, "index", j, "GasConsumed", f.GasConsumed.String(), "GasPrice", f.GasPrice.String(), "OverallFee", f.OverallFee.String(), "FeeUnit", string(f.FeeUnit))
			if f.FeeUnit == "FRI" {
				friEstimate = &feeEstimate[j]
			}
		}
		if friEstimate != nil {
			return friEstimate, nil
		}

		txm.lggr.Errorw("No FRI estimate was returned", "attempt", i)
	}

	txm.lggr.Errorw("all attempts to estimate fee failed")
	return nil, fmt.Errorf("all attempts to estimate fee failed")
}

func (txm *starktxm) broadcast(ctx context.Context, publicKey *felt.Felt, accountAddress *felt.Felt, call starknetrpc.FunctionCall) (txhash string, err error) {
	client, err := txm.client.Get()
	if err != nil {
		txm.client.Reset()
		return txhash, fmt.Errorf("broadcast: failed to fetch client: %+w", err)
	}
	// create new account
	cairoVersion := 2
	account, err := starknetaccount.NewAccount(client.Provider, accountAddress, publicKey.String(), txm.ks, cairoVersion)
	if err != nil {
		return txhash, fmt.Errorf("failed to create new account: %+w", err)
	}

	tx := starknetrpc.InvokeTxnV3{
		Type:          starknetrpc.TransactionType_Invoke,
		SenderAddress: account.AccountAddress,
		Version:       starknetrpc.TransactionV3,
		Signature:     []*felt.Felt{},
		Nonce:         &felt.Zero, // filled in below
		ResourceBounds: starknetrpc.ResourceBoundsMapping{ // TODO: use proper values
			L1Gas: starknetrpc.ResourceBounds{
				MaxAmount:       "0x0",
				MaxPricePerUnit: "0x0",
			},
			L2Gas: starknetrpc.ResourceBounds{
				MaxAmount:       "0x0",
				MaxPricePerUnit: "0x0",
			},
		},
		Tip:                   "0x0",
		PayMasterData:         []*felt.Felt{},
		AccountDeploymentData: []*felt.Felt{},
		NonceDataMode:         starknetrpc.DAModeL1, // TODO: confirm
		FeeMode:               starknetrpc.DAModeL1, // TODO: confirm
	}

	// Building the Calldata with the help of FmtCalldata where we pass in the FnCall struct along with the Cairo version
	tx.Calldata, err = account.FmtCalldata([]starknetrpc.FunctionCall{call})
	if err != nil {
		return txhash, err
	}

	friEstimate, err := txm.estimateFriFee(ctx, client, accountAddress, tx)
	if err != nil {
		return txhash, fmt.Errorf("failed to get FRI estimate: %+w", err)
	}

	// TODO: consider making this configurable
	// pad estimate to 150% (add extra because estimate did not include validation)
	gasConsumed := friEstimate.GasConsumed.BigInt(new(big.Int))
	expandedGas := new(big.Int).Mul(gasConsumed, big.NewInt(250))
	maxGas := new(big.Int).Div(expandedGas, big.NewInt(100))
	tx.ResourceBounds.L1Gas.MaxAmount = starknetrpc.U64(starknetutils.BigIntToFelt(maxGas).String())

	// pad by 150%
	gasPrice := friEstimate.GasPrice.BigInt(new(big.Int))
	expandedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(250))
	maxGasPrice := new(big.Int).Div(expandedGasPrice, big.NewInt(100))
	tx.ResourceBounds.L1Gas.MaxPricePerUnit = starknetrpc.U128(starknetutils.BigIntToFelt(maxGasPrice).String())

	txm.lggr.Infow("Set resource bounds", "L1MaxAmount", tx.ResourceBounds.L1Gas.MaxAmount, "L1MaxPricePerUnit", tx.ResourceBounds.L1Gas.MaxPricePerUnit)

	nonce, err := txm.nonce.NextSequence(publicKey)
	if err != nil {
		return txhash, fmt.Errorf("failed to get nonce: %+w", err)
	}
	tx.Nonce = nonce
	// Re-sign transaction now that we've determined MaxFee
	// TODO: SignInvokeTransaction for V3 is missing so we do it by hand
	hash, err := account.TransactionHashInvoke(tx)
	if err != nil {
		return txhash, err
	}
	signature, err := account.Sign(ctx, hash)
	if err != nil {
		return txhash, err
	}
	tx.Signature = signature

	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout())
	defer execCancel()

	// finally, transmit the invoke
	res, err := account.AddInvokeTransaction(execCtx, tx)
	if err != nil {
		// TODO: handle initial broadcast errors - what kind of errors occur?
		var dataErr ethrpc.DataError
		var dataStr string
		if errors.As(err, &dataErr) {
			data := dataErr.ErrorData()
			dataStr = fmt.Sprintf("%+v", data)
		}
		txm.lggr.Errorw("failed to invoke tx from address", accountAddress, "error", err, "data", dataStr)
		return txhash, fmt.Errorf("failed to invoke tx: %+w", err)
	}
	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	// update nonce if transaction is successful
	txhash = res.TransactionHash.String()
	err = errors.Join(
		txm.nonce.IncrementNextSequence(publicKey, nonce),
		txm.accountStore.GetTxStore(accountAddress).Save(nonce, txhash, &call, publicKey),
	)
	return txhash, err
}

func (txm *starktxm) confirmLoop() {
	defer txm.done.Done()

	ctx, cancel := utils.ContextFromChan(txm.stop)
	defer cancel()

	tick := time.After(txm.cfg.ConfirmationPoll())

	txm.lggr.Debugw("confirmLoop: started")

	for {
		var start time.Time
		select {
		case <-tick:
			start = time.Now()
			client, err := txm.client.Get()
			if err != nil {
				txm.lggr.Errorw("failed to load client", "error", err)
				break
			}

			hashes := txm.accountStore.GetAllUnconfirmed()
			for addr := range hashes {
				for i := range hashes[addr] {
					hash := hashes[addr][i]
					f, err := starknetutils.HexToFelt(hash)
					if err != nil {
						txm.lggr.Errorw("invalid felt value", "hash", hash)
						continue
					}
					response, err := client.Provider.GetTransactionStatus(ctx, f)

					// tx can be rejected due to a nonce error. but we cannot know from the Starknet RPC directly  so we have to wait for
					// a broadcasted tx to fail in order to fix the nonce errors

					if err != nil {
						txm.lggr.Errorw("failed to fetch transaction status", "hash", hash, "error", err)
						continue
					}

					finalityStatus := response.FinalityStatus
					executionStatus := response.ExecutionStatus

					// any finalityStatus other than received
					if finalityStatus == starknetrpc.TxnStatus_Accepted_On_L1 || finalityStatus == starknetrpc.TxnStatus_Accepted_On_L2 || finalityStatus == starknetrpc.TxnStatus_Rejected {
						txm.lggr.Debugw(fmt.Sprintf("tx confirmed: %s", finalityStatus), "hash", hash, "finalityStatus", finalityStatus)
						if err := txm.accountStore.GetTxStore(addr).Confirm(hash); err != nil {
							txm.lggr.Errorw("failed to confirm tx in TxStore", "hash", hash, "sender", addr, "error", err)
						}
					}

					// currently, feeder client is only way to get rejected reason
					if finalityStatus == starknetrpc.TxnStatus_Rejected {
						go txm.logFeederError(ctx, hash, f)
					}

					if executionStatus == starknetrpc.TxnExecutionStatusREVERTED {
						// TODO: get revert reason?
						txm.lggr.Errorw("transaction reverted", "hash", hash)
					}
				}
			}
		case <-txm.stop:
			txm.lggr.Debugw("confirmLoop: stopped")
			return
		}
		t := txm.cfg.ConfirmationPoll() - time.Since(start)
		tick = time.After(utils.WithJitter(t.Abs()))
	}
}

func (txm *starktxm) logFeederError(ctx context.Context, hash string, f *felt.Felt) {
	feederClient, err := txm.feederClient.Get()
	if err != nil {
		txm.lggr.Errorw("failed to load feeder client", "error", err)
		return
	}

	rejectedTx, err := feederClient.TransactionFailure(ctx, f)
	if err != nil {
		txm.lggr.Errorw("failed to fetch reason for transaction failure", "hash", hash, "error", err)
		return
	}

	txm.lggr.Errorw("feeder rejected reason", "hash", hash, "errorMessage", rejectedTx.ErrorMessage)
}

func (txm *starktxm) Close() error {
	return txm.starter.StopOnce("starktxm", func() error {
		close(txm.stop)
		txm.done.Wait()
		return nil
	})
}

func (txm *starktxm) Healthy() error {
	return txm.starter.Healthy()
}

func (txm *starktxm) Ready() error {
	return txm.starter.Ready()
}

func (txm *starktxm) HealthReport() map[string]error {
	return map[string]error{txm.Name(): txm.Healthy()}
}

func (txm *starktxm) Enqueue(accountAddress, publicKey *felt.Felt, tx starknetrpc.FunctionCall) error {
	// validate key exists for sender
	// use the embedded Loopp Keystore to do this; the spec and design
	// encourage passing nil data to the loop.Keystore.Sign as way to test
	// existence of a key
	if _, err := txm.ks.Loopp().Sign(context.Background(), publicKey.String(), nil); err != nil {
		return fmt.Errorf("enqueue: failed to sign: %+w", err)
	}

	client, err := txm.client.Get()
	if err != nil {
		txm.client.Reset()
		return fmt.Errorf("broadcast: failed to fetch client: %+w", err)
	}

	// register account for nonce manager
	if err := txm.nonce.Register(context.TODO(), accountAddress, publicKey, client); err != nil {
		return fmt.Errorf("failed to register nonce: %+w", err)
	}

	select {
	case txm.queue <- Tx{publicKey: publicKey, accountAddress: accountAddress, call: tx}: // TODO fix naming here
	default:
		return fmt.Errorf("failed to enqueue transaction: %+v", tx)
	}

	return nil
}

func (txm *starktxm) InflightCount() (queue int, unconfirmed int) {
	list := maps.Values(txm.accountStore.GetAllInflightCount())
	for _, count := range list {
		unconfirmed += count
	}
	return len(txm.queue), unconfirmed
}
