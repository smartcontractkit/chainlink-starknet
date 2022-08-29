package ocr2

import (
	"math/big"
	"testing"

	"github.com/dontpanicdao/caigo"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/stretchr/testify/assert"
)

var (
	testConfigSetSelector = "config_set"
)

func TestIsEventFromContract(t *testing.T) {
	event := caigotypes.Event{
		Order:       0,
		FromAddress: "0x00",
		Keys:        []*caigotypes.Felt{caigotypes.BigToFelt(caigo.GetSelectorFromName(testConfigSetSelector))},
		Data:        []*caigotypes.Felt{},
	}

	// test zeros
	assert.True(t, isEventFromContract(&event, "0x000000", testConfigSetSelector))

	// test mismatch selector
	assert.False(t, isEventFromContract(&event, "0x00", "bad_selector"))

	// test mismatch addresses
	event.FromAddress = "0x00002432012bcda2bfa339c51b3be731118f2bd3bac6b63c5ca664c154bf636f"
	assert.False(t, isEventFromContract(&event, "0x3002432012bcda2bfa339c51b3be731118f2bd3bac6b63c5ca664c154bf6", testConfigSetSelector))

	// test different length addresses
	assert.True(t, isEventFromContract(&event, "0x2432012bcda2bfa339c51b3be731118f2bd3bac6b63c5ca664c154bf636f", testConfigSetSelector))
}

func TestParseAnswer(t *testing.T) {
	// Positive value (99)
	answer := parseAnswer("0x63")
	assert.Equal(t, big.NewInt(99), answer)

	// Negative value (-10)
	answer = parseAnswer("0x800000000000010fffffffffffffffffffffffffffffffffffffffffffffff7")
	assert.Equal(t, big.NewInt(-10), answer)
}
