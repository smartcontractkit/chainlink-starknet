package txm

import (
	"context"
	"sync"

	"github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
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
	done    sync.WaitGroup
	stop    chan struct{}
	queue   chan types.Transaction
}

// TODO: pass in
// - method to fetch client
// - signer for signing tx
// - logger
func New() StarkTXM {
	return &starktxm{
		queue: make(chan types.Transaction, MaxQueueLen),
		stop:  make(chan struct{}),
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

	// TODO: func not available without importing core
	// ctx, cancel := utils.ContextFromChan(txm.stop)
	// defer cancel()

	for {
		select {
		case <-txm.queue:
			// process + broadcast transactions
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
