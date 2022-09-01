package main

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/monitoring"
)

func main() {
	ctx := context.Background()

	log, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if serr := log.Sync(); serr != nil {
			fmt.Printf("Error while closing Logger: %v\n", serr)
		}
	}()

	starknetConfig, err := monitoring.ParseStarknetConfig()
	if err != nil {
		log.Fatalw("failed to parse terra specific configuration", "error", err)
		return
	}

	starknetClient, err := starknet.NewClient(
		starknetConfig.ChainID,
		starknetConfig.RPCEndpoint,
		logger.With(log, "component", "starknet-client"),
		&starknetConfig.ReadTimeout,
	)
	if err != nil {
		log.Fatalw("failed to build a starknet.Client", "error", err)
	}

	envelopeSourceFactory := monitoring.NewEnvelopeSourceFactory(
		starknetClient,
		logger.With(log, "component", "source-envelope"),
	)
	txResultsFactory := monitoring.NewTxResultsSourceFactory(
		starknetClient,
		logger.With(log, "component", "source-txresults"),
	)

	monitor, err := relayMonitoring.NewMonitor(
		ctx,
		logger.With(log, "component", "monitor"),
		starknetConfig,
		envelopeSourceFactory,
		txResultsFactory,
		monitoring.StarknetFeedsParser,
		monitoring.StarknetNodesParser,
	)
	if err != nil {
		log.Fatalw("failed to build monitor", "error", err)
		return
	}

	proxySourceFactory := monitoring.NewProxySourceFactory(
		starknetClient,
		logger.With(log, "component", "source-proxy"),
	)
	monitor.SourceFactories = append(monitor.SourceFactories, proxySourceFactory)

	prometheusExporterFactory := monitoring.NewPrometheusExporterFactory(
		logger.With(log, "component", "starknet-prometheus-exporter"),
		monitoring.NewMetrics(logger.With(log, "component", "starknet-metrics")),
	)
	monitor.ExporterFactories = append(monitor.ExporterFactories, prometheusExporterFactory)

	monitor.Run()
	log.Info("monitor stopped")
}
