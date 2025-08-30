package pkg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func strBody(s string) io.ReadCloser { return io.NopCloser(bytes.NewBufferString(s)) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       strBody(body),
	}
}

func textResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       strBody(body),
	}
}

func newArcadyan(client *resty.Client, username, password, token string, exp time.Time) *ArcadyanGateway {
	ag := &ArcadyanGateway{
		GatewayCommon: &GatewayCommon{
			Username: username,
			Password: password,
			Client:   client,
		},
	}
	if token != "" {
		ag.credentials = arcadianLoginData{
			Token:      token,
			Expiration: int(exp.Unix()),
		}
	}
	return ag
}

func TestArcadyanGateway_Login_Success(t *testing.T) {
	body := `{"auth":{"expiration":1234567890,"refreshCountLeft":5,"refreshCountMax":10,"token":"testtoken"}}`
	client := NewTestClient(jsonResponse(http.StatusOK, body), nil)

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	assert.NoError(t, err)
	assert.Equal(t, 1234567890, gw.credentials.Expiration)
	assert.Equal(t, "testtoken", gw.credentials.Token)
}

func TestArcadyanGateway_Reboot_Failure(t *testing.T) {
	client := NewTestClient(textResponse(http.StatusInternalServerError, "server error"), nil)

	gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

	err := gw.Reboot(false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reboot failed")
}

func TestArcadyanGateway_Login_Non200Status(t *testing.T) {
	client := NewTestClient(textResponse(http.StatusUnauthorized, "unauthorized"), nil)

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}

func TestArcadyanGateway_Login_InvalidJSON(t *testing.T) {
	client := NewTestClient(jsonResponse(http.StatusOK, "{invalid json"), nil)

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode login response")
}

func TestArcadyanGateway_Login_HTTPClientError(t *testing.T) {
	client := NewTestClient(nil, errors.New("network error"))

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login request failed")
}

func TestArcadyanGateway_Info_Success(t *testing.T) {
	client := NewTestClient(jsonResponse(http.StatusOK, `{"system": {"model": "TEST123"}}`), nil)

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Info()
	assert.NoError(t, err)
}

func TestArcadyanGateway_isLoggedIn(t *testing.T) {
	t.Run("valid login", func(t *testing.T) {
		gw := &ArcadyanGateway{
			GatewayCommon: NewGatewayCommon(),
			credentials: arcadianLoginData{
				Expiration: int(time.Now().Add(1 * time.Hour).Unix()),
				Token:      "valid",
			},
		}
		assert.True(t, gw.isLoggedIn())
	})

	t.Run("expired token", func(t *testing.T) {
		gw := &ArcadyanGateway{
			GatewayCommon: NewGatewayCommon(),
			credentials: arcadianLoginData{
				Expiration: int(time.Now().Add(-1 * time.Hour).Unix()),
				Token:      "expired",
			},
		}
		assert.False(t, gw.isLoggedIn())
	})

	t.Run("no token", func(t *testing.T) {
		gw := &ArcadyanGateway{GatewayCommon: NewGatewayCommon()}
		assert.False(t, gw.isLoggedIn())
	})
}

func TestArcadyanGateway_Request_Methods(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		client := NewTestClient(jsonResponse(http.StatusOK, `{"status": "ok"}`), nil)

		gw := newArcadyan(client, "", "", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("GET", "/test")
		assert.NoError(t, err)
	})

	t.Run("POST request", func(t *testing.T) {
		client := NewTestClient(jsonResponse(http.StatusOK, `{"status": "created"}`), nil)

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("POST", "/test")
		assert.NoError(t, err)
	})

	t.Run("non-JSON response", func(t *testing.T) {
		client := NewTestClient(textResponse(http.StatusOK, "plain text response"), nil)

		gw := newArcadyan(client, "", "", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("GET", "/test")
		assert.NoError(t, err)
	})

	t.Run("empty response", func(t *testing.T) {
		client := NewTestClient(textResponse(http.StatusNoContent, ""), nil)

		gw := newArcadyan(client, "", "", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("GET", "/test")
		assert.NoError(t, err)
	})
}
