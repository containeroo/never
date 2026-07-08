package cli

import (
	"net/http"
	"time"

	"github.com/containeroo/tinyflags"
)

// registerHTTPFlags registers HTTP-related flags and binds them to cfg.
func registerHTTPFlags(tf *tinyflags.FlagSet) {
	httpGroup := tf.DynamicGroup("http").Title("HTTP")
	httpGroup.String("name", "", "Name of the HTTP checker. Defaults to <ID>.")
	tinyflags.DynamicEnum(
		httpGroup, "method", http.MethodGet, "HTTP method to use.",
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
		Validate(validateHTTPAddress).
		Required()
	httpGroup.Duration("interval", 0*time.Second, "Time between HTTP requests. Defaults to --default-interval when unset or 0.").
		Validate(validateNonNegativeDuration("interval")).
		Placeholder("DURATION")
	httpGroup.Int("max-attempts", 0, "Maximum attempts before giving up. Defaults to --max-attempts when unset or 0.").
		Validate(validateOptionalMaxAttempts).
		Placeholder("N")
	registerRetryFlags(httpGroup)
	httpGroup.StringSlice("header", []string{}, "HTTP headers to send").
		Placeholder("KEY=VALUE)")
	httpGroup.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	httpGroup.StringSlice(
		"expected-status-codes",
		[]string{"200"},
		"Expected HTTP status codes. Comma-separated list of status codes, ranges possible (eg \"200-299\", \"300,301\")",
	).
		Validate(validateHTTPStatusCodes).
		Placeholder("CODES...")

	httpGroup.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	httpGroup.Duration("timeout", 2*time.Second, "Request timeout").
		Validate(validatePositiveDuration("timeout")).
		Placeholder("DURATION")
}
