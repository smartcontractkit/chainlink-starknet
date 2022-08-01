package monitoring

import (
	"context"
	"fmt"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func NewEnvelopeSourceFactory(
	client starknet.Reader,
	log relayMonitoring.Logger,
) relayMonitoring.SourceFactory {
	return &envelopeSourceFactory{
		client,
		log,
	}
}

type envelopeSourceFactory struct {
	client starknet.Reader
	log    relayMonitoring.Logger
}

func (s *envelopeSourceFactory) NewSource(
	_ relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	starknetFeedConfig, ok := feedConfig.(StarknetFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	return &envelopeSource{
		s.client,
		starknetFeedConfig,
	}, nil
}

func (s *envelopeSourceFactory) GetType() string {
	return "envelope"
}

type envelopeSource struct {
	client     starknet.Reader
	feedConfig StarknetFeedConfig
}

func (s *envelopeSource) Fetch(ctx context.Context) (interface{}, error) {
	envelope := relayMonitoring.Envelope{}
	return envelope, nil
}
