package pkg

import (
	"context"
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
	credentials            nokiaLoginData
	client                 *http.Client
}

type nonceResp struct {
	Nonce     string `json:"nonce"`
	Pubkey    string `json:"pubkey"`
	RandomKey string `json:"randomKey"`
}

type nokiaLoginData struct {
	Success   bool
	SID       string
	CSRFToken string
}

type nokiaLoginResp struct {
	Success   int    `json:"success"`
	Reason    int    `json:"reason"`
	Sid       string `json:"sid"`
	CsrfToken string `json:"token"`
}

var ErrNokiaAuthenticationProcessStart = errors.New("could not start authentication process")

func NewNokiaGateway(username, password, ip string) *NokiaGateway {
	return &NokiaGateway{
		Username: username,
		Password: password,
		IP:       ip,
		client:   &http.Client{},
	}
}

func AuthenticationProcessStartError(details string) error {
	return fmt.Errorf("%w: %s", ErrNokiaAuthenticationProcessStart, details)
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
	logrus.WithField("nonce", nonceResp.Nonce).Debug("got nonce")
	return &nonceResp, nil
}

func (l *nokiaLoginResp) success() bool {
	return l.Sid != "" && l.CsrfToken != ""
}

func (n *NokiaGateway) Login() error {
	nonceResp, nonceErr := getNonce(n.IP)
	if nonceErr != nil {
		return fmt.Errorf("error getting nonce: %w", nonceErr)
	}

	loginResp, loginErr := n.getCredentials(*nonceResp)
	if loginErr != nil {
		return fmt.Errorf("login failed: %w", loginErr)
	}
	n.credentials.SID = loginResp.Sid
	n.credentials.CSRFToken = loginResp.CsrfToken
	n.credentials.Success = true
	logrus.WithField("credentials", n.credentials).Debug("authenticated")
	return nil
}

func (n *NokiaGateway) Reboot(dryRun bool) error {
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

	return doReboot(n.client, req, dryRun)
}

func (n *NokiaGateway) ensureLoggedIn() error {
	if !n.credentials.Success {
		return n.Login()
	}
	return nil
}

func (n *NokiaGateway) getCredentials(nonceResp nonceResp) (*nokiaLoginResp, error) {
	passHashInput := strings.ToLower(n.Password)
	userPassHash := Sha256Hash(n.Username, passHashInput)
	userPassNonceHash := Sha256Url(userPassHash, nonceResp.Nonce)
	reqURL := "http://" + n.IP + "/login_web_app.cgi"
	reqParams := url.Values{
		"userhash":      {Sha256Url(n.Username, nonceResp.Nonce)},
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

	// Build the POST request manually to avoid PostForm (noctx)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		reqURL,
		strings.NewReader(reqParams.Encode()),
	)
	if err != nil {
		logrus.WithError(err).Error("error creating login request")
		return nil, AuthenticationError(err.Error())
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	login, loginErr := n.client.Do(req)
	if loginErr != nil {
		logrus.WithError(loginErr).Error("error while making login request")
		return nil, AuthenticationError(loginErr.Error())
	}
	defer login.Body.Close()
	if !HTTPRequestSuccessful(login) {
		logrus.WithFields(LogHTTPResponseFields(login)).Error("error while making login request")
		return nil, AuthenticationError(GetBody(login))
	}

	var loginResp nokiaLoginResp
	jsonErr := json.NewDecoder(login.Body).Decode(&loginResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("error parsing login response: %w", jsonErr)
	}
	logrus.WithField("response", loginResp).Debug("got login response")
	var authErr error
	if loginResp.success() {
		authErr = nil
	} else {
		authErr = AuthenticationError("no valid credentials returned")
	}
	return &loginResp, authErr
}
