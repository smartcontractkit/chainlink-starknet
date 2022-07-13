package ocr2

import (
	"math/big"
	"time"

	"github.com/pkg/errors"

	junotypes "github.com/NethermindEth/juno/pkg/types"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type ContractConfigDetails struct {
	Block  uint64
	Digest types.ConfigDigest
}

func NewContractConfigDetails(blockFelt junotypes.Felt, digestFelt junotypes.Felt) (ccd ContractConfigDetails, err error) {
	block := blockFelt.Big()

	digest, err := types.BytesToConfigDigest(digestFelt.Bytes())
	if err != nil {
		return ccd, errors.Wrap(err, "couldn't decode config digest")
	}

	return ContractConfigDetails{
		Block:  block.Uint64(),
		Digest: digest,
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

func NewBillingDetails(observationPaymentFelt junotypes.Felt, transmissionPaymentFelt junotypes.Felt) (bd BillingDetails, err error) {
	observationPaymentGJuels := observationPaymentFelt.Big()
	transmissionPaymentGJuels := transmissionPaymentFelt.Big()

	return BillingDetails{
		observationPaymentGJuels:  observationPaymentGJuels.Uint64(),
		transmissionPaymentGJuels: transmissionPaymentGJuels.Uint64(),
	}, nil
}
