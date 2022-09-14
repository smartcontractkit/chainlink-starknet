package starknet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	"github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"
)

const chunkSize = 31

// padd bytes to specific length
func PadBytes(a []byte, length int) []byte {
	if len(a) < length {
		pad := make([]byte, length-len(a))
		return append(pad, a...)
	}

	// return original if length is >= to specified length
	return a
}

// convert 32 byte to "0" + 31 bytes
func EnsureFelt(b [32]byte) (out []byte) {
	out = make([]byte, 32)
	copy(out[:], b[:])
	out[0] = 0
	return out
}

func NilResultError(funcName string) error {
	return fmt.Errorf("nil result in %s", funcName)
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// EncodeFelts takes a byte slice and splits as bunch of felts. First felt indicates the total byte size.
func EncodeFelts(data []byte) (felts []*big.Int) {
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

// DecodeFelts is the reverse of EncodeFelts
func DecodeFelts(felts []*big.Int) ([]byte, error) {
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

// SignedBigToFelt wraps negative values correctly into felt
func SignedBigToFelt(num *big.Int) *big.Int {
	return new(big.Int).Mod(num, caigotypes.MaxFelt.Big())
}

// FeltToSigned unwraps felt into negative values
func FeltToSignedBig(felt *caigotypes.Felt) (num *big.Int) {
	num = felt.Big()
	prime := caigotypes.MaxFelt.Big()
	half := new(big.Int).Div(prime, big.NewInt(2))
	// if num > PRIME/2, then -PRIME to convert to negative value
	if num.Cmp(half) > 0 {
		return new(big.Int).Sub(num, prime)
	}
	return num
}

func HexToSignedBig(str string) (num *big.Int) {
	felt := junotypes.HexToFelt(str)
	return FeltToSignedBig(&caigotypes.Felt{Int: felt.Big()})
}

func FeltsToBig(in []*caigotypes.Felt) (out []*big.Int) {
	for _, f := range in {
		out = append(out, f.Int)
	}

	return out
}

// StringsToFelt maps felts from 'string' (hex) representation to 'caigo.Felt' representation
func StringsToFelt(in []string) (out []*caigotypes.Felt) {
	for _, f := range in {
		out = append(out, caigotypes.StrToFelt(f))
	}

	return out
}

// CompareAddress compares different hex starknet addresses with potentially different 0 padding
func CompareAddress(a, b string) bool {
	aBytes, err := caigo.HexToBytes(a)
	if err != nil {
		return false
	}

	bBytes, err := caigo.HexToBytes(b)
	if err != nil {
		return false
	}

	return bytes.Compare(PadBytes(aBytes, 32), PadBytes(bBytes, 32)) == 0
}

/* Testing utils - do not use (XXX) outside testing context */

func XXXMustHexDecodeString(data string) []byte {
	bytes, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	return bytes
}
