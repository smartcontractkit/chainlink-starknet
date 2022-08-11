package ocr2

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/exp/constraints"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	caigo "github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

const chunkSize = 31

// Encodes a byte slice as a bunch of felts. First felt indicates the total byte size.
func EncodeBytes(data []byte) (felts []*big.Int) {
	// prefix with len
	length := big.NewInt(int64(len(data)))
	felts = append(felts, length)

	// chunk every 31 bytes
	for i := 0; i < len(data); i += chunkSize {
		chunk := data[i:Min(i+chunkSize, len(data))]
		// cast to int
		felt := new(big.Int).SetBytes(chunk)
		felts = append(felts, felt)
	}

	return felts
}

func DecodeBytes(felts []*big.Int) ([]byte, error) {
	if len(felts) == 0 {
		return []byte{}, nil
	}

	data := []byte{}
	buf := make([]byte, chunkSize)
	length := int(felts[0].Int64())

	for _, felt := range felts[1:] {
		buf := buf[:Min(chunkSize, length)]

		felt.FillBytes(buf)
		data = append(data, buf...)

		length -= len(buf)
	}

	if length != 0 {
		return nil, errors.New("invalid: contained less bytes than the specified length")
	}

	return data, nil
}

func caigoStringToJunoFelt(str string) (felt junotypes.Felt, err error) {
	big, ok := new(big.Int).SetString(str, 0)
	if !ok {
		return felt, errors.New("wrong format of caigo string")
	}

	return junotypes.BigToFelt(big), nil
}

func caigoFeltsToJunoFelts(cFelts []*caigotypes.Felt) (jFelts []*big.Int) {
	for _, felt := range cFelts {
		jFelts = append(jFelts, felt.Int)
	}

	return jFelts
}

func isEventFromContract(event *caigotypes.Event, address string, eventName string) bool {
	// temp: disable because of address format differences (with/without leading zeros)
	// if event.FromAddress != address {
	// 	return false
	// }

	eventKey := caigo.GetSelectorFromName(eventName)
	// encoded event name guaranteed to be at index 0
	return event.Keys[0].Cmp(eventKey) == 0
}

func parseTransmissionEventData(eventData []*caigotypes.Felt) (TransmissionDetails, error) {
	// round_id - skip
	// answer
	index := 1
	latestAnswer := eventData[index].Big()

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
	var epochAndRound [junotypes.FeltLength]byte
	epochAndRoundFelt := eventData[index].Big()
	epochAndRoundFelt.FillBytes(epochAndRound[:])
	epoch := binary.BigEndian.Uint32(epochAndRound[junotypes.FeltLength-5 : junotypes.FeltLength-1])
	round := epochAndRound[junotypes.FeltLength-1]

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

	// onchain_config
	index += 1
	onchainConfigFelts := eventData[index:(index + int(onchainConfigLen))]
	onchainConfig, err := DecodeBytes(caigoFeltsToJunoFelts(onchainConfigFelts))
	if err != nil {
		return types.ContractConfig{}, errors.Wrap(err, "error in decoding onchainConfig")
	}

	// TODO: remove the below once onchain config is implemented
	temp := median.OnchainConfig{
		Min: big.NewInt(math.MinInt64),
		Max: big.NewInt(math.MaxInt64),
	}
	onchainConfig, err = temp.Encode()
	if err != nil {
		return types.ContractConfig{}, errors.Wrap(err, "err in generating placeholder onchain config")
	}
	// -----------------------------------------

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
	offchainConfig, err := DecodeBytes(caigoFeltsToJunoFelts(offchainConfigFelts))
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
