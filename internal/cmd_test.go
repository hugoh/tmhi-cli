package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	tmhi "github.com/hugoh/tmhi-gateway/v2"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
	capturer "github.com/zenizh/go-capturer"
)

func TestCaptureOutput(t *testing.T) {
	output := capturer.CaptureOutput(func() {
		fmt.Print("hello world") //nolint:forbidigo
	})

	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain 'hello world', got %q", output)
	}
}

func TestCaptureOutput_MultipleWrites(t *testing.T) {
	output := capturer.CaptureOutput(func() {
		fmt.Print("line1") //nolint:forbidigo
		fmt.Print("line2") //nolint:forbidigo
	})

	if !strings.Contains(output, "line1") || !strings.Contains(output, "line2") {
		t.Errorf("expected output to contain 'line1' and 'line2', got %q", output)
	}
}

func TestFetchWithFeedback_Success(t *testing.T) {
	result, err := fetchWithFeedback(
		t.Context(),
		newTestApp(nil).newSpinner,
		"Test operation",
		func(context.Context) (string, error) {
			return "success", nil
		},
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestFetchWithFeedback_Error(t *testing.T) {
	expectedErr := errors.New("operation failed")
	result, err := fetchWithFeedback(
		t.Context(),
		newTestApp(nil).newSpinner,
		"Test operation",
		func(context.Context) (string, error) {
			return "partial", expectedErr
		},
		nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Test operation")
	assert.Contains(t, err.Error(), "operation failed")
	assert.Equal(t, "partial", result)
}

func TestFetchWithFeedback_WithPointerType(t *testing.T) {
	type testResult struct {
		Value int
	}

	result, err := fetchWithFeedback(
		t.Context(),
		newTestApp(nil).newSpinner,
		"Test operation",
		func(context.Context) (*testResult, error) {
			return &testResult{Value: 42}, nil
		},
		nil,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 42, result.Value)
}

func TestFetchWithFeedback_ErrorWrapping(t *testing.T) {
	originalErr := errors.New("underlying error")
	_, err := fetchWithFeedback(
		t.Context(),
		newTestApp(nil).newSpinner,
		"Doing something",
		func(context.Context) (int, error) {
			return 0, originalErr
		},
		nil,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, originalErr, "error should wrap the original error")
}

func TestFetchWithFeedback_WithDisplay(t *testing.T) {
	displayCalled := false

	var displayedResult string

	result, err := fetchWithFeedback(
		t.Context(),
		newTestApp(nil).newSpinner,
		"Test operation",
		func(context.Context) (string, error) {
			return "test value", nil
		},
		func(r string) {
			displayCalled = true
			displayedResult = r
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "test value", result)
	assert.True(t, displayCalled, "display function should be called")
	assert.Equal(t, "test value", displayedResult)
}

func TestFetchWithFeedback_SpinnerError(t *testing.T) {
	a := newTestApp(nil)
	a.newSpinner = func(_ string) (spinner, error) {
		return nil, errors.New("spinner failed")
	}

	_, err := fetchWithFeedback(
		t.Context(),
		a.newSpinner,
		"Test operation",
		func(context.Context) (string, error) {
			return "should not reach", nil
		},
		nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "spinner failed")
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

	flags := newApp().flags(&configFile, nil)

	require.Len(t, flags, 11)
}

func TestBuildCommands(t *testing.T) {
	commands := newApp().commands()

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
	os.Args = []string{appName, "--help"}

	t.Cleanup(func() { os.Args = oldArgs })

	var err error

	out := capturer.CaptureOutput(func() {
		err = Cmd("test-version")
	})

	require.NoError(t, err)
	assert.Contains(t, out, "Utility to interact with T-Mobile Home Internet gateway")
}

func TestCmd_Version(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{appName, "--version"}

	t.Cleanup(func() { os.Args = oldArgs })

	testVersion := "test-version-123"

	var err error

	out := capturer.CaptureOutput(func() {
		err = Cmd(testVersion)
	})

	require.NoError(t, err)
	assert.Contains(
		t,
		out,
		testVersion,
		"expected version output to contain %q, got: %q",
		testVersion,
		out,
	)
}

func TestCmd_UsageError(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{appName, "--invalid-flag"}

	t.Cleanup(func() { os.Args = oldArgs })

	var err error

	out := capturer.CaptureOutput(func() {
		err = Cmd("test-version")
	})

	require.Error(t, err)
	assert.Contains(t, out, "error:")
}

func TestCmd_InvalidConfig(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{appName, "login"}

	t.Cleanup(func() { os.Args = oldArgs })

	var err error

	capturer.CaptureOutput(func() {
		err = Cmd("test-version")
	})

	require.ErrorIs(t, err, ErrInvalidConfig)
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

	_, err := setupColor(t.Context(), cmd)
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

	_, err := setupColor(t.Context(), cmd)
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
	_, err := setupColor(t.Context(), cmd)
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
		Name: appName,
		OnUsageError: func(_ context.Context, cmd *cli.Command, err error, _ bool) error {
			_, _ = fmt.Fprintf(cmd.ErrWriter, "error: %v\n", err)

			return err
		},
	}

	err := app.Run(t.Context(), []string{appName, "--invalid-flag"})
	require.Error(t, err)
}

func TestDebugFlagAction(t *testing.T) {
	pterm.DisableDebugMessages()
	t.Cleanup(pterm.DisableDebugMessages)

	var configFile string

	flags := newApp().flags(&configFile, nil)

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
	err := debugFlag.Action(t.Context(), cmd, true)
	require.NoError(t, err)
}

func TestQuietFlagAction(t *testing.T) {
	pterm.EnableOutput()
	t.Cleanup(pterm.EnableOutput)

	var configFile string

	flags := newApp().flags(&configFile, nil)

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
	err := quietFlag.Action(t.Context(), cmd, true)
	require.NoError(t, err)
}

func TestGatewayInitErrors(t *testing.T) {
	tests := []struct {
		name    string
		handler func(*app) cli.ActionFunc
	}{
		{cmdLogin, func(a *app) cli.ActionFunc { return a.login }},
		{cmdStatus, func(a *app) cli.ActionFunc { return a.status }},
		{cmdSignal, func(a *app) cli.ActionFunc { return a.signal }},
		{"reboot", func(a *app) cli.ActionFunc { return a.reboot }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp(nil)
			a.initGateway = func(_ *Config) (tmhi.Gateway, error) {
				return nil, errors.New("gateway init failed")
			}

			err := tt.handler(a)(t.Context(), nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "gateway init failed")
		})
	}
}
