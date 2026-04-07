package internal

import (
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
	"github.com/sirupsen/logrus"
)

// WithPatchedLogrusExit patches logrus's ExitFunc during the test.
func WithPatchedLogrusExit(tb testing.TB) (func(), *int) {
	tb.Helper()
	orig := logrus.StandardLogger().ExitFunc
	var code int
	logrus.StandardLogger().ExitFunc = func(c int) { code = c }
	restore := func() {
		logrus.StandardLogger().ExitFunc = orig
	}

	return restore, &code
}

func CaptureStdout(tb testing.TB, fn func()) string {
	tb.Helper()

	return testutil.CaptureStdout(tb, fn)
}
