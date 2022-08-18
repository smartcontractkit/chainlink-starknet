package starknet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

var (
	chainID = gateway.GOERLI_ID
	timeout = 10 * time.Second
)

func TestGatewayClient(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1) // mock endpoint only called once
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"parent_block_hash": "0x0", "status": "ACCEPTED_ON_L2", "timestamp": 1660840417, "gas_price": "0x174876e800", "sequencer_address": "0x4bbfb0d1aab5bf33eec5ada3a1040c41ed902a1eeb38c78a753d6f6359f1666", "transactions": [], "transaction_receipts": [], "state_root": "030c9b7339aabef2d6c293c40d4f0ec6ffae303cb7df5b705dce7acc00306b06", "starknet_version": "0.9.1", "block_hash": "0x0", "block_number": 0}`))
		require.NoError(t, err)
		wg.Done()
	}))
	defer mockServer.Close()

	// todo: adjust for e2e tests
	lggr := logger.Test(t)

	client, err := NewClient(chainID, mockServer.URL, lggr, &timeout)
	require.NoError(t, err)
	assert.Equal(t, timeout, client.defaultTimeout)

	// does not call endpoint - chainID returned from gateway client
	t.Run("get chain id", func(t *testing.T) {
		id, err := client.ChainID(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, id, chainID)
	})

	t.Run("get block height", func(t *testing.T) {
		_, err := client.LatestBlockHeight(context.Background())
		assert.NoError(t, err)
	})
	wg.Wait()
}

func TestGateWayClient_DefaultTimeout(t *testing.T) {
	client, err := NewClient(gateway.GOERLI_ID, "", logger.Test(t), nil)
	require.NoError(t, err)
	assert.Zero(t, client.defaultTimeout)
}
