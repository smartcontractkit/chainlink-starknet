package ocr2

import (
	"context"

	"github.com/smartcontractkit/chainlink-starknet/pkg/relay/chainlink/starknet"
	junorpc "github.com/NethermindEth/juno/pkg/rpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	relaytypes "github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ relaytypes.ConfigProvider = (*configProvider)(nil)

type configProvider struct {
	utils.StartStopOnce

	reader        *contractReader
	contractCache *contractCache
	digester      types.OffchainConfigDigester
	transmitter   types.ContractTransmitter

	lggr logger.Logger
}

func NewConfigProvider(chainID string, contractAddress string, cfg starknet.Config, lggr logger.Logger) (*configProvider, error) {
	chainReader, err := NewClient(chainID, lggr)
	if err != nil {
		return nil, err
	}

	reader := NewContractReader(contractAddress, chainReader, lggr)
	cache := NewContractCache(cfg, reader, lggr)
	digester := NewOffchainConfigDigester(junorpc.ChainID(chainID), junorpc.Address(contractAddress))
	transmitter := NewContractTransmitter(reader)

	return &configProvider{
		reader:        reader,
		contractCache: cache,
		digester:      digester,
		transmitter:   transmitter,
		lggr:          lggr,
	}, nil
}

func (p *configProvider) Start(context.Context) error {
	return p.StartOnce("ConfigProvider", func() error {
		p.lggr.Debugf("Config provider starting")
		return p.contractCache.Start()
	})
}

func (p *configProvider) Close() error {
	return p.StopOnce("ConfigProvider", func() error {
		p.lggr.Debugf("Config provider stopping")
		return p.contractCache.Close()
	})
}

func (p *configProvider) ContractConfigTracker() types.ContractConfigTracker {
	return p.reader
}

func (p *configProvider) OffchainConfigDigester() types.OffchainConfigDigester {
	return p.digester
}

var _ relaytypes.MedianProvider = (*medianProvider)(nil)

type medianProvider struct {
	*configProvider
	transmissionsCache *transmissionsCache
	reportCodec        median.ReportCodec
}

func NewMedianProvider(chainID string, contractAddress string, cfg starknet.Config, lggr logger.Logger) (*medianProvider, error) {
	configProvider, err := NewConfigProvider(chainID, contractAddress, cfg, lggr)
	if err != nil {
		return nil, err
	}

	cache := NewTransmissionsCache(configProvider.reader, configProvider.lggr)

	return &medianProvider{
		configProvider:     configProvider,
		transmissionsCache: cache,
		reportCodec:        reportCodec{},
	}, nil
}

func (p *medianProvider) Start(context.Context) error {
	return p.StartOnce("MedianProvider", func() error {
		p.lggr.Debugf("Median provider starting")
		// starting both cache services here
		// todo: find a better way
		if err := p.configProvider.contractCache.Start(); err != nil {
			return err
		}
		return p.transmissionsCache.Start()
	})
}

func (p *medianProvider) Close() error {
	return p.StopOnce("MedianProvider", func() error {
		p.lggr.Debugf("Median provider stopping")
		// stopping both cache services here
		// todo: find a better way
		if err := p.configProvider.contractCache.Close(); err != nil {
			return err
		}
		return p.transmissionsCache.Close()
	})
}

func (p *medianProvider) ContractTransmitter() types.ContractTransmitter {
	return p.transmitter
}

func (p *medianProvider) ReportCodec() median.ReportCodec {
	return p.reportCodec
}

func (p *medianProvider) MedianContract() median.MedianContract {
	return p.transmissionsCache
}
