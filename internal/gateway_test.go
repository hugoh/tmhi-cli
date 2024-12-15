package internal

import (
	"testing"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/stretchr/testify/assert"
)

func TestGateway(t *testing.T) {
	var g pkg.GatewayI
	var err error
	const u = "u"
	g, err = getGateway(NOK5G21, u, u, u, false)
	assert.NoError(t, err)
	assert.NotNil(t, g)
	assert.IsType(t, &pkg.NokiaGateway{}, g)
	g, err = getGateway("foo", u, u, u, false)
	assert.Error(t, err)
	assert.Nil(t, g)
}
