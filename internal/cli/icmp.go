package cli

import (
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/tinyflags"
)

// registerICMPFlags registers ICMP-related flags and binds them to cfg.
func registerICMPFlags(tf *tinyflags.FlagSet) {
	icmp := tf.DynamicGroup("icmp").Title("ICMP")
	icmp.String("name", "", "Name of the ICMP checker. Defaults to <ID>.")
	icmp.String("address", "", "ICMP target address").
		Validate(validateICMPAddress).
		Required()
	icmp.Duration("interval", 0*time.Second, "Time between ICMP requests. Defaults to --default-interval when unset or 0.").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[time.Duration](),
		)).
		Placeholder("DURATION")
	icmp.Int("max-attempts", 0, "Maximum attempts before giving up. Inherits --max-attempts when unset; 0 means endless retries.").
		Validate(tinyflags.Optional(
			tinyflags.NonNegative[int](),
		)).
		Placeholder("N")
	registerRetryFlags(icmp)
	icmp.Duration("timeout", checker.DefaultICMPConfig().ReadTimeout, "Timeout for ICMP read and write").
		Validate(tinyflags.NonNegative[time.Duration]()).
		Placeholder("DURATION")
	icmp.Duration("read-timeout", 0*time.Second, "Advanced: override the ICMP read timeout. Defaults to --icmp.<ID>.timeout when unset or 0.").
		Validate(tinyflags.NonNegative[time.Duration]()).
		Placeholder("DURATION")
	icmp.Duration("write-timeout", 0*time.Second, "Advanced: override the ICMP write timeout. Defaults to --icmp.<ID>.timeout when unset or 0.").
		Validate(tinyflags.NonNegative[time.Duration]()).
		Placeholder("DURATION")
}
