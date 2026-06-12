package internal

import (
	"testing"
	"time"

	tmhi "github.com/hugoh/tmhi-gateway/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUser = "u"
	testPass = "p"
	testIP   = "192.168.1.1"
)

func testGatewayCreation(t *testing.T, model string, expectedType any) {
	t.Helper()

	cfg := &Config{
		Model:    model,
		Username: testUser,
		Password: testPass,
		IP:       testIP,
		Timeout:  DefaultTimeout,
	}
	g, err := getGateway(cfg)
	require.NoError(t, err)
	assert.NotNil(t, g)
	assert.IsType(t, expectedType, g)
}

func TestGateway(t *testing.T) {
	t.Run("Nokia gateway creation", func(t *testing.T) {
		testGatewayCreation(t, NOK5G21, &tmhi.NokiaGateway{})
	})

	t.Run("Arcadyan gateway creation", func(t *testing.T) {
		testGatewayCreation(t, ARCADYAN, &tmhi.ArcadyanGateway{})
	})

	t.Run("Unknown gateway error", func(t *testing.T) {
		cfg := &Config{
			Model:    "invalid",
			Username: testUser,
			Password: testPass,
			IP:       testIP,
			Timeout:  DefaultTimeout,
		}
		g, err := getGateway(cfg)
		require.Error(t, err)
		assert.Nil(t, g)
	})

	t.Run("Missing credentials is not an error", func(t *testing.T) {
		cfg := &Config{
			Model:    NOK5G21,
			Username: "",
			Password: "",
			IP:       testIP,
			Timeout:  DefaultTimeout,
		}
		g, err := getGateway(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)
	})

	t.Run("Gateway can be created with various config options", func(t *testing.T) {
		cfg := &Config{
			Model:    NOK5G21,
			Username: testUser,
			Password: testPass,
			IP:       testIP,
			Timeout:  5 * time.Second,
			Retries:  3,
			Debug:    true,
			DryRun:   true,
		}
		g, err := getGateway(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.IsType(t, &tmhi.NokiaGateway{}, g)
	})
}
