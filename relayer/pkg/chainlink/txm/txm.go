package txm

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
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
	InflightCount() int
}

type starktxm struct {
	starter utils.StartStopOnce
	lggr    logger.Logger
	done    sync.WaitGroup
	stop    chan struct{}
	queue   chan types.Transaction
	ks      keys.Keystore
	cfg     Config
	client  *utils.LazyLoad[types.Provider]
	txs     txStatuses
}

const (
	QUEUED int = iota
	BROADCAST
	CONFIRMED
	ERRORED
)

type transaction struct {
	tx     types.Transaction
	id     string
	status int
}

// placeholder for managing transaction states
type txStatuses struct {
	txs  map[string]transaction
	lock sync.RWMutex
}

func (t *txStatuses) Queued(tx types.Transaction) (id string, err error) {
	id = uuid.New().String()

	// validate it exists
	if t.Exists(id) {
		return id, fmt.Errorf("transaction with id: %s already exists", id)
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	t.txs[id] = transaction{
		tx:     tx,
		id:     id,
		status: QUEUED,
	}
	return id, nil
}

func (t *txStatuses) Exists(id string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	_, exists := t.txs[id]
	return exists
}

func (t *txStatuses) Broadcast(id, txhash string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].status != QUEUED {
		return fmt.Errorf("tx must be queued before broadcast")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.tx.TransactionHash = txhash
	tx.status = BROADCAST
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) GetUnconfirmed() (out []transaction) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for k := range t.txs {
		if t.txs[k].status == BROADCAST {
			out = append(out, t.txs[k])
		}
	}
	return out
}

func (t *txStatuses) Confirmed(id string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.RLock()
	if t.txs[id].status != BROADCAST {
		return fmt.Errorf("tx must be broadcast before confirmed")
	}
	t.lock.RUnlock()

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.status = CONFIRMED
	t.txs[id] = tx
	return nil
}

func (t *txStatuses) Errored(id string) error {
	if !t.Exists(id) {
		return fmt.Errorf("id does not exist")
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	tx := t.txs[id]
	tx.status = ERRORED
	t.txs[id] = tx
	return nil
}

func New(lggr logger.Logger, keystore keys.Keystore, cfg Config, getClient func() (types.Provider, error)) (StarkTXM, error) {
	return &starktxm{
		lggr:   lggr,
		queue:  make(chan types.Transaction, MaxQueueLen),
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
			// add transaction to storage to preserve in case of later error
			id, err := txm.txs.Queued(tx)
			if err != nil {
				txm.lggr.Errorw("unable to save transaction status", "error", err, "tx", tx)
				continue
			}

			// fetch key matching sender address
			key, err := txm.ks.Get(tx.SenderAddress)
			if err != nil {
				txm.lggr.Errorw("failed to retrieve key", "id", tx.SenderAddress, "error", err)
				continue
			}

			// parse key to expected format
			privKeyBytes := key.Raw()
			privKey := caigo.BigToHex(caigo.BytesToBig(privKeyBytes))

			// broadcast
			hash, err := txm.broadcast(ctx, privKey, tx)
			if err != nil {
				txm.lggr.Errorw("transaction failed to broadcast", "error", err, "tx", tx)
				if err = txm.txs.Errored(id); err != nil {
					txm.lggr.Errorw("unable to save transaction status", "error", err, "tx", tx)
				}
				continue
			}
			txm.lggr.Infow("transaction broadcast", "txhash", hash)
			if err = txm.txs.Broadcast(id, hash); err != nil {
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
		return txhash, errors.Errorf("failed to fetch client: %s", err)
	}

	// create new account
	account, err := caigo.NewAccount(privKey, tx.SenderAddress, client)
	if err != nil {
		return txhash, errors.Errorf("failed to create account: %s", err)
	}

	// TODO: investigate if nonce management is needed (nonce is requested queried by the sdk for now)
	// get nonce
	nonce, err := client.AccountNonce(ctx, tx.SenderAddress)
	details := caigo.ExecuteDetails{
		Nonce: nonce,
	}

	// get fee for txm
	fee, err := account.EstimateFee(ctx, txs, details)
	if err != nil {
		return txhash, errors.Errorf("failed to estimate fee: %s", err)
	}

	details.MaxFee = &types.Felt{
		Int: new(big.Int).SetUint64((fee.OverallFee * FEE_MARGIN) / 100),
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

func (txm *starktxm) confirmer(ctx context.Context) {
	defer txm.done.Done()

	tick := time.After(0) // immediately try confirming any unconfirmed
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			txs := txm.txs.GetUnconfirmed()

			client, err := txm.client.Get()
			if err != nil {
				txm.lggr.Errorw("confirmer: failed to fetch client", "error", err)
				break // exit select
			}

			var wg sync.WaitGroup
			wg.Add(len(txs))
			for _, tx := range txs {
				go func(tx transaction) {
					defer wg.Done()

					receipt, err := client.TransactionReceipt(ctx, tx.tx.TransactionHash)
					if err != nil {
						txm.lggr.Errorw("failed to fetch receipt", "hash", tx.tx.TransactionHash)
						return
					}
					txm.lggr.Infow("tx status", "hash", tx.tx.TransactionHash, "status", receipt.Status)

					if strings.Contains(receipt.Status, "ACCEPTED") {
						if err = txm.txs.Confirmed(tx.id); err != nil {
							txm.lggr.Errorw("unable to save transaction status", "error", err, "id", tx.id)
							return
						}
					}

				}(tx)
			}
			wg.Wait()
		}
		tick = time.After(5 * time.Second) // TODO: set in config
	}
}

// look through txs to determine if tx should be retried
func (txm *starktxm) retryer(ctx context.Context) {
	defer txm.done.Done()

	tick := time.After(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			// do something
		}
	}
}

func (txm *starktxm) InflightCount() int {
	return len(txm.txs.GetUnconfirmed())
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
	select {
	case txm.queue <- tx:
	default:
		return errors.Errorf("failed to enqueue transaction: %+v", tx)
	}

	return nil
}
