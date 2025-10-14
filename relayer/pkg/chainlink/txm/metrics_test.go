package txm

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTxMetrics is a mock implementation of TxMetrics for testing
type mockTxMetrics struct {
	mu                     sync.RWMutex
	successfulTransactions map[string]int
	revertedTransactions   map[string]int
	finalizedTransactions  map[string]int
	txAttemptCounts        map[string]int
}

func newMockTxMetrics() *mockTxMetrics {
	return &mockTxMetrics{
		successfulTransactions: make(map[string]int),
		revertedTransactions:   make(map[string]int),
		finalizedTransactions:  make(map[string]int),
		txAttemptCounts:        make(map[string]int),
	}
}

func (m *mockTxMetrics) IncrementSuccessfulTransactions(chainID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.successfulTransactions[chainID]++
}

func (m *mockTxMetrics) IncrementRevertedTransactions(chainID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revertedTransactions[chainID]++
}

func (m *mockTxMetrics) IncrementFinalizedTransactions(chainID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.finalizedTransactions[chainID]++
}

func (m *mockTxMetrics) SetTxAttemptCount(chainID string, count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txAttemptCounts[chainID] = count
}

func (m *mockTxMetrics) GetSuccessfulCount(chainID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.successfulTransactions[chainID]
}

func (m *mockTxMetrics) GetRevertedCount(chainID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.revertedTransactions[chainID]
}

func (m *mockTxMetrics) GetFinalizedCount(chainID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.finalizedTransactions[chainID]
}

func (m *mockTxMetrics) GetTxAttemptCount(chainID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.txAttemptCounts[chainID]
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

	// Verify NoOpTxMetrics implements TxMetrics interface
	var _ TxMetrics = NoOpTxMetrics{}
	var _ TxMetrics = &mockTxMetrics{}
}

func TestNoOpTxMetrics(t *testing.T) {
	t.Parallel()

	metrics := NoOpTxMetrics{}

	// Should not panic
	metrics.IncrementSuccessfulTransactions("test-chain")
	metrics.IncrementRevertedTransactions("test-chain")
	metrics.IncrementFinalizedTransactions("test-chain")
	metrics.SetTxAttemptCount("test-chain", 10)
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

func TestInflightCount_UpdatesMetrics(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	mockLggr := &mockLogger{Logger: logger.Test(t)}
	chainID := "test-chain-id"

	txm, err := NewWithMetrics(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chainID,
		mockMetrics,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)

	require.NoError(t, err)

	// Call InflightCount
	queue, unconfirmed := txm.InflightCount()

	// Verify counts
	assert.Equal(t, 0, queue)
	assert.Equal(t, 0, unconfirmed)

	// Verify metrics were updated
	attemptCount := mockMetrics.GetTxAttemptCount(chainID)
	assert.Equal(t, 0, attemptCount)
}

func TestTxMetrics_Methods(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	chainID := "test-chain-id"

	// Test IncrementSuccessfulTransactions
	mockMetrics.IncrementSuccessfulTransactions(chainID)
	mockMetrics.IncrementSuccessfulTransactions(chainID)
	assert.Equal(t, 2, mockMetrics.GetSuccessfulCount(chainID))

	// Test IncrementRevertedTransactions
	mockMetrics.IncrementRevertedTransactions(chainID)
	assert.Equal(t, 1, mockMetrics.GetRevertedCount(chainID))

	// Test IncrementFinalizedTransactions
	mockMetrics.IncrementFinalizedTransactions(chainID)
	mockMetrics.IncrementFinalizedTransactions(chainID)
	mockMetrics.IncrementFinalizedTransactions(chainID)
	assert.Equal(t, 3, mockMetrics.GetFinalizedCount(chainID))

	// Test SetTxAttemptCount
	mockMetrics.SetTxAttemptCount(chainID, 42)
	assert.Equal(t, 42, mockMetrics.GetTxAttemptCount(chainID))

	// Update with new value
	mockMetrics.SetTxAttemptCount(chainID, 100)
	assert.Equal(t, 100, mockMetrics.GetTxAttemptCount(chainID))
}

func TestTxMetrics_MultipleCalls(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	chainID := "test-chain-id"

	// Multiple concurrent calls
	var wg sync.WaitGroup
	numGoroutines := 10
	incrementsPerGoroutine := 10

	// Test concurrent IncrementSuccessfulTransactions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				mockMetrics.IncrementSuccessfulTransactions(chainID)
			}
		}()
	}

	wg.Wait()

	// Should have exactly numGoroutines * incrementsPerGoroutine
	expected := numGoroutines * incrementsPerGoroutine
	assert.Equal(t, expected, mockMetrics.GetSuccessfulCount(chainID))
}

func TestTxMetrics_MultipleChains(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	chain1 := "chain-1"
	chain2 := "chain-2"

	// Increment metrics for chain1
	mockMetrics.IncrementSuccessfulTransactions(chain1)
	mockMetrics.IncrementSuccessfulTransactions(chain1)

	// Increment metrics for chain2
	mockMetrics.IncrementFinalizedTransactions(chain2)

	// Verify chain1 metrics
	assert.Equal(t, 2, mockMetrics.GetSuccessfulCount(chain1))
	assert.Equal(t, 0, mockMetrics.GetFinalizedCount(chain1))

	// Verify chain2 metrics
	assert.Equal(t, 0, mockMetrics.GetSuccessfulCount(chain2))
	assert.Equal(t, 1, mockMetrics.GetFinalizedCount(chain2))
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
