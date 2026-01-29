package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"slices"
	"time"
)

const (
	defaultHTTPTimeout       time.Duration = 1 * time.Second
	defaultHTTPMethod        string        = http.MethodGet
	defaultHTTPSkipTLSVerify bool          = false
)

var defaultHTTPExpectedStatusCodes = []int{200}

// HTTPChecker implements the Checker interface for HTTP checks.
type HTTPChecker struct {
	name                string
	address             string
	method              string
	headers             http.Header
	expectedStatusCodes []int
	skipTLSVerify       bool
	timeout             time.Duration
	client              *http.Client
}

func (c *HTTPChecker) Address() string { return c.address }
func (c *HTTPChecker) Name() string    { return c.name }
func (c *HTTPChecker) Type() string    { return HTTP.String() }
func (c *HTTPChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, c.method, c.address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	if slices.Contains(c.expectedStatusCodes, resp.StatusCode) {
		return nil
	}

	return fmt.Errorf("unexpected status code: got %d, expected one of %v", resp.StatusCode, c.expectedStatusCodes)
}

// newHTTPChecker creates a new HTTPChecker with functional options.
func newHTTPChecker(name, address string, opts ...Option) (*HTTPChecker, error) { // nolint:unparam
	checker := &HTTPChecker{
		name:                name,
		address:             address,
		method:              defaultHTTPMethod,
		headers:             make(http.Header),
		expectedStatusCodes: defaultHTTPExpectedStatusCodes,
		skipTLSVerify:       defaultHTTPSkipTLSVerify,
		timeout:             defaultHTTPTimeout,
	}

	for _, opt := range opts {
		opt.apply(checker)
	}

	checker.client = &http.Client{
		Timeout: checker.timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: checker.skipTLSVerify,
			},
		},
	}

	return checker, nil
}

// WithHTTPMethod sets the HTTP method for the HTTPChecker.
func WithHTTPMethod(method string) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			httpChecker.method = method
		}
	})
}

// WithHTTPHeaders sets the HTTP headers for the HTTPChecker.
func WithHTTPHeaders(headers http.Header) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			httpChecker.headers = headers
		}
	})
}

// WithExpectedStatusCodes sets the expected status codes for the HTTPChecker.
func WithExpectedStatusCodes(codes []int) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			httpChecker.expectedStatusCodes = codes
		}
	})
}

// WithHTTPSkipTLSVerify sets the TLS verification flag for the HTTPChecker.
func WithHTTPSkipTLSVerify(skip bool) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			httpChecker.skipTLSVerify = skip
		}
	})
}

// WithHTTPTimeout sets the timeout for the HTTPChecker.
func WithHTTPTimeout(timeout time.Duration) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			httpChecker.timeout = timeout
		}
	})
}
