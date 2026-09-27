package cli

import (
	"time"

	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/tinyflags"
)

// registerAppFlags registers core application flags and binds them to cfg.
func registerAppFlags(tf *tinyflags.FlagSet, cfg *Config) {
	tf.DurationVar(&cfg.DefaultCheckInterval, "default-interval", defaultCheckInterval, "Default interval between checks. Can be overridden for each target.").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[time.Duration](),
		)).
		Placeholder("DURATION").
		Value()

	tf.IntVar(&cfg.MaxAttempts, "max-attempts", 0, "Maximum attempts before giving up. Set to 0 for endless retries.").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[int](),
		)).
		Placeholder("N").
		Value()

	tinyflags.EnumVar(tf, &cfg.LogFormat, "log-format", logging.LogFormatJSON, "Log format",
		logging.LogFormatJSON, logging.LogFormatText,
	).
		Value()
}
