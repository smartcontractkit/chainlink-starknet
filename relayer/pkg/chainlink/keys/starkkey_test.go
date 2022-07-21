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
		name string
		priv string
		pub  string
	}{
		{"0", "0x04c4bfa272a15b2f00f36a72af9db41701e14d40b14621aa60ce3e34cfd2428b", "0x04c4bfa272a15b2f00f36a72af9db41701e14d40b14621aa60ce3e34cfd2428b"},
		// {"1", "015c6230af8da00dfe725b638ba8a7255693006f69b273b3f9bb56615393cf9a", "0x00e044f17514266f0c9a3d9d831ef077031a3ec2443e0ffc7bb4210e2d8b0191"},
		// {"2", "047fb848b91a174254c760fe0946c4fbda002b2d238dc40180afaeb2d6957aa5", "0x01ef823ddc1041c8d63b36941168b6133a69a50f132b1620096c1110c06e5b53"},
	}

	for _, k := range keys {
		t.Run(k.name, func(t *testing.T) {
			b, err := hex.DecodeString(k.priv[2:])
			require.NoError(t, err)
			key := StarkRaw(b).Key()

			assert.Equal(t, k.pub, key.PublicKeyStr(), "address calculated from private key does not match expected contract address")
		})
	}
}
