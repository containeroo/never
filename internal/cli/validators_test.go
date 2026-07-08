package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateHTTPAddress verifies HTTP address validation accepts supported inputs.
func TestValidateHTTPAddress(t *testing.T) {
	t.Parallel()

	t.Run("http URL", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress("http://example.com"))
	})

	t.Run("https URL", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress("https://example.com/ready"))
	})

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress(" https://example.com "))
	})

	t.Run("resolvable value", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress("env:NEVER_HTTP_ADDRESS"))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateHTTPAddress(""), `invalid URL: ""`)
	})

	t.Run("missing host", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateHTTPAddress("http://"), `invalid URL: "http://"`)
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateHTTPAddress("ftp://example.com"), `unsupported scheme: "ftp"`)
	})
}

// TestValidateICMPAddress verifies ICMP address validation accepts supported inputs.
func TestValidateICMPAddress(t *testing.T) {
	t.Parallel()

	t.Run("IPv4", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("127.0.0.1"))
	})

	t.Run("IPv6", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("2001:db8::1"))
	})

	t.Run("hostname", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("example.com"))
	})

	t.Run("localhost", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("localhost"))
	})

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress(" example.com "))
	})

	t.Run("resolvable value", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("env:NEVER_ICMP_ADDRESS"))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress(""), "ICMP address cannot be empty")
	})

	t.Run("scheme", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("icmp://example.com"), "ICMP check cannot have a scheme")
	})

	t.Run("path", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("example.com/ready"), "ICMP address must be a hostname or IP without path or port")
	})

	t.Run("port", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("example.com:80"), "ICMP address must be a hostname or IP without path or port")
	})

	t.Run("invalid hostname", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("exa_mple.com"), `invalid hostname: "exa_mple.com"`)
	})
}

// TestValidateTCPAddress verifies TCP address validation accepts supported inputs.
func TestValidateTCPAddress(t *testing.T) {
	t.Parallel()

	t.Run("IPv4 with port", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress("127.0.0.1:80"))
	})

	t.Run("hostname with port", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress("example.com:443"))
	})

	t.Run("IPv6 with port", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress("[2001:db8::1]:443"))
	})

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress(" example.com:443 "))
	})

	t.Run("resolvable value", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress("env:NEVER_TCP_ADDRESS"))
	})

	t.Run("missing port", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateTCPAddress("example.com"), "TCP address must be host:port")
	})

	t.Run("scheme", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateTCPAddress("tcp://example.com:80"), "TCP address must be host:port")
	})
}

// TestValidateHTTPStatusCodes verifies HTTP status code validation accepts supported expressions.
func TestValidateHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	t.Run("single code", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPStatusCodes("200"))
	})

	t.Run("comma list", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPStatusCodes("200,204,301"))
	})

	t.Run("range", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPStatusCodes("200-299"))
	})

	t.Run("descending range", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateHTTPStatusCodes("299-200"), "invalid status code")
	})

	t.Run("not numeric", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateHTTPStatusCodes("ok"), "invalid status code")
	})
}

// TestValidateMaxAttempts verifies global max-attempts validation.
func TestValidateMaxAttempts(t *testing.T) {
	t.Parallel()

	t.Run("endless", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateMaxAttempts(-1))
	})

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateMaxAttempts(1))
	})

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateMaxAttempts(0), "max-attempts must be -1 or positive")
	})

	t.Run("below endless", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateMaxAttempts(-2), "max-attempts must be -1 or positive")
	})
}

// TestValidateOptionalMaxAttempts verifies per-target max-attempts validation.
func TestValidateOptionalMaxAttempts(t *testing.T) {
	t.Parallel()

	t.Run("inherits global", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateOptionalMaxAttempts(0))
	})

	t.Run("endless", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateOptionalMaxAttempts(-1))
	})

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateOptionalMaxAttempts(3))
	})

	t.Run("below endless", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateOptionalMaxAttempts(-2), "max-attempts must be -1 or positive")
	})
}

// TestValidatePositiveDuration verifies positive duration validation.
func TestValidatePositiveDuration(t *testing.T) {
	t.Parallel()

	validateTimeout := validatePositiveDuration("timeout")

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTimeout(time.Nanosecond))
	})

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateTimeout(0), "timeout must be positive")
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateTimeout(-time.Second), "timeout must be positive")
	})
}

// TestValidateNonNegativeDuration verifies non-negative duration validation.
func TestValidateNonNegativeDuration(t *testing.T) {
	t.Parallel()

	validateInterval := validateNonNegativeDuration("interval")

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateInterval(0))
	})

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateInterval(time.Second))
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateInterval(-time.Second), "interval must be non-negative")
	})
}

// assertNoValidationError verifies a validator accepted the value.
func assertNoValidationError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// assertExactValidationError verifies a validator returned the exact expected error.
func assertExactValidationError(t *testing.T, err error, want string) {
	t.Helper()
	require.Error(t, err)
	assert.EqualError(t, err, want)
}

// assertValidationErrorContains verifies a validator returned an error containing the expected text.
func assertValidationErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	require.Error(t, err)
	assert.Contains(t, err.Error(), want)
}
