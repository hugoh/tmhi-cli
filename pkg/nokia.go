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
	Username, Password, Ip string
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
	loginNonce, loginNonceErr := http.Get(fmt.Sprintf("http://%s/login_web_app.cgi?nonce", ip))
	if loginNonceErr != nil {
		return nil, fmt.Errorf("Error getting nonce: %v", loginNonceErr)
	}
	defer loginNonce.Body.Close()

	if loginNonce.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Could not start authentication process: %s", loginNonce.Body)
	}
	var nonceResp nonceResp
	jsonErr := json.NewDecoder(loginNonce.Body).Decode(&nonceResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("Error getting nonce: %v", jsonErr)
	}
	logrus.Debugf("Got nonce: %s", nonceResp.Nonce)
	return &nonceResp, nil
}

func getCredentials(username, password, ip string, nonceResp nonceResp) (*loginResp, error) {
	passHashInput := strings.ToLower(password)
	userPassHash := internal.Sha256Hash(username, passHashInput)
	userPassNonceHash := internal.Sha256Url(userPassHash, nonceResp.Nonce)
	loginRequestUrl := fmt.Sprintf("http://%s/login_web_app.cgi", ip)
	loginRequestParams := url.Values{
		"userhash":      {internal.Sha256Url(username, nonceResp.Nonce)},
		"RandomKeyhash": {internal.Sha256Url(nonceResp.RandomKey, nonceResp.Nonce)},
		"response":      {userPassNonceHash},
		"nonce":         {internal.Base64urlEscape(nonceResp.Nonce)},
		"enckey":        {internal.Random16bytes()},
		"enciv":         {internal.Random16bytes()},
	}
	logrus.WithFields(logrus.Fields{
		"url":    loginRequestUrl,
		"params": loginRequestParams,
	}).Info("sending login request")
	login, loginErr := http.PostForm(loginRequestUrl, loginRequestParams)
	if loginErr != nil {
		logrus.WithField("err", loginErr).Error("error while making login request")
		return nil, fmt.Errorf("Could not authenticate: %v", loginErr)
	}
	defer login.Body.Close()
	if !internal.HttpRequestSuccessful(login) {
		logrus.WithFields(internal.LogHttpResponseFields(login)).Error("error while making login request")
		return nil, fmt.Errorf("Could not authenticate: %s", login.Body)
	}

	var loginResp loginResp
	jsonErr := json.NewDecoder(login.Body).Decode(&loginResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("Error parsing login response: %v", jsonErr)
	}
	logrus.Debugf("Got login response: %s", loginResp)
	return &loginResp, nil
}

func (n *NokiaGateway) Login() error {
	nonceResp, nonceErr := getNonce(n.Ip)
	if nonceErr != nil {
		return fmt.Errorf("Error getting nonce: %v", nonceErr)
	}

	loginResp, loginErr := getCredentials(n.Username, n.Password, n.Ip, *nonceResp)
	if loginErr != nil {
		return fmt.Errorf("Could not authenticate: %v", loginErr)
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
		return fmt.Errorf("Cannot reboot without successful login flow: %v", err)
	}

	rebootRequestUri := fmt.Sprintf("http://%s/reboot_web_app.cgi", n.Ip)
	formData := url.Values{
		"csrf_token": {n.credentials.CSRFToken},
	}
	req, err := http.NewRequest("POST", rebootRequestUri, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("Error creating reboot request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cookie := fmt.Sprintf("sid=%s", n.credentials.SID)
	req.Header.Set("Cookie", cookie)

	logrus.WithFields(logrus.Fields{
		"url":    rebootRequestUri,
		"cookie": cookie,
		"params": formData,
	}).Debug("reboot request prepared")

	if !n.DryRun {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Error sending reboot request: %v", err)
		}
		defer resp.Body.Close()

		logrus.WithFields(internal.LogHttpResponseFields(resp)).Debug("reboot response")
		if internal.HttpRequestSuccessful(resp) {
			logrus.Info("successfully requested gateway rebooted")
		}
	}
	return nil
}
