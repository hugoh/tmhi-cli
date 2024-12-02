package pkg

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

type GatewayI interface {
	Login() error
	Reboot() error
}

var ErrGatewayUnknown = errors.New("unsupported gateway")

func GatewayUnknownError(gateway string) error {
	return fmt.Errorf("%w: %s", ErrGatewayUnknown, gateway)
}

func NewGateway(gateway, username, password, ip string, dryRun bool) (GatewayI, error) { //nolint:ireturn
	switch gateway {
	case "NOK5G21":
		return &NokiaGateway{
			Username: username,
			Password: password,
			IP:       ip,
			DryRun:   dryRun,
		}, nil
	default:
		logrus.WithField("gateway", gateway).Error("unsupported gateway")
		return nil, GatewayUnknownError(gateway)
	}
}
