package monitoring

import (
	"context"
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	starknetutils "github.com/NethermindEth/starknet.go/utils"

	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
)

type ProxyData struct {
	Answer *big.Int
}

func NewProxySourceFactory(
	ocr2Reader ocr2.OCR2Reader,
) relayMonitoring.SourceFactory {
	return &proxySourceFactory{
		ocr2Reader,
	}
}

type proxySourceFactory struct {
	ocr2Reader ocr2.OCR2Reader
}

func (s *proxySourceFactory) NewSource(
	_ relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	starknetFeedConfig, ok := feedConfig.(StarknetFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	contractAddress, err := starknetutils.HexToFelt(starknetFeedConfig.ProxyAddress)
	if err != nil {
		return nil, err
	}
	return &proxySource{
		contractAddress,
		s.ocr2Reader,
	}, nil
}

func (s *proxySourceFactory) GetType() string {
	return "proxy"
}

type proxySource struct {
	contractAddress *felt.Felt
	ocr2Reader      ocr2.OCR2Reader
}

func (s *proxySource) Fetch(ctx context.Context) (interface{}, error) {
	latestRoundData, err := s.ocr2Reader.LatestRoundData(ctx, s.contractAddress)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch latest_round_data: %w", err)
	}
	return ProxyData{
		Answer: latestRoundData.Answer,
	}, nil
}
