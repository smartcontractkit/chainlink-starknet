package chainlink

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	relaytypes "github.com/smartcontractkit/chainlink-common/pkg/types"

	starkchain "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/chain"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
)

var _ relaytypes.Relayer = (*relayer)(nil) //nolint:staticcheck

type relayer struct {
	chain  starkchain.Chain
	stopCh services.StopChan

	lggr logger.Logger
}

func NewRelayer(lggr logger.Logger, chain starkchain.Chain) *relayer {
	return &relayer{
		chain:  chain,
		stopCh: make(chan struct{}),
		lggr:   lggr,
	}
}

func (r *relayer) Name() string {
	return r.lggr.Name()
}

func (r *relayer) Start(context.Context) error {
	return nil
}

func (r *relayer) Close() error {
	close(r.stopCh)
	return nil
}

func (r *relayer) Ready() error {
	return r.chain.Ready()
}

func (r *relayer) Healthy() error { return nil }

func (r *relayer) HealthReport() map[string]error {
	return map[string]error{r.Name(): r.Healthy()}
}

func (r *relayer) NewConfigProvider(args relaytypes.RelayArgs) (relaytypes.ConfigProvider, error) {
	var relayConfig RelayConfig

	err := json.Unmarshal(args.RelayConfig, &relayConfig)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal RelayConfig: %w", err)
	}

	reader, err := r.chain.Reader()
	if err != nil {
		return nil, fmt.Errorf("error in NewConfigProvider chain.Reader: %w", err)
	}
	configProvider, err := ocr2.NewConfigProvider(r.chain.ID(), args.ContractID, reader, r.chain.Config(), r.lggr)
	if err != nil {
		return nil, fmt.Errorf("coudln't initialize ConfigProvider: %w", err)
	}

	return configProvider, nil
}

func (r *relayer) NewMedianProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.MedianProvider, error) {
	var relayConfig RelayConfig

	err := json.Unmarshal(rargs.RelayConfig, &relayConfig)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal RelayConfig: %w", err)
	}

	if relayConfig.AccountAddress == "" {
		return nil, errors.New("no account address in relay config")
	}

	// todo: use pargs for median provider
	reader, err := r.chain.Reader()
	if err != nil {
		return nil, fmt.Errorf("error in NewMedianProvider chain.Reader: %w", err)
	}
	medianProvider, err := ocr2.NewMedianProvider(r.chain.ID(), rargs.ContractID, pargs.TransmitterID, relayConfig.AccountAddress, reader, r.chain.Config(), r.chain.TxManager(), r.lggr)
	if err != nil {
		return nil, fmt.Errorf("couldn't initilize MedianProvider: %w", err)
	}

	return medianProvider, nil
}

func (r *relayer) NewMercuryProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.MercuryProvider, error) {
	return nil, errors.New("mercury is not supported for starknet")
}

func (r *relayer) NewLLOProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.LLOProvider, error) {
	return nil, errors.New("data streams is not supported for starknet")
}

func (r *relayer) NewFunctionsProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.FunctionsProvider, error) {
	return nil, errors.New("functions are not supported for solana")
}

func (r *relayer) NewAutomationProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.AutomationProvider, error) {
	return nil, errors.New("automation is not supported for starknet")
}

func (r *relayer) NewPluginProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.PluginProvider, error) {
	return nil, errors.New("plugin provider is not supported for starknet")
}
