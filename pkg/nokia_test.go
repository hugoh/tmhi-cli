package pkg

import (
	"errors"
	"net/http"
	"testing"

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
	gw := NewNokiaGateway("user", "pass", "1.2.3.4")
	err := gw.Info()
	assert.ErrorIs(t, err, ErrNotImplemented)
}

func TestNokiaGateway_Request_NotImplemented(t *testing.T) {
	gw := NewNokiaGateway("user", "pass", "1.2.3.4")
	err := gw.Request("GET", "/test", false, false)
	assert.ErrorIs(t, err, ErrNotImplemented)
}

func TestNokiaGateway_Reboot_Success(t *testing.T) {
	client := newTestClient(&http.Response{StatusCode: http.StatusOK}, nil)
	gw := &NokiaGateway{
		client: client,
		credentials: nokiaLoginData{
			Success:   true,
			SID:       "valid-sid",
			CSRFToken: "valid-token",
		},
	}

	err := gw.Reboot(false)
	assert.NoError(t, err)
}

func TestNokiaGateway_Reboot_DryRun(t *testing.T) {
	client := newTestClient(nil, errors.New("should not be called"))
	gw := &NokiaGateway{
		client: client,
		credentials: nokiaLoginData{
			Success:   true,
			SID:       "valid-sid",
			CSRFToken: "valid-token",
		},
	}

	err := gw.Reboot(true)
	assert.NoError(t, err)
}
