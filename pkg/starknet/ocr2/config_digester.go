package ocr2

import (
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.OffchainConfigDigester = (*offchainConfigDigester)(nil)

type offchainConfigDigester struct {
	// todo: add params
}

func NewOffchainConfigDigester() offchainConfigDigester {
	return offchainConfigDigester{}
}

func (d offchainConfigDigester) ConfigDigest(cfg types.ContractConfig) (types.ConfigDigest, error) {
	// todo: implement
	return types.ConfigDigest{}, nil
}

func (offchainConfigDigester) ConfigDigestPrefix() types.ConfigDigestPrefix {
	// todo: implement
	return 0
}
