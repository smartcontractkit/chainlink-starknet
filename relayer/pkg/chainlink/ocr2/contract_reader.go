package ocr2

import (
	"context"
	"github.com/pkg/errors"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.ContractConfigTracker = (*contractReader)(nil)
var _ median.MedianContract = (*contractReader)(nil)

type contractReader struct {
	address string
	reader  OCR2Reader
	lggr    logger.Logger
}

func NewContractReader(address string, reader OCR2Reader, lggr logger.Logger) *contractReader {
	return &contractReader{
		address: address,
		reader:  reader,
		lggr:    lggr,
	}
}

func (c *contractReader) Notify() <-chan struct{} {
	return nil
}

func (c *contractReader) LatestConfigDetails(ctx context.Context) (changedInBlock uint64, configDigest types.ConfigDigest, err error) {
	resp, err := c.reader.LatestConfigDetails(ctx, c.address)
	if err != nil {
		return changedInBlock, configDigest, errors.Wrap(err, "couldn't get latest config details")
	}

	changedInBlock = resp.Block
	configDigest = resp.Digest

	return
}

func (c *contractReader) LatestConfig(ctx context.Context, changedInBlock uint64) (config types.ContractConfig, err error) {
	resp, err := c.reader.ConfigFromEventAt(ctx, c.address, changedInBlock)
	if err != nil {
		return config, errors.Wrap(err, "couldn't get latest config")
	}

	config = resp.Config

	return
}

func (c *contractReader) LatestBlockHeight(ctx context.Context) (blockHeight uint64, err error) {
	blockHeight, err = c.reader.BaseClient().LatestBlockHeight(ctx)
	if err != nil {
		return blockHeight, errors.Wrap(err, "couldn't get latest block height")
	}

	return
}

func (c *contractReader) LatestTransmissionDetails(
	ctx context.Context,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	round uint8,
	latestAnswer *big.Int,
	latestTimestamp time.Time,
	err error,
) {
	transmissionDetails, err := c.reader.LatestTransmissionDetails(ctx, c.address)
	if err != nil {
		err = errors.Wrap(err, "couldn't get transmission details")
	}

	configDigest = transmissionDetails.digest
	epoch = transmissionDetails.epoch
	round = transmissionDetails.round
	latestAnswer = transmissionDetails.latestAnswer
	latestTimestamp = transmissionDetails.latestTimestamp

	return
}

func (c *contractReader) LatestRoundRequested(
	ctx context.Context,
	lookback time.Duration,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	round uint8,
	err error,
) {
	transmissionDetails, err := c.reader.LatestTransmissionDetails(ctx, c.address)
	if err != nil {
		err = errors.Wrap(err, "couldn't get transmission details")
	}

	configDigest = transmissionDetails.digest
	epoch = transmissionDetails.epoch
	round = transmissionDetails.round

	return
}
