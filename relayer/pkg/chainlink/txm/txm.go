package txm

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
)

const (
	MaxQueueLen = 1000
)

type TxManager interface {
	Enqueue(types.Transaction) error
}

type StarkTXM interface {
	relaytypes.Service
	TxManager
}

type starktxm struct {
	starter utils.StartStopOnce
	lggr    logger.Logger
	done    sync.WaitGroup
	stop    chan struct{}
	queue   chan types.Transaction
	curve   *caigo.StarkCurve
	ks      keys.Keystore
	cfg     TxConfig

	// TODO: use lazy loaded client
	client    types.Provider
	getClient func() (types.Provider, error)
}

func New(lggr logger.Logger, keystore keys.Keystore, cfg TxConfig, getClient func() (types.Provider, error)) (StarkTXM, error) {
	curve, err := caigo.SC(caigo.WithConstants())
	if err != nil {
		return nil, errors.Errorf("failed to build curve: %s", err)
	}
	return &starktxm{
		lggr:      lggr,
		queue:     make(chan types.Transaction, MaxQueueLen),
		stop:      make(chan struct{}),
		getClient: getClient,
		curve:     &curve,
		ks:        keystore,
		cfg:       cfg,
	}, nil
}

func (txm *starktxm) Start(ctx context.Context) error {
	return txm.starter.StartOnce("starktxm", func() error {
		txm.done.Add(1) // waitgroup: tx sender
		go txm.run()
		return nil
	})
}

func (txm *starktxm) run() {
	defer txm.done.Done()

	// TODO: func not available without importing core
	// ctx, cancel := utils.ContextFromChan(txm.stop)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tick := time.After(0)

	for {
		select {
		case <-tick:
			start := time.Now()

			// calculate total txs to process
			txLen := len(txm.queue)
			if txLen > txm.cfg.TxMaxBatchSize() {
				txLen = txm.cfg.TxMaxBatchSize()
			}

			// fetch batch and split by sender
			txsBySender := map[string][]types.Transaction{}
			for i := 0; i < txLen; i++ {
				tx := <-txm.queue
				txsBySender[tx.SenderAddress] = append(txsBySender[tx.SenderAddress], tx)
			}
			txm.lggr.Infow("creating batch", "totalTxCount", txLen, "batchCount", len(txsBySender))
			txm.lggr.Debugw("batch details", "batches", txsBySender)

			// fetch client if needed
			if txm.client == nil {
				var err error
				txm.client, err = txm.getClient()
				if err != nil {
					txm.lggr.Errorw("unable to fetch client", "error", err)
				}
			}

			// async process of tx batches
			var wg sync.WaitGroup
			wg.Add(len(txsBySender))
			for sender, txs := range txsBySender {
				go func(sender string, txs []types.Transaction) {

					// fetch key matching sender address
					key, err := txm.ks.Get(sender)
					if err != nil {
						txm.lggr.Errorw("failed to retrieve key", "id", sender, "error", err)
					} else {
						// parse key to expected format
						privKeyBytes := key.Raw()
						privKey := caigo.BigToHex(caigo.BytesToBig(privKeyBytes))

						// broadcast batch based on sender
						hash, err := txm.broadcastBatch(ctx, privKey, sender, txs)
						if err != nil {
							txm.lggr.Errorw("transaction failed to broadcast", "error", err, "batchTx", txs)
						} else {
							txm.lggr.Infow("transaction broadcast", "txhash", hash)
						}
					}
					wg.Done()
				}(sender, txs)
			}
			wg.Wait()
			tick = time.After(utils.WithJitter(txm.cfg.TxSendFrequency()) - time.Since(start))
		case <-txm.stop:
			return
		}
	}
}

func (txm *starktxm) broadcastBatch(ctx context.Context, privKey, sender string, txs []types.Transaction) (txhash string, err error) {
	// create new account
	account, err := caigo.NewAccount(txm.curve, privKey, sender, txm.client)
	if err != nil {
		return txhash, errors.Errorf("failed to create new account: %s", err)
	}

	// get fee for txm
	fee, err := account.EstimateFee(ctx, txs)
	if err != nil {
		return txhash, errors.Errorf("failed to estimate fee: %s", err)
	}

	// TODO: investigate if nonce management is needed (nonce is requested queried by the sdk for now)
	// transmit txs
	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout()*time.Second)
	defer execCancel()
	res, err := account.Execute(execCtx, types.StrToFelt(strconv.Itoa(int(fee.Amount))), txs)
	if err != nil {
		return txhash, errors.Errorf("failed to invoke tx: %s", err)
	}

	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	return res.TransactionHash, nil
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
