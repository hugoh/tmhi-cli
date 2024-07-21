package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/hugoh/thmi-cli/internal"
	"github.com/sirupsen/logrus"
)

type NokiaGateway struct {
	Username, Password, Ip string
	DryRun                 bool
}

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
	userPassHash := internal.Sha256Hash(username, passHashInput)
	userPassNonceHash := internal.Sha256Url(userPassHash, nonceResp.Nonce)
	loginRequestUrl := fmt.Sprintf("http://%s/login_web_app.cgi", ip)
	loginRequestParams := url.Values{}
	loginRequestParams.Add("userhash", internal.Sha256Url(username, nonceResp.Nonce))
	loginRequestParams.Add("RandomKeyhash", internal.Sha256Url(nonceResp.RandomKey, nonceResp.Nonce))
	loginRequestParams.Add("response", userPassNonceHash)
	loginRequestParams.Add("nonce", internal.Base64urlEscape(nonceResp.Nonce))
	loginRequestParams.Add("enckey", internal.Random16bytes())
	loginRequestParams.Add("enciv", internal.Random16bytes())
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

func (n *NokiaGateway) Login() (LoginData, error) {
	ret := LoginData{}
	nonceResp, nonceErr := getNonce(n.Ip)
	if nonceErr != nil {
		return ret, fmt.Errorf("Error getting nonce: %v", nonceErr)
	}

	loginResp, loginErr := getCredentials(n.Username, n.Password, n.Ip, *nonceResp)
	if loginErr != nil {
		return ret, fmt.Errorf("Could not authenticate: %v", loginErr)
	}
	ret.SID = loginResp.Sid
	ret.CSRFToken = loginResp.CsrfToken
	ret.Success = true
	logrus.Debugf("Authenticated: %v", ret)
	return ret, nil
}

func (n *NokiaGateway) Reboot() error {
	loginData, err := n.Login()
	if !loginData.Success || err != nil {
		return fmt.Errorf("Cannot reboot without successful login flow: %v", err)
	}
	rebootRequest := map[string]interface{}{
		"uri": fmt.Sprintf("http://%s/reboot_web_app.cgi", n.Ip),
		"headers": map[string]string{
			"Cookie": fmt.Sprintf("sid=%s", loginData.SID),
		},
		"body": map[string]string{
			"csrf_token": loginData.CSRFToken,
		},
	}
	logrus.Debug(fmt.Sprintf("Reboot request: %+v", rebootRequest))
	rebootMsg := "T-Mobile Internet Router reboot successfully requested"
	if !n.DryRun {
		// httpPost(rebootRequest, func(rebootResp *http.Response) {
		// 	defer rebootResp.Body.Close()
		// 	var respData map[string]interface{}
		// 	json.NewDecoder(rebootResp.Body).Decode(&respData)
		// 	if rebootResp.StatusCode == http.StatusOK {
		// 		logrus.Debug(fmt.Sprintf("Reboot response: %+v", respData))
		// 		log.Println(rebootMsg)
		// 	} else {
		// 		log.Println(fmt.Sprintf("Reboot request failed: %+v", respData))
		// 	}
		// })
	} else {
		logrus.Infof("[DRY-RUN] %s [/DRY-RUN]", rebootMsg)
	}
	return nil
}

func httpPost(request map[string]interface{}, handler func(resp *http.Response)) {
	uri := request["uri"].(string)
	body := request["body"].(map[string]string)
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Println("Error marshalling request body:", err)
		return
	}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return
	}
	for k, v := range request["headers"].(map[string]string) {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making HTTP request:", err)
		return
	}
	handler(resp)
}
