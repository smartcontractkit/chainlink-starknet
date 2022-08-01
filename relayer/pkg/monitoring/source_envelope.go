package monitoring

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

func NewEnvelopeSourceFactory(
	reader starknet.Reader,
	log relayMonitoring.Logger,
) relayMonitoring.SourceFactory {
	return &envelopeSourceFactory{
		reader,
		log,
	}
}

type envelopeSourceFactory struct {
	reader starknet.Reader
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
	client, err := ocr2.NewClient(s.reader, logger.With(s.log, "component", "ocr2-client"))
	if err != nil {
		return nil, fmt.Errorf("failed to build an ocr2.Client instance: %w", err)
	}
	// TODO (dru) maybe use the cached version!
	reader := ocr2.NewContractReader(
		starknetFeedConfig.ContractAddress,
		client,
		logger.With(s.log, "component", "ocr2-contract-reader"),
	)
	return &envelopeSource{
		reader,
		starknetFeedConfig,
	}, nil
}

func (s *envelopeSourceFactory) GetType() string {
	return "envelope"
}

type envelopeSource struct {
	reader     ocr2.Reader
	feedConfig StarknetFeedConfig
}

func (s *envelopeSource) Fetch(ctx context.Context) (interface{}, error) {
	changedInBlock, _, err := s.reader.LatestConfigDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest config details: %w", err)
	}
	config, err := s.reader.LatestConfig(ctx, changedInBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest config: %w", err)
	}
	configDigest, epoch, round, latestAnswer, latestTimestamp, err := s.reader.LatestTransmissionDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest transmission details: %w", err)
	}

	envelope := relayMonitoring.Envelope{
		// latest transmission details
		ConfigDigest:    configDigest,
		Epoch:           epoch,
		Round:           round,
		LatestAnswer:    latestAnswer,
		LatestTimestamp: latestTimestamp,

		// latest contract config
		ContractConfig: config,

		// extra
		BlockNumber:             0,                 // TODO (dru)
		Transmitter:             types.Account(""), // TODO (dru)
		LinkBalance:             new(big.Int),      // TODO (dru)
		LinkAvailableForPayment: new(big.Int),      // TODO (dru)

		// The "fee coin" is different for each chain.
		JuelsPerFeeCoin:   new(big.Int), // TODO (dru)
		AggregatorRoundID: 0,            // TODO (dru)
	}
	return envelope, nil
}
