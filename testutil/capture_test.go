package testutil

import (
	"fmt"
	"strings"
	"testing"
)

func TestCaptureStdout(t *testing.T) {
	output := CaptureStdout(t, func() {
		fmt.Print("hello world") //nolint:forbidigo
	})

	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain 'hello world', got %q", output)
	}
}

func TestCaptureStdout_MultipleWrites(t *testing.T) {
	output := CaptureStdout(t, func() {
		fmt.Print("line1") //nolint:forbidigo
		fmt.Print("line2") //nolint:forbidigo
	})

	if !strings.Contains(output, "line1") || !strings.Contains(output, "line2") {
		t.Errorf("expected output to contain 'line1' and 'line2', got %q", output)
	}
}
