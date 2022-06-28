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
	MaxQueueLen     = 1000
	DefaultTimeout  = 1 //s TODO: move to config
	TxSendFrequency = 1 //s TODO: move to config
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

	// TODO: use lazy loaded client
	client    *gateway.GatewayProvider
	getClient func(...gateway.Option) *gateway.GatewayProvider
}

// TODO: pass in
// - method to fetch client
// - signer for signing tx
// - logger
func New(lggr logger.Logger) (StarkTXM, error) {
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
			tx := <-txm.queue // process + broadcast transactions

			txs := []types.Transaction{tx}

			// fetch client
			if txm.client == nil {
				txm.client = txm.getClient()
			}

			// fetch key matching tx.SenderAddress
			privKey := "0"

			hash, err := txm.broadcastBatch(ctx, privKey, tx.SenderAddress, txs)
			if err != nil {
				txm.lggr.Errorw("transaction failed to broadcast", "error", err, "batchTx", txs)
				continue
			}
			txm.lggr.Infow("transaction broadcast", "txhash", hash)

			tick = time.After(utils.WithJitter(TxSendFrequency) - time.Since(start))
		case <-txm.stop:
			return
		}
	}
}

func (txm *starktxm) broadcastBatch(ctx context.Context, privKey, sender string, txs []types.Transaction) (txhash string, err error) {
	// create new account
	account, err := caigo.NewAccount(txm.curve, privKey, sender, txm.client)
	if err != nil {
		return txhash, errors.Errorf("failed to create new account:", err)
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
	// TODO: investigate if nonce management is needed
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
