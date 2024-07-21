package pkg

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type GatewayI interface {
	Login() error
	Reboot() error
}

func NewGateway(gateway, username, password, ip string, dryRun bool) (GatewayI, error) { //nolint:ireturn //FIXME:
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
		return nil, fmt.Errorf("unsupported gateway: %s", gateway)
	}
}
