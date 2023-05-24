package keys

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	s := &signature{
		x: big.NewInt(7),
		y: big.NewInt(11),
	}

	b, err := s.bytes()
	require.NoError(t, err)
	require.NotNil(t, b)
	require.Len(t, b, signatureLen)

	roundTrip, err := signatureFromBytes(b)
	require.NoError(t, err)
	require.Equal(t, s.x, roundTrip.x)
	require.Equal(t, s.y, roundTrip.y)
}
