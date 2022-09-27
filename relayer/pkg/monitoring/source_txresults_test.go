package monitoring

import (
	"context"
	"testing"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	ocr2Mocks "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTxResultsSource(t *testing.T) {
	// This test makes sure that the mapping between the response from the ocr2.Client
	// method calls and the output of the TxResults source is correct.

	chainConfig := generateChainConfig()
	feedConfig := generateFeedConfig()

	ocr2Reader := ocr2Mocks.NewOCR2Reader(t)
	ocr2Reader.On(
		"LatestRoundData",
		mock.Anything, // ctx
		feedConfig.ContractAddress,
	).Return(ocr2ClientLatestRoundDataResponseForTxResults1, nil).Once()
	ocr2Reader.On(
		"LatestRoundData",
		mock.Anything, // ctx
		feedConfig.ContractAddress,
	).Return(ocr2ClientLatestRoundDataResponseForTxResults2, nil).Once()

	factory := NewTxResultsSourceFactory(ocr2Reader)
	source, err := factory.NewSource(chainConfig, feedConfig)
	require.NoError(t, err)
	// First call identifies no new transactions.
	rawTxResults, err := source.Fetch(context.Background())
	require.NoError(t, err)
	txResults, ok := rawTxResults.(relayMonitoring.TxResults)
	require.True(t, ok)
	require.Equal(t, txResults.NumSucceeded, uint64(0))
	require.Equal(t, txResults.NumFailed, uint64(0))
	// Second call identifies new transactions
	rawTxResults, err = source.Fetch(context.Background())
	require.NoError(t, err)
	txResults, ok = rawTxResults.(relayMonitoring.TxResults)
	require.True(t, ok)
	require.Equal(t, txResults.NumSucceeded, uint64(1))
	require.Equal(t, txResults.NumFailed, uint64(0))
}

var (
	ocr2ClientLatestRoundDataResponseForTxResults1 = ocr2.RoundData{
		RoundID: 100,
	}
	ocr2ClientLatestRoundDataResponseForTxResults2 = ocr2.RoundData{
		RoundID: 101,
	}
)
