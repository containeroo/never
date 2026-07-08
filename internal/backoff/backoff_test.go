package backoff

import (
	"testing"
	"time"
)

func TestNextIntervalNoneKeepsBaseInterval(t *testing.T) {
	t.Parallel()

	got := NextInterval(ModeNone, 100*time.Millisecond, 5, 0)
	want := 100 * time.Millisecond
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestNextIntervalExponentialDoublesByAttempt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{name: "first failure uses base interval", attempt: 1, want: 100 * time.Millisecond},
		{name: "second failure doubles interval", attempt: 2, want: 200 * time.Millisecond},
		{name: "third failure doubles again", attempt: 3, want: 400 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NextInterval(ModeExponential, 100*time.Millisecond, tt.attempt, 0)
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestNextIntervalExponentialCapsAtMaxInterval(t *testing.T) {
	t.Parallel()

	got := NextInterval(ModeExponential, 100*time.Millisecond, 10, 500*time.Millisecond)
	want := 500 * time.Millisecond
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
