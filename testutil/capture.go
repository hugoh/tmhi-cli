// Package testutil provides test helpers with terminal output
package testutil

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/pterm/pterm"
)

// CaptureOutput captures stdout and stderr.
func CaptureOutput(tb testing.TB, action func()) string {
	tb.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		tb.Fatalf("os.Pipe failed: %v", err)
	}

	rErr, wErr, err := os.Pipe()
	if err != nil {
		tb.Fatalf("os.Pipe failed: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	pterm.SetDefaultOutput(os.Stdout)

	pterm.DefaultSpinner.Writer = os.Stderr

	action()

	pterm.SetDefaultOutput(oldStdout)

	pterm.DefaultSpinner.Writer = oldStderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	_ = wOut.Close()
	_ = wErr.Close()

	var bufOut, bufErr bytes.Buffer
	if _, err := io.Copy(&bufOut, rOut); err != nil {
		tb.Fatalf("reading from stdout pipe failed: %v", err)
	}

	if _, err := io.Copy(&bufErr, rErr); err != nil {
		tb.Fatalf("reading from stderr pipe failed: %v", err)
	}

	_ = rOut.Close()
	_ = rErr.Close()

	out := bufOut.String() + bufErr.String()
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	out = ansi.ReplaceAllString(out, "")

	return out
}
