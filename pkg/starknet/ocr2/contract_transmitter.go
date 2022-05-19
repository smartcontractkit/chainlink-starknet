package ocr2

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.ContractTransmitter = (*ContractTransmitter)(nil)

type ContractTransmitter struct {
	*ContractReader
	// todo: add params
}

func NewContractTransmitter(
	reader *ContractReader,
) *ContractTransmitter {
	return &ContractTransmitter{
		ContractReader: reader,
	}
}

func (c *ContractTransmitter) Transmit(
	ctx context.Context,
	reportCtx types.ReportContext,
	report types.Report,
	sigs []types.AttributedOnchainSignature,
) error {
	// todo: implement
	return nil
}

func (c *ContractTransmitter) LatestConfigDigestAndEpoch(
	ctx context.Context,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	err error,
) {
	// todo: implement
	return types.ConfigDigest{}, 0, err
}

func (c *ContractTransmitter) FromAccount() types.Account {
	// todo: implement
	return ""
}
