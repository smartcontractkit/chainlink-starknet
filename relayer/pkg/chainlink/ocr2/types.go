package ocr2

import (
	"encoding/binary"
	"fmt"
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

type RoundData struct {
	RoundID     uint32
	Answer      *big.Int
	BlockNumber uint64
	StartedAt   time.Time
	UpdatedAt   time.Time
}

func NewRoundData(felts []junotypes.Felt) (data RoundData, err error) {
	if len(felts) < 5 {
		return data, fmt.Errorf("expected number of felts to be 5 but got %d", len(felts))
	}
	data.RoundID = binary.BigEndian.Uint32(felts[0][:])
	data.Answer = felts[1].Big()
	blockNumber := felts[2].Big()
	if !blockNumber.IsUint64() {
		return data, fmt.Errorf("block number '%s' does not fit into uint64", blockNumber.String())
	}
	data.BlockNumber = blockNumber.Uint64()
	startedAt := felts[3].Big()
	if !startedAt.IsInt64() {
		return data, fmt.Errorf("startedAt '%s' does not fit into int64", startedAt.String())
	}
	data.StartedAt = time.Unix(startedAt.Int64(), 0)
	updatedAt := felts[3].Big()
	if !updatedAt.IsInt64() {
		return data, fmt.Errorf("updatedAt '%s' does not fit into int64", startedAt.String())
	}
	data.StartedAt = time.Unix(startedAt.Int64(), 0)
	return data, nil
}
