package pkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type Gateway interface {
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

func (gc *GatewayCommon) StatusCore() {
	resp, err := gc.Client.R().Head("/")
	EchoStatus("Web interface up", err == nil && resp.IsSuccess())
}

func (gc *GatewayCommon) AddCredentials(username, password string) {
	gc.Username = username
	gc.Password = password
}
