package wait

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/never/internal/checker"
)

// ErrMaxAttemptsExceeded is returned when the maximum number of attempts is reached.
var ErrMaxAttemptsExceeded = errors.New("max attempts reached")

// WaitUntilReady continuously attempts to connect to the specified target until it becomes available or the context is canceled.
func WaitUntilReady(
	ctx context.Context,
	interval time.Duration,
	maxAttempts int,
	checker checker.Checker,
	logger *slog.Logger,
) error {
	logger = logger.With(
		slog.String("target", checker.Name()),
		slog.String("type", checker.Type()),
		slog.String("address", checker.Address()),
		slog.Duration("interval", interval),
		slog.Int("max_attempts", maxAttempts),
	)

	logger.Info(fmt.Sprintf("Waiting for %s to become ready...", checker.Name()))

	timer := newStoppedTimer(interval)
	defer timer.Stop()

	attempt := 0

	for {
		attempt++
		err := checker.Check(ctx)
		if err == nil {
			logger.Info(fmt.Sprintf("%s is ready ✓", checker.Name()), slog.Int("attempt", attempt))
			return nil // Successfully connected to the target
		}

		logger.Warn(fmt.Sprintf("%s is not ready ✗", checker.Name()),
			slog.String("error", err.Error()),
			slog.Int("attempt", attempt),
		)

		if maxAttempts > 0 && attempt >= maxAttempts {
			return ErrMaxAttemptsExceeded
		}

		timer.Reset(interval) // Reset starts the timer again
		select {
		case <-timer.C:
			// Wait until the timer expires
			// Continue to the next connection attempt after the interval
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil // Treat context cancellation as expected behavior
			}
			return ctx.Err()
		}
	}
}

// newStoppedTimer returns a new timer that is stopped and reset.
func newStoppedTimer(d time.Duration) *time.Timer {
	timer := time.NewTimer(d)
	if !timer.Stop() {
		<-timer.C
	}
	return timer
}
