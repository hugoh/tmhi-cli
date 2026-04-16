package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

type mockGateway struct {
	loginCalled   bool
	loginErr      error
	infoCalled    bool
	infoErr       error
	statusCalled  bool
	statusErr     error
	rebootCalled  bool
	rebootErr     error
	requestCalled bool
	signalCalled  bool
	signalErr     error
}

func (m *mockGateway) Login() (*tmhi.LoginResult, error) {
	m.loginCalled = true
	if m.loginErr != nil {
		return nil, m.loginErr
	}

	return &tmhi.LoginResult{Success: true}, nil
}

func (m *mockGateway) Reboot() error {
	m.rebootCalled = true

	return m.rebootErr
}

func (m *mockGateway) Request(_, _ string) (*tmhi.InfoResult, error) {
	m.requestCalled = true

	return &tmhi.InfoResult{}, nil
}

func (m *mockGateway) Info() (*tmhi.InfoResult, error) {
	m.infoCalled = true
	if m.infoErr != nil {
		return nil, m.infoErr
	}

	return &tmhi.InfoResult{}, nil
}

func (m *mockGateway) Status() (*tmhi.StatusResult, error) {
	m.statusCalled = true
	if m.statusErr != nil {
		return nil, m.statusErr
	}

	return &tmhi.StatusResult{WebInterfaceUp: true}, nil
}

func (m *mockGateway) Signal() (*tmhi.SignalResult, error) {
	m.signalCalled = true
	if m.signalErr != nil {
		return nil, m.signalErr
	}

	return &tmhi.SignalResult{}, nil
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

//nolint:dupl
func TestLogin_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("success", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := login(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.loginCalled)
	})

	t.Run("failure", func(t *testing.T) {
		mg := &mockGateway{loginErr: errors.New("login failed")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := login(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
		assert.True(t, mg.loginCalled)
	})
}

//nolint:dupl
func TestInfo_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("success", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := info(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.infoCalled)
	})

	t.Run("failure", func(t *testing.T) {
		mg := &mockGateway{infoErr: errors.New("info boom")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := info(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "info command failed")
		assert.True(t, mg.infoCalled)
	})
}

//nolint:dupl
func TestStatus_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("success", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := status(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.statusCalled)
	})

	t.Run("failure", func(t *testing.T) {
		mg := &mockGateway{statusErr: errors.New("status boom")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := status(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Checking gateway status...")
		assert.True(t, mg.statusCalled)
	})
}

func TestReboot_DryRunFlagAndFailure(t *testing.T) {
	original := initGatewayFunc
	originalConfig := appConfig

	defer func() {
		initGatewayFunc = original
		appConfig = originalConfig
	}()

	t.Run("dry-run true returns early without calling gateway", func(t *testing.T) {
		appConfig = &Config{DryRun: true}
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		cmd := newRebootCmd(true)
		err := reboot(context.Background(), cmd)
		require.NoError(t, err)
		assert.False(t, mg.rebootCalled)
	})

	t.Run("dry-run false with error", func(t *testing.T) {
		appConfig = &Config{DryRun: false}
		mg := &mockGateway{rebootErr: errors.New("reboot boom")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		cmd := newRebootCmd(false)
		err := reboot(context.Background(), cmd)
		require.Error(t, err)
		assert.True(t, mg.rebootCalled)
	})
}

func TestReboot_ConfirmationDefaultsToNo(t *testing.T) {
	original := initGatewayFunc
	originalConfig := appConfig

	defer func() {
		initGatewayFunc = original
		appConfig = originalConfig
	}()

	t.Run("enter accepts default no", func(t *testing.T) {
		testRebootConfirmCancel(t, keys.Enter)
	})

	t.Run("y confirms reboot", func(t *testing.T) {
		testRebootConfirmProceed(t, 'y')
	})

	t.Run("n cancels reboot", func(t *testing.T) {
		testRebootConfirmCancel(t, 'n')
	})

	t.Run("auto confirm skips prompt", func(t *testing.T) {
		appConfig = &Config{DryRun: false}
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		cmd := &cli.Command{
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: ConfigDryRun, Value: false},
				&cli.BoolFlag{Name: ConfigAutoConfirm, Value: true},
			},
		}

		err := reboot(context.Background(), cmd)
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
	initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
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

	err := reboot(context.Background(), cmd)
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
	initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
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

	err := reboot(context.Background(), cmd)
	require.NoError(t, err)
	assert.True(t, mg.rebootCalled, "reboot should proceed")
}

//nolint:dupl
func TestSignal_SuccessAndFailure(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("success", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := signalCmd(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, mg.signalCalled)
	})

	t.Run("failure", func(t *testing.T) {
		mg := &mockGateway{signalErr: errors.New("signal boom")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		err := signalCmd(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Fetching signal information...")
		assert.True(t, mg.signalCalled)
	})
}

func TestReq_Command(t *testing.T) {
	original := initGatewayFunc

	defer func() { initGatewayFunc = original }()

	t.Run("wrong number of arguments", func(t *testing.T) {
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }

		reqCmd := &cli.Command{
			Name:   "req",
			Action: req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "login", Value: false},
			},
		}

		originalExiter := cli.OsExiter
		cli.OsExiter = func(_ int) {}

		defer func() { cli.OsExiter = originalExiter }()

		err := reqCmd.Run(context.Background(), []string{"tmhi-cli", "req"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly 2 arguments required")
	})
}

func TestInitGateway(t *testing.T) {
	t.Run("returns gateway on success", func(t *testing.T) {
		originalConfig := appConfig

		defer func() { appConfig = originalConfig }()

		appConfig = &Config{
			Model:    NOK5G21,
			IP:       "192.168.12.1",
			Username: "admin",
			Password: "test",
			Timeout:  5 * time.Second,
		}

		g, err := initGateway(appConfig)
		require.NoError(t, err)
		assert.NotNil(t, g)
	})
}
