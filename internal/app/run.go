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
	// Setup logger immediately so startup errors are correctly logged.
	logger := logging.SetupLogger(version, output)

	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			_, _ = fmt.Fprint(output, err.Error())
			return nil
		}
		logger.Error("failed to parse command-line flags", "err", err)
		return err
	}

	// Initialize target checkers
	checkers, err := factory.BuildCheckers(flags.DynamicGroups, flags.DefaultCheckInterval)
	if err != nil {
		logger.Error("failed to initialize target checkers", "err", err)
		return err
	}

	// Run all checkers
	if err = runner.RunAll(ctx, checkers, flags.MaxAttempts, logger); err != nil {
		logger.Error("failed to run all checkers", "err", err)
		return err
	}

	return nil
}
