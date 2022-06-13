package ocr2

import (
	"errors"
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type ContractConfigDetails struct {
	Block  uint64
	Digest types.ConfigDigest
}

func NewContractConfigDetails(blockFelt string, digestFelt string) (ccd ContractConfigDetails, err error) {
	block, ok := new(big.Int).SetString(blockFelt, 0)
	if !ok {
		return ccd, errors.New("wrong format of block")
	}

	digest, ok := new(big.Int).SetString(digestFelt, 0)
	if !ok {
		return ccd, errors.New("wrong format of digest")
	}

	digestBytes, err := types.BytesToConfigDigest(digest.Bytes())
	if err != nil {
		return
	}

	return ContractConfigDetails{
		Block:  block.Uint64(),
		Digest: digestBytes,
	}, nil
}

type ContractConfig struct {
	Config      types.ContractConfig
	ConfigBlock uint64
}

type TransmissionDetails struct {
	digest          types.ConfigDigest
	epoch           uint32
	round           uint8
	latestAnswer    *big.Int
	latestTimestamp time.Time
}

type BillingDetails struct {
	observationPaymentGJuels  uint64
	transmissionPaymentGJuels uint64
}

func NewBillingDetails(observationPaymentFelt string, transmissionPaymentFelt string) (bd BillingDetails, err error) {
	observationPaymentGJuels, ok := new(big.Int).SetString(observationPaymentFelt, 0)
	if !ok {
		return bd, errors.New("wrong format of observationPaymentGJuels")
	}

	transmissionPaymentGJuels, ok := new(big.Int).SetString(transmissionPaymentFelt, 0)
	if !ok {
		return bd, errors.New("wrong format of transmissionPaymentGJuels")
	}

	return BillingDetails{
		observationPaymentGJuels:  observationPaymentGJuels.Uint64(),
		transmissionPaymentGJuels: transmissionPaymentGJuels.Uint64(),
	}, nil
}
