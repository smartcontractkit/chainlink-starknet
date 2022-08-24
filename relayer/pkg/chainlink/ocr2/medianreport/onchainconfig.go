package medianreport

import (
	"fmt"
	"math/big"

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

func (codec OnchainConfigCodec) Decode(b []byte) (median.OnchainConfig, error) {
	bigInts, err := codec.DecodeBigInt(b)
	if err != nil {
		return median.OnchainConfig{}, err
	}
	return median.OnchainConfig{Min: bigInts[1], Max: bigInts[2]}, nil
}

func (codec OnchainConfigCodec) Encode(c median.OnchainConfig) ([]byte, error) {
	versionBytes, err := bigbigendian.SerializeSigned(byteWidth, big.NewInt(OnchainConfigVersion))
	if err != nil {
		return nil, err
	}
	minBytes, err := bigbigendian.SerializeSigned(byteWidth, c.Min)
	if err != nil {
		return nil, err
	}
	maxBytes, err := bigbigendian.SerializeSigned(byteWidth, c.Max)
	if err != nil {
		return nil, err
	}
	result := []byte{}
	result = append(result, versionBytes...)
	result = append(result, minBytes...)
	result = append(result, maxBytes...)

	return result, nil
}