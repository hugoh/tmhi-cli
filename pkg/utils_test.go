package pkg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func Test_GetBody(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		resp := &http.Response{Body: io.NopCloser(bytes.NewBufferString("test body"))}
		result := GetBody(resp)
		assert.Equal(t, "test body", result)
	})

	t.Run("with read error", func(t *testing.T) {
		resp := &http.Response{Body: io.NopCloser(errorReader{})}
		result := GetBody(resp)
		assert.Contains(t, result, "test error")
	})

	t.Run("successful read with error in body", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(errorReader{}),
		}
		result := GetBody(resp)
		assert.Contains(t, result, "test error")
	})
}

func Test_EchoStatus(t *testing.T) {
	out := CaptureStdout(t, func() {
		testMessage := "test status"
		EchoStatus(testMessage, true)
		EchoStatus(testMessage, false)
	})
	assert.Contains(t, out, "✅")
	assert.Contains(t, out, "❌")
}

func Test_HTTPRequestSuccessful(t *testing.T) {
	assert.True(t, HTTPRequestSuccessful(&http.Response{StatusCode: 200}))
	assert.True(t, HTTPRequestSuccessful(&http.Response{StatusCode: 299}))
	assert.False(t, HTTPRequestSuccessful(&http.Response{StatusCode: 400}))
}

func Test_LogHTTPResponseFields(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("test body")),
		}
		fields := LogHTTPResponseFields(resp)
		assert.Equal(t, 200, fields["status"])
		assert.Equal(t, "test body", fields["body"])
	})

	t.Run("read error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(errorReader{}),
		}
		fields := LogHTTPResponseFields(resp)
		assert.Equal(t, 200, fields["status"])
		assert.Contains(t, fields["body"], "error reading HTTP body")
	})

	t.Run("empty body", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 204,
			Body:       io.NopCloser(bytes.NewBufferString("")),
		}
		fields := LogHTTPResponseFields(resp)
		assert.Equal(t, 204, fields["status"])
		assert.Equal(t, "", fields["body"])
	})
}

func Test_Sha256Hash(t *testing.T) {
	result := Sha256Hash("admin", "password")
	assert.Equal(t, "ux+w+s92nXMGACVBFqXMzkpsDxdWeI/aFC8GPNGAKqM=", result)
}

func Test_Base64urlEscape(t *testing.T) {
	out := Base64urlEscape("efbgOrynhgggULfrXxDu9FveT+q2fXegZs6rXIbiky4=")
	assert.Equal(t, "efbgOrynhgggULfrXxDu9FveT-q2fXegZs6rXIbiky4.", out)
}

func Test_Sha256Url(t *testing.T) {
	out := Sha256Url("admin", "efbgOrynhgggULfrXxDu9FveT+q2fXegZs6rXIbiky4=")
	assert.Equal(t, "xrNe9hWWlAiL14wfvJxcXOBmMKLBOPIXX1nESQpvaOk.", out)
}

func Test_Random16bytes(t *testing.T) {
	out1 := Random16bytes()
	assert.NotEqual(t, "", out1)
	out2 := Random16bytes()
	assert.NotEqual(t, "", out2)
	assert.NotEqual(t, out1, out2)
}

func Test_EchoOut(t *testing.T) {
	testString := "test echo output"
	out := CaptureStdout(t, func() {
		EchoOut(testString)
	})
	assert.Equal(t, testString+"\n", out)
}
