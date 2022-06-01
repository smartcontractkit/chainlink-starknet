package starknet

import (
	"context"
	"testing"

	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

func TestGatewayClient(t *testing.T) {
	chainID := gateway.GOERLI_ID
	lggr := logger.Test(t)

	client, err := NewClient(chainID, lggr)
	assert.NoError(t, err)

	t.Run("get chain id", func(t *testing.T) {
		id, err := client.ChainID(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, id, chainID)
	})
}
