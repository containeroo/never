package flag

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/containeroo/tinyflags"
)

const (
	paramDefaultInterval             string        = "default-interval"
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
	DynamicGroups        []*tinyflags.DynamicGroup
}

// ParseFlags parses command-line arguments and returns the parsed flags.
func ParseFlags(args []string, version string) (*ParsedFlags, error) {
	tf := tinyflags.NewFlagSet("never", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.SortedFlags()
	tf.SortedGroups()

	interval := tf.Duration(
		paramDefaultInterval,
		defaultCheckInterval,
		"Default interval between checks. Can be overridden for each target.",
	).
		Placeholder("DURATION").
		Value()

	tf.Note("\nFor more information, see https://github.com/containeroo/never")

	// HTTP flags
	http := tf.DynamicGroup("http").Title("HTTP")
	http.String("name", "", "Name of the HTTP checker. Defaults to <ID>.")
	http.String("method", "GET", "HTTP method to use")
	http.String("address", "", "HTTP target URL").
		Validate(func(s string) error {
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
	http.Duration("interval", 1*time.Second, "Time between HTTP requests. Can be overwritten with --default-interval.").
		Placeholder("DURATION")
	http.StringSlice("header", []string{}, "HTTP headers to send")
	http.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	http.String("expected-status-codes", "200", "Expected HTTP status codes").
		Placeholder("CODE")
	http.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	http.Duration("timeout", 2*time.Second, "Timeout in seconds").
		Placeholder("DURATION")

	// ICMP flags
	icmp := tf.DynamicGroup("icmp").Title("ICMP")
	icmp.String("name", "", "Name of the ICMP checker. Defaults to <ID>.")
	icmp.String("address", "", "ICMP target address").
		Validate(func(s string) error {
			if ip := net.ParseIP(s); ip != nil {
				return nil
			}
			u, err := url.Parse(s)
			if err != nil || u.Host == "" {
				return fmt.Errorf("invalid URL: %q", s)
			}
			if u.Scheme != "" {
				return errors.New("ICMP check cannot have a scheme")
			}
			return nil
		}).
		Required()
	icmp.Duration("interval", 1*time.Second, "Time between ICMP requests. Can be overwritten with --default-interval.").
		Placeholder("DURATION")
	icmp.Duration("read-timeout", 2*time.Second, "Timeout for ICMP read").
		Placeholder("DURATION")
	icmp.Duration("write-timeout", 2*time.Second, "Timeout for ICMP write").
		Placeholder("DURATION")

	// TCP flags
	tcp := tf.DynamicGroup("tcp").Title("TCP")
	tcp.String("name", "", "Name of the TCP checker. Defaults to <ID>.")
	tcp.String("address", "", "TCP target address").
		Validate(func(s string) error {
			if _, _, err := net.SplitHostPort(s); err != nil {
				return fmt.Errorf("TCP address must be host:port (e.g. 127.0.0.1:80): %w", err)
			}
			return nil
		}).
		Required()
	tcp.Duration("timeout", 2*time.Second, "Timeout for TCP connection").
		Placeholder("DURATION")
	tcp.Duration("interval", 1*time.Second, "Time between TCP requests. Can be overwritten with --default-interval.").
		Placeholder("DURATION")

	// Parse unknown arguments with dynamic flags
	if err := tf.Parse(args); err != nil {
		return nil, err
	}

	return &ParsedFlags{
		DefaultCheckInterval: *interval,
		DynamicGroups:        tf.DynamicGroups(),
	}, nil
}
