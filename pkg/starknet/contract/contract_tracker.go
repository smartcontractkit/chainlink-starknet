package contract

import (
	"context"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/client"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/config"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/logger"

	"github.com/smartcontractkit/chainlink/core/services"
	"github.com/smartcontractkit/chainlink/core/utils"
)

type Tracker interface {
	services.ServiceCtx

	ReadState() (State, error)
}

var _ Tracker = (*ContractTracker)(nil)

type ContractTracker struct {
	state     State
	stateLock *sync.RWMutex
	stateTime time.Time

	reader client.Reader
	cfg    config.Config
	lggr   logger.Logger

	utils.StartStopOnce
}

func NewTracker(spec OCR2Spec, cfg config.Config, reader client.Reader, lggr logger.Logger) *ContractTracker {
	return &ContractTracker{
		reader:    reader,
		cfg:       cfg,
		lggr:      lggr,
		stateLock: &sync.RWMutex{},
	}
}

func (c *ContractTracker) ReadState() (State, error) {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()

	// todo: assert state is not stale (if necessary)
	return c.state, nil
}

func (c *ContractTracker) Start(ctx context.Context) error {
	// todo: implement
	return nil
}

func (c *ContractTracker) Close() error {
	// todo: implement
	return nil
}
