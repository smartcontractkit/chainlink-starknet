package ocr2

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.ContractTransmitter = (*contractTransmitter)(nil)

type contractTransmitter struct {
	*contractReader
	// todo: add params
}

func NewContractTransmitter(
	reader *contractReader,
) *contractTransmitter {
	return &contractTransmitter{
		contractReader: reader,
	}
}

func (c *contractTransmitter) Transmit(
	ctx context.Context,
	reportCtx types.ReportContext,
	report types.Report,
	sigs []types.AttributedOnchainSignature,
) error {
	// todo: implement
	return nil
}

func (c *contractTransmitter) LatestConfigDigestAndEpoch(
	ctx context.Context,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	err error,
) {
	// todo: implement
	return types.ConfigDigest{}, 0, err
}

func (c *contractTransmitter) FromAccount() types.Account {
	// todo: implement
	return ""
}
