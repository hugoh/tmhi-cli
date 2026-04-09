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
	"github.com/pterm/pterm"
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
	pterm.Debug.Println("sending login request:", reqPath)

	spinner, _ := pterm.DefaultSpinner.Start("Logging in...")

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
		spinner.Fail("Login request failed")

		return fmt.Errorf("login request failed: failed to decode login response: %w", err)
	}

	if resp.IsError() {
		spinner.Fail(fmt.Sprintf("Login failed (status %d)", resp.StatusCode()))

		return NewAuthError(resp.StatusCode(), resp.String())
	}

	if loginResp.Auth.Token == "" {
		spinner.Fail("Login response missing auth token")

		return NewAuthError(0, "login response missing auth token")
	}

	a.credentials = arcadianLoginData{
		Expiration: loginResp.Auth.Expiration,
		Token:      loginResp.Auth.Token,
	}
	a.Client.SetAuthToken(a.credentials.Token)
	a.Authenticated = true

	_ = spinner.WithRemoveWhenDone().Stop()

	return nil
}

// Reboot restarts the Arcadyan gateway. If dryRun is true, it logs without executing.
func (a *ArcadyanGateway) Reboot(dryRun bool) error {
	err := a.Login()
	if err != nil {
		return fmt.Errorf("cannot reboot without successful login flow: %w", err)
	}

	rebootRequestPath := "/TMI/v1/gateway/reset?set=reboot"

	pterm.Debug.Println("reboot request prepared:", rebootRequestPath)

	if dryRun {
		pterm.Info.Println("Dry run - would send reboot request")

		return nil
	}

	spinner, _ := pterm.DefaultSpinner.Start("Rebooting gateway...")

	resp, err := a.Client.R().
		Post(rebootRequestPath)
	if err != nil {
		spinner.Fail("Reboot request failed")

		return fmt.Errorf("reboot request failed: %w", err)
	}

	if !resp.IsSuccess() {
		spinner.Fail(fmt.Sprintf("Reboot failed (status %d)", resp.StatusCode()))

		return NewGatewayError("reboot", resp.StatusCode(), resp.String(), ErrRebootFailed)
	}

	spinner.Success("Reboot command sent successfully")

	return nil
}

// Info retrieves and displays gateway information.
func (a *ArcadyanGateway) Info() error {
	return a.Request("GET", InfoURL)
}

// Request makes an HTTP request to the gateway and displays the response.
func (a *ArcadyanGateway) Request(method, path string) error {
	pterm.Debug.Println("making request:", method, path)

	resp, err := a.Client.R().Execute(method, path)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	contentType := resp.Header().Get("Content-Type")
	body := resp.Body()

	if strings.HasPrefix(contentType, "application/json") {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			pterm.DefaultBasicText.Println(string(body))
		} else {
			pterm.DefaultBasicText.Println(prettyJSON.String())
		}
	} else {
		pterm.DefaultBasicText.Println(string(body))
	}

	return nil
}

// Status checks and displays the gateway connection status.
func (a *ArcadyanGateway) Status() error {
	a.StatusCore()

	spinner, _ := pterm.DefaultSpinner.Start("Fetching registration status...")

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
		spinner.Fail("Request failed: " + err.Error())

		return NewGatewayError("status", 0, "failed to get registration status", err)
	}

	regStatus := "unknown"
	if info.IsSuccess() {
		regStatus = result.Signal.Generic.Registration
	}

	spinner.Info("Registration status: " + regStatus)

	if !info.IsSuccess() {
		return NewGatewayError(
			"status",
			info.StatusCode(),
			ErrSignalFailed.Error(),
			ErrSignalFailed,
		)
	}

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
		return NewGatewayError("signal", 0, "failed to get signal info", err)
	}

	if !info.IsSuccess() {
		return NewGatewayError(
			"signal",
			info.StatusCode(),
			ErrSignalFailed.Error(),
			ErrSignalFailed,
		)
	}

	a.printSignalResult(result.Signal)

	return nil
}

func (a *ArcadyanGateway) printSignalResult(sig signalResult) {
	if sig.FourG != nil {
		a.printSignalMetrics(
			"4G LTE Signal",
			&sig.FourG.signalData,
			[]string{"eNBID", strconv.Itoa(sig.FourG.ENBID)},
		)
	}

	if sig.FiveG != nil {
		var fiveGExtras [][]string
		if sig.FiveG.AntennaUsed != "" {
			fiveGExtras = append(fiveGExtras, []string{"Antenna", sig.FiveG.AntennaUsed})
		}

		fiveGExtras = append(fiveGExtras, []string{"gNBID", strconv.Itoa(sig.FiveG.GNBID)})

		a.printSignalMetrics("5G Signal", &sig.FiveG.signalData, fiveGExtras...)
	}

	pterm.DefaultHeader.Println("Generic Info")

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"APN", sig.Generic.APN},
		{"IPv6", strconv.FormatBool(sig.Generic.HasIPv6)},
		{"Registration", sig.Generic.Registration},
		{"Roaming", strconv.FormatBool(sig.Generic.Roaming)},
	}
	if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		pterm.Error.Println("Failed to render table:", err)
	}
}

func (a *ArcadyanGateway) printSignalMetrics(
	header string,
	metrics *signalData,
	extras ...[]string,
) {
	const signalMetricRows = 6

	rater := signal.NewRater()

	pterm.DefaultHeader.Println(header)

	tableData := make(pterm.TableData, 0, 2+len(extras)+signalMetricRows)
	tableData = append(tableData,
		[]string{"Metric", "Value"},
		[]string{"Signal bars", fmt.Sprintf("%.0f", metrics.Bars)},
	)

	for _, extra := range extras {
		tableData = append(tableData, extra)
	}

	const signalFormat = "%v %u - %q %s"

	tableData = append(tableData,
		[]string{"Bands", fmt.Sprintf("%v", metrics.Bands)},
		[]string{"RSRP", rater.FormatWith(signalFormat, rater.RateRSRP(metrics.RSRP))},
		[]string{"RSRQ", rater.FormatWith(signalFormat, rater.RateRSRQ(metrics.RSRQ))},
		[]string{"RSSI", rater.FormatWith(signalFormat, rater.RateRSSI(metrics.RSSI))},
		[]string{"SINR", rater.FormatWith(signalFormat, rater.RateSINR(metrics.SINR))},
		[]string{"CID", strconv.Itoa(metrics.CID)},
	)

	if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		pterm.Error.Println("Failed to render table:", err)
	}
}

func (a *ArcadyanGateway) isLoggedIn() bool {
	now := int(time.Now().Unix())

	return a.credentials.Token != "" && a.credentials.Expiration > now
}
