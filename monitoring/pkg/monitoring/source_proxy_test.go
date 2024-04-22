package monitoring

import (
	"context"
	"math/big"
	"testing"
	"time"

	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	ocr2Mocks "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/mocks"
)

func TestProxySource(t *testing.T) {
	// This test makes sure that the mapping between the response from the ocr2.Client
	// method calls and the output of the Proxy source is correct.

	chainConfig := generateChainConfig()
	feedConfig := generateFeedConfig()

	proxyContractAddressFelt, err := starknetutils.HexToFelt(feedConfig.ProxyAddress)
	require.NoError(t, err)

	ocr2Reader := ocr2Mocks.NewOCR2Reader(t)
	ocr2Reader.On(
		"LatestRoundData",
		mock.Anything, // ctx
		proxyContractAddressFelt,
	).Return(ocr2ClientLatestRoundDataResponseForProxy, nil).Once()

	factory := NewProxySourceFactory(ocr2Reader)
	source, err := factory.NewSource(chainConfig, feedConfig)
	require.NoError(t, err)
	rawProxyData, err := source.Fetch(context.Background())
	require.NoError(t, err)
	proxyData, ok := rawProxyData.(ProxyData)
	require.True(t, ok)

	require.Equal(t,
		ocr2ClientLatestRoundDataResponseForProxy.Answer.String(),
		proxyData.Answer.String(),
	)
}

var (
	ocr2ClientLatestRoundDataResponseForProxy = ocr2.RoundData{
		RoundID:     9,
		Answer:      big.NewInt(10000),
		BlockNumber: 777,
		StartedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
)
