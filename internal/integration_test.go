package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginIntegration_FullFlow(t *testing.T) {
	t.Run("successful login flow", func(t *testing.T) {
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)

		err := Login(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.loginCalled, "login should be called")
	})

	t.Run("failed login returns error", func(t *testing.T) {
		mg := &mockGateway{loginErr: errors.New("authentication failed")}
		ctx := ctxWithGateway(mg)

		err := Login(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
		assert.True(t, mg.loginCalled, "login should still be attempted")
	})
}

func TestInfoIntegration_FullFlow(t *testing.T) {
	t.Run("successful info retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)

		err := Info(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.infoCalled, "info should be called")
	})

	t.Run("info failure returns error", func(t *testing.T) {
		mg := &mockGateway{infoErr: errors.New("info unavailable")}
		ctx := ctxWithGateway(mg)

		err := Info(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "info command failed")
	})
}

func TestStatusIntegration_FullFlow(t *testing.T) {
	t.Run("successful status check", func(t *testing.T) {
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)

		err := Status(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.statusCalled, "status should be called")
	})

	t.Run("status failure returns error", func(t *testing.T) {
		mg := &mockGateway{statusErr: errors.New("status unavailable")}
		ctx := ctxWithGateway(mg)

		err := Status(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status check failed")
	})
}

func TestSignalIntegration_FullFlow(t *testing.T) {
	t.Run("successful signal retrieval", func(t *testing.T) {
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)

		err := Signal(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.signalCalled, "signal should be called")
	})

	t.Run("signal failure returns error", func(t *testing.T) {
		mg := &mockGateway{signalErr: errors.New("signal unavailable")}
		ctx := ctxWithGateway(mg)

		err := Signal(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signal command failed")
	})
}

func TestRebootIntegration_FullFlow(t *testing.T) {
	t.Run("successful reboot with auto-confirm", func(t *testing.T) {
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)
		cmd := newRebootCmd(false)

		err := Reboot(ctx, cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should be called")
		assert.False(t, mg.rebootDryRun, "dry-run should be false")
	})

	t.Run("successful reboot with dry-run", func(t *testing.T) {
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)
		cmd := newRebootCmd(true)

		err := Reboot(ctx, cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should be called")
		assert.True(t, mg.rebootDryRun, "dry-run should be true")
	})

	t.Run("reboot failure returns error", func(t *testing.T) {
		mg := &mockGateway{rebootErr: errors.New("reboot failed")}
		ctx := ctxWithGateway(mg)
		cmd := newRebootCmd(false)

		err := Reboot(ctx, cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not reboot gateway")
	})
}

func TestContextValueExtraction(t *testing.T) {
	t.Run("gateway can be extracted from context", func(t *testing.T) {
		mg := &mockGateway{}
		ctx := context.WithValue(context.Background(), gatewayContextKey, mg)

		extracted, ok := ctx.Value(gatewayContextKey).(*mockGateway)
		require.True(t, ok, "gateway should be extractable from context")
		assert.NotNil(t, extracted)
	})
}
