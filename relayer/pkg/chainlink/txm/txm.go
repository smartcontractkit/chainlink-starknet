package txm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
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

	client  *utils.LazyLoad[*starknet.Client]
	txStore *ChainTxStore
}

func New(lggr logger.Logger, keystore loop.Keystore, cfg Config, getClient func() (*starknet.Client, error)) (StarkTXM, error) {
	txm := &starktxm{
		lggr:    logger.Named(lggr, "StarknetTxm"),
		queue:   make(chan Tx, MaxQueueLen),
		stop:    make(chan struct{}),
		client:  utils.NewLazyLoad(getClient),
		ks:      NewKeystoreAdapter(keystore),
		cfg:     cfg,
		txStore: NewChainTxStore(),
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
		default:
			// skip processing if queue is empty
			if len(txm.queue) == 0 {
				continue
			}

			// preserve tx queue: don't pull tx from queue until client is known to work
			if _, err := txm.client.Get(); err != nil {
				txm.lggr.Errorw("failed to fetch client: skipping processing tx", "error", err)
				continue
			}
			tx := <-txm.queue

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

	chainID, err := client.Provider.ChainID(ctx)
	if err != nil {
		return txhash, fmt.Errorf("failed to get chainID: %+w", err)
	}

	nonce, err := txm.nonce.NextSequence(publicKey, chainID)
	if err != nil {
		return txhash, fmt.Errorf("failed to get nonce: %+w", err)
	}

	tx := starknetrpc.InvokeTxnV3{
		Type:          starknetrpc.TransactionType_Invoke,
		SenderAddress: account.AccountAddress,
		Version:       starknetrpc.TransactionV3,
		Signature:     []*felt.Felt{},
		Nonce:         nonce,
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

	// TODO: if we estimate with sig then the hash changes and we have to re-sign
	// if we don't then the signature is invalid??

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

	// get fee for tx
	// optional - pass nonce to fee estimate (if nonce gets ahead, estimate may fail)
	// can we estimate fee without calling estimate - tbd with 1.0
	simFlags := []starknetrpc.SimulationFlag{}
	feeEstimate, err := account.EstimateFee(ctx, []starknetrpc.BroadcastTxn{tx}, simFlags, starknetrpc.BlockID{Tag: "latest"})
	if err != nil {
		var data any
		if err, ok := err.(ethrpc.DataError); ok {
			data = err.ErrorData()
		}
		txm.lggr.Errorw("failed to estimate fee", "error", err, "data", data)
		return txhash, fmt.Errorf("failed to estimate fee: %T %+w", err, err)
	}

	txm.lggr.Infow("Account", "account", account.AccountAddress)

	var friEstimate *starknetrpc.FeeEstimate
	for i, f := range feeEstimate {
		txm.lggr.Infow("Estimated fee", "index", i, "GasConsumed", f.GasConsumed.String(), "GasPrice", f.GasPrice.String(), "OverallFee", f.OverallFee.String(), "FeeUnit", string(f.FeeUnit))
		if f.FeeUnit == "FRI" && friEstimate == nil {
			friEstimate = &feeEstimate[i]
		}
	}
	if friEstimate == nil {
		return txhash, fmt.Errorf("failed to get FRI estimate")
	}

	// TODO: does v3 tx uses fri instead of wei? check feeEstimate[0].FeeUnit?
	// pad estimate to 110%
	gasConsumed := friEstimate.GasConsumed.BigInt(new(big.Int))
	expandedGas := new(big.Int).Mul(gasConsumed, big.NewInt(140))
	maxGas := new(big.Int).Div(expandedGas, big.NewInt(100))
	tx.ResourceBounds.L1Gas.MaxAmount = starknetrpc.U64(starknetutils.BigIntToFelt(maxGas).String())

	// TODO: add margin
	tx.ResourceBounds.L1Gas.MaxPricePerUnit = starknetrpc.U128(friEstimate.GasPrice.String())

	txm.lggr.Infow("Set resource bounds", "L1MaxAmount", tx.ResourceBounds.L1Gas.MaxAmount, "L1MaxPricePerUnit", tx.ResourceBounds.L1Gas.MaxPricePerUnit)

	// Re-sign transaction now that we've determined MaxFee
	// TODO: SignInvokeTransaction for V3 is missing so we do it by hand
	hash, err = account.TransactionHashInvoke(tx)
	if err != nil {
		return txhash, err
	}
	signature, err = account.Sign(ctx, hash)
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
		var data any
		if err, ok := err.(ethrpc.DataError); ok {
			data = err.ErrorData()
		}
		txm.lggr.Errorw("failed to invoke tx", "error", err, "data", data)
		return txhash, fmt.Errorf("failed to invoke tx: %+w", err)
	}
	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	// update nonce if transaction is successful
	txhash = res.TransactionHash.String()
	err = errors.Join(
		txm.nonce.IncrementNextSequence(publicKey, chainID, nonce),
		txm.txStore.Save(accountAddress, nonce, txhash),
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

			hashes := txm.txStore.GetAllUnconfirmed()
			for addr := range hashes {
				for i := range hashes[addr] {
					hash := hashes[addr][i]
					f, err := starknetutils.HexToFelt(hash)
					if err != nil {
						txm.lggr.Errorw("invalid felt value", "hash", hash)
						continue
					}
					response, err := client.Provider.GetTransactionStatus(ctx, f)
					if err != nil {
						txm.lggr.Errorw("failed to fetch transaction status", "hash", hash, "error", err)
						continue
					}

					status := response.FinalityStatus

					// any status other than received
					if status == starknetrpc.TxnStatus_Accepted_On_L1 || status == starknetrpc.TxnStatus_Accepted_On_L2 || status == starknetrpc.TxnStatus_Rejected {
						txm.lggr.Debugw(fmt.Sprintf("tx confirmed: %s", status), "hash", hash, "status", status)
						if err := txm.txStore.Confirm(addr, hash); err != nil {
							txm.lggr.Errorw("failed to confirm tx in TxStore", "hash", hash, "sender", addr, "error", err)
						}
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

	chainID, err := client.Provider.ChainID(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get chainID: %+w", err)
	}

	// register account for nonce manager
	if err := txm.nonce.Register(context.TODO(), accountAddress, publicKey, chainID, client); err != nil {
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
	list := maps.Values(txm.txStore.GetAllInflightCount())
	for _, count := range list {
		unconfirmed += count
	}
	return len(txm.queue), unconfirmed
}
