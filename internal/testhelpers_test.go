package internal

import (
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
)

// WithPatchedExit patches exit behavior during the test (no-op shim).
func WithPatchedExit(tb testing.TB) (func(), *int) {
	tb.Helper()
	// No-op: pterm handles output; tests shouldn't depend on exit codes here.
	restore := func() {}

	var code int

	return restore, &code
}

func CaptureStdout(tb testing.TB, fn func()) string {
	tb.Helper()

	return testutil.CaptureStdout(tb, fn)
}
