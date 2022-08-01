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

	var chainReader starknet.Reader
	chainReader, err = starknet.NewClient(
		starknetConfig.ChainID,
		starknetConfig.RPCEndpoint,
		logger.With(log, "component", "chain-reader"),
		&starknetConfig.ReadTimeout,
	)
	if err != nil {
		log.Fatalw("failed to build a chain reader", "error", err)
	}

	envelopeSourceFactory := monitoring.NewEnvelopeSourceFactory(
		chainReader,
		logger.With(log, "component", "source-envelope"),
	)
	txResultsFactory := monitoring.NewTxResultsSourceFactory(
		chainReader,
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

	monitor.Run()
	log.Info("monitor stopped")
}
