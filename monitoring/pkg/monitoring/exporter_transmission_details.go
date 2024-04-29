package monitoring

import (
	"context"
	"fmt"

	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
)

// NewPrometheusExporterFactory builds an implementation of the Exporter for prometheus.
func NewTransmissionDetailsExporterFactory(
	metrics Metrics,
) relayMonitoring.ExporterFactory {
	return &transmissionDetailsExporterFactory{
		metrics,
	}
}

type transmissionDetailsExporterFactory struct {
	metrics Metrics
}

func (p *transmissionDetailsExporterFactory) NewExporter(
	params relayMonitoring.ExporterParams,
) (relayMonitoring.Exporter, error) {
	starknetFeedConfig, ok := params.FeedConfig.(StarknetFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", params.FeedConfig)
	}
	return &transmissionDetailsExporter{
		params.ChainConfig,
		starknetFeedConfig,
		p.metrics,
	}, nil
}

type transmissionDetailsExporter struct {
	chainConfig relayMonitoring.ChainConfig
	feedConfig  StarknetFeedConfig
	metrics     Metrics
}

func (p *transmissionDetailsExporter) Export(ctx context.Context, data interface{}) {
	transmissionsEnvelope, found := data.(TransmissionsEnvelope)
	if !found {
		return
	}

	for _, t := range transmissionsEnvelope.Transmissions {
		observationLength := float64(t.ObservationLength)
		p.metrics.SetReportObservations(
			observationLength,
			p.feedConfig.ContractAddress,
			p.feedConfig.GetID(),
			p.chainConfig.GetChainID(),
			p.feedConfig.GetContractStatus(),
			p.feedConfig.GetContractType(),
			p.feedConfig.Name,
			p.feedConfig.Path,
			p.chainConfig.GetNetworkID(),
			p.chainConfig.GetNetworkName(),
		)
	}
}

func (p *transmissionDetailsExporter) Cleanup(_ context.Context) {
	p.metrics.CleanupReportObservations(
		p.feedConfig.GetContractAddress(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
}
