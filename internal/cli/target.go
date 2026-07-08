package cli

import (
	"net/http"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/tinyflags"
)

// parseTargetConfigs converts parsed dynamic flag groups into typed target config.
func parseTargetConfigs(dynamicGroups []*tinyflags.DynamicGroup) ([]factory.TargetConfig, error) {
	var targets []factory.TargetConfig

	for _, group := range dynamicGroups {
		checkType, err := checker.ParseCheckType(group.Name())
		if err != nil {
			return nil, err
		}

		for _, id := range group.Instances() {
			target := factory.TargetConfig{
				ID:          id,
				Type:        checkType,
				Name:        id,
				Address:     tinyflags.GetOrDefaultDynamic[string](group, id, "address"),
				Interval:    getDynamicDuration(group, id, "interval"),
				MaxAttempts: getDynamicInt(group, id, "max-attempts"),
				Backoff:     getDynamicBackoffMode(group, id, "backoff"),
				MaxInterval: getDynamicDuration(group, id, "max-interval"),
			}

			if name := tinyflags.GetOrDefaultDynamic[string](group, id, "name"); name != "" {
				target.Name = name
			}

			applyTargetTypeConfig(&target, group, id, checkType)
			targets = append(targets, target)
		}
	}

	return targets, nil
}

// applyTargetTypeConfig fills target fields that are specific to the checker type.
func applyTargetTypeConfig(target *factory.TargetConfig, group *tinyflags.DynamicGroup, id string, checkType checker.CheckType) {
	switch checkType {
	case checker.HTTP:
		target.HTTPMethod = tinyflags.GetOrDefaultDynamic[string](group, id, "method")
		if target.HTTPMethod == "" {
			target.HTTPMethod = http.MethodGet
		}
		target.HTTPHeaders = tinyflags.GetOrDefaultDynamic[[]string](group, id, "header")
		target.HTTPAllowDuplicateHeaders = tinyflags.GetOrDefaultDynamic[bool](group, id, "allow-duplicate-headers")
		target.HTTPExpectedStatusCodes = tinyflags.GetOrDefaultDynamic[[]string](group, id, "expected-status-codes")
		target.HTTPSkipTLSVerify = tinyflags.GetOrDefaultDynamic[bool](group, id, "skip-tls-verify")
		target.HTTPTimeout = tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout")

	case checker.TCP:
		target.TCPTimeout = tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout")

	case checker.ICMP:
		target.ICMPTimeout = tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout")
		target.ICMPReadTimeout = getDynamicDuration(group, id, "read-timeout")
		target.ICMPWriteTimeout = getDynamicDuration(group, id, "write-timeout")
	}
}

// getDynamicDuration returns a dynamic duration flag value or zero when unset.
func getDynamicDuration(group *tinyflags.DynamicGroup, id, name string) time.Duration {
	v, _ := tinyflags.GetDynamic[time.Duration](group, id, name)
	return v
}

// getDynamicInt returns a dynamic int flag value or zero when unset.
func getDynamicInt(group *tinyflags.DynamicGroup, id, name string) int {
	v, _ := tinyflags.GetDynamic[int](group, id, name)
	return v
}

// getDynamicBackoffMode returns the configured backoff mode or ModeNone when unset.
func getDynamicBackoffMode(group *tinyflags.DynamicGroup, id, name string) backoff.Mode {
	v := tinyflags.GetOrDefaultDynamic[backoff.Mode](group, id, name)
	if v == "" {
		return backoff.ModeLinear
	}
	return v
}
