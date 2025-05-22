package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

// Mocking a reader that always returns an error
type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError // Using assert.AnError for a generic error
}

func Test_GetBody(t *testing.T) {
	t.Run("valid body", func(t *testing.T) {
		bodyStr := "Hello, world!"
		resp := &http.Response{
			Body: io.NopCloser(strings.NewReader(bodyStr)),
		}
		assert.Equal(t, bodyStr, GetBody(resp))
	})

	t.Run("empty body", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(strings.NewReader("")),
		}
		assert.Equal(t, "", GetBody(resp))
	})

	t.Run("error reading body", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(&errorReader{}),
		}
		// The current GetBody implementation appends the error message to the (empty) details string.
		// assert.AnError.Error() will give the string representation of the generic error.
		expectedBodyWithError := "\n" + assert.AnError.Error()
		assert.Equal(t, expectedBodyWithError, GetBody(resp))
	})
}

func Test_HTTPRequestSuccessful(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{"OK", 200, true},
		{"Created", 201, true},
		{"No Content", 204, true},
		{"Multiple Choices", 300, false},
		{"Bad Request", 400, false},
		{"Not Found", 404, false},
		{"Internal Server Error", 500, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.statusCode,
			}
			assert.Equal(t, tc.expected, HTTPRequestSuccessful(resp))
		})
	}
}

func Test_LogHTTPResponseFields(t *testing.T) {
	bodyStr := "Response body content"
	statusCode := 200
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(bodyStr)),
	}

	fields := LogHTTPResponseFields(resp)

	assert.Equal(t, statusCode, fields["status"])
	assert.Equal(t, bodyStr, fields["body"])

	// Test with error reading body
	respError := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(&errorReader{}),
	}
	fieldsError := LogHTTPResponseFields(respError)
	assert.Equal(t, 500, fieldsError["status"])
	// When body read fails, the body field in logs should be empty,
	// and an error should be logged by LogHTTPResponseFields itself (not tested here directly).
	assert.Equal(t, "", fieldsError["body"])
}

func Test_Sha256Hash(t *testing.T) {
	// Test vectors can be generated using an online tool or command line:
	// e.g., echo -n "admin:testpassword" | sha256sum | xxd -r -p | base64
	testCases := []struct {
		name     string
		val1     string
		val2     string
		expected string
	}{
		{"standard case", "admin", "testpassword", "uSSTqP18RXMHCmH+j2AQ9nL8GNK9HuKI+bgM4WvxGk8="},
		{"empty val1", "", "testpassword", "j42yZyt2T4NlAdbFfB/jOuGUL7kC+7+3Y2mJbMR0TE4="},
		{"empty val2", "admin", "", "bhVfPMyD+8VPG1N2uz4H0zQO0p4xL0N7jSQuCFmc+Zc="},
		{"both empty", "", "", "frcCV1k9oG9oKj3dpUqdJg1VOZZxDTRL8X6EPwAEvx8="},
		{"longer strings", "averylongusernameexample", "averysecureandlongpasswordexample", "dG+j3qrfz0NWHbPNQ8L4qgqiBJSy3Wp7uN7OKlFfG2o="},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := Sha256Hash(tc.val1, tc.val2)
			assert.Equal(t, tc.expected, hash)
		})
	}
}
