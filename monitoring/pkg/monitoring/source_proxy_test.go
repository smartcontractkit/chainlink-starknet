package monitoring

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	ocr2Mocks "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/mocks"
)

func TestProxySource(t *testing.T) {
	// This test makes sure that the mapping between the response from the ocr2.Client
	// method calls and the output of the Proxy source is correct.

	chainConfig := generateChainConfig()
	feedConfig := generateFeedConfig()

	ocr2Reader := ocr2Mocks.NewOCR2Reader(t)
	ocr2Reader.On(
		"LatestTransmissionDetails",
		mock.Anything, // ctx
		feedConfig.ContractAddress,
	).Return(ocr2ClientLatestTransmissionDetailsResponseForProxy, nil).Once()

	factory := NewProxySourceFactory(ocr2Reader)
	source, err := factory.NewSource(relayMonitoring.Params{
		ChainConfig: chainConfig,
		FeedConfig:  feedConfig,
	})
	require.NoError(t, err)
	rawProxyData, err := source.Fetch(context.Background())
	require.NoError(t, err)
	proxyData, ok := rawProxyData.(ProxyData)
	require.True(t, ok)

	require.Equal(t,
		ocr2ClientLatestTransmissionDetailsResponseForProxy.LatestAnswer.String(),
		proxyData.Answer.String(),
	)
}

var (
	ocr2ClientLatestTransmissionDetailsResponseForProxy = ocr2.TransmissionDetails{
		Digest:          types.ConfigDigest{0x0, 0x4, 0x18, 0xe5, 0x44, 0xab, 0xa8, 0x18, 0x15, 0xa5, 0x2b, 0xf0, 0x11, 0x58, 0xc6, 0x9b, 0x38, 0x8a, 0x48, 0x9f, 0x76, 0xd, 0xd8, 0x3d, 0x84, 0x3f, 0x1d, 0x31, 0x22, 0xdb, 0x78, 0xa},
		Epoch:           0x1,
		Round:           0x9,
		LatestAnswer:    big.NewInt(10000),
		LatestTimestamp: time.Now(),
	}
)
