package pkg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUsername = "user"
	testPassword = "pass"
)

func Test_LoginSuccess(t *testing.T) {
	success := &nokiaLoginResp{
		Success:   0,
		Reason:    0,
		Sid:       "foo",
		CsrfToken: "bar",
	}
	assert.True(t, success.success())
}

func Test_LoginFailure(t *testing.T) {
	fail := &nokiaLoginResp{
		Success: 0,
		Reason:  600,
	}
	assert.False(t, fail.success())
}

func TestNokiaGateway_Info_NotImplemented(t *testing.T) {
	client := resty.New()
	gw := NewNokiaGateway()
	gw.Client = client
	err := gw.Info()
	assert.ErrorIs(t, err, ErrNotImplemented)
}

func TestNokiaGateway_Request_NotImplemented(t *testing.T) {
	client := resty.New()
	gw := NewNokiaGateway()
	gw.Client = client
	err := gw.Request("GET", "/test")
	assert.ErrorIs(t, err, ErrNotImplemented)
}

func TestNokiaGateway_getCredentials_ErrorResponse(t *testing.T) {
	t.Run("server error", func(t *testing.T) {
		client := NewTestClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("server error")),
		}, nil)

		gw := NewNokiaGateway()
		gw.Client = client
		gw.Username = testUsername
		gw.Password = testPassword

		_, err := gw.getCredentials(nonceResp{Nonce: "test"})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAuthentication)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		client := NewTestClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"success":0,"reason":600}`)),
		}, nil)

		gw := NewNokiaGateway()
		gw.Client = client
		gw.Username = "user"
		gw.Password = "pass"

		_, err := gw.getCredentials(nonceResp{Nonce: "test"})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAuthentication)
	})
}

func TestNokiaGateway_Reboot_Success(t *testing.T) {
	client := NewTestClient(&http.Response{StatusCode: http.StatusOK}, nil)
	gw := &NokiaGateway{
		GatewayCommon: &GatewayCommon{
			Client:        client,
			Username:      testUsername,
			Password:      testPassword,
			Authenticated: true,
		},
		credentials: nokiaLoginData{
			SID:       "valid-sid",
			CSRFToken: "valid-token",
		},
	}

	err := gw.Reboot(false)
	assert.NoError(t, err)
}

func TestNokiaGateway_Status(t *testing.T) {
	client := NewTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}, nil)
	gw := NewNokiaGateway()
	gw.Client = client

	var err error
	out := CaptureStdout(t, func() {
		err = gw.Status()
	})
	require.NoError(t, err)
	assert.Contains(t, out, "Web interface up")
}

func TestNokiaGateway_getNonce_ErrorResponse(t *testing.T) {
	client := NewTestClient(nil, errors.New("network error"))
	gw := NewNokiaGateway()
	gw.Client = client

	_, err := gw.getNonce()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting nonce")
}

func TestNokiaGateway_getNonce_Success(t *testing.T) {
	body := `{"nonce":"testNonce","pubkey":"testPubkey","randomKey":"testRandomKey"}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	client := NewTestClient(resp, nil)

	gw := NewNokiaGateway()
	gw.Client = client

	nonceResp, err := gw.getNonce()
	require.NoError(t, err)
	assert.Equal(t, "testNonce", nonceResp.Nonce)
	assert.Equal(t, "testPubkey", nonceResp.Pubkey)
	assert.Equal(t, "testRandomKey", nonceResp.RandomKey)
}

func TestNokiaGateway_getCredentials_Success(t *testing.T) {
	body := `{"success":0,"reason":0,"sid":"testSid","token":"testToken"}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	client := NewTestClient(resp, nil)

	gw := NewNokiaGateway()
	gw.Client = client
	gw.Username = testUsername
	gw.Password = testPassword

	loginResp, err := gw.getCredentials(nonceResp{Nonce: "testNonce", RandomKey: "testRandomKey"})
	require.NoError(t, err)
	assert.Equal(t, "testSid", loginResp.Sid)
	assert.Equal(t, "testToken", loginResp.CsrfToken)
}

func TestNokiaGateway_Login_AlreadyAuthenticated(t *testing.T) {
	gw := NewNokiaGateway()
	gw.Authenticated = true

	err := gw.Login()
	assert.NoError(t, err)
}

func TestNokiaGateway_Login_NonceError(t *testing.T) {
	client := NewTestClient(nil, errors.New("network error"))

	gw := NewNokiaGateway()
	gw.Client = client

	err := gw.Login()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting nonce")
}

func TestNokiaGateway_Login_CredentialsError(t *testing.T) {
	nonceBody := `{"nonce":"testNonce","pubkey":"testPubkey","randomKey":"testRandomKey"}`
	client := NewTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(nonceBody)),
	}, nil)

	gw := NewNokiaGateway()
	gw.Client = client
	gw.Username = testUsername
	gw.Password = testPassword

	err := gw.Login()
	assert.Error(t, err)
}

func TestNokiaGateway_Reboot_DryRun(t *testing.T) {
	client := NewTestClient(nil, errors.New("should not be called"))
	gw := &NokiaGateway{
		GatewayCommon: &GatewayCommon{
			Client:        client,
			Authenticated: true,
		},
		credentials: nokiaLoginData{
			SID:       "valid-sid",
			CSRFToken: "valid-token",
		},
	}

	err := gw.Reboot(true)
	assert.NoError(t, err)
}
