package starknet

import (
	"context"
	"testing"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeederClient(t *testing.T) {
	client := NewTestClient(t)
	tx, err := client.TransactionFailure(context.TODO(), &felt.Zero)
	require.NoError(t, err)

	// test server will return this for a transaction failure
	// so, the test server will never return a nonce error
	assert.Equal(t, tx.Code, "SOME_ERROR")
	assert.Equal(t, tx.ErrorMessage, "some error was encountered")
}
