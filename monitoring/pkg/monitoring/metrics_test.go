package monitoring

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	relayMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
)

func TestMetrics_TransactionMetrics(t *testing.T) {
	// Create a test logger
	logger := relayMonitoring.NewTestLogger(t)
	metrics := NewMetrics(logger)

	// Test successful transactions
	chainID := "SN_MAIN"
	metrics.IncrementSuccessfulTransactions(chainID)
	metrics.IncrementSuccessfulTransactions(chainID)
	metrics.IncrementSuccessfulTransactions(chainID)

	// Test reverted transactions
	metrics.IncrementRevertedTransactions(chainID)
	metrics.IncrementRevertedTransactions(chainID)

	// Test finalized transactions
	metrics.IncrementFinalizedTransactions(chainID)
	metrics.IncrementFinalizedTransactions(chainID)
	metrics.IncrementFinalizedTransactions(chainID)
	metrics.IncrementFinalizedTransactions(chainID)

	// Test tx attempt count
	metrics.SetTxAttemptCount(chainID, 5)
	metrics.SetTxAttemptCount(chainID, 10)

	// Verify successful transactions counter
	successfulCount := testutil.ToFloat64(successfulTransactions.With(prometheus.Labels{"chainID": chainID}))
	assert.Equal(t, 3.0, successfulCount, "Successful transactions count should be 3")

	// Verify reverted transactions counter
	revertedCount := testutil.ToFloat64(revertedTransactions.With(prometheus.Labels{"chainID": chainID}))
	assert.Equal(t, 2.0, revertedCount, "Reverted transactions count should be 2")

	// Verify finalized transactions counter
	finalizedCount := testutil.ToFloat64(finalizedTransactions.With(prometheus.Labels{"chainID": chainID}))
	assert.Equal(t, 4.0, finalizedCount, "Finalized transactions count should be 4")

	// Verify tx attempt count gauge
	attemptCount := testutil.ToFloat64(txAttemptCount.With(prometheus.Labels{"chainID": chainID}))
	assert.Equal(t, 10.0, attemptCount, "Tx attempt count should be 10")
}

func TestMetrics_MultipleChains(t *testing.T) {
	// Create a test logger
	logger := relayMonitoring.NewTestLogger(t)
	metrics := NewMetrics(logger)

	chainID1 := "SN_MAIN"
	chainID2 := "SN_SEPOLIA"

	// Test metrics for multiple chains
	metrics.IncrementSuccessfulTransactions(chainID1)
	metrics.IncrementSuccessfulTransactions(chainID2)
	metrics.IncrementRevertedTransactions(chainID1)
	metrics.IncrementFinalizedTransactions(chainID2)
	metrics.SetTxAttemptCount(chainID1, 5)
	metrics.SetTxAttemptCount(chainID2, 3)

	// Verify chain1 metrics
	successfulCount1 := testutil.ToFloat64(successfulTransactions.With(prometheus.Labels{"chainID": chainID1}))
	revertedCount1 := testutil.ToFloat64(revertedTransactions.With(prometheus.Labels{"chainID": chainID1}))
	attemptCount1 := testutil.ToFloat64(txAttemptCount.With(prometheus.Labels{"chainID": chainID1}))

	assert.Equal(t, 1.0, successfulCount1, "Chain1 successful transactions should be 1")
	assert.Equal(t, 1.0, revertedCount1, "Chain1 reverted transactions should be 1")
	assert.Equal(t, 5.0, attemptCount1, "Chain1 attempt count should be 5")

	// Verify chain2 metrics
	successfulCount2 := testutil.ToFloat64(successfulTransactions.With(prometheus.Labels{"chainID": chainID2}))
	finalizedCount2 := testutil.ToFloat64(finalizedTransactions.With(prometheus.Labels{"chainID": chainID2}))
	attemptCount2 := testutil.ToFloat64(txAttemptCount.With(prometheus.Labels{"chainID": chainID2}))

	assert.Equal(t, 1.0, successfulCount2, "Chain2 successful transactions should be 1")
	assert.Equal(t, 1.0, finalizedCount2, "Chain2 finalized transactions should be 1")
	assert.Equal(t, 3.0, attemptCount2, "Chain2 attempt count should be 3")
}

func TestMetrics_ZeroValues(t *testing.T) {
	// Create a test logger
	logger := relayMonitoring.NewTestLogger(t)
	metrics := NewMetrics(logger)

	chainID := "SN_MAIN"

	// Test setting attempt count to 0
	metrics.SetTxAttemptCount(chainID, 0)

	// Verify the gauge is set to 0
	attemptCount := testutil.ToFloat64(txAttemptCount.With(prometheus.Labels{"chainID": chainID}))
	assert.Equal(t, 0.0, attemptCount, "Tx attempt count should be 0")

	// Verify counters start at 0
	successfulCount := testutil.ToFloat64(successfulTransactions.With(prometheus.Labels{"chainID": chainID}))
	revertedCount := testutil.ToFloat64(revertedTransactions.With(prometheus.Labels{"chainID": chainID}))
	finalizedCount := testutil.ToFloat64(finalizedTransactions.With(prometheus.Labels{"chainID": chainID}))

	assert.Equal(t, 0.0, successfulCount, "Successful transactions should start at 0")
	assert.Equal(t, 0.0, revertedCount, "Reverted transactions should start at 0")
	assert.Equal(t, 0.0, finalizedCount, "Finalized transactions should start at 0")
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	// Create a test logger
	logger := relayMonitoring.NewTestLogger(t)
	metrics := NewMetrics(logger)

	chainID := "SN_MAIN"
	numGoroutines := 10
	numIncrements := 100

	// Test concurrent access to metrics
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numIncrements; j++ {
				metrics.IncrementSuccessfulTransactions(chainID)
				metrics.IncrementRevertedTransactions(chainID)
				metrics.IncrementFinalizedTransactions(chainID)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final counts
	expectedCount := float64(numGoroutines * numIncrements)

	successfulCount := testutil.ToFloat64(successfulTransactions.With(prometheus.Labels{"chainID": chainID}))
	revertedCount := testutil.ToFloat64(revertedTransactions.With(prometheus.Labels{"chainID": chainID}))
	finalizedCount := testutil.ToFloat64(finalizedTransactions.With(prometheus.Labels{"chainID": chainID}))

	assert.Equal(t, expectedCount, successfulCount, "Concurrent successful transactions count should be correct")
	assert.Equal(t, expectedCount, revertedCount, "Concurrent reverted transactions count should be correct")
	assert.Equal(t, expectedCount, finalizedCount, "Concurrent finalized transactions count should be correct")
}

func TestMetrics_InterfaceCompliance(t *testing.T) {
	// Test that our implementation satisfies the Metrics interface
	var _ Metrics = &defaultMetrics{}

	// Test that NoOpTxMetrics satisfies the TxMetrics interface
	var _ TxMetrics = NoOpTxMetrics{}
}

func TestNoOpTxMetrics(t *testing.T) {
	// Test that NoOpTxMetrics methods don't panic
	noOp := NoOpTxMetrics{}

	// These should not panic
	require.NotPanics(t, func() {
		noOp.IncrementSuccessfulTransactions("SN_MAIN")
		noOp.IncrementRevertedTransactions("SN_MAIN")
		noOp.IncrementFinalizedTransactions("SN_MAIN")
		noOp.SetTxAttemptCount("SN_MAIN", 5)
	})
}
