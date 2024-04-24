package monitoring

import (
	"context"
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	starknetutils "github.com/NethermindEth/starknet.go/utils"

	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
)

type TransmissionInfo struct {
	GasPrice          *big.Int
	ObservationLength uint32
}

type TransmissionsEnvelope struct {
	Transmissions []TransmissionInfo
}

func NewTransmissionDetailsSourceFactory(
	ocr2Reader ocr2.OCR2Reader,
) relayMonitoring.SourceFactory {
	return &transmissionDetailsSourceFactory{
		ocr2Reader,
	}
}

type transmissionDetailsSourceFactory struct {
	ocr2Reader ocr2.OCR2Reader
}

func (s *transmissionDetailsSourceFactory) NewSource(
	_ relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	starknetFeedConfig, ok := feedConfig.(StarknetFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	contractAddress, err := starknetutils.HexToFelt(starknetFeedConfig.ContractAddress)
	if err != nil {
		return nil, err
	}
	return &transmissionDetailsSource{
		contractAddress,
		s.ocr2Reader,
	}, nil
}

func (s *transmissionDetailsSourceFactory) GetType() string {
	return "transmission details"
}

type transmissionDetailsSource struct {
	contractAddress *felt.Felt
	ocr2Reader      ocr2.OCR2Reader
}

func (s *transmissionDetailsSource) Fetch(ctx context.Context) (interface{}, error) {
	latestRound, err := s.ocr2Reader.LatestRoundData(ctx, s.contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest_round_data: %w", err)
	}
	transmissions, err := s.ocr2Reader.NewTransmissionsFromEventsAt(ctx, s.contractAddress, latestRound.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch transmission events: %w", err)
	}
	var envelope TransmissionsEnvelope
	for _, t := range transmissions {
		envelope.Transmissions = append(
			envelope.Transmissions,
			TransmissionInfo{GasPrice: t.GasPrice, ObservationLength: t.ObservationsLen},
		)
	}

	return envelope, nil
}
