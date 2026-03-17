package pkg

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func Base64urlEscape(b64 string) string {
	r := strings.NewReplacer("+", "-", "/", "_", "=", ".")
	return r.Replace(b64)
}

func Sha256Hash(val1, val2 string) string {
	h := sha256.New()
	h.Write(fmt.Appendf(nil, "%s:%s", val1, val2))
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

func EchoOut(str string) {
	_, err := os.Stdout.WriteString(str + "\n")
	if err != nil {
		logrus.WithError(err).Error("error writing output")
	}
}

func EchoStatus(str string, status bool) {
	EchoOut(str + ": " + BoolEmoji(status))
}

func BoolEmoji(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}
