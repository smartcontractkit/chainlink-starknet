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
)

const (
	MaxQueueLen    = 1000
	DefaultTimeout = 1 //s TODO: move to config
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

	// TODO: use lazy loaded client
	client    *gateway.GatewayProvider
	getClient func(...gateway.Option) *gateway.GatewayProvider
}

// TODO: pass in
// - method to fetch client
// - signer for signing tx
// - logger
func New(lggr logger.Logger) StarkTXM {
	return &starktxm{
		lggr:      lggr,
		queue:     make(chan types.Transaction, MaxQueueLen),
		stop:      make(chan struct{}),
		getClient: gateway.NewProvider,
	}
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

	curve, err := caigo.SC(caigo.WithConstants())
	if err != nil {
		txm.lggr.Errorw("failed to build curve", "error", err)
	}

	// TODO: func not available without importing core
	// ctx, cancel := utils.ContextFromChan(txm.stop)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case tx := <-txm.queue: // process + broadcast transactions

			// fetch client
			if txm.client == nil {
				txm.client = txm.getClient()
			}

			// create new account
			privKey := "0"
			address := "PLACEHOLDER"
			account, err := caigo.NewAccount(&curve, privKey, address, txm.client)
			if err != nil {
				txm.lggr.Errorw("failed to create new account", "error", err)
				continue
			}

			// get fee for tx
			// TODO: move ctx construction into Reader/Writer client
			feeCtx, feeCancel := context.WithTimeout(ctx, DefaultTimeout*time.Second)
			defer feeCancel()
			fee, err := account.EstimateFee(feeCtx, []types.Transaction{tx})
			if err != nil {
				txm.lggr.Errorw("failed to estimate fee", "error", err, "transaction", tx)
				continue // exit loop
			}

			// TODO: move ctx construction into Reader/Writer client
			// TODO: nonce management => batching?
			// transmit txs
			execCtx, execCancel := context.WithTimeout(ctx, DefaultTimeout*time.Second)
			defer execCancel()
			res, err := account.Execute(execCtx, types.StrToFelt(strconv.Itoa(int(fee.Amount))), []types.Transaction{tx})
			if err != nil {
				txm.lggr.Errorw("failed to invoke tx", "error", err, "transaction", tx)
				continue
			}

			// handle nil pointer
			if res == nil {
				txm.lggr.Warnw("invoke response is nil", "tx", tx)
				continue
			}

			txm.lggr.Infow("transaction broadcast", "txhash", res.TransactionHash)
		case <-txm.stop:
			return
		}
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
	// TODO
	return nil
}

func (txm *starktxm) Ready() error {
	// TODO
	return nil
}

func (txm *starktxm) Enqueue(tx types.Transaction) error {
	select {
	case txm.queue <- tx:
	default:
		return errors.New("failed to enqueue transaction")
	}

	return nil
}
