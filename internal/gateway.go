package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
)

var (
	ErrUnknownGateway     = errors.New("unknown gateway")
	ErrMissingCredentials = errors.New("missing required credentials")
)

//nolint:ireturn
func getGateway(version, model, username, password, ip string, timeout time.Duration, retries int, debug bool,
) (pkg.GatewayI, error) {
	LogSetup(debug)
	var gateway pkg.GatewayI
	switch model {
	case "ARCADYAN":
		gateway = pkg.NewArcadyanGateway()
	case "NOK5G21":
		gateway = pkg.NewNokiaGateway()
	default:
		logrus.WithField("gateway", model).Error("unsupported gateway")
		return nil, fmt.Errorf("%w: %s", ErrUnknownGateway, model)
	}

	if username == "" || password == "" {
		return nil, ErrMissingCredentials
	}
	gateway.NewClient(version, ip, timeout, retries, debug)
	gateway.AddCredentials(username, password)
	return gateway, nil
}
