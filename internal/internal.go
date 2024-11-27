package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func HTTPRequestSuccessful(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func LogHTTPResponseFields(resp *http.Response) logrus.Fields {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("error reading HTTP body")
	}
	return logrus.Fields{
		"status": resp.StatusCode,
		"body":   string(body),
	}
}

func Base64urlEscape(b64 string) string {
	r := strings.NewReplacer("+", "-", "/", "_", "=", ".")
	return r.Replace(b64)
}

func Sha256Hash(val1, val2 string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%s", val1, val2)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Sha256Url(val1, val2 string) string {
	return Base64urlEscape(Sha256Hash(val1, val2))
}

func Random16bytes() string {
	const length = 16
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return Base64urlEscape(base64.StdEncoding.EncodeToString(b))
}
