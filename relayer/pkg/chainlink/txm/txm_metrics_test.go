package txm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMetrics_Registration(t *testing.T) {
	// Note: Cannot use t.Parallel() because we need to ensure metrics are initialized

	// Initialize metrics by calling them once (promauto registers on first use)
	testChainID := "test-registration"
	metrics := NewPrometheusMetrics(testChainID)
	ctx := context.Background()
	metrics.IncrementNumBroadcastedTxs(ctx)
	metrics.IncrementNumConfirmedTxs(ctx, 1)
	metrics.IncrementNumNonceGaps(ctx)
	metrics.ReachedMaxAttempts(ctx, true)
	metrics.RecordTimeUntilTxConfirmed(ctx, 1.5)

	// Test that all metrics are registered with Prometheus
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	metricsFound := map[string]bool{
		"txm_num_broadcasted_transactions": false,
		"txm_num_confirmed_transactions":   false,
		"txm_num_nonce_gaps":               false,
		"txm_reached_max_attempts":         false,
		"txm_time_until_tx_confirmed":      false,
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

func TestNewPrometheusMetrics(t *testing.T) {
	t.Parallel()

	chainID := "test-chain-prometheus"
	metrics := NewPrometheusMetrics(chainID)

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

	chainID := "test-chain-increment"
	metrics := NewPrometheusMetrics(chainID)
	ctx := context.Background()

	// Get initial values
	initialBroadcasted := getCounterValue(t, "txm_num_broadcasted_transactions", chainID)
	initialConfirmed := getCounterValue(t, "txm_num_confirmed_transactions", chainID)
	initialNonceGaps := getCounterValue(t, "txm_num_nonce_gaps", chainID)

	// Increment metrics
	metrics.IncrementNumBroadcastedTxs(ctx)
	metrics.IncrementNumBroadcastedTxs(ctx)
	metrics.IncrementNumConfirmedTxs(ctx, 1)
	metrics.IncrementNumNonceGaps(ctx)
	metrics.IncrementNumNonceGaps(ctx)
	metrics.IncrementNumNonceGaps(ctx)

	// Verify increments
	finalBroadcasted := getCounterValue(t, "txm_num_broadcasted_transactions", chainID)
	finalConfirmed := getCounterValue(t, "txm_num_confirmed_transactions", chainID)
	finalNonceGaps := getCounterValue(t, "txm_num_nonce_gaps", chainID)

	assert.Equal(t, initialBroadcasted+2, finalBroadcasted, "Broadcasted transactions should increment by 2")
	assert.Equal(t, initialConfirmed+1, finalConfirmed, "Confirmed transactions should increment by 1")
	assert.Equal(t, initialNonceGaps+3, finalNonceGaps, "Nonce gaps should increment by 3")
}

func TestPrometheusMetrics_SetGauge(t *testing.T) {
	t.Parallel()

	chainID := "test-chain-gauge"
	metrics := NewPrometheusMetrics(chainID)
	ctx := context.Background()

	// Set gauge values
	metrics.ReachedMaxAttempts(ctx, true)
	value := getGaugeValue(t, "txm_reached_max_attempts", chainID)
	assert.Equal(t, 1.0, value, "Gauge should be set to 1 (true)")

	// Update gauge
	metrics.ReachedMaxAttempts(ctx, false)
	value = getGaugeValue(t, "txm_reached_max_attempts", chainID)
	assert.Equal(t, 0.0, value, "Gauge should be updated to 0 (false)")
}

func TestPrometheusMetrics_MultipleChains(t *testing.T) {
	t.Parallel()

	chain1 := "chain-1-multi"
	chain2 := "chain-2-multi"

	metrics1 := NewPrometheusMetrics(chain1)
	metrics2 := NewPrometheusMetrics(chain2)
	ctx := context.Background()

	// Get initial values
	initialChain1 := getCounterValue(t, "txm_num_broadcasted_transactions", chain1)
	initialChain2 := getCounterValue(t, "txm_num_broadcasted_transactions", chain2)

	// Increment different chains
	metrics1.IncrementNumBroadcastedTxs(ctx)
	metrics1.IncrementNumBroadcastedTxs(ctx)
	metrics2.IncrementNumBroadcastedTxs(ctx)

	// Verify separate tracking
	finalChain1 := getCounterValue(t, "txm_num_broadcasted_transactions", chain1)
	finalChain2 := getCounterValue(t, "txm_num_broadcasted_transactions", chain2)

	assert.Equal(t, initialChain1+2, finalChain1, "Chain 1 should increment by 2")
	assert.Equal(t, initialChain2+1, finalChain2, "Chain 2 should increment by 1")
}

// Helper function to get counter value from Prometheus
func getCounterValue(t *testing.T, metricName, chainID string) float64 {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if mf.GetName() == metricName {
			for _, metric := range mf.GetMetric() {
				for _, labelPair := range metric.GetLabel() {
					if labelPair.GetName() == "chainID" && labelPair.GetValue() == chainID {
						return metric.GetCounter().GetValue()
					}
				}
			}
		}
	}
	return 0
}

// Helper function to get gauge value from Prometheus
func getGaugeValue(t *testing.T, metricName, chainID string) float64 {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if mf.GetName() == metricName {
			for _, metric := range mf.GetMetric() {
				for _, labelPair := range metric.GetLabel() {
					if labelPair.GetName() == "chainID" && labelPair.GetValue() == chainID {
						return metric.GetGauge().GetValue()
					}
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
	metrics := NewPrometheusMetrics(chainID)

	ctx := context.Background()

	// Test incrementing enqueue failed events
	metrics.IncrementEnqueueFailed(ctx)
	metrics.IncrementEnqueueFailed(ctx)
	metrics.IncrementEnqueueFailed(ctx)

	// Verify the metric was incremented by checking the counter value
	finalValue := getCounterValue(t, "txm_enqueue_failed", chainID)
	assert.Equal(t, 3.0, finalValue)
}
