package cli

import (
	"time"

	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/tinyflags"
)

const (
	defaultCheckInterval             time.Duration = 2 * time.Second
	defaultHTTPAllowDuplicateHeaders bool          = false
	defaultHTTPSkipTLSVerify         bool          = false
)

// Config holds the parsed command-line configuration.
type Config struct {
	ShowHelp             bool
	ShowVersion          bool
	Version              string
	DefaultCheckInterval time.Duration
	MaxAttempts          int
	LogFormat            logging.LogFormat
	Targets              []factory.TargetConfig
}

// ParseFlags parses command-line arguments and returns application configuration.
func ParseFlags(args []string, version string) (*Config, error) {
	cfg := Config{}

	tf := tinyflags.NewFlagSet("never", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.SortedFlags()
	tf.SortedGroups()

	tf.Note("\nFor more information, see https://github.com/containeroo/never")

	registerAppFlags(tf, &cfg)
	registerHTTPFlags(tf, &cfg)
	registerTCPFlags(tf, &cfg)
	registerICMPFlags(tf, &cfg)

	if err := tf.Parse(args); err != nil {
		return nil, err
	}

	targets, err := parseTargetConfigs(tf.DynamicGroups())
	if err != nil {
		return nil, err
	}
	cfg.Targets = targets

	return &cfg, nil
}
