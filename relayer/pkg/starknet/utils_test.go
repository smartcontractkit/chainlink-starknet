package starknet

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	caigotypes "github.com/smartcontractkit/caigo/types"
)

var (
	lengths = []int{1, 31, 32, 40}
	length  = 32
)

func TestPadBytes(t *testing.T) {
	for i, v := range lengths {

		// generate random
		in := make([]byte, v)
		_, err := rand.Read(in)
		require.NoError(t, err)
		require.Equal(t, lengths[i], len(in))

		out := PadBytes(in, length)
		expectLen := length
		if v > expectLen {
			expectLen = v
		}
		assert.Equal(t, expectLen, len(out))

		start := 0
		if v < length {
			start = length - v
		}
		assert.Equal(t, in, out[start:])
	}
}

func TestFeltToUnsignedBig(t *testing.T) {

	negativeBig := caigotypes.BigToFelt(big.NewInt(-100))
	assert.False(t, negativeBig.IsUint64())

	// negative felts are not supported, so it'll be interpreted as non-negative
	num, err := FeltToUnsignedBig(negativeBig)
	assert.NoError(t, err)
	assert.Equal(t, num, big.NewInt(100))

	positiveBig := caigotypes.BigToFelt(big.NewInt(100))
	num, err = FeltToUnsignedBig(positiveBig)
	assert.NoError(t, err)
	assert.Equal(t, num, big.NewInt(100))

}

func TestHexToUnSignedBig(t *testing.T) {
	// Positive value (99)
	answer, err := HexToUnsignedBig("0x63")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(99), answer)
}

func TestDecodeFeltFails(t *testing.T) {
	val, _ := new(big.Int).SetString("1231927389172389172983712738127391273891", 10)

	array := make([]*big.Int, 2)
	array[0] = val
	array[1] = val

	_, err := DecodeFelts(array)
	assert.Equal(t, err.Error(), "invalid: contained less bytes than the specified length")
}

func TestDecodeFeltsSuccesses(t *testing.T) {
	b, _ := new(big.Int).SetString("1231927389172389172983712738127391273891", 10)
	bBytesLen := int64(len(b.Bytes()))
	a := big.NewInt(bBytesLen)

	array := make([]*big.Int, 2)
	array[0] = a
	array[1] = b

	fBytes, err := DecodeFelts(array)
	assert.Equal(t, int64(len(fBytes)), bBytesLen)
	require.NoError(t, err)
}

func TestEncodeThenDecode(t *testing.T) {
	const dataLen = 232

	// create random bytes
	data := make([]byte, dataLen)
	_, err := rand.Read(data)
	require.NoError(t, err)

	encodedData := EncodeFelts(data)
	decodedData, err := DecodeFelts(encodedData)
	require.NoError(t, err)

	assert.Equal(t, data, decodedData)
}

func TestDecodeThenEncode(t *testing.T) {
	const numOfFelts = 12

	// Create random array of felts
	felts := make([]*big.Int, numOfFelts+1)
	felts[0] = big.NewInt(numOfFelts * 31)
	for i := 1; i <= numOfFelts; i++ {
		feltRaw := make([]byte, 31)
		_, err := rand.Read(feltRaw)
		require.NoError(t, err)

		felt := new(big.Int).SetBytes(feltRaw)
		felts[i] = felt
	}

	decodedFelts, err := DecodeFelts(felts)
	require.NoError(t, err)

	encodedFelts := EncodeFelts(decodedFelts)
	assert.Equal(t, encodedFelts, felts)
}

func TestIncorrectDecode(t *testing.T) {
	felts := make([]*big.Int, 3)
	felts[0] = big.NewInt(35)
	// bytes length can be less than 31
	felts[1] = new(big.Int).SetBytes([]byte{0x1, 0x2, 0x3, 0x4})
	felts[2] = new(big.Int).SetBytes([]byte{0x11, 0x21})

	_, err := DecodeFelts(felts)

	require.NoError(t, err)
}

func TestDecodeFelts(t *testing.T) {
	a, _ := new(big.Int).SetString("1231927389172389172983712738127391273891", 10)
	array := make([]*big.Int, 2)
	array[0] = a
	array[1] = a
	_, err := DecodeFelts(array)
	require.Error(t, err)
}

func TestDecodeFelts2(t *testing.T) {
	a, _ := new(big.Int).SetString("1", 10)
	b, _ := new(big.Int).SetString("1231927389172389172983712738127391273891", 10)
	array := make([]*big.Int, 2)
	array[0] = a
	array[1] = b
	_, err := DecodeFelts(array)
	require.Error(t, err)
}
