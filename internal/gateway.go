package internal

import (
	"errors"
	"fmt"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var ErrUnknownGateway = errors.New("unknown gateway")

func getGateway(model, username, password, ip string, debug bool) (pkg.GatewayI, error) { //nolint:ireturn
	LogSetup(debug)
	var gateway pkg.GatewayI
	switch model {
	case "NOK5G21":
		gateway = pkg.NewNokiaGateway(username, password, ip)
	default:
		logrus.WithField("gateway", model).Error("unsupported gateway")
		return nil, fmt.Errorf("%w: %s", ErrUnknownGateway, model)
	}
	return gateway, nil
}

func getGatewayFromCtxOrFail(cCtx *cli.Context) pkg.GatewayI { //nolint:ireturn
	gateway, err := getGateway(cCtx.String(ConfigModel),
		cCtx.String(ConfigUsername),
		cCtx.String(ConfigPassword),
		cCtx.String(ConfigIP),
		cCtx.Bool(ConfigDebug))
	if err != nil {
		logrus.WithError(err).Fatal("unsupported gateway")
		// NOTREACHED
	}
	return gateway
}
