package monitoring

import (
	"context"
	"fmt"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func NewTxResultsSourceFactory(
	client starknet.Reader,
	log relayMonitoring.Logger,
) relayMonitoring.SourceFactory {
	return &txResultsSourceFactory{
		client,
		log,
	}
}

type txResultsSourceFactory struct {
	client starknet.Reader
	log    relayMonitoring.Logger
}

func (s *txResultsSourceFactory) NewSource(
	_ relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	starknetFeedConfig, ok := feedConfig.(StarknetFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarkNetFeedConfig not %T", feedConfig)
	}
	return &txResultsSource{
		s.client,
		s.log,
		starknetFeedConfig,
	}, nil
}

func (s *txResultsSourceFactory) GetType() string {
	return "txresults"
}

type txResultsSource struct {
	client     starknet.Reader
	log        relayMonitoring.Logger
	feedConfig StarknetFeedConfig
}

func (t *txResultsSource) Fetch(ctx context.Context) (interface{}, error) {
	return relayMonitoring.TxResults{}, nil
}
