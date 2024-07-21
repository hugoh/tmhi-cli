package pkg

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type GatewayI interface {
	Login() (LoginData, error)
	Reboot() error
}

func NewGateway(gateway, username, password, ip string, dryRun bool) (GatewayI, error) {
	switch gateway {
	case "NOK5G21":
		return &NokiaGateway{
			Username: username,
			Password: password,
			Ip:       ip,
			DryRun:   dryRun,
		}, nil
	default:
		logrus.WithField("gateway", gateway).Error("unsupported gateway")
		return nil, fmt.Errorf("Unsupported gateway: %s", gateway)
	}
}
