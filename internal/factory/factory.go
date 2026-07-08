package factory

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/containeroo/httputils"
	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/resolver"
	"github.com/containeroo/tinyflags"
)

// CheckerWithInterval represents a checker with its interval.
type CheckerWithInterval struct {
	Interval time.Duration
	Checker  checker.Checker
	// MaxAttempts == 0 means use global; -1 means endless retries.
	MaxAttempts int
	Backoff     backoff.Mode
	MaxInterval time.Duration
}

// BuildCheckers creates a list of CheckerWithInterval from the parsed dynflags configuration.
func BuildCheckers(dynamicGroups []*tinyflags.DynamicGroup, defaultInterval time.Duration) ([]CheckerWithInterval, error) {
	var checkers []CheckerWithInterval

	for _, group := range dynamicGroups {
		checkType, err := checker.ParseCheckType(group.Name())
		if err != nil {
			return nil, err
		}

		for _, id := range group.Instances() {
			address := tinyflags.GetOrDefaultDynamic[string](group, id, "address")
			resolvedAddr, err := resolver.ResolveVariable(address)
			if err != nil {
				return nil, fmt.Errorf("invalid variable in address: %w", err)
			}

			interval := defaultInterval
			if v, _ := tinyflags.GetDynamic[time.Duration](group, id, "interval"); v != 0 {
				interval = v
			}
			maxAttempts := 0
			if v, _ := tinyflags.GetDynamic[int](group, id, "max-attempts"); v != 0 {
				maxAttempts = v
			}

			backoffMode := tinyflags.GetOrDefaultDynamic[backoff.Mode](group, id, "backoff")
			if backoffMode == "" {
				backoffMode = backoff.ModeNone
			}

			maxInterval := time.Duration(0)
			if v, _ := tinyflags.GetDynamic[time.Duration](group, id, "max-interval"); v != 0 {
				maxInterval = v
			}

			var opts []checker.Option

			switch checkType {
			case checker.HTTP:
				method := tinyflags.GetOrDefaultDynamic[string](group, id, "method")
				opts = append(opts, checker.WithHTTPMethod(method))

				headers := tinyflags.GetOrDefaultDynamic[[]string](group, id, "header")
				allowDup := tinyflags.GetOrDefaultDynamic[bool](group, id, "allow-duplicate-headers")
				headersMap, err := createHTTPHeadersMap(headers, allowDup)
				if err != nil {
					return nil, fmt.Errorf("invalid \"--%s.%s.header\": %w", group.Name(), id, err)
				}
				opts = append(opts, checker.WithHTTPHeaders(headersMap))

				// Status codes can be passed multiple times and containt ranges.
				// Get all status codes as a slices. Can produce something like []string{"200-299", "300", "301"}
				codeStr := tinyflags.GetOrDefaultDynamic[[]string](group, id, "expected-status-codes")
				codes, err := httputils.ParseStatusCodes(strings.Join(codeStr, ",")) // Join the slice again so ParseStatusCodes can parse it.
				if err != nil {
					return nil, fmt.Errorf("invalid --%s.%s.expected-status-codes: %w", group.Name(), id, err)
				}
				opts = append(opts, checker.WithExpectedStatusCodes(codes))

				skipTLS := tinyflags.GetOrDefaultDynamic[bool](group, id, "skip-tls-verify")
				opts = append(opts, checker.WithHTTPSkipTLSVerify(skipTLS))

				timeout := tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout")
				opts = append(opts, checker.WithHTTPTimeout(timeout))

			case checker.TCP:
				timeout := tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout")
				opts = append(opts, checker.WithTCPTimeout(timeout))

			case checker.ICMP:
				timeout := tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout")
				opts = append(opts, checker.WithICMPTimeout(timeout))

				if rt, _ := tinyflags.GetDynamic[time.Duration](group, id, "read-timeout"); rt != 0 {
					opts = append(opts, checker.WithICMPReadTimeout(rt))
				}

				if wt, _ := tinyflags.GetDynamic[time.Duration](group, id, "write-timeout"); wt != 0 {
					opts = append(opts, checker.WithICMPWriteTimeout(wt))
				}
			}

			name := id
			if n := tinyflags.GetOrDefaultDynamic[string](group, id, "name"); n != "" {
				name = n
			}

			instance, err := checker.NewChecker(checkType, name, resolvedAddr, opts...)
			if err != nil {
				return nil, fmt.Errorf("failed to create %s checker: %w", checkType, err)
			}

			checkers = append(checkers, CheckerWithInterval{
				Interval:    interval,
				Checker:     instance,
				MaxAttempts: maxAttempts,
				Backoff:     backoffMode,
				MaxInterval: maxInterval,
			})
		}
	}

	return checkers, nil
}

// createHTTPHeadersMap creates an http.Header from a slice of key=value strings.
// If allowDuplicateHeaders is true, headers with the same key will be appended.
func createHTTPHeadersMap(headers []string, allowDuplicateHeaders bool) (http.Header, error) {
	headersMap := make(http.Header)

	if headers == nil {
		return headersMap, nil
	}

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

		if allowDuplicateHeaders {
			headersMap.Add(key, resolved)
			continue
		}

		headersMap.Set(key, resolved)
	}

	return headersMap, nil
}
