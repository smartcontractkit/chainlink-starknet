package keys_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	k, err := keys.GenerateKey(rand.Reader)
	require.NoError(t, err)

	x := big.NewInt(123456)

	signature, err := keys.Sign(x, k)
	require.NoError(t, err)
	require.Len(t, signature, 3*32)
	require.Equal(t, []byte(keys.PubKeyToStarkKey(k.PublicKey())), signature[:32])
}
