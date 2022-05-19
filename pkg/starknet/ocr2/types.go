package ocr2

import (
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type ContractConfig struct {
	config      types.ContractConfig
	configBlock uint64
}

type TransmissionDetails struct {
	digest          types.ConfigDigest
	epoch           uint32
	round           uint8
	latestAnswer    *big.Int
	latestTimestamp time.Time
}

type OCR2Spec struct {
	// todo: add spec
	ID      int32
	ChainID string
}
