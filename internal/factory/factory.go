package factory

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/utils"

	"github.com/containeroo/httputils"
	"github.com/containeroo/resolver"
)

// TargetConfig describes one checker target after CLI parsing.
type TargetConfig struct {
	ID          string
	Type        checker.CheckType
	Name        string
	Address     string
	Interval    time.Duration
	MaxAttempts int
	Backoff     backoff.Mode
	MaxInterval time.Duration

	HTTPMethod                string
	HTTPHeaders               []string
	HTTPAllowDuplicateHeaders bool
	HTTPExpectedStatusCodes   []string
	HTTPSkipTLSVerify         bool
	HTTPTimeout               time.Duration

	TCPTimeout time.Duration

	ICMPTimeout      time.Duration
	ICMPReadTimeout  time.Duration
	ICMPWriteTimeout time.Duration
}

// CheckerWithInterval represents a checker with its interval.
type CheckerWithInterval struct {
	Interval time.Duration
	Checker  checker.Checker
	// MaxAttempts == 0 means use global; -1 means endless retries.
	MaxAttempts int
	Backoff     backoff.Mode
	MaxInterval time.Duration
}

// BuildCheckers creates a list of CheckerWithInterval from typed target configuration.
func BuildCheckers(targets []TargetConfig, defaultInterval time.Duration) ([]CheckerWithInterval, error) {
	checkers := make([]CheckerWithInterval, 0, len(targets))

	for _, target := range targets {
		resolvedAddr, err := resolver.ResolveVariable(target.Address)
		if err != nil {
			return nil, fmt.Errorf("invalid variable in address: %w", err)
		}

		interval := utils.DefaultIfZero(target.Interval, defaultInterval)
		backoffMode := utils.DefaultIfZero(target.Backoff, backoff.ModeLinear)
		name := utils.DefaultIfZero(target.Name, target.ID)

		var opts []checker.Option

		switch target.Type {
		case checker.HTTP:
			if target.HTTPMethod != "" {
				opts = append(opts, checker.WithHTTPMethod(target.HTTPMethod))
			}

			headersMap, err := createHTTPHeadersMap(target.HTTPHeaders, target.HTTPAllowDuplicateHeaders)
			if err != nil {
				return nil, fmt.Errorf("invalid \"--%s.%s.header\": %w", strings.ToLower(target.Type.String()), target.ID, err)
			}
			opts = append(opts, checker.WithHTTPHeaders(headersMap))

			if len(target.HTTPExpectedStatusCodes) > 0 {
				// Status codes can be passed multiple times and contain ranges.
				// Get all status codes as a slice. Can produce something like []string{"200-299", "300", "301"}.
				codes, err := httputils.ParseStatusCodes(strings.Join(target.HTTPExpectedStatusCodes, ","))
				if err != nil {
					return nil, fmt.Errorf("invalid --%s.%s.expected-status-codes: %w", strings.ToLower(target.Type.String()), target.ID, err)
				}
				opts = append(opts, checker.WithExpectedStatusCodes(codes))
			}

			opts = append(opts, checker.WithHTTPSkipTLSVerify(target.HTTPSkipTLSVerify))

			if target.HTTPTimeout > 0 {
				opts = append(opts, checker.WithHTTPTimeout(target.HTTPTimeout))
			}

		case checker.TCP:
			if target.TCPTimeout > 0 {
				opts = append(opts, checker.WithTCPTimeout(target.TCPTimeout))
			}

		case checker.ICMP:
			if target.ICMPTimeout > 0 {
				opts = append(opts, checker.WithICMPTimeout(target.ICMPTimeout))
			}

			if target.ICMPReadTimeout > 0 {
				opts = append(opts, checker.WithICMPReadTimeout(target.ICMPReadTimeout))
			}

			if target.ICMPWriteTimeout > 0 {
				opts = append(opts, checker.WithICMPWriteTimeout(target.ICMPWriteTimeout))
			}

		default:
			return nil, fmt.Errorf("unsupported check type: %s", target.Type)
		}

		instance, err := checker.NewChecker(target.Type, name, resolvedAddr, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s checker: %w", target.Type, err)
		}

		checkers = append(checkers, CheckerWithInterval{
			Interval:    interval,
			Checker:     instance,
			MaxAttempts: target.MaxAttempts,
			Backoff:     backoffMode,
			MaxInterval: target.MaxInterval,
		})
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
