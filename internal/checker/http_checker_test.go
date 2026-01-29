package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP check with default configuration", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		require.NoError(t, err)
		assert.Equal(t, checker.Address(), server.URL)
	})

	t.Run("HTTP check with custom headers", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPHeaders(http.Header{
			"Authorization": []string{"Bearer token"},
		}))
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		require.NoError(t, err)
	})

	t.Run("HTTP check with unexpected status code", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		require.Error(t, err)
		assert.EqualError(t, err, "unexpected status code: got 404, expected one of [200]")
	})

	t.Run("Invalid URL for HTTP check", func(t *testing.T) {
		t.Parallel()

		checker, err := newHTTPChecker("example", "://invalid-url")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = checker.Check(context.Background()) // Run the check to trigger the error.
		require.Error(t, err)
		assert.EqualError(t, err, "failed to create request: parse \"://invalid-url\": missing protocol scheme")
	})

	t.Run("Timeout during HTTP check", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second) // Simulate delay
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPTimeout(1*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)

		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("HTTP request failed: Get \"http://%s\": context deadline exceeded", server.Listener.Addr().String()))
	})

	t.Run("Custom expected status codes", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithExpectedStatusCodes([]int{202}))
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		require.NoError(t, err)
	})

	t.Run("Custom HTTP method", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPMethod(http.MethodPost))
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		require.NoError(t, err)
	})

	t.Run("Skip TLS verification", func(t *testing.T) {
		t.Parallel()

		// Create a test server with a self-signed certificate
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		checker, err := newHTTPChecker("example", server.URL, WithHTTPSkipTLSVerify(true))
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = checker.Check(ctx)
		require.NoError(t, err)
	})
}
