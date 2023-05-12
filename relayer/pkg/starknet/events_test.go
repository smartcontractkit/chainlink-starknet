package starknet

import (
	"testing"

	caigogw "github.com/smartcontractkit/caigo/gateway"
	caigotypes "github.com/smartcontractkit/caigo/types"
	"github.com/stretchr/testify/assert"
)

var (
	testEventSelector = "transfer"
)

func TestIsEventFromContract(t *testing.T) {
	event := caigogw.Event{
		Order:       0,
		FromAddress: "0x00",
		Keys:        []*caigotypes.Felt{caigotypes.BigToFelt(caigotypes.GetSelectorFromName(testEventSelector))},
		Data:        []*caigotypes.Felt{},
	}

	// test zeros
	assert.True(t, IsEventFromContract(&event, caigotypes.HexToHash("0x000000"), testEventSelector))

	// test mismatch selector
	assert.False(t, IsEventFromContract(&event, caigotypes.HexToHash("0x00"), "bad_selector"))

	// test mismatch addresses
	event.FromAddress = "0x00002432012bcda2bfa339c51b3be731118f2bd3bac6b63c5ca664c154bf636f"
	assert.False(t, IsEventFromContract(&event, caigotypes.HexToHash("0x3002432012bcda2bfa339c51b3be731118f2bd3bac6b63c5ca664c154bf6"), testEventSelector))

	// test different length addresses
	assert.True(t, IsEventFromContract(&event, caigotypes.HexToHash("0x2432012bcda2bfa339c51b3be731118f2bd3bac6b63c5ca664c154bf636f"), testEventSelector))
}
