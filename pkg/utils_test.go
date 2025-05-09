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
