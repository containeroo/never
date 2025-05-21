package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/never/internal/config"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/never/internal/wait"
	"golang.org/x/sync/errgroup"
)

// Run is the main function of the application.
func Run(ctx context.Context, version string, args []string, output io.Writer) error {
	// Create a new context that listens for interrupt signals
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Parse command-line flags
	parsedFlags, err := config.ParseFlags(args, version)
	if err != nil {
		if errors.Is(err, &config.HelpRequested{}) {
			fmt.Fprint(output, err.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("configuration error: %w", err)
	}

	// Initialize target checkers
	checkers, err := factory.BuildCheckers(parsedFlags.DynFlags, parsedFlags.DefaultCheckInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize target checkers: %w", err)
	}

	if len(checkers) == 0 {
		return errors.New("configuration error: no checkers configured")
	}

	logger := logging.SetupLogger(version, output)

	// Run checkers concurrently
	eg, ctx := errgroup.WithContext(ctx)
	for _, chk := range checkers {
		checker := chk // Capture loop variable
		eg.Go(func() error {
			err := wait.WaitUntilReady(ctx, checker.Interval, checker.Checker, logger)
			if err != nil {
				return fmt.Errorf("checker '%s' failed: %w", checker.Checker.Name(), err)
			}
			return nil
		})
	}

	// Wait for all checkers to finish or return error
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
