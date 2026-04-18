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

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir) // skipcq: GO-W1032

	// Override spinnerFunc with a mock to prevent data races from async goroutines
	spinnerFunc = func(_ string) (spinner, error) {
		return &mockSpinner{}, nil
	}

	// Override confirmDialog to prevent data races from async keyboard listeners
	confirmDialog = func(_ string, defaultVal bool) (bool, error) {
		return defaultVal, nil
	}

	code := m.Run()

	_ = os.Setenv("HOME", origHome) // skipcq: GO-W1032

	if err := os.RemoveAll(dir); err != nil {
		os.Exit(1)
	}

	os.Exit(code)
}
