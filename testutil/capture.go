package testutil

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func CaptureStdout(tb testing.TB, fn func()) string {
	tb.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		tb.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	if err := w.Close(); err != nil {
		tb.Fatalf("closing write pipe failed: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		tb.Fatalf("reading from pipe failed: %v", err)
	}
	if err := r.Close(); err != nil {
		tb.Fatalf("closing read pipe failed: %v", err)
	}
	return buf.String()
}
