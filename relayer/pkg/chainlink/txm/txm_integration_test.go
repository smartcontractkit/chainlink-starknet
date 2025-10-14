//go:build integration

package txm

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_TXM_Metrics(t *testing.T) {
	t.Parallel()

	mockLggr := &mockLogger{Logger: logger.Test(t)}
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
	require.NotNil(t, txm)

	// Exercise all metrics to ensure they're registered
	stxm := txm.(*starktxm)
	stxm.metrics.IncrementSuccessfulTransactions(chainID)
	stxm.metrics.IncrementRevertedTransactions(chainID)
	stxm.metrics.IncrementFinalizedTransactions(chainID)
	stxm.metrics.SetTxAttemptCount(chainID, 0)

	// Test InflightCount which should also update metrics
	queue, unconfirmed := txm.InflightCount()
	assert.Equal(t, 0, queue)
	assert.Equal(t, 0, unconfirmed)

	// Verify all metrics are now registered in Prometheus
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	foundMetrics := map[string]bool{
		"tx_manager_num_successful_transactions": false,
		"tx_manager_num_tx_reverted":             false,
		"tx_manager_tx_attempt_count":            false,
		"tx_manager_num_finalized_transactions":  false,
	}

	for _, mf := range metricFamilies {
		if _, exists := foundMetrics[mf.GetName()]; exists {
			foundMetrics[mf.GetName()] = true
		}
	}

	for metricName, found := range foundMetrics {
		assert.True(t, found, "Metric %s should be registered", metricName)
	}

	// Verify actual values
	assert.Greater(t, getMetricValue(t, "tx_manager_num_successful_transactions", chainID), 0.0)
	assert.Greater(t, getMetricValue(t, "tx_manager_num_tx_reverted", chainID), 0.0)
	assert.Greater(t, getMetricValue(t, "tx_manager_num_finalized_transactions", chainID), 0.0)
	assert.Equal(t, 0.0, getMetricValue(t, "tx_manager_tx_attempt_count", chainID))
}

// TestIntegration_TXM_InflightCountMetrics tests that InflightCount properly updates metrics
func TestIntegration_TXM_InflightCountMetrics(t *testing.T) {
	t.Parallel()

	mockLggr := &mockLogger{Logger: logger.Test(t)}
	chainID := "inflightcount-integration-test"

	txm, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chainID,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)
	require.NoError(t, err)

	// Call InflightCount multiple times
	for i := 0; i < 5; i++ {
		queue, unconfirmed := txm.InflightCount()
		assert.Equal(t, 0, queue)
		assert.Equal(t, 0, unconfirmed)
	}

	// Verify metric is consistently set
	finalValue := getMetricValue(t, "tx_manager_tx_attempt_count", chainID)
	assert.Equal(t, 0.0, finalValue)
}

// Helper function to get metric value from Prometheus
func getMetricValue(t *testing.T, metricName, chainID string) float64 {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if mf.GetName() == metricName {
			for _, m := range mf.GetMetric() {
				if hasLabelValue(m, "chainID", chainID) {
					if m.Gauge != nil {
						return m.GetGauge().GetValue()
					}
					if m.Counter != nil {
						return m.GetCounter().GetValue()
					}
				}
			}
		}
	}
	return 0
}

// Helper to check if metric has specific label value
func hasLabelValue(m *dto.Metric, labelName, labelValue string) bool {
	for _, label := range m.GetLabel() {
		if label.GetName() == labelName && label.GetValue() == labelValue {
			return true
		}
	}
	return false
}
