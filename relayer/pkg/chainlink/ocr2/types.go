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
	Digest          types.ConfigDigest
	Epoch           uint32
	Round           uint8
	LatestAnswer    *big.Int
	LatestTimestamp time.Time
}

type BillingDetails struct {
	ObservationPaymentGJuels  uint64
	TransmissionPaymentGJuels uint64
}

func NewBillingDetails(observationPaymentFelt junotypes.Felt, transmissionPaymentFelt junotypes.Felt) (bd BillingDetails, err error) {
	observationPaymentGJuels := observationPaymentFelt.Big()
	transmissionPaymentGJuels := transmissionPaymentFelt.Big()

	return BillingDetails{
		ObservationPaymentGJuels:  observationPaymentGJuels.Uint64(),
		TransmissionPaymentGJuels: transmissionPaymentGJuels.Uint64(),
	}, nil
}
