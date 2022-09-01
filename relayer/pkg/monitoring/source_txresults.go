package monitoring

import (
	"context"
	"fmt"
	"sync"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/monitoring/encoding"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func NewTxResultsSourceFactory(
	starknetClient *starknet.Client,
	log relayMonitoring.Logger,
) relayMonitoring.SourceFactory {
	return &proxySourceFactory{
		starknetClient,
		log,
	}
}

type txResultsSourceFactory struct {
	starknetClient *starknet.Client
	log            relayMonitoring.Logger
}

func (s *txResultsSourceFactory) NewSource(
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

func (s *txResultsSourceFactory) GetType() string {
	return "txresults"
}

type txResultsSource struct {
	chainConfig    StarknetConfig
	feedConfig     StarknetFeedConfig
	starknetClient *starknet.Client

	prevRoundID   uint32
	prevRoundIDMu sync.Mutex
}

func (s *txResultsSource) Fetch(ctx context.Context) (interface{}, error) {
	results, err := s.starknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: s.feedConfig.ContractAddress,
		Selector:        encoding.LatestRoundDataViewName,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't call the contract selector %s on contract %s: %w", encoding.LatestRoundDataViewName, s.feedConfig.ContractAddress, err)
	}
	var latestRoundData encoding.RoundData
	if err := latestRoundData.Unmarshal(results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LatestTransmissionDetails from %v: %w", results, err)
	}
	s.prevRoundIDMu.Lock()
	defer s.prevRoundIDMu.Unlock()
	var numSucceeded uint32
	if s.prevRoundID != 0 {
		numSucceeded = latestRoundData.RoundID - s.prevRoundID
		s.prevRoundID = latestRoundData.RoundID
	}
	return relayMonitoring.TxResults{NumSucceeded: uint64(numSucceeded), NumFailed: 0}, nil
}
