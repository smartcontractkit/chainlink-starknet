package ocr2

import (
	"context"

	"github.com/pkg/errors"

	junorpc "github.com/NethermindEth/juno/pkg/rpc"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/config"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/medianreport"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

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

	lggr logger.Logger
}

func NewConfigProvider(chainID string, contractAddress string, basereader starknet.Reader, cfg config.Config, lggr logger.Logger) (*configProvider, error) {
	chainReader, err := NewClient(basereader, lggr)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initialize chain client")
	}

	reader := NewContractReader(contractAddress, chainReader, lggr)
	cache := NewContractCache(cfg, reader, lggr)
	digester := NewOffchainConfigDigester(junorpc.ChainID(chainID), junorpc.Address(contractAddress))

	return &configProvider{
		reader:        reader,
		contractCache: cache,
		digester:      digester,
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
	return p.contractCache
}

func (p *configProvider) OffchainConfigDigester() types.OffchainConfigDigester {
	return p.digester
}

var _ relaytypes.MedianProvider = (*medianProvider)(nil)

type medianProvider struct {
	*configProvider
	transmitter        types.ContractTransmitter
	transmissionsCache *transmissionsCache
	reportCodec        median.ReportCodec
}

func NewMedianProvider(chainID string, contractAddress string, senderAddress string, basereader starknet.Reader, cfg config.Config, lggr logger.Logger) (*medianProvider, error) {
	configProvider, err := NewConfigProvider(chainID, contractAddress, basereader, cfg, lggr)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initialize ConfigProvider")
	}

	cache := NewTransmissionsCache(cfg, configProvider.reader, configProvider.lggr)
	transmitter := NewContractTransmitter(cache, contractAddress, senderAddress)

	return &medianProvider{
		configProvider:     configProvider,
		transmitter:        transmitter,
		transmissionsCache: cache,
		reportCodec:        medianreport.ReportCodec{},
	}, nil
}

func (p *medianProvider) Start(context.Context) error {
	return p.StartOnce("MedianProvider", func() error {
		p.lggr.Debugf("Median provider starting")
		// starting both cache services here
		// todo: find a better way
		if err := p.configProvider.contractCache.Start(); err != nil {
			return errors.Wrap(err, "couldn't start contractCache")
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
			return errors.Wrap(err, "coulnd't stop contractCache")
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
