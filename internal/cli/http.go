package cli

import (
	"net/http"
	"strconv"
	"time"

	"github.com/containeroo/never/internal/checker"

	"github.com/containeroo/httputils"
	"github.com/containeroo/tinyflags"
)

// registerHTTPFlags registers HTTP-related flags.
func registerHTTPFlags(tf *tinyflags.FlagSet) {
	httpGroup := tf.DynamicGroup("http").Title("HTTP")
	httpGroup.String("name", "", "Name of the HTTP checker. Defaults to <ID>.")
	tinyflags.DynamicEnum(
		httpGroup, "method", checker.DefaultHTTPConfig().Method, "HTTP method to use.",
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
	tinyflags.DynamicEnum(
		httpGroup,
		"address-detail",
		checker.HTTPAddressOrigin,
		"HTTP address detail shown in logs (query and full may expose secrets).",
		checker.HTTPAddressOrigin,
		checker.HTTPAddressPath,
		checker.HTTPAddressQuery,
		checker.HTTPAddressFull,
	).
		Placeholder("DETAIL")
	httpGroup.Duration("interval", 0*time.Second, "Time between HTTP requests. Defaults to --default-interval when unset or 0.").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[time.Duration](),
		)).
		Placeholder("DURATION")
	httpGroup.Int("max-attempts", 0, "Maximum attempts before giving up. Inherits --max-attempts when unset; 0 means endless retries.").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[int](),
		)).
		Placeholder("N")
	registerRetryFlags(httpGroup)
	// Headers are repeated flags; a newline delimiter preserves commas inside header values.
	httpGroup.StringSlice("header", []string{}, "HTTP headers to send").
		Delimiter("\n").
		Placeholder("KEY=VALUE")
	httpGroup.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	tinyflags.DynamicSlice(
		httpGroup,
		"expected-status-codes",
		[]int{http.StatusOK},
		"Expected HTTP status codes. Comma-separated list of status codes, ranges possible (eg \"200-299\", \"300,301\")",
		httputils.ParseStatusCodes,
		strconv.Itoa,
	).
		Placeholder("CODES...")
	httpGroup.Bool("follow-redirects", defaultHTTPFollowRedirects, "Follow HTTP redirects").Strict()
	httpGroup.Int("max-redirects", defaultHTTPMaxRedirects, "Maximum number of redirects to follow. Set to 0 to disable redirects.").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[int](),
		)).
		Placeholder("N")

	httpGroup.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	httpGroup.Duration("timeout", checker.DefaultHTTPConfig().Timeout, "Request timeout").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[time.Duration](),
		)).
		Placeholder("DURATION")
}
