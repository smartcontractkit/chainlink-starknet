package starknet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreConfirmedBlockID(t *testing.T) {
	blockID := PreConfirmedBlockID()
	assert.Equal(t, BlockTagPreConfirmed, string(blockID.Tag))

	raw, err := blockIDFromJSON(BlockTagPreConfirmed)
	assert.NoError(t, err)
	assert.Equal(t, `"pre_confirmed"`, string(raw))
}
