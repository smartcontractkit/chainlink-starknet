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
	rand.Read(random)

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
