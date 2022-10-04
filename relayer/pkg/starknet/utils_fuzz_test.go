//go:build go1.18
// +build go1.18

package starknet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzEncodeDecodeFelts(f *testing.F) {
	f.Add([]byte{1, 2, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		encodedData := EncodeFelts(data)
		decodedData, err := DecodeFelts(encodedData)
		require.NoError(t, err)

		require.Equal(t, data, decodedData)
	})
}
