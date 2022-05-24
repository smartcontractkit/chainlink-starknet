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

var _ Tracker = (*TransmissionsCache)(nil)
var _ median.MedianContract = (*TransmissionsCache)(nil)

type TransmissionsCache struct {
	transmissionDetails TransmissionDetails
	tdLock              sync.RWMutex
	tdTime              time.Time

	stop, done chan struct{}

	reader *ContractReader
	lggr   logger.Logger
}

func NewTransmissionsCache(reader *ContractReader, lggr logger.Logger) *TransmissionsCache {
	return &TransmissionsCache{
		reader: reader,
		lggr:   lggr,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (c *TransmissionsCache) updateTransmission(ctx context.Context) error {
	// todo: update transmission details with the reader
	// todo: assert reading was successful, return error otherwise
	transmissionDetails := TransmissionDetails{}

	c.tdLock.Lock()
	c.transmissionDetails = transmissionDetails
	c.tdLock.Unlock()

	return nil
}

func (c *TransmissionsCache) Start() error {
	ctx, cancel := utils.ContextFromChan(c.stop)
	defer cancel()
	if err := c.updateTransmission(ctx); err != nil {
		c.lggr.Warnf("failed to populate initial transmission details: %v", err)
	}
	go c.poll()
	return nil
}

func (c *TransmissionsCache) Close() error {
	close(c.stop)
	return nil
}

func (c *TransmissionsCache) poll() {
	defer close(c.done)
	tick := time.After(0)
	for {
		select {
		case <-c.stop:
			return
		case <-tick:
			ctx, cancel := utils.ContextFromChan(c.stop)

			if err := c.updateTransmission(ctx); err != nil {
				c.lggr.Errorf("Failed to update transmission: %v", err)
			}
			cancel()

			// todo: adjust tick with values from config
			tick = time.After(utils.WithJitter(0))
		}
	}
}

func (c *TransmissionsCache) LatestTransmissionDetails(
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

func (c *TransmissionsCache) LatestRoundRequested(
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
