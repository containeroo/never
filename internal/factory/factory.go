package factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/dynflags"
	"github.com/containeroo/httputils"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/resolver"
)

// CheckerWithInterval represents a checker with its interval.
type CheckerWithInterval struct {
	Interval time.Duration
	Checker  checker.Checker
}

// BuildCheckers creates a list of CheckerWithInterval from the parsed dynflags configuration.
func BuildCheckers(dynFlags *dynflags.DynFlags, defaultInterval time.Duration) ([]CheckerWithInterval, error) {
	var checkers []CheckerWithInterval

	// Iterate over all parsed groups
	for parentName, childGroups := range dynFlags.Parsed().Groups() {
		checkType, err := checker.ParseCheckType(parentName)
		if err != nil {
			return nil, fmt.Errorf("invalid check type '%s': %w", parentName, err)
		}

		// Process each parsed group (child) under the parent group
		for _, group := range childGroups {
			address, err := group.GetString("address")
			if err != nil {
				return nil, fmt.Errorf("missing address for %s checker: %w", parentName, err)
			}

			resolvedAddress, err := resolver.ResolveVariable(address)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve variable in address: %w", err)
			}

			// Default interval for the checker
			interval := defaultInterval
			if customInterval, err := group.GetDuration("interval"); err == nil {
				interval = customInterval
			}

			// Prepare options based on the checker type
			var opts []checker.Option

			switch checkType {
			case checker.HTTP:
				if method, err := group.GetString("method"); err == nil {
					opts = append(opts, checker.WithHTTPMethod(method))
				}

				allowDuplicateHeaders, _ := group.GetBool("allow-duplicate-headers") // Type is checked when parsing
				if headers, err := group.GetStringSlices("header"); err == nil {
					headersMap, err := createHTTPHeadersMap(headers, allowDuplicateHeaders)
					if err != nil {
						return nil, fmt.Errorf("invalid \"--%s.%s.header\": %w", parentName, group.Name, err)
					}
					opts = append(opts, checker.WithHTTPHeaders(headersMap))
				}

				if allowedStatusCodes, err := group.GetString("expected-status-codes"); err == nil {
					statusCodes, err := httputils.ParseStatusCodes(allowedStatusCodes)
					if err != nil {
						return nil, fmt.Errorf("invalid \"--%s.%s.expected-status-codes\": %w", parentName, group.Name, err)
					}

					opts = append(opts, checker.WithExpectedStatusCodes(statusCodes))
				}

				if skipTLS, err := group.GetBool("skip-tls-verify"); err == nil {
					opts = append(opts, checker.WithHTTPSkipTLSVerify(skipTLS))
				}

				if timeout, err := group.GetDuration("timeout"); err == nil {
					opts = append(opts, checker.WithHTTPTimeout(timeout))
				}

			case checker.TCP:
				if timeout, err := group.GetDuration("timeout"); err == nil {
					opts = append(opts, checker.WithHTTPTimeout(timeout)) // Could have a TCP-specific timeout option
				}

			case checker.ICMP:
				if readTimeout, err := group.GetDuration("read-timeout"); err == nil {
					opts = append(opts, checker.WithICMPReadTimeout(readTimeout))
				}
				if writeTimeout, err := group.GetDuration("write-timeout"); err == nil {
					opts = append(opts, checker.WithICMPWriteTimeout(writeTimeout))
				}
			}

			name, _ := group.GetString("name")
			if name == "" {
				name = group.Name
			}

			instance, err := checker.NewChecker(checkType, name, resolvedAddress, opts...)
			if err != nil {
				return nil, fmt.Errorf("failed to create %s checker: %w", parentName, err)
			}

			// Wrap the checker with its interval and add to the list
			checkers = append(checkers, CheckerWithInterval{
				Interval: interval,
				Checker:  instance,
			})
		}
	}

	return checkers, nil
}

// createHTTPHeadersMap creates a map or slice-based map of HTTP headers from a slice of strings.
// If allowDuplicateHeaders is true, headers with the same key will be overwritten.
func createHTTPHeadersMap(headers []string, allowDuplicateHeaders bool) (map[string]string, error) {
	if headers == nil {
		return nil, fmt.Errorf("headers cannot be nil")
	}

	headersMap := make(map[string]string)

	for _, header := range headers {
		parts := strings.SplitN(header, "=", 2)

		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid header format: %q", header)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		resolved, err := resolver.ResolveVariable(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve variable in header: %w", err)
		}

		if _, exists := headersMap[key]; exists && !allowDuplicateHeaders {
			return nil, fmt.Errorf("duplicate header: %q", header)
		}

		headersMap[key] = resolved
	}

	return headersMap, nil
}
