package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestWithSpinner_Success(t *testing.T) {
	pterm.DisableStyling()

	defer pterm.EnableStyling()

	result, err := withSpinner("Test operation", func() (string, error) {
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestWithSpinner_Error(t *testing.T) {
	pterm.DisableStyling()

	defer pterm.EnableStyling()

	expectedErr := errors.New("operation failed")
	result, err := withSpinner("Test operation", func() (string, error) {
		return "partial", expectedErr
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Test operation")
	assert.Contains(t, err.Error(), "operation failed")
	assert.Equal(t, "partial", result)
}

func TestWithSpinner_WithPointerType(t *testing.T) {
	pterm.DisableStyling()

	defer pterm.EnableStyling()

	type testResult struct {
		Value int
	}

	result, err := withSpinner("Test operation", func() (*testResult, error) {
		return &testResult{Value: 42}, nil
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 42, result.Value)
}

func TestWithSpinner_ErrorWrapping(t *testing.T) {
	pterm.DisableStyling()

	defer pterm.EnableStyling()

	originalErr := errors.New("underlying error")
	_, err := withSpinner("Doing something", func() (int, error) {
		return 0, originalErr
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, originalErr, "error should wrap the original error")
}

func TestDefaultConfigPath(t *testing.T) {
	path := defaultConfigPath()
	require.NotEmpty(t, path)
	require.Contains(t, path, ".tmhi-cli.toml")
}

func TestDefaultConfigPath_HomeError(t *testing.T) {
	t.Setenv("HOME", "")

	path := defaultConfigPath()
	require.Equal(t, ".tmhi-cli.toml", path)
}

func TestBuildFlags(t *testing.T) {
	var configFile string

	flags := cmdFlags(&configFile, nil)

	require.Len(t, flags, 11)
}

func TestBuildCommands(t *testing.T) {
	commands := cmdCommands()

	require.Len(t, commands, 6)
	require.Equal(t, "login", commands[0].Name)
	require.Equal(t, "reboot", commands[1].Name)
	require.Equal(t, "info", commands[2].Name)
	require.Equal(t, "status", commands[3].Name)
	require.Equal(t, "signal", commands[4].Name)
	require.Equal(t, "req", commands[5].Name)
}

func TestCmd_Help(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"tmhi-cli", "--help"}

	defer func() { os.Args = oldArgs }()

	out := testutil.CaptureOutput(t, func() {
		Cmd("test-version")
	})

	assert.Contains(t, out, "Utility to interact with T-Mobile Home Internet gateway")
}

func TestCmd_Version(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"tmhi-cli", "--version"}

	defer func() { os.Args = oldArgs }()

	testVersion := "test-version-123"
	out := testutil.CaptureOutput(t, func() {
		Cmd(testVersion)
	})

	assert.Contains(
		t,
		out,
		testVersion,
		"expected version output to contain %q, got: %q",
		testVersion,
		out,
	)
}

func TestSetupColor_Never(t *testing.T) {
	// Reset pterm state before test
	pterm.EnableStyling()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  ConfigColor,
				Value: "never",
			},
		},
	}

	// Simulate parsing the flag
	_ = cmd.Set("color", "never")

	_, err := setupColor(context.Background(), cmd)
	require.NoError(t, err)
	assert.True(t, pterm.RawOutput, "pterm.RawOutput should be true after --color=never")
}

func TestSetupColor_Always(t *testing.T) {
	// First disable, then test that "always" keeps colors enabled
	pterm.DisableStyling()
	pterm.EnableStyling()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  ConfigColor,
				Value: "always",
			},
		},
	}

	_ = cmd.Set("color", "always")

	_, err := setupColor(context.Background(), cmd)
	require.NoError(t, err)
	assert.False(t, pterm.RawOutput, "pterm.RawOutput should be false after --color=always")
}

func TestSetupColor_AutoDefault(t *testing.T) {
	// Test that --color=auto (the default) correctly detects terminal state
	// When stdout is not a terminal (like in tests), styling should be disabled
	pterm.EnableStyling()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  ConfigColor,
				Value: "auto",
			},
		},
	}

	// Don't set the flag - test default behavior
	_, err := setupColor(context.Background(), cmd)
	require.NoError(t, err)

	// In test environment, stdout is typically not a terminal
	// so styling should be disabled
	assert.True(
		t,
		pterm.RawOutput,
		"pterm.RawOutput should be true in non-terminal environment with --color=auto",
	)
}

func TestOnUsageError(t *testing.T) {
	app := &cli.Command{
		Name: "tmhi-cli",
		OnUsageError: func(_ context.Context, cmd *cli.Command, err error, _ bool) error {
			_, _ = fmt.Fprintf(cmd.ErrWriter, "error: %v\n", err)

			return err
		},
	}

	err := app.Run(context.Background(), []string{"tmhi-cli", "--invalid-flag"})
	require.Error(t, err)
}

func TestDebugFlagAction(t *testing.T) {
	pterm.DisableDebugMessages()
	defer pterm.DisableDebugMessages()

	var configFile string

	flags := cmdFlags(&configFile, nil)

	var debugFlag *cli.BoolFlag

	for _, f := range flags {
		if bf, ok := f.(*cli.BoolFlag); ok && bf.Name == ConfigDebug {
			debugFlag = bf

			break
		}
	}

	require.NotNil(t, debugFlag)
	require.NotNil(t, debugFlag.Action)

	cmd := &cli.Command{Flags: flags}
	err := debugFlag.Action(context.Background(), cmd, true)
	require.NoError(t, err)
}

func TestQuietFlagAction(t *testing.T) {
	pterm.EnableOutput()
	defer pterm.EnableOutput()

	var configFile string

	flags := cmdFlags(&configFile, nil)

	var quietFlag *cli.BoolFlag

	for _, f := range flags {
		if bf, ok := f.(*cli.BoolFlag); ok && bf.Name == ConfigQuiet {
			quietFlag = bf

			break
		}
	}

	require.NotNil(t, quietFlag)
	require.NotNil(t, quietFlag.Action)

	cmd := &cli.Command{Flags: flags}
	err := quietFlag.Action(context.Background(), cmd, true)
	require.NoError(t, err)
}
