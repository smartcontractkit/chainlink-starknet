package txm

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const (
	MaxQueueLen        = 1000
	FEE_MARGIN  uint64 = 115
)

type TxManager interface {
	Enqueue(types.Transaction) error
}

type StarkTXM interface {
	relaytypes.Service
	TxManager
	TxCount(int) int
}

type starktxm struct {
	starter utils.StartStopOnce
	lggr    logger.Logger
	done    sync.WaitGroup
	stop    chan struct{}
	queue   chan transaction
	ks      keys.Keystore
	cfg     Config
	client  *utils.LazyLoad[starknet.ReaderWriter]
	txs     txStatuses
}

const (
	QUEUED int = iota
	RETRY      // can be reached from ERRORED
	BROADCAST
	CONFIRMED // ending state for happy path
	ERRORED
	FATAL // ending state for failed txs (reverts, etc)
)

type transaction struct {
	TX         types.Transaction
	ID         string
	Status     int
	Err        string
	RetryCount uint
}

// placeholder for managing transaction states
type txStatuses struct {
	txs  map[string]transaction
	lock sync.RWMutex
}

func (t *txStatuses) Get(state int) (out []transaction) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for k := range t.txs {
		if t.txs[k].Status == state {
			out = append(out, t.txs[k])
		}
	}
	return out
}

func (t *txStatuses) Exists(id string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	_, exists := t.txs[id]
	return exists
}

func (t *txStatuses) Queued(tx types.Transaction) (transaction, error) {
	id := uuid.New().String()

	// validate it exists
	if t.Exists(id) {
		return transaction{}, fmt.Errorf("transaction with id: %s already exists", id)
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	t.txs[id] = transaction{
		TX:     tx,
		ID:     id,
		Status: QUEUED,
	}
	return t.txs[id], nil
}

func (t *txStatuses) Retry(id string) (transaction, error) {
	if !t.Exists(id) {
		return transaction{}, fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].Status != ERRORED {
		return transaction{}, fmt.Errorf("tx must have errored before retry")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.TX.Nonce = ""  // clear if retrying
	tx.TX.MaxFee = "" // clear if retrying
	tx.Status = RETRY
	tx.RetryCount += 1
	t.txs[id] = tx
	return t.txs[id], nil
}

func (t *txStatuses) Broadcast(id, txhash string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].Status != QUEUED && t.txs[id].Status != RETRY {
		return fmt.Errorf("tx must be queued before broadcast")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.TX.TransactionHash = txhash
	tx.Status = BROADCAST
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) Confirmed(id string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].Status != BROADCAST {
		return fmt.Errorf("tx must be broadcast before confirmed")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.Status = CONFIRMED
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) Errored(id, err string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.Status = ERRORED
	tx.Err = err
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) Fatal(id string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].Status != ERRORED {
		return fmt.Errorf("tx must have errored before fatal")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.Status = FATAL
	t.txs[id] = tx
	return nil
}

func New(lggr logger.Logger, keystore keys.Keystore, cfg Config, getClient func() (starknet.ReaderWriter, error)) (StarkTXM, error) {
	return &starktxm{
		lggr:   lggr,
		queue:  make(chan transaction, MaxQueueLen),
		stop:   make(chan struct{}),
		ks:     keystore,
		cfg:    cfg,
		client: utils.NewLazyLoad(getClient),
		txs: txStatuses{
			txs: map[string]transaction{},
		},
	}, nil
}

func (txm *starktxm) Start(ctx context.Context) error {
	return txm.starter.StartOnce("starktxm", func() error {
		txm.done.Add(3) // waitgroup: tx sender, confirmer, retryer
		go txm.run()
		return nil
	})
}

func (txm *starktxm) run() {
	defer txm.done.Done()

	ctx, cancel := utils.ContextFromChan(txm.stop)
	defer cancel()

	// start retryer and confirmer
	go txm.confirmer(ctx)
	go txm.retryer(ctx)

	// process new txs as they come in
	for {
		select {
		case tx := <-txm.queue:
			// fetch key matching sender address
			key, err := txm.ks.Get(tx.TX.SenderAddress)
			if err != nil {
				txm.lggr.Errorw("failed to retrieve key", "id", tx.TX.SenderAddress, "error", err)
				continue
			}

			// parse key to expected format
			privKeyBytes := key.Raw()
			privKey := caigo.BigToHex(caigo.BytesToBig(privKeyBytes))

			// broadcast original transaction with custom nonce and max fee (if present)
			hash, broadcastErr := txm.broadcast(ctx, privKey, tx.TX)
			if broadcastErr != nil {
				txm.lggr.Errorw("transaction failed to broadcast", "error", broadcastErr, "tx", tx)
				if err = txm.txs.Errored(tx.ID, broadcastErr.Error()); err != nil {
					txm.lggr.Errorw("unable to save transaction status", "error", err, "tx", tx)
				}
				continue
			}
			txm.lggr.Infow("transaction broadcast", "txhash", hash)
			if err = txm.txs.Broadcast(tx.ID, hash); err != nil {
				txm.lggr.Errorw("unable to save transaction status", "error", err, "tx", tx)
			}
		case <-txm.stop:
			return
		}
	}
}

func (txm *starktxm) broadcast(ctx context.Context, privKey string, tx types.Transaction) (txhash string, err error) {
	txs := []types.Transaction{tx} // assemble into slice
	client, err := txm.client.Get()
	if err != nil {
		return txhash, errors.Errorf("broadcast: failed to fetch client: %s", err)
	}

	// create new account
	account, err := caigo.NewAccount(privKey, tx.SenderAddress, client)
	if err != nil {
		return txhash, errors.Errorf("failed to create account: %s", err)
	}

	details := caigo.ExecuteDetails{}

	// allow custom passed nonce
	if tx.Nonce != "" {
		nonce, ok := new(big.Int).SetString(tx.Nonce, 0) // use base 0 to dynamically determine base
		if !ok {
			return txhash, errors.Errorf("failed to decode custom nonce: %s", tx.Nonce)
		}
		details.Nonce = nonce
	} else {
		// TODO: investigate if nonce management is needed (nonce is requested queried by the sdk for now)
		// get nonce
		nonce, err := client.AccountNonce(ctx, tx.SenderAddress)
		if err != nil {
			return txhash, errors.Errorf("failed to fetch nonce: %s", err)
		}
		details.Nonce = nonce
	}

	// allow fustom passed max fee
	if tx.MaxFee != "" {
		fee, ok := new(big.Int).SetString(tx.MaxFee, 0) // use base 0 to dynamically determine base
		if !ok {
			return txhash, errors.Errorf("failed to decode custom max fee: %s", tx.MaxFee)
		}
		details.MaxFee = types.BigToFelt(fee)
	} else {
		// nonce management + fee estimator go together (otherwise too high of nonce will cause estimate to fail)
		// get fee for txm
		fee, err := account.EstimateFee(ctx, txs, details)
		if err != nil {
			return txhash, errors.Errorf("failed to estimate fee: %s", err)
		}
		details.MaxFee = types.BigToFelt(new(big.Int).SetUint64((fee.OverallFee * FEE_MARGIN) / 100))
	}

	// transmit txs
	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout())
	defer execCancel()
	res, err := account.Execute(execCtx, txs, details)
	if err != nil {
		return txhash, errors.Errorf("failed to invoke tx: %s", err)
	}

	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	return res.TransactionHash, nil
}

// look through transactions to see which can be confirmed
func (txm *starktxm) confirmer(ctx context.Context) {
	defer txm.done.Done()

	tick := time.After(0) // immediately try confirming any unconfirmed
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			start = time.Now()
			txs := txm.txs.Get(BROADCAST)

			client, err := txm.client.Get()
			if err != nil {
				txm.lggr.Errorw("confirmer: failed to fetch client", "error", err)
				break // exit select
			}

			var wg sync.WaitGroup
			wg.Add(len(txs))
			txm.lggr.Debugw("txm confirmer", "count", len(txs), "txs", txs)
			for _, tx := range txs {
				go func(tx transaction) {
					defer wg.Done()

					status, err := client.TransactionStatus(ctx, gateway.TransactionStatusOptions{
						TransactionHash: tx.TX.TransactionHash,
					})
					if err != nil {
						txm.lggr.Errorw("failed to fetch receipt", "hash", tx.TX.TransactionHash)
						return
					}

					txm.lggr.Debugw("tx status", "hash", tx.TX.TransactionHash, "status", status.TxStatus)

					if strings.Contains(status.TxStatus, "ACCEPTED") {
						if err = txm.txs.Confirmed(tx.ID); err != nil {
							txm.lggr.Errorw("unable to save transaction status", "error", err, "id", tx.ID)
							return
						}
					}

					if status.TxStatus == "REJECTED" {
						txm.lggr.Errorw("tx rejected by sequencer", "hash", tx.TX.TransactionHash, "error", status.TxFailureReason.ErrorMessage)
						if err = txm.txs.Errored(tx.ID, status.TxFailureReason.ErrorMessage); err != nil {
							txm.lggr.Errorw("unable to save transaction status", "error", err, "id", tx.ID)
							return
						}
					}

				}(tx)
			}
			wg.Wait()
		}
		tick = time.After(utils.WithJitter(txm.cfg.TxConfirmFrequency()) - time.Since(start))
	}
}

// look through txs to determine if tx should be retried
// TODO: define what conditions to (or not to) retry, everything else will retry
func (txm *starktxm) retryer(ctx context.Context) {
	defer txm.done.Done()

	tick := time.After(0)
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			start = time.Now()
			txs := txm.txs.Get(ERRORED)
			txm.lggr.Debugw("txm retryer", "count", len(txs), "txs", txs)
			for i := range txs {
				tx := txs[i]

				// Fatal condition - error revert in contract
				// https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/starknet/business_logic/execution/execute_entry_point.py#L237
				// TODO: verify if this is true
				if strings.Contains(tx.Err, "Error in the called contract") {
					txm.lggr.Errorw("transaction fatally errored", "tx", tx.TX)
					if err := txm.txs.Fatal(tx.ID); err != nil {
						txm.lggr.Errorw("unable to save transaction", "error", err, "id", tx.ID)
					}
					continue
				}

				// retry other transactions (nonce error, gas error, endpoint goes down
				var err error
				tx, err = txm.txs.Retry(tx.ID)
				if err != nil {
					txm.lggr.Errorw("unable to update transaction", "error", err, "id", tx.ID)
				}
				txm.queue <- tx // requeue for retry
			}
		}
		tick = time.After(utils.WithJitter(txm.cfg.TxRetryFrequency()) - time.Since(start))
	}
}

func (txm *starktxm) TxCount(state int) int {
	return len(txm.txs.Get(state))
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

func (txm *starktxm) Enqueue(tx types.Transaction) error {
	// add transaction to storage to preserve in case of later error
	queuedTx, err := txm.txs.Queued(tx)
	if err != nil {
		return fmt.Errorf("unable to save transaction status: %s : %+v", err, tx)
	}

	select {
	case txm.queue <- queuedTx:
	default:
		return errors.Errorf("failed to enqueue transaction: %+v", tx)
	}

	return nil
}
