package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
	Success   int    `json:"success"`
	Reason    int    `json:"reason"`
	Sid       string `json:"sid"`
	CsrfToken string `json:"token"`
}

var ErrAuthenticationProcessStart = errors.New("could not start authentication process")

func AuthenticationProcessStartError(details string) error {
	return fmt.Errorf("%w: %s", ErrAuthenticationProcessStart, details)
}

func getNonce(ip string) (*nonceResp, error) {
	reqURL := fmt.Sprintf("http://%s/login_web_app.cgi?nonce", ip)
	loginNonce, loginNonceErr := http.Get(reqURL) /* #nosec G107 */ //nolint:gosec,noctx,nolintlint //FIXME:
	if loginNonceErr != nil {
		return nil, fmt.Errorf("error getting nonce: %w", loginNonceErr)
	}
	defer loginNonce.Body.Close()

	if loginNonce.StatusCode != http.StatusOK {
		return nil, AuthenticationProcessStartError(GetBody(loginNonce))
	}
	var nonceResp nonceResp
	jsonErr := json.NewDecoder(loginNonce.Body).Decode(&nonceResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("error getting nonce: %w", jsonErr)
	}
	logrus.WithField("nonce", nonceResp.Nonce).Debug("Got nonce")
	return &nonceResp, nil
}

var ErrAuthentication = errors.New("could not authenticate")

func AuthenticationError(details string) error {
	return fmt.Errorf("%w: %s", ErrAuthentication, details)
}

func (l *loginResp) success() bool {
	return l.Sid != "" && l.CsrfToken != ""
}

func getCredentials(username, password, ip string, nonceResp nonceResp) (*loginResp, error) {
	passHashInput := strings.ToLower(password)
	userPassHash := Sha256Hash(username, passHashInput)
	userPassNonceHash := Sha256Url(userPassHash, nonceResp.Nonce)
	reqURL := fmt.Sprintf("http://%s/login_web_app.cgi", ip)
	reqParams := url.Values{
		"userhash":      {Sha256Url(username, nonceResp.Nonce)},
		"RandomKeyhash": {Sha256Url(nonceResp.RandomKey, nonceResp.Nonce)},
		"response":      {userPassNonceHash},
		"nonce":         {Base64urlEscape(nonceResp.Nonce)},
		"enckey":        {Random16bytes()},
		"enciv":         {Random16bytes()},
	}
	logrus.WithFields(logrus.Fields{
		"url":    reqURL,
		"params": reqParams,
	}).Info("sending login request")
	login, loginErr := http.PostForm(reqURL, reqParams) /* #nosec G107 */ //nolint:gosec,noctx,nolintlint //FIXME:
	if loginErr != nil {
		logrus.WithError(loginErr).Error("error while making login request")
		return nil, AuthenticationError(loginErr.Error())
	}
	defer login.Body.Close()
	if !HTTPRequestSuccessful(login) {
		logrus.WithFields(LogHTTPResponseFields(login)).Error("error while making login request")
		return nil, AuthenticationError(GetBody(login))
	}

	var loginResp loginResp
	jsonErr := json.NewDecoder(login.Body).Decode(&loginResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("error parsing login response: %w", jsonErr)
	}
	logrus.WithField("response", loginResp).Debug("got login response")
	var err error
	if loginResp.success() {
		err = nil
	} else {
		err = AuthenticationError("no valid credentials returned")
	}
	return &loginResp, err
}

func (n *NokiaGateway) Login() error {
	nonceResp, nonceErr := getNonce(n.IP)
	if nonceErr != nil {
		return fmt.Errorf("error getting nonce: %w", nonceErr)
	}

	loginResp, loginErr := getCredentials(n.Username, n.Password, n.IP, *nonceResp)
	if loginErr != nil {
		return fmt.Errorf("login failed: %w", loginErr)
	}
	n.credentials.SID = loginResp.Sid
	n.credentials.CSRFToken = loginResp.CsrfToken
	n.credentials.Success = true
	logrus.WithField("credentials", n.credentials).Debug("uthenticated")
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

		logrus.WithFields(LogHTTPResponseFields(resp)).Debug("reboot response")
		if HTTPRequestSuccessful(resp) {
			logrus.Info("successfully requested gateway rebooted")
		}
	}
	return nil
}
