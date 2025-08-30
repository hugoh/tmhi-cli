package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogSetup(t *testing.T) {
	t.Run("debug logging enabled", func(t *testing.T) {
		LogSetup(true)
		assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
	})

	t.Run("debug logging disabled", func(t *testing.T) {
		LogSetup(false)
		assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
	})
}

func TestCmd_Help(t *testing.T) {
	// Prevent logrus Fatal from exiting the test process
	restore, _ := WithPatchedLogrusExit(t)
	defer restore()

	// Preserve and set args
	oldArgs := os.Args
	os.Args = []string{"tmhi-cli", "--help"}
	defer func() { os.Args = oldArgs }()

	out := CaptureStdout(t, func() {
		Cmd("test-version")
	})

	assert.Contains(t, out, "Utility to interact with T-Mobile Home Internet gateway")
}

func TestCmd_Version(t *testing.T) {
	// Prevent logrus Fatal from exiting the test process
	restore, _ := WithPatchedLogrusExit(t)
	defer restore()

	// Preserve and set args
	oldArgs := os.Args
	os.Args = []string{"tmhi-cli", "--version"}
	defer func() { os.Args = oldArgs }()

	testVersion := "test-version-123"
	out := CaptureStdout(t, func() {
		Cmd(testVersion)
	})

	assert.True(t, strings.Contains(out, testVersion), "expected version output to contain %q, got: %q", testVersion, out)
}
