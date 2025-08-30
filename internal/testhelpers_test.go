package internal

import (
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
	"github.com/sirupsen/logrus"
)

// WithPatchedLogrusExit patches logrus's ExitFunc during the test and returns a restore func and a pointer to captured exit code.
func WithPatchedLogrusExit(t testing.TB) (func(), *int) {
	t.Helper()
	orig := logrus.StandardLogger().ExitFunc
	var code int
	logrus.StandardLogger().ExitFunc = func(c int) { code = c }
	restore := func() {
		logrus.StandardLogger().ExitFunc = orig
	}
	return restore, &code
}

func CaptureStdout(t testing.TB, fn func()) string {
	return testutil.CaptureStdout(t, fn)
}
