package txm

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMetrics_Registration(t *testing.T) {
	// Note: Cannot use t.Parallel() because we need to ensure metrics are initialized

	// Initialize metrics by calling them once (promauto registers on first use)
	testChainID := "test-registration"
	metrics := NewPrometheusMetrics(testChainID)
	metrics.IncrementNumSuccessfulTxs()
	metrics.IncrementNumRevertedTxs()
	metrics.IncrementNumFinalizedTxs()
	metrics.SetTxAttemptCount(1)

	// Test that all metrics are registered with Prometheus
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	metricsFound := map[string]bool{
		"tx_manager_num_successful_transactions": false,
		"tx_manager_num_tx_reverted":             false,
		"tx_manager_tx_attempt_count":            false,
		"tx_manager_num_finalized_transactions":  false,
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

	// Test that NewPrometheusMetrics returns a valid instance
	metrics := NewPrometheusMetrics("test-chain")
	require.NotNil(t, metrics)

	// Verify it implements TxMetrics interface
	var _ TxMetrics = metrics
}

func TestPrometheusMetrics_ImplementsInterface(t *testing.T) {
	t.Parallel()

	// Compile-time check that prometheusMetrics implements TxMetrics
	var _ TxMetrics = &prometheusMetrics{chainID: "test"}
	var _ TxMetrics = (*prometheusMetrics)(nil)
}

func TestPrometheusMetrics_Increment(t *testing.T) {
	t.Parallel()

	chainID := "test-chain-increment"
	metrics := NewPrometheusMetrics(chainID)

	// Get initial values
	initialSuccessful := getCounterValue(t, "tx_manager_num_successful_transactions", chainID)
	initialReverted := getCounterValue(t, "tx_manager_num_tx_reverted", chainID)
	initialFinalized := getCounterValue(t, "tx_manager_num_finalized_transactions", chainID)

	// Increment metrics
	metrics.IncrementNumSuccessfulTxs()
	metrics.IncrementNumSuccessfulTxs()
	metrics.IncrementNumRevertedTxs()
	metrics.IncrementNumFinalizedTxs()
	metrics.IncrementNumFinalizedTxs()
	metrics.IncrementNumFinalizedTxs()

	// Verify increments
	finalSuccessful := getCounterValue(t, "tx_manager_num_successful_transactions", chainID)
	finalReverted := getCounterValue(t, "tx_manager_num_tx_reverted", chainID)
	finalFinalized := getCounterValue(t, "tx_manager_num_finalized_transactions", chainID)

	assert.Equal(t, initialSuccessful+2, finalSuccessful, "Successful transactions should increment by 2")
	assert.Equal(t, initialReverted+1, finalReverted, "Reverted transactions should increment by 1")
	assert.Equal(t, initialFinalized+3, finalFinalized, "Finalized transactions should increment by 3")
}

func TestPrometheusMetrics_SetGauge(t *testing.T) {
	t.Parallel()

	chainID := "test-chain-gauge"
	metrics := NewPrometheusMetrics(chainID)

	// Set gauge values
	metrics.SetTxAttemptCount(42)
	value := getGaugeValue(t, "tx_manager_tx_attempt_count", chainID)
	assert.Equal(t, 42.0, value, "Gauge should be set to 42")

	// Update gauge
	metrics.SetTxAttemptCount(100)
	value = getGaugeValue(t, "tx_manager_tx_attempt_count", chainID)
	assert.Equal(t, 100.0, value, "Gauge should be updated to 100")
}

func TestPrometheusMetrics_MultipleChains(t *testing.T) {
	t.Parallel()

	chain1 := "chain-1-multi"
	chain2 := "chain-2-multi"

	metrics1 := NewPrometheusMetrics(chain1)
	metrics2 := NewPrometheusMetrics(chain2)

	// Get initial values
	initialChain1 := getCounterValue(t, "tx_manager_num_successful_transactions", chain1)
	initialChain2 := getCounterValue(t, "tx_manager_num_successful_transactions", chain2)

	// Increment different chains
	metrics1.IncrementNumSuccessfulTxs()
	metrics1.IncrementNumSuccessfulTxs()
	metrics2.IncrementNumSuccessfulTxs()

	// Verify separate tracking
	finalChain1 := getCounterValue(t, "tx_manager_num_successful_transactions", chain1)
	finalChain2 := getCounterValue(t, "tx_manager_num_successful_transactions", chain2)

	assert.Equal(t, initialChain1+2, finalChain1, "Chain 1 should increment by 2")
	assert.Equal(t, initialChain2+1, finalChain2, "Chain 2 should increment by 1")
}

// Helper function to get counter value from Prometheus
func getCounterValue(t *testing.T, metricName, chainID string) float64 {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if mf.GetName() == metricName {
			for _, m := range mf.GetMetric() {
				if hasLabel(m, "chainID", chainID) {
					return m.GetCounter().GetValue()
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
			for _, m := range mf.GetMetric() {
				if hasLabel(m, "chainID", chainID) {
					return m.GetGauge().GetValue()
				}
			}
		}
	}
	return 0
}

// Helper function to check if metric has a specific label value
func hasLabel(m *dto.Metric, labelName, labelValue string) bool {
	for _, label := range m.GetLabel() {
		if label.GetName() == labelName && label.GetValue() == labelValue {
			return true
		}
	}
	return false
}
