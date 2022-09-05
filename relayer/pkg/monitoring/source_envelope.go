package monitoring

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	relayUtils "github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"go.uber.org/multierr"
)

func NewEnvelopeSourceFactory(
	ocr2Reader ocr2.OCR2Reader,
) relayMonitoring.SourceFactory {
	return &envelopeSourceFactory{
		ocr2Reader,
	}
}

type envelopeSourceFactory struct {
	ocr2Reader ocr2.OCR2Reader
}

func (s *envelopeSourceFactory) NewSource(
	chainConfig relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	starknetChainConfig, ok := chainConfig.(StarknetConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	return &envelopeSource{
		feedConfig.GetContractAddress(),
		starknetChainConfig.LinkTokenAddress,
		s.ocr2Reader,
	}, nil
}

func (s *envelopeSourceFactory) GetType() string {
	return "envelope"
}

type envelopeSource struct {
	contractAddress  string
	linkTokenAddress string
	ocr2Reader       ocr2.OCR2Reader
}

func (s *envelopeSource) Fetch(ctx context.Context) (interface{}, error) {
	envelope := relayMonitoring.Envelope{}
	var envelopeMu sync.Mutex
	var envelopeErr error
	subs := &relayUtils.Subprocesses{}

	subs.Go(func() {
		latestRoundData, newTransmissionEvent, err := s.fetchNewTransmissionEvent(ctx, s.contractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetchNewTransmissionEvent failed: %w", err))
			return
		}
		envelope.BlockNumber = latestRoundData.BlockNumber
		envelope.AggregatorRoundID = latestRoundData.RoundID
		envelope.ConfigDigest = newTransmissionEvent.ConfigDigest
		envelope.Epoch = newTransmissionEvent.Epoch
		envelope.Round = newTransmissionEvent.Round
		envelope.LatestAnswer = newTransmissionEvent.LatestAnswer
		envelope.LatestTimestamp = newTransmissionEvent.LatestTimestamp
		envelope.JuelsPerFeeCoin = newTransmissionEvent.JuelsPerFeeCoin
	})

	subs.Go(func() {
		contractConfig, err := s.fetchContractConfig(ctx, s.contractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetchContractConfig failed: %w", err))
			return
		}
		envelope.ContractConfig = contractConfig.Config
	})

	subs.Go(func() {
		linkAvailable, err := s.fetchLinkAvailableForPayment(ctx, s.contractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetchLinkAvailableForPayment failed: %w", err))
			return
		}
		envelope.LinkAvailableForPayment = linkAvailable
	})

	subs.Go(func() {
		balance, err := s.fetchLinkBalance(ctx, s.linkTokenAddress, s.contractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetchLinkBalance failed: %w", err))
			return
		}
		envelope.LinkBalance = balance
	})

	subs.Wait()
	return envelope, envelopeErr
}

func (s *envelopeSource) fetchNewTransmissionEvent(ctx context.Context, contractAddress string) (latestRound ocr2.RoundData, transmission ocr2.NewTransmissionEvent, err error) {
	latestRound, err = s.ocr2Reader.LatestRoundData(ctx, contractAddress)
	if err != nil {
		return latestRound, transmission, fmt.Errorf("failed to fetch latest_round_data: %w", err)
	}
	transmission, err = s.ocr2Reader.NewTransmissionEventAt(ctx, contractAddress, latestRound.BlockNumber)
	if err != nil {
		return latestRound, transmission, fmt.Errorf("failed to fetch last new_transmission event: %w", err)
	}
	return latestRound, transmission, nil
}

func (s *envelopeSource) fetchContractConfig(ctx context.Context, contractAddress string) (config ocr2.ContractConfig, err error) {
	configDetails, err := s.ocr2Reader.LatestConfigDetails(ctx, contractAddress)
	if err != nil {
		return config, fmt.Errorf("couldn't fetch latest config details for contract '%s': %w", contractAddress, err)
	}
	config, err = s.ocr2Reader.ConfigFromEventAt(ctx, contractAddress, configDetails.Block)
	if err != nil {
		return config, fmt.Errorf("couldn't fetch config at block '%d' for contract '%s': %w", configDetails.Block, contractAddress, err)
	}
	return config, nil
}

func (s *envelopeSource) fetchLinkAvailableForPayment(ctx context.Context, contractAddress string) (*big.Int, error) {
	results, err := s.ocr2Reader.BaseReader().CallContract(ctx, starknet.CallOps{
		ContractAddress: contractAddress,
		Selector:        "link_available_for_payments",
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't call the contract '%s' with selector 'link_available_for_payments': %w", contractAddress, err)
	}
	if len(results) < 1 {
		return nil, fmt.Errorf("insufficient data from contract %s with selector 'link_available_for_payments'. Expected 1 got %d", contractAddress, len(results))
	}
	return junotypes.HexToFelt(results[0]).Big(), nil
}

func (s *envelopeSource) fetchLinkBalance(ctx context.Context, linkTokenAddress, contractAddress string) (*big.Int, error) {
	results, err := s.ocr2Reader.BaseReader().CallContract(ctx, starknet.CallOps{
		ContractAddress: linkTokenAddress,
		Selector:        "balanceOf",
		Calldata:        []string{contractAddress},
	})
	if err != nil {
		return nil, fmt.Errorf("failed call to ECR20 contract, balanceOf method: %w", err)
	}
	if len(results) < 1 {
		return nil, fmt.Errorf("insufficient data from balanceOf '%v': %w", results, err)
	}
	return junotypes.HexToFelt(results[0]).Big(), nil
}
