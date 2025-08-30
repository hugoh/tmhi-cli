package pkg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
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
		gw.Username = "user"
		gw.Password = "pass"

		_, err := gw.getCredentials(nonceResp{Nonce: "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication")
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
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication")
	})
}

func TestNokiaGateway_Reboot_Success(t *testing.T) {
	client := NewTestClient(&http.Response{StatusCode: http.StatusOK}, nil)
	gw := &NokiaGateway{
		GatewayCommon: &GatewayCommon{
			Client:        client,
			Username:      "user",
			Password:      "pass",
			Authenticated: true,
		},
		credentials: nokiaLoginData{
			Success:   true,
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
	assert.NoError(t, err)
	assert.Contains(t, out, "Web interface up")
}

func TestNokiaGateway_getNonce_ErrorResponse(t *testing.T) {
	client := NewTestClient(&http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewBufferString("server error")),
	}, nil)
	gw := NewNokiaGateway()
	gw.Client = client

	_, err := gw.getNonce()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication")
}

func TestNokiaGateway_Reboot_DryRun(t *testing.T) {
	client := NewTestClient(nil, errors.New("should not be called"))
	gw := &NokiaGateway{
		GatewayCommon: &GatewayCommon{
			Client:        client,
			Authenticated: true,
		},
		credentials: nokiaLoginData{
			Success:   true,
			SID:       "valid-sid",
			CSRFToken: "valid-token",
		},
	}

	err := gw.Reboot(true)
	assert.NoError(t, err)
}
