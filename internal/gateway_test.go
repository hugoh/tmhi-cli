package internal

import (
	"testing"
	"time"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGateway(t *testing.T) {
	const (
		testUser = "u"
		testPass = "p"
		testIP   = "192.168.1.1"
	)

	t.Run("Nokia gateway creation", func(t *testing.T) {
		g, err := getGateway("test-version", NOK5G21, testUser, testPass, testIP, 0, 0, false)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.IsType(t, &pkg.NokiaGateway{}, g)
	})

	t.Run("Arcadyan gateway creation", func(t *testing.T) {
		g, err := getGateway("test-version", ARCADYAN, testUser, testPass, testIP, 0, 0, false)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.IsType(t, &pkg.ArcadyanGateway{}, g)
	})

	t.Run("Unknown gateway error", func(t *testing.T) {
		g, err := getGateway("test-version", "invalid", testUser, testPass, testIP, 0, 0, false)
		require.Error(t, err)
		assert.Nil(t, g)
	})

	t.Run("Missing credentials is not an error", func(t *testing.T) {
		g, err := getGateway("test-version", NOK5G21, "", "", testIP, 0, 0, false)
		require.NoError(t, err)
		assert.NotNil(t, g)
	})

	t.Run("Client configuration", func(t *testing.T) {
		g, err := getGateway(
			"test-version",
			NOK5G21,
			testUser,
			testPass,
			testIP,
			5*time.Second,
			3,
			true,
		)
		require.NoError(t, err)

		nokia, ok := g.(*pkg.NokiaGateway)
		require.True(t, ok)
		assert.Equal(t, "http://192.168.1.1", nokia.Client.BaseURL)
		assert.Equal(t, 3, nokia.Client.RetryCount)
		assert.True(t, nokia.Client.Debug)
	})
}
