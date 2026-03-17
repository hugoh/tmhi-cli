package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
)

// ErrUnknownGateway is returned when an unsupported gateway model is specified.
var ErrUnknownGateway = errors.New("unknown gateway")

// getGateway returns a gateway instance based on the model type.
// Returns interface because this is a factory function that creates
// different concrete types (ArcadyanGateway, NokiaGateway) based on input.
//
//nolint:ireturn
func getGateway(version, model, username, password, ip string, timeout time.Duration, retries int, debug bool,
) (pkg.Gateway, error) {
	LogSetup(debug)
	var gateway pkg.Gateway
	switch model {
	case "ARCADYAN":
		gateway = pkg.NewArcadyanGateway()
	case "NOK5G21":
		gateway = pkg.NewNokiaGateway()
	default:
		logrus.WithField("gateway", model).Error("unsupported gateway")
		return nil, fmt.Errorf("%w: %s", ErrUnknownGateway, model)
	}

	gateway.NewClient(version, ip, timeout, retries, debug)
	gateway.AddCredentials(username, password)
	return gateway, nil
}
