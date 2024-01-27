package monitoring

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"go.uber.org/multierr"

	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
	relayUtils "github.com/smartcontractkit/chainlink-common/pkg/utils"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
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
		caigotypes.StrToFelt(feedConfig.GetContractAddress()),
		caigotypes.StrToFelt(starknetChainConfig.GetLinkTokenAddress()),
		s.ocr2Reader,
	}, nil
}

func (s *envelopeSourceFactory) GetType() string {
	return "envelope"
}

type envelopeSource struct {
	contractAddress  caigotypes.Felt
	linkTokenAddress caigotypes.Felt
	ocr2Reader       ocr2.OCR2Reader
}

func (s *envelopeSource) Fetch(ctx context.Context) (interface{}, error) {
	envelope := relayMonitoring.Envelope{}
	var envelopeMu sync.Mutex
	var envelopeErr error
	subs := &relayUtils.Subprocesses{}

	subs.Go(func() {
		latestRoundData, newTransmissionEvent, err := s.fetchLatestNewTransmissionEvent(ctx, s.contractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetchLatestNewTransmissionEvent failed: %w", err))
			return
		}
		envelope.BlockNumber = latestRoundData.BlockNumber
		envelope.Transmitter = types.Account(newTransmissionEvent.Transmitter.String())
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
		availableLink, err := s.ocr2Reader.LinkAvailableForPayment(ctx, s.contractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetch LinkAvailableForPayment failed: %w", err))
			return
		}
		envelope.LinkAvailableForPayment = availableLink
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

func (s *envelopeSource) fetchLatestNewTransmissionEvent(ctx context.Context, contractAddress caigotypes.Felt) (
	latestRound ocr2.RoundData,
	transmission ocr2.NewTransmissionEvent,
	err error,
) {
	latestRound, err = s.ocr2Reader.LatestRoundData(ctx, contractAddress)
	if err != nil {
		return latestRound, transmission, fmt.Errorf("failed to fetch latest_round_data: %w", err)
	}
	transmissions, err := s.ocr2Reader.NewTransmissionsFromEventsAt(ctx, contractAddress, latestRound.BlockNumber)
	if err != nil {
		return latestRound, transmission, fmt.Errorf("failed to fetch new_transmission events: %w", err)
	}
	if len(transmissions) == 0 {
		// NOTE This shouldn't happen! LatestRound says this block should have a transmission and we didn't find any!
		return latestRound, transmission, fmt.Errorf("no transmissions found in the block %d", latestRound.BlockNumber)
	}
	for _, transmission = range transmissions {
		if transmission.RoundId == latestRound.RoundID {
			return latestRound, transmission, nil
		}
	}
	// NOTE! This also shouldn't happen! We found transmissions in the block suggested by LatestRound but they have a different round id!
	return latestRound, transmission, fmt.Errorf("no new_trasmission event found to correspond with the round id %d in block %d", latestRound.RoundID, latestRound.BlockNumber)
}

func (s *envelopeSource) fetchContractConfig(ctx context.Context, contractAddress caigotypes.Felt) (config ocr2.ContractConfig, err error) {
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

var zeroBigInt = big.NewInt(0)

func (s *envelopeSource) fetchLinkBalance(ctx context.Context, linkTokenAddress, contractAddress caigotypes.Felt) (*big.Int, error) {
	results, err := s.ocr2Reader.BaseReader().CallContract(ctx, starknet.CallOps{
		ContractAddress: linkTokenAddress,
		Selector:        "balanceOf",
		Calldata:        []string{contractAddress.String()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed call to ECR20 contract, balanceOf method: %w", err)
	}
	if len(results) < 1 {
		return nil, fmt.Errorf("insufficient data from balanceOf '%v': %w", results, err)
	}
	linkBalance := caigotypes.HexToBN(results[0])
	if linkBalance.Cmp(zeroBigInt) == 0 {
		return nil, fmt.Errorf("contract's LINK balance should not be zero")
	}
	return linkBalance, nil
}
