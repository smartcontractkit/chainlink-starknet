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

var _ Tracker = (*transmissionsCache)(nil)
var _ median.MedianContract = (*transmissionsCache)(nil)

type transmissionsCache struct {
	transmissionDetails TransmissionDetails
	tdLock              sync.RWMutex
	tdTime              time.Time

	stop, done chan struct{}

	reader *contractReader
	lggr   logger.Logger
}

func NewTransmissionsCache(reader *contractReader, lggr logger.Logger) *transmissionsCache {
	return &transmissionsCache{
		reader: reader,
		lggr:   lggr,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (c *transmissionsCache) updateTransmission(ctx context.Context) error {
	// todo: update transmission details with the reader
	// todo: assert reading was successful, return error otherwise
	transmissionDetails := TransmissionDetails{}

	c.tdLock.Lock()
	defer c.tdLock.Unlock()
	c.transmissionDetails = transmissionDetails

	return nil
}

func (c *transmissionsCache) Start() error {
	ctx, cancel := utils.ContextFromChan(c.stop)
	defer cancel()
	if err := c.updateTransmission(ctx); err != nil {
		c.lggr.Warnf("failed to populate initial transmission details: %v", err)
	}
	go c.poll()
	return nil
}

func (c *transmissionsCache) Close() error {
	close(c.stop)
	return nil
}

func (c *transmissionsCache) poll() {
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

func (c *transmissionsCache) LatestTransmissionDetails(
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
	defer c.tdLock.RUnlock()
	configDigest = c.transmissionDetails.digest
	epoch = c.transmissionDetails.epoch
	round = c.transmissionDetails.round
	latestAnswer = c.transmissionDetails.latestAnswer
	latestTimestamp = c.transmissionDetails.latestTimestamp
	return
}

func (c *transmissionsCache) LatestRoundRequested(
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
