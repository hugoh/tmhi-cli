package pkg

import (
	"io"
	"net/http"
)

func GetBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	details := string(body)
	if err != nil {
		details = details + "\n" + err.Error()
	}
	return details
}
