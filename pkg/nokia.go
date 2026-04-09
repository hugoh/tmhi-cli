package pkg

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

// NokiaGateway implements Gateway for Nokia-based T-Mobile gateways.
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
	SID       string
	CSRFToken string
}

type nokiaLoginResp struct {
	Success   int    `json:"success"`
	Reason    int    `json:"reason"`
	Sid       string `json:"sid"`
	CsrfToken string `json:"token"`
}

// NewNokiaGateway creates a new Nokia gateway instance.
func NewNokiaGateway() *NokiaGateway {
	return &NokiaGateway{GatewayCommon: &GatewayCommon{}}
}

func (l *nokiaLoginResp) success() bool {
	return l.Sid != "" && l.CsrfToken != ""
}

// Login authenticates with the Nokia gateway.
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
	pterm.Debug.Println("authenticated", n.credentials)

	return nil
}

// Reboot restarts the Nokia gateway. If dryRun is true, it logs without executing.
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

	pterm.Debug.Println("reboot request prepared:",
		rebootRequestURL,
		"cookie=", "sid="+n.credentials.SID,
		"params=", formData,
	)

	if dryRun {
		pterm.Info.Println("simulating gateway rebooted")

		return nil
	}

	resp, err := req.Execute("POST", req.URL)
	if err != nil {
		return fmt.Errorf("error sending reboot request: %w", err)
	}

	pterm.Debug.Println("reboot response:", resp.StatusCode(), resp.String())

	if resp.IsError() {
		pterm.Error.Println("reboot failed", resp.StatusCode(), resp.String())

		return NewGatewayError("reboot", resp.StatusCode(), resp.String(), ErrRebootFailed)
	}

	pterm.Info.Println("successfully requested gateway rebooted")

	return nil
}

// Request is not implemented for Nokia gateway.
func (n *NokiaGateway) Request(_, _ string) error {
	return ErrNotImplemented
}

// Info is not implemented for Nokia gateway.
func (n *NokiaGateway) Info() error {
	return ErrNotImplemented
}

// Status checks and displays the gateway connection status.
func (n *NokiaGateway) Status() error {
	n.StatusCore()

	return nil
}

// Signal is not implemented for Nokia gateway.
func (n *NokiaGateway) Signal() error {
	return ErrNotImplemented
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
	pterm.Info.Println("sending login request", reqURL, reqParams)

	var loginResp nokiaLoginResp

	resp, err := n.Client.R().
		SetFormData(reqParams).
		SetResult(&loginResp).
		Post(reqURL)
	if err != nil {
		pterm.Error.Println("error while making login request", err)

		return nil, NewAuthError(0, err.Error())
	}

	if resp.IsError() {
		pterm.Error.Println("error while making login request", resp.StatusCode(), resp.String())

		return nil, NewAuthError(resp.StatusCode(), resp.String())
	}

	pterm.Debug.Println("got login response", loginResp)

	var authErr error
	if loginResp.success() {
		authErr = nil
	} else {
		authErr = NewAuthError(0, "no valid credentials returned")
	}

	return &loginResp, authErr
}

func (n *NokiaGateway) getNonce() (*nonceResp, error) {
	var resp nonceResp

	_, err := n.Client.R().
		SetResult(&resp).
		Get("/login_web_app.cgi?nonce")
	if err != nil {
		return nil, fmt.Errorf("error getting nonce: %w", err)
	}

	pterm.Debug.Println("got nonce:", resp.Nonce)

	return &resp, nil
}
