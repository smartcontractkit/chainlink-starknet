package ocr2

import (
	"encoding/hex"
	"math/big"
	"time"

	"github.com/pkg/errors"

	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/medianreport"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var (
	MaxObservers = 31
)

// NewTransmissionEvent represents the 'new_transmission' event
type NewTransmissionEvent struct {
	RoundId         uint32
	LatestAnswer    *big.Int
	Transmitter     *caigotypes.Felt
	LatestTimestamp time.Time
	Observers       []uint8
	ObservationsLen uint32
	Observations    []*big.Int
	JuelsPerFeeCoin *big.Int
	ConfigDigest    types.ConfigDigest
	Epoch           uint32
	Round           uint8
	Reimbursement   *big.Int
}

// ParseNewTransmissionEvent is decoding binary felt data as the NewTransmissionEvent type
func ParseNewTransmissionEvent(eventData []*caigotypes.Felt) (NewTransmissionEvent, error) {
	// round_id
	index := 0
	roundId := uint32(eventData[index].Uint64())

	// answer
	index += 1
	latestAnswer := starknet.HexToSignedBig(eventData[index].String())

	// transmitter
	index += 1
	transmitter := eventData[index]

	// observation_timestamp
	index += 1
	unixTime := eventData[index].Int64()
	latestTimestamp := time.Unix(unixTime, 0)

	// observers (raw) max 31
	index += 1
	observersRaw := starknet.PadBytes(eventData[index].Big().Bytes(), MaxObservers)

	// observation_len
	index += 1
	observationsLen := uint32(eventData[index].Uint64())

	// observers (based on observationsLen)
	var observers []uint8
	for i := 0; i < int(observationsLen); i++ {
		observers = append(observers, observersRaw[i])
	}

	// observations (based on observationsLen)
	var observations []*big.Int
	for i := 0; i < int(observationsLen); i++ {
		observations = append(observations, eventData[index+i+1].Big())
	}

	// juels_per_fee_coin
	index += int(observationsLen) + 1
	juelsPerFeeCoin := eventData[index].Big()

	// config digest
	index += 1
	digest, err := types.BytesToConfigDigest(starknet.PadBytes(eventData[index].Bytes(), len(types.ConfigDigest{})))
	if err != nil {
		return NewTransmissionEvent{}, errors.Wrap(err, "couldn't convert bytes to ConfigDigest")
	}

	// epoch_and_round
	index += 1
	epoch, round := parseEpochAndRound(eventData[index].Big())

	// reimbursement
	index += 1
	reimbursement := eventData[index].Big()

	return NewTransmissionEvent{
		RoundId:         roundId,
		LatestAnswer:    latestAnswer,
		Transmitter:     transmitter,
		LatestTimestamp: latestTimestamp,
		Observers:       observers,
		ObservationsLen: observationsLen,
		Observations:    observations,
		JuelsPerFeeCoin: juelsPerFeeCoin,
		ConfigDigest:    digest,
		Epoch:           epoch,
		Round:           round,
		Reimbursement:   reimbursement,
	}, nil
}

// ParseConfigSetEvent is decoding binary felt data as the libocr ContractConfig type
func ParseConfigSetEvent(eventData []*caigotypes.Felt) (types.ContractConfig, error) {
	index := 0
	// previous_config_block_number - skip

	// latest_config_digest
	index += 1
	digest, err := types.BytesToConfigDigest(starknet.PadBytes(eventData[index].Bytes(), len(types.ConfigDigest{})))
	if err != nil {
		return types.ContractConfig{}, errors.Wrap(err, "couldn't convert bytes to ConfigDigest")
	}

	// config_count
	index += 1
	configCount := eventData[index].Uint64()

	// oracles_len
	index += 1
	oraclesLen := eventData[index].Uint64()

	// oracles
	index += 1
	oracleMembers := eventData[index:(index + int(oraclesLen)*2)]
	var signers []types.OnchainPublicKey
	var transmitters []types.Account
	for i, member := range oracleMembers {
		if i%2 == 0 {
			signers = append(signers, starknet.PadBytes(member.Bytes(), 32)) // pad to 32 bytes
		} else {
			transmitters = append(transmitters, types.Account("0x"+hex.EncodeToString(starknet.PadBytes(member.Bytes(), 32)))) // pad to 32 byte length then re-encode
		}
	}

	// f
	index = index + int(oraclesLen)*2
	f := eventData[index].Uint64()

	// onchain_config length
	index += 1
	onchainConfigLen := eventData[index].Uint64()

	// onchain_config (version=1, min, max)
	index += 1
	onchainConfigFelts := eventData[index:(index + int(onchainConfigLen))]
	onchainConfig, err := medianreport.OnchainConfigCodec{}.EncodeFromFelt(
		onchainConfigFelts[0].Big(),
		onchainConfigFelts[1].Big(),
		onchainConfigFelts[2].Big(),
	)
	if err != nil {
		return types.ContractConfig{}, errors.Wrap(err, "err in encoding onchain config from felts")
	}

	// offchain_config_version
	index += int(onchainConfigLen)
	offchainConfigVersion := eventData[index].Uint64()

	// offchain_config_len
	index += 1
	offchainConfigLen := eventData[index].Uint64()

	// offchain_config
	index += 1
	offchainConfigFelts := eventData[index:(index + int(offchainConfigLen))]
	// todo: get rid of caigoToJuno workaround
	offchainConfig, err := starknet.DecodeFelts(starknet.FeltsToBig(offchainConfigFelts))
	if err != nil {
		return types.ContractConfig{}, errors.Wrap(err, "couldn't decode offchain config")
	}

	return types.ContractConfig{
		ConfigDigest:          digest,
		ConfigCount:           configCount,
		Signers:               signers,
		Transmitters:          transmitters,
		F:                     uint8(f),
		OnchainConfig:         onchainConfig,
		OffchainConfigVersion: offchainConfigVersion,
		OffchainConfig:        offchainConfig,
	}, nil
}
