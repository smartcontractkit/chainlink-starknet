package txm

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	"go.opentelemetry.io/otel/metric/noop"
)

// mockTxMetrics is a mock implementation of TxMetrics for testing
type mockTxMetrics struct {
	mu                   sync.RWMutex
	numBroadcastedTxs    int
	numConfirmedTxs      int
	numNonceGaps         int
	timeUntilTxConfirmed []float64
	enqueueFailed        int
	nonceRebroadcast     int
	nextNonce            map[string]int64 // accountAddress -> nonce
}

func newMockTxMetrics() *mockTxMetrics {
	return &mockTxMetrics{
		timeUntilTxConfirmed: make([]float64, 0),
		nextNonce:            make(map[string]int64),
	}
}

func (m *mockTxMetrics) IncrementNumBroadcastedTxs(ctx context.Context, accountAddress string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.numBroadcastedTxs++
}

func (m *mockTxMetrics) IncrementNumConfirmedTxs(ctx context.Context, accountAddress string, confirmedTransactions int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.numConfirmedTxs += confirmedTransactions
}

func (m *mockTxMetrics) IncrementNumNonceGaps(ctx context.Context, accountAddress string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.numNonceGaps++
}

func (m *mockTxMetrics) IncrementNonceRebroadcast(ctx context.Context, accountAddress string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nonceRebroadcast++
}

func (m *mockTxMetrics) UpdateNextNonceMetric(ctx context.Context, accountAddress string, nonce *felt.Felt) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextNonce[accountAddress] = nonce.BigInt(new(big.Int)).Int64()
}

func (m *mockTxMetrics) RecordTimeUntilTxConfirmed(ctx context.Context, accountAddress string, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeUntilTxConfirmed = append(m.timeUntilTxConfirmed, duration)
}

func (m *mockTxMetrics) IncrementEnqueueFailed(ctx context.Context, accountAddress string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enqueueFailed++
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

func (m *mockTxMetrics) GetNonceRebroadcastCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nonceRebroadcast
}

func (m *mockTxMetrics) GetNextNonce(accountAddress string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nextNonce[accountAddress]
}

func (m *mockTxMetrics) GetTimeUntilTxConfirmed() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]float64{}, m.timeUntilTxConfirmed...)
}

func (m *mockTxMetrics) GetEnqueueFailedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enqueueFailed
}

// mockLogger implements logger.Logger interface for testing
type mockLogger struct {
	logger.Logger
}

// mockKeystore implements loop.Keystore interface for testing
type mockKeystore struct{}

func (m *mockKeystore) Decrypt(ctx context.Context, account string, encrypted []byte) (decrypted []byte, err error) {
	return []byte("decrypted-data"), nil
}

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

func (m *mockConfig) FeeEstimationMaxAttempts() int {
	return 3
}

func TestTxMetrics_EnqueueFailed(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Test initial state
	assert.Equal(t, 0, mockMetrics.GetEnqueueFailedCount())

	// Test incrementing enqueue failed events
	mockMetrics.IncrementEnqueueFailed(ctx, "0x123")
	assert.Equal(t, 1, mockMetrics.GetEnqueueFailedCount())

	mockMetrics.IncrementEnqueueFailed(ctx, "0x123")
	mockMetrics.IncrementEnqueueFailed(ctx, "0x456")
	assert.Equal(t, 3, mockMetrics.GetEnqueueFailedCount())
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
	noopMeter := noop.NewMeterProvider().Meter("test")

	txm, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		"test-chain-id",
		noopMeter,
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

func TestInflightCount_ReturnsQueueAndUnconfirmedCounts(t *testing.T) {
	t.Parallel()

	mockLggr := &mockLogger{Logger: logger.Test(t)}
	chainID := "test-chain-inflight"
	noopMeter := noop.NewMeterProvider().Meter("test")

	// Use real Prometheus metrics, not mocks
	txm, err := New(
		mockLggr,
		&mockKeystore{},
		&mockConfig{},
		chainID,
		noopMeter,
		func() (*starknet.Client, error) { return nil, nil },
		func() (*starknet.FeederClient, error) { return nil, nil },
	)

	require.NoError(t, err)

	// Call InflightCount (this doesn't update metrics in v2, but we can verify the method works)
	queueCount, unconfirmedCount := txm.InflightCount()
	assert.Equal(t, 0, queueCount)
	assert.Equal(t, 0, unconfirmedCount)
}

func TestMetrics_AllMethodsWorkCorrectly(t *testing.T) {
	t.Parallel()

	mockMetrics := newMockTxMetrics()
	ctx := context.Background()

	// Test IncrementNumBroadcastedTxs
	mockMetrics.IncrementNumBroadcastedTxs(ctx, "0x123")
	mockMetrics.IncrementNumBroadcastedTxs(ctx, "0x123")
	assert.Equal(t, 2, mockMetrics.GetBroadcastedCount())

	// Test IncrementNumConfirmedTxs
	mockMetrics.IncrementNumConfirmedTxs(ctx, "0x123", 1)
	assert.Equal(t, 1, mockMetrics.GetConfirmedCount())

	// Test IncrementNumNonceGaps
	mockMetrics.IncrementNumNonceGaps(ctx, "0x123")
	mockMetrics.IncrementNumNonceGaps(ctx, "0x123")
	mockMetrics.IncrementNumNonceGaps(ctx, "0x456")
	assert.Equal(t, 3, mockMetrics.GetNonceGapsCount())

	// Test IncrementNonceRebroadcast
	mockMetrics.IncrementNonceRebroadcast(ctx, "0x123")
	mockMetrics.IncrementNonceRebroadcast(ctx, "0x123")
	assert.Equal(t, 2, mockMetrics.GetNonceRebroadcastCount())

	// Test UpdateNextNonceMetric
	testAccount := "0x123"
	testNonce := new(felt.Felt).SetUint64(42)
	mockMetrics.UpdateNextNonceMetric(ctx, testAccount, testNonce)
	assert.Equal(t, int64(42), mockMetrics.GetNextNonce(testAccount))

	// Test RecordTimeUntilTxConfirmed
	mockMetrics.RecordTimeUntilTxConfirmed(ctx, "0x123", 1.5)
	mockMetrics.RecordTimeUntilTxConfirmed(ctx, "0x123", 2.3)
	times := mockMetrics.GetTimeUntilTxConfirmed()
	assert.Equal(t, 2, len(times))
	assert.Equal(t, 1.5, times[0])
	assert.Equal(t, 2.3, times[1])
}

func TestTxMetrics_Isolation(t *testing.T) {
	t.Parallel()

	// Test that two separate metric instances are independent
	metrics1 := newMockTxMetrics()
	metrics2 := newMockTxMetrics()
	ctx := context.Background()

	// Increment metrics1
	metrics1.IncrementNumBroadcastedTxs(ctx, "0x123")
	metrics1.IncrementNumBroadcastedTxs(ctx, "0x123")

	// Increment metrics2
	metrics2.IncrementNumConfirmedTxs(ctx, "0x456", 1)

	// Verify metrics1
	assert.Equal(t, 2, metrics1.GetBroadcastedCount())
	assert.Equal(t, 0, metrics1.GetConfirmedCount())

	// Verify metrics2
	assert.Equal(t, 0, metrics2.GetBroadcastedCount())
	assert.Equal(t, 1, metrics2.GetConfirmedCount())
}
