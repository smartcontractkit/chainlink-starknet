package starknet

import (
	"context"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/report"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
)

var _ relaytypes.Relayer = (*Relayer)(nil)

type Relayer struct {
	chainSet ChainSet
	ctx      context.Context

	lggr logger.Logger

	cancel func()
}

func NewRelayer(lggr logger.Logger, chainSet ChainSet) *Relayer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Relayer{
		chainSet: chainSet,
		ctx:      ctx,
		lggr:     lggr,
		cancel:   cancel,
	}
}

func (r *Relayer) Start(context.Context) error {
	if r.chainSet == nil {
		return errors.New("chain unavailable")
	}
	return nil
}

func (r *Relayer) Close() error {
	r.cancel()
	return nil
}

func (r *Relayer) Ready() error {
	return r.chainSet.Ready()
}

func (r *Relayer) Healthy() error {
	return r.chainSet.Healthy()
}

func (r *Relayer) NewConfigProvider(args relaytypes.RelayArgs) (relaytypes.ConfigProvider, error) {
	configProvider, err := NewConfigProvider(r.ctx, r.lggr, r.chainSet, args)

	if err != nil {
		return nil, err
	}

	return configProvider, nil
}

func (r *Relayer) NewMedianProvider(rargs relaytypes.RelayArgs, pargs relaytypes.PluginArgs) (relaytypes.MedianProvider, error) {
	configProvider, err := NewConfigProvider(r.ctx, r.lggr, r.chainSet, rargs)

	if err != nil {
		return nil, err
	}

	return &MedianProvider{
		ConfigProvider: configProvider,
		reportCodec:    report.ReportCodec{},
	}, nil
}
