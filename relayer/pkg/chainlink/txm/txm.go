package txm

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/caigo"
	caigotypes "github.com/smartcontractkit/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
)

const (
	MaxQueueLen = 1000
)

type TxManager interface {
	Enqueue(caigotypes.Hash, caigotypes.Hash, caigotypes.FunctionCall) error
}

type Tx struct {
	senderAddress  caigotypes.Hash
	accountAddress caigotypes.Hash
	call           caigotypes.FunctionCall
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
	queue   chan Tx
	ks      keys.Keystore
	cfg     Config

	// TODO: use lazy loaded client
	client    *starknet.Client
	getClient func() (*starknet.Client, error)
}

func New(lggr logger.Logger, keystore keys.Keystore, cfg Config, getClient func() (*starknet.Client, error)) (StarkTXM, error) {
	return &starktxm{
		lggr:      logger.Named(lggr, "StarknetTxm"),
		queue:     make(chan Tx, MaxQueueLen),
		stop:      make(chan struct{}),
		getClient: getClient,
		ks:        keystore,
		cfg:       cfg,
	}, nil
}

func (txm *starktxm) Name() string {
	return txm.lggr.Name()
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

			// fetch client if needed (before processing txs to preserve queue)
			if txm.client == nil {
				var err error
				txm.client, err = txm.getClient()
				if err != nil {
					txm.lggr.Errorw("unable to fetch client", "error", err)
					tick = time.After(utils.WithJitter(txm.cfg.TxSendFrequency()) - time.Since(start)) // reset tick
					txm.client = nil                                                                   // reset
					continue
				}
			}

			// calculate total txs to process
			txLen := len(txm.queue)
			if txLen > txm.cfg.TxMaxBatchSize() {
				txLen = txm.cfg.TxMaxBatchSize()
			}

			type senderAccountPair struct {
				senderAddress  caigotypes.Hash
				accountAddress caigotypes.Hash
			}

			// split the transactions by sender and account address for batching.
			// this avoids assuming that there is always a 1:1 mapping from sender addresses
			// onto account addresses.
			txsByAccount := map[senderAccountPair][]caigotypes.FunctionCall{}
			for i := 0; i < txLen; i++ {
				tx := <-txm.queue
				key := senderAccountPair{
					senderAddress:  tx.senderAddress,
					accountAddress: tx.accountAddress,
				}
				txsByAccount[key] = append(txsByAccount[key], tx.call)
			}
			txm.lggr.Infow("creating batch", "totalTxCount", txLen, "batchCount", len(txsByAccount))

			// async process of tx batches
			var wg sync.WaitGroup
			wg.Add(len(txsByAccount))
			for key, txs := range txsByAccount {
				txm.lggr.Debugw("batch details", "accountAddress", key.accountAddress, "senderAddress", key.senderAddress, "txs", txs)
				go func(senderAddress caigotypes.Hash, accountAddress caigotypes.Hash, txs []caigotypes.FunctionCall) {
					// fetch key matching sender address
					key, err := txm.ks.Get(senderAddress.String())
					if err != nil {
						txm.lggr.Errorw("failed to retrieve key", "id", senderAddress, "error", err)
					} else {
						// parse key to expected format
						privKeyBytes := key.Raw()
						privKey := caigotypes.BigToHex(caigotypes.BytesToBig(privKeyBytes))

						// broadcast batch based on account address
						hash, err := txm.broadcastBatch(ctx, privKey, accountAddress, txs)
						if err != nil {
							txm.lggr.Errorw("transaction failed to broadcast", "error", err, "account", accountAddress, "batchTx", txs)
						} else {
							txm.lggr.Infow("transaction broadcast", "txhash", hash)
						}
					}
					wg.Done()
				}(key.senderAddress, key.accountAddress, txs)
			}
			wg.Wait()
			tick = time.After(utils.WithJitter(txm.cfg.TxSendFrequency()) - time.Since(start))
		case <-txm.stop:
			return
		}
	}
}

const FEE_MARGIN uint64 = 115

func (txm *starktxm) broadcastBatch(ctx context.Context, privKey string, accountAddress caigotypes.Hash, txs []caigotypes.FunctionCall) (txhash string, err error) {
	// create new account
	account, err := caigo.NewGatewayAccount(privKey, accountAddress.String(), txm.client.Gw, caigo.AccountVersion1)
	if err != nil {
		return txhash, errors.Errorf("failed to create new account: %s", err)
	}

	// get fee for txm
	feeEstimate, err := account.EstimateFee(ctx, txs, caigotypes.ExecuteDetails{})
	if err != nil {
		return txhash, errors.Errorf("failed to estimate fee: %s", err)
	}

	fee, _ := big.NewInt(0).SetString(string(feeEstimate.OverallFee), 0)
	expandedFee := big.NewInt(0).Mul(fee, big.NewInt(int64(FEE_MARGIN)))
	max := big.NewInt(0).Div(expandedFee, big.NewInt(100))
	details := caigotypes.ExecuteDetails{
		MaxFee: max,
	}

	// TODO: investigate if nonce management is needed (nonce is requested queried by the sdk for now)
	// transmit txs
	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout()*time.Second)
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

func (txm *starktxm) Enqueue(senderAddress caigotypes.Hash, accountAddress caigotypes.Hash, tx caigotypes.FunctionCall) error {
	select {
	case txm.queue <- Tx{senderAddress: senderAddress, accountAddress: accountAddress, call: tx}:
	default:
		return errors.Errorf("failed to enqueue transaction: %+v", tx)
	}

	return nil
}
