package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestLoginIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful login flow", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Login(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.loginCalled, "login should be called")
	})

	t.Run("failed login returns error", func(t *testing.T) {
		mg := &mockGateway{loginErr: errors.New("authentication failed")}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Login(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
		assert.True(t, mg.loginCalled, "login should still be attempted")
	})
}

//nolint:dupl
func TestInfoIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful info retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Info(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.infoCalled, "info should be called")
	})

	t.Run("info failure returns error", func(t *testing.T) {
		mg := &mockGateway{infoErr: errors.New("info unavailable")}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Info(context.Background(), nil)
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
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Status(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.statusCalled, "status should be called")
	})

	t.Run("status failure returns error", func(t *testing.T) {
		mg := &mockGateway{statusErr: errors.New("status unavailable")}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Status(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status check failed")
	})
}

//nolint:dupl
func TestSignalIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful signal retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Signal(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.signalCalled, "signal should be called")
	})

	t.Run("signal failure returns error", func(t *testing.T) {
		mg := &mockGateway{signalErr: errors.New("signal unavailable")}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}

		err := Signal(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signal command failed")
	})
}

func TestRebootIntegration_FullFlow(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("successful reboot with auto-confirm", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(false)

		err := Reboot(context.Background(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should be called")
		assert.False(t, mg.rebootDryRun, "dry-run should be false")
	})

	t.Run("successful reboot with dry-run", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(true)

		err := Reboot(context.Background(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should be called")
		assert.True(t, mg.rebootDryRun, "dry-run should be true")
	})

	t.Run("reboot failure returns error", func(t *testing.T) {
		mg := &mockGateway{rebootErr: errors.New("reboot failed")}
		initGatewayFunc = func(_ *cli.Command) (pkg.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(false)

		err := Reboot(context.Background(), cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not reboot gateway")
	})
}
