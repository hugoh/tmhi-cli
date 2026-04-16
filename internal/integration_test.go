package internal

import (
	"context"
	"errors"
	"testing"

	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful login flow", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := login(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.loginCalled, "login should be called")
	})

	t.Run("failed login returns error", func(t *testing.T) {
		mg := &mockGateway{loginErr: errors.New("authentication failed")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := login(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Checking logging in...")
		assert.True(t, mg.loginCalled, "login should still be attempted")
	})
}

//nolint:dupl
func TestInfoIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful info retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := info(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.infoCalled, "info should be called")
	})

	t.Run("info failure returns error", func(t *testing.T) {
		mg := &mockGateway{infoErr: errors.New("info unavailable")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := info(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "info command failed")
	})
}

//nolint:dupl
func TestStatusIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful status check", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := status(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.statusCalled, "status should be called")
	})

	t.Run("status failure returns error", func(t *testing.T) {
		mg := &mockGateway{statusErr: errors.New("status unavailable")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := status(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Checking gateway status...")
	})
}

//nolint:dupl
func TestSignalIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful signal retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := signalCmd(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.signalCalled, "signal should be called")
	})

	t.Run("signal failure returns error", func(t *testing.T) {
		mg := &mockGateway{signalErr: errors.New("signal unavailable")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}

		err := signalCmd(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Fetching signal information...")
	})
}

func TestRebootIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc
	originalConfig := appConfig

	defer func() {
		initGatewayFunc = original
		appConfig = originalConfig
	}()

	t.Run("successful reboot with auto-confirm", func(t *testing.T) {
		appConfig = &Config{DryRun: false}
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(false)

		err := reboot(context.Background(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should be called")
	})

	t.Run("successful reboot with dry-run returns early", func(t *testing.T) {
		appConfig = &Config{DryRun: true}
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(true)

		err := reboot(context.Background(), cmd)
		require.NoError(t, err)
		assert.False(t, mg.rebootCalled, "reboot should not be called in dry-run mode")
	})

	t.Run("reboot failure returns error", func(t *testing.T) {
		appConfig = &Config{DryRun: false}
		mg := &mockGateway{rebootErr: errors.New("test failure")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(false)

		err := reboot(context.Background(), cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Reboot failed: test failure")
	})
}
