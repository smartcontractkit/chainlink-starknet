package contract

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.ContractTransmitter = (*ContractTracker)(nil)

func (c *ContractTracker) Transmit(
	ctx context.Context,
	reportCtx types.ReportContext,
	report types.Report,
	sigs []types.AttributedOnchainSignature,
) error {
	// todo: implement
	return nil
}

func (c *ContractTracker) LatestConfigDigestAndEpoch(
	ctx context.Context,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	err error,
) {
	state, err := c.ReadState()
	return state.Config.LatestConfigDigest, state.Config.Epoch, err
}

func (c *ContractTracker) FromAccount() types.Account {
	// todo: implement
	return ""
}
