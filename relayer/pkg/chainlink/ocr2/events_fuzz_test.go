package ocr2

import (
	"fmt"
	"testing"

	junotypes "github.com/NethermindEth/juno/pkg/types"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/assert"
)

func test(lels ...[]byte) {
	for _, lel := range lels {
		fmt.Println(lel)
	}
}

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
	test(arr...)

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
