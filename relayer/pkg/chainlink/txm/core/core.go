package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
)

const (
	MaxQueueLen = 1000
)

type txm[T any, K any, N any, E any] struct {
	starter utils.StartStopOnce
	lggr    logger.Logger
	done    sync.WaitGroup
	stop    chan struct{}
	queue   chan Tx[T]
	ks      Keystore[K]
	cfg     Config
	client  ChainClient[K, T, N, E]
	txs     TxStatuses[T]
}

func New[T any, K any, N any, E any](
	lggr logger.Logger,
	ks Keystore[K],
	cfg Config,
	client ChainClient[K, T, N, E],
	txStatuses TxStatuses[T],
) (TxManager[T], error) {
	// validate interface inputs are not nil
	if lggr == nil || ks == nil || cfg == nil || client == nil || txStatuses == nil {
		return nil, fmt.Errorf("txm inputs cannot be nil")
	}

	return &txm[T, K, N, E]{
		lggr:   lggr,
		queue:  make(chan Tx[T], MaxQueueLen),
		stop:   make(chan struct{}),
		ks:     ks,
		cfg:    cfg,
		client: client,
		txs:    txStatuses,
	}, nil
}

func (txm *txm[T, K, N, E]) Start(ctx context.Context) error {
	return txm.starter.StartOnce("txm", func() error {
		txm.done.Add(3) // waitgroup: tx sender, confirmer, retryer
		go txm.run()
		return nil
	})
}

func (txm *txm[T, K, N, E]) run() {
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
			key, err := txm.ks.Get(tx.Sender())
			if err != nil {
				txm.lggr.Errorw("failed to retrieve key", "id", tx.Sender(), "error", err)
				continue
			}

			// broadcast original transaction with custom nonce and max fee (if present)
			hash, broadcastErr := txm.broadcast(ctx, key, tx.Tx())
			if broadcastErr != nil {
				txm.lggr.Errorw("transaction failed to broadcast", "error", broadcastErr, "tx", tx)
				if err = txm.txs.Errored(tx.ID(), broadcastErr.Error()); err != nil {
					txm.lggr.Errorw("unable to save transaction status", "error", err, "tx", tx)
				}
				continue
			}
			txm.lggr.Infow("transaction broadcast", "txhash", hash)
			if err = txm.txs.Broadcast(tx.ID(), hash); err != nil {
				txm.lggr.Errorw("unable to save transaction status", "error", err, "tx", tx)
			}
		case <-txm.stop:
			return
		}
	}
}

func (txm *txm[T, K, N, E]) broadcast(ctx context.Context, key K, tx T) (txhash string, err error) {
	// get Nonce
	nonce, err := txm.client.GetNonce(ctx, key, tx)
	if err != nil {
		return txhash, errors.Wrap(err, "err in txm.getNonce")
	}

	// estimate/simulate Tx
	fee, err := txm.client.EstimateTx(ctx, key, tx, nonce)
	if err != nil {
		return txhash, errors.Wrap(err, "err in txm.estimateFee")
	}

	// broadcast transcation
	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout())
	defer execCancel()
	return txm.client.SendTx(execCtx, key, tx, nonce, fee)
}

func (txm *txm[T, K, N, E]) confirmer(ctx context.Context) {
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

			var wg sync.WaitGroup
			wg.Add(len(txs))
			txm.lggr.Debugw("txm confirmer", "count", len(txs), "txs", txs)
			for _, tx := range txs {
				go func(tx Tx[T]) {
					defer wg.Done()

					// get transaction status
					status, statusErrStr, err := txm.client.TxStatus(ctx, tx.Hash())
					if err != nil {
						txm.lggr.Errorw("failed to fetch tx status", "hash", tx.Hash())
						return
					}

					txm.lggr.Debugw("tx status", "hash", tx.Hash(), "status", status.String())

					// check if status is confirmed
					if status == CONFIRMED {
						if err = txm.txs.Confirmed(tx.ID()); err != nil {
							txm.lggr.Errorw("unable to save transaction status", "error", err, "id", tx.ID)
							return
						}
					}

					if status == ERRORED {
						txm.lggr.Errorw("tx rejected by sequencer", "hash", tx.Hash(), "error", statusErrStr)
						if err = txm.txs.Errored(tx.ID(), statusErrStr); err != nil {
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

func (txm *txm[T, K, N, E]) retryer(ctx context.Context) {
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

				// determine if fatal based on chain specific error
				if txm.client.IsFatalError(tx.Err()) {
					txm.lggr.Errorw("transaction fatally errored", "tx", tx.Tx(), "error", tx.Err())
					if err := txm.txs.Fatal(tx.ID()); err != nil {
						txm.lggr.Errorw("unable to save transaction", "error", err, "id", tx.ID())
					}
					continue
				}

				// retry other transactions (nonce error, gas error, endpoint goes down
				var err error
				tx, err = txm.txs.Retry(tx.ID())
				if err != nil {
					txm.lggr.Errorw("unable to update transaction", "error", err, "id", tx.ID())
				}
				txm.queue <- tx // requeue for retry
			}
		}
		tick = time.After(utils.WithJitter(txm.cfg.TxRetryFrequency()) - time.Since(start))
	}
}

func (txm *txm[T, K, N, E]) TxCount(state Status) int {
	return len(txm.txs.Get(state))
}

func (txm *txm[T, K, N, E]) Close() error {
	return txm.starter.StopOnce("txm", func() error {
		close(txm.stop)
		txm.done.Wait()
		return nil
	})
}

func (txm *txm[T, K, N, E]) Healthy() error {
	return txm.starter.Healthy()
}

func (txm *txm[T, K, N, E]) Ready() error {
	return txm.starter.Ready()
}

func (txm *txm[T, K, N, E]) Enqueue(tx T) error {
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