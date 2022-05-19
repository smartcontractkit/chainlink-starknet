package ocr2

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ relaytypes.ConfigProvider = (*ConfigProvider)(nil)

type ConfigProvider struct {
	utils.StartStopOnce

	reader      *ContractReader
	cache       *ContractCache
	digester    types.OffchainConfigDigester
	transmitter types.ContractTransmitter

	lggr logger.Logger
}

func NewConfigProvider(chainReader Reader, lggr logger.Logger) (*ConfigProvider, error) {
	reader := NewContractReader(chainReader, lggr)
	cache := NewContractCache(reader, lggr)
	digester := NewOffchainConfigDigester()
	transmitter := NewContractTransmitter(reader)

	return &ConfigProvider{
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

func NewMedianProvider(configProvider *ConfigProvider) (*MedianProvider, error) {
	return &MedianProvider{
		ConfigProvider: configProvider,
		reportCodec:    ReportCodec{},
	}, nil
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
