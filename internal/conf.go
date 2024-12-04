package internal

import (
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfig string = ".tmhi-cli.yaml"
)

func LogSetup(debugFlag bool) {
	if debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
