package starknet

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestEnsureFelt(t *testing.T) {
	// create random bytes
	random := make([]byte, 32)
	_, err := rand.Read(random)
	require.NoError(t, err)

	// fit into [32]byte
	val := [32]byte{}
	copy(val[:], random[:])

	// validate replace first char with 0
	out := EnsureFelt(val)
	assert.Equal(t, 32, len(out))
	assert.Equal(t, uint8(0), out[0])
	assert.Equal(t, random[1:], out[1:])

	// validate always fills 64 characters
	out = EnsureFelt([32]byte{})
	assert.Equal(t, 32, len(out))
}

func TestHexToSignedBig(t *testing.T) {
	// Positive value (99)
	answer := HexToSignedBig("0x63")
	assert.Equal(t, big.NewInt(99), answer)

	// Negative value (-10)
	answer = HexToSignedBig("0x800000000000010fffffffffffffffffffffffffffffffffffffffffffffff7")
	assert.Equal(t, big.NewInt(-10), answer)
}

func TestDecodeFeltFails(t *testing.T) {
	val, _ := new(big.Int).SetString("1231927389172389172983712738127391273891", 10)

	array := make([]*big.Int, 2)
	array[0] = val
	array[1] = val

	_, error := DecodeFelts(array)
	assert.Equal(t, error.Error(), "invalid: felt array can't be decoded")
}

func TestDecodeFeltsSuccesses(t *testing.T) {
	b, _ := new(big.Int).SetString("1231927389172389172983712738127391273891", 10)
	bBytesLen := int64(len(b.Bytes()))
	a := big.NewInt(bBytesLen)

	array := make([]*big.Int, 2)
	array[0] = a
	array[1] = b

	bytes, error := DecodeFelts(array)
	assert.Equal(t, int64(len(bytes)), bBytesLen)
	require.NoError(t, error)
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
