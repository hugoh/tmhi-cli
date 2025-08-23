package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ArcadyanGateway struct {
	Username, Password, IP string
	client                 *http.Client
	credentials            arcadianLoginData
}

type arcadianLoginData struct {
	Expiration int
	Token      string
}

type arcadianLoginResp struct {
	Auth struct {
		Expiration       int    `json:"expiration"`
		RefreshCountLeft int    `json:"refreshCountLeft"`
		RefreshCountMax  int    `json:"refreshCountMax"`
		Token            string `json:"token"`
	} `json:"auth"`
}

func NewArcadyanGateway(username, password, ip string) *ArcadyanGateway {
	return &ArcadyanGateway{
		client:   &http.Client{},
		Username: username,
		Password: password,
		IP:       ip,
	}
}

func (a *ArcadyanGateway) Login() error {
	// Prepare request body
	bodyMap := map[string]string{
		"username": a.Username,
		"password": a.Password,
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("failed to marshal login request body: %w", err)
	}

	// Send POST request
	reqURL := "http://" + a.IP + "/TMI/v1/auth/login"
	logrus.WithFields(logrus.Fields{
		"url":    reqURL,
		"params": bodyMap,
	}).Debug("sending login request")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status %d and failed to read body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body)) //nolint:err113
	}

	// Parse response
	var loginResp arcadianLoginResp
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}
	logrus.WithField("response", loginResp).Debug("got login response")

	// Populate return type
	a.credentials = arcadianLoginData{
		Expiration: loginResp.Auth.Expiration,
		Token:      loginResp.Auth.Token,
	}

	return nil
}

func (a *ArcadyanGateway) Reboot(dryRun bool) error {
	err := a.ensureLoggedIn()
	if err != nil {
		return fmt.Errorf("cannot reboot without successful login flow: %w", err)
	}

	rebootRequestURL := "http://" + a.IP + "/TMI/v1/gateway/reset?set=reboot"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, rebootRequestURL, nil)
	if err != nil {
		return fmt.Errorf("error creating reboot request: %w", err)
	}
	a.addRequestCredentials(req)

	logrus.WithFields(logrus.Fields{
		"url": rebootRequestURL,
	}).Debug("reboot request prepared")

	return doReboot(a.client, req, dryRun)
}

func (a *ArcadyanGateway) Info() error {
	return a.Request("GET", "/TMI/v1/gateway/?get=all", false, false)
}

func (a *ArcadyanGateway) Request(method, path string, loginFirst bool, details bool) error {
	if loginFirst {
		if err := a.ensureLoggedIn(); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	fullURL := "http://" + a.IP + path
	req, err := http.NewRequestWithContext(context.Background(), method, fullURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if a.credentials.Token != "" {
		a.addRequestCredentials(req)
	}

	logrus.WithFields(logrus.Fields{
		"method": method,
		"url":    fullURL,
	}).Debug("making request")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if details {
		fields := logrus.Fields{}
		for k, v := range resp.Header {
			fields[k] = strings.Join(v, ", ")
		}
		logrus.WithFields(fields).Info("HTTP response headers")
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		EchoOut(string(body))
	} else {
		EchoOut(prettyJSON.String())
	}

	return nil
}

func (a *ArcadyanGateway) addRequestCredentials(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.credentials.Token)
}

func (a *ArcadyanGateway) ensureLoggedIn() error {
	now := int(time.Now().Unix())
	if a.credentials.Token == "" || a.credentials.Expiration <= now {
		return a.Login()
	}
	return nil
}
