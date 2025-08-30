package pkg

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
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

func NewTestClient(resp *http.Response, err error) *resty.Client {
	return resty.NewWithClient(&http.Client{
		Transport: &mockRoundTripper{resp: resp, err: err},
	})
}

func TestNewGatewayCommon(t *testing.T) {
	gc := NewGatewayCommon()
	assert.NotNil(t, gc.Client)
	assert.Equal(t, 0, gc.Client.RetryCount)
	assert.False(t, gc.Authenticated)
	assert.Empty(t, gc.Username)
	assert.Empty(t, gc.Password)

	// Test client configuration
	version := "test-version"
	ip := "192.168.1.1"
	timeout := 5 * time.Second
	retries := 3
	gc.NewClient(version, ip, timeout, retries, true)

	assert.Equal(t, "http://192.168.1.1", gc.Client.BaseURL)
	assert.Equal(t, "tmhi-cli/test-version", gc.Client.Header.Get("User-Agent"))
	assert.Equal(t, timeout, gc.Client.GetClient().Timeout)
	assert.Equal(t, retries, gc.Client.RetryCount)
	assert.True(t, gc.Client.Debug)
}

func TestAuthenticationError(t *testing.T) {
	err := AuthenticationError("test error")
	assert.ErrorContains(t, err, "could not authenticate: test error")
}

func TestAddCredentials(t *testing.T) {
	gc := NewGatewayCommon()
	gc.AddCredentials("admin", "password")
	assert.Equal(t, "admin", gc.Username)
	assert.Equal(t, "password", gc.Password)
}

func TestGatewayCommon_StatusCore(t *testing.T) {
	cases := []struct {
		name      string
		resp      *http.Response
		err       error
		wantEmoji string
	}{
		{
			name:      "successful web interface check",
			resp:      &http.Response{StatusCode: http.StatusOK, Body: http.NoBody},
			err:       nil,
			wantEmoji: "✅",
		},
		{
			name:      "failed web interface status code",
			resp:      &http.Response{StatusCode: http.StatusInternalServerError, Body: http.NoBody},
			err:       nil,
			wantEmoji: "❌",
		},
		{
			name:      "not found web interface",
			resp:      &http.Response{StatusCode: http.StatusNotFound, Body: http.NoBody},
			err:       nil,
			wantEmoji: "❌",
		},
		{
			name:      "failed web interface check",
			resp:      nil,
			err:       errors.New("connection refused"),
			wantEmoji: "❌",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewTestClient(tc.resp, tc.err)
			gc := NewGatewayCommon()
			gc.Client = client

			out := CaptureStdout(t, func() {
				gc.StatusCore()
			})
			assert.Contains(t, out, tc.wantEmoji)
		})
	}
}
