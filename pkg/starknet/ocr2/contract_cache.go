package ocr2

import (
	"context"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type Tracker interface {
	Start() error
	Close() error
	poll()
}

var _ Tracker = (*ContractCache)(nil)
var _ types.ContractConfigTracker = (*ContractCache)(nil)

type ContractCache struct {
	contractConfig ContractConfig
	ccLock         sync.RWMutex
	ccTime         time.Time

	stop, done chan struct{}

	reader *ContractReader
	lggr   logger.Logger
}

func NewContractCache(reader *ContractReader, lggr logger.Logger) *ContractCache {
	return &ContractCache{
		reader: reader,
		lggr:   lggr,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (c *ContractCache) updateConfig(ctx context.Context) error {
	// todo: update config with the reader
	// todo: assert reading was successful, return error otherwise
	newConfig := ContractConfig{}

	c.ccLock.Lock()
	defer c.ccLock.Unlock()
	c.contractConfig = newConfig

	return nil
}

func (c *ContractCache) Start() error {
	ctx, cancel := utils.ContextFromChan(c.stop)
	defer cancel()
	if err := c.updateConfig(ctx); err != nil {
		c.lggr.Warnf("failed to populate initial config: %v", err)
	}
	go c.poll()
	return nil
}

func (c *ContractCache) Close() error {
	close(c.stop)
	return nil
}

func (c *ContractCache) poll() {
	defer close(c.done)
	tick := time.After(0)
	for {
		select {
		case <-c.stop:
			return
		case <-tick:
			ctx, cancel := utils.ContextFromChan(c.stop)

			if err := c.updateConfig(ctx); err != nil {
				c.lggr.Errorf("Failed to update config: %v", err)
			}
			cancel()

			// todo: adjust tick with values from config
			tick = time.After(utils.WithJitter(0))
		}
	}
}

func (c *ContractCache) Notify() <-chan struct{} {
	return nil
}

func (c *ContractCache) LatestConfigDetails(ctx context.Context) (changedInBlock uint64, configDigest types.ConfigDigest, err error) {
	c.ccLock.RLock()
	defer c.ccLock.RUnlock()
	changedInBlock = c.contractConfig.configBlock
	configDigest = c.contractConfig.config.ConfigDigest
	return
}

func (c *ContractCache) LatestConfig(ctx context.Context, changedInBlock uint64) (config types.ContractConfig, err error) {
	c.ccLock.RLock()
	defer c.ccLock.RUnlock()
	config = c.contractConfig.config
	return
}

func (c *ContractCache) LatestBlockHeight(ctx context.Context) (blockHeight uint64, err error) {
	// todo: implement
	return 0, nil
}
