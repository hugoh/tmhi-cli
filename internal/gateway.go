package internal

import (
	"errors"
	"fmt"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
)

var ErrUnknownGateway = errors.New("unknown gateway")

func getGateway(model, username, password, ip string, debug bool) (pkg.GatewayI, error) { //nolint:ireturn
	LogSetup(debug)
	var gateway pkg.GatewayI
	switch model {
	case "ARCADYAN":
		gateway = pkg.NewArcadyanGateway(username, password, ip)
	case "NOK5G21":
		gateway = pkg.NewNokiaGateway(username, password, ip)
	default:
		logrus.WithField("gateway", model).Error("unsupported gateway")
		return nil, fmt.Errorf("%w: %s", ErrUnknownGateway, model)
	}
	return gateway, nil
}
