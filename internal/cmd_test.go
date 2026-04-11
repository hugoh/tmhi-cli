package internal

import (
	"context"
	"os"
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

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
