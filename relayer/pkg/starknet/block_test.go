package starknet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatestBlockID(t *testing.T) {
	blockID := LatestBlockID()
	assert.Equal(t, BlockTagLatest, blockID.Tag)
}

func TestPreConfirmedBlockID(t *testing.T) {
	blockID := PreConfirmedBlockID()
	assert.Equal(t, BlockTagPreConfirmed, blockID.Tag)

	raw, err := blockIDFromJSON(BlockTagPreConfirmed)
	assert.NoError(t, err)
	assert.Equal(t, `"pre_confirmed"`, string(raw))
}
