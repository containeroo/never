package backoff

import (
	"math"
	"time"
)

// Mode controls how retry intervals are calculated between failed checks.
type Mode string

const (
	// ModeLinear keeps the retry interval constant.
	ModeLinear Mode = "linear"
	// ModeExponential doubles the retry interval after each failed attempt.
	ModeExponential Mode = "exponential"
)

// String returns the user-facing mode value.
func (m Mode) String() string { return string(m) }

// NextInterval returns the delay before the next retry attempt.
// intervalLimit caps the returned interval when greater than zero.
// A zero or negative intervalLimit means unlimited.
func NextInterval(mode Mode, base time.Duration, attempt int, intervalLimit time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	if usesBaseInterval(mode, attempt) {
		return limitInterval(base, intervalLimit)
	}

	interval := base
	for i := 1; i < attempt; i++ {
		if wouldOverflowOnDouble(interval) {
			return limitInterval(time.Duration(math.MaxInt64), intervalLimit)
		}

		interval *= 2
		if reachesIntervalLimit(interval, intervalLimit) {
			return intervalLimit
		}
	}

	return limitInterval(interval, intervalLimit)
}

// usesBaseInterval returns true when no exponential growth should be applied yet.
func usesBaseInterval(mode Mode, attempt int) bool {
	return mode != ModeExponential || attempt <= 1
}

// wouldOverflowOnDouble returns true when doubling interval would overflow time.Duration.
func wouldOverflowOnDouble(interval time.Duration) bool {
	return interval > time.Duration(math.MaxInt64)/2
}

// reachesIntervalLimit returns true when interval reached or exceeded a positive limit.
func reachesIntervalLimit(interval, intervalLimit time.Duration) bool {
	return intervalLimit > 0 && interval >= intervalLimit
}

// limitInterval caps interval to intervalLimit when intervalLimit is greater than zero.
func limitInterval(interval, intervalLimit time.Duration) time.Duration {
	if intervalLimit <= 0 {
		return interval
	}
	return min(interval, intervalLimit)
}
