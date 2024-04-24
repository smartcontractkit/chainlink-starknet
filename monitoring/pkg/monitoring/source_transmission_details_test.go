package monitoring

import (
	"context"
	"math/big"
	"testing"

	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ocr2Mocks "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/mocks"
)

func TestTransmissionDetailsSource(t *testing.T) {
	chainConfig := generateChainConfig()
	feedConfig := generateFeedConfig()

	contractAddressFelt, err := starknetutils.HexToFelt(feedConfig.ContractAddress)
	require.NoError(t, err)

	ocr2Reader := ocr2Mocks.NewOCR2Reader(t)
	blockNumber := uint64(777)
	ocr2Reader.On(
		"LatestRoundData",
		mock.Anything, // ctx
		contractAddressFelt,
	).Return(ocr2.RoundData{BlockNumber: blockNumber}, nil).Once()
	ocr2Reader.On(
		"NewTransmissionsFromEventsAt",
		mock.Anything, // ctx
		contractAddressFelt,
		blockNumber,
	).Return(
		[]ocr2.NewTransmissionEvent{
			{
				GasPrice:        new(big.Int).SetUint64(7),
				ObservationsLen: 7,
			},
		},
		nil,
	).Once()

	factory := NewTransmissionDetailsSourceFactory(ocr2Reader)
	source, err := factory.NewSource(chainConfig, feedConfig)
	require.NoError(t, err)

	transmissionsEnvelope, err := source.Fetch(context.Background())
	require.NoError(t, err)
	envelope, ok := transmissionsEnvelope.(TransmissionsEnvelope)
	require.True(t, ok)

	require.Equal(t, len(envelope.Transmissions), 1)
	require.Equal(t, envelope.Transmissions[0].GasPrice.Uint64(), uint64(7))
	require.Equal(t, envelope.Transmissions[0].ObservationLength, uint32(7))
}
