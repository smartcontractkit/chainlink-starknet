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
		{"0", "0x0398eca85a333bc5de78f87d70d26f6e1f2438da6d163424b20f6190d3c38a21", "0x057605d472e1478b66396d8abec8f6c58348d9278d25049d9d73dafab40cde0c"},
		{"1", "0x18e693006a3dc4db5adf7812c2e4ab8d7729707fcb3c439de0939f39de8d2b", "0x05c96456a9d58aa45e997182050f07a0649638a8f1d955935b42b6898d99e63d"},
		{"2", "0x0358549759856b585a7b74ce5462e0ec0e56dbcc8fc729255150da2b62b702a6", "0x025d26785bc488193674b4e504f1ea0fc0bc28b0b92b7ce3e4b63ea5514bc3ab"},
	}

	for _, k := range keys {
		t.Run(k.name, func(t *testing.T) {
			b, err := hex.DecodeString(k.priv[2:])
			require.NoError(t, err)
			key := Raw(b).Key()

			assert.Equal(t, k.contract, key.PublicKeyStr(), "address calculated from private key does not match expected contract address")
		})
	}
}
