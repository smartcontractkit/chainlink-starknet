package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/erc20"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-starknet/monitoring/pkg/monitoring"
)

func main() {
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
		log.Fatalw("failed to parse starknet specific configuration", "error", err)
		return
	}

	readTimeout := starknetConfig.GetReadTimeout()
	starknetClient, err := starknet.NewClient(
		starknetConfig.GetChainID(),
		starknetConfig.GetRPCEndpoint(),
		starknetConfig.GetRPCApiKey(),
		logger.With(log, "component", "starknet-client"),
		&readTimeout,
	)
	if err != nil {
		log.Fatalw("failed to build a starknet.Client", "error", err)
	}
	ocr2Client, err := ocr2.NewClient(
		starknetClient,
		logger.With(log, "component", "ocr2-client"),
	)
	if err != nil {
		log.Fatalw("failed to build a ocr2.Client", "error", err)
	}

	strTokenClient, err := erc20.NewClient(
		starknetClient,
		logger.With(log, "component", "erc20-client"),
		starknetConfig.GetStrkTokenAddress(),
	)

	if err != nil {
		log.Fatalw("failed to build erc20-client", "error", err)
	}

	envelopeSourceFactory := monitoring.NewEnvelopeSourceFactory(ocr2Client)
	txResultsFactory := monitoring.NewTxResultsSourceFactory(ocr2Client)

	monitor, err := monitoring.NewMonitorPrometheusOnly(
		make(chan struct{}),
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

	// per-feed factories
	proxySourceFactory := monitoring.NewProxySourceFactory(ocr2Client)
	transmissionsDetailsSourceFactory := monitoring.NewTransmissionDetailsSourceFactory(ocr2Client)
	monitor.SourceFactories = append(monitor.SourceFactories, proxySourceFactory, transmissionsDetailsSourceFactory)

	metricsBuilder := monitoring.NewMetrics(logger.With(log, "component", "starknet-metrics-builder"))

	prometheusExporterFactory := monitoring.NewPrometheusExporterFactory(metricsBuilder)
	transmissionsDetailsExporterFactory := monitoring.NewTransmissionDetailsExporterFactory(metricsBuilder)
	monitor.ExporterFactories = append(monitor.ExporterFactories, prometheusExporterFactory, transmissionsDetailsExporterFactory)

	// network factories
	nodeBalancesSourceFactory := monitoring.NewNodeBalancesSourceFactory(strTokenClient)
	monitor.NetworkSourceFactories = append(monitor.NetworkSourceFactories, nodeBalancesSourceFactory)

	nodeBalancesExporterFactory := monitoring.NewNodeBalancesExporterFactory(
		logger.With(log, "node-balances-exporter"),
		metricsBuilder,
	)
	monitor.NetworkExporterFactories = append(monitor.NetworkExporterFactories, nodeBalancesExporterFactory)

	monitor.Run()
	log.Info("monitor stopped")
}
