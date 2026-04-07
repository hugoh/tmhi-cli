package internal

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// mockGateway implements pkg.Gateway for testing internal command handlers.
type mockGateway struct {
	loginCalled   bool
	loginErr      error
	infoCalled    bool
	infoErr       error
	statusCalled  bool
	statusErr     error
	rebootCalled  bool
	rebootDryRun  bool
	rebootErr     error
	requestCalled bool
	signalCalled  bool
	signalErr     error
}

func (m *mockGateway) NewClient(_ string, _ string, _ time.Duration, _ int, _ bool) {}
func (m *mockGateway) AddCredentials(_ string, _ string)                            {}
func (m *mockGateway) Login() error {
	m.loginCalled = true

	return m.loginErr
}

func (m *mockGateway) Reboot(dryRun bool) error {
	m.rebootCalled = true
	m.rebootDryRun = dryRun

	return m.rebootErr
}

func (m *mockGateway) Request(_ string, _ string) error {
	m.requestCalled = true

	return nil
}

func (m *mockGateway) Info() error {
	m.infoCalled = true

	return m.infoErr
}

func (m *mockGateway) Status() error {
	m.statusCalled = true

	return m.statusErr
}

func (m *mockGateway) Signal() error {
	m.signalCalled = true

	return m.signalErr
}

func ctxWithGateway(mg *mockGateway) context.Context {
	return context.WithValue(context.Background(), gatewayContextKey, mg)
}

func newRebootCmd(dry bool) *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  ConfigDryRun,
				Value: dry,
			},
		},
	}
}

func TestLogin_SuccessAndFailure(t *testing.T) {
	// Prevent logrus Fatal from exiting the test process
	restore, exitCode := WithPatchedLogrusExit(t)
	defer restore()

	// Success case
	{
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)
		err := Login(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.loginCalled)
		assert.Equal(t, 0, *exitCode, "should not exit on success")
	}

	// Failure case
	{
		mg := &mockGateway{loginErr: errors.New("login failed")}
		ctx := ctxWithGateway(mg)
		err := Login(ctx, nil)
		require.NoError(t, err, "Login handler always returns nil")
		assert.True(t, mg.loginCalled)
		assert.Equal(t, 1, *exitCode, "expected fatal exit code to be 1 on error")
	}
}

func TestInfo_SuccessAndFailure(t *testing.T) {
	// Success
	{
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)
		err := Info(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.infoCalled)
	}

	// Failure
	{
		mg := &mockGateway{infoErr: errors.New("info boom")}
		ctx := ctxWithGateway(mg)
		err := Info(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "info command failed")
		assert.True(t, mg.infoCalled)
	}
}

func TestStatus_SuccessAndFailure(t *testing.T) {
	// Success
	{
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)
		err := Status(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.statusCalled)
	}

	// Failure
	{
		mg := &mockGateway{statusErr: errors.New("status boom")}
		ctx := ctxWithGateway(mg)
		err := Status(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status check failed")
		assert.True(t, mg.statusCalled)
	}
}

func TestReboot_DryRunFlagAndFailure(t *testing.T) {
	// Dry-run true
	{
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)
		cmd := newRebootCmd(true)
		err := Reboot(ctx, cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled)
		assert.True(t, mg.rebootDryRun)
	}

	// Dry-run false with error
	{
		mg := &mockGateway{rebootErr: errors.New("reboot boom")}
		ctx := ctxWithGateway(mg)
		cmd := newRebootCmd(false)
		err := Reboot(ctx, cmd)
		require.Error(t, err)
		assert.True(t, mg.rebootCalled)
		assert.False(t, mg.rebootDryRun)
	}
}

func TestReq_LoginError(t *testing.T) {
	restore, _ := WithPatchedLogrusExit(t)
	defer restore()

	oldArgs := os.Args
	os.Args = []string{
		"tmhi-cli",
		"--gateway.model",
		"ARCADYAN",
		"--gateway.ip",
		"192.168.12.1",
		"req",
		"-l",
		"GET",
		"/test",
	}
	defer func() { os.Args = oldArgs }()

	Cmd("test-version")
}

func TestSignal_SuccessAndFailure(t *testing.T) {
	// Success
	{
		mg := &mockGateway{}
		ctx := ctxWithGateway(mg)
		err := Signal(ctx, nil)
		require.NoError(t, err)
		assert.True(t, mg.signalCalled)
	}

	// Failure
	{
		mg := &mockGateway{signalErr: errors.New("signal boom")}
		ctx := ctxWithGateway(mg)
		err := Signal(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signal command failed")
		assert.True(t, mg.signalCalled)
	}
}
