package ocr2

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"time"

	"github.com/pkg/errors"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	caigo "github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/medianreport"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

func parseAnswer(str string) (num *big.Int) {
	felt := junotypes.HexToFelt(str)
	num = felt.Big()
	return starknet.FeltToBigInt(&caigotypes.Felt{Int: num})
}

func parseEpochAndRound(felt *big.Int) (epoch uint32, round uint8) {
	var epochAndRound [junotypes.FeltLength]byte
	felt.FillBytes(epochAndRound[:])
	epoch = binary.BigEndian.Uint32(epochAndRound[junotypes.FeltLength-5 : junotypes.FeltLength-1])
	round = epochAndRound[junotypes.FeltLength-1]
	return epoch, round
}

func isEventFromContract(event *caigotypes.Event, address string, eventName string) bool {
	eventKey := caigo.GetSelectorFromName(eventName)
	// encoded event name guaranteed to be at index 0
	return CompareAddress(event.FromAddress, address) && event.Keys[0].Cmp(eventKey) == 0
}

// CompareAddress compares different hex starknet addresses with potentially different 0 padding
func CompareAddress(a, b string) bool {
	aBytes, err := keys.HexToBytes(a)
	if err != nil {
		return false
	}

	bBytes, err := keys.HexToBytes(b)
	if err != nil {
		return false
	}

	return bytes.Compare(starknet.PadBytes(aBytes, 32), starknet.PadBytes(bBytes, 32)) == 0
}

type NewTransmission struct {
	RoundId         uint32
	Epoch           uint32
	Round           uint8
	ConfigDigest    types.ConfigDigest
	LatestAnswer    *big.Int
	LatestTimestamp time.Time
	Transmitter     *caigotypes.Felt
	Observers       []uint8
	Observations    []*big.Int
	JuelsPerFeeCoin *big.Int
	Reimbursement   *big.Int
}

// ParseTransmissionEventData is used by the monitoring component
func ParseTransmissionEventData(eventData []*caigotypes.Felt) (NewTransmission, error) {
	// round_id
	index := 0
	roundId := uint32(eventData[index].Big().Uint64())

	// answer
	index += 1
	latestAnswer := starknet.FeltToBigInt(eventData[index])

	// transmitter
	index += 1
	transmitter := eventData[index]

	// observation_timestamp
	index += 1
	unixTime := eventData[index].Int64()
	latestTimestamp := time.Unix(unixTime, 0)

	// observers - TODO
	index += 1

	// observation_len
	index += 1
	observationLen := eventData[index].Int64()

	// observations - (based on observationLen)
	var observations []*big.Int
	for i := 0; i < int(observationLen); i++ {
		observations = append(observations, eventData[index+1].Big())
	}

	// juels_per_fee_coin
	index += 1 + int(observationLen)
	juelsPerFeeCoin := eventData[index].Big()

	// config digest
	index += 1
	digest, err := types.BytesToConfigDigest(starknet.PadBytes(eventData[index].Bytes(), len(types.ConfigDigest{})))
	if err != nil {
		return NewTransmission{}, errors.Wrap(err, "couldn't convert bytes to ConfigDigest")
	}

	// epoch_and_round
	index += 1
	epoch, round := parseEpochAndRound(eventData[index].Big())

	// reimbursement - skip
	index += 1
	reimbursement := eventData[index].Big()

	return NewTransmission{
		RoundId:         roundId,
		Epoch:           epoch,
		Round:           round,
		ConfigDigest:    digest,
		LatestAnswer:    latestAnswer,
		LatestTimestamp: latestTimestamp,
		Transmitter:     transmitter,
		// Observers: observers,
		Observations:    observations,
		JuelsPerFeeCoin: juelsPerFeeCoin,
		Reimbursement:   reimbursement,
	}, nil
}

func parseConfigEventData(eventData []*caigotypes.Felt) (types.ContractConfig, error) {
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

	//onchain_config length
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
	offchainConfig, err := starknet.DecodeBytes(starknet.CaigoFeltsToJunoFelts(offchainConfigFelts))
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
