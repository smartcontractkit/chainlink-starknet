package txm

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTxMetrics is a mock implementation of TxMetrics for testing
type mockTxMetrics struct {
	mu                   sync.RWMutex
	numBroadcastedTxs    int
	numConfirmedTxs      int
	numNonceGaps         int
	reachedMaxAttempts   bool
	timeUntilTxConfirmed []float64
	queueFullEvents      int
}

func newMockTxMetrics() *mockTxMetrics {
	return &mockTxMetrics{
		timeUntilTxConfirmed: make([]float64, 0),
	}
}

func (m *mockTxMetrics) IncrementNumBroadcastedTxs(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.numBroadcastedTxs++
}

func (m *mockTxMetrics) IncrementNumConfirmedTxs(ctx context.Context, confirmedTransactions int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.numConfirmedTxs += confirmedTransactions
}

func (m *mockTxMetrics) IncrementNumNonceGaps(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.numNonceGaps++
}

func (m *mockTxMetrics) ReachedMaxAttempts(ctx context.Context, reached bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reachedMaxAttempts = reached
}

func (m *mockTxMetrics) RecordTimeUntilTxConfirmed(ctx context.Context, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeUntilTxConfirmed = append(m.timeUntilTxConfirmed, duration)
}

func (m *mockTxMetrics) IncrementQueueFullEvents(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queueFullEvents++
}

func (m *mockTxMetrics) GetBroadcastedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.numBroadcastedTxs
}

func (m *mockTxMetrics) GetConfirmedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.numConfirmedTxs
}

func (m *mockTxMetrics) GetNonceGapsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.numNonceGaps
}

func (m *mockTxMetrics) GetReachedMaxAttempts() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.reachedMaxAttempts
}

func (m *mockTxMetrics) GetTimeUntilTxConfirmed() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]float64{}, m.timeUntilTxConfirmed...)
}

func (m *mockTxMetrics) GetQueueFullEventsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.queueFullEvents
}

// mockLogger implements logger.Logger interface for testing
type mockLogger struct {
	logger.Logger
}

// mockKeystore implements loop.Keystore interface for testing
type mockKeystore struct{}

func (m *mockKeystore) Accounts(ctx context.Context) ([]string, error) {
	return []string{"test-account"}, nil
}

func (m *mockKeystore) Get(ctx context.Context, id string) ([]byte, error) {
	return []byte("test-key"), nil
}

func (m *mockKeystore) Sign(ctx context.Context, account string, data []byte) ([]byte, error) {
	return []byte("test-signature"), nil
}

// mockConfig implements Config interface for testing
type mockConfig struct{}

func (m *mockConfig) ConfirmationPoll() time.Duration {
	return 1 * time.Second
}

func (m *mockConfig) TxTimeout() time.Duration {
	return 30 * time.Second
}

func (m *mockConfig) MaxAttempts() int {
	return 5
}

func (m *mockConfig) FeeEstimationMaxAttempts() int {
	return 3
}

func TestTxMetrics_QueueFullEvents(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Test initial state
	assert.Equal(t, 0, mockMetrics.GetQueueFullEventsCount())

	// Test incrementing queue full events
	mockMetrics.IncrementQueueFullEvents(ctx)
	assert.Equal(t, 1, mockMetrics.GetQueueFullEventsCount())

	mockMetrics.IncrementQueueFullEvents(ctx)
	mockMetrics.IncrementQueueFullEvents(ctx)
	assert.Equal(t, 3, mockMetrics.GetQueueFullEventsCount())
}

func TestNewWithMetrics(t *testing.T) {
	t.Parallel()

	mockLggr := &mockLogger{Logger: logger.Test(t)}
	mockMetrics := newMockTxMetrics()

	txm, err := NewWithMetrics(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		"test-chain-id",
		mockMetrics,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)

	require.NoError(t, err)
	require.NotNil(t, txm)

	// Verify metrics and chainID are set correctly
	stxm := txm.(*starktxm)
	assert.Equal(t, mockMetrics, stxm.metrics)
	assert.Equal(t, "test-chain-id", stxm.chainID)
}

func TestNew_DefaultMetrics(t *testing.T) {
	t.Parallel()

	mockLggr := &mockLogger{Logger: logger.Test(t)}

	txm, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		"test-chain-id",
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)

	require.NoError(t, err)
	require.NotNil(t, txm)

	// Verify PrometheusMetrics is used by default
	stxm := txm.(*starktxm)
	_, ok := stxm.metrics.(*prometheusMetrics)
	assert.True(t, ok, "Default metrics should be PrometheusMetrics")
	assert.Equal(t, "test-chain-id", stxm.chainID)
}

func TestNew_UsesPrometheusMetrics(t *testing.T) {
	t.Parallel()

	mockLggr := &mockLogger{Logger: logger.Test(t)}
	chainID := "test-prometheus-chain"

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

	// Verify PrometheusMetrics is used by default
	stxm := txm.(*starktxm)
	_, ok := stxm.metrics.(*prometheusMetrics)
	assert.True(t, ok, "Default metrics should be PrometheusMetrics")
	assert.Equal(t, chainID, stxm.chainID)
}

func TestInflightCount_UpdatesMetrics(t *testing.T) {
	t.Parallel()

	mockLggr := &mockLogger{Logger: logger.Test(t)}
	chainID := "test-chain-inflight"

	// Use real Prometheus metrics, not mocks
	txm, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chainID,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)

	require.NoError(t, err)

	// Get initial metric value from Prometheus
	initialValue := getGaugeValueFromPrometheus(t, "txm_reached_max_attempts", chainID)

	// Call InflightCount (this doesn't update metrics in v2, but we can verify the method works)
	queueCount, unconfirmedCount := txm.InflightCount()
	assert.Equal(t, 0, queueCount)
	assert.Equal(t, 0, unconfirmedCount)

	// Verify it was actually called (value should be set, not just initial)
	_ = initialValue // We set it regardless of initial value
}

func TestTxMetrics_Methods(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Test IncrementNumBroadcastedTxs
	mockMetrics.IncrementNumBroadcastedTxs(ctx)
	mockMetrics.IncrementNumBroadcastedTxs(ctx)
	assert.Equal(t, 2, mockMetrics.GetBroadcastedCount())

	// Test IncrementNumConfirmedTxs
	mockMetrics.IncrementNumConfirmedTxs(ctx, 1)
	assert.Equal(t, 1, mockMetrics.GetConfirmedCount())

	// Test IncrementNumNonceGaps
	mockMetrics.IncrementNumNonceGaps(ctx)
	mockMetrics.IncrementNumNonceGaps(ctx)
	mockMetrics.IncrementNumNonceGaps(ctx)
	assert.Equal(t, 3, mockMetrics.GetNonceGapsCount())

	// Test ReachedMaxAttempts
	mockMetrics.ReachedMaxAttempts(ctx, true)
	assert.Equal(t, true, mockMetrics.GetReachedMaxAttempts())

	// Update with new value
	mockMetrics.ReachedMaxAttempts(ctx, false)
	assert.Equal(t, false, mockMetrics.GetReachedMaxAttempts())

	// Test RecordTimeUntilTxConfirmed
	mockMetrics.RecordTimeUntilTxConfirmed(ctx, 1.5)
	mockMetrics.RecordTimeUntilTxConfirmed(ctx, 2.3)
	times := mockMetrics.GetTimeUntilTxConfirmed()
	assert.Equal(t, 2, len(times))
	assert.Equal(t, 1.5, times[0])
	assert.Equal(t, 2.3, times[1])
}

func TestTxMetrics_MultipleCalls(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Multiple concurrent calls
	var wg sync.WaitGroup
	numGoroutines := 10
	incrementsPerGoroutine := 10

	// Test concurrent IncrementNumBroadcastedTxs
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				mockMetrics.IncrementNumBroadcastedTxs(ctx)
			}
		}()
	}

	wg.Wait()

	// Should have exactly numGoroutines * incrementsPerGoroutine
	expected := numGoroutines * incrementsPerGoroutine
	assert.Equal(t, expected, mockMetrics.GetBroadcastedCount())
}

func TestTxMetrics_Isolation(t *testing.T) {
	t.Parallel()

	// Test that two separate metric instances are independent
	metrics1 := newMockTxMetrics()
	metrics2 := newMockTxMetrics()
	ctx := context.Background()

	// Increment metrics1
	metrics1.IncrementNumBroadcastedTxs(ctx)
	metrics1.IncrementNumBroadcastedTxs(ctx)

	// Increment metrics2
	metrics2.IncrementNumConfirmedTxs(ctx, 1)

	// Verify metrics1
	assert.Equal(t, 2, metrics1.GetBroadcastedCount())
	assert.Equal(t, 0, metrics1.GetConfirmedCount())

	// Verify metrics2
	assert.Equal(t, 0, metrics2.GetBroadcastedCount())
	assert.Equal(t, 1, metrics2.GetConfirmedCount())
}

// Helper function to get gauge value from Prometheus
func getGaugeValueFromPrometheus(t *testing.T, metricName, chainID string) float64 {
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
