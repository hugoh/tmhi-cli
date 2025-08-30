package pkg

import (
	"testing"

	"github.com/hugoh/tmhi-cli/testutil"
)

func CaptureStdout(t testing.TB, fn func()) string {
	return testutil.CaptureStdout(t, fn)
}
