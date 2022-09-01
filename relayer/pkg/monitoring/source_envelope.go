package monitoring

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	relayMonitoring "github.com/smartcontractkit/chainlink-relay/pkg/monitoring"
	relayUtils "github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/monitoring/encoding"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"go.uber.org/multierr"
)

func NewEnvelopeSourceFactory(
	starknetClient *starknet.Client,
	log relayMonitoring.Logger,
) relayMonitoring.SourceFactory {
	return &envelopeSourceFactory{
		starknetClient,
		log,
	}
}

type envelopeSourceFactory struct {
	starknetClient *starknet.Client
	log            relayMonitoring.Logger
}

func (s *envelopeSourceFactory) NewSource(
	chainConfig relayMonitoring.ChainConfig,
	feedConfig relayMonitoring.FeedConfig,
) (relayMonitoring.Source, error) {
	starknetChainConfig, ok := chainConfig.(StarknetConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	starknetFeedConfig, ok := feedConfig.(StarknetFeedConfig)
	if !ok {
		return nil, fmt.Errorf("expected feedConfig to be of type StarknetFeedConfig not %T", feedConfig)
	}
	return &envelopeSource{
		starknetChainConfig,
		starknetFeedConfig,
		s.starknetClient,
	}, nil
}

func (s *envelopeSourceFactory) GetType() string {
	return "envelope"
}

type envelopeSource struct {
	chainConfig    StarknetConfig
	feedConfig     StarknetFeedConfig
	starknetClient *starknet.Client
}

func (s *envelopeSource) Fetch(ctx context.Context) (interface{}, error) {
	envelope := relayMonitoring.Envelope{}
	var envelopeMu sync.Mutex
	var envelopeErr error
	subs := &relayUtils.Subprocesses{}

	subs.Go(func() {
		latestRoundData, newTransmissionEvent, err := s.fetchNewTransmissionEvent(ctx, s.feedConfig.ContractAddress)
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
		envelope.LatestAnswer = newTransmissionEvent.Answer
		envelope.LatestTimestamp = newTransmissionEvent.ObservationTimestamp
		envelope.JuelsPerFeeCoin = newTransmissionEvent.JuelsPerFeeCoin
	})

	subs.Go(func() {
		configSetEvent, err := s.fetchConfigSetEvent(ctx, s.feedConfig.ContractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetchConfigSetEvent failed: %w", err))
			return
		}
		envelope.ContractConfig = types.ContractConfig{
			ConfigDigest:          configSetEvent.LatestConfigDigest,
			ConfigCount:           configSetEvent.ConfigCount,
			Signers:               configSetEvent.Signers,
			Transmitters:          configSetEvent.Transmitters,
			F:                     configSetEvent.F,
			OnchainConfig:         configSetEvent.OnchainConfig,
			OffchainConfigVersion: configSetEvent.OffchainConfigVersion,
			OffchainConfig:        configSetEvent.OffchainConfig,
		}
	})

	subs.Go(func() {
		linkAvailable, err := s.fetchLinkAvailableForPayment(ctx, s.feedConfig.ContractAddress)
		envelopeMu.Lock()
		defer envelopeMu.Unlock()
		if err != nil {
			envelopeErr = multierr.Combine(envelopeErr, fmt.Errorf("fetchLinkAvailableForPayment failed: %w", err))
			return
		}
		envelope.LinkAvailableForPayment = linkAvailable
	})

	subs.Go(func() {
		balance, err := s.fetchLinkBalance(ctx, s.chainConfig.LinkTokenAddress, s.feedConfig.ContractAddress)
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

func (s *envelopeSource) fetchNewTransmissionEvent(ctx context.Context, contractAddress string) (encoding.RoundData, encoding.NewTransmisisonEvent, error) {
	results, err := s.starknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: contractAddress,
		Selector:        encoding.LatestRoundDataViewName,
	})
	if err != nil {
		return encoding.RoundData{}, encoding.NewTransmisisonEvent{}, fmt.Errorf("couldn't call the contract selector %s: %w", encoding.LatestRoundDataViewName, err)
	}
	var latestRoundData encoding.RoundData
	if err := latestRoundData.Unmarshal(results); err != nil {
		return encoding.RoundData{}, encoding.NewTransmisisonEvent{}, fmt.Errorf("failed to unmarshal RoundData from %v: %w", results, err)
	}
	block, err := s.starknetClient.BlockByNumberGateway(ctx, latestRoundData.BlockNumber)
	if err != nil {
		return encoding.RoundData{}, encoding.NewTransmisisonEvent{}, fmt.Errorf("failed to fetch block number %d: %w", latestRoundData.BlockNumber, err)
	}
	resultss, err := FilterEvents(block, contractAddress, encoding.NewTransmissionEventName)
	if err != nil {
		return encoding.RoundData{}, encoding.NewTransmisisonEvent{}, fmt.Errorf("failed to filter events of type '%s' from block number %d: %w", encoding.NewTransmissionEventName, latestRoundData.BlockNumber, err)
	}
	if len(resultss) == 0 {
		return encoding.RoundData{}, encoding.NewTransmisisonEvent{}, fmt.Errorf("could not find any events of type '%s' from block number %d", encoding.NewTransmissionEventName, latestRoundData.BlockNumber)
	}
	var newTransmissionEvent encoding.NewTransmisisonEvent
	if err := newTransmissionEvent.Unmarshal(resultss[0]); err != nil {
		return encoding.RoundData{}, encoding.NewTransmisisonEvent{}, fmt.Errorf("failed to unmarshal NewTransmissionEvent from %v: %w", resultss[0], err)
	}
	return latestRoundData, newTransmissionEvent, nil
}

func (s *envelopeSource) fetchConfigSetEvent(ctx context.Context, contractAddress string) (encoding.ConfigSetEvent, error) {
	results, err := s.starknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: contractAddress,
		Selector:        encoding.LatestConfigDetailsViewName,
	})
	if err != nil {
		return encoding.ConfigSetEvent{}, fmt.Errorf("couldn't call the contract selector %s: %w", encoding.LatestConfigDetailsViewName, err)
	}
	var latestConfigDetails encoding.LatestConfigDetails
	if err := latestConfigDetails.Unmarshal(results); err != nil {
		return encoding.ConfigSetEvent{}, fmt.Errorf("failed to unmarshal LatestConfigDetails from %v: %w", results, err)
	}
	block, err := s.starknetClient.BlockByNumberGateway(ctx, latestConfigDetails.BlockNumber)
	if err != nil {
		return encoding.ConfigSetEvent{}, fmt.Errorf("failed to fetch block number %d: %w", latestConfigDetails.BlockNumber, err)
	}
	resultss, err := FilterEvents(block, contractAddress, encoding.ConfigSetEventName)
	if err != nil {
		return encoding.ConfigSetEvent{}, fmt.Errorf("failed to filter events of type '%s' from block number %d: %w", encoding.ConfigSetEventName, latestConfigDetails.BlockNumber, err)
	}
	if len(resultss) == 0 {
		return encoding.ConfigSetEvent{}, fmt.Errorf("could not find any events of type '%s' from block number %d", encoding.ConfigSetEventName, latestConfigDetails.BlockNumber)
	}
	var configSetEvent encoding.ConfigSetEvent
	if err := configSetEvent.Unmarshal(resultss[0]); err != nil {
		return encoding.ConfigSetEvent{}, fmt.Errorf("failed to unmarshal ConfigSetEvent from %v: %w", resultss[0], err)
	}
	return configSetEvent, nil
}

func (s *envelopeSource) fetchLinkAvailableForPayment(ctx context.Context, contractAddress string) (*big.Int, error) {
	results, err := s.starknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: contractAddress,
		Selector:        encoding.LinkAvailableForPaymentViewName,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't call the contract selector %s: %w", encoding.LinkAvailableForPaymentViewName, err)
	}
	if len(results) < 1 {
		return nil, fmt.Errorf("insufficient data from %s event '%v': %w", encoding.LinkAvailableForPaymentViewName, results, err)
	}
	return encoding.DecodeBigInt(results[0])

}

func (s *envelopeSource) fetchLinkBalance(ctx context.Context, linkTokenAddress, contractAddress string) (*big.Int, error) {
	results, err := s.starknetClient.CallContract(ctx, starknet.CallOps{
		ContractAddress: linkTokenAddress,
		Selector:        encoding.BalanceOfMethod,
		Calldata:        []string{contractAddress},
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't call the contract selector %s: %w", encoding.BalanceOfMethod, err)
	}
	if len(results) < 1 {
		return nil, fmt.Errorf("insufficient data from balanceOf '%v': %w", results, err)
	}
	return encoding.DecodeBigInt(results[0])
}
