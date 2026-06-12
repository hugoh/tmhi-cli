package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginIntegration_FullFlow(t *testing.T) {
	t.Run("successful login flow", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		err := a.login(t.Context(), nil)
		require.NoError(t, err)
		assert.True(t, mg.loginCalled, "login should be called")
	})

	t.Run("failed login returns error", func(t *testing.T) {
		mg := &mockGateway{loginErr: errors.New("authentication failed")}
		a := newTestApp(mg)

		err := a.login(t.Context(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Logging in...")
		assert.True(t, mg.loginCalled, "login should still be attempted")
	})
}

func TestInfoIntegration_FullFlow(t *testing.T) {
	t.Run("successful info retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		err := a.info(t.Context(), nil)
		require.NoError(t, err)
		assert.True(t, mg.infoCalled, "info should be called")
	})

	t.Run("info failure returns error", func(t *testing.T) {
		mg := &mockGateway{infoErr: errors.New("info unavailable")}
		a := newTestApp(mg)

		err := a.info(t.Context(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Fetching gateway info")
		assert.Contains(t, err.Error(), "info unavailable")
	})
}

func TestStatusIntegration_FullFlow(t *testing.T) {
	t.Run("successful status check", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		err := a.status(t.Context(), nil)
		require.NoError(t, err)
		assert.True(t, mg.statusCalled, "status should be called")
	})

	t.Run("status failure returns error", func(t *testing.T) {
		mg := &mockGateway{statusErr: errors.New("status unavailable")}
		a := newTestApp(mg)

		err := a.status(t.Context(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Checking gateway status...")
	})
}

func TestSignalIntegration_FullFlow(t *testing.T) {
	t.Run("successful signal retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		err := a.signal(t.Context(), nil)
		require.NoError(t, err)
		assert.True(t, mg.signalCalled, "signal should be called")
	})

	t.Run("signal failure returns error", func(t *testing.T) {
		mg := &mockGateway{signalErr: errors.New("signal unavailable")}
		a := newTestApp(mg)

		err := a.signal(t.Context(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Fetching signal information...")
	})
}

func TestRebootIntegration_FullFlow(t *testing.T) {
	t.Run("successful reboot with auto-confirm", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)
		a.config = &Config{DryRun: false}
		cmd := newRebootCmd(false)

		err := a.reboot(t.Context(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should be called")
	})

	t.Run("successful reboot with dry-run returns early", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)
		a.config = &Config{DryRun: true}
		cmd := newRebootCmd(true)

		err := a.reboot(t.Context(), cmd)
		require.NoError(t, err)
		assert.False(t, mg.rebootCalled, "reboot should not be called in dry-run mode")
	})

	t.Run("reboot failure returns error", func(t *testing.T) {
		mg := &mockGateway{rebootErr: errors.New("test failure")}
		a := newTestApp(mg)
		a.config = &Config{DryRun: false}
		cmd := newRebootCmd(false)

		err := a.reboot(t.Context(), cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Rebooting gateway...: test failure")
	})
}
