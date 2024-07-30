package chainlink

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"

	starkchain "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/chain"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
)

var _ relaytypes.Relayer = (*relayer)(nil) //nolint:staticcheck

type relayer struct {
	chain starkchain.Chain

	lggr logger.Logger
}

func NewRelayer(lggr logger.Logger, chain starkchain.Chain, capRegistry core.CapabilitiesRegistry) *relayer {
	return &relayer{
		chain: chain,
		lggr:  logger.Named(lggr, "Relayer"),
	}
}

func (r *relayer) Name() string {
	return r.lggr.Name()
}

func (r *relayer) Start(context.Context) error {
	return nil
}

func (r *relayer) Close() error {
	return nil
}

func (r *relayer) Ready() error {
	return r.chain.Ready()
}

func (r *relayer) Healthy() error { return nil }

func (r *relayer) HealthReport() map[string]error {
	return map[string]error{r.Name(): r.Healthy()}
}

func (r *relayer) NewChainWriter(_ context.Context, _ []byte) (relaytypes.ChainWriter, error) {
	return nil, errors.New("chain writer is not supported for starknet")
}

func (r *relayer) NewContractReader(ctx context.Context, _ []byte) (relaytypes.ContractReader, error) {
	return nil, errors.New("contract reader is not supported for starknet")
}

func (r *relayer) LatestHead(ctx context.Context) (relaytypes.Head, error) {
	return r.chain.LatestHead(ctx)
}

func (r *relayer) GetChainStatus(ctx context.Context) (relaytypes.ChainStatus, error) {
	return r.chain.GetChainStatus(ctx)
}

func (r *relayer) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) (stats []relaytypes.NodeStatus, nextPageToken string, total int, err error) {
	return r.chain.ListNodeStatuses(ctx, pageSize, pageToken)
}

func (r *relayer) Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error {
	return r.chain.Transact(ctx, from, to, amount, balanceCheck)
}

func (r *relayer) NewConfigProvider(ctx context.Context, args relaytypes.RelayArgs) (relaytypes.ConfigProvider, error) {
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

func (r *relayer) NewMedianProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.MedianProvider, error) {
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

func (r *relayer) NewMercuryProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.MercuryProvider, error) {
	return nil, errors.New("mercury is not supported for starknet")
}

func (r *relayer) NewLLOProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.LLOProvider, error) {
	return nil, errors.New("data streams is not supported for starknet")
}

func (r *relayer) NewFunctionsProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.FunctionsProvider, error) {
	return nil, errors.New("functions are not supported for solana")
}

func (r *relayer) NewAutomationProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.AutomationProvider, error) {
	return nil, errors.New("automation is not supported for starknet")
}

func (r *relayer) NewPluginProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.PluginProvider, error) {
	return nil, errors.New("plugin provider is not supported for starknet")
}

func (r *relayer) NewOCR3CapabilityProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.OCR3CapabilityProvider, error) {
	return nil, errors.New("ocr3 capability provider is not supported for starknet")
}

func (r *relayer) NewCCIPCommitProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.CCIPCommitProvider, error) {
	return nil, errors.New("ccip.commit is not supported for starknet")
}

func (r *relayer) NewCCIPExecProvider(ctx context.Context, rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.CCIPExecProvider, error) {
	return nil, errors.New("ccip.exec is not supported for starknet")
}
