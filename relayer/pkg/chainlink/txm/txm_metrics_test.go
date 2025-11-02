package txm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMetrics_Registration(t *testing.T) {
	// Note: Cannot use t.Parallel() because we need to ensure metrics are initialized

	// Initialize metrics by calling them once (promauto registers on first use)
	testChainID := "test-registration"
	metrics := NewTxmMetrics(testChainID)
	ctx := context.Background()
	testAccount := "0x123"
	metrics.IncrementNumBroadcastedTxs(ctx, testAccount)
	metrics.IncrementNumConfirmedTxs(ctx, testAccount, 1)
	metrics.IncrementNumNonceGaps(ctx, testAccount)
	metrics.IncrementNonceRebroadcast(ctx, testAccount)
	metrics.IncrementEnqueueFailed(ctx, testAccount)
	metrics.RecordTimeUntilTxConfirmed(ctx, testAccount, 1.5)
	testNonce := new(felt.Felt).SetUint64(42)
	metrics.UpdateNextNonceMetric(ctx, testAccount, testNonce)

	// Test that all metrics are registered with Prometheus
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	metricsFound := map[string]bool{
		"txm_num_broadcasted_transactions": false,
		"txm_num_confirmed_transactions":   false,
		"txm_num_nonce_gaps":               false,
		"txm_time_until_tx_confirmed":      false,
		"txm_enqueue_failed":               false,
		"txm_nonce_rebroadcast":            false,
	}

	for _, mf := range metricFamilies {
		if _, exists := metricsFound[mf.GetName()]; exists {
			metricsFound[mf.GetName()] = true
		}
	}

	for metricName, found := range metricsFound {
		assert.True(t, found, "Metric %s should be registered with Prometheus", metricName)
	}
}

func TestNewTxmMetrics(t *testing.T) {
	t.Parallel()

	chainID := "test-chain-prometheus"
	metrics := NewTxmMetrics(chainID)

	// Verify it returns the correct type
	_, ok := metrics.(*prometheusMetrics)
	assert.True(t, ok, "Should return prometheusMetrics instance")
}

func TestPrometheusMetrics_ImplementsInterface(t *testing.T) {
	t.Parallel()

	// Test that prometheusMetrics implements TxMetrics interface
	var _ TxMetrics = (*prometheusMetrics)(nil)
}

func TestPrometheusMetrics_Increment(t *testing.T) {
	t.Parallel()

	// Use unique chainID and accountAddress to avoid interference from other test runs
	chainID := fmt.Sprintf("test-chain-increment-%d", time.Now().UnixNano())
	metrics := NewTxmMetrics(chainID)
	ctx := context.Background()

	// Get initial values
	testAccount := fmt.Sprintf("0x%x", time.Now().UnixNano())
	initialBroadcasted := getCounterValue(t, "txm_num_broadcasted_transactions", chainID, testAccount)
	initialConfirmed := getCounterValue(t, "txm_num_confirmed_transactions", chainID, testAccount)
	initialNonceGaps := getCounterValue(t, "txm_num_nonce_gaps", chainID, testAccount)

	// Increment metrics
	metrics.IncrementNumBroadcastedTxs(ctx, testAccount)
	metrics.IncrementNumBroadcastedTxs(ctx, testAccount)
	metrics.IncrementNumConfirmedTxs(ctx, testAccount, 1)
	metrics.IncrementNumNonceGaps(ctx, testAccount)
	metrics.IncrementNumNonceGaps(ctx, testAccount)
	metrics.IncrementNumNonceGaps(ctx, testAccount)

	// Verify increments
	finalBroadcasted := getCounterValue(t, "txm_num_broadcasted_transactions", chainID, testAccount)
	finalConfirmed := getCounterValue(t, "txm_num_confirmed_transactions", chainID, testAccount)
	finalNonceGaps := getCounterValue(t, "txm_num_nonce_gaps", chainID, testAccount)

	assert.Equal(t, initialBroadcasted+2, finalBroadcasted, "Broadcasted transactions should increment by 2")
	assert.Equal(t, initialConfirmed+1, finalConfirmed, "Confirmed transactions should increment by 1")
	assert.Equal(t, initialNonceGaps+3, finalNonceGaps, "Nonce gaps should increment by 3")
}

func TestPrometheusMetrics_SetGauge(t *testing.T) {
	t.Parallel()

	// Use unique chainID and accountAddress to avoid interference from other test runs
	chainID := fmt.Sprintf("test-chain-gauge-%d", time.Now().UnixNano())
	metrics := NewTxmMetrics(chainID)
	ctx := context.Background()

	// Test IncrementNonceRebroadcast
	testAccount := fmt.Sprintf("0x%x", time.Now().UnixNano())
	initialValue := getCounterValue(t, "txm_nonce_rebroadcast", chainID, testAccount)

	metrics.IncrementNonceRebroadcast(ctx, testAccount)
	value := getCounterValue(t, "txm_nonce_rebroadcast", chainID, testAccount)
	assert.Equal(t, initialValue+1.0, value, "Nonce rebroadcast should increment by 1")

	// Test IncrementNonceRebroadcast again
	metrics.IncrementNonceRebroadcast(ctx, testAccount)
	value = getCounterValue(t, "txm_nonce_rebroadcast", chainID, testAccount)
	assert.Equal(t, initialValue+2.0, value, "Nonce rebroadcast should increment to 2")
}

func TestPrometheusMetrics_MultipleChains(t *testing.T) {
	t.Parallel()

	// Use unique chainIDs and accountAddress to avoid interference from other test runs
	baseID := time.Now().UnixNano()
	chain1 := fmt.Sprintf("chain-1-multi-%d", baseID)
	chain2 := fmt.Sprintf("chain-2-multi-%d", baseID+1)

	metrics1 := NewTxmMetrics(chain1)
	metrics2 := NewTxmMetrics(chain2)
	ctx := context.Background()

	// Get initial values
	testAccount := fmt.Sprintf("0x%x", baseID)
	initialChain1 := getCounterValue(t, "txm_num_broadcasted_transactions", chain1, testAccount)
	initialChain2 := getCounterValue(t, "txm_num_broadcasted_transactions", chain2, testAccount)

	// Increment different chains
	metrics1.IncrementNumBroadcastedTxs(ctx, testAccount)
	metrics1.IncrementNumBroadcastedTxs(ctx, testAccount)
	metrics2.IncrementNumBroadcastedTxs(ctx, testAccount)

	// Verify separate tracking
	finalChain1 := getCounterValue(t, "txm_num_broadcasted_transactions", chain1, testAccount)
	finalChain2 := getCounterValue(t, "txm_num_broadcasted_transactions", chain2, testAccount)

	assert.Equal(t, initialChain1+2, finalChain1, "Chain 1 should increment by 2")
	assert.Equal(t, initialChain2+1, finalChain2, "Chain 2 should increment by 1")
}

// Helper function to get counter value from Prometheus
func getCounterValue(t *testing.T, metricName, chainID, accountAddress string) float64 {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if mf.GetName() == metricName {
			for _, metric := range mf.GetMetric() {
				chainMatch := false
				accountMatch := false
				for _, labelPair := range metric.GetLabel() {
					if labelPair.GetName() == "chainID" && labelPair.GetValue() == chainID {
						chainMatch = true
					}
					if labelPair.GetName() == "accountAddress" && labelPair.GetValue() == accountAddress {
						accountMatch = true
					}
				}
				if chainMatch && accountMatch {
					return metric.GetCounter().GetValue()
				}
			}
		}
	}
	return 0
}

func TestPrometheusMetrics_EnqueueFailed(t *testing.T) {
	t.Parallel()

	// Create metrics with unique test chain ID to avoid conflicts across test runs
	chainID := fmt.Sprintf("test-chain-enqueue-%d", time.Now().UnixNano())
	metrics := NewTxmMetrics(chainID)

	ctx := context.Background()

	// Test incrementing enqueue failed events
	testAccount := fmt.Sprintf("0x%x", time.Now().UnixNano())
	initialValue := getCounterValue(t, "txm_enqueue_failed", chainID, testAccount)

	metrics.IncrementEnqueueFailed(ctx, testAccount)
	metrics.IncrementEnqueueFailed(ctx, testAccount)
	metrics.IncrementEnqueueFailed(ctx, testAccount)

	// Verify the metric was incremented by checking the counter value
	finalValue := getCounterValue(t, "txm_enqueue_failed", chainID, testAccount)
	assert.Equal(t, initialValue+3.0, finalValue)
}

func TestPrometheusMetrics_BeholderMetricsInitialized(t *testing.T) {
	t.Parallel()

	chainID := "test-beholder-init"
	metrics := NewTxmMetrics(chainID)

	// Verify that the metrics struct has beholder metrics initialized
	pm, ok := metrics.(*prometheusMetrics)
	require.True(t, ok, "Should be prometheusMetrics type")

	// Verify beholder metrics are not nil (they should be initialized)
	// Note: Even if beholder is not configured, they might be noop metrics, not nil
	assert.NotNil(t, pm.numBroadcastedTxs)
	assert.NotNil(t, pm.numConfirmedTxs)
	assert.NotNil(t, pm.numNonceGaps)
	assert.NotNil(t, pm.timeUntilTxConfirmed)
	assert.NotNil(t, pm.enqueueFailed)
	assert.NotNil(t, pm.nonceRebroadcast)
	assert.NotNil(t, pm.nextNonce)
}

func TestPrometheusMetrics_BeholderMetricsCanBeCalled(t *testing.T) {
	t.Parallel()

	chainID := "test-beholder-call"
	metrics := NewTxmMetrics(chainID)
	ctx := context.Background()

	// Verify all beholder metrics can be called without panicking
	// This ensures they're properly initialized and can handle method calls
	testAccount := "0x123"
	assert.NotPanics(t, func() {
		metrics.IncrementNumBroadcastedTxs(ctx, testAccount)
		metrics.IncrementNumConfirmedTxs(ctx, testAccount, 5)
		metrics.IncrementNumNonceGaps(ctx, testAccount)
		metrics.IncrementNonceRebroadcast(ctx, testAccount)
		metrics.RecordTimeUntilTxConfirmed(ctx, testAccount, 1.5)
		metrics.IncrementEnqueueFailed(ctx, testAccount)
		testNonce := new(felt.Felt).SetUint64(42)
		metrics.UpdateNextNonceMetric(ctx, testAccount, testNonce)
	})
}
