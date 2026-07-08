package cli

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/containeroo/httputils"
	"github.com/containeroo/never/internal/utils"
)

// validateHTTPAddress validates HTTP target addresses and resolvable values.
func validateHTTPAddress(s string) error {
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
}

// validateICMPAddress validates ICMP target hostnames, IPs, and resolvable values.
func validateICMPAddress(s string) error {
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

	if strings.Contains(s, "://") {
		return errors.New("ICMP check cannot have a scheme")
	}
	if strings.Contains(s, "/") || strings.Contains(s, ":") {
		return errors.New("ICMP address must be a hostname or IP without path or port")
	}
	if !utils.IsHostnameLike(s) {
		return fmt.Errorf("invalid hostname: %q", s)
	}

	return nil
}

// validateTCPAddress validates TCP target host:port values and resolvable values.
func validateTCPAddress(s string) error {
	s = strings.TrimSpace(s)
	if utils.IsResolvableValue(s) {
		return nil
	}
	if _, _, err := net.SplitHostPort(s); err != nil {
		return fmt.Errorf("TCP address must be host:port (e.g. 127.0.0.1:80): %w", err)
	}

	return nil
}

// validateHTTPStatusCodes validates a single status code value or range expression.
func validateHTTPStatusCodes(codes string) error {
	for _, code := range strings.Split(codes, ",") {
		_, err := httputils.ParseStatusCodes(code)
		if err != nil {
			return fmt.Errorf("invalid status code: %w", err)
		}
	}

	return nil
}

// validateMaxAttempts validates the global max-attempts flag.
func validateMaxAttempts(v int) error {
	if v == 0 {
		return errors.New("max-attempts must be -1 or positive")
	}
	if v < -1 {
		return errors.New("max-attempts must be -1 or positive")
	}

	return nil
}

// validateOptionalMaxAttempts validates per-target max-attempts where zero means inherit global.
func validateOptionalMaxAttempts(v int) error {
	if v == 0 {
		return nil
	}

	return validateMaxAttempts(v)
}

// validatePositiveDuration returns a validator that requires a duration greater than zero.
func validatePositiveDuration(name string) func(time.Duration) error {
	return func(d time.Duration) error {
		if d <= 0 {
			return fmt.Errorf("%s must be positive", name)
		}

		return nil
	}
}

// validateNonNegativeDuration returns a validator that rejects negative durations.
func validateNonNegativeDuration(name string) func(time.Duration) error {
	return func(d time.Duration) error {
		if d < 0 {
			return fmt.Errorf("%s must be non-negative", name)
		}

		return nil
	}
}
