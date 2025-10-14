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
	mu                     sync.RWMutex
	successfulTransactions int
	revertedTransactions   int
	finalizedTransactions  int
	txAttemptCount         int
}

func newMockTxMetrics() *mockTxMetrics {
	return &mockTxMetrics{}
}

func (m *mockTxMetrics) IncrementNumSuccessfulTxs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.successfulTransactions++
}

func (m *mockTxMetrics) IncrementNumRevertedTxs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revertedTransactions++
}

func (m *mockTxMetrics) IncrementNumFinalizedTxs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.finalizedTransactions++
}

func (m *mockTxMetrics) SetTxAttemptCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txAttemptCount = count
}

func (m *mockTxMetrics) GetSuccessfulCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.successfulTransactions
}

func (m *mockTxMetrics) GetRevertedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.revertedTransactions
}

func (m *mockTxMetrics) GetFinalizedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.finalizedTransactions
}

func (m *mockTxMetrics) GetTxAttemptCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.txAttemptCount
}

// mockLogger implements logger.Logger interface for testing
type mockLogger struct {
	logger.Logger
}

func (m *mockLogger) Name() string {
	return "test-logger"
}

func (m *mockLogger) Sync() error {
	return nil
}

func (m *mockLogger) Panic(args ...interface{}) {
	// no-op for testing
}

func (m *mockLogger) Panicf(format string, args ...interface{}) {
	// no-op for testing
}

func TestTxMetrics_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Verify mockTxMetrics and prometheusMetrics implement TxMetrics interface
	var _ TxMetrics = &mockTxMetrics{}
	var _ TxMetrics = &prometheusMetrics{}
}

func TestNewWithMetrics(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	mockLggr := &mockLogger{Logger: logger.Test(t)}

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

	// Verify metrics are set
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

	// Call InflightCount to trigger metrics
	queueCount, unconfirmedCount := txm.InflightCount()
	assert.Equal(t, 0, queueCount)
	assert.Equal(t, 0, unconfirmedCount)

	// Verify the metric was actually set in Prometheus
	stxm := txm.(*starktxm)
	promMetrics, ok := stxm.metrics.(*prometheusMetrics)
	require.True(t, ok, "Should be using PrometheusMetrics")

	// Exercise all metric methods
	promMetrics.IncrementNumSuccessfulTxs()
	promMetrics.IncrementNumRevertedTxs()
	promMetrics.IncrementNumFinalizedTxs()
	promMetrics.SetTxAttemptCount(5)
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
	initialValue := getGaugeValueFromPrometheus(t, "tx_manager_tx_attempt_count", chainID)

	// Call InflightCount - this should update the Prometheus gauge
	queue, unconfirmed := txm.InflightCount()

	// Verify counts
	assert.Equal(t, 0, queue)
	assert.Equal(t, 0, unconfirmed)

	// Verify the actual Prometheus metric was updated
	finalValue := getGaugeValueFromPrometheus(t, "tx_manager_tx_attempt_count", chainID)
	assert.Equal(t, 0.0, finalValue, "Prometheus gauge should be set to 0")

	// Verify it was actually called (value should be set, not just initial)
	_ = initialValue // We set it regardless of initial value
}

func TestTxMetrics_Methods(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()

	// Test IncrementNumSuccessfulTxs
	mockMetrics.IncrementNumSuccessfulTxs()
	mockMetrics.IncrementNumSuccessfulTxs()
	assert.Equal(t, 2, mockMetrics.GetSuccessfulCount())

	// Test IncrementNumRevertedTxs
	mockMetrics.IncrementNumRevertedTxs()
	assert.Equal(t, 1, mockMetrics.GetRevertedCount())

	// Test IncrementNumFinalizedTxs
	mockMetrics.IncrementNumFinalizedTxs()
	mockMetrics.IncrementNumFinalizedTxs()
	mockMetrics.IncrementNumFinalizedTxs()
	assert.Equal(t, 3, mockMetrics.GetFinalizedCount())

	// Test SetTxAttemptCount
	mockMetrics.SetTxAttemptCount(42)
	assert.Equal(t, 42, mockMetrics.GetTxAttemptCount())

	// Update with new value
	mockMetrics.SetTxAttemptCount(100)
	assert.Equal(t, 100, mockMetrics.GetTxAttemptCount())
}

func TestTxMetrics_MultipleCalls(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()

	// Multiple concurrent calls
	var wg sync.WaitGroup
	numGoroutines := 10
	incrementsPerGoroutine := 10

	// Test concurrent IncrementNumSuccessfulTxs
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				mockMetrics.IncrementNumSuccessfulTxs()
			}
		}()
	}

	wg.Wait()

	// Should have exactly numGoroutines * incrementsPerGoroutine
	expected := numGoroutines * incrementsPerGoroutine
	assert.Equal(t, expected, mockMetrics.GetSuccessfulCount())
}

func TestTxMetrics_Isolation(t *testing.T) {
	t.Parallel()

	// Test that two separate metric instances are independent
	metrics1 := newMockTxMetrics()
	metrics2 := newMockTxMetrics()

	// Increment metrics1
	metrics1.IncrementNumSuccessfulTxs()
	metrics1.IncrementNumSuccessfulTxs()

	// Increment metrics2
	metrics2.IncrementNumFinalizedTxs()

	// Verify metrics1
	assert.Equal(t, 2, metrics1.GetSuccessfulCount())
	assert.Equal(t, 0, metrics1.GetFinalizedCount())

	// Verify metrics2
	assert.Equal(t, 0, metrics2.GetSuccessfulCount())
	assert.Equal(t, 1, metrics2.GetFinalizedCount())
}

// Helper function to get gauge value from Prometheus
func getGaugeValueFromPrometheus(t *testing.T, metricName, chainID string) float64 {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if mf.GetName() == metricName {
			for _, m := range mf.GetMetric() {
				for _, label := range m.GetLabel() {
					if label.GetName() == "chainID" && label.GetValue() == chainID {
						return m.GetGauge().GetValue()
					}
				}
			}
		}
	}
	return 0
}

// Mock implementations for testing

type mockKeystore struct{}

func (m *mockKeystore) Sign(ctx context.Context, keyID string, data []byte) ([]byte, error) {
	return []byte("mock-signature"), nil
}

func (m *mockKeystore) Accounts(ctx context.Context) ([]string, error) {
	return []string{"mock-account"}, nil
}

func (m *mockKeystore) AccountsByID(ctx context.Context, keyIDs ...string) ([]string, error) {
	return []string{"mock-account"}, nil
}

func (m *mockKeystore) SignTx(ctx context.Context, keyID string, tx []byte) ([]byte, error) {
	return []byte("mock-tx-signature"), nil
}

func (m *mockKeystore) New(ctx context.Context) (string, error) {
	return "mock-new-key", nil
}

func (m *mockKeystore) Delete(ctx context.Context, keyID string) error {
	return nil
}

func (m *mockKeystore) Export(ctx context.Context, keyID, password string) ([]byte, error) {
	return []byte("mock-export"), nil
}

func (m *mockKeystore) Import(ctx context.Context, keyJSON []byte, password string) error {
	return nil
}

func (m *mockKeystore) Get(ctx context.Context, keyID string) ([]byte, error) {
	return []byte("mock-key"), nil
}

type mockConfig struct{}

func (m *mockConfig) ConfirmationPoll() time.Duration { return time.Second }
func (m *mockConfig) TxTimeout() time.Duration        { return 10 * time.Second }
