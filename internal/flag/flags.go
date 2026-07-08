package flag

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/containeroo/httputils"
	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/never/internal/utils"
	"github.com/containeroo/tinyflags"
)

const (
	defaultCheckInterval             time.Duration = 2 * time.Second
	defaultHTTPAllowDuplicateHeaders bool          = false
	defaultHTTPSkipTLSVerify         bool          = false
)

// ParsedFlags holds the parsed command-line flags.
type ParsedFlags struct {
	ShowHelp             bool
	ShowVersion          bool
	Version              string
	DefaultCheckInterval time.Duration
	MaxAttempts          int
	LogFormat            string
	DynamicGroups        []*tinyflags.DynamicGroup
}

// ParseFlags parses command-line arguments and returns the parsed flags.
func ParseFlags(args []string, version string) (*ParsedFlags, error) {
	tf := tinyflags.NewFlagSet("never", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.SortedFlags()
	tf.SortedGroups()

	interval := tf.Duration("default-interval", defaultCheckInterval, "Default interval between checks. Can be overridden for each target.").
		Placeholder("DURATION").
		Value()

	maxAttempts := tf.Int("max-attempts", -1, "Maximum attempts before giving up (-1 for endless).").
		Validate(validateMaxAttempts).
		Placeholder("N").
		Value()

	logFormat := tf.Enum("log-format", logging.FormatText, "Log output format.", logging.FormatText, logging.FormatJSON).
		Placeholder("FORMAT").
		Value()

	tf.Note("\nFor more information, see https://github.com/containeroo/never")

	// HTTP flags
	httpGroup := tf.DynamicGroup("http").Title("HTTP")
	httpGroup.String("name", "", "Name of the HTTP checker. Defaults to <ID>.")
	tinyflags.DynamicEnum(httpGroup, "method", http.MethodGet, "HTTP method to use.",
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	).
		Placeholder("METHOD")
	httpGroup.String("address", "", "HTTP target URL").
		Validate(func(s string) error {
			s = strings.TrimSpace(s)
			if utils.IsResolvableValue(s) {
				return nil
			}
			u, err := url.Parse(s)
			if err != nil || u.Host == "" {
				return fmt.Errorf("invalid URL: %q", s)
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				return fmt.Errorf("unsupported scheme: %q", u.Scheme)
			}
			return nil
		}).
		Required()
	httpGroup.Duration("interval", 0*time.Second, "Time between HTTP requests. Defaults to --default-interval when unset or 0.").
		Validate(func(d time.Duration) error {
			if d < 0 {
				return errors.New("interval must be non-negative")
			}
			return nil
		}).
		Placeholder("DURATION")
	httpGroup.Int("max-attempts", 0, "Maximum attempts before giving up. Defaults to --max-attempts when unset or 0.").
		Validate(func(v int) error {
			if v == 0 {
				return nil
			}
			return validateMaxAttempts(v)
		}).
		Placeholder("N")
	registerRetryFlags(httpGroup)
	httpGroup.StringSlice("header", []string{}, "HTTP headers to send").
		Placeholder("KEY=VALUE)")
	httpGroup.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	httpGroup.StringSlice("expected-status-codes",
		[]string{"200"},
		"Expected HTTP status codes. Comma-separated list of status codes, ranges possible (eg \"200-299\", \"300,301\")",
	).
		Validate(func(codes string) error {
			for _, code := range strings.Split(codes, ",") {
				_, err := httputils.ParseStatusCodes(code)
				if err != nil {
					return fmt.Errorf("invalid status code: %w", err)
				}
			}
			return nil
		}).
		Placeholder("CODES...")

	httpGroup.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	httpGroup.Duration("timeout", 2*time.Second, "Request timeout").
		Validate(func(d time.Duration) error {
			if d <= 0 {
				return errors.New("timeout must be positive")
			}
			return nil
		}).
		Placeholder("DURATION")

	// ICMP flags
	icmp := tf.DynamicGroup("icmp").Title("ICMP")
	icmp.String("name", "", "Name of the ICMP checker. Defaults to <ID>.")
	icmp.String("address", "", "ICMP target address").
		Validate(func(s string) error {
			s = strings.TrimSpace(s)
			if s == "" {
				return errors.New("ICMP address cannot be empty")
			}
			if utils.IsResolvableValue(s) {
				return nil
			}
			if ip := net.ParseIP(s); ip != nil {
				return nil
			}
			u, err := url.Parse(s)
			if err != nil {
				return fmt.Errorf("invalid address: %q", s)
			}
			if u.Scheme != "" {
				return errors.New("ICMP check cannot have a scheme")
			}
			if strings.Contains(s, "/") || strings.Contains(s, ":") {
				return errors.New("ICMP address must be a hostname or IP without path or port")
			}
			if !utils.IsHostnameLike(s) {
				return fmt.Errorf("invalid hostname: %q", s)
			}
			return nil
		}).
		Required()
	icmp.Duration("interval", 0*time.Second, "Time between ICMP requests. Defaults to --default-interval when unset or 0.").
		Validate(func(d time.Duration) error {
			if d < 0 {
				return errors.New("interval must be non-negative")
			}
			return nil
		}).
		Placeholder("DURATION")
	icmp.Int("max-attempts", 0, "Maximum attempts before giving up. Defaults to --max-attempts when unset or 0.").
		Validate(func(v int) error {
			if v == 0 {
				return nil
			}
			return validateMaxAttempts(v)
		}).
		Placeholder("N")
	registerRetryFlags(icmp)
	icmp.Duration("timeout", 2*time.Second, "Timeout for ICMP read and write").
		Validate(func(d time.Duration) error {
			if d <= 0 {
				return errors.New("timeout must be positive")
			}
			return nil
		}).
		Placeholder("DURATION")
	icmp.Duration("read-timeout", 0*time.Second, "Advanced: override the ICMP read timeout. Defaults to --icmp.<ID>.timeout when unset or 0.").
		Validate(func(d time.Duration) error {
			if d < 0 {
				return errors.New("read-timeout must be non-negative")
			}
			return nil
		}).
		Placeholder("DURATION")
	icmp.Duration("write-timeout", 0*time.Second, "Advanced: override the ICMP write timeout. Defaults to --icmp.<ID>.timeout when unset or 0.").
		Validate(func(d time.Duration) error {
			if d < 0 {
				return errors.New("write-timeout must be non-negative")
			}
			return nil
		}).
		Placeholder("DURATION")

	// TCP flags
	tcp := tf.DynamicGroup("tcp").Title("TCP")
	tcp.String("name", "", "Name of the TCP checker. Defaults to <ID>.")
	tcp.String("address", "", "TCP target address").
		Validate(func(s string) error {
			s = strings.TrimSpace(s)
			if utils.IsResolvableValue(s) {
				return nil
			}
			if _, _, err := net.SplitHostPort(s); err != nil {
				return fmt.Errorf("TCP address must be host:port (e.g. 127.0.0.1:80): %w", err)
			}
			return nil
		}).
		Required()
	tcp.Duration("timeout", 2*time.Second, "Timeout for TCP connection").
		Validate(func(d time.Duration) error {
			if d <= 0 {
				return errors.New("timeout must be positive")
			}
			return nil
		}).
		Placeholder("DURATION")
	tcp.Duration("interval", 0*time.Second, "Time between TCP requests. Defaults to --default-interval when unset or 0.").
		Validate(func(d time.Duration) error {
			if d < 0 {
				return errors.New("interval must be non-negative")
			}
			return nil
		}).
		Placeholder("DURATION")
	tcp.Int("max-attempts", 0, "Maximum attempts before giving up. Defaults to --max-attempts when unset or 0.").
		Validate(func(v int) error {
			if v == 0 {
				return nil
			}
			return validateMaxAttempts(v)
		}).
		Placeholder("N")
	registerRetryFlags(tcp)

	// Parse unknown arguments with dynamic flags
	if err := tf.Parse(args); err != nil {
		return nil, err
	}

	return &ParsedFlags{
		DefaultCheckInterval: *interval,
		MaxAttempts:          *maxAttempts,
		LogFormat:            *logFormat,
		DynamicGroups:        tf.DynamicGroups(),
	}, nil
}

// validateMaxAttempts validates the max-attempts flag.
func validateMaxAttempts(v int) error {
	if v == 0 {
		return errors.New("max-attempts must be -1 or positive")
	}
	if v < -1 {
		return errors.New("max-attempts must be -1 or positive")
	}
	return nil
}

func registerRetryFlags(group *tinyflags.DynamicGroup) {
	tinyflags.DynamicEnum(group, "backoff", backoff.ModeNone, "Retry backoff mode.", backoff.ModeNone, backoff.ModeExponential).
		Placeholder("MODE")
	group.Duration("max-interval", 0*time.Second, "Maximum retry interval when backoff increases the delay. Defaults to uncapped when unset or 0.").
		Validate(func(d time.Duration) error {
			if d < 0 {
				return errors.New("max-interval must be non-negative")
			}
			return nil
		}).
		Placeholder("DURATION")
}
