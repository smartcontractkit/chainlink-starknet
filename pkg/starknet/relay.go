package starknet

import (
	"context"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/contract"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/logger"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/report"

	relaytypes "github.com/smartcontractkit/chainlink/core/services/relay/types"
)

var _ relaytypes.RelayerCtx = (*Relayer)(nil)

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

func (r *Relayer) NewOCR2Provider(externalJobID uuid.UUID, s interface{}) (relaytypes.OCR2ProviderCtx, error) {
	var provider Ocr2Provider

	spec, ok := s.(contract.OCR2Spec)
	if !ok {
		return &provider, errors.New("unsuccessful cast to OCR2Spec in NewOCR2Provider")
	}

	// todo: insert digester values from spec
	configDigester := contract.OffchainConfigDigester{}

	chain, err := r.chainSet.Chain(r.ctx, spec.ChainID)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing Chain in NewOCR2Provider")
	}

	reader, err := chain.Reader()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing Reader in NewOCR2Provider")
	}

	contractTracker := contract.NewTracker(spec, chain.Config(), reader, r.lggr)

	return &Ocr2Provider{
		offchainConfigDigester: configDigester,
		reportCodec:            report.ReportCodec{},
		tracker:                contractTracker,
	}, nil
}
