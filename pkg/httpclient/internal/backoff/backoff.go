package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Exponential calculates the exponential backoff duration for a given attempt.
// It implements exponential backoff with jitter to prevent thundering herd.
//
// The formula is: min(max * 2^attempt * (1 ± jitter), maxDuration)
// where jitter is a random value between 0 and 0.3 (30% variation).
//
// Parameters:
//   - attempt: The retry attempt number (0-indexed)
//   - min: The minimum wait duration
//   - max: The maximum wait duration
//
// Returns the calculated wait duration with jitter applied.
func Exponential(attempt int, min, max time.Duration) time.Duration {
	// Calculate base wait time: min * 2^attempt
	mult := math.Pow(2, float64(attempt))
	wait := time.Duration(float64(min) * mult)

	// Apply maximum cap
	if wait > max {
		wait = max
	}

	// Apply jitter (±30%)
	jitter := float64(wait) * (rand.Float64()*0.6 - 0.3)
	wait = time.Duration(float64(wait) + jitter)

	// Ensure we don't go below minimum
	if wait < min {
		wait = min
	}

	// Ensure we don't exceed maximum
	if wait > max {
		wait = max
	}

	return wait
}
