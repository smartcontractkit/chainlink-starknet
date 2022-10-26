package ocr2

import (
	junotypes "github.com/NethermindEth/juno/pkg/types"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
		// TODO: Random test data can be correct
		assert.Equal(t, err != nil, true)
	})
}
