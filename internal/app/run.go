package app

import (
	"context"
	"fmt"
	"io"

	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/flag"
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/never/internal/runner"

	"github.com/containeroo/httpgrace/server"
	"github.com/containeroo/tinyflags"
)

// Run is the main function of the application.
func Run(ctx context.Context, version string, args []string, stdOut, stdErr io.Writer) error {
	// Parse command-line flags
	flags, err := flag.ParseFlags(args, version)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			_, _ = fmt.Fprint(stdOut, err)
			return nil
		}
		_, _ = fmt.Fprintln(stdErr, err)
		return err
	}

	// Setup logger immediately so startup errors are correctly logged.
	logger := logging.SetupLogger(version, stdOut)

	// Initialize target checkers
	checkers, err := factory.BuildCheckers(flags.DynamicGroups, flags.DefaultCheckInterval)
	if err != nil {
		logger.Error("failed to initialize target checkers", "err", err)
		return err
	}

	// Create a signal-aware context that preserves the shutdown cause.
	ctx, stop := server.SignalContext(ctx)
	defer stop()

	// Run all checkers.
	err = runner.RunAll(ctx, checkers, flags.MaxAttempts, logger)

	if cause := context.Cause(ctx); cause != nil {
		logger.Info("context stopped", "cause", cause)
	}

	if err != nil {
		logger.Error("failed to run all checkers", "err", err)
		return err
	}

	return nil
}
