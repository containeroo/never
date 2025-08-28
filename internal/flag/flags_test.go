package flag

import (
	"errors"
	"testing"
	"time"

	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("Successful Parsing", func(t *testing.T) {
		t.Parallel()

		args := []string{"--default-interval=5s"}

		parsedFlags, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, 5*time.Second, parsedFlags.DefaultCheckInterval)
	})

	t.Run("Handle Help Flag", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--help"}, "1.0.0")
		require.Error(t, err)
	})

	t.Run("Show Version Flag", func(t *testing.T) {
		t.Parallel()

		args := []string{"--version"}

		_, err := ParseFlags(args, "1.0.0")
		require.Error(t, err)
		var verr *tinyflags.VersionRequested
		assert.True(t, errors.As(err, &verr), "expected VersionRequested error")
		assert.EqualError(t, err, "1.0.0")
	})

	t.Run("Invalid Duration Flag", func(t *testing.T) {
		t.Parallel()

		args := []string{"--default-interval=invalid"}

		_, err := ParseFlags(args, "1.0.0")
		require.Error(t, err)

		assert.EqualError(t, err, "invalid value for flag --default-interval: time: invalid duration \"invalid\".")
	})
}
