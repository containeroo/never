package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/flag"
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/never/internal/runner"
	"github.com/containeroo/tinyflags"
)

// Run is the main function of the application.
func Run(ctx context.Context, version string, args []string, output io.Writer) error {
	// Create a new context that listens for interrupt signals
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Parse command-line flags
	flags, err := flag.ParseFlags(args, version)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(output, err.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("configuration error: %w", err)
	}

	// Initialize target checkers
	checkers, err := factory.BuildCheckers(flags.DynamicGroups, flags.DefaultCheckInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize target checkers: %w", err)
	}

	// Setup logger
	logger := logging.SetupLogger(version, output)

	// Run all checkers
	return runner.RunAll(ctx, checkers, logger)
}
