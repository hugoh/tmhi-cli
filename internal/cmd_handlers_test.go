package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	tmhi "github.com/hugoh/tmhi-gateway/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const (
	testReqMethod = "GET"
	testReqPath   = "/test"
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
	requestErr    error
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
	if m.requestErr != nil {
		return nil, m.requestErr
	}

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
		handler   func(*app) cli.ActionFunc
		called    func(*mockGateway) bool
		setupFail func(*mockGateway)
		errChecks []string
	}{
		{
			name:    cmdLogin,
			handler: func(a *app) cli.ActionFunc { return a.login },
			called:  func(mg *mockGateway) bool { return mg.loginCalled },
			setupFail: func(mg *mockGateway) {
				mg.loginErr = errors.New("login failed")
			},
			errChecks: []string{"login failed"},
		},
		{
			name:    "info",
			handler: func(a *app) cli.ActionFunc { return a.info },
			called:  func(mg *mockGateway) bool { return mg.infoCalled },
			setupFail: func(mg *mockGateway) {
				mg.infoErr = errors.New("info boom")
			},
			errChecks: []string{"Fetching gateway info", "info boom"},
		},
		{
			name:    cmdStatus,
			handler: func(a *app) cli.ActionFunc { return a.status },
			called:  func(mg *mockGateway) bool { return mg.statusCalled },
			setupFail: func(mg *mockGateway) {
				mg.statusErr = errors.New("status boom")
			},
			errChecks: []string{"Checking gateway status..."},
		},
		{
			name:    cmdSignal,
			handler: func(a *app) cli.ActionFunc { return a.signal },
			called:  func(mg *mockGateway) bool { return mg.signalCalled },
			setupFail: func(mg *mockGateway) {
				mg.signalErr = errors.New("signal boom")
			},
			errChecks: []string{"Fetching signal information..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/success", func(t *testing.T) {
			mg := &mockGateway{}
			a := newTestApp(mg)

			err := tt.handler(a)(t.Context(), nil)
			require.NoError(t, err)
			assert.True(t, tt.called(mg))
		})

		t.Run(tt.name+"/failure", func(t *testing.T) {
			mg := &mockGateway{}
			tt.setupFail(mg)

			a := newTestApp(mg)

			err := tt.handler(a)(t.Context(), nil)
			require.Error(t, err)

			for _, check := range tt.errChecks {
				assert.Contains(t, err.Error(), check)
			}

			assert.True(t, tt.called(mg))
		})
	}
}

func TestReboot_DryRunFlagAndFailure(t *testing.T) {
	t.Run("dry-run true returns early without calling gateway", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)
		a.config = &Config{DryRun: true}
		cmd := newRebootCmd(true)
		err := a.reboot(t.Context(), cmd)
		require.NoError(t, err)
		assert.False(t, mg.rebootCalled)
	})

	t.Run("dry-run false with error", func(t *testing.T) {
		mg := &mockGateway{rebootErr: errors.New("reboot boom")}
		a := newTestApp(mg)
		a.config = &Config{DryRun: false}
		cmd := newRebootCmd(false)
		err := a.reboot(t.Context(), cmd)
		require.Error(t, err)
		assert.True(t, mg.rebootCalled)
	})
}

func TestReboot_ConfirmationDefaultsToNo(t *testing.T) {
	t.Run("enter accepts default no", func(t *testing.T) {
		testRebootConfirm(t, false, false, "reboot should be cancelled")
	})

	t.Run("y confirms reboot", func(t *testing.T) {
		testRebootConfirm(t, true, true, "reboot should proceed")
	})

	t.Run("n cancels reboot", func(t *testing.T) {
		testRebootConfirm(t, false, false, "reboot should be cancelled")
	})

	t.Run("confirmation error aborts reboot", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)
		a.config = &Config{DryRun: false}
		a.confirm = func(_ string, _ bool) (bool, error) {
			return false, errors.New("prompt broken")
		}
		cmd := &cli.Command{
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: ConfigDryRun, Value: false},
				&cli.BoolFlag{Name: ConfigAutoConfirm, Value: false},
			},
		}

		err := a.reboot(t.Context(), cmd)
		require.Error(t, err)
		assert.False(t, mg.rebootCalled, "reboot should not proceed on prompt failure")
	})

	t.Run("auto confirm skips prompt", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)
		a.config = &Config{DryRun: false}
		cmd := &cli.Command{
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: ConfigDryRun, Value: false},
				&cli.BoolFlag{Name: ConfigAutoConfirm, Value: true},
			},
		}

		err := a.reboot(t.Context(), cmd)
		require.NoError(t, err)
		assert.True(t, mg.rebootCalled, "reboot should proceed with auto-confirm")
	})
}

func testRebootConfirm(t *testing.T, confirmResult bool, expectCalled bool, msg string) {
	t.Helper()

	mg := &mockGateway{}
	a := newTestApp(mg)
	a.config = &Config{DryRun: false}
	a.confirm = func(_ string, _ bool) (bool, error) {
		return confirmResult, nil
	}
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: ConfigDryRun, Value: false},
			&cli.BoolFlag{Name: ConfigAutoConfirm, Value: false},
		},
	}

	err := a.reboot(t.Context(), cmd)
	require.NoError(t, err)
	assert.Equal(t, expectCalled, mg.rebootCalled, msg)
}

func TestReq_Command(t *testing.T) {
	t.Run("wrong number of arguments", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		reqCmd := &cli.Command{
			Name:   "req",
			Action: a.req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "login", Value: false},
			},
		}

		exited := false
		origExiter := cli.OsExiter
		cli.OsExiter = func(_ int) { exited = true }

		t.Cleanup(func() { cli.OsExiter = origExiter })

		err := reqCmd.Run(t.Context(), []string{appName, cmdReq})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly 2 arguments required")
		assert.False(
			t,
			exited,
			"argument validation should return a normal error, not exit the process",
		)
	})

	t.Run("empty method is rejected", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		reqCmd := &cli.Command{
			Name:   cmdReq,
			Action: a.req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: cmdLogin, Value: false},
			},
		}

		err := reqCmd.Run(t.Context(), []string{cmdReq, "", testReqPath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP method must not be empty")
		assert.False(t, mg.requestCalled, "request should not be performed with an empty method")
	})

	t.Run("dry-run does not perform the request", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)
		a.config = &Config{DryRun: true}

		reqCmd := &cli.Command{
			Name:   cmdReq,
			Action: a.req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: cmdLogin, Value: false},
			},
		}

		err := reqCmd.Run(t.Context(), []string{cmdReq, testReqMethod, testReqPath})
		require.NoError(t, err)
		assert.False(t, mg.requestCalled, "request should be skipped in dry-run mode")
		assert.False(t, mg.loginCalled, "login should be skipped in dry-run mode")
	})

	t.Run("performs request", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		reqCmd := &cli.Command{
			Name:   cmdReq,
			Action: a.req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: cmdLogin, Value: false},
			},
		}

		err := reqCmd.Run(t.Context(), []string{cmdReq, testReqMethod, testReqPath})
		require.NoError(t, err)
		assert.True(t, mg.requestCalled, "request should be performed")
		assert.False(t, mg.loginCalled, "login should be skipped without --login")
	})

	t.Run("logs in first with --login", func(t *testing.T) {
		mg := &mockGateway{}
		a := newTestApp(mg)

		reqCmd := &cli.Command{
			Name:   cmdReq,
			Action: a.req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: cmdLogin, Value: false},
			},
		}

		err := reqCmd.Run(t.Context(), []string{cmdReq, "--login", testReqMethod, testReqPath})
		require.NoError(t, err)
		assert.True(t, mg.loginCalled, "login should be performed first")
		assert.True(t, mg.requestCalled, "request should be performed")
	})

	t.Run("login failure aborts request", func(t *testing.T) {
		mg := &mockGateway{loginErr: errors.New("bad credentials")}
		a := newTestApp(mg)

		reqCmd := &cli.Command{
			Name:   cmdReq,
			Action: a.req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: cmdLogin, Value: false},
			},
		}

		err := reqCmd.Run(t.Context(), []string{cmdReq, "--login", testReqMethod, testReqPath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
		assert.False(t, mg.requestCalled, "request should not be performed")
	})

	t.Run("request failure returns error", func(t *testing.T) {
		mg := &mockGateway{requestErr: errors.New("gateway timeout")}
		a := newTestApp(mg)

		reqCmd := &cli.Command{
			Name:   cmdReq,
			Action: a.req,
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: cmdLogin, Value: false},
			},
		}

		err := reqCmd.Run(t.Context(), []string{cmdReq, testReqMethod, testReqPath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
		assert.Contains(t, err.Error(), "gateway timeout")
		assert.True(t, mg.requestCalled, "request should have been attempted")
	})
}

func TestInitGateway(t *testing.T) {
	t.Run("returns gateway on success", func(t *testing.T) {
		cfg := &Config{
			Model:    NOK5G21,
			IP:       defaultIP,
			Username: defaultUser,
			Password: "test",
			Timeout:  5 * time.Second,
		}

		g, err := initGateway(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)
	})
}
