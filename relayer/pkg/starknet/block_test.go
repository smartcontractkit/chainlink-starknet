package starknet

import (
	"encoding/json"
	"testing"

	starknetrpc "github.com/NethermindEth/starknet.go/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatestBlockID(t *testing.T) {
	blockID := LatestBlockID()
	assert.Equal(t, starknetrpc.BlockTagLatest, blockID.Tag)
}

func TestPreConfirmedBlockID(t *testing.T) {
	blockID := PreConfirmedBlockID()
	assert.Equal(t, starknetrpc.BlockTagPreConfirmed, blockID.Tag)

	raw, err := json.Marshal(blockID.Tag)
	require.NoError(t, err)
	assert.Equal(t, `"pre_confirmed"`, string(raw))
}
