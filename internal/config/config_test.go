package config

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("Successful Parsing", func(t *testing.T) {
		t.Parallel()

		args := []string{"--default-interval=5s"}
		var output strings.Builder

		parsedFlags, err := ParseFlags(args, "1.0.0", &output)
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, parsedFlags.DefaultCheckInterval)
	})

	t.Run("Handle Help Flag", func(t *testing.T) {
		t.Parallel()

		var output strings.Builder
		flagSet := setupGlobalFlags()
		flagSet.SetOutput(&output) // Ensure output is properly set
		_ = flagSet.Parse([]string{"--help"})

		flagSet.Usage = func() {
			fmt.Fprintln(&output, "Usage: portpatrol [FLAGS] [DYNAMIC FLAGS..]")
		}

		err := handleSpecialFlags(flagSet, "1.0.0")
		assert.Error(t, err)
		assert.IsType(t, &HelpRequested{}, err)
		assert.Contains(t, output.String(), "Usage: portpatrol [FLAGS] [DYNAMIC FLAGS..]")
	})

	t.Run("Show Version Flag", func(t *testing.T) {
		t.Parallel()

		args := []string{"--version"}
		var output bytes.Buffer

		_, err := ParseFlags(args, "1.0.0", &output)
		assert.Error(t, err)
		assert.IsType(t, &HelpRequested{}, err)
		assert.Contains(t, err.Error(), "PortPatrol version 1.0.0")
	})

	t.Run("Invalid Duration Flag", func(t *testing.T) {
		t.Parallel()

		args := []string{"--default-interval=invalid"}
		var output bytes.Buffer

		_, err := ParseFlags(args, "1.0.0", &output)
		assert.Error(t, err)

		assert.EqualError(t, err, "flag parsing error: invalid argument \"invalid\" for \"--default-interval\" flag: time: invalid duration \"invalid\"")
	})
}

func TestIsHelpRequested(t *testing.T) {
	t.Parallel()

	t.Run("Help requested", func(t *testing.T) {
		t.Parallel()

		err := &HelpRequested{Message: "Help requested"}
		assert.ErrorIs(t, err, &HelpRequested{})
	})
}

func TestSetupGlobalFlags(t *testing.T) {
	t.Parallel()

	flagSet := setupGlobalFlags()
	assert.NotNil(t, flagSet.Lookup("default-interval"))
	assert.NotNil(t, flagSet.Lookup("version"))
	assert.NotNil(t, flagSet.Lookup("help"))
}

func TestSetupDynamicFlags(t *testing.T) {
	t.Parallel()

	dynFlags := setupDynamicFlags()
	assert.NotNil(t, dynFlags.Group("http"))
	assert.NotNil(t, dynFlags.Group("tcp"))
	assert.NotNil(t, dynFlags.Group("icmp"))

	httpGroup := dynFlags.Group("http")
	assert.NotNil(t, httpGroup.Lookup("name"))
	assert.NotNil(t, httpGroup.Lookup("method"))
	assert.NotNil(t, httpGroup.Lookup("address"))
}

func TestSetupUsage(t *testing.T) {
	t.Parallel()

	var output strings.Builder
	flagSet := setupGlobalFlags()
	flagSet.SetOutput(&output)

	dynFlags := setupDynamicFlags()
	dynFlags.SetOutput(&output)

	setupUsage(flagSet, dynFlags)
	flagSet.Usage()

	usageOutput := output.String()
	assert.Contains(t, usageOutput, "Usage: portpatrol [FLAGS] [DYNAMIC FLAGS..]")
	assert.Contains(t, usageOutput, "Global Flags:")
	assert.Contains(t, usageOutput, "--default-interval")
	assert.Contains(t, usageOutput, "Dynamic Flags:")
	assert.Contains(t, usageOutput, "http")
}

func TestHandleSpecialFlags(t *testing.T) {
	t.Parallel()

	t.Run("Handle Help Flag", func(t *testing.T) {
		t.Parallel()

		var output strings.Builder
		flagSet := setupGlobalFlags()
		flagSet.SetOutput(&output)

		flagSet.Usage = func() {
			fmt.Fprintln(&output, "Usage: portpatrol [FLAGS] [DYNAMIC FLAGS..]")
		}

		args := []string{"--help"}
		err := flagSet.Parse(args)
		assert.NoError(t, err)

		err = handleSpecialFlags(flagSet, "1.0.0")
		assert.Error(t, err)
	})

	t.Run("Handle Version Flag", func(t *testing.T) {
		t.Parallel()

		flagSet := setupGlobalFlags()
		_ = flagSet.Parse([]string{"--version"})

		err := handleSpecialFlags(flagSet, "1.0.0")
		assert.Error(t, err)
		assert.IsType(t, &HelpRequested{}, err)
		assert.Contains(t, err.Error(), "PortPatrol version 1.0.0")
	})

	t.Run("No Special Flags", func(t *testing.T) {
		t.Parallel()

		flagSet := setupGlobalFlags()
		_ = flagSet.Parse([]string{})

		err := handleSpecialFlags(flagSet, "1.0.0")
		assert.NoError(t, err)
	})
}

func TestGetDurationFlag(t *testing.T) {
	t.Parallel()

	t.Run("Valid Duration Flag", func(t *testing.T) {
		t.Parallel()

		flagSet := setupGlobalFlags()
		_ = flagSet.Set("default-interval", "10s")

		duration, err := getDurationFlag(flagSet, "default-interval", time.Second)
		assert.NoError(t, err)
		assert.Equal(t, 10*time.Second, duration)
	})

	t.Run("Invalid Duration", func(t *testing.T) {
		t.Parallel()

		flagSet := pflag.NewFlagSet("portpatrol", pflag.ContinueOnError)
		flagSet.String("invalid-flag", "invalid", "Invalid flag")
		err := flagSet.Set("invalid-flag", "invalid")
		assert.NoError(t, err)

		_, err = getDurationFlag(flagSet, "invalid-flag", time.Second)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid duration for flag 'invalid'")
	})

	t.Run("Missing Duration Flag", func(t *testing.T) {
		t.Parallel()

		flagSet := setupGlobalFlags()

		duration, err := getDurationFlag(flagSet, "non-existent-flag", time.Second)
		assert.NoError(t, err)
		assert.Equal(t, time.Second, duration)
	})
}
