package starknet

import (
	"context"
	"testing"
	"time"

	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

func TestGatewayClient(t *testing.T) {
	// todo: adjust for e2e tests
	chainID := gateway.GOERLI_ID
	lggr := logger.Test(t)
	timeout := 10 * time.Second

	client, err := NewClient(chainID, "", lggr, &timeout)
	require.NoError(t, err)
	assert.Equal(t, timeout, *client.defaultTimeout)

	t.Run("get chain id", func(t *testing.T) {
		id, err := client.ChainID(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, id, chainID)
	})

	t.Run("get block height", func(t *testing.T) {
		_, err := client.LatestBlockHeight(context.Background())
		assert.NoError(t, err)
	})
}

func TestGateWayClient_DefaultTimeout(t *testing.T) {
	client, err := NewClient(gateway.GOERLI_ID, "", logger.Test(t), nil)
	require.NoError(t, err)
	assert.Nil(t, client.defaultTimeout)
}
