package pkg

import (
	"errors"
	"time"

	"github.com/go-resty/resty/v2"
)

// Gateway defines the interface for T-Mobile gateway implementations.
type Gateway interface {
	NewClient(version, ip string, timeout time.Duration, retries int, debug bool)
	AddCredentials(username, password string)
	Login() error
	Reboot(dryRun bool) error
	Request(method, path string) error
	Info() error
	Status() error
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
)

// NewGatewayCommon creates a new GatewayCommon with default client.
func NewGatewayCommon() *GatewayCommon {
	return &GatewayCommon{Client: resty.New()}
}

// NewClient configures the HTTP client for the gateway.
func (gc *GatewayCommon) NewClient(version, ip string, timeout time.Duration, retries int, debug bool) {
	if gc.Client == nil {
		gc.Client = resty.New()
	}
	gc.Client.
		SetBaseURL("http://"+ip).
		SetHeader("User-Agent", "tmhi-cli/"+version).
		SetDebug(debug).
		SetTimeout(timeout)
	if retries > 0 {
		gc.Client.SetRetryCount(retries)
	}
}

// StatusCore checks if the gateway web interface is accessible.
func (gc *GatewayCommon) StatusCore() {
	resp, err := gc.Client.R().Head("/")
	EchoStatus("Web interface up", err == nil && resp.IsSuccess())
}

// AddCredentials sets the username and password for gateway authentication.
func (gc *GatewayCommon) AddCredentials(username, password string) {
	gc.Username = username
	gc.Password = password
}
