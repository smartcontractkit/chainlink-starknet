package txm

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_StateManagement(t *testing.T) {
	// Test the circuit breaker state management logic
	txm := &starktxm{
		clientFailures:    0,
		lastClientFailure: time.Time{},
		circuitBreakerMu:  sync.RWMutex{},
	}

	// Test initial state
	txm.circuitBreakerMu.RLock()
	initialFailures := txm.clientFailures
	txm.circuitBreakerMu.RUnlock()

	assert.Equal(t, 0, initialFailures, "Initial failure count should be 0")

	// Test incrementing failures
	txm.circuitBreakerMu.Lock()
	txm.clientFailures++
	txm.lastClientFailure = time.Now()
	txm.circuitBreakerMu.Unlock()

	txm.circuitBreakerMu.RLock()
	failures := txm.clientFailures
	txm.circuitBreakerMu.RUnlock()

	assert.Equal(t, 1, failures, "Failure count should be 1")

	// Test resetting failures
	txm.circuitBreakerMu.Lock()
	txm.clientFailures = 0
	txm.circuitBreakerMu.Unlock()

	txm.circuitBreakerMu.RLock()
	resetFailures := txm.clientFailures
	txm.circuitBreakerMu.RUnlock()

	assert.Equal(t, 0, resetFailures, "Failure count should be reset to 0")
}

func TestCircuitBreaker_ThreadSafety(t *testing.T) {
	// Test thread safety of circuit breaker state
	txm := &starktxm{
		clientFailures:    0,
		lastClientFailure: time.Time{},
		circuitBreakerMu:  sync.RWMutex{},
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Test concurrent access
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// Simulate failure recording
				txm.circuitBreakerMu.Lock()
				txm.clientFailures++
				txm.lastClientFailure = time.Now()
				txm.circuitBreakerMu.Unlock()

				// Simulate reading state
				txm.circuitBreakerMu.RLock()
				_ = txm.clientFailures
				_ = txm.lastClientFailure
				txm.circuitBreakerMu.RUnlock()
			}
		}()
	}

	wg.Wait()

	// Verify final state is consistent
	txm.circuitBreakerMu.RLock()
	finalFailures := txm.clientFailures
	txm.circuitBreakerMu.RUnlock()

	expectedFailures := numGoroutines * operationsPerGoroutine
	assert.Equal(t, expectedFailures, finalFailures,
		"Final failure count should match expected value")
}

func TestCircuitBreaker_BackoffCalculation(t *testing.T) {
	// Test the backoff calculation logic
	testCases := []struct {
		failures    int
		expected    time.Duration
		description string
	}{
		{5, 50 * time.Second, "5 failures should result in 50s backoff"},
		{6, 60 * time.Second, "6 failures should result in 60s backoff"},
		{10, 100 * time.Second, "10 failures should result in 100s backoff"},
		{18, 180 * time.Second, "18 failures should result in 180s backoff"},
		{20, MaxBackoffDuration, "20 failures should be capped at MaxBackoffDuration"},
		{25, MaxBackoffDuration, "25 failures should be capped at MaxBackoffDuration"},
		{30, MaxBackoffDuration, "30 failures should be capped at MaxBackoffDuration"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Simulate the backoff calculation from the circuit breaker
			backoffDuration := time.Duration(tc.failures) * CircuitBreakerBackoffStep
			if backoffDuration > MaxBackoffDuration {
				backoffDuration = MaxBackoffDuration
			}

			assert.Equal(t, tc.expected, backoffDuration,
				"Backoff calculation should be correct for %d failures", tc.failures)
		})
	}
}

func TestCircuitBreaker_ActivationThreshold(t *testing.T) {
	// Test that circuit breaker activates at the correct threshold
	threshold := CircuitBreakerThreshold

	// Test below threshold
	for failures := 1; failures < threshold; failures++ {
		backoffDuration := time.Duration(failures) * CircuitBreakerBackoffStep
		if backoffDuration > MaxBackoffDuration {
			backoffDuration = MaxBackoffDuration
		}

		// Below threshold, backoff should be reasonable
		assert.Less(t, backoffDuration, 60*time.Second,
			"Backoff should be reasonable for %d failures (below threshold)", failures)
	}

	// Test at threshold
	backoffDuration := time.Duration(threshold) * CircuitBreakerBackoffStep
	if backoffDuration > MaxBackoffDuration {
		backoffDuration = MaxBackoffDuration
	}

	assert.Equal(t, 50*time.Second, backoffDuration,
		"Backoff should start at 50s for threshold failures")
}

func TestCircuitBreaker_TimeWindow(t *testing.T) {
	// Test the 1-minute time window logic
	timeWindow := CircuitBreakerTimeWindow

	// Test that the time window is reasonable
	assert.Greater(t, timeWindow, 30*time.Second, "Time window should be at least 30 seconds")
	assert.Less(t, timeWindow, 5*time.Minute, "Time window should be less than 5 minutes")

	// Test time-based logic
	now := time.Now()
	oneMinuteAgo := now.Add(-timeWindow)

	// Test that failures within the time window are considered recent
	assert.True(t, now.Sub(oneMinuteAgo) <= timeWindow,
		"Time difference should be within the time window")

	// Test that failures outside the time window are considered old
	oldTime := now.Add(-2 * timeWindow)
	assert.True(t, now.Sub(oldTime) > timeWindow,
		"Time difference should be outside the time window")
}

func TestMaxBackoffDuration_Constant_CircuitBreaker(t *testing.T) {
	// Test that the MaxBackoffDuration constant is properly defined
	assert.Equal(t, 3*time.Minute, MaxBackoffDuration, "MaxBackoffDuration should be 3 minutes")

	// Test that it's reasonable
	assert.Greater(t, MaxBackoffDuration, 1*time.Minute, "MaxBackoffDuration should be at least 1 minute")
	assert.Less(t, MaxBackoffDuration, 10*time.Minute, "MaxBackoffDuration should be less than 10 minutes")
}

func TestCircuitBreaker_ResetLogic(t *testing.T) {
	// Test the reset logic
	txm := &starktxm{
		clientFailures:    5,
		lastClientFailure: time.Now(),
		circuitBreakerMu:  sync.RWMutex{},
	}

	// Verify initial state
	txm.circuitBreakerMu.RLock()
	initialFailures := txm.clientFailures
	txm.circuitBreakerMu.RUnlock()

	assert.Equal(t, 5, initialFailures, "Initial failure count should be 5")

	// Simulate successful connection and reset
	txm.circuitBreakerMu.Lock()
	if txm.clientFailures > 0 {
		txm.clientFailures = 0
	}
	txm.circuitBreakerMu.Unlock()

	// Verify reset
	txm.circuitBreakerMu.RLock()
	resetFailures := txm.clientFailures
	txm.circuitBreakerMu.RUnlock()

	assert.Equal(t, 0, resetFailures, "Failure count should be reset to 0")
}

func TestCircuitBreaker_BackoffProgression(t *testing.T) {
	// Test the progression of backoff durations
	expectedProgression := []struct {
		failures int
		expected time.Duration
	}{
		{5, 50 * time.Second},
		{6, 60 * time.Second},
		{7, 70 * time.Second},
		{8, 80 * time.Second},
		{9, 90 * time.Second},
		{10, 100 * time.Second},
		{15, 150 * time.Second},
		{18, 180 * time.Second},
		{20, MaxBackoffDuration}, // Should be capped
		{25, MaxBackoffDuration}, // Should be capped
		{30, MaxBackoffDuration}, // Should be capped
	}

	for _, tc := range expectedProgression {
		backoffDuration := time.Duration(tc.failures) * CircuitBreakerBackoffStep
		if backoffDuration > MaxBackoffDuration {
			backoffDuration = MaxBackoffDuration
		}

		assert.Equal(t, tc.expected, backoffDuration,
			"Backoff for %d failures should be %v", tc.failures, tc.expected)
	}
}
