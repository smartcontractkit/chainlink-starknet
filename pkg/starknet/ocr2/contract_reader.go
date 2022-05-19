package ocr2

import (
	"context"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/client"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.ContractConfigTracker = (*ContractReader)(nil)
var _ median.MedianContract = (*ContractReader)(nil)

type ContractReader struct {
	chainReader client.Reader
	lggr        logger.Logger
}

func NewContractReader(chainReader client.Reader, lggr logger.Logger) *ContractReader {
	return &ContractReader{
		chainReader: chainReader,
		lggr:        lggr,
	}
}

func (c *ContractReader) Notify() <-chan struct{} {
	return nil
}

func (c *ContractReader) LatestConfigDetails(ctx context.Context) (changedInBlock uint64, configDigest types.ConfigDigest, err error) {
	// todo: implement
	return 0, types.ConfigDigest{}, nil
}

func (c *ContractReader) LatestConfig(ctx context.Context, changedInBlock uint64) (types.ContractConfig, error) {
	// todo: implement
	return types.ContractConfig{}, nil
}

func (c *ContractReader) LatestBlockHeight(ctx context.Context) (blockHeight uint64, err error) {
	// todo: implement
	return 0, nil
}

func (c *ContractReader) LatestTransmissionDetails(
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

func (c *ContractReader) LatestRoundRequested(
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
