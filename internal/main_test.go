package internal

import (
	"os"
	"testing"
)

type mockSpinner struct{}

func (*mockSpinner) Fail(_ ...any) {}

func (*mockSpinner) Success(_ ...any) {}

func (*mockSpinner) Stop() error {
	return nil
}

func TestMain(m *testing.M) {
	// Use a temp directory as HOME to avoid reading user's config
	dir, err := os.MkdirTemp("", "tmhi-test-home")
	if err != nil {
		os.Exit(1)
	}

	if err := os.Setenv("HOME", dir); err != nil {
		os.Exit(1)
	}

	// Override spinnerFunc with a mock to prevent data races from async goroutines
	spinnerFunc = func(_ string) (spinner, error) {
		return &mockSpinner{}, nil
	}

	// Override confirmDialog to prevent data races from async keyboard listeners
	confirmDialog = func(_ string, defaultVal bool) (bool, error) {
		return defaultVal, nil
	}

	code := m.Run()

	if err := os.RemoveAll(dir); err != nil {
		os.Exit(1)
	}

	os.Exit(code)
}
