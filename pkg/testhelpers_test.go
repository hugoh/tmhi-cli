package pkg

import (
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
)

func CaptureStdout(tb testing.TB, fn func()) string {
	tb.Helper()

	return testutil.CaptureStdout(tb, fn)
}
