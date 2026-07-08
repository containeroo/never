package backoff

import (
	"testing"
	"time"
)

// TestNextIntervalNoneKeepsBaseInterval verifies the expected behavior.
func TestNextIntervalNoneKeepsBaseInterval(t *testing.T) {
	t.Parallel()

	got := NextInterval(ModeLinear, 100*time.Millisecond, 5, 0)
	want := 100 * time.Millisecond
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

// TestNextIntervalExponentialDoublesByAttempt verifies the expected behavior.
func TestNextIntervalExponentialDoublesByAttempt(t *testing.T) {
	t.Parallel()

	t.Run("first failure uses base interval", func(t *testing.T) {
		t.Parallel()

		got := NextInterval(ModeExponential, 100*time.Millisecond, 1, 0)
		want := 100 * time.Millisecond
		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("second failure doubles interval", func(t *testing.T) {
		t.Parallel()

		got := NextInterval(ModeExponential, 100*time.Millisecond, 2, 0)
		want := 200 * time.Millisecond
		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("third failure doubles again", func(t *testing.T) {
		t.Parallel()

		got := NextInterval(ModeExponential, 100*time.Millisecond, 3, 0)
		want := 400 * time.Millisecond
		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})
}

// TestNextIntervalExponentialCapsAtMaxInterval verifies the expected behavior.
func TestNextIntervalExponentialCapsAtMaxInterval(t *testing.T) {
	t.Parallel()

	got := NextInterval(ModeExponential, 100*time.Millisecond, 10, 500*time.Millisecond)
	want := 500 * time.Millisecond
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
