package ocr2

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type Tracker interface {
	Start() error
	Close() error
	poll()
}

var _ Tracker = (*contractCache)(nil)
var _ types.ContractConfigTracker = (*contractCache)(nil)

type contractCache struct {
	contractConfig  ContractConfig
	ccLock          sync.RWMutex
	ccLastCheckedAt time.Time

	stop, done chan struct{}

	reader *contractReader
	lggr   logger.Logger
}

func NewContractCache(reader *contractReader, lggr logger.Logger) *contractCache {
	return &contractCache{
		reader: reader,
		lggr:   lggr,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (c *contractCache) updateConfig(ctx context.Context) error {
	configBlock, configDigest, err := c.reader.LatestConfigDetails(ctx)
	if err != nil {
		return errors.Wrap(err, "fetch latest config details")
	}

	isSameConfig := c.contractConfig.ConfigBlock == configBlock && c.contractConfig.Config.ConfigDigest == configDigest
	if isSameConfig {
		now := time.Now()
		c.ccLock.Lock()
		defer c.ccLock.Unlock()

		if c.ccLastCheckedAt.Before(now) {
			c.ccLastCheckedAt = now
		}

		return nil
	}

	newConfig, err := c.reader.LatestConfig(ctx, configBlock)
	if err != nil {
		return errors.Wrap(err, "fetch latest config")
	}

	c.ccLock.Lock()
	defer c.ccLock.Unlock()
	c.contractConfig = ContractConfig{
		Config:      newConfig,
		ConfigBlock: configBlock,
	}
	c.ccLastCheckedAt = time.Now()

	return nil
}

func (c *contractCache) Start() error {
	ctx, cancel := utils.ContextFromChan(c.stop)
	defer cancel()
	if err := c.updateConfig(ctx); err != nil {
		c.lggr.Warnf("failed to populate initial config: %v", err)
	}
	go c.poll()
	return nil
}

func (c *contractCache) Close() error {
	close(c.stop)
	return nil
}

func (c *contractCache) poll() {
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

func (c *contractCache) Notify() <-chan struct{} {
	return nil
}

func (c *contractCache) LatestConfigDetails(ctx context.Context) (changedInBlock uint64, configDigest types.ConfigDigest, err error) {
	c.ccLock.RLock()
	defer c.ccLock.RUnlock()
	changedInBlock = c.contractConfig.ConfigBlock
	configDigest = c.contractConfig.Config.ConfigDigest
	return
}

func (c *contractCache) LatestConfig(ctx context.Context, changedInBlock uint64) (config types.ContractConfig, err error) {
	c.ccLock.RLock()
	defer c.ccLock.RUnlock()
	config = c.contractConfig.Config
	return
}

func (c *contractCache) LatestBlockHeight(ctx context.Context) (blockHeight uint64, err error) {
	// todo: implement
	return 0, nil
}
