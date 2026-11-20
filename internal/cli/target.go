package cli

import (
	"cmp"
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/httputils"
	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/resolver"
	"github.com/containeroo/tinyflags"
)

// parseTargetConfigs converts parsed dynamic flag groups into typed target config.
func parseTargetConfigs(
	dynamicGroups []*tinyflags.DynamicGroup,
	version string,
	defaultMaxAttempts int,
) ([]factory.TargetConfig, error) {
	var targets []factory.TargetConfig

	for _, group := range dynamicGroups {
		for _, id := range group.Instances() {
			checkerConfig, address, err := parseCheckerConfig(group, id, version)
			if err != nil {
				return nil, fmt.Errorf("%s target %q: %w", strings.ToUpper(group.Name()), id, err)
			}

			maxAttempts := defaultMaxAttempts
			if value, err := tinyflags.GetDynamic[int](group, id, "max-attempts"); err == nil {
				maxAttempts = value
			}

			targets = append(targets, factory.TargetConfig{
				ID:          id,
				Name:        tinyflags.GetOrDefaultDynamic[string](group, id, "name"),
				Address:     address,
				Interval:    getDynamicDuration(group, id, "interval"),
				MaxAttempts: maxAttempts,
				Backoff:     getDynamicBackoffMode(group, id, "backoff"),
				MaxInterval: getDynamicDuration(group, id, "max-interval"),
				Config:      checkerConfig,
			})
		}
	}

	return targets, nil
}

// parseCheckerConfig converts one protocol's parsed flags into its checker config.
func parseCheckerConfig(
	group *tinyflags.DynamicGroup,
	id string,
	version string,
) (checker.Config, string, error) {
	rawAddress := tinyflags.GetOrDefaultDynamic[string](group, id, "address")

	switch group.Name() {
	case "http":
		address, err := resolveTargetAddress(rawAddress, validateResolvedHTTPAddress)
		if err != nil {
			return nil, "", err
		}

		headers, err := httputils.ParseHeaders(
			tinyflags.GetOrDefaultDynamic[[]string](group, id, "header"),
			tinyflags.GetOrDefaultDynamic[bool](group, id, "allow-duplicate-headers"),
		)
		if err != nil {
			return nil, "", fmt.Errorf("invalid HTTP header: %w", err)
		}
		if err := resolveHTTPHeaderValues(headers); err != nil {
			return nil, "", fmt.Errorf("invalid HTTP header: %w", err)
		}

		return checker.HTTPConfig{
			Method:              tinyflags.GetOrDefaultDynamic[string](group, id, "method"),
			Headers:             headers,
			ExpectedStatusCodes: tinyflags.GetOrDefaultDynamic[[]int](group, id, "expected-status-codes"),
			FollowRedirects:     tinyflags.GetOrDefaultDynamic[bool](group, id, "follow-redirects"),
			MaxRedirects:        tinyflags.GetOrDefaultDynamic[int](group, id, "max-redirects"),
			SkipTLSVerify:       tinyflags.GetOrDefaultDynamic[bool](group, id, "skip-tls-verify"),
			Timeout:             tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout"),
			UserAgent:           strings.TrimSuffix(httpUserAgentPrefix, "/") + "/" + version,
			AddressDetail:       tinyflags.GetOrDefaultDynamic[checker.HTTPAddressDetail](group, id, "address-detail"),
		}, address, nil
	case "tcp":
		address, err := resolveTargetAddress(rawAddress, validateResolvedTCPAddress)
		if err != nil {
			return nil, "", err
		}
		return checker.TCPConfig{
			Timeout: tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout"),
		}, address, nil
	case "icmp":
		address, err := resolveTargetAddress(rawAddress, validateResolvedICMPAddress)
		if err != nil {
			return nil, "", err
		}
		timeout := tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout")
		return checker.ICMPConfig{
			ReadTimeout:  cmp.Or(getDynamicDuration(group, id, "read-timeout"), timeout),
			WriteTimeout: cmp.Or(getDynamicDuration(group, id, "write-timeout"), timeout),
		}, address, nil
	default:
		return nil, "", fmt.Errorf("unsupported check type: %s", group.Name())
	}
}

// resolveTargetAddress resolves target address references and validates their concrete value.
// Literal addresses were already validated by the flag validator.
func resolveTargetAddress(value string, validate func(string) error) (string, error) {
	raw := strings.TrimSpace(value)
	address, err := resolver.ResolveVariable(raw)
	if err != nil {
		return "", fmt.Errorf("failed to resolve address: %w", err)
	}

	address = strings.TrimSpace(address)
	if resolver.CanResolve(raw) {
		if err := validate(address); err != nil {
			return "", err
		}
	}

	return address, nil
}

// getDynamicDuration returns a dynamic duration flag value or zero when unset.
func getDynamicDuration(group *tinyflags.DynamicGroup, id, name string) time.Duration {
	v, _ := tinyflags.GetDynamic[time.Duration](group, id, name)
	return v
}

// getDynamicBackoffMode returns the configured backoff mode or the zero value when unset.
func getDynamicBackoffMode(group *tinyflags.DynamicGroup, id, name string) backoff.Mode {
	return tinyflags.GetOrDefaultDynamic[backoff.Mode](group, id, name)
}
