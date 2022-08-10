package starknet

import (
	"crypto/rand"
	"fmt"
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
		fmt.Println(in, out)
		assert.Equal(t, in, out[start:])
	}

}