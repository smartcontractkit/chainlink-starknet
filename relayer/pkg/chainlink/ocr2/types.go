package ocr2

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type ContractConfigDetails struct {
	Block  uint64
	Digest types.ConfigDigest
}

func NewContractConfigDetails(blockNum *big.Int, digest [32]byte) (ccd ContractConfigDetails, err error) {
	return ContractConfigDetails{
		Block:  blockNum.Uint64(),
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

func NewBillingDetails(observationPaymentGJuels *big.Int, transmissionPaymentGJuels *big.Int) (bd BillingDetails, err error) {
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

var feltLowU128Mask = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

// splitFelt mirrors chainlink::utils::split_felt (u128s_from_felt252).
func splitFelt(f *felt.Felt) (high, low *big.Int) {
	n := f.BigInt(new(big.Int))
	low = new(big.Int).And(n, feltLowU128Mask)
	high = new(big.Int).Rsh(new(big.Int).Set(n), 128)
	return high, low
}

func roundIDFromFelt(f *felt.Felt) (uint32, error) {
	_, low := splitFelt(f)
	if !low.IsUint64() {
		return 0, fmt.Errorf("aggregator round id does not fit in a uint64 '%s'", f.String())
	}
	roundID64 := low.Uint64()
	if roundID64 > math.MaxUint32 {
		return 0, fmt.Errorf("aggregator round id does not fit in a uint32 '%s'", f.String())
	}
	return uint32(roundID64), nil
}

func NewRoundData(felts []*felt.Felt) (data RoundData, err error) {
	if len(felts) != 5 {
		return data, fmt.Errorf("expected number of felts to be 5 but got %d", len(felts))
	}
	data.RoundID, err = roundIDFromFelt(felts[0])
	if err != nil {
		return data, err
	}
	data.Answer = felts[1].BigInt(big.NewInt(0))
	blockNumber := felts[2].BigInt(big.NewInt(0))
	if !blockNumber.IsUint64() {
		return data, fmt.Errorf("block number '%s' does not fit into uint64", blockNumber.String())
	}
	data.BlockNumber = blockNumber.Uint64()
	startedAt := felts[3].BigInt(big.NewInt(0))
	if !startedAt.IsInt64() {
		return data, fmt.Errorf("startedAt '%s' does not fit into int64", startedAt.String())
	}
	data.StartedAt = time.Unix(startedAt.Int64(), 0)
	updatedAt := felts[4].BigInt(big.NewInt(0))
	if !updatedAt.IsInt64() {
		return data, fmt.Errorf("updatedAt '%s' does not fit into int64", startedAt.String())
	}
	data.UpdatedAt = time.Unix(updatedAt.Int64(), 0)
	return data, nil
}
