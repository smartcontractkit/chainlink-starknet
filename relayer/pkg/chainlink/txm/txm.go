package txm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/smartcontractkit/caigo"
	"github.com/smartcontractkit/caigo/gateway"
	caigotypes "github.com/smartcontractkit/caigo/types"
	"golang.org/x/exp/maps"

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
	InflightCount() int
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
	nonce   keys.NonceManager

	client  *utils.LazyLoad[*starknet.Client]
	txStore ChainTxStore
}

func New(lggr logger.Logger, keystore keys.Keystore, cfg Config, getClient func() (*starknet.Client, error)) (StarkTXM, error) {
	txm := &starktxm{
		lggr:    logger.Named(lggr, "StarknetTxm"),
		queue:   make(chan Tx, MaxQueueLen),
		stop:    make(chan struct{}),
		client:  utils.NewLazyLoad(getClient),
		ks:      keystore,
		cfg:     cfg,
		txStore: ChainTxStore{},
	}
	client, err := txm.client.Get()
	if err != nil {
		return txm, err
	}
	txm.nonce = keys.NewNonceManager(txm.lggr, client, txm.ks)

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

			// fetch key matching sender address
			key, err := txm.ks.Get(tx.senderAddress.String())
			if err != nil {
				txm.lggr.Errorw("failed to retrieve key", "id", tx.senderAddress, "error", err)
				continue
			}

			// parse key to expected format
			privKeyBytes := key.Raw()
			privKey := caigotypes.BigToHex(caigotypes.BytesToBig(privKeyBytes))

			// broadcast tx serially - wait until accepted by mempool before processing next
			hash, err := txm.broadcast(ctx, privKey, tx.accountAddress, tx.call)
			if err != nil {
				txm.lggr.Errorw("transaction failed to broadcast", "error", err, "tx", tx.call)
			} else {
				txm.lggr.Infow("transaction broadcast", "txhash", hash)
			}
		}
	}
}

const FEE_MARGIN uint64 = 115

func (txm *starktxm) broadcast(ctx context.Context, privKey string, accountAddress caigotypes.Hash, tx caigotypes.FunctionCall) (txhash string, err error) {
	txs := []caigotypes.FunctionCall{tx}
	client, err := txm.client.Get()
	if err != nil {
		return txhash, fmt.Errorf("broadcast: failed to fetch client: %+w", err)
	}
	// create new account
	account, err := caigo.NewGatewayAccount(privKey, accountAddress.String(), client.Gw, caigo.AccountVersion1)
	if err != nil {
		return txhash, fmt.Errorf("failed to create new account: %+w", err)
	}

	// nonce, err := txm.nonce.NextSequence(accountAddress, client.Gw.ChainId)
	// if err != nil {
	// 	return txhash, fmt.Errorf("failed to get nonce: %+w", err)
	// }

	// // get fee for txm
	// // optional - pass nonce to fee estimate (if nonce gets ahead, estimate may fail)
	// // can we estimate fee without calling estimate - tbd with 1.0
	// feeEstimate, err := account.EstimateFee(ctx, txs, caigotypes.ExecuteDetails{})
	// if err != nil {
	// 	return txhash, fmt.Errorf("failed to estimate fee: %+w", err)
	// }

	// fee, _ := big.NewInt(0).SetString(string(feeEstimate.OverallFee), 0)
	// expandedFee := big.NewInt(0).Mul(fee, big.NewInt(int64(FEE_MARGIN)))
	// max := big.NewInt(0).Div(expandedFee, big.NewInt(100))
	// details := caigotypes.ExecuteDetails{
	// 	MaxFee: max,
	// 	Nonce:  nonce,
	// }
	details := caigotypes.ExecuteDetails{}

	// transmit txs
	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout()*10*time.Second) // TODO: for some reason the TxTimeout isn't correct and immediately times out
	defer execCancel()
	res, err := account.Execute(execCtx, txs, details)
	if err != nil {
		// TODO: handle nonce errors - what kind of errors occur?
		return txhash, fmt.Errorf("failed to invoke tx: %+w", err)
	}

	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	// update nonce if transaction is successful
	// err = errors.Join(
	// 	txm.nonce.IncrementNextSequence(accountAddress, client.Gw.ChainId, nonce),
	// 	txm.txStore.Save(accountAddress, nonce, res.TransactionHash),
	// )
	return res.TransactionHash, err
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

			hashes := txm.txStore.GetUnconfirmed()
			for addr := range hashes {
				for i := range hashes[addr] {
					hash := hashes[addr][i]
					status, err := client.Gw.TransactionStatus(ctx, gateway.TransactionStatusOptions{
						TransactionHash: hashes[addr][i],
					})
					if err != nil {
						txm.lggr.Errorw("failed to fetch transaction status", "hash", hash, "error", err)
					}

					if status.TxStatus == "ACCEPTED_ON_L1" || status.TxStatus == "ACCEPTED_ON_L2" || status.TxStatus == "REJECTED" {
						txm.lggr.Debugw("tx confirmed", "hash", hash, "status", status)
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

		tick = time.After(utils.WithJitter(txm.cfg.ConfirmationPoll() - time.Since(start)))
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

func (txm *starktxm) Enqueue(senderAddress, accountAddress caigotypes.Hash, tx caigotypes.FunctionCall) error {
	// validate key exists for sender
	if _, err := txm.ks.Get(senderAddress.String()); err != nil {
		return err
	}

	client, err := txm.client.Get()
	if err != nil {
		return fmt.Errorf("broadcast: failed to fetch client: %+w", err)
	}

	// register account for nonce manager
	if err := txm.nonce.Register(context.TODO(), accountAddress, client.Gw.ChainId); err != nil {
		return err
	}

	select {
	case txm.queue <- Tx{senderAddress: senderAddress, accountAddress: accountAddress, call: tx}:
	default:
		return fmt.Errorf("failed to enqueue transaction: %+v", tx)
	}

	return nil
}

func (txm *starktxm) InflightCount() (count int) {
	list := maps.Values(txm.txStore.GetUnconfirmed())
	for i := range list {
		count += len(list[i])
	}
	return
}
