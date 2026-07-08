package cli

import (
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/tinyflags"
)

// registerAppFlags registers core application flags and binds them to cfg.
func registerAppFlags(tf *tinyflags.FlagSet, cfg *Config) {
	tf.DurationVar(&cfg.DefaultCheckInterval, "default-interval", defaultCheckInterval, "Default interval between checks. Can be overridden for each target.").
		Placeholder("DURATION").
		Value()

	tf.IntVar(&cfg.MaxAttempts, "max-attempts", -1, "Maximum attempts before giving up (-1 for endless).").
		Validate(validateMaxAttempts).
		Placeholder("N").
		Value()

	tinyflags.EnumVar(
		tf,
		&cfg.LogFormat,
		"log-format",
		logging.LogFormatJSON,
		"Log format",
		logging.LogFormatJSON,
		logging.LogFormatText,
	).Value()
}
