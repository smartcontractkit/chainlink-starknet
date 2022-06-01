package ocr2

import (
	"context"
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type ContractConfigDetails struct {
	Block  uint64
	Digest types.ConfigDigest
}

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

type Reader interface {
	OCR2ReadLatestConfigDetails(context.Context, string) (ContractConfigDetails, error)
}

type Config interface {
	// todo: add ocr2 specific config read func
}
