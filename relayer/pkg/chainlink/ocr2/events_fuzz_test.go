package ocr2

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"testing"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransmissiionEvent3(t *testing.T) {
	const constNumOfElements = 11
	const ObservationMaxBytes = 16

	roundId := mathrand.Int31()
	latestAnswer, _ := starknet.RandomFelt()
	transmitter, _ := starknet.RandomFelt()
	unixTime := mathrand.Int63()
	fmt.Println(roundId, latestAnswer, transmitter, unixTime)

	observationsLen := mathrand.Intn(MaxObservers)

	observations := []*caigotypes.Felt{}
	data := make([]byte, ObservationMaxBytes)
	for i := 0; i < observationsLen; i++ {
		_, err := cryptorand.Read(data)
		require.NoError(t, err)

		// felt := caigotypes.Felt{new(big.Int).SetBytes(data)}
		observations = append(observations, &caigotypes.Felt{new(big.Int).SetBytes(data)})
	}
}

// func FuzzTransmissionEvent2(f *testing.F) {
// 	data := [][]byte{}
// 	for _, feltRaw := range newTransmissionEventRaw {
// 		felt := junotypes.HexToFelt(feltRaw)
// 		data = append(data, felt.Bytes())
// 	}

// 	f.Add(data...)
// 	f.Fuzz(func(t *testing.T, data ...[]byte) {
// 		felts := starknet.EncodeFelts(data)
// 		caigoFelts := []*caigotypes.Felt{}
// 		for _, felt := range felts {
// 			caigoFelt := caigotypes.Felt{felt}
// 			caigoFelts = append(caigoFelts, &caigoFelt)
// 		}

// 		_, err := ParseNewTransmissionEvent(caigoFelts)
// 		assert.Equal(t, err.Error(), "invalid: event data")
// 	})
// }

func FuzzTransmissionEvent(f *testing.F) {
	const chunkSize = 31

	data := []byte{}
	arr := [][]byte{}
	for _, feltRaw := range newTransmissionEventRaw {
		felt := junotypes.HexToFelt(feltRaw)
		feltBytes := felt.Bytes()
		arr = append(arr, feltBytes)
		data = append(data, feltBytes...)
	}

	f.Add(data)
	f.Add([]byte{0, 1, 2, 3, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		felts := starknet.EncodeFelts(data)
		caigoFelts := []*caigotypes.Felt{}
		for _, felt := range felts {
			caigoFelt := caigotypes.Felt{felt}
			caigoFelts = append(caigoFelts, &caigoFelt)
		}

		_, err := ParseNewTransmissionEvent(caigoFelts)
		assert.Equal(t, err.Error(), "invalid: event data")
	})
}
