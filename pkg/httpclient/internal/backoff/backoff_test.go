package backoff

import (
	"testing"
	"time"
)

func TestExponential(t *testing.T) {
	min := 1 * time.Second
	max := 30 * time.Second

	tests := []struct {
		name    string
		attempt int
		min     time.Duration
		max     time.Duration
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			name:    "attempt 0",
			attempt: 0,
			min:     min,
			max:     max,
			wantMin: min / 2,    // min - 30% jitter
			wantMax: min*2 + min, // min * 2 (base) + 30% jitter
		},
		{
			name:    "attempt 1",
			attempt: 1,
			min:     min,
			max:     max,
			wantMin: min,      // 2s - jitter
			wantMax: min*3 + 1, // 2s + jitter
		},
		{
			name:    "attempt 2",
			attempt: 2,
			min:     min,
			max:     max,
			wantMin: min*2 + min, // 4s - jitter
			wantMax: min * 6,      // 4s + jitter
		},
		{
			name:    "attempt 10 (exceeds max)",
			attempt: 10,
			min:     min,
			max:     max,
			wantMin: max - max/3, // capped at max with jitter
			wantMax: max,          // capped at max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Exponential(tt.attempt, tt.min, tt.max)

			// Check that result is within expected range (accounting for jitter)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("Exponential(%d, %v, %v) = %v, want between %v and %v",
					tt.attempt, tt.min, tt.max, got, tt.wantMin, tt.wantMax)
			}

			// Verify it doesn't exceed max
			if got > tt.max {
				t.Errorf("Exponential(%d, %v, %v) = %v, exceeds max %v",
					tt.attempt, tt.min, tt.max, got, tt.max)
			}
		})
	}
}

func TestExponential_AlwaysPositive(t *testing.T) {
	min := 100 * time.Millisecond
	max := 5 * time.Second

	for attempt := 0; attempt < 20; attempt++ {
		got := Exponential(attempt, min, max)
		if got <= 0 {
			t.Errorf("Exponential(%d) = %v, want positive duration", attempt, got)
		}
	}
}

func TestExponential_Jitter(t *testing.T) {
	min := 1 * time.Second
	max := 30 * time.Second
	attempt := 2

	// Run multiple times to verify jitter produces different values
	results := make(map[time.Duration]bool)

	for i := 0; i < 100; i++ {
		duration := Exponential(attempt, min, max)
		results[duration] = true
	}

	// With jitter, we should get multiple different values
	// (not guaranteed, but extremely likely over 100 attempts)
	if len(results) < 5 {
		t.Errorf("Exponential with jitter produced only %d unique values over 100 attempts, expected more variation",
			len(results))
	}
}

func TestExponential_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name    string
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{"zero min and max", 0, 0, 0},
		{"same min and max", 0, 5 * time.Second, 5 * time.Second},
		{"very small values", 0, 1 * time.Nanosecond, 10 * time.Nanosecond},
		{"very large values", 5, 1 * time.Hour, 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Exponential(tt.attempt, tt.min, tt.max)

			// Should always be within bounds
			if got < 0 {
				t.Errorf("Exponential returned negative value: %v", got)
			}

			if tt.max > 0 && got > tt.max {
				t.Errorf("Exponential returned %v, exceeds max %v", got, tt.max)
			}
		})
	}
}
