package factory_test

import (
	"testing"
	"time"

	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCheckers(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP Checker", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		http := tf.DynamicGroup("http")
		http.String("name", "", "Name of the HTTP checker")
		http.String("method", "GET", "HTTP method to use")
		http.String("address", "", "HTTP target URL")
		http.Duration("interval", 1*time.Second, "Time between HTTP requests. Can be overwritten with --default-interval.")
		http.StringSlice("header", nil, "HTTP headers to send")
		http.Bool("allow-duplicate-headers", true, "Allow duplicate HTTP headers")
		http.StringSlice("expected-status-codes", []string{"200"}, "Expected HTTP status codes")
		http.Bool("skip-tls-verify", true, "Skip TLS verification")
		http.Duration("timeout", 22*time.Second, "Timeout in seconds")

		args := []string{
			"--http.mygroup.address=http://example.com",
			"--http.mygroup.method=GET",
			"--http.mygroup.interval=5s",
			"--http.mygroup.header=Content-Type=application/json",
			"--http.mygroup.skip-tls-verify=true",
			"--http.mygroup.timeout=33s",
		}
		err := tf.Parse(args)
		require.NoError(t, err)

		checkers, err := factory.BuildCheckers(tf.DynamicGroups(), 9*time.Second)
		require.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "http://example.com", checkers[0].Checker.Address())
		assert.Equal(t, 5*time.Second, checkers[0].Interval)
	})

	t.Run("Invalid Check Type", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		invalidGroup := tf.DynamicGroup("invalid")
		invalidGroup.String("address", "invalid-address", "Invalid target address")

		args := []string{"--invalid.mygroup.address=invalid-address"}
		err := tf.Parse(args)
		require.NoError(t, err)

		checkers, err := factory.BuildCheckers(tf.DynamicGroups(), 2*time.Second)
		assert.Nil(t, checkers)
		assert.EqualError(t, err, "unsupported check type: invalid")
	})

	t.Run("Invalid Header Parsing", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		httpGroup := tf.DynamicGroup("http")
		httpGroup.String("address", "http://example.com", "HTTP target address")
		httpGroup.StringSlice("header", []string{}, "HTTP headers")
		httpGroup.String("method", "GET", "HTTP method to use")
		httpGroup.Bool("allow-duplicate-headers", true, "Allow duplicate HTTP headers")

		args := []string{
			"--http.mygroup.address=http://example.com",
			"--http.mygroup.header=InvalidHeaderFormat",
		}
		err := tf.Parse(args)
		require.NoError(t, err)

		checkers, err := factory.BuildCheckers(tf.DynamicGroups(), 2*time.Second)

		require.Error(t, err)
		assert.EqualError(t, err, "invalid \"--http.mygroup.header\": invalid header format: \"InvalidHeaderFormat\"")
		assert.Nil(t, checkers)
		assert.ErrorContains(t, err, "invalid \"--http.mygroup.header\"")
	})

	t.Run("Inalid HTTP Status codes", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		http := tf.DynamicGroup("http")
		http.String("name", "", "Name of the HTTP checker")
		http.String("method", "GET", "HTTP method to use")
		http.String("address", "", "HTTP target URL")
		http.Duration("interval", 1*time.Second, "Time between HTTP requests. Can be overwritten with --default-interval.")
		http.StringSlice("header", nil, "HTTP headers to send")
		http.Bool("allow-duplicate-headers", true, "Allow duplicate HTTP headers")
		http.String("expected-status-codes", "200", "Expected HTTP status codes")
		http.Bool("skip-tls-verify", true, "Skip TLS verification")
		http.Duration("timeout", 2*time.Second, "Timeout in seconds")

		args := []string{
			"--http.myid.address=http://example.com",
			"--http.myid.expected-status-codes=201-200",
		}
		err := tf.Parse(args)
		require.NoError(t, err)

		checkers, err := factory.BuildCheckers(tf.DynamicGroups(), 2*time.Second)
		require.Error(t, err)
		assert.Len(t, checkers, 0)
	})

	t.Run("Valid HTTP Status codes", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		httpGroup := tf.DynamicGroup("http")
		httpGroup.String("name", "", "Name of the HTTP checker. Defaults to <ID>.")
		httpGroup.String("address", "http://example.com", "HTTP target address")
		httpGroup.String("expected-status-codes", "200,201", "HTTP expected status codes")
		httpGroup.StringSlice("header", []string{}, "HTTP headers")
		httpGroup.String("method", "GET", "HTTP method to use")
		httpGroup.Bool("allow-duplicate-headers", true, "Allow duplicate HTTP headers")
		httpGroup.Bool("skip-tls-verify", true, "Skip TLS verification")
		httpGroup.Duration("timeout", 2*time.Second, "Timeout in seconds")

		args := []string{
			"--http.mygroup.address=http://example.com",
			"--http.mygroup.expected-status-codes=200,201",
		}
		err := tf.Parse(args)
		require.NoError(t, err)

		checkers, err := factory.BuildCheckers(tf.DynamicGroups(), 2*time.Second)
		require.NoError(t, err)
		assert.Len(t, checkers, 1)
	})

	t.Run("Valid TCP Checker", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		tcpGroup := tf.DynamicGroup("tcp")
		tcpGroup.String("name", "", "Name of the HTTP checker. Defaults to <ID>.")
		tcpGroup.String("address", "127.0.0.1:8080", "TCP target address")
		tcpGroup.Duration("timeout", 3*time.Second, "Timeout")

		args := []string{
			"--tcp.mygroup.address=127.0.0.1:8080",
			"--tcp.mygroup.timeout=3s",
		}
		err := tf.Parse(args)
		require.NoError(t, err)

		checkers, err := factory.BuildCheckers(tf.DynamicGroups(), 2*time.Second)
		require.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "127.0.0.1:8080", checkers[0].Checker.Address())
	})

	t.Run("Valid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		icmpGroup := tf.DynamicGroup("icmp")
		icmpGroup.String("name", "", "Name of the ICMP checker. Defaults to <ID>.")
		icmpGroup.String("address", "8.8.8.8", "ICMP target address")
		icmpGroup.Duration("read-timeout", 2*time.Second, "Read timeout")
		icmpGroup.Duration("write-timeout", 2*time.Second, "Write timeout")

		args := []string{
			"--icmp.mygroup.address=8.8.8.8",
			"--icmp.mygroup.read-timeout=2s",
			"--icmp.mygroup.write-timeout=2s",
		}
		err := tf.Parse(args)
		require.NoError(t, err)

		checkers, err := factory.BuildCheckers(tf.DynamicGroups(), 2*time.Second)
		require.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "8.8.8.8", checkers[0].Checker.Address())
	})

	t.Run("Invalid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		tf := tinyflags.NewFlagSet("test.exe", tinyflags.ContinueOnError)
		icmpGroup := tf.DynamicGroup("icmp")
		icmpGroup.String("name", "", "Name of the TCP checker. Defaults to <ID>.")
		icmpGroup.String("address", "8.8.8.8", "ICMP target address")
		icmpGroup.Duration("read-timeout", 2*time.Second, "Read timeout")
		icmpGroup.Duration("write-timeout", 2*time.Second, "Write timeout")

		args := []string{
			"--icmp.mygroup.address=://invalid-url",
		}

		err := tf.Parse(args)
		require.NoError(t, err)

		checker, err := factory.BuildCheckers(tf.DynamicGroups(), 2*time.Second)
		assert.Nil(t, checker)
		require.Error(t, err)
	})
}
