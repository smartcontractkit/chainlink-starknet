package ocr2

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type Tracker interface {
	Start() error
	Close() error
	poll()

	updateConfig(context.Context) error
	updateTransmission(ctx context.Context) error
}

var _ Tracker = (*ContractCache)(nil)
var _ median.MedianContract = (*ContractCache)(nil)
var _ types.ContractConfigTracker = (*ContractCache)(nil)

type ContractCache struct {
	contractConfig ContractConfig
	ccLock         sync.RWMutex
	ccTime         time.Time

	transmissionDetails TransmissionDetails
	tdLock              sync.RWMutex
	tdTime              time.Time

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
	c.contractConfig = newConfig
	c.ccLock.Unlock()

	return nil
}

func (c *ContractCache) updateTransmission(ctx context.Context) error {
	// todo: update transmission details with the reader
	// todo: assert reading was successful, return error otherwise
	transmissionDetails := TransmissionDetails{}

	c.tdLock.Lock()
	c.transmissionDetails = transmissionDetails
	c.tdLock.Unlock()

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

			if err := c.updateTransmission(ctx); err != nil {
				c.lggr.Errorf("Failed to update transmission: %v", err)
			}
			cancel()

			// todo: adjust tick with values from config
			tick = time.After(utils.WithJitter(0))
		}
	}
}

func (c *ContractCache) LatestTransmissionDetails(
	ctx context.Context,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	round uint8,
	latestAnswer *big.Int,
	latestTimestamp time.Time,
	err error,
) {
	c.tdLock.RLock()
	configDigest = c.transmissionDetails.digest
	epoch = c.transmissionDetails.epoch
	round = c.transmissionDetails.round
	latestAnswer = c.transmissionDetails.latestAnswer
	latestTimestamp = c.transmissionDetails.latestTimestamp
	c.tdLock.RUnlock()
	return
}

func (c *ContractCache) LatestRoundRequested(
	ctx context.Context,
	lookback time.Duration,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	round uint8,
	err error,
) {
	// todo: implement
	return types.ConfigDigest{}, 0, 0, err
}

func (c *ContractCache) Notify() <-chan struct{} {
	return nil
}

func (c *ContractCache) LatestConfigDetails(ctx context.Context) (changedInBlock uint64, configDigest types.ConfigDigest, err error) {
	c.ccLock.RLock()
	changedInBlock = c.contractConfig.configBlock
	configDigest = c.contractConfig.config.ConfigDigest
	c.ccLock.RUnlock()
	return
}

func (c *ContractCache) LatestConfig(ctx context.Context, changedInBlock uint64) (config types.ContractConfig, err error) {
	c.ccLock.RLock()
	config = c.contractConfig.config
	c.ccLock.RUnlock()
	return
}

func (c *ContractCache) LatestBlockHeight(ctx context.Context) (blockHeight uint64, err error) {
	// todo: implement
	return 0, nil
}
