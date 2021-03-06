package chainlink

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ relaytypes.Relayer = (*relayer)(nil)

type relayer struct {
	chainSet starknet.ChainSet
	ctx      context.Context

	lggr logger.Logger

	cancel func()
}

func NewRelayer(lggr logger.Logger, chainSet starknet.ChainSet) *relayer {
	ctx, cancel := context.WithCancel(context.Background())
	return &relayer{
		chainSet: chainSet,
		ctx:      ctx,
		lggr:     lggr,
		cancel:   cancel,
	}
}

func (r *relayer) Start(context.Context) error {
	if r.chainSet == nil {
		return errors.New("chain unavailable")
	}
	return nil
}

func (r *relayer) Close() error {
	r.cancel()
	return nil
}

func (r *relayer) Ready() error {
	return r.chainSet.Ready()
}

func (r *relayer) Healthy() error {
	return r.chainSet.Healthy()
}

func (r *relayer) NewConfigProvider(args relaytypes.RelayArgs) (relaytypes.ConfigProvider, error) {
	var relayConfig RelayConfig

	err := json.Unmarshal(args.RelayConfig, &relayConfig)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't unmarshal RelayConfig")
	}

	chain, err := r.chainSet.Chain(r.ctx, relayConfig.ChainID)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initilize Chain")
	}

	url := "" // TODO: retrieve from reader from nodes/chains config
	configProvider, err := ocr2.NewConfigProvider(relayConfig.ChainID, args.ContractID, url, chain.Config(), r.lggr)
	if err != nil {
		return nil, errors.Wrap(err, "coudln't initialize ConfigProvider")
	}

	return configProvider, nil
}

func (r *relayer) NewMedianProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.MedianProvider, error) {
	var relayConfig RelayConfig

	err := json.Unmarshal(rargs.RelayConfig, &relayConfig)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't unmarshal RelayConfig")
	}

	chain, err := r.chainSet.Chain(r.ctx, relayConfig.ChainID)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initilize Chain")
	}

	// todo: use pargs for median provider
	url := "" // TODO: retrieve from reader from nodes/chains config
	medianProvider, err := ocr2.NewMedianProvider(relayConfig.ChainID, rargs.ContractID, url, chain.Config(), r.lggr)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initilize MedianProvider")
	}

	return medianProvider, nil
}
