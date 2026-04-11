package internal

import (
	"errors"
	"fmt"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/pterm/pterm"
)

// ErrUnknownGateway is returned when an unsupported gateway model is specified.
var ErrUnknownGateway = errors.New("unknown gateway")

// getGateway returns a gateway instance based on the model type.
//
//nolint:ireturn
func getGateway(cfg *Config) (pkg.Gateway, error) {
	var gateway pkg.Gateway

	switch cfg.Model {
	case "ARCADYAN":
		gateway = pkg.NewArcadyanGateway()
	case "NOK5G21":
		gateway = pkg.NewNokiaGateway()
	default:
		pterm.Error.Println("unsupported gateway:", cfg.Model)

		return nil, fmt.Errorf("%w: %s", ErrUnknownGateway, cfg.Model)
	}

	gateway.NewClient(&pkg.GatewayConfig{
		IP:       cfg.IP,
		Username: cfg.Username,
		Password: cfg.Password,
		Timeout:  cfg.Timeout,
		Retries:  cfg.Retries,
		Debug:    cfg.Debug,
	})
	gateway.AddCredentials(cfg.Username, cfg.Password)

	return gateway, nil
}
