package internal

import (
	"os"
	"testing"

	tmhi "github.com/hugoh/tmhi-gateway/v2"
)

type mockSpinner struct{}

func (*mockSpinner) Fail(_ ...any) {}

func (*mockSpinner) Success(_ ...any) {}

// trackingSpinner records whether Fail/Success feedback was shown, so tests
// can assert a command surfaced spinner feedback on failure.
type trackingSpinner struct {
	failCalled    bool
	successCalled bool
}

func (s *trackingSpinner) Fail(_ ...any) { s.failCalled = true }

func (s *trackingSpinner) Success(_ ...any) { s.successCalled = true }

// newTestApp returns an app wired with test doubles: a no-op spinner (the
// real one spawns goroutines that race in tests), a confirm stub returning
// its default, and a gateway factory returning gw.
func newTestApp(gw tmhi.Gateway) *app {
	a := newApp()
	a.newSpinner = func(_ string) (spinner, error) { return &mockSpinner{}, nil }
	a.confirm = func(_ string, defaultVal bool) (bool, error) { return defaultVal, nil }
	a.initGateway = func(_ *Config) (tmhi.Gateway, error) { return gw, nil }

	return a
}

func TestMain(m *testing.M) {
	// Use a temp directory as HOME to avoid reading user's config
	dir, err := os.MkdirTemp("", "tmhi-test-home")
	if err != nil {
		os.Exit(1)
	}

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir) // skipcq: GO-W1032

	code := m.Run()

	_ = os.Setenv("HOME", origHome) // skipcq: GO-W1032

	if err := os.RemoveAll(dir); err != nil && code == 0 {
		code = 1
	}

	os.Exit(code)
}
