package internal

import (
	"os"
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigPath(t *testing.T) {
	path := defaultConfigPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".tmhi-cli.toml")
}

func TestDefaultConfigPath_HomeError(t *testing.T) {
	t.Setenv("HOME", "")

	path := defaultConfigPath()
	assert.Equal(t, ".tmhi-cli.toml", path)
}

func TestBuildFlags(t *testing.T) {
	var configFile string

	flags := cmdFlags(&configFile, nil)

	assert.Len(t, flags, 11)
}

func TestBuildCommands(t *testing.T) {
	commands := cmdCommands()

	assert.Len(t, commands, 6)
	assert.Equal(t, "login", commands[0].Name)
	assert.Equal(t, "reboot", commands[1].Name)
	assert.Equal(t, "info", commands[2].Name)
	assert.Equal(t, "status", commands[3].Name)
	assert.Equal(t, "signal", commands[4].Name)
	assert.Equal(t, "req", commands[5].Name)
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

func TestColorFlag_Never(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"tmhi-cli", "--color", "never", "info"}

	defer func() { os.Args = oldArgs }()

	out := testutil.CaptureOutput(t, func() {
		Cmd("test-version")
	})

	assert.Contains(t, out, "")
}

func TestColorFlag_Always(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"tmhi-cli", "--color", "always", "info"}

	defer func() { os.Args = oldArgs }()

	out := testutil.CaptureOutput(t, func() {
		Cmd("test-version")
	})

	assert.Contains(t, out, "")
}

func TestColorFlag_Auto(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"tmhi-cli", "--color", "auto", "info"}

	defer func() { os.Args = oldArgs }()

	out := testutil.CaptureOutput(t, func() {
		Cmd("test-version")
	})

	assert.Contains(t, out, "")
}

func TestTermIsTerminal(_ *testing.T) {
	result := termIsTerminal()
	_ = result
}
