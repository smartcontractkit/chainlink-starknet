package monitoring

import (
	"context"
	"fmt"
	"math/big"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/monitoring/encoding"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

type ProxyData struct {
	Answer *big.Int
}

func NewProxySourceFactory(
	starknetClient *starknet.Client,
	log relayMonitoring.Logger,
) relayMonitoring.SourceFactory {
	return &proxySourceFactory{
		starknetClient,
		log,
	}
}

type proxySourceFactory struct {
	starknetClient *starknet.Client
	log            relayMonitoring.Logger
}

func (s *proxySourceFactory) NewSource(
	chainConfig relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	starknetChainConfig, ok := chainConfig.(StarknetConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	starknetFeedConfig, ok := feedConfig.(StarknetFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	return &proxySource{
		starknetChainConfig,
		starknetFeedConfig,
		s.starknetClient,
	}, nil
}

func (s *proxySourceFactory) GetType() string {
	return "proxy"
}

type proxySource struct {
	chainConfig    StarknetConfig
	feedConfig     StarknetFeedConfig
	starknetClient *starknet.Client
}

func (s *proxySource) Fetch(ctx context.Context) (interface{}, error) {
	results, err := s.starknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: s.feedConfig.ProxyAddress,
		Selector:        encoding.LatestTransmissionDetailsViewName,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't call the contract selector %s on proxy address %s: %w", encoding.LatestTransmissionDetailsViewName, s.feedConfig.ProxyAddress, err)
	}
	var latestTransmissionDetails encoding.LatestTransmissionDetails
	if err := latestTransmissionDetails.Unmarshal(results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LatestTransmissionDetails from %v: %w", results, err)
	}
	return ProxyData{
		Answer: latestTransmissionDetails.LatestAnswer,
	}, nil
}
