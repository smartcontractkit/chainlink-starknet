package main

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-starknet/monitoring/pkg/monitoring"
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
		log.Fatalw("failed to parse starknet specific configuration", "error", err)
		return
	}

	readTimeout := starknetConfig.GetReadTimeout()
	starknetClient, err := starknet.NewClient(
		starknetConfig.GetChainID(),
		starknetConfig.GetRPCEndpoint(),
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

	envelopeSourceFactory := monitoring.NewEnvelopeSourceFactory(ocr2Client)
	txResultsFactory := monitoring.NewTxResultsSourceFactory(ocr2Client)

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

	proxySourceFactory := monitoring.NewProxySourceFactory(ocr2Client)
	monitor.SourceFactories = append(monitor.SourceFactories, proxySourceFactory)

	prometheusExporterFactory := monitoring.NewPrometheusExporterFactory(
		monitoring.NewMetrics(logger.With(log, "component", "starknet-metrics")),
	)
	monitor.ExporterFactories = append(monitor.ExporterFactories, prometheusExporterFactory)

	monitor.Run()
	log.Info("monitor stopped")
}
