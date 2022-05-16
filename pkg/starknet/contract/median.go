package contract

import (
	"context"
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ median.MedianContract = (*ContractTracker)(nil)

func (c *ContractTracker) LatestTransmissionDetails(
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

func (c *ContractTracker) LatestRoundRequested(
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
