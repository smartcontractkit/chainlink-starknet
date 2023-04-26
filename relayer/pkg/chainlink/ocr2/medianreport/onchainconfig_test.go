package medianreport

import (
	"math/big"
	"testing"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/reportingplugin/median"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnchainConfigCodec(t *testing.T) {
	var params = []struct {
		name      string
		val       []*big.Int
		expectErr bool
	}{
		{
			name:      "positive min < positive max",
			val:       []*big.Int{big.NewInt(1), big.NewInt(1000)},
			expectErr: false,
		},
		{
			name:      "negative min < positive max",
			val:       []*big.Int{big.NewInt(-1000), big.NewInt(1000)},
			expectErr: false,
		},
		{
			name:      "negative min < negative max",
			val:       []*big.Int{big.NewInt(-1000), big.NewInt(-1)},
			expectErr: false,
		},
		{
			name:      "positive min > positive max",
			val:       []*big.Int{big.NewInt(1000), big.NewInt(1)},
			expectErr: true,
		},
		{
			name:      "negative min > negative max",
			val:       []*big.Int{big.NewInt(-1), big.NewInt(-1000)},
			expectErr: true,
		},
		{
			name:      "positive min > negative max",
			val:       []*big.Int{big.NewInt(1), big.NewInt(-1000)},
			expectErr: true,
		},
		{
			name:      "min = max",
			val:       []*big.Int{big.NewInt(0), big.NewInt(0)},
			expectErr: false,
		},
	}

	codec := OnchainConfigCodec{}

	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			cfg := median.OnchainConfig{
				Min: p.val[0],
				Max: p.val[1],
			}

			configBytes, err := codec.Encode(cfg)
			require.NoError(t, err)

			newCfg, err := codec.Decode(configBytes)
			if p.expectErr {
				assert.Error(t, err)
				return // exit func if error is verified
			}
			require.NoError(t, err)

			assert.True(t, cfg.Min.Cmp(newCfg.Min) == 0)
			assert.True(t, cfg.Max.Cmp(newCfg.Max) == 0)
		})
	}

	t.Run("incorrect length", func(t *testing.T) {
		b := make([]byte, length-1)
		_, err := codec.Decode(b)
		assert.Error(t, err)
	})

	t.Run("incorrect version", func(t *testing.T) {
		b := make([]byte, length)
		b[0] = 1
		_, err := codec.Decode(b)
		assert.Error(t, err)
	})
}
