package ocr2

import (
	"context"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.ContractConfigTracker = (*contractReader)(nil)
var _ median.MedianContract = (*contractReader)(nil)

type contractReader struct {
	reader Reader
	lggr   logger.Logger
}

func NewContractReader(chainReader Reader, lggr logger.Logger) *contractReader {
	return &contractReader{
		reader: chainReader,
		lggr:   lggr,
	}
}

func (c *contractReader) Notify() <-chan struct{} {
	return nil
}

func (c *contractReader) LatestConfigDetails(ctx context.Context) (changedInBlock uint64, configDigest types.ConfigDigest, err error) {
	// todo: implement
	return 0, types.ConfigDigest{}, nil
}

func (c *contractReader) LatestConfig(ctx context.Context, changedInBlock uint64) (types.ContractConfig, error) {
	// todo: implement
	return types.ContractConfig{}, nil
}

func (c *contractReader) LatestBlockHeight(ctx context.Context) (blockHeight uint64, err error) {
	// todo: implement
	return 0, nil
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
	// todo: implement
	return types.ConfigDigest{}, 0, 0, nil, time.Now(), nil
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
	// todo: implement
	return types.ConfigDigest{}, 0, 0, err
}
