package txm

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/keys"
)

const (
	MaxQueueLen     = 1000
	DefaultTimeout  = 1   //s TODO: move to config
	TxSendFrequency = 1   //s TODO: move to config
	MaxTxsPerBatch  = 100 // TODO: move to config
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

	// TODO: use lazy loaded client
	client    *gateway.GatewayProvider
	getClient func(...gateway.Option) *gateway.GatewayProvider
}

// TODO: pass in
// - method to fetch client
func New(lggr logger.Logger, keystore keys.Keystore) (StarkTXM, error) {
	curve, err := caigo.SC(caigo.WithConstants())
	if err != nil {
		return nil, errors.Errorf("failed to build curve: %s", err)
	}
	return &starktxm{
		lggr:      lggr,
		queue:     make(chan types.Transaction, MaxQueueLen),
		stop:      make(chan struct{}),
		getClient: gateway.NewProvider,
		curve:     &curve,
		ks:        keystore,
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
			if txLen > MaxTxsPerBatch {
				txLen = MaxTxsPerBatch
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
				txm.client = txm.getClient() // TODO: chains + nodes config for proper endpoint
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

			tick = time.After(utils.WithJitter(TxSendFrequency*time.Second) - time.Since(start))
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
	// TODO: move ctx construction into Reader/Writer client
	feeCtx, feeCancel := context.WithTimeout(ctx, DefaultTimeout*time.Second)
	defer feeCancel()
	fee, err := account.EstimateFee(feeCtx, txs)
	if err != nil {
		return txhash, errors.Errorf("failed to estimate fee: %s", err)
	}

	// TODO: move ctx construction into Reader/Writer client
	// TODO: investigate if nonce management is needed (nonce is requested queried by the sdk for now)
	// transmit txs
	execCtx, execCancel := context.WithTimeout(ctx, DefaultTimeout*time.Second)
	defer execCancel()
	res, err := account.Execute(execCtx, types.StrToFelt(strconv.Itoa(int(fee.Amount))), txs)
	if err != nil {
		return txhash, errors.Errorf("failed to invoke tx: %s", err)
	}

	// handle nil pointer
	if res == nil {
		return txhash, errors.Errorf("execute response and error are nil")
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
	// TODO
	return txm.starter.Healthy()
}

func (txm *starktxm) Ready() error {
	return txm.starter.Ready()
}

func (txm *starktxm) Enqueue(tx types.Transaction) error {
	select {
	case txm.queue <- tx:
	default:
		return errors.New("failed to enqueue transaction")
	}

	return nil
}
