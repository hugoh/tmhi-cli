package internal

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Use a temp directory as HOME to avoid reading user's config
	dir, err := os.MkdirTemp("", "tmhi-test-home")
	if err != nil {
		os.Exit(1)
	}

	os.Setenv("HOME", dir)

	code := m.Run()

	os.RemoveAll(dir)
	os.Exit(code)
}
