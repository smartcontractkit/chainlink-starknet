package ocr2

import (
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.OffchainConfigDigester = (*OffchainConfigDigester)(nil)

type OffchainConfigDigester struct {
	// todo: add params
}

func NewOffchainConfigDigester() OffchainConfigDigester {
	return OffchainConfigDigester{}
}

func (d OffchainConfigDigester) ConfigDigest(cfg types.ContractConfig) (types.ConfigDigest, error) {
	// todo: implement
	return types.ConfigDigest{}, nil
}

func (OffchainConfigDigester) ConfigDigestPrefix() types.ConfigDigestPrefix {
	// todo: implement
	return 0
}
