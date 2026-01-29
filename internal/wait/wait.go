package wait

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/never/internal/checker"
)

// WaitUntilReady continuously attempts to connect to the specified target until it becomes available or the context is canceled.
func WaitUntilReady(ctx context.Context, interval time.Duration, checker checker.Checker, logger *slog.Logger) error {
	logger = logger.With(
		slog.String("target", checker.Name()),
		slog.String("type", checker.Type()),
		slog.String("address", checker.Address()),
		slog.Duration("interval", interval),
	)

	logger.Info(fmt.Sprintf("Waiting for %s to become ready...", checker.Name()))

	timer := newStoppedTimer(interval)
	defer timer.Stop()

	for {
		err := checker.Check(ctx)
		if err == nil {
			logger.Info(fmt.Sprintf("%s is ready ✓", checker.Name()))
			return nil // Successfully connected to the target
		}

		logger.Warn(fmt.Sprintf("%s is not ready ✗", checker.Name()), slog.String("error", err.Error()))

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
