package pkg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	resp *http.Response
	err  error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

func newTestClient(resp *http.Response, err error) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{resp: resp, err: err},
	}
}

func TestArcadyanGateway_Login_Success(t *testing.T) {
	// Prepare mock response
	body := `{"auth":{"expiration":1234567890,"refreshCountLeft":5,"refreshCountMax":10,"token":"testtoken"}}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	client := newTestClient(resp, nil)

	gw := &ArcadyanGateway{
		Username: "user",
		Password: "pass",
		IP:       "1.2.3.4",
		client:   client,
	}

	err := gw.Login()
	assert.NoError(t, err)
	assert.Equal(t, 1234567890, gw.credentials.Expiration)
	assert.Equal(t, "testtoken", gw.credentials.Token)
}

func TestArcadyanGateway_Login_Non200Status(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(bytes.NewBufferString("unauthorized")),
	}
	client := newTestClient(resp, nil)

	gw := &ArcadyanGateway{
		Username: "user",
		Password: "pass",
		IP:       "1.2.3.4",
		client:   client,
	}

	err := gw.Login()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 401")
}

func TestArcadyanGateway_Login_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("{invalid json")),
	}
	client := newTestClient(resp, nil)

	gw := &ArcadyanGateway{
		Username: "user",
		Password: "pass",
		IP:       "1.2.3.4",
		client:   client,
	}

	err := gw.Login()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode login response")
}

func TestArcadyanGateway_Login_HTTPClientError(t *testing.T) {
	client := newTestClient(nil, errors.New("network error"))

	gw := &ArcadyanGateway{
		Username: "user",
		Password: "pass",
		IP:       "1.2.3.4",
		client:   client,
	}

	err := gw.Login()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login request failed")
}
