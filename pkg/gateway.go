package pkg

import (
	"errors"
	"fmt"
)

type GatewayI interface {
	Login() error
	Reboot(dryRun bool) error
}

var ErrAuthentication = errors.New("could not authenticate")

func AuthenticationError(details string) error {
	return fmt.Errorf("%w: %s", ErrAuthentication, details)
}

var ErrRebootFailed = errors.New("reboot failed")
