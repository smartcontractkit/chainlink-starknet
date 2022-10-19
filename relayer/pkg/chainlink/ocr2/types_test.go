package ocr2

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestNewRoundData(t *testing.T) {
	raw := []string{
		"0x121e",
		"0x800000000000010ffffffffffffffffffffffffffffffffffffffffca5b1701",
		"0x1087",
		"0x633344a3",
		"0x633344a5",
	}

	felts := starknet.StringsToJunoFelts(raw)
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

// Helpers

func bigIntFromString(s string) *big.Int {
	out, _ := new(big.Int).SetString(s, 10)
	return out
}
