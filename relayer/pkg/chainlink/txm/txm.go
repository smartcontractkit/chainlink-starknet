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

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const (
	MaxQueueLen           = 1000
	ConfirmationThreshold = 4
)

type TxManager interface {
	Enqueue(ctx context.Context, accountAddress *felt.Felt, publicKey *felt.Felt, txFn starknetrpc.FunctionCall) error
	InflightCount() (int, int)
}

// TxMetrics interface for v2 TXM metrics
type TxMetrics interface {
	IncrementNumBroadcastedTxs(ctx context.Context)
	IncrementNumConfirmedTxs(ctx context.Context, confirmedTransactions int)
	IncrementNumNonceGaps(ctx context.Context)
	ReachedMaxAttempts(ctx context.Context, reached bool)
	RecordTimeUntilTxConfirmed(ctx context.Context, duration float64)
	IncrementQueueFullEvents(ctx context.Context)
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

	client       *utils.LazyLoad[*starknet.Client]
	feederClient *utils.LazyLoad[*starknet.FeederClient]
	accountStore *AccountStore
	metrics      TxMetrics
	chainID      string

	// Track broadcast times for confirmation duration metrics
	broadcastTimes sync.Map // map[string]time.Time (tx hash -> broadcast time)

	// Track retry attempts for max attempts detection
	retryAttempts sync.Map // map[string]int (tx hash -> attempt count)
	maxAttempts   int      // Maximum number of attempts before marking as failed
}

func New(lggr logger.Logger, keystore loop.Keystore, cfg Config, chainID string, getClient func() (*starknet.Client, error),
	getFeederClient func() (*starknet.FeederClient, error)) (StarkTXM, error) {
	return NewWithMetrics(lggr, keystore, cfg, chainID, NewPrometheusMetrics(chainID), getClient, getFeederClient)
}

func NewWithMetrics(lggr logger.Logger, keystore loop.Keystore, cfg Config, chainID string, metrics TxMetrics, getClient func() (*starknet.Client, error),
	getFeederClient func() (*starknet.FeederClient, error)) (StarkTXM, error) {
	txm := &starktxm{
		lggr:         logger.Named(lggr, "Txm"),
		queue:        make(chan Tx, MaxQueueLen),
		stop:         make(chan struct{}),
		client:       utils.NewLazyLoad(getClient),
		feederClient: utils.NewLazyLoad(getFeederClient),
		ks:           NewKeystoreAdapter(keystore),
		cfg:          cfg,
		accountStore: NewAccountStore(),
		metrics:      metrics,
		chainID:      chainID,
		maxAttempts:  cfg.MaxAttempts(),
	}

	return txm, nil
}

func (txm *starktxm) Name() string {
	return txm.lggr.Name()
}

func (txm *starktxm) Start(ctx context.Context) error {
	return txm.starter.StartOnce("Txm", func() error {
		txm.done.Add(2) // waitgroup: broadcast loop and confirm loop
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
				// Track retry attempts for max attempts detection
				txm.trackRetryAttempt(ctx, hash)
			} else {
				txm.lggr.Infow("transaction broadcast", "txhash", hash)
				// Increment broadcasted transactions metric
				txm.metrics.IncrementNumBroadcastedTxs(ctx)
				// Track broadcast time for confirmation duration metrics
				txm.broadcastTimes.Store(hash, time.Now())
				// Clear retry attempts on successful broadcast
				txm.retryAttempts.Delete(hash)
			}
		}
	}
}

const FeeMargin uint32 = 115
const RPCNonceErrMsg = "Invalid transaction nonce"

func (txm *starktxm) estimateFriFee(ctx context.Context, client *starknet.Client, accountAddress *felt.Felt, tx starknetrpc.BroadcastInvokeTxnV3) (*starknetrpc.FeeEstimation, *felt.Felt, error) {
	// skip prevalidation, which is known to overestimate amount of gas needed and error with L1GasBoundsExceedsBalance
	simFlags := []starknetrpc.SimulationFlag{starknetrpc.SKIP_VALIDATE}

	var largestEstimateNonce *felt.Felt

	for i := 1; i <= txm.maxAttempts; i++ {
		txm.lggr.Infow("attempt to estimate fee", "attempt", i)

		estimateNonce, err := client.AccountNonce(ctx, accountAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check account nonce: %+w", err)
		}
		tx.InvokeTxnV3.Nonce = estimateNonce

		if largestEstimateNonce == nil || estimateNonce.Cmp(largestEstimateNonce) > 0 {
			largestEstimateNonce = estimateNonce
		}

		feeEstimate, err := client.Provider.EstimateFee(ctx, []starknetrpc.BroadcastTxn{tx}, simFlags, starknetrpc.BlockID{Tag: "pending"})
		if err != nil {
			var dataErr *starknetrpc.RPCError
			if !errors.As(err, &dataErr) {
				return nil, nil, fmt.Errorf("failed to read EstimateFee error: %T %+v", err, err)
			}
			data := dataErr.Data
			dataStr := fmt.Sprintf("%+v", data)

			txm.lggr.Errorw("failed to estimate fee", "attempt", i, "error", err, "data", dataStr)

			if strings.Contains(dataStr, RPCNonceErrMsg) {
				continue
			}

			return nil, nil, fmt.Errorf("Failed to estimate fee: %T %+v", err, err)
		}

		// track the FRI estimate, but keep looping so we print out all estimates
		var friEstimate *starknetrpc.FeeEstimation
		for j, f := range feeEstimate {
			txm.lggr.Infow("Estimated fee", "attempt", i, "index", j, "EstimateNonce", estimateNonce, "L1GasConsumed", f.L1GasConsumed, "L1GasPrice", f.L1GasPrice, "L1DataGasConsumed", f.L1DataGasConsumed, "L1DataGasPrice", f.L1DataGasPrice,
				"L2GasConsumed", f.L2GasConsumed, "L2GasPrice", f.L2GasPrice, "OverallFee", f.OverallFee, "FeeUnit", string(f.FeeUnit))
			if f.FeeUnit == "FRI" {
				friEstimate = &feeEstimate[j]
			}
		}
		if friEstimate != nil {
			return friEstimate, largestEstimateNonce, nil
		}

		txm.lggr.Errorw("No FRI estimate was returned", "attempt", i)
	}

	txm.lggr.Errorw("all attempts to estimate fee failed")
	txm.metrics.ReachedMaxAttempts(ctx, true)
	return nil, nil, fmt.Errorf("all attempts to estimate fee failed")
}

func (txm *starktxm) broadcast(ctx context.Context, publicKey *felt.Felt, accountAddress *felt.Felt, call starknetrpc.FunctionCall) (txhash string, err error) {
	client, err := txm.client.Get()
	if err != nil {
		txm.client.Reset()
		return txhash, fmt.Errorf("broadcast: failed to fetch client: %+w", err)
	}

	txStore := txm.accountStore.GetTxStore(accountAddress)
	if txStore == nil {
		initialNonce, accountNonceErr := client.AccountNonce(ctx, accountAddress)
		if accountNonceErr != nil {
			return txhash, fmt.Errorf("failed to check account nonce during TxStore creation: %+w", accountNonceErr)
		}
		newTxStore, createErr := txm.accountStore.CreateTxStore(accountAddress, initialNonce, txm.lggr)
		if createErr != nil {
			return txhash, fmt.Errorf("failed to create TxStore: %+w", createErr)
		}
		txStore = newTxStore
	}

	// create new account
	cairoVersion := 2
	account, err := starknetaccount.NewAccount(client.Provider, accountAddress, publicKey.String(), txm.ks, cairoVersion)
	if err != nil {
		return txhash, fmt.Errorf("failed to create new account: %+w", err)
	}

	tx := starknetrpc.InvokeTxnV3{
		Type:          starknetrpc.TransactionType_Invoke,
		SenderAddress: account.Address,
		Version:       starknetrpc.TransactionV3,
		Signature:     []*felt.Felt{},
		Nonce:         &felt.Zero, // filled in below
		ResourceBounds: starknetrpc.ResourceBoundsMapping{
			L1Gas: starknetrpc.ResourceBounds{
				MaxAmount:       "0x0",
				MaxPricePerUnit: "0x0",
			},
			// New starknet cannot resolve amounts as 0x0
			L1DataGas: starknetrpc.ResourceBounds{
				MaxAmount:       "0x0",
				MaxPricePerUnit: "0x0",
			},
			L2Gas: starknetrpc.ResourceBounds{
				MaxAmount:       "0x01",
				MaxPricePerUnit: "0x01",
			},
		},
		Tip:                   "0x0",
		PayMasterData:         []*felt.Felt{},
		AccountDeploymentData: []*felt.Felt{},
		NonceDataMode:         starknetrpc.DAModeL1,
		FeeMode:               starknetrpc.DAModeL1,
	}

	// Building the Calldata with the help of FmtCalldata where we pass in the FnCall struct along with the Cairo version
	tx.Calldata, err = account.FmtCalldata([]starknetrpc.FunctionCall{call})
	if err != nil {
		return txhash, err
	}

	broadcastTxnV3 := starknetrpc.BroadcastInvokeTxnV3{
		InvokeTxnV3: tx,
	}

	friEstimate, largestEstimateNonce, err := txm.estimateFriFee(ctx, client, accountAddress, broadcastTxnV3)
	if err != nil {
		return txhash, fmt.Errorf("failed to get FRI estimate: %+w", err)
	}

	nonce := txStore.GetNextNonce()
	if largestEstimateNonce.Cmp(nonce) > 0 {
		// The nonce value returned from the node during estimation is greater than our expected next nonce
		// - which means that we are behind, due to a resync. Fast forward our locally tracked nonce value.
		// See resyncNonce for a more detailed explanation.
		staleTxs := txStore.SetNextNonce(largestEstimateNonce)
		txm.lggr.Infow("fast-forwarding nonce after resync", "previousNonce", nonce, "updatedNonce", largestEstimateNonce, "staleTxs", len(staleTxs))
		if len(staleTxs) > 0 {
			txm.lggr.Errorw("unexpected stale transactions after nonce fast-forward", "accountAddress", accountAddress)
		}
		nonce = largestEstimateNonce
	}

	L2GasConsumed := friEstimate.L2GasConsumed.BigInt(new(big.Int))
	broadcastTxnV3.InvokeTxnV3.ResourceBounds.L2Gas.MaxAmount = txm.updateMaxAmountBounds(L2GasConsumed, 150)

	L1GasPrice := friEstimate.L1GasPrice.BigInt(new(big.Int))
	L2GasPrice := friEstimate.L2GasPrice.BigInt(new(big.Int))

	L1GasConsumed := friEstimate.L1GasConsumed.BigInt(new(big.Int))
	// TODO: consider making this configurable
	// pad estimate to 150% (add extra because estimate did not include validation)
	broadcastTxnV3.InvokeTxnV3.ResourceBounds.L1Gas.MaxAmount = txm.updateMaxAmountBounds(L1GasConsumed, 150)

	// pad by 150%
	broadcastTxnV3.InvokeTxnV3.ResourceBounds.L1Gas.MaxPricePerUnit = txm.updateMaxPriceUnitBounds(L1GasPrice, 150)
	broadcastTxnV3.InvokeTxnV3.ResourceBounds.L2Gas.MaxPricePerUnit = txm.updateMaxPriceUnitBounds(L2GasPrice, 150)

	txm.lggr.Infow("Set resource bounds", "L1MaxAmount", tx.ResourceBounds.L1Gas.MaxAmount, "L1MaxPricePerUnit", tx.ResourceBounds.L1Gas.MaxPricePerUnit, "FinalNonce", nonce)

	L1DataGasConsumed := friEstimate.L1DataGasConsumed.BigInt(new(big.Int))
	L1DataGasPrice := friEstimate.L1DataGasPrice.BigInt(new(big.Int))
	broadcastTxnV3.InvokeTxnV3.ResourceBounds.L1DataGas.MaxAmount = txm.updateMaxAmountBounds(L1DataGasConsumed, 150)
	broadcastTxnV3.InvokeTxnV3.ResourceBounds.L1DataGas.MaxPricePerUnit = txm.updateMaxPriceUnitBounds(L1DataGasPrice, 150)

	broadcastTxnV3.InvokeTxnV3.Nonce = nonce

	err = account.SignInvokeTransaction(ctx, &broadcastTxnV3.InvokeTxnV3)
	if err != nil {
		return txhash, err
	}

	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout())
	defer execCancel()

	// finally, transmit the invoke
	res, err := account.Provider.AddInvokeTransaction(execCtx, &broadcastTxnV3)
	if err != nil {
		txm.lggr.Errorw("failed to invoke tx", "accountAddress", accountAddress, "error", err)
		if strings.Contains(err.Error(), RPCNonceErrMsg) {
			// if we see an invalid nonce error at the broadcast stage, that means that we are out of sync.
			// see the comment at resyncNonce for more details.
			if resyncErr := txm.resyncNonce(ctx, client, accountAddress); resyncErr != nil {
				txm.lggr.Errorw("failed to resync nonce after unsuccessful invoke", "error", err, "resyncError", resyncErr)
				return txhash, fmt.Errorf("failed to resync after bad invoke: %+w", err)
			}
		}
		return txhash, fmt.Errorf("failed to invoke tx: %+w", err)
	}
	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	// update nonce if transaction is successful
	txhash = res.TransactionHash.String()
	err = txStore.AddUnconfirmed(nonce, txhash, call, publicKey)
	if err != nil {
		return txhash, fmt.Errorf("failed to add unconfirmed tx: %+w", err)
	}
	return txhash, nil
}

func (txm *starktxm) updateMaxAmountBounds(gasConsumed *big.Int, padding int64) starknetrpc.U64 {
	expandedGas := new(big.Int).Mul(gasConsumed, big.NewInt(padding))
	maxGas := new(big.Int).Div(expandedGas, big.NewInt(100))

	return starknetrpc.U64(starknetutils.BigIntToFelt(maxGas).String())
}

func (txm *starktxm) updateMaxPriceUnitBounds(gasPrice *big.Int, padding int64) starknetrpc.U128 {
	expandedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(padding))
	maxGasPrice := new(big.Int).Div(expandedGasPrice, big.NewInt(100))

	return starknetrpc.U128(starknetutils.BigIntToFelt(maxGasPrice).String())
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

			for _, accountAddressStr := range txm.accountStore.Accounts() {
				accountAddress, err := new(felt.Felt).SetString(accountAddressStr)
				// this should never occur because the acccount address string key was created from the account address felt.
				if err != nil {
					txm.lggr.Errorw("could not recreate account address felt", "accountAddress", accountAddressStr)
					continue
				}
				nonce, err := client.AccountNonceLatest(ctx, accountAddress)
				if err != nil {
					txm.lggr.Errorf("failed to fetch latest nonce for account %v, err: %v", accountAddress, err)
					continue
				}
				// Confirm all transactions with nonce lower than the latest.
				confirmed, highestUnconfirmed := txm.accountStore.GetTxStore(accountAddress).Confirm(nonce)
				txm.lggr.Infow("Confirmation loop", "accountAddress", accountAddress, "latestNonce", nonce,
					"transactionsConfirmed", confirmed, "highestUnconfirmed", highestUnconfirmed)

				// Increment confirmed transactions metric
				if confirmed > 0 {
					txm.metrics.IncrementNumConfirmedTxs(ctx, confirmed)
					// Record confirmation duration for confirmed transactions
					// Since we can't track individual transactions, we'll estimate based on confirmation poll interval
					// This is a reasonable approximation for batch confirmations
					estimatedDuration := float64(txm.cfg.ConfirmationPoll().Seconds()) * 0.5 // Half the poll interval as average
					txm.metrics.RecordTimeUntilTxConfirmed(ctx, estimatedDuration)
				}

				// We add a maximum threshold between latest nonce and highest unconfirmed. This prevents the TXM from sending a very large
				// number of unconfirmed transactions in the mempool and triggers a resync to prevent nonce gaps since the RPC responses are unreliable.
				// The nonce stored here won't necessarily be picked up by the next transaction since there is a fast-forward functionality in broadcasting.
				// But it ensures that if for whatever reason the diff between mined and uncofirmed transactions starts to grow, the TXM will be able to
				// go back on the nonce and fill any nonce gaps.
				hu := highestUnconfirmed.BigInt(new(big.Int))
				n := nonce.BigInt(new(big.Int))
				threshold := big.NewInt(ConfirmationThreshold)
				if new(big.Int).Sub(hu, n).Cmp(threshold) == 1 {
					if resyncErr := txm.resyncNonce(ctx, client, accountAddress); resyncErr != nil {
						txm.lggr.Errorw("resync failed for rejected tx", "error", resyncErr)
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

func (txm *starktxm) resyncNonce(ctx context.Context, client *starknet.Client, accountAddress *felt.Felt) error {
	/*
	   the follow errors indicate that there could be a problem with our locally tracked nonce value:
	       1. a EstimateFee was successful, but broadcasting using the locally tracked nonce results in a nonce error,
	       2. a transaction was rejected after a successful broadcast.

	   for these cases, we call starknet_getNonce from the RPC node and resync the locally tracked next nonce
	   with the RPC node's value.

	   however, while the value returned by starknet_getNonce is eventually consistent, it can be lower than the actual
	   next nonce value when pending transactions haven't yet been processed - resulting in more category 1
	   invalid nonce broadcast errors.

	   in order to recover from these cases, each time we do starknet_getNonce during estimation (see estimateFriFee),
	   we compare it with our locally tracked nonce - if it is greater, than that means our locally tracked value is
	   behind, and we fast forward. this ensures our locally tracked value will also eventually be correct.
	*/

	rpcNonce, err := client.AccountNonce(ctx, accountAddress)
	if err != nil {
		return fmt.Errorf("failed to check nonce during resync: %+w", err)
	}

	txStore := txm.accountStore.GetTxStore(accountAddress)
	currentNonce := txStore.GetNextNonce()

	if rpcNonce.Cmp(currentNonce) == 0 {
		txm.lggr.Infow("resync nonce skipped, nonce value is the same", "accountAddress", accountAddress, "nonce", currentNonce)
		return nil
	}

	staleTxs := txStore.SetNextNonce(rpcNonce)

	txm.lggr.Infow("resynced nonce", "accountAddress", "accountAddress", "previousNonce", currentNonce, "updatedNonce", rpcNonce, "staleTxCount", len(staleTxs))

	// Increment nonce gaps metric when stale transactions are found
	if len(staleTxs) > 0 {
		txm.metrics.IncrementNumNonceGaps(ctx)
	}

	return nil
}

// trackRetryAttempt tracks retry attempts and updates max attempts metric
func (txm *starktxm) trackRetryAttempt(ctx context.Context, txHash string) {
	if txHash == "" {
		return // Skip empty hashes
	}

	// Get current attempt count
	currentAttempts := 0
	if val, ok := txm.retryAttempts.Load(txHash); ok {
		currentAttempts = val.(int)
	}

	// Increment attempt count
	newAttempts := currentAttempts + 1
	txm.retryAttempts.Store(txHash, newAttempts)

	// Check if max attempts reached
	if newAttempts >= txm.maxAttempts {
		txm.lggr.Warnw("transaction reached max attempts", "txHash", txHash, "attempts", newAttempts)
		txm.metrics.ReachedMaxAttempts(ctx, true)
	} else {
		txm.metrics.ReachedMaxAttempts(ctx, false)
	}
}

func (txm *starktxm) Close() error {
	return txm.starter.StopOnce("Txm", func() error {
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

func (txm *starktxm) Enqueue(ctx context.Context, accountAddress, publicKey *felt.Felt, tx starknetrpc.FunctionCall) error {
	// validate key exists for sender
	// use the embedded Loopp Keystore to do this; the spec and design
	// encourage passing nil data to the loop.Keystore.Sign as way to test
	// existence of a key
	if _, err := txm.ks.Loopp().Sign(ctx, publicKey.String(), nil); err != nil {
		return fmt.Errorf("enqueue: failed to sign: %+w", err)
	}

	select {
	case txm.queue <- Tx{publicKey: publicKey, accountAddress: accountAddress, call: tx}: // TODO fix naming here
	default:
		// Queue is full - this could indicate high load or processing issues
		txm.metrics.IncrementQueueFullEvents(ctx)
		return fmt.Errorf("failed to enqueue transaction: %+v", tx)
	}

	return nil
}

func (txm *starktxm) InflightCount() (queue int, unconfirmed int) {
	return len(txm.queue), txm.accountStore.GetTotalInflightCount()
}
