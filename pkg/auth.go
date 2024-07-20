package pkg

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type LoginData struct {
	Success   bool
	SID       string
	CSRFToken string
}

func login(username, password, ip string) LoginData {
	ret := LoginData{Success: false}
	resp, err := http.Get(fmt.Sprintf("http://%s/login_web_app.cgi?nonce", ip))
	if err != nil {
		log.Println("Error getting nonce:", err)
		return ret
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var nonceJson map[string]string
		json.NewDecoder(resp.Body).Decode(&nonceJson)
		logrus.Debug(fmt.Sprintf("Got nonce JSON: %+v", nonceJson))
		nonce := nonceJson["nonce"]
		logrus.Debug(fmt.Sprintf("Got nonce: %s", nonce))
		passHashInput := strings.ToLower(password)
		userPassHash := sha256Hash(username, passHashInput)
		userPassNonceHash := sha256Url(userPassHash, nonce)
		loginRequest := map[string]interface{}{
			"uri": fmt.Sprintf("http://%s/login_web_app.cgi", ip),
			"body": map[string]string{
				"userhash":      sha256Url(username, nonce),
				"RandomKeyhash": sha256Url(nonceJson["randomKey"], nonce),
				"response":      userPassNonceHash,
				"nonce":         base64urlEscape(nonce),
				"enckey":        random16bytes(),
				"enciv":         random16bytes(),
			},
		}
		logrus.Debug(fmt.Sprintf("Login request: %+v", loginRequest))
		httpPost(loginRequest, func(loginResp *http.Response) {
			defer loginResp.Body.Close()
			var loginJson map[string]string
			json.NewDecoder(loginResp.Body).Decode(&loginJson)
			logrus.Debug(fmt.Sprintf("Login response: %+v", loginJson))
			ret.SID = loginJson["sid"]
			logrus.Debug(fmt.Sprintf("SID: %s", ret.SID))
			ret.CSRFToken = loginJson["token"]
			logrus.Debug(fmt.Sprintf("Token: %s", ret.CSRFToken))
			ret.Success = true
		})
	}
	return ret
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
		logrus.Debug("Error generating random bytes:", err)
		return ""
	}
	return base64urlEscape(base64.StdEncoding.EncodeToString(b))
}
