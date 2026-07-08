package backoff

import (
	"math"
	"time"
)

// Mode controls how retry intervals are calculated between failed checks.
type Mode string

const (
	// ModeNone keeps the retry interval constant.
	ModeNone Mode = "none"
	// ModeExponential doubles the retry interval after each failed attempt.
	ModeExponential Mode = "exponential"
)

// String returns the user-facing mode value.
func (m Mode) String() string { return string(m) }

// NextInterval returns the delay before the next retry attempt.
func NextInterval(mode Mode, base time.Duration, attempt int, max time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	if mode != ModeExponential || attempt <= 1 {
		return capInterval(base, max)
	}

	interval := base
	for i := 1; i < attempt; i++ {
		if interval > time.Duration(math.MaxInt64)/2 {
			return capInterval(time.Duration(math.MaxInt64), max)
		}
		interval *= 2
		if max > 0 && interval >= max {
			return max
		}
	}

	return capInterval(interval, max)
}

func capInterval(interval, max time.Duration) time.Duration {
	if max > 0 && interval > max {
		return max
	}
	return interval
}
