package medianreport

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
)

const (
	OnchainConfigVersion = 1
	byteWidth            = 32
	length               = 3 * byteWidth
)

// report format
// 32 bytes - version
// 32 bytes - min
// 32 bytes - max

type OnchainConfigCodec struct{}

var _ median.OnchainConfigCodec = &OnchainConfigCodec{}

// DecodeToFelts decodes the onchainconfig into felt values (used in config digest hashing)
func (codec OnchainConfigCodec) DecodeToFelts(b []byte) ([]*big.Int, error) {
	if len(b) != length {
		return []*big.Int{}, fmt.Errorf("unexpected length of OnchainConfig, expected %v, got %v", length, len(b))
	}

	// convert from bytes to *big.Int
	configVersion := new(big.Int).SetBytes(b[:32])

	if OnchainConfigVersion != configVersion.Int64() {
		return []*big.Int{}, fmt.Errorf("unexpected version of OnchainConfig, expected %v, got %v", OnchainConfigVersion, configVersion.Int64())
	}

	minVal := new(big.Int).SetBytes(b[byteWidth : 2*byteWidth])
	maxVal := new(big.Int).SetBytes(b[2*byteWidth:])

	return []*big.Int{configVersion, minVal, maxVal}, nil
}

// Decode converts the onchainconfig via the outputs of DecodeToFelts into unsigned big.Ints that libocr expects
func (codec OnchainConfigCodec) Decode(ctx context.Context, b []byte) (median.OnchainConfig, error) {
	felts, err := codec.DecodeToFelts(b)
	if err != nil {
		return median.OnchainConfig{}, err
	}

	minVal := felts[1]
	maxVal := felts[2]

	if !(minVal.Cmp(maxVal) <= 0) {
		return median.OnchainConfig{}, fmt.Errorf("OnchainConfig min (%v) should not be greater than max(%v)", minVal, maxVal)
	}

	return median.OnchainConfig{Min: minVal, Max: maxVal}, nil
}

// EncodeFromFelt encodes the config where min & max are big.Int representations of a felt
// Cairo has no notion of signed values: min and max values will be non-negative
func (codec OnchainConfigCodec) EncodeFromFelt(version, minVal, maxVal *big.Int) ([]byte, error) {
	if version.Uint64() != OnchainConfigVersion {
		return nil, fmt.Errorf("unexpected version of OnchainConfig, expected %v, got %v", OnchainConfigVersion, version.Int64())
	}

	if minVal.Sign() == -1 || maxVal.Sign() == -1 {
		return nil, fmt.Errorf("starknet does not support negative values: min = (%v) and max = (%v)", minVal, maxVal)
	}

	result := []byte{}
	versionBytes := make([]byte, byteWidth)
	minBytes := make([]byte, byteWidth)
	maxBytes := make([]byte, byteWidth)
	result = append(result, version.FillBytes(versionBytes)...)
	result = append(result, minVal.FillBytes(minBytes)...)
	result = append(result, maxVal.FillBytes(maxBytes)...)

	return result, nil
}

// Encode takes the interface that libocr uses (big.Ints) and serializes it into 3 felts
func (codec OnchainConfigCodec) Encode(ctx context.Context, c median.OnchainConfig) ([]byte, error) {
	return codec.EncodeFromFelt(big.NewInt(OnchainConfigVersion), c.Min, c.Max)
}
