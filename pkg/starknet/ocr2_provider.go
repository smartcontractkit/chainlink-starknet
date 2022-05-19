package starknet

import (
	"context"
	"encoding/json"

	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/ocr2"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ relaytypes.ConfigProvider = (*ConfigProvider)(nil)

type ConfigProvider struct {
	utils.StartStopOnce

	chain       Chain
	reader      *ocr2.ContractReader
	cache       *ocr2.ContractCache
	digester    types.OffchainConfigDigester
	transmitter types.ContractTransmitter

	lggr logger.Logger
}

func NewConfigProvider(ctx context.Context, lggr logger.Logger, chainSet ChainSet, args relaytypes.RelayArgs) (*ConfigProvider, error) {
	var relayConfig RelayConfig

	err := json.Unmarshal(args.RelayConfig, &relayConfig)
	if err != nil {
		return nil, err
	}

	chain, err := chainSet.Chain(ctx, relayConfig.ChainID)
	if err != nil {
		return nil, err
	}

	chainReader, err := chain.Reader()
	if err != nil {
		return nil, err
	}

	reader := ocr2.NewContractReader(chainReader, lggr)
	cache := ocr2.NewContractCache(reader, lggr)
	digester := ocr2.NewOffchainConfigDigester()
	transmitter := ocr2.NewContractTransmitter(reader)

	return &ConfigProvider{
		chain:       chain,
		reader:      reader,
		cache:       cache,
		digester:    digester,
		transmitter: transmitter,
		lggr:        lggr,
	}, nil
}

func (p *ConfigProvider) Start(context.Context) error {
	return p.StartOnce("Relay", func() error {
		p.lggr.Debugf("Relay starting")
		return p.cache.Start()
	})
}

func (p *ConfigProvider) Close() error {
	return p.StopOnce("Relay", func() error {
		p.lggr.Debugf("Relay stopping")
		return p.cache.Close()
	})
}

func (p *ConfigProvider) ContractConfigTracker() types.ContractConfigTracker {
	return p.reader
}

func (p *ConfigProvider) OffchainConfigDigester() types.OffchainConfigDigester {
	return p.digester
}

var _ relaytypes.MedianProvider = (*MedianProvider)(nil)

type MedianProvider struct {
	*ConfigProvider
	reportCodec median.ReportCodec
}

func (p *MedianProvider) ContractTransmitter() types.ContractTransmitter {
	return p.transmitter
}

func (p *MedianProvider) ReportCodec() median.ReportCodec {
	return p.reportCodec
}

func (p *MedianProvider) MedianContract() median.MedianContract {
	return p.cache
}
