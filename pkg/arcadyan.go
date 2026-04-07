// Package pkg provides gateway implementations for T-Mobile Home Internet devices.
package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	signal "github.com/hugoh/cellular-signal"
	"github.com/sirupsen/logrus"
)

// ArcadyanGateway implements Gateway for Arcadyan-based T-Mobile gateways.
type ArcadyanGateway struct {
	*GatewayCommon

	credentials arcadianLoginData
}

type arcadianLoginData struct {
	Expiration int
	Token      string
}

// InfoURL is the endpoint for gateway information.
const InfoURL = "/TMI/v1/gateway/?get=all"

// NewArcadyanGateway creates a new Arcadyan gateway instance.
func NewArcadyanGateway() *ArcadyanGateway {
	ret := &ArcadyanGateway{GatewayCommon: NewGatewayCommon()}
	ret.Client.SetHeader("Accept", "application/json")

	return ret
}

// Login authenticates with the Arcadyan gateway.
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
		return fmt.Errorf(
			"%w: unexpected status %d: %s",
			ErrAuthentication,
			resp.StatusCode(),
			resp.String(),
		)
	}

	if loginResp.Auth.Token == "" {
		return fmt.Errorf("%w: login response missing auth token", ErrAuthentication)
	}

	a.credentials = arcadianLoginData{
		Expiration: loginResp.Auth.Expiration,
		Token:      loginResp.Auth.Token,
	}
	a.Client.SetAuthToken(a.credentials.Token)
	a.Authenticated = true

	return nil
}

// Reboot restarts the Arcadyan gateway. If dryRun is true, it logs without executing.
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

// Info retrieves and displays gateway information.
func (a *ArcadyanGateway) Info() error {
	return a.Request("GET", InfoURL)
}

// Request makes an HTTP request to the gateway and displays the response.
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

// Status checks and displays the gateway connection status.
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

type signalData struct {
	Bands []string `json:"bands"`
	Bars  float64  `json:"bars"`
	CID   int      `json:"cid"`
	RSRP  int      `json:"rsrp"`
	RSRQ  int      `json:"rsrq"`
	RSSI  int      `json:"rssi"`
	SINR  int      `json:"sinr"`
}

type fourGSignal struct {
	signalData

	ENBID int `json:"eNBID"` //nolint:tagliatelle
}

type fiveGSignal struct {
	signalData

	AntennaUsed string `json:"antennaUsed"`
	GNBID       int    `json:"gNBID"` //nolint:tagliatelle
}

type signalResult struct {
	FourG   *fourGSignal `json:"4g"`
	FiveG   *fiveGSignal `json:"5g"`
	Generic struct {
		APN          string `json:"apn"`
		HasIPv6      bool   `json:"hasIPv6"`
		Registration string `json:"registration"`
		Roaming      bool   `json:"roaming"`
	} `json:"generic"`
}

// Signal retrieves and displays signal strength information.
func (a *ArcadyanGateway) Signal() error {
	var result struct {
		Signal signalResult `json:"signal"`
	}

	info, err := a.Client.R().SetResult(&result).Get(InfoURL)
	if err != nil {
		return fmt.Errorf("failed to get signal info: %w", err)
	}

	if !info.IsSuccess() {
		return fmt.Errorf("%w: status %d", ErrSignalFailed, info.StatusCode())
	}

	a.printSignalResult(result.Signal)

	return nil
}

func (a *ArcadyanGateway) printSignalResult(sig signalResult) {
	if sig.FourG != nil {
		a.printSignalMetrics(
			"4G LTE Signal",
			&sig.FourG.signalData,
			"eNBID: "+strconv.Itoa(sig.FourG.ENBID),
		)
	}

	if sig.FiveG != nil {
		var fiveGExtras []string
		if sig.FiveG.AntennaUsed != "" {
			fiveGExtras = append(fiveGExtras, "Antenna: "+sig.FiveG.AntennaUsed)
		}
		fiveGExtras = append(fiveGExtras, "gNBID: "+strconv.Itoa(sig.FiveG.GNBID))

		a.printSignalMetrics("5G Signal", &sig.FiveG.signalData, fiveGExtras...)
	}

	EchoOut("")
	EchoOut("=== Generic Info ===")
	EchoOut("APN: " + sig.Generic.APN)
	EchoOut(fmt.Sprintf("IPv6: %t", sig.Generic.HasIPv6))
	EchoOut("Registration: " + sig.Generic.Registration)
	EchoOut(fmt.Sprintf("Roaming: %t", sig.Generic.Roaming))
}

func (a *ArcadyanGateway) printSignalMetrics(header string, metrics *signalData, extras ...string) {
	rater := signal.NewRater()
	EchoOut(fmt.Sprintf("=== %s ===", header))
	EchoOut(fmt.Sprintf("Signal bars: %.0f", metrics.Bars))
	for _, extra := range extras {
		EchoOut(extra)
	}
	EchoOut(fmt.Sprintf("Bands: %v", metrics.Bands))
	EchoOut(rater.Format(rater.RateRSRP(metrics.RSRP)))
	EchoOut(rater.Format(rater.RateRSRQ(metrics.RSRQ)))
	EchoOut(rater.Format(rater.RateRSSI(metrics.RSSI)))
	EchoOut(rater.Format(rater.RateSINR(metrics.SINR)))
	EchoOut(fmt.Sprintf("CID: %d", metrics.CID))
}

func (a *ArcadyanGateway) isLoggedIn() bool {
	now := int(time.Now().Unix())

	return a.credentials.Token != "" && a.credentials.Expiration > now
}
