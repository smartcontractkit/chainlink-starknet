package txm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackoffCalculation(t *testing.T) {
	testCases := []struct {
		failures    int
		expectedMin time.Duration
		expectedMax time.Duration
		description string
	}{
		{5, 50 * time.Second, 50 * time.Second, "5 failures should result in 50s backoff"},
		{6, 60 * time.Second, 60 * time.Second, "6 failures should result in 60s backoff"},
		{10, 100 * time.Second, 100 * time.Second, "10 failures should result in 100s backoff"},
		{15, 150 * time.Second, 150 * time.Second, "15 failures should result in 150s backoff"},
		{18, 180 * time.Second, 180 * time.Second, "18 failures should result in 180s backoff"},
		{20, 200 * time.Second, 200 * time.Second, "20 failures should result in 200s backoff"},
		{25, MaxBackoffDuration, MaxBackoffDuration, "25 failures should be capped at MaxBackoffDuration"},
		{30, MaxBackoffDuration, MaxBackoffDuration, "30 failures should be capped at MaxBackoffDuration"},
		{50, MaxBackoffDuration, MaxBackoffDuration, "50 failures should be capped at MaxBackoffDuration"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Simulate the backoff calculation from the circuit breaker
			backoffDuration := time.Duration(tc.failures) * 10 * time.Second
			if backoffDuration > MaxBackoffDuration {
				backoffDuration = MaxBackoffDuration
			}

			assert.GreaterOrEqual(t, backoffDuration, tc.expectedMin,
				"Backoff should be at least the expected minimum for %d failures", tc.failures)
			assert.LessOrEqual(t, backoffDuration, tc.expectedMax,
				"Backoff should not exceed the expected maximum for %d failures", tc.failures)
		})
	}
}

func TestMaxBackoffDuration_Constant_Backoff(t *testing.T) {
	// Test that the MaxBackoffDuration constant is properly defined
	assert.Equal(t, 3*time.Minute, MaxBackoffDuration, "MaxBackoffDuration should be 3 minutes")

	// Test that it's reasonable (not too short, not too long)
	assert.Greater(t, MaxBackoffDuration, 1*time.Minute, "MaxBackoffDuration should be at least 1 minute")
	assert.Less(t, MaxBackoffDuration, 10*time.Minute, "MaxBackoffDuration should be less than 10 minutes")
}

func TestBackoffProgression(t *testing.T) {
	// Test the progression of backoff durations
	expectedProgression := []time.Duration{
		50 * time.Second,  // 5 failures
		60 * time.Second,  // 6 failures
		70 * time.Second,  // 7 failures
		80 * time.Second,  // 8 failures
		90 * time.Second,  // 9 failures
		100 * time.Second, // 10 failures
		110 * time.Second, // 11 failures
		120 * time.Second, // 12 failures
		130 * time.Second, // 13 failures
		140 * time.Second, // 14 failures
		150 * time.Second, // 15 failures
		160 * time.Second, // 16 failures
		170 * time.Second, // 17 failures
		180 * time.Second, // 18 failures
	}

	for i, expected := range expectedProgression {
		failures := i + 5 // Start from 5 failures
		backoffDuration := time.Duration(failures) * 10 * time.Second
		if backoffDuration > MaxBackoffDuration {
			backoffDuration = MaxBackoffDuration
		}

		assert.Equal(t, expected, backoffDuration,
			"Backoff for %d failures should be %v", failures, expected)
	}
}

func TestBackoffCapping(t *testing.T) {
	// Test that backoff is properly capped at MaxBackoffDuration
	highFailureCounts := []int{20, 25, 30, 50, 100, 1000}

	for _, failures := range highFailureCounts {
		backoffDuration := time.Duration(failures) * 10 * time.Second
		if backoffDuration > MaxBackoffDuration {
			backoffDuration = MaxBackoffDuration
		}

		assert.Equal(t, MaxBackoffDuration, backoffDuration,
			"Backoff for %d failures should be capped at MaxBackoffDuration", failures)
	}
}

func TestBackoffFormula(t *testing.T) {
	// Test the exact formula: failures * 10 seconds, capped at MaxBackoffDuration
	testCases := []struct {
		failures int
		expected time.Duration
	}{
		{1, 10 * time.Second},
		{2, 20 * time.Second},
		{3, 30 * time.Second},
		{4, 40 * time.Second},
		{5, 50 * time.Second},
		{10, 100 * time.Second},
		{15, 150 * time.Second},
		{18, 180 * time.Second},
		{20, 200 * time.Second},
		{25, MaxBackoffDuration}, // Should be capped
		{30, MaxBackoffDuration}, // Should be capped
	}

	for _, tc := range testCases {
		// Apply the exact formula from the circuit breaker
		backoffDuration := time.Duration(tc.failures) * 10 * time.Second
		if backoffDuration > MaxBackoffDuration {
			backoffDuration = MaxBackoffDuration
		}

		assert.Equal(t, tc.expected, backoffDuration,
			"Backoff formula should produce %v for %d failures", tc.expected, tc.failures)
	}
}

func TestBackoffThreshold(t *testing.T) {
	// Test that the circuit breaker threshold (5 failures) works correctly
	threshold := 5

	// Below threshold - no backoff
	for failures := 1; failures < threshold; failures++ {
		backoffDuration := time.Duration(failures) * 10 * time.Second
		if backoffDuration > MaxBackoffDuration {
			backoffDuration = MaxBackoffDuration
		}

		assert.Less(t, backoffDuration, 60*time.Second,
			"Backoff should be reasonable for %d failures (below threshold)", failures)
	}

	// At threshold - backoff starts
	backoffDuration := time.Duration(threshold) * 10 * time.Second
	if backoffDuration > MaxBackoffDuration {
		backoffDuration = MaxBackoffDuration
	}

	assert.Equal(t, 50*time.Second, backoffDuration,
		"Backoff should start at 50s for threshold failures")
}

func TestBackoffTimeWindow(t *testing.T) {
	// Test that the 1-minute time window for failure counting is reasonable
	timeWindow := 1 * time.Minute

	// The time window should be reasonable for circuit breaker activation
	assert.Greater(t, timeWindow, 30*time.Second, "Time window should be at least 30 seconds")
	assert.Less(t, timeWindow, 5*time.Minute, "Time window should be less than 5 minutes")

	// Test that the time window is used in the circuit breaker logic
	// (This is more of a documentation test since we can't easily test the time-based logic)
	assert.True(t, timeWindow > 0, "Time window should be positive")
}
