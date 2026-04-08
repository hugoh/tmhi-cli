// Package testutil provides testing utilities for tmhi-cli.
package testutil

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/pterm/pterm"
)

// CaptureStdout captures stdout during function execution and returns it.
func CaptureStdout(tb testing.TB, action func()) string {
	tb.Helper()

	old := os.Stdout

	reader, writer, err := os.Pipe()
	if err != nil {
		tb.Fatalf("os.Pipe failed: %v", err)
	}

	os.Stdout = writer
	pterm.SetDefaultOutput(os.Stdout)

	action()

	pterm.SetDefaultOutput(old)
	os.Stdout = old

	err = writer.Close()
	if err != nil {
		tb.Fatalf("closing write pipe failed: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		tb.Fatalf("reading from pipe failed: %v", err)
	}

	if err := reader.Close(); err != nil {
		tb.Fatalf("closing read pipe failed: %v", err)
	}

	out := buf.String()
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	out = ansi.ReplaceAllString(out, "")

	return out
}
