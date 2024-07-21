package internal

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Fatal logs the error to the standard output and exits with status 1.
func FatalIfError(err error) {
	if err == nil {
		return
	}
	logrus.Fatal(err)
	os.Exit(1)
}
