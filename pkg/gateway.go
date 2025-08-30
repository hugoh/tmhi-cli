package pkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type GatewayI interface {
	NewClient(version, ip string, timeout time.Duration, retries int, debug bool)
	AddCredentials(username, password string)
	Login() error
	Reboot(dryRun bool) error
	Request(method, path string) error
	Info() error
	Status() error
}

type GatewayCommon struct {
	Client        *resty.Client
	Username      string
	Password      string
	Authenticated bool
}

var (
	ErrAuthentication = errors.New("could not authenticate")
	ErrNotImplemented = errors.New("command not implemented")
	ErrRebootFailed   = errors.New("reboot failed")
)

func AuthenticationError(details string) error {
	return fmt.Errorf("%w: %s", ErrAuthentication, details)
}

func NewGatewayCommon() *GatewayCommon {
	return &GatewayCommon{Client: resty.New()}
}

func (gatewayCommon *GatewayCommon) NewClient(version, ip string, timeout time.Duration, retries int, debug bool) {
	if gatewayCommon.Client == nil {
		gatewayCommon.Client = resty.New()
	}
	gatewayCommon.Client.
		SetBaseURL("http://"+ip).
		SetHeader("User-Agent", "tmhi-cli/"+version).
		SetDebug(debug).
		SetTimeout(timeout)
	if retries > 0 {
		gatewayCommon.Client.SetRetryCount(retries)
	}
}

func (gatewayCommon *GatewayCommon) StatusCore() {
	// Web interface
	resp, err := gatewayCommon.Client.R().Head("/")
	EchoStatus("Web interface up", err == nil && resp.IsSuccess())
}

func (gatewayCommon *GatewayCommon) AddCredentials(username, password string) {
	gatewayCommon.Username = username
	gatewayCommon.Password = password
}
