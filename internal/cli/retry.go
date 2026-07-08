package cli

import (
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/tinyflags"
)

// registerRetryFlags registers flags related to retry behavior.
func registerRetryFlags(group *tinyflags.DynamicGroup) {
	tinyflags.DynamicEnum(group, "backoff", backoff.ModeLinear, "Retry backoff mode.", backoff.ModeLinear, backoff.ModeExponential).
		Placeholder("MODE")
	group.Duration("max-interval", 0*time.Second, "Maximum retry interval when backoff increases the delay. Defaults to uncapped when unset or 0.").
		Validate(validateNonNegativeDuration("max-interval")).
		Placeholder("DURATION")
}
