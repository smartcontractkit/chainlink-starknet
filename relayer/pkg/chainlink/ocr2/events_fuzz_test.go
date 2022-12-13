package ocr2

import (
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"

	fuzz "github.com/google/gofuzz"
)

// TODO: ParseNewTransmissionEvent doesn't check for type overflows
// Ex: caigotypes.Felt may contain big.Int larger than Prime
// roundId := uint32(eventData[index].Uint64()) may have value larger UINT32_MAX
func TestRandomStruct(t *testing.T) {
	const ObservationMaxBytes = 16
	// Locally used struct to fill eventData array
	type CallData struct {
		RoundId           *caigotypes.Felt
		LatestAnswer      *caigotypes.Felt
		Transmitter       *caigotypes.Felt
		UnixTime          *caigotypes.Felt
		ObserversRaw      *caigotypes.Felt
		Observations      []*caigotypes.Felt
		JuelsPerFeeCoin   *caigotypes.Felt
		GasPrice          *caigotypes.Felt
		DigestData        *caigotypes.Felt
		EpochAndRoundData *caigotypes.Felt
		Reimbursement     *caigotypes.Felt
	}

	eventData := []*caigotypes.Felt{}
	f := fuzz.New().RandSource(rand.NewSource(0)).NilChance(0).Funcs(
		func(felt *caigotypes.Felt, c fuzz.Continue) {
			feltRaw := make([]byte, 32)
			size, err := c.Read(feltRaw)

			assert.NoError(t, err)
			require.Equal(t, size, len(feltRaw))

			*felt = caigotypes.Felt{new(big.Int).SetBytes(feltRaw)}
			eventData = append(eventData, felt)
		},
		func(felts *[]*caigotypes.Felt, c fuzz.Continue) {
			observationsLen := c.Intn(MaxObservers)

			observations := []*caigotypes.Felt{}
			data := make([]byte, ObservationMaxBytes)
			for i := 0; i < observationsLen; i++ {
				_, err := c.Read(data)
				require.NoError(t, err)

				observations = append(observations, &caigotypes.Felt{new(big.Int).SetBytes(data)})
			}

			*felts = observations
			eventData = append(
				append(
					eventData,
					&caigotypes.Felt{big.NewInt(int64(observationsLen))},
				),
				observations...,
			)
		},
	)

	f.Fuzz(&CallData{})

	_, err := ParseNewTransmissionEvent(eventData)
	require.NoError(t, err)
}
