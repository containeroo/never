package factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/httputils"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/resolver"
	"github.com/containeroo/tinyflags"
)

// CheckerWithInterval represents a checker with its interval.
type CheckerWithInterval struct {
	Interval time.Duration
	Checker  checker.Checker
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

				codeStr := tinyflags.GetOrDefaultDynamic[string](group, id, "expected-status-codes")
				codes, err := httputils.ParseStatusCodes(codeStr)
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
				opts = append(opts, checker.WithHTTPTimeout(timeout))

			case checker.ICMP:
				rt := tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "read-timeout")
				opts = append(opts, checker.WithICMPReadTimeout(rt))

				wt := tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "write-timeout")
				opts = append(opts, checker.WithICMPWriteTimeout(wt))
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
