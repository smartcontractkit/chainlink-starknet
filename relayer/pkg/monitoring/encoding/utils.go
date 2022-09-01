package encoding

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/NethermindEth/juno/pkg/common"
	junotypes "github.com/NethermindEth/juno/pkg/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/medianreport"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

func DecodeBigInt(result string) (*big.Int, error) {
	felt := junotypes.HexToFelt(result)
	return felt.Big(), nil
}

func DecodeUint64(result string) (uint64, error) {
	bigNum, _ := DecodeBigInt(result)
	if !bigNum.IsUint64() {
		return 0, fmt.Errorf("doesn't fit in a uint64: %s", bigNum.String())
	}
	return bigNum.Uint64(), nil
}

func DecodeInt64(result string) (int64, error) {
	bigNum, _ := DecodeBigInt(result)
	if !bigNum.IsInt64() {
		return 0, fmt.Errorf("doesn't fit in a int64: %s", bigNum.String())
	}
	return bigNum.Int64(), nil
}

func DecodeUint8(result string) (uint8, error) {
	num, err := DecodeUint64(result)
	if err != nil {
		return 0, err
	}
	if num >= 256 {
		return 0, fmt.Errorf("number '%d' is too big to fit in a uint8", num)
	}
	return uint8(num), nil
}

func DecodeUint32(result string) (uint32, error) {
	buf := common.FromHex(result)
	return binary.BigEndian.Uint32(buf), nil
}

func DecodeConfigDigest(result string) (types.ConfigDigest, error) {
	felt := junotypes.HexToFelt(result)
	padded := starknet.PadBytes(felt.Bytes(), len(types.ConfigDigest{}))
	return types.BytesToConfigDigest(padded)
}

func DecodeEpochAndRound(result string) (epoch uint32, round uint8, err error) {
	bigNum, _ := DecodeBigInt(result)
	var epochAndRound [junotypes.FeltLength]byte
	bigNum.FillBytes(epochAndRound[:])
	epoch = binary.BigEndian.Uint32(epochAndRound[junotypes.FeltLength-5 : junotypes.FeltLength-1])
	round = epochAndRound[junotypes.FeltLength-1]
	return epoch, round, nil
}

func DecodeTime(result string) (time.Time, error) {
	timestamp, err := DecodeInt64(result)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to decode time.Time from '%s': %w", result, err)
	}
	return time.Unix(timestamp, 0), nil
}

func DecodeAccount(result string) (types.Account, error) {
	felt := junotypes.HexToFelt(result)
	padded := starknet.PadBytes(felt.Bytes(), 32)
	encoded := hex.EncodeToString(padded)
	return types.Account("0x" + encoded), nil
}

func DecodeObservations(results []string, lenFieldIndex int) ([]*big.Int, error) {
	numObservations, err := DecodeInt64(results[lenFieldIndex])
	if err != nil {
		return nil, fmt.Errorf("unable to decode the number of observations from '%s': %w", results[lenFieldIndex], err)
	}
	observations := make([]*big.Int, numObservations)
	for i := 0; i < int(numObservations); i++ {
		observations[i], _ = DecodeBigInt(results[lenFieldIndex+i+1])
	}
	return observations, nil
}

func DecodeSigner(result string) (types.OnchainPublicKey, error) {
	felt := junotypes.HexToFelt(result)
	padded := starknet.PadBytes(felt.Bytes(), 32)
	return types.OnchainPublicKey(padded), nil
}

func DecodeOracles(results []string, lenFieldIndex int) ([]types.OnchainPublicKey, []types.Account, error) {
	oraclesLen, err := DecodeUint64(results[lenFieldIndex])
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode OraclesLen from '%s': %w", results[lenFieldIndex], err)
	}
	if len(results) < lenFieldIndex+2*int(oraclesLen) {
		return nil, nil, fmt.Errorf("insufficient fields in the response, expected %d but got %d", lenFieldIndex+2*int(oraclesLen), len(results))
	}
	signers := []types.OnchainPublicKey{}
	transmitters := []types.Account{}
	var i uint64
	for i = 0; i < oraclesLen; i += 2 {
		signer, err := DecodeSigner(results[i])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode signer at position '%d': %w", i, err)
		}
		signers = append(signers, signer)
		transmitter, err := DecodeAccount(results[i+1])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode account at position '%d': %w", i+1, err)
		}
		transmitters = append(transmitters, transmitter)
	}
	return signers, transmitters, nil
}

func DecodeOnchainConfig(results []string, lenFieldIndex int) ([]byte, uint64, error) {
	configLen, err := DecodeUint64(results[lenFieldIndex])
	if err != nil {
		return nil, 0, fmt.Errorf("unable to decode OnchainConfigLen from '%s': %w", results[lenFieldIndex], err)
	}
	onchainConfig, err := medianreport.OnchainConfigCodec{}.EncodeFromFelt(
		junotypes.HexToFelt(results[lenFieldIndex+1]).Big(),
		junotypes.HexToFelt(results[lenFieldIndex+2]).Big(),
		junotypes.HexToFelt(results[lenFieldIndex+3]).Big(),
	)
	if err != nil {
		return nil, 0, errors.Wrap(err, "err in encoding onchain config from felts")
	}
	return onchainConfig, configLen, nil
}

func DecodeOffchainConfig(results []string, lenFieldIndex int) ([]byte, uint64, error) {
	configLen, err := DecodeUint64(results[lenFieldIndex])
	if err != nil {
		return nil, 0, fmt.Errorf("unable to decode OffchainConfigLen from '%s': %w", results[lenFieldIndex], err)
	}
	caigoFelts := []junotypes.Felt{}
	for i := lenFieldIndex + 1; i < lenFieldIndex+int(configLen)+1; i++ {
		caigoFelts = append(caigoFelts, junotypes.HexToFelt(results[i]))
	}
	offchainConfig, err := starknet.DecodeBytes(toBigNums(caigoFelts))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode offchain config: %w", err)
	}
	return offchainConfig, configLen, nil
}

func toBigNums(felts []junotypes.Felt) []*big.Int {
	output := []*big.Int{}
	for _, felt := range felts {
		output = append(output, felt.Big())
	}
	return output
}
