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

	reader        *ContractReader
	contractCache *ContractCache
	digester      types.OffchainConfigDigester
	transmitter   types.ContractTransmitter

	lggr logger.Logger
}

func NewConfigProvider(chainReader Reader, lggr logger.Logger) (*ConfigProvider, error) {
	reader := NewContractReader(chainReader, lggr)
	cache := NewContractCache(reader, lggr)
	digester := NewOffchainConfigDigester()
	transmitter := NewContractTransmitter(reader)

	return &ConfigProvider{
		reader:        reader,
		contractCache: cache,
		digester:      digester,
		transmitter:   transmitter,
		lggr:          lggr,
	}, nil
}

func (p *ConfigProvider) Start(context.Context) error {
	return p.StartOnce("ConfigProvider", func() error {
		p.lggr.Debugf("Config provider starting")
		return p.contractCache.Start()
	})
}

func (p *ConfigProvider) Close() error {
	return p.StopOnce("ConfigProvider", func() error {
		p.lggr.Debugf("Config provider stopping")
		return p.contractCache.Close()
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
	transmissionsCache *TransmissionsCache
	reportCodec        median.ReportCodec
}

func NewMedianProvider(configProvider *ConfigProvider) (*MedianProvider, error) {
	cache := NewTransmissionsCache(configProvider.reader, configProvider.lggr)

	return &MedianProvider{
		ConfigProvider:     configProvider,
		transmissionsCache: cache,
		reportCodec:        ReportCodec{},
	}, nil
}

func (p *MedianProvider) Start(context.Context) error {
	return p.StartOnce("MedianProvider", func() error {
		p.lggr.Debugf("Median provider starting")
		return p.transmissionsCache.Start()
	})
}

func (p *MedianProvider) Close() error {
	return p.StopOnce("MedianProvider", func() error {
		p.lggr.Debugf("Median provider stopping")
		return p.transmissionsCache.Close()
	})
}

func (p *MedianProvider) ContractTransmitter() types.ContractTransmitter {
	return p.transmitter
}

func (p *MedianProvider) ReportCodec() median.ReportCodec {
	return p.reportCodec
}

func (p *MedianProvider) MedianContract() median.MedianContract {
	return p.transmissionsCache
}
