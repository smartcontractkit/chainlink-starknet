package keys

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStarkKey_PublicKeyStr(t *testing.T) {
	// randomly generated private key (for testing only)
	privKey := "0x04c4bfa272a15b2f00f36a72af9db41701e14d40b14621aa60ce3e34cfd2428b"
	contractAddr := "0x001b1ae54cd4d718b7ae4b689ffc52d260061c6ba9304bbb229b28b1aeb410da"

	privBytes, err := hex.DecodeString(privKey[2:])
	require.NoError(t, err)

	key := Raw(privBytes).Key()
	assert.Equal(t, contractAddr, key.PublicKeyStr(), "address calculated from private key does not match expected contract address")
}
