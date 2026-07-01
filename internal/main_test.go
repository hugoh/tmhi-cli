package internal

import (
	"context"
	"os"
	"testing"

	tmhi "github.com/hugoh/tmhi-gateway/v2"
	"go.uber.org/goleak"
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
	a.confirm = func(_ context.Context, _ string, defaultVal bool) (bool, error) { return defaultVal, nil }
	a.initGateway = func(_ *Config) (tmhi.Gateway, error) { return gw, nil }

	return a
}

// ptermConfirmLeakIgnores returns goleak options for ptermConfirm's
// background goroutine (internal/cmd.go), which keeps reading stdin after
// ctx cancellation since pterm offers no way to interrupt it; it blocks
// forever on the underlying keyboard listener rather than exiting, so it's
// ignored here rather than treated as a leak.
func ptermConfirmLeakIgnores() []goleak.Option {
	return []goleak.Option{
		goleak.IgnoreTopFunction("atomicgo.dev/keyboard.Listen"),
		goleak.IgnoreTopFunction("atomicgo.dev/keyboard.Listen.func3"),
	}
}

// TestMain sets a temp HOME for the test binary (avoiding the real user's
// config) and wires up goleak so any goroutine leak fails the run.
// goleak.VerifyTestMain exits the process itself, so the temp dir is best
// left for the OS to reclaim rather than cleaned up via defer here.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tmhi-test-home")
	if err != nil {
		os.Exit(1)
	}

	_ = os.Setenv("HOME", dir) // skipcq: GO-W1032

	goleak.VerifyTestMain(m, ptermConfirmLeakIgnores()...)
}
