package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/wait"
	"golang.org/x/sync/errgroup"
)

// ErrNoCheckers is returned when RunAll is invoked with an empty checker list.
var ErrNoCheckers = errors.New("no checkers to run")

// RunAll runs all checkers concurrently and returns the first error or context cancellation.
func RunAll(ctx context.Context, checkers []factory.CheckerWithInterval, logger *slog.Logger) error {
	if len(checkers) == 0 {
		return ErrNoCheckers
	}

	// Run checkers concurrently
	eg, ctx := errgroup.WithContext(ctx)
	for _, chk := range checkers {
		checker := chk             // Capture loop variable
		name := chk.Checker.Name() // avoid re-calling in error path
		eg.Go(func() error {
			err := wait.WaitUntilReady(ctx, checker.Interval, checker.Checker, logger)
			if err != nil {
				return fmt.Errorf("checker '%s' failed: %w", name, err)
			}
			return nil
		})
	}

	// Wait for all checkers to finish or return error
	return eg.Wait()
}
