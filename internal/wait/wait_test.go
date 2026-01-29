package wait

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/never/internal/checker"
)

// TestWaitUntilReady_ReadyHTTP ensures WaitUntilReady returns success when the HTTP target is ready.
func TestWaitUntilReady_ReadyHTTP(t *testing.T) {
	t.Parallel()

	server := &http.Server{Addr: ":9082"}
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.Handler = mux
	go func() { _ = server.ListenAndServe() }()
	defer server.Close() // nolint:errcheck

	checker, err := checker.NewChecker(checker.HTTP, "HTTPServer", "http://localhost:9082/ready")
	if err != nil {
		t.Fatalf("Failed to create HTTPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "HTTPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_HTTPFailsInitially tests HTTP target readiness after initial failures.
func TestWaitUntilReady_HTTPFailsInitially(t *testing.T) {
	t.Parallel()

	server := &http.Server{Addr: ":9083"}
	mux := http.NewServeMux()
	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond) // Simulate a delayed start
		w.WriteHeader(http.StatusOK)
	})
	server.Handler = mux
	go func() { _ = server.ListenAndServe() }()
	defer server.Close() // nolint:errcheck

	checker, err := checker.NewChecker(checker.HTTP, "HTTPServer", "http://localhost:9083/fail")
	if err != nil {
		t.Fatalf("Failed to create HTTPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "HTTPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_HTTPContextCanceled tests behavior when the context is canceled.
func TestWaitUntilReady_HTTPContextCanceled(t *testing.T) {
	t.Parallel()

	server := &http.Server{Addr: ":9084"}
	mux := http.NewServeMux()
	mux.HandleFunc("/canceled", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	server.Handler = mux
	go func() { _ = server.ListenAndServe() }()
	defer server.Close() // nolint:errcheck

	checker, err := checker.NewChecker(checker.HTTP, "HTTPServer", "http://localhost:9084/canceled")
	if err != nil {
		t.Fatalf("Failed to create HTTPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = WaitUntilReady(ctx, 50*time.Millisecond, -1, checker, logger)
	if err == nil {
		t.Fatalf("Expected context cancellation error, got nil")
	}

	expectedLog := "Waiting for HTTPServer to become ready..."
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_ReadyTCP ensures WaitUntilReady succeeds for a ready TCP target.
func TestWaitUntilReady_ReadyTCP(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "localhost:9085")
	if err != nil {
		t.Fatalf("Failed to create TCP server: %v", err)
	}
	defer ln.Close() // nolint:errcheck

	checker, err := checker.NewChecker(checker.TCP, "TCPServer", "localhost:9085")
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "TCPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_TCPFailsInitially tests TCP readiness after initial failures.
func TestWaitUntilReady_TCPFailsInitially(t *testing.T) {
	t.Parallel()

	var ln net.Listener
	go func() {
		time.Sleep(500 * time.Millisecond) // Simulate a delayed server start
		var err error
		ln, err = net.Listen("tcp", "localhost:9086")
		if err != nil {
			panic("Failed to start TCP server")
		}
	}()
	defer func() {
		if ln != nil {
			ln.Close() // nolint:errcheck
		}
	}()

	checker, err := checker.NewChecker(checker.TCP, "TCPServer", "localhost:9086")
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 100*time.Millisecond, -1, checker, logger)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "TCPServer is ready ✓"
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

// TestWaitUntilReady_TCPContextCanceled tests behavior when the TCP target's context is canceled.
func TestWaitUntilReady_TCPContextCanceled(t *testing.T) {
	t.Parallel()

	checker, err := checker.NewChecker(checker.TCP, "TCPServer", "localhost:9087")
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = WaitUntilReady(ctx, 50*time.Millisecond, -1, checker, logger)
	if err == nil {
		t.Fatalf("Expected context cancellation error, got nil")
	}

	expectedLog := "Waiting for TCPServer to become ready..."
	if !strings.Contains(output.String(), expectedLog) {
		t.Errorf("Expected log to contain %q, got %q", expectedLog, output.String())
	}
}

func TestNewStoppedTimer(t *testing.T) {
	t.Parallel()

	timer := newStoppedTimer(10 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
		t.Fatal("expected timer to be stopped")
	default:
	}

	timer.Reset(5 * time.Millisecond)
	select {
	case <-timer.C:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected timer to fire after reset")
	}
}

func TestWaitUntilReady_MaxAttempts(t *testing.T) {
	t.Parallel()

	checker, err := checker.NewChecker(checker.TCP, "TCPServer", "localhost:59999")
	if err != nil {
		t.Fatalf("Failed to create TCPChecker: %v", err)
	}

	var output strings.Builder
	logger := slog.New(slog.NewTextHandler(&output, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = WaitUntilReady(ctx, 10*time.Millisecond, 2, checker, logger)
	if err == nil {
		t.Fatal("Expected max attempts error, got nil")
	}
	if !errors.Is(err, ErrMaxAttemptsExceeded) {
		t.Fatalf("Expected ErrMaxAttemptsExceeded, got %v", err)
	}
}
