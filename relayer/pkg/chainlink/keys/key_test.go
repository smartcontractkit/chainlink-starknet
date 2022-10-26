package keys

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStarkKey_PublicKeyStr(t *testing.T) {
	// randomly generated private key (for testing only)
	keys := []struct {
		name     string
		priv     string
		contract string
	}{
		{"0", "0x0115fe99e7137319dc20c29511cadfb3650d256d6298f3c411bcfe8730967c51", "0x070cb19fbed32437e6ba7dfc4a957080b2caa2821a15f420b14670f2b8f8d365"},
	}

	for _, k := range keys {
		t.Run(k.name, func(t *testing.T) {
			b, err := hex.DecodeString(k.priv[2:])
			require.NoError(t, err)
			key := Raw(b).Key()

			assert.Equal(t, k.contract, key.AccountAddressStr(), "address calculated from private key does not match expected contract address")
		})
	}
}
