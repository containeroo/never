package cli

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// validateHTTPAddress validates HTTP target addresses and resolver references.
func validateHTTPAddress(s string) error {
	s = strings.TrimSpace(s)
	if isResolvableValue(s) {
		return nil
	}

	return validateResolvedHTTPAddress(s)
}

// validateResolvedHTTPAddress validates a concrete HTTP target URL.
func validateResolvedHTTPAddress(s string) error {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil || u.Host == "" {
		return errors.New("invalid HTTP URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %q", u.Scheme)
	}

	return nil
}

// validateICMPAddress validates ICMP target addresses and resolver references.
func validateICMPAddress(s string) error {
	s = strings.TrimSpace(s)
	if isResolvableValue(s) {
		return nil
	}

	return validateResolvedICMPAddress(s)
}

// validateResolvedICMPAddress validates a concrete ICMP hostname or IP.
func validateResolvedICMPAddress(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("ICMP address cannot be empty")
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
	if !isHostnameLike(s) {
		return fmt.Errorf("invalid hostname: %q", s)
	}

	return nil
}

// validateTCPAddress validates TCP target addresses and resolver references.
func validateTCPAddress(s string) error {
	s = strings.TrimSpace(s)
	if isResolvableValue(s) {
		return nil
	}

	return validateResolvedTCPAddress(s)
}

// validateResolvedTCPAddress validates a concrete TCP host:port address.
func validateResolvedTCPAddress(s string) error {
	s = strings.TrimSpace(s)
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return fmt.Errorf("TCP address must be host:port (e.g. 127.0.0.1:80): %w", err)
	}

	if host == "" || strings.ContainsAny(host, "/?#") {
		return errors.New("TCP address must contain a host")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return errors.New("TCP port must be between 1 and 65535")
	}

	return nil
}

// isResolvableValue reports whether s uses one of resolver's supported prefixes.
func isResolvableValue(s string) bool {
	for _, prefix := range [...]string{"env:", "file:", "json:", "yaml:", "ini:"} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	return false
}

// isHostnameLike reports whether s is an ASCII hostname without scheme, port, or path.
func isHostnameLike(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}

	for label := range strings.SplitSeq(s, ".") {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		for i := range len(label) {
			ch := label[i]
			if ch == '-' {
				if i == 0 || i == len(label)-1 {
					return false
				}
				continue
			}
			if !isAlphaNum(ch) {
				return false
			}
		}
	}

	return true
}

// isAlphaNum reports whether ch is an ASCII letter or digit.
func isAlphaNum(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ('0' <= ch && ch <= '9')
}
