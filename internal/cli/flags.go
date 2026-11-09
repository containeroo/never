package cli

import (
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/tinyflags"
)

const (
	defaultCheckInterval             time.Duration = 2 * time.Second
	defaultHTTPAllowDuplicateHeaders               = false
)

var (
	defaultHTTPFollowRedirects = checker.DefaultHTTPConfig().FollowRedirects
	defaultHTTPMaxRedirects    = checker.DefaultHTTPConfig().MaxRedirects
	defaultHTTPSkipTLSVerify   = checker.DefaultHTTPConfig().SkipTLSVerify
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
	// tinyflags inserts the separator underscore after this prefix.
	tf.EnvPrefix("NEVER_")
	tf.HideEnvs()
	tf.SortedFlags()
	tf.SortedGroups()

	tf.Note("\nFlags can also be set through environment variables with the NEVER__ prefix. " +
		"For example, --default-interval becomes NEVER__DEFAULT_INTERVAL and --http.web.address becomes NEVER__HTTP_WEB_ADDRESS.")

	registerAppFlags(tf, &cfg)
	registerHTTPFlags(tf)
	registerTCPFlags(tf)
	registerICMPFlags(tf)

	if err := tf.Parse(args); err != nil {
		return nil, err
	}

	targets, err := parseTargetConfigs(tf.DynamicGroups(), version, cfg.MaxAttempts)
	if err != nil {
		return nil, err
	}
	cfg.Targets = targets

	return &cfg, nil
}
