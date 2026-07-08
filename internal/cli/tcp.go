package cli

import (
	"time"

	"github.com/containeroo/tinyflags"
)

// registerTCPFlags registers TCP-related flags and binds them to cfg.
func registerTCPFlags(tf *tinyflags.FlagSet, cfg *Config) {
	tcp := tf.DynamicGroup("tcp").Title("TCP")
	tcp.String("name", "", "Name of the TCP checker. Defaults to <ID>.")
	tcp.String("address", "", "TCP target address").
		Validate(validateTCPAddress).
		Required()
	tcp.Duration("timeout", 2*time.Second, "Timeout for TCP connection").
		Validate(validatePositiveDuration("timeout")).
		Placeholder("DURATION")
	tcp.Duration("interval", 0*time.Second, "Time between TCP requests. Defaults to --default-interval when unset or 0.").
		Validate(validateNonNegativeDuration("interval")).
		Placeholder("DURATION")
	tcp.Int("max-attempts", 0, "Maximum attempts before giving up. Defaults to --max-attempts when unset or 0.").
		Validate(validateOptionalMaxAttempts).
		Placeholder("N")
	registerRetryFlags(tcp)
}
