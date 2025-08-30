package pkg

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type NokiaGateway struct {
	*GatewayCommon

	credentials nokiaLoginData
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

func NewNokiaGateway() *NokiaGateway {
	return &NokiaGateway{GatewayCommon: &GatewayCommon{}}
}

func (l *nokiaLoginResp) success() bool {
	return l.Sid != "" && l.CsrfToken != ""
}

func (n *NokiaGateway) Login() error {
	if n.Authenticated {
		return nil
	}
	nonceResp, nonceErr := n.getNonce()
	if nonceErr != nil {
		return fmt.Errorf("error getting nonce: %w", nonceErr)
	}

	loginResp, loginErr := n.getCredentials(*nonceResp)
	if loginErr != nil {
		return fmt.Errorf("login failed: %w", loginErr)
	}
	n.credentials.SID = loginResp.Sid
	n.credentials.CSRFToken = loginResp.CsrfToken
	n.Authenticated = true
	n.Client.SetHeader("Cookie", "sid="+n.credentials.SID)
	logrus.WithField("credentials", n.credentials).Debug("authenticated")
	return nil
}

func (n *NokiaGateway) Reboot(dryRun bool) error {
	if err := n.Login(); err != nil {
		return fmt.Errorf("cannot reboot without successful login flow: %w", err)
	}

	rebootRequestURL := "/reboot_web_app.cgi"
	formData := map[string]string{
		"csrf_token": n.credentials.CSRFToken,
	}
	req := n.Client.R().
		SetFormData(formData)

	logrus.WithFields(logrus.Fields{
		"url":    rebootRequestURL,
		"cookie": "sid=" + n.credentials.SID,
		"params": formData,
	}).Debug("reboot request prepared")

	if dryRun {
		logrus.Info("simulating gateway rebooted")
		return nil
	}

	resp, err := req.Execute("POST", req.URL)
	if err != nil {
		return fmt.Errorf("error sending reboot request: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"status": resp.StatusCode(),
		"body":   resp.String(),
	}).Debug("reboot response")

	if resp.IsError() {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode(),
			"body":   resp.String(),
		}).Error("reboot failed")
		return ErrRebootFailed
	}
	logrus.Info("successfully requested gateway rebooted")
	return nil
}

func (n *NokiaGateway) Request(_, _ string) error {
	return ErrNotImplemented
}

func (n *NokiaGateway) Info() error {
	return ErrNotImplemented
}

func (n *NokiaGateway) Status() error {
	n.StatusCore()
	return nil
}

func (n *NokiaGateway) getCredentials(nonceResp nonceResp) (*nokiaLoginResp, error) {
	passHashInput := strings.ToLower(n.Password)
	userPassHash := Sha256Hash(n.Username, passHashInput)
	userPassNonceHash := Sha256Url(userPassHash, nonceResp.Nonce)
	reqParams := map[string]string{
		"userhash":      Sha256Url(n.Username, nonceResp.Nonce),
		"RandomKeyhash": Sha256Url(nonceResp.RandomKey, nonceResp.Nonce),
		"response":      userPassNonceHash,
		"nonce":         Base64urlEscape(nonceResp.Nonce),
		"enckey":        Random16bytes(),
		"enciv":         Random16bytes(),
	}

	reqURL := "/login_web_app.cgi"
	logrus.WithFields(logrus.Fields{
		"url":    reqURL,
		"params": reqParams,
	}).Info("sending login request")

	var loginResp nokiaLoginResp
	resp, err := n.Client.R().
		SetFormData(reqParams).
		SetResult(loginResp).
		Post(reqURL)
	if err != nil {
		logrus.WithError(err).Error("error while making login request")
		return nil, AuthenticationError("authentication failed: " + err.Error())
	}
	if resp.IsError() {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode(),
			"body":   resp.String(),
		}).Error("error while making login request")
		return nil, AuthenticationError("authentication failed: " + resp.String())
	}

	logrus.WithField("response", loginResp).Debug("got login response")
	var authErr error
	if loginResp.success() {
		authErr = nil
	} else {
		authErr = AuthenticationError("authentication failed: no valid credentials returned")
	}
	return &loginResp, authErr
}

func (n *NokiaGateway) getNonce() (*nonceResp, error) {
	var nonceResp nonceResp
	resp, err := n.Client.R().
		SetResult(nonceResp).
		Get("/login_web_app.cgi?nonce")
	if err != nil {
		return nil, fmt.Errorf("error getting nonce: %w", err)
	}
	if resp.IsError() {
		return nil, AuthenticationError("authentication failed: " + resp.String())
	}
	logrus.WithField("nonce", nonceResp.Nonce).Debug("got nonce")
	return &nonceResp, nil
}
