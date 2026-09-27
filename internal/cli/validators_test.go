package cli

import (
	"testing"

	"github.com/containeroo/never/internal/testutils"

	"github.com/stretchr/testify/assert"
)

// TestValidateHTTPAddress verifies HTTP address validation accepts supported inputs.
func TestValidateHTTPAddress(t *testing.T) {
	t.Parallel()

	t.Run("http URL", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validateHTTPAddress("http://example.com"))
	})

	t.Run("https URL", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validateHTTPAddress("https://example.com/ready"))
	})

	t.Run("resolver reference", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validateHTTPAddress("env:TARGET_URL"))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, validateHTTPAddress(""), "invalid HTTP URL")
	})

	t.Run("missing host", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, validateHTTPAddress("http://"), "invalid HTTP URL")
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, validateHTTPAddress("ftp://example.com"), `unsupported scheme: "ftp"`)
	})
}

// TestValidateICMPAddress verifies ICMP address validation accepts supported inputs.
func TestValidateICMPAddress(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		for _, address := range []string{testutils.LocalhostIPv4, "2001:db8::1", "example.com", "localhost", "env:TARGET_HOST"} {
			assert.NoError(t, validateICMPAddress(address))
		}
	})

	t.Run("scheme", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, validateICMPAddress("icmp://example.com"), "ICMP check cannot have a scheme")
	})

	t.Run("path", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, validateICMPAddress("example.com/ready"), "ICMP address must be a hostname or IP without path or port")
	})

	t.Run("port", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, validateICMPAddress("example.com:80"), "ICMP address must be a hostname or IP without path or port")
	})

	t.Run("invalid hostname", func(t *testing.T) {
		t.Parallel()
		assert.EqualError(t, validateICMPAddress("exa_mple.com"), `invalid hostname: "exa_mple.com"`)
	})
}

// TestValidateTCPAddress verifies TCP address validation accepts supported inputs.
func TestValidateTCPAddress(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		for _, address := range []string{testutils.LocalhostAddr("80"), "example.com:443", "[2001:db8::1]:443", "env:TARGET_ADDRESS"} {
			assert.NoError(t, validateTCPAddress(address))
		}
	})

	t.Run("missing port", func(t *testing.T) {
		t.Parallel()
		assert.ErrorContains(t, validateTCPAddress("example.com"), "TCP address must be host:port")
	})

	t.Run("scheme", func(t *testing.T) {
		t.Parallel()
		assert.ErrorContains(t, validateTCPAddress("tcp://example.com:80"), "TCP address must be host:port")
	})
}

// TestIsResolvableValue verifies supported resolver prefixes are recognized.
func TestIsResolvableValue(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"env:TARGET",
		"file:/config/app.txt//Target",
		"json:/config/app.json//target",
		"yaml:/config/app.yaml//target",
		"ini:/config/app.ini//Target.Address",
	} {
		assert.True(t, isResolvableValue(value), value)
	}

	assert.False(t, isResolvableValue("http://example.com"))
}

// TestIsHostnameLike verifies hostname validation edge cases.
func TestIsHostnameLike(t *testing.T) {
	t.Parallel()

	for _, hostname := range []string{"example.com", "sub.domain.local", "localhost", "a-b.c"} {
		assert.True(t, isHostnameLike(hostname), hostname)
	}

	for _, hostname := range []string{"", "-bad.example", "bad-.example", "bad..example", "example.com/", "example.com:80", "exa_mple.com"} {
		assert.False(t, isHostnameLike(hostname), hostname)
	}
}

// TestIsAlphaNum verifies ASCII hostname characters.
func TestIsAlphaNum(t *testing.T) {
	t.Parallel()

	for _, ch := range []byte{'a', 'A', '0', '9'} {
		assert.True(t, isAlphaNum(ch), string(ch))
	}
	for _, ch := range []byte{'-', '_', '!', '$', '%', '*', '/', ':', '@', '[', '\\', ']', '^'} {
		assert.False(t, isAlphaNum(ch), string(ch))
	}
}
