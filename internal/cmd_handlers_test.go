package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/hugoh/tmhi-cli/pkg"
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

func (m *mockGateway) NewClient(_ *pkg.GatewayConfig)    {}
func (m *mockGateway) AddCredentials(_ string, _ string) {}
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

func newRebootCmd(dry bool) *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  ConfigDryRun,
				Value: dry,
			},
			&cli.BoolFlag{
				Name:  ConfigAutoConfirm,
				Value: true,
			},
		},
	}
}

func TestLogin_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	// Success case
	{
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Login(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.loginCalled)
	}

	// Failure case
	{
		mg := &mockGateway{loginErr: errors.New("login failed")}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Login(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
		assert.True(t, mg.loginCalled)
	}
}

func TestInfo_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	// Success
	{
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Info(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.infoCalled)
	}

	// Failure
	{
		mg := &mockGateway{infoErr: errors.New("info boom")}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Info(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "info command failed")
		assert.True(t, mg.infoCalled)
	}
}

func TestStatus_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	// Success
	{
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Status(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.statusCalled)
	}

	// Failure
	{
		mg := &mockGateway{statusErr: errors.New("status boom")}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Status(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status check failed")
		assert.True(t, mg.statusCalled)
	}
}

func TestReboot_DryRunFlagAndFailure(t *testing.T) {
	original := initGatewayFunc
	originalConfig := appConfig

	defer func() {
		initGatewayFunc = original
		appConfig = originalConfig
	}()

	// Dry-run true
	{
		appConfig = &Config{DryRun: true}
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(true)
		err := Reboot(context.Background(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled)
		assert.True(t, mg.rebootDryRun)
	}

	// Dry-run false with error
	{
		appConfig = &Config{DryRun: false}
		mg := &mockGateway{rebootErr: errors.New("reboot boom")}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		cmd := newRebootCmd(false)
		err := Reboot(context.Background(), cmd)
		require.Error(t, err)
		assert.True(t, mg.rebootCalled)
		assert.False(t, mg.rebootDryRun)
	}
}

func TestReboot_ConfirmationDefaultsToNo(t *testing.T) {
	original := initGatewayFunc
	originalConfig := appConfig

	defer func() {
		initGatewayFunc = original
		appConfig = originalConfig
	}()

	t.Run("enter_accepts_default_no", func(t *testing.T) {
		testRebootConfirmCancel(t, keys.Enter)
	})

	t.Run("y_confirms_reboot", func(t *testing.T) {
		testRebootConfirmProceed(t, 'y')
	})

	t.Run("n_cancels_reboot", func(t *testing.T) {
		testRebootConfirmCancel(t, 'n')
	})

	t.Run("auto_confirm_skips_prompt", func(t *testing.T) {
		appConfig = &Config{DryRun: false}
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		cmd := &cli.Command{
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: ConfigDryRun, Value: false},
				&cli.BoolFlag{Name: ConfigAutoConfirm, Value: true},
			},
		}

		err := Reboot(context.Background(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should proceed with auto-confirm")
	})
}

//nolint:dupl
func testRebootConfirmCancel(t *testing.T, key any) {
	t.Helper()

	original := initGatewayFunc
	originalConfig := appConfig

	defer func() {
		initGatewayFunc = original
		appConfig = originalConfig
	}()

	appConfig = &Config{DryRun: false}
	mg := &mockGateway{}
	initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
		return mg, nil
	}
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: ConfigDryRun, Value: false},
			&cli.BoolFlag{Name: ConfigAutoConfirm, Value: false},
		},
	}

	go func() {
		time.Sleep(50 * time.Millisecond)

		_ = keyboard.SimulateKeyPress(key)
	}()

	err := Reboot(context.Background(), cmd)
	require.NoError(t, err)
	assert.False(t, mg.rebootCalled, "reboot should be cancelled")
}

//nolint:dupl
func testRebootConfirmProceed(t *testing.T, key any) {
	t.Helper()

	original := initGatewayFunc
	originalConfig := appConfig

	defer func() {
		initGatewayFunc = original
		appConfig = originalConfig
	}()

	appConfig = &Config{DryRun: false}
	mg := &mockGateway{}
	initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
		return mg, nil
	}
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: ConfigDryRun, Value: false},
			&cli.BoolFlag{Name: ConfigAutoConfirm, Value: false},
		},
	}

	go func() {
		time.Sleep(50 * time.Millisecond)

		_ = keyboard.SimulateKeyPress(key)
	}()

	err := Reboot(context.Background(), cmd)
	require.NoError(t, err)
	assert.True(t, mg.rebootCalled, "reboot should proceed")
}

func TestSignal_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	// Success
	{
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Signal(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.signalCalled)
	}

	// Failure
	{
		mg := &mockGateway{signalErr: errors.New("signal boom")}
		initGatewayFunc = func(_ *Config) (pkg.Gateway, error) {
			return mg, nil
		}
		err := Signal(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signal command failed")
		assert.True(t, mg.signalCalled)
	}
}
