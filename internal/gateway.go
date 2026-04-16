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
	gwConfig := &tmhi.GatewayConfig{
		IP:       cfg.IP,
		Username: cfg.Username,
		Password: cfg.Password,
		Timeout:  cfg.Timeout,
		Retries:  cfg.Retries,
		DryRun:   cfg.DryRun,
		Debug:    cfg.Debug,
	}

	switch cfg.Model {
	case ARCADYAN:
		return tmhi.NewArcadyanGateway(gwConfig), nil
	case NOK5G21:
		return tmhi.NewNokiaGateway(gwConfig), nil
	default:
		pterm.Error.Println("unsupported gateway:", cfg.Model)

		return nil, fmt.Errorf("%w: %s", errUnknownGateway, cfg.Model)
	}
}
