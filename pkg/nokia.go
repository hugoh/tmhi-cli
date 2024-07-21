package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hugoh/thmi-cli/internal"
	"github.com/sirupsen/logrus"
)

type NokiaGateway struct {
	Username, Password, IP string
	DryRun                 bool
	credentials            loginData
}

type nonceResp struct {
	Nonce     string `json:"nonce"`
	Pubkey    string `json:"pubkey"`
	RandomKey string `json:"randomKey"`
}

type loginData struct {
	Success   bool
	SID       string
	CSRFToken string
}

type loginResp struct {
	Sid       string `json:"sid"`
	CsrfToken string `json:"token"`
}

func getNonce(ip string) (*nonceResp, error) {
	reqURL := fmt.Sprintf("http://%s/login_web_app.cgi?nonce", ip)
	loginNonce, loginNonceErr := http.Get(reqURL) /* #nosec G107 */ //nolint:gosec,noctx,nolintlint //FIXME:
	if loginNonceErr != nil {
		return nil, fmt.Errorf("error getting nonce: %w", loginNonceErr)
	}
	defer loginNonce.Body.Close()

	if loginNonce.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not start authentication process: %s", loginNonce.Body)
	}
	var nonceResp nonceResp
	jsonErr := json.NewDecoder(loginNonce.Body).Decode(&nonceResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("error getting nonce: %w", jsonErr)
	}
	logrus.Debugf("Got nonce: %s", nonceResp.Nonce)
	return &nonceResp, nil
}

func getCredentials(username, password, ip string, nonceResp nonceResp) (*loginResp, error) {
	passHashInput := strings.ToLower(password)
	userPassHash := internal.Sha256Hash(username, passHashInput)
	userPassNonceHash := internal.Sha256Url(userPassHash, nonceResp.Nonce)
	reqURL := fmt.Sprintf("http://%s/login_web_app.cgi", ip)
	reqParams := url.Values{
		"userhash":      {internal.Sha256Url(username, nonceResp.Nonce)},
		"RandomKeyhash": {internal.Sha256Url(nonceResp.RandomKey, nonceResp.Nonce)},
		"response":      {userPassNonceHash},
		"nonce":         {internal.Base64urlEscape(nonceResp.Nonce)},
		"enckey":        {internal.Random16bytes()},
		"enciv":         {internal.Random16bytes()},
	}
	logrus.WithFields(logrus.Fields{
		"url":    reqURL,
		"params": reqParams,
	}).Info("sending login request")
	login, loginErr := http.PostForm(reqURL, reqParams) /* #nosec G107 */ //nolint:gosec,noctx,nolintlint //FIXME:
	if loginErr != nil {
		logrus.WithField("err", loginErr).Error("error while making login request")
		return nil, fmt.Errorf("could not authenticate: %w", loginErr)
	}
	defer login.Body.Close()
	if !internal.HTTPRequestSuccessful(login) {
		logrus.WithFields(internal.LogHTTPResponseFields(login)).Error("error while making login request")
		return nil, fmt.Errorf("could not authenticate: %s", login.Body)
	}

	var loginResp loginResp
	jsonErr := json.NewDecoder(login.Body).Decode(&loginResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("error parsing login response: %w", jsonErr)
	}
	logrus.Debugf("Got login response: %s", loginResp)
	return &loginResp, nil
}

func (n *NokiaGateway) Login() error {
	nonceResp, nonceErr := getNonce(n.IP)
	if nonceErr != nil {
		return fmt.Errorf("error getting nonce: %w", nonceErr)
	}

	loginResp, loginErr := getCredentials(n.Username, n.Password, n.IP, *nonceResp)
	if loginErr != nil {
		return fmt.Errorf("could not authenticate: %w", loginErr)
	}
	n.credentials.SID = loginResp.Sid
	n.credentials.CSRFToken = loginResp.CsrfToken
	n.credentials.Success = true
	logrus.Debugf("Authenticated: %v", n.credentials)
	return nil
}

func (n *NokiaGateway) ensureLoggedIn() error {
	if !n.credentials.Success {
		return n.Login()
	}
	return nil
}

func (n *NokiaGateway) Reboot() error {
	err := n.ensureLoggedIn()
	if err != nil {
		return fmt.Errorf("cannot reboot without successful login flow: %w", err)
	}

	rebootRequestURL := fmt.Sprintf("http://%s/reboot_web_app.cgi", n.IP)
	formData := url.Values{
		"csrf_token": {n.credentials.CSRFToken},
	}
	req, err := http.NewRequest(http.MethodPost, rebootRequestURL, strings.NewReader(formData.Encode())) //nolint:noctx
	if err != nil {
		return fmt.Errorf("error creating reboot request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cookie := "sid=" + n.credentials.SID
	req.Header.Set("Cookie", cookie)

	logrus.WithFields(logrus.Fields{
		"url":    rebootRequestURL,
		"cookie": cookie,
		"params": formData,
	}).Debug("reboot request prepared")

	if !n.DryRun {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending reboot request: %w", err)
		}
		defer resp.Body.Close()

		logrus.WithFields(internal.LogHTTPResponseFields(resp)).Debug("reboot response")
		if internal.HTTPRequestSuccessful(resp) {
			logrus.Info("successfully requested gateway rebooted")
		}
	}
	return nil
}
