package contract

import (
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type ContractConfig struct {
	config      types.ContractConfig
	configBlock uint64
}

type OCR2Spec struct {
	// todo: add spec
	ID      int32
	ChainID string
}
