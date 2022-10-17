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
		{"0", "0x0398eca85a333bc5de78f87d70d26f6e1f2438da6d163424b20f6190d3c38a21", "0x06d368786b4dbf0d464a750c7b220079175aac577fd8e2f6fbe514237f8dd7ca"},
		{"1", "0x18e693006a3dc4db5adf7812c2e4ab8d7729707fcb3c439de0939f39de8d2b", "0x053091e77194fb61256b55ca25aad94d673f669163623f671cee50d5e8b387a2"},
		{"2", "0x0358549759856b585a7b74ce5462e0ec0e56dbcc8fc729255150da2b62b702a6", "0x07405687e273ca255ba689ae65c5681e8219e26a973c7e577590e566e20913ed"},
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
