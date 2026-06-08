package internal

import (
	"context"
	"errors"
	"testing"
	"time"

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

func (m *mockGateway) Login(context.Context) error {
	m.loginCalled = true

	return m.loginErr
}

func (m *mockGateway) Reboot(context.Context) error {
	m.rebootCalled = true

	return m.rebootErr
}

func (m *mockGateway) Request(context.Context, string, string) (*tmhi.InfoResult, error) {
	m.requestCalled = true

	return &tmhi.InfoResult{}, nil
}

func (m *mockGateway) Info(context.Context) (*tmhi.InfoResult, error) {
	m.infoCalled = true
	if m.infoErr != nil {
		return nil, m.infoErr
	}

	return &tmhi.InfoResult{}, nil
}

func (m *mockGateway) Status(context.Context) (*tmhi.StatusResult, error) {
	m.statusCalled = true
	if m.statusErr != nil {
		return nil, m.statusErr
	}

	return &tmhi.StatusResult{WebInterfaceUp: true}, nil
}

func (m *mockGateway) Signal(context.Context) (*tmhi.SignalResult, error) {
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

func TestHandlerSuccessAndFailure(t *testing.T) {
	tests := []struct {
		name      string
		handler   func(context.Context, *cli.Command) error
		called    func(*mockGateway) bool
		setupFail func(*mockGateway)
		errChecks []string
	}{
		{
			name:    cmdLogin,
			handler: login,
			called:  func(mg *mockGateway) bool { return mg.loginCalled },
			setupFail: func(mg *mockGateway) {
				mg.loginErr = errors.New("login failed")
			},
			errChecks: []string{"login failed"},
		},
		{
			name:    "info",
			handler: info,
			called:  func(mg *mockGateway) bool { return mg.infoCalled },
			setupFail: func(mg *mockGateway) {
				mg.infoErr = errors.New("info boom")
			},
			errChecks: []string{"Fetching gateway info", "info boom"},
		},
		{
			name:    cmdStatus,
			handler: status,
			called:  func(mg *mockGateway) bool { return mg.statusCalled },
			setupFail: func(mg *mockGateway) {
				mg.statusErr = errors.New("status boom")
			},
			errChecks: []string{"Checking gateway status..."},
		},
		{
			name:    cmdSignal,
			handler: signalCmd,
			called:  func(mg *mockGateway) bool { return mg.signalCalled },
			setupFail: func(mg *mockGateway) {
				mg.signalErr = errors.New("signal boom")
			},
			errChecks: []string{"Fetching signal information..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/success", func(t *testing.T) {
			orig := initGatewayFunc

			t.Cleanup(func() { initGatewayFunc = orig })

			mg := &mockGateway{}
			initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }

			err := tt.handler(t.Context(), nil)
			require.NoError(t, err)
			assert.True(t, tt.called(mg))
		})

		t.Run(tt.name+"/failure", func(t *testing.T) {
			orig := initGatewayFunc

			t.Cleanup(func() { initGatewayFunc = orig })

			mg := &mockGateway{}
			tt.setupFail(mg)

			initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }

			err := tt.handler(t.Context(), nil)
			require.Error(t, err)

			for _, check := range tt.errChecks {
				assert.Contains(t, err.Error(), check)
			}

			assert.True(t, tt.called(mg))
		})
	}
}

func TestReboot_DryRunFlagAndFailure(t *testing.T) {
	origInit := initGatewayFunc
	origConfig := appConfig

	t.Cleanup(func() {
		initGatewayFunc = origInit
		appConfig = origConfig
	})

	t.Run("dry-run true returns early without calling gateway", func(t *testing.T) {
		appConfig = &Config{DryRun: true}
		mg := &mockGateway{}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		cmd := newRebootCmd(true)
		err := reboot(t.Context(), cmd)
		require.NoError(t, err)
		assert.False(t, mg.rebootCalled)
	})

	t.Run("dry-run false with error", func(t *testing.T) {
		appConfig = &Config{DryRun: false}
		mg := &mockGateway{rebootErr: errors.New("reboot boom")}
		initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
		cmd := newRebootCmd(false)
		err := reboot(t.Context(), cmd)
		require.Error(t, err)
		assert.True(t, mg.rebootCalled)
	})
}

func TestReboot_ConfirmationDefaultsToNo(t *testing.T) {
	origInit := initGatewayFunc
	origConfig := appConfig

	t.Cleanup(func() {
		initGatewayFunc = origInit
		appConfig = origConfig
	})

	t.Run("enter accepts default no", func(t *testing.T) {
		testRebootConfirm(t, false, false, "reboot should be cancelled")
	})

	t.Run("y confirms reboot", func(t *testing.T) {
		testRebootConfirm(t, true, true, "reboot should proceed")
	})

	t.Run("n cancels reboot", func(t *testing.T) {
		testRebootConfirm(t, false, false, "reboot should be cancelled")
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

		err := reboot(t.Context(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should proceed with auto-confirm")
	})
}

func testRebootConfirm(t *testing.T, confirmResult bool, expectCalled bool, msg string) {
	t.Helper()

	origInit := initGatewayFunc
	origConfig := appConfig
	origConfirm := confirmDialog

	t.Cleanup(func() {
		initGatewayFunc = origInit
		appConfig = origConfig
		confirmDialog = origConfirm
	})

	appConfig = &Config{DryRun: false}
	mg := &mockGateway{}
	initGatewayFunc = func(_ *Config) (tmhi.Gateway, error) { return mg, nil }
	confirmDialog = func(_ string, _ bool) (bool, error) {
		return confirmResult, nil
	}
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: ConfigDryRun, Value: false},
			&cli.BoolFlag{Name: ConfigAutoConfirm, Value: false},
		},
	}

	err := reboot(t.Context(), cmd)
	require.NoError(t, err)
	assert.Equal(t, expectCalled, mg.rebootCalled, msg)
}

func TestReq_Command(t *testing.T) {
	orig := initGatewayFunc

	t.Cleanup(func() { initGatewayFunc = orig })

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

		origExiter := cli.OsExiter
		cli.OsExiter = func(_ int) {}

		t.Cleanup(func() { cli.OsExiter = origExiter })

		err := reqCmd.Run(t.Context(), []string{appName, cmdReq})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly 2 arguments required")
	})
}

func TestInitGateway(t *testing.T) {
	t.Run("returns gateway on success", func(t *testing.T) {
		origConfig := appConfig

		t.Cleanup(func() { appConfig = origConfig })

		appConfig = &Config{
			Model:    NOK5G21,
			IP:       defaultIP,
			Username: defaultUser,
			Password: "test",
			Timeout:  5 * time.Second,
		}

		g, err := initGateway(appConfig)
		require.NoError(t, err)
		assert.NotNil(t, g)
	})
}
