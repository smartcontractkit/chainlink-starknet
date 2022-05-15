package starknet

import (
	"context"

	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/contract"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/report"

	relaytypes "github.com/smartcontractkit/chainlink/core/services/relay/types"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ relaytypes.OCR2ProviderCtx = (*Ocr2Provider)(nil)

type Ocr2Provider struct {
	offchainConfigDigester contract.OffchainConfigDigester
	reportCodec            report.ReportCodec
	tracker                *contract.ContractTracker
}

func (p Ocr2Provider) Start(ctx context.Context) error {
	return p.tracker.Start(ctx)
}

func (p Ocr2Provider) Close() error {
	return p.tracker.Close()
}

func (p Ocr2Provider) Ready() error {
	return p.tracker.Ready()
}

func (p Ocr2Provider) Healthy() error {
	return p.tracker.Healthy()
}

func (p Ocr2Provider) ContractTransmitter() types.ContractTransmitter {
	return p.tracker
}

func (p Ocr2Provider) ContractConfigTracker() types.ContractConfigTracker {
	return p.tracker
}

func (p Ocr2Provider) OffchainConfigDigester() types.OffchainConfigDigester {
	return p.offchainConfigDigester
}

func (p Ocr2Provider) ReportCodec() median.ReportCodec {
	return p.reportCodec
}

func (p Ocr2Provider) MedianContract() median.MedianContract {
	return p.tracker
}
