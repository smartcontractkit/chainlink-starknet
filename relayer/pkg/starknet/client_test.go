package starknet

import (
	"context"
	"testing"

	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/db"
)

func TestGatewayClient(t *testing.T) {
	// todo: adjust for e2e tests
	chainID := gateway.GOERLI_ID
	lggr := logger.Test(t)

	client, err := NewClient(chainID, "", lggr, NewConfig(db.ChainCfg{}, lggr))
	assert.NoError(t, err)

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
