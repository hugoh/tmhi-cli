package internal

import (
	"testing"
	"time"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/stretchr/testify/assert"
)

func TestGateway(t *testing.T) {
	const testUser = "u"
	const testPass = "p"
	const testIP = "192.168.1.1"

	t.Run("Nokia gateway creation", func(t *testing.T) {
		g, err := getGateway("test-version", NOK5G21, testUser, testPass, testIP, 0, 0, false)
		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.IsType(t, &pkg.NokiaGateway{}, g)
	})

	t.Run("Arcadyan gateway creation", func(t *testing.T) {
		g, err := getGateway("test-version", ARCADYAN, testUser, testPass, testIP, 0, 0, false)
		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.IsType(t, &pkg.ArcadyanGateway{}, g)
	})

	t.Run("Unknown gateway error", func(t *testing.T) {
		g, err := getGateway("test-version", "invalid", testUser, testPass, testIP, 0, 0, false)
		assert.Error(t, err)
		assert.Nil(t, g)
	})

	t.Run("Missing credentials error", func(t *testing.T) {
		g, err := getGateway("test-version", NOK5G21, "", "", testIP, 0, 0, false)
		assert.Error(t, err)
		assert.Nil(t, g)
	})

	t.Run("Client configuration", func(t *testing.T) {
		g, err := getGateway("test-version", NOK5G21, testUser, testPass, testIP, 5*time.Second, 3, true)
		assert.NoError(t, err)
		assert.Equal(t, "http://192.168.1.1", g.(*pkg.NokiaGateway).Client.BaseURL)
		assert.Equal(t, 3, g.(*pkg.NokiaGateway).Client.RetryCount)
		assert.True(t, g.(*pkg.NokiaGateway).Client.Debug)
	})
}
