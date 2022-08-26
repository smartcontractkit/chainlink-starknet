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
	return starknet.FeltToSignedBig(&caigotypes.Felt{Int: num})
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

// NOTE: currently unused, could be used by monitoring component
func parseTransmissionEventData(eventData []*caigotypes.Felt) (TransmissionDetails, error) {
	// round_id - skip
	// answer
	index := 1
	latestAnswer := starknet.FeltToSignedBig(eventData[index])

	// transmitter - skip
	// observation_timestamp
	index += 2
	unixTime := eventData[index].Int64()
	latestTimestamp := time.Unix(unixTime, 0)

	// observers - skip
	// observation_len
	index += 2
	observationLen := eventData[index].Int64()

	// observations - skip (based on observationLen)
	// juels_per_fee_coin - skip
	// config digest
	index += int(observationLen) + 2
	digest, err := types.BytesToConfigDigest(starknet.PadBytes(eventData[index].Bytes(), len(types.ConfigDigest{})))
	if err != nil {
		return TransmissionDetails{}, errors.Wrap(err, "couldn't convert bytes to ConfigDigest")
	}

	// epoch_and_round
	index += 1
	epoch, round := parseEpochAndRound(eventData[index].Big())

	// reimbursement - skip

	return TransmissionDetails{
		Digest:          digest,
		Epoch:           epoch,
		Round:           round,
		LatestAnswer:    latestAnswer,
		LatestTimestamp: latestTimestamp,
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
	offchainConfig, err := starknet.DecodeFelts(starknet.CaigoFeltsToJunoFelts(offchainConfigFelts))
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
