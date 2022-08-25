package medianreport

import (
	"fmt"
	"math/big"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/libocr/bigbigendian"
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

func (codec OnchainConfigCodec) DecodeBigInt(b []byte) ([]*big.Int, error) {
	if len(b) != length {
		return []*big.Int{}, fmt.Errorf("unexpected length of OnchainConfig, expected %v, got %v", length, len(b))
	}

	configVersion, err := bigbigendian.DeserializeSigned(byteWidth, b[:32])
	if err != nil {
		return []*big.Int{}, fmt.Errorf("unable to decode version: %s", err)
	}
	if OnchainConfigVersion != configVersion.Int64() {
		return []*big.Int{}, fmt.Errorf("unexpected version of OnchainConfig, expected %v, got %v", OnchainConfigVersion, configVersion.Int64())
	}

	min, err := bigbigendian.DeserializeSigned(byteWidth, b[byteWidth:2*byteWidth])
	if err != nil {
		return []*big.Int{}, err
	}
	max, err := bigbigendian.DeserializeSigned(byteWidth, b[2*byteWidth:])
	if err != nil {
		return []*big.Int{}, err
	}

	if !(min.Cmp(max) <= 0) {
		return []*big.Int{}, fmt.Errorf("OnchainConfig min (%v) should not be greater than max(%v)", min, max)
	}

	return []*big.Int{configVersion, min, max}, nil
}

func (codec OnchainConfigCodec) DecodeFelts(b []byte) ([]*big.Int, error) {
	bigInts, err := codec.DecodeBigInt(b)
	if err != nil {
		return []*big.Int{}, err
	}

	// ensure felt: this wraps negative values correctly into felts
	min := new(big.Int).Mod(bigInts[1], caigotypes.MaxFelt.Big())
	max := new(big.Int).Mod(bigInts[2], caigotypes.MaxFelt.Big())
	return []*big.Int{bigInts[0], min, max}, nil
}

func (codec OnchainConfigCodec) Decode(b []byte) (median.OnchainConfig, error) {
	bigInts, err := codec.DecodeBigInt(b)
	if err != nil {
		return median.OnchainConfig{}, err
	}
	return median.OnchainConfig{Min: bigInts[1], Max: bigInts[2]}, nil
}

func (codec OnchainConfigCodec) EncodeBigInt(version, min, max *big.Int) ([]byte, error) {
	if version.Uint64() != OnchainConfigVersion {
		return nil, fmt.Errorf("unexpected version of OnchainConfig, expected %v, got %v", OnchainConfigVersion, version.Int64())
	}

	versionBytes, err := bigbigendian.SerializeSigned(byteWidth, version)
	if err != nil {
		return nil, err
	}
	minBytes, err := bigbigendian.SerializeSigned(byteWidth, min)
	if err != nil {
		return nil, err
	}
	maxBytes, err := bigbigendian.SerializeSigned(byteWidth, max)
	if err != nil {
		return nil, err
	}
	result := []byte{}
	result = append(result, versionBytes...)
	result = append(result, minBytes...)
	result = append(result, maxBytes...)

	return result, nil
}

func (codec OnchainConfigCodec) Encode(c median.OnchainConfig) ([]byte, error) {
	return codec.EncodeBigInt(big.NewInt(OnchainConfigVersion), c.Min, c.Max)
}