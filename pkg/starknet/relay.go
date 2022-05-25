package starknet

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/ocr2"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ relaytypes.Relayer = (*relayer)(nil)

type relayer struct {
	chainSet ChainSet
	ctx      context.Context

	lggr logger.Logger

	cancel func()
}

func NewRelayer(lggr logger.Logger, chainSet ChainSet) *relayer {
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
	chainReader, err := r.newChainReader(args)
	if err != nil {
		return nil, err
	}

	configProvider, err := ocr2.NewConfigProvider(chainReader, r.lggr)
	if err != nil {
		return nil, err
	}

	return configProvider, nil
}

func (r *relayer) NewMedianProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.MedianProvider, error) {
	chainReader, err := r.newChainReader(rargs)
	if err != nil {
		return nil, err
	}

	// todo: use pargs for median provider

	medianProvider, err := ocr2.NewMedianProvider(chainReader, r.lggr)
	if err != nil {
		return nil, err
	}

	return medianProvider, nil
}

func (r *relayer) newChainReader(args relaytypes.RelayArgs) (Reader, error) {
	var relayConfig RelayConfig

	err := json.Unmarshal(args.RelayConfig, &relayConfig)
	if err != nil {
		return nil, err
	}

	chain, err := r.chainSet.Chain(r.ctx, relayConfig.ChainID)
	if err != nil {
		return nil, err
	}

	chainReader, err := chain.Reader()
	if err != nil {
		return nil, err
	}

	return chainReader, nil
}
