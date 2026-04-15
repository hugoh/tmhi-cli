package internal

import (
	"errors"
	"fmt"

	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/pterm/pterm"
)

var errUnknownGateway = errors.New("unknown gateway")

//nolint:ireturn
func getGateway(cfg *Config) (tmhi.Gateway, error) {
	var gateway tmhi.Gateway

	switch cfg.Model {
	case ARCADYAN:
		gateway = tmhi.NewArcadyanGateway()
	case NOK5G21:
		gateway = tmhi.NewNokiaGateway()
	default:
		pterm.Error.Println("unsupported gateway:", cfg.Model)

		return nil, fmt.Errorf("%w: %s", errUnknownGateway, cfg.Model)
	}

	gateway.NewClient(&tmhi.GatewayConfig{
		IP:       cfg.IP,
		Username: cfg.Username,
		Password: cfg.Password,
		Timeout:  cfg.Timeout,
		Retries:  cfg.Retries,
		DryRun:   cfg.DryRun,
		Debug:    cfg.Debug,
	})
	gateway.AddCredentials(cfg.Username, cfg.Password)

	return gateway, nil
}
