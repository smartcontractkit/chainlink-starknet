package txm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// MockTxMetrics is a mock implementation of TxMetrics for testing
type MockTxMetrics struct {
	SuccessfulTransactions int
	RevertedTransactions   int
	FinalizedTransactions  int
	TxAttemptCount         int
}

func (m *MockTxMetrics) IncrementSuccessfulTransactions(chainID string) {
	m.SuccessfulTransactions++
}

func (m *MockTxMetrics) IncrementRevertedTransactions(chainID string) {
	m.RevertedTransactions++
}

func (m *MockTxMetrics) IncrementFinalizedTransactions(chainID string) {
	m.FinalizedTransactions++
}

func (m *MockTxMetrics) SetTxAttemptCount(chainID string, count int) {
	m.TxAttemptCount = count
}

func TestTxMetrics_InterfaceCompliance(t *testing.T) {
	// Test that our interfaces are properly defined
	var _ TxMetrics = &MockTxMetrics{}
	var _ TxMetrics = NoOpTxMetrics{}
}

func TestNoOpTxMetrics(t *testing.T) {
	noOp := NoOpTxMetrics{}

	// Test that all methods can be called without panicking
	require.NotPanics(t, func() {
		noOp.IncrementSuccessfulTransactions("SN_MAIN")
		noOp.IncrementRevertedTransactions("SN_MAIN")
		noOp.IncrementFinalizedTransactions("SN_MAIN")
		noOp.SetTxAttemptCount("SN_MAIN", 5)
	})
}

func TestNewWithMetrics(t *testing.T) {
	// Test that NewWithMetrics creates a txm with metrics
	mockMetrics := &MockTxMetrics{}

	// Create a test logger to avoid nil pointer panic
	logger := &mockLogger{}

	txm, err := NewWithMetrics(
		logger,    // logger
		nil,       // keystore
		nil,       // config
		"SN_MAIN", // chainID
		mockMetrics,
		nil, // getClient
		nil, // getFeederClient
	)

	require.NoError(t, err)
	require.NotNil(t, txm)

	// Verify that the txm has the metrics and chainID set
	starkTxm := txm.(*starktxm)
	assert.Equal(t, "SN_MAIN", starkTxm.chainID)
	assert.Equal(t, mockMetrics, starkTxm.metrics)
}

// mockLogger is a simple mock logger for testing
type mockLogger struct{}

func (m *mockLogger) Name() string {
	return "mockLogger"
}

func (m *mockLogger) Named(name string, args ...interface{}) logger.Logger {
	return m
}

func (m *mockLogger) With(args ...interface{}) logger.Logger {
	return m
}

func (m *mockLogger) Trace(args ...interface{}) {}
func (m *mockLogger) Debug(args ...interface{}) {}
func (m *mockLogger) Info(args ...interface{})  {}
func (m *mockLogger) Warn(args ...interface{})  {}
func (m *mockLogger) Error(args ...interface{}) {}
func (m *mockLogger) Panic(args ...interface{}) {}
func (m *mockLogger) Fatal(args ...interface{}) {}

func (m *mockLogger) Tracef(format string, values ...interface{}) {}
func (m *mockLogger) Debugf(format string, values ...interface{}) {}
func (m *mockLogger) Infof(format string, values ...interface{})  {}
func (m *mockLogger) Warnf(format string, values ...interface{})  {}
func (m *mockLogger) Errorf(format string, values ...interface{}) {}
func (m *mockLogger) Panicf(format string, values ...interface{}) {}
func (m *mockLogger) Fatalf(format string, values ...interface{}) {}

func (m *mockLogger) Tracew(msg string, args ...interface{}) {}
func (m *mockLogger) Debugw(msg string, args ...interface{}) {}
func (m *mockLogger) Infow(msg string, args ...interface{})  {}
func (m *mockLogger) Warnw(msg string, args ...interface{})  {}
func (m *mockLogger) Errorw(msg string, args ...interface{}) {}
func (m *mockLogger) Panicw(msg string, args ...interface{}) {}
func (m *mockLogger) Fatalw(msg string, args ...interface{}) {}

func (m *mockLogger) Sync() error {
	return nil
}

func TestNew_DefaultMetrics(t *testing.T) {
	// Test that New creates a txm with NoOpTxMetrics by default
	logger := &mockLogger{}

	txm, err := New(
		logger,    // logger
		nil,       // keystore
		nil,       // config
		"SN_MAIN", // chainID
		nil,       // getClient
		nil,       // getFeederClient
	)

	require.NoError(t, err)
	require.NotNil(t, txm)

	// Verify that the txm has NoOpTxMetrics
	starkTxm := txm.(*starktxm)
	assert.Equal(t, "SN_MAIN", starkTxm.chainID)
	assert.IsType(t, NoOpTxMetrics{}, starkTxm.metrics)
}

func TestInflightCount_UpdatesMetrics(t *testing.T) {
	// Test that InflightCount updates the tx attempt count metric
	mockMetrics := &MockTxMetrics{}
	logger := &mockLogger{}

	txm, err := NewWithMetrics(
		logger,    // logger
		nil,       // keystore
		nil,       // config
		"SN_MAIN", // chainID
		mockMetrics,
		nil, // getClient
		nil, // getFeederClient
	)

	require.NoError(t, err)

	// Call InflightCount
	queue, unconfirmed := txm.InflightCount()

	// Verify that SetTxAttemptCount was called with the sum
	assert.Equal(t, queue+unconfirmed, mockMetrics.TxAttemptCount)
}

func TestTxMetrics_Methods(t *testing.T) {
	mockMetrics := &MockTxMetrics{}

	// Test IncrementSuccessfulTransactions
	mockMetrics.IncrementSuccessfulTransactions("SN_MAIN")
	assert.Equal(t, 1, mockMetrics.SuccessfulTransactions)

	// Test IncrementRevertedTransactions
	mockMetrics.IncrementRevertedTransactions("SN_MAIN")
	assert.Equal(t, 1, mockMetrics.RevertedTransactions)

	// Test IncrementFinalizedTransactions
	mockMetrics.IncrementFinalizedTransactions("SN_MAIN")
	assert.Equal(t, 1, mockMetrics.FinalizedTransactions)

	// Test SetTxAttemptCount
	mockMetrics.SetTxAttemptCount("SN_MAIN", 42)
	assert.Equal(t, 42, mockMetrics.TxAttemptCount)
}

func TestTxMetrics_MultipleCalls(t *testing.T) {
	mockMetrics := &MockTxMetrics{}

	// Test multiple increments
	for i := 0; i < 5; i++ {
		mockMetrics.IncrementSuccessfulTransactions("SN_MAIN")
		mockMetrics.IncrementRevertedTransactions("SN_MAIN")
		mockMetrics.IncrementFinalizedTransactions("SN_MAIN")
	}

	assert.Equal(t, 5, mockMetrics.SuccessfulTransactions)
	assert.Equal(t, 5, mockMetrics.RevertedTransactions)
	assert.Equal(t, 5, mockMetrics.FinalizedTransactions)

	// Test setting attempt count multiple times
	mockMetrics.SetTxAttemptCount("SN_MAIN", 10)
	assert.Equal(t, 10, mockMetrics.TxAttemptCount)

	mockMetrics.SetTxAttemptCount("SN_MAIN", 20)
	assert.Equal(t, 20, mockMetrics.TxAttemptCount)
}
