package internal

import (
	"reflect"
	"testing"
	"time"

	tmhi "github.com/hugoh/tmhi-gateway"
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
		cfg := &Config{Model: NOK5G21, Username: testUser, Password: testPass, IP: testIP}
		g, err := getGateway(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.IsType(t, &tmhi.NokiaGateway{}, g)
	})

	t.Run("Arcadyan gateway creation", func(t *testing.T) {
		cfg := &Config{Model: ARCADYAN, Username: testUser, Password: testPass, IP: testIP}
		g, err := getGateway(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.IsType(t, &tmhi.ArcadyanGateway{}, g)
	})

	t.Run("Unknown gateway error", func(t *testing.T) {
		cfg := &Config{Model: "invalid", Username: testUser, Password: testPass, IP: testIP}
		g, err := getGateway(cfg)
		require.Error(t, err)
		assert.Nil(t, g)
	})

	t.Run("Missing credentials is not an error", func(t *testing.T) {
		cfg := &Config{Model: NOK5G21, Username: "", Password: "", IP: testIP}
		g, err := getGateway(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)
	})

	t.Run("Client configuration", func(t *testing.T) {
		cfg := &Config{
			Model:    NOK5G21,
			Username: testUser,
			Password: testPass,
			IP:       testIP,
			Timeout:  5 * time.Second,
			Retries:  3,
			Debug:    true,
		}
		g, err := getGateway(cfg)
		require.NoError(t, err)

		nokia, ok := g.(*tmhi.NokiaGateway)
		require.True(t, ok)
		assert.Equal(t, "http://192.168.1.1", nokia.Client.BaseURL)
		assert.Equal(t, 3, nokia.Client.RetryCount)
		assert.True(t, nokia.Client.Debug)
	})

	t.Run("DryRun is passed to GatewayConfig", func(t *testing.T) {
		cfg := &Config{
			Model:    NOK5G21,
			Username: testUser,
			Password: testPass,
			IP:       testIP,
			Timeout:  5 * time.Second,
			DryRun:   true,
		}
		g, err := getGateway(cfg)
		require.NoError(t, err)

		// Use reflection to access the unexported config field
		gwValue := reflect.ValueOf(g).Elem()
		configField := gwValue.FieldByName("config")
		require.True(t, configField.IsValid(), "config field should exist")

		// config is a pointer to GatewayConfig
		if configField.Kind() == reflect.Ptr && !configField.IsNil() {
			gwConfig := configField.Elem()
			dryRunField := gwConfig.FieldByName("DryRun")
			require.True(t, dryRunField.IsValid(), "DryRun field should exist in GatewayConfig")
			assert.True(t, dryRunField.Bool(), "DryRun should be true in GatewayConfig")
		} else {
			// Check GatewayCommon's config field (for Arcadyan)
			commonField := gwValue.FieldByName("GatewayCommon")
			if commonField.IsValid() {
				commonConfig := commonField.Elem().FieldByName("config")
				if commonConfig.IsValid() && commonConfig.Kind() == reflect.Ptr &&
					!commonConfig.IsNil() {
					gwConfig := commonConfig.Elem()
					dryRunField := gwConfig.FieldByName("DryRun")
					require.True(
						t,
						dryRunField.IsValid(),
						"DryRun field should exist in GatewayConfig",
					)
					assert.True(t, dryRunField.Bool(), "DryRun should be true in GatewayConfig")
				}
			}
		}
	})
}
