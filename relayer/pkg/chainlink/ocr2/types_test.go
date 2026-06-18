package ocr2

import (
	"math/big"
	"testing"
	"time"

	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/require"
)

func TestNewRoundData(t *testing.T) {
	raw := []string{
		"0x121e",
		"0x800000000000010ffffffffffffffffffffffffffffffffffffffffca5b1701",
		"0x1087",
		"0x633344a3",
		"0x633344a5",
	}

	felts, err := starknetutils.HexArrToFelt(raw)
	require.NoError(t, err)
	actualRound, err := NewRoundData(felts)
	require.NoError(t, err)
	expectedRound := RoundData{
		RoundID:     0x121e,
		Answer:      bigIntFromString("3618502788666131213697322783095070105623107215331596699973092056134972020481"),
		BlockNumber: 0x1087,
		StartedAt:   time.Unix(int64(0x633344a3), 0),
		UpdatedAt:   time.Unix(int64(0x633344a5), 0),
	}
	require.Equal(t, expectedRound, actualRound)
}

func TestNewRoundData_ProxyPhasePrefixedRoundID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		roundIDHex    string
		expectedRound uint32
	}{
		{
			name:          "DAI/USD phase 1",
			roundIDHex:    "0x100000000000000000000000000010c80",
			expectedRound: 0x10c80,
		},
		{
			name:          "USDC/USD phase 1",
			roundIDHex:    "0x10000000000000000000000000001192e",
			expectedRound: 0x1192e,
		},
		{
			name:          "ETH/USD phase 1",
			roundIDHex:    "0x10000000000000000000000000001664b",
			expectedRound: 0x1664b,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			raw := []string{
				tc.roundIDHex,
				"0x5f5e100",
				"0x1087",
				"0x633344a3",
				"0x633344a5",
			}
			felts, err := starknetutils.HexArrToFelt(raw)
			require.NoError(t, err)

			actualRound, err := NewRoundData(felts)
			require.NoError(t, err)
			require.Equal(t, tc.expectedRound, actualRound.RoundID)
		})
	}
}

// Helpers

func bigIntFromString(s string) *big.Int {
	out, _ := new(big.Int).SetString(s, 10)
	return out
}
