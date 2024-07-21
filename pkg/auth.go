package pkg

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type LoginData struct {
	Success   bool
	SID       string
	CSRFToken string
}

type NonceResp struct {
	Nonce     string `json:"nonce"`
	Pubkey    string `json:"pubkey"`
	RandomKey string `json:"randomKey"`
}

type LoginResp struct {
	Sid       string `json:"sid"`
	CsrfToken string `json:"token"`
}

func getNonce(ip string) (*NonceResp, error) {
	loginNonce, loginNonceErr := http.Get(fmt.Sprintf("http://%s/login_web_app.cgi?nonce", ip))
	if loginNonceErr != nil {
		return nil, fmt.Errorf("Error getting nonce: %v", loginNonceErr)
	}
	defer loginNonce.Body.Close()

	if loginNonce.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Could not start authentication process: %s", loginNonce.Body)
	}
	var nonceResp NonceResp
	jsonErr := json.NewDecoder(loginNonce.Body).Decode(&nonceResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("Error getting nonce: %v", jsonErr)
	}
	logrus.Debugf("Got nonce: %s", nonceResp.Nonce)
	return &nonceResp, nil
}

func httpRequestSuccessful(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func getCredentials(username, password, ip string, nonceResp NonceResp) (*LoginResp, error) {
	passHashInput := strings.ToLower(password)
	userPassHash := sha256Hash(username, passHashInput)
	userPassNonceHash := sha256Url(userPassHash, nonceResp.Nonce)
	loginRequestUrl := fmt.Sprintf("http://%s/login_web_app.cgi", ip)
	loginRequestParams := url.Values{}
	loginRequestParams.Add("userhash", sha256Url(username, nonceResp.Nonce))
	loginRequestParams.Add("RandomKeyhash", sha256Url(nonceResp.RandomKey, nonceResp.Nonce))
	loginRequestParams.Add("response", userPassNonceHash)
	loginRequestParams.Add("nonce", base64urlEscape(nonceResp.Nonce))
	loginRequestParams.Add("enckey", random16bytes())
	loginRequestParams.Add("enciv", random16bytes())
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
	if !httpRequestSuccessful(login) {
		body, err := io.ReadAll(login.Body)
		if err != nil {
			logrus.Errorf("Error reading HTTP body: %v", err)
		}
		logrus.WithFields(logrus.Fields{
			"status": login.StatusCode,
			"body":   string(body),
		}).Error("error while making login request")
		return nil, fmt.Errorf("Could not authenticate: %s", login.Body)
	}

	var loginResp LoginResp
	jsonErr := json.NewDecoder(login.Body).Decode(&loginResp)
	if jsonErr != nil {
		return nil, fmt.Errorf("Error parsing login response: %v", jsonErr)
	}
	logrus.Debugf("Got login response: %s", loginResp)
	return &loginResp, nil
}

func Login(username, password, ip string) (LoginData, error) {
	ret := LoginData{}
	nonceResp, nonceErr := getNonce(ip)
	if nonceErr != nil {
		return ret, fmt.Errorf("Error getting nonce: %v", nonceErr)
	}

	loginResp, loginErr := getCredentials(username, password, ip, *nonceResp)
	if loginErr != nil {
		return ret, fmt.Errorf("Could not authenticate: %v", loginErr)
	}
	ret.SID = loginResp.Sid
	ret.CSRFToken = loginResp.CsrfToken
	ret.Success = true
	logrus.Debugf("Authenticated: %v", ret)
	return ret, nil
}

func base64urlEscape(b64 string) string {
	r := strings.NewReplacer("+", "-", "/", "_", "=", ".")
	return r.Replace(b64)
}

func sha256Hash(val1, val2 string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%s", val1, val2)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func sha256Url(val1, val2 string) string {
	return base64urlEscape(sha256Hash(val1, val2))
}

func random16bytes() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64urlEscape(base64.StdEncoding.EncodeToString(b))
}
