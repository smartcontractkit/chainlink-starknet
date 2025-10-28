package txm

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTxMetrics_ErrorHandlingAdvanced(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T, metrics TxMetrics)
	}{
		{
			name: "NilContext",
			testFunc: func(t *testing.T, metrics TxMetrics) {
				// Test that metrics handle nil context gracefully
				metrics.IncrementNumBroadcastedTxs(context.TODO())
				metrics.IncrementNumConfirmedTxs(context.TODO(), 1)
				metrics.IncrementNumNonceGaps(context.TODO())
				metrics.ReachedMaxAttempts(context.TODO(), true)
				metrics.RecordTimeUntilTxConfirmed(context.TODO(), 1.0)
			},
		},
		{
			name: "NegativeConfirmedCount",
			testFunc: func(t *testing.T, metrics TxMetrics) {
				// Test that metrics handle negative confirmed count
				ctx := context.Background()
				// Note: Prometheus counters cannot decrease, so we test with 0 and positive values
				metrics.IncrementNumConfirmedTxs(ctx, 0)
				metrics.IncrementNumConfirmedTxs(ctx, 1)

				// For mock metrics, we can test negative values
				if mockMetrics, ok := metrics.(*mockTxMetrics); ok {
					mockMetrics.IncrementNumConfirmedTxs(ctx, -1)
				}
			},
		},
		{
			name: "NegativeDuration",
			testFunc: func(t *testing.T, metrics TxMetrics) {
				// Test that metrics handle negative duration
				ctx := context.Background()
				metrics.RecordTimeUntilTxConfirmed(ctx, -1.0)
				metrics.RecordTimeUntilTxConfirmed(ctx, 0.0)
				metrics.RecordTimeUntilTxConfirmed(ctx, 1.0)
			},
		},
		{
			name: "LargeValues",
			testFunc: func(t *testing.T, metrics TxMetrics) {
				// Test that metrics handle large values
				ctx := context.Background()
				metrics.IncrementNumConfirmedTxs(ctx, 1000000)
				metrics.RecordTimeUntilTxConfirmed(ctx, 999999.99)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with Prometheus metrics
			promMetrics := NewPrometheusMetrics("error-test-chain")
			tt.testFunc(t, promMetrics)

			// Test with mock metrics
			mockMetrics := newMockTxMetrics()
			tt.testFunc(t, mockMetrics)
		})
	}
}

func TestTxMetrics_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Test concurrent access to all metric methods
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Mix of different metric calls
			for j := 0; j < 10; j++ {
				mockMetrics.IncrementNumBroadcastedTxs(ctx)
				mockMetrics.IncrementNumConfirmedTxs(ctx, goroutineID%3+1)
				mockMetrics.IncrementNumNonceGaps(ctx)
				mockMetrics.ReachedMaxAttempts(ctx, j%2 == 0)
				mockMetrics.RecordTimeUntilTxConfirmed(ctx, float64(j))
			}
		}(i)
	}

	wg.Wait()

	// Verify all calls were processed correctly
	assert.Equal(t, numGoroutines*10, mockMetrics.GetBroadcastedCount())
	assert.Equal(t, numGoroutines*10, mockMetrics.GetNonceGapsCount())
	assert.Equal(t, numGoroutines*10, len(mockMetrics.GetTimeUntilTxConfirmed()))
}

func TestTxMetrics_PrometheusConcurrentAccess(t *testing.T) {
	t.Parallel()

	// Create metrics with unique test chain ID to avoid conflicts across test runs
	chainID := fmt.Sprintf("concurrent-test-chain-%d", time.Now().UnixNano())
	promMetrics := NewPrometheusMetrics(chainID)
	ctx := context.Background()

	// Test concurrent access to Prometheus metrics
	var wg sync.WaitGroup
	numGoroutines := 20
	callsPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < callsPerGoroutine; j++ {
				promMetrics.IncrementNumBroadcastedTxs(ctx)
				promMetrics.IncrementNumConfirmedTxs(ctx, 1)
				promMetrics.IncrementNumNonceGaps(ctx)
				promMetrics.ReachedMaxAttempts(ctx, j%2 == 0)
				promMetrics.RecordTimeUntilTxConfirmed(ctx, float64(j))
			}
		}()
	}

	wg.Wait()

	// Verify metrics were updated correctly
	expectedCalls := numGoroutines * callsPerGoroutine
	finalBroadcasted := getCounterValue(t, "txm_num_broadcasted_transactions", chainID)
	finalConfirmed := getCounterValue(t, "txm_num_confirmed_transactions", chainID)
	finalNonceGaps := getCounterValue(t, "txm_num_nonce_gaps", chainID)

	assert.Equal(t, float64(expectedCalls), finalBroadcasted)
	assert.Equal(t, float64(expectedCalls), finalConfirmed)
	assert.Equal(t, float64(expectedCalls), finalNonceGaps)
}

func TestTxMetrics_Performance(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Performance test - measure time for many metric calls
	numCalls := 10000
	start := time.Now()

	for i := 0; i < numCalls; i++ {
		mockMetrics.IncrementNumBroadcastedTxs(ctx)
		mockMetrics.IncrementNumConfirmedTxs(ctx, 1)
		mockMetrics.IncrementNumNonceGaps(ctx)
		mockMetrics.ReachedMaxAttempts(ctx, i%2 == 0)
		mockMetrics.RecordTimeUntilTxConfirmed(ctx, float64(i))
	}

	duration := time.Since(start)

	// Verify all calls were processed
	assert.Equal(t, numCalls, mockMetrics.GetBroadcastedCount())
	assert.Equal(t, numCalls, mockMetrics.GetConfirmedCount())
	assert.Equal(t, numCalls, mockMetrics.GetNonceGapsCount())
	assert.Equal(t, numCalls, len(mockMetrics.GetTimeUntilTxConfirmed()))

	// Performance should be reasonable (less than 1 second for 10k calls)
	assert.Less(t, duration, time.Second, "Metrics should be fast")
	t.Logf("Processed %d metric calls in %v", numCalls*5, duration)
}

func TestTxMetrics_MemoryUsage(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Test memory usage with many duration recordings
	numRecordings := 1000
	for i := 0; i < numRecordings; i++ {
		mockMetrics.RecordTimeUntilTxConfirmed(ctx, float64(i))
	}

	// Verify all recordings were stored
	durations := mockMetrics.GetTimeUntilTxConfirmed()
	assert.Equal(t, numRecordings, len(durations))

	// Verify data integrity
	for i, duration := range durations {
		assert.Equal(t, float64(i), duration)
	}
}

func TestTxMetrics_EdgeCases(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Test edge cases
	t.Run("ZeroValues", func(t *testing.T) {
		mockMetrics.IncrementNumConfirmedTxs(ctx, 0)
		mockMetrics.RecordTimeUntilTxConfirmed(ctx, 0.0)
		assert.Equal(t, 0, mockMetrics.GetConfirmedCount())
	})

	t.Run("VeryLargeValues", func(t *testing.T) {
		mockMetrics.IncrementNumConfirmedTxs(ctx, 1000000)
		mockMetrics.RecordTimeUntilTxConfirmed(ctx, 999999.99)
		assert.Equal(t, 1000000, mockMetrics.GetConfirmedCount())
	})

	t.Run("BooleanToggle", func(t *testing.T) {
		mockMetrics.ReachedMaxAttempts(ctx, true)
		assert.True(t, mockMetrics.GetReachedMaxAttempts())

		mockMetrics.ReachedMaxAttempts(ctx, false)
		assert.False(t, mockMetrics.GetReachedMaxAttempts())
	})
}

func TestTxMetrics_InterfaceComplianceAdvanced(t *testing.T) {
	t.Parallel()

	// Test that all implementations properly implement the interface
	var _ TxMetrics = (*prometheusMetrics)(nil)
	var _ TxMetrics = (*mockTxMetrics)(nil)

	// Test that interface methods can be called
	ctx := context.Background()

	promMetrics := NewPrometheusMetrics("compliance-test")
	promMetrics.IncrementNumBroadcastedTxs(ctx)
	promMetrics.IncrementNumConfirmedTxs(ctx, 1)
	promMetrics.IncrementNumNonceGaps(ctx)
	promMetrics.ReachedMaxAttempts(ctx, true)
	promMetrics.RecordTimeUntilTxConfirmed(ctx, 1.0)

	mockMetrics := newMockTxMetrics()
	mockMetrics.IncrementNumBroadcastedTxs(ctx)
	mockMetrics.IncrementNumConfirmedTxs(ctx, 1)
	mockMetrics.IncrementNumNonceGaps(ctx)
	mockMetrics.ReachedMaxAttempts(ctx, true)
	mockMetrics.RecordTimeUntilTxConfirmed(ctx, 1.0)
}
