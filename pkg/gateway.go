package pkg

import (
	"errors"
	"fmt"
)

type GatewayI interface {
	Login() error
	Reboot(dryRun bool) error
	Request(method, path string, loginFirst bool, details bool) error
	Info() error
}

var (
	ErrAuthentication = errors.New("could not authenticate")
	ErrNotImplemented = errors.New("command not implemented")
)

func AuthenticationError(details string) error {
	return fmt.Errorf("%w: %s", ErrAuthentication, details)
}

var ErrRebootFailed = errors.New("reboot failed")
