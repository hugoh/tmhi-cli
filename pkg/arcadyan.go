package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ArcadyanGateway struct {
	*GatewayCommon

	credentials arcadianLoginData
}

type arcadianLoginData struct {
	Expiration int
	Token      string
}

const InfoURL = "/TMI/v1/gateway/?get=all"

func NewArcadyanGateway() *ArcadyanGateway {
	ret := &ArcadyanGateway{GatewayCommon: NewGatewayCommon()}
	ret.Client.SetHeader("Accept", "application/json")
	return ret
}

func (a *ArcadyanGateway) Login() error {
	if a.isLoggedIn() {
		return nil
	}

	bodyMap := map[string]string{
		"username": a.Username,
		"password": a.Password,
	}

	reqPath := "/TMI/v1/auth/login"
	logrus.WithFields(logrus.Fields{
		"url":    reqPath,
		"params": bodyMap,
	}).Debug("sending login request")

	var loginResp struct {
		Auth struct {
			Expiration       int
			RefreshCountLeft int
			RefreshCountMax  int
			Token            string
		}
	}
	resp, err := a.Client.R().
		SetBody(bodyMap).
		SetResult(&loginResp).
		Post(reqPath)
	if err != nil {
		return fmt.Errorf("login request failed: failed to decode login response: %w", err)
	}

	if resp.IsError() {
		return AuthenticationError(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode(), resp.String()))
	}

	if loginResp.Auth.Token == "" {
		return AuthenticationError("login response missing auth token")
	}

	a.credentials = arcadianLoginData{
		Expiration: loginResp.Auth.Expiration,
		Token:      loginResp.Auth.Token,
	}
	a.Client.SetAuthToken(a.credentials.Token)
	a.Authenticated = true

	return nil
}

func (a *ArcadyanGateway) Reboot(dryRun bool) error {
	err := a.Login()
	if err != nil {
		return fmt.Errorf("cannot reboot without successful login flow: %w", err)
	}

	rebootRequestPath := "/TMI/v1/gateway/reset?set=reboot"

	logrus.WithFields(logrus.Fields{
		"url": rebootRequestPath,
	}).Debug("reboot request prepared")

	if dryRun {
		logrus.Info("Dry run - would send reboot request")
		return nil
	}

	resp, err := a.Client.R().
		Post(rebootRequestPath)
	if err != nil {
		return fmt.Errorf("reboot request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("%w: status %d: %s", ErrRebootFailed, resp.StatusCode(), resp.String())
	}

	return nil
}

func (a *ArcadyanGateway) Info() error {
	return a.Request("GET", InfoURL)
}

func (a *ArcadyanGateway) Request(method, path string) error {
	logrus.WithFields(logrus.Fields{
		"method": method,
		"url":    path,
	}).Debug("making request")
	resp, err := a.Client.R().Execute(method, path)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	contentType := resp.Header().Get("Content-Type")
	body := resp.Body()
	if strings.HasPrefix(contentType, "application/json") {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			EchoOut(string(body))
		} else {
			EchoOut(prettyJSON.String())
		}
	} else {
		EchoOut(string(body))
	}
	return nil
}

func (a *ArcadyanGateway) Status() error {
	a.StatusCore()

	// Info
	var result struct {
		Signal struct {
			Generic struct {
				Registration string
			}
		}
	}
	info, err := a.Client.R().SetResult(&result).Get(InfoURL)
	if err != nil {
		return fmt.Errorf("failed to get registration status: %w",
			err)
	}
	regStatus := "unknown"
	if info.IsSuccess() {
		regStatus = result.Signal.Generic.Registration
	}
	EchoOut("Registration status: " + regStatus)

	return nil
}

func (a *ArcadyanGateway) isLoggedIn() bool {
	now := int(time.Now().Unix())
	return a.credentials.Token != "" && a.credentials.Expiration > now
}
