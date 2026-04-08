package pkg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hugoh/tmhi-cli/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func newArcadyan(
	client *resty.Client,
	username, password, token string,
	exp time.Time,
) *ArcadyanGateway {
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
	client := NewTestClient(jsonResponse(http.StatusOK, body), nil) //nolint:bodyclose // test mock

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	require.NoError(t, err)
	assert.Equal(t, 1234567890, gw.credentials.Expiration)
	assert.Equal(t, "testtoken", gw.credentials.Token)
}

func TestArcadyanGateway_Reboot_Failure(t *testing.T) {
	//nolint:bodyclose // test mock
	client := NewTestClient(textResponse(http.StatusInternalServerError, "server error"), nil)

	gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

	err := gw.Reboot(false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reboot failed")
}

func TestArcadyanGateway_Login_Non200Status(t *testing.T) {
	//nolint:bodyclose // test mock
	client := NewTestClient(textResponse(http.StatusUnauthorized, "unauthorized"), nil)

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}

func TestArcadyanGateway_Login_InvalidJSON(t *testing.T) {
	//nolint:bodyclose // test mock
	client := NewTestClient(jsonResponse(http.StatusOK, "{invalid json"), nil)

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode login response")
}

func TestArcadyanGateway_Login_HTTPClientError(t *testing.T) {
	client := NewTestClient(nil, errors.New("network error"))

	gw := newArcadyan(client, "user", "pass", "", time.Time{})

	err := gw.Login()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login request failed")
}

func TestArcadyanGateway_Info_Success(t *testing.T) {
	//nolint:bodyclose // test mock
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

func TestNewArcadyanGateway(t *testing.T) {
	gw := NewArcadyanGateway()
	assert.NotNil(t, gw)
	assert.NotNil(t, gw.Client)
	assert.Equal(t, "application/json", gw.Client.Header.Get("Accept"))
}

func TestArcadyanGateway_Status(t *testing.T) {
	t.Run("successful status with registration info", func(t *testing.T) {
		headResp := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
		infoBody := `{"signal":{"generic":{"registration":"registered"}}}`
		//nolint:bodyclose // test mock
		infoResp := jsonResponse(http.StatusOK, infoBody)
		client := NewMultiTestClient([]*http.Response{headResp, infoResp}, []error{nil, nil})

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		var err error

		out := testutil.CaptureStdout(t, func() {
			err = gw.Status()
		})
		require.NoError(t, err)
		assert.Contains(t, out, "Web interface")
		assert.Contains(t, out, "up")
		assert.Contains(t, out, "Registration status")
		assert.Contains(t, out, "registered")
	})

	t.Run("status with network error returns unknown", func(t *testing.T) {
		headResp := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
		client := NewMultiTestClient([]*http.Response{headResp}, []error{nil})

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Status()
		assert.Error(t, err)
	})
}

func TestArcadyanGateway_Request_Methods(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		//nolint:bodyclose // test mock
		client := NewTestClient(jsonResponse(http.StatusOK, `{"status": "ok"}`), nil)

		gw := newArcadyan(client, "", "", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("GET", "/test")
		assert.NoError(t, err)
	})

	t.Run("POST request", func(t *testing.T) {
		//nolint:bodyclose // test mock
		client := NewTestClient(jsonResponse(http.StatusOK, `{"status": "created"}`), nil)

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("POST", "/test")
		assert.NoError(t, err)
	})

	t.Run("non-JSON response", func(t *testing.T) {
		//nolint:bodyclose // test mock
		client := NewTestClient(textResponse(http.StatusOK, "plain text response"), nil)

		gw := newArcadyan(client, "", "", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("GET", "/test")
		assert.NoError(t, err)
	})

	t.Run("empty response", func(t *testing.T) {
		//nolint:bodyclose // test mock
		client := NewTestClient(textResponse(http.StatusNoContent, ""), nil)

		gw := newArcadyan(client, "", "", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Request("GET", "/test")
		assert.NoError(t, err)
	})
}

//nolint:funlen // test function with multiple sub-tests
func TestArcadyanGateway_Signal(t *testing.T) {
	t.Run("successful signal retrieval with 4g and 5g", func(t *testing.T) {
		body := `{
	"signal": {
		"4g": {
			"bands": ["b2"],
			"bars": 4.0,
			"cid": 12,
			"eNBID": 310463,
			"rsrp": -95,
			"rsrq": -8,
			"rssi": -85,
			"sinr": 15
		},
		"5g": {
			"antennaUsed": "Internal_directional",
			"bands": ["n41"],
			"bars": 5.0,
			"cid": 311,
			"gNBID": 1076984,
			"rsrp": -84,
			"rsrq": -10,
			"rssi": -72,
			"sinr": 28
		},
		"generic": {
			"apn": "FBB.HOME",
			"hasIPv6": true,
			"registration": "registered",
			"roaming": false
		}
	}
}`
		//nolint:bodyclose // test mock
		client := NewTestClient(jsonResponse(http.StatusOK, body), nil)

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		var err error

		out := testutil.CaptureStdout(t, func() {
			err = gw.Signal()
		})
		require.NoError(t, err)
		assert.Contains(t, out, "4G LTE Signal")
		assert.Contains(t, out, "4")
		assert.Contains(t, out, "[b2]")
		assert.Contains(t, out, "-95")
		assert.Contains(t, out, "-8")
		assert.Contains(t, out, "-85")
		assert.Contains(t, out, "15")
		assert.Contains(t, out, "310463")
		assert.Contains(t, out, "5G Signal")
		assert.Contains(t, out, "5")
		assert.Contains(t, out, "Internal_directional")
		assert.Contains(t, out, "[n41]")
		assert.Contains(t, out, "-84")
		assert.Contains(t, out, "-10")
		assert.Contains(t, out, "-72")
		assert.Contains(t, out, "28")
		assert.Contains(t, out, "1076984")
		assert.Contains(t, out, "Generic Info")
		assert.Contains(t, out, "FBB.HOME")
		assert.Contains(t, out, "registered")
	})

	t.Run("successful signal retrieval 5g only", func(t *testing.T) {
		body := `{
	"signal": {
		"5g": {
			"antennaUsed": "",
			"bands": ["n41"],
			"bars": 5.0,
			"cid": 311,
			"gNBID": 1076984,
			"rsrp": -84,
			"rsrq": -10,
			"rssi": -72,
			"sinr": 28
		},
		"generic": {
			"apn": "FBB.HOME",
			"hasIPv6": true,
			"registration": "registered",
			"roaming": false
		}
	}
}`
		//nolint:bodyclose // test mock
		client := NewTestClient(jsonResponse(http.StatusOK, body), nil)

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		var err error

		out := testutil.CaptureStdout(t, func() {
			err = gw.Signal()
		})
		require.NoError(t, err)
		assert.NotContains(t, out, "4G LTE Signal")
		assert.Contains(t, out, "5G Signal")
		assert.Contains(t, out, "5")
		assert.Contains(t, out, "[n41]")
		assert.NotContains(t, out, "Antenna")
	})

	t.Run("successful signal retrieval 4g only", func(t *testing.T) {
		body := `{
	"signal": {
		"4g": {
			"bands": ["b2"],
			"bars": 4.0,
			"cid": 12,
			"eNBID": 310463,
			"rsrp": -95,
			"rsrq": -8,
			"rssi": -85,
			"sinr": 15
		},
		"generic": {
			"apn": "FBB.HOME",
			"hasIPv6": true,
			"registration": "registered",
			"roaming": false
		}
	}
}`
		//nolint:bodyclose // test mock
		client := NewTestClient(jsonResponse(http.StatusOK, body), nil)

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		var err error

		out := testutil.CaptureStdout(t, func() {
			err = gw.Signal()
		})
		require.NoError(t, err)
		assert.Contains(t, out, "4G LTE Signal")
		assert.Contains(t, out, "4")
		assert.Contains(t, out, "310463")
		assert.NotContains(t, out, "5G Signal")
	})

	t.Run("signal with network error", func(t *testing.T) {
		client := NewTestClient(nil, errors.New("network error"))

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Signal()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get signal info")
	})

	t.Run("signal with non-200 status", func(t *testing.T) {
		//nolint:bodyclose // test mock
		client := NewTestClient(jsonResponse(http.StatusInternalServerError, "{}"), nil)

		gw := newArcadyan(client, "user", "pass", "valid-token", time.Now().Add(1*time.Hour))

		err := gw.Signal()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signal failed")
	})
}
