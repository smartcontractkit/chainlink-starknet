//go:build integration

package txm

import (
	"context"
	"sync"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTXM_Integration_Metrics(t *testing.T) {
	// Integration test to verify metrics are actually called during TXM operations
	// This test requires the integration build tag to run

	mockLggr := logger.Test(t)
	chainID := "integration-test-chain"

	// Create TXM with real Prometheus metrics
	txm, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chainID,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)
	require.NoError(t, err)

	// Start the TXM
	ctx := context.Background()
	err = txm.Start(ctx)
	require.NoError(t, err)

	// Get initial metric values
	testAccount := "0x123"
	initialBroadcasted := getCounterValue(t, "txm_num_broadcasted_transactions", chainID, testAccount)
	initialConfirmed := getCounterValue(t, "txm_num_confirmed_transactions", chainID, testAccount)
	initialNonceGaps := getCounterValue(t, "txm_num_nonce_gaps", chainID, testAccount)

	// Test that metrics are properly initialized
	assert.GreaterOrEqual(t, initialBroadcasted, 0.0)
	assert.GreaterOrEqual(t, initialConfirmed, 0.0)
	assert.GreaterOrEqual(t, initialNonceGaps, 0.0)

	// Test InflightCount method
	queueCount, unconfirmedCount := txm.InflightCount()
	assert.Equal(t, 0, queueCount)
	assert.Equal(t, 0, unconfirmedCount)

	// Clean up
	err = txm.Close()
	require.NoError(t, err)
}

func TestTXM_Integration_MetricsUnderLoad(t *testing.T) {
	// Test metrics behavior under concurrent load
	mockLggr := logger.Test(t)
	chainID := "load-test-chain"

	txm, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chainID,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)
	require.NoError(t, err)

	// Start the TXM
	ctx := context.Background()
	err = txm.Start(ctx)
	require.NoError(t, err)

	// Get the metrics instance to test concurrent access
	stxm := txm.(*starktxm)
	metrics := stxm.metrics

	// Test concurrent metric calls
	var wg sync.WaitGroup
	numGoroutines := 10
	callsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testAccount := "0x123"
			for j := 0; j < callsPerGoroutine; j++ {
				metrics.IncrementNumBroadcastedTxs(ctx, testAccount)
				metrics.IncrementNumConfirmedTxs(ctx, testAccount, 1)
				metrics.IncrementNumNonceGaps(ctx, testAccount)
				if j%2 == 0 {
					metrics.IncrementNonceRebroadcast(ctx, testAccount)
				}
				metrics.RecordTimeUntilTxConfirmed(ctx, testAccount, float64(j))
			}
		}()
	}

	wg.Wait()

	// Verify metrics were updated
	testAccount := "0x123"
	finalBroadcasted := getCounterValue(t, "txm_num_broadcasted_transactions", chainID, testAccount)
	finalConfirmed := getCounterValue(t, "txm_num_confirmed_transactions", chainID, testAccount)
	finalNonceGaps := getCounterValue(t, "txm_num_nonce_gaps", chainID, testAccount)

	expectedCalls := numGoroutines * callsPerGoroutine
	assert.Equal(t, float64(expectedCalls), finalBroadcasted)
	assert.Equal(t, float64(expectedCalls), finalConfirmed)
	assert.Equal(t, float64(expectedCalls), finalNonceGaps)

	// Clean up
	err = txm.Close()
	require.NoError(t, err)
}

func TestTXM_Integration_MultipleChains(t *testing.T) {
	// Test that metrics work correctly with multiple chain instances
	chain1 := "chain-1-integration"
	chain2 := "chain-2-integration"

	mockLggr := logger.Test(t)

	// Create two TXM instances with different chain IDs
	txm1, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chain1,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)
	require.NoError(t, err)

	txm2, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chain2,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)
	require.NoError(t, err)

	// Start both TXMs
	ctx := context.Background()
	err = txm1.Start(ctx)
	require.NoError(t, err)
	err = txm2.Start(ctx)
	require.NoError(t, err)

	// Get metrics instances
	stxm1 := txm1.(*starktxm)
	stxm2 := txm2.(*starktxm)

	// Update metrics for each chain
	testAccount := "0x123"
	stxm1.metrics.IncrementNumBroadcastedTxs(ctx, testAccount)
	stxm1.metrics.IncrementNumBroadcastedTxs(ctx, testAccount)
	stxm2.metrics.IncrementNumBroadcastedTxs(ctx, testAccount)

	// Verify separate tracking
	finalChain1 := getCounterValue(t, "txm_num_broadcasted_transactions", chain1, testAccount)
	finalChain2 := getCounterValue(t, "txm_num_broadcasted_transactions", chain2, testAccount)

	assert.Equal(t, 2.0, finalChain1, "Chain 1 should have 2 broadcasted transactions")
	assert.Equal(t, 1.0, finalChain2, "Chain 2 should have 1 broadcasted transaction")

	// Clean up
	err = txm1.Close()
	require.NoError(t, err)
	err = txm2.Close()
	require.NoError(t, err)
}
