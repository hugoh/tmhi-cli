package internal

import (
	"os"
	"testing"
)

type mockSpinner struct{}

func (m *mockSpinner) Fail(_ ...any) {}

func (m *mockSpinner) Success(_ ...any) {}

func (m *mockSpinner) Stop() error {
	return nil
}

func TestMain(m *testing.M) {
	// Use a temp directory as HOME to avoid reading user's config
	dir, err := os.MkdirTemp("", "tmhi-test-home")
	if err != nil {
		os.Exit(1)
	}

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)

	// Override spinnerFunc with a mock to prevent data races from async goroutines
	spinnerFunc = func(_ string) (spinner, error) {
		return &mockSpinner{}, nil
	}

	// Override confirmDialog to prevent data races from async keyboard listeners
	confirmDialog = func(_ string, defaultVal bool) (bool, error) {
		return defaultVal, nil
	}

	code := m.Run()

	_ = os.Setenv("HOME", origHome)

	if err := os.RemoveAll(dir); err != nil {
		os.Exit(1)
	}

	os.Exit(code)
}
