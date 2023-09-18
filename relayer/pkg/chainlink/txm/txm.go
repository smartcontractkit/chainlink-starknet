package txm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	starknetgo "github.com/NethermindEth/starknet.go"
	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	starknettypes "github.com/NethermindEth/starknet.go/types"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const (
	MaxQueueLen = 1000
)

type TxManager interface {
	Enqueue(senderAddress *felt.Felt, accountAddress *felt.Felt, txFn starknettypes.FunctionCall) error
	InflightCount() (int, int)
}

type Tx struct {
	senderAddress  *felt.Felt
	accountAddress *felt.Felt
	call           starknettypes.FunctionCall
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
	ks      KeystoreAdapter
	cfg     Config
	nonce   NonceManager

	client  *utils.LazyLoad[*starknet.Client]
	txStore *ChainTxStore
}

func New(lggr logger.Logger, keystore loop.Keystore, cfg Config, getClient func() (*starknet.Client, error)) (StarkTXM, error) {
	txm := &starktxm{
		lggr:    logger.Named(lggr, "StarknetTxm"),
		queue:   make(chan Tx, MaxQueueLen),
		stop:    make(chan struct{}),
		client:  utils.NewLazyLoad(getClient),
		ks:      NewKeystoreAdapter(keystore),
		cfg:     cfg,
		txStore: NewChainTxStore(),
	}
	txm.nonce = NewNonceManager(txm.lggr)

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

			// broadcast tx serially - wait until accepted by mempool before processing next
			hash, err := txm.broadcast(ctx, tx.senderAddress, tx.accountAddress, tx.call)
			if err != nil {
				txm.lggr.Errorw("transaction failed to broadcast", "error", err, "tx", tx.call)
			} else {
				txm.lggr.Infow("transaction broadcast", "txhash", hash)
			}
		}
	}
}

const FEE_MARGIN uint64 = 115

func (txm *starktxm) broadcast(ctx context.Context, senderAddress *felt.Felt, accountAddress *felt.Felt, tx starknettypes.FunctionCall) (txhash string, err error) {
	txs := []starknettypes.FunctionCall{tx}
	client, err := txm.client.Get()
	if err != nil {
		txm.client.Reset()
		return txhash, fmt.Errorf("broadcast: failed to fetch client: %+w", err)
	}
	// create new account
	account, err := starknetgo.NewRPCAccount(senderAddress, accountAddress, txm.ks, client.Provider, starknetgo.AccountVersion1)
	if err != nil {
		return txhash, fmt.Errorf("failed to create new account: %+w", err)
	}

	chainID, err := client.Provider.ChainID(ctx)
	if err != nil {
		return txhash, fmt.Errorf("failed to get chainID: %+w", err)
	}

	nonce, err := txm.nonce.NextSequence(accountAddress, chainID)
	if err != nil {
		return txhash, fmt.Errorf("failed to get nonce: %+w", err)
	}

	// get fee for txm
	// optional - pass nonce to fee estimate (if nonce gets ahead, estimate may fail)
	// can we estimate fee without calling estimate - tbd with 1.0
	feeEstimate, err := account.EstimateFee(ctx, txs, starknettypes.ExecuteDetails{})
	if err != nil {
		return txhash, fmt.Errorf("failed to estimate fee: %+w", err)
	}

	fee, _ := big.NewInt(0).SetString(string(feeEstimate.OverallFee), 0)
	expandedFee := big.NewInt(0).Mul(fee, big.NewInt(int64(FEE_MARGIN)))
	max := big.NewInt(0).Div(expandedFee, big.NewInt(100))
	details := starknettypes.ExecuteDetails{
		MaxFee: max,
		Nonce:  nonce,
	}

	// transmit txs
	execCtx, execCancel := context.WithTimeout(ctx, txm.cfg.TxTimeout())
	defer execCancel()
	res, err := account.Execute(execCtx, txs, details)
	if err != nil {
		// TODO: handle initial broadcast errors - what kind of errors occur?
		return txhash, fmt.Errorf("failed to invoke tx: %+w", err)
	}

	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	// update nonce if transaction is successful
	hash := res.TransactionHash.String()
	err = errors.Join(
		txm.nonce.IncrementNextSequence(accountAddress, chainID, nonce),
		txm.txStore.Save(accountAddress, nonce, hash),
	)
	return hash, err
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

			hashes := txm.txStore.GetAllUnconfirmed()
			for addr := range hashes {
				for i := range hashes[addr] {
					hash := hashes[addr][i]
					f, err := starknetutils.HexToFelt(hash)
					if err != nil {
						txm.lggr.Errorw("invalid felt value", "hash", hash)
						continue
					}
					// according to v0.4 rpc changes, if no error, then tx is confirmed
					response, err := client.Provider.TransactionReceipt(ctx, f)
					if err != nil {
						txm.lggr.Errorw("failed to fetch transaction status", "hash", hash, "error", err)
						continue
					}
					receipt, ok := response.(starknetrpc.InvokeTransactionReceipt)
					if !ok {
						txm.lggr.Errorw("wrong receipt type", "type", reflect.TypeOf(response))
						continue
					}

					status := receipt.FinalityStatus

					txm.lggr.Debugw(fmt.Sprintf("tx confirmed: %s", status), "hash", hash, "finality status", status)
					if err := txm.txStore.Confirm(addr, hash); err != nil {
						txm.lggr.Errorw("failed to confirm tx in TxStore", "hash", hash, "sender", addr, "error", err)
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

func (txm *starktxm) Enqueue(senderAddress, accountAddress *felt.Felt, tx starknettypes.FunctionCall) error {
	// validate key exists for sender
	// use the embedded Loopp Keystore to do this; the spec and design
	// encourage passing nil data to the loop.Keystore.Sign as way to test
	// existence of a key
	if _, err := txm.ks.Loopp().Sign(context.Background(), senderAddress.String(), nil); err != nil {
		return err
	}

	client, err := txm.client.Get()
	if err != nil {
		txm.client.Reset()
		return fmt.Errorf("broadcast: failed to fetch client: %+w", err)
	}

	chainID, err := client.Provider.ChainID(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get chainID: %+w", err)
	}

	// register account for nonce manager
	if err := txm.nonce.Register(context.TODO(), accountAddress, chainID, client); err != nil {
		return err
	}

	select {
	case txm.queue <- Tx{senderAddress: senderAddress, accountAddress: accountAddress, call: tx}:
	default:
		return fmt.Errorf("failed to enqueue transaction: %+v", tx)
	}

	return nil
}

func (txm *starktxm) InflightCount() (queue int, unconfirmed int) {
	list := maps.Values(txm.txStore.GetAllInflightCount())
	for _, count := range list {
		unconfirmed += count
	}
	return len(txm.queue), unconfirmed
}
