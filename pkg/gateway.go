package pkg

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/pterm/pterm"
)

// Gateway defines the interface for T-Mobile gateway implementations.
type Gateway interface {
	NewClient(cfg *GatewayConfig)
	AddCredentials(username, password string)
	Login() error
	Reboot(dryRun bool) error
	Request(method, path string) error
	Info() error
	Status() error
	Signal() error
}

// GatewayCommon provides shared functionality for gateway implementations.
type GatewayCommon struct {
	Client        *resty.Client
	Username      string
	Password      string
	Authenticated bool
}

// Sentinel errors for gateway operations.
var (
	// ErrAuthentication indicates an authentication failure.
	ErrAuthentication = errors.New("could not authenticate")
	// ErrNotImplemented indicates an unsupported operation.
	ErrNotImplemented = errors.New("command not implemented")
	// ErrRebootFailed indicates a reboot operation failed.
	ErrRebootFailed = errors.New("reboot failed")
	// ErrSignalFailed indicates a signal operation failed.
	ErrSignalFailed = errors.New("signal failed")
	// ErrNoResponse indicates no response available from mock.
	ErrNoResponse = errors.New("no response available")
)

// NewGatewayCommon creates a new GatewayCommon with default client.
func NewGatewayCommon() *GatewayCommon {
	return &GatewayCommon{Client: resty.New()}
}

// NewClient configures the HTTP client for the gateway.
func (gc *GatewayCommon) NewClient(cfg *GatewayConfig) {
	if gc.Client == nil {
		gc.Client = resty.New()
	}

	gc.Client.
		SetBaseURL("http://" + cfg.IP).
		SetDebug(cfg.Debug).
		SetTimeout(cfg.Timeout)

	if cfg.Retries > 0 {
		gc.Client.SetRetryCount(cfg.Retries)
	}
}

// StatusCore checks if the gateway web interface is accessible.
func (gc *GatewayCommon) StatusCore() {
	spinner, _ := pterm.DefaultSpinner.Start("Checking web interface...")

	resp, err := gc.Client.R().Head("/")

	switch {
	case err == nil && resp.IsSuccess():
		spinner.Success("Web interface up")
	case err != nil:
		spinner.Fail("Web interface down: " + err.Error())
	default:
		spinner.Fail(fmt.Sprintf("Web interface down: status %d", resp.StatusCode()))
	}
}

// AddCredentials sets the username and password for gateway authentication.
func (gc *GatewayCommon) AddCredentials(username, password string) {
	gc.Username = username
	gc.Password = password
}
