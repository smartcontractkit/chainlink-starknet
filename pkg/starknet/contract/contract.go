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

	ReadCCFromCache() (ContractConfig, error)
	updateConfig(context.Context) error
}

var _ Tracker = (*ContractTracker)(nil)

type ContractTracker struct {
	contractConfig ContractConfig
	ccLock         *sync.RWMutex
	ccTime         time.Time

	reader client.Reader
	cfg    config.Config
	lggr   logger.Logger

	utils.StartStopOnce
}

func NewTracker(spec OCR2Spec, cfg config.Config, reader client.Reader, lggr logger.Logger) *ContractTracker {
	return &ContractTracker{
		reader: reader,
		cfg:    cfg,
		lggr:   lggr,
		ccLock: &sync.RWMutex{},
	}
}

func (c *ContractTracker) ReadCCFromCache() (ContractConfig, error) {
	c.ccLock.RLock()
	defer c.ccLock.RUnlock()

	// todo: assert cache is not stale (if necessary)
	return c.contractConfig, nil
}

func (c *ContractTracker) updateConfig(ctx context.Context) error {
	// todo: read latest config through the reader
	// todo: assert reading was successful, return error otherwise
	newConfig := ContractConfig{}

	c.ccLock.Lock()
	c.contractConfig = newConfig
	c.ccLock.Unlock()

	return nil
}

func (c *ContractTracker) Start(ctx context.Context) error {
	// todo: implement
	return nil
}

func (c *ContractTracker) Close() error {
	// todo: implement
	return nil
}
