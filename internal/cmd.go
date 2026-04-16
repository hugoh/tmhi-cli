// Package internal provides CLI command handling for tmhi-cli.
package internal

import (
	"context"
	"fmt"
	"os"
	"time"

	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/pterm/pterm"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

type spinner interface {
	Fail(message ...any)
	Success(message ...any)
	Stop() error
}

// spinnerFunc creates a new spinner. Overrideable for testing.
//
//nolint:gochecknoglobals
var spinnerFunc = func(message string) (spinner, error) {
	sp, err := pterm.DefaultSpinner.Start(message)
	if err != nil {
		return nil, fmt.Errorf("failed to start spinner: %w", err)
	}

	return &spinnerWrapper{spinnerPrinter: sp}, nil
}

// spinnerWrapper wraps pterm.SpinnerPrinter to implement the spinner interface.
type spinnerWrapper struct {
	spinnerPrinter *pterm.SpinnerPrinter
}

func (w *spinnerWrapper) Fail(message ...any) {
	w.spinnerPrinter.Fail(message...)
}

func (w *spinnerWrapper) Success(message ...any) {
	if message != nil {
		w.spinnerPrinter.Success(message...)

		return
	}

	_ = w.spinnerPrinter.WithRemoveWhenDone().Stop()
}

func (w *spinnerWrapper) Stop() error {
	if err := w.spinnerPrinter.Stop(); err != nil {
		return fmt.Errorf("failed to stop spinner: %w", err)
	}

	return nil
}

// confirmDialog prompts the user for confirmation. Overrideable for testing.
//
//nolint:gochecknoglobals
var confirmDialog = func(msg string, defaultVal bool) (bool, error) {
	return pterm.DefaultInteractiveConfirm.
		WithDefaultValue(defaultVal).
		Show(msg)
}

// Gateway model constants.
const (
	ARCADYAN       string        = "ARCADYAN"
	NOK5G21        string        = "NOK5G21"
	DefaultTimeout time.Duration = 5 * time.Second
)

//nolint:gochecknoglobals
var initGatewayFunc = initGateway

// withSpinner runs an operation with a spinner, handling success/failure.
// It starts a spinner with the given message, executes the function,
// and properly stops the spinner on success or failure.
//
//nolint:ireturn
func withSpinner[T any](
	message string,
	operation func() (T, error),
	successMessage ...any,
) (T, error) {
	spinnerInstance, err := spinnerFunc(message)
	if err != nil {
		return *new(T), err
	}

	result, opErr := operation()
	if opErr != nil {
		spinnerInstance.Fail(fmt.Sprintf("%s: %v", message, opErr))

		return result, fmt.Errorf("%s: %w", message, opErr)
	}

	spinnerInstance.Success(successMessage...)

	return result, nil
}

//nolint:ireturn
func initGateway(_ *Config) (tmhi.Gateway, error) {
	gateway, err := getGateway(appConfig)
	if err != nil {
		pterm.Error.Println("could not instantiate gateway:", err)

		return nil, err
	}

	return gateway, nil
}

func login(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	result, err := withSpinner("Checking logging in...", func() (*tmhi.LoginResult, error) {
		return gateway.Login()
	})
	if err != nil {
		return err
	}

	displayLoginResult(result)

	return nil
}

func req(_ context.Context, cmd *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	const requiredArgsCount = 2
	if cmd.NArg() != requiredArgsCount {
		return cli.Exit("exactly 2 arguments required (HTTP method and path)", 1)
	}

	method := cmd.Args().Get(0)
	path := cmd.Args().Get(1)
	loginFirst := cmd.Bool("login")

	if loginFirst {
		if _, err := gateway.Login(); err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
	}

	result, err := gateway.Request(method, path)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	displayInfoResult(result)

	return nil
}

func info(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	result, err := gateway.Info()
	if err != nil {
		return fmt.Errorf("info command failed: %w", err)
	}

	displayInfoResult(result)

	return nil
}

func status(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	result, err := withSpinner("Checking gateway status...", func() (*tmhi.StatusResult, error) {
		return gateway.Status()
	})
	if err != nil {
		return err
	}

	displayStatusResult(result)

	return nil
}

func signalCmd(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	result, err := withSpinner(
		"Fetching signal information...",
		func() (*tmhi.SignalResult, error) {
			return gateway.Signal()
		},
	)
	if err != nil {
		return err
	}

	displaySignalResult(result)

	return nil
}

func reboot(_ context.Context, cmd *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	if !cmd.Bool(ConfigAutoConfirm) {
		confirmed, confirmErr := confirmDialog(
			"Are you sure you want to reboot the gateway?",
			false,
		)
		if confirmErr != nil {
			return fmt.Errorf("confirmation failed: %w", confirmErr)
		}

		if !confirmed {
			pterm.Warning.Println("Reboot cancelled")

			return nil
		}
	}

	if appConfig.DryRun {
		pterm.Info.Println("Dry run - would send reboot request")

		return nil
	}

	_, ret := withSpinner(
		"Rebooting gateway...",
		func() (*tmhi.SignalResult, error) {
			rebootErr := gateway.Reboot()
			if rebootErr != nil {
				return nil, fmt.Errorf("Reboot failed: %w", rebootErr)
			}

			return nil, nil //nolint:nilnil // no error to report
		},
		"Reboot command sent successfully",
	)

	return ret
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".tmhi-cli.toml"
	}

	return home + "/.tmhi-cli.toml"
}

func setupColor(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	colorValue := cmd.String(ConfigColor)
	switch colorValue {
	case "always":
	case "auto":
		//nolint:gosec // Fd() returns a valid int on all platforms
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			pterm.DisableStyling()
		}
	case "never":
		pterm.DisableStyling()
	}

	return ctx, nil
}

// Cmd runs the CLI application with the given version string.
func Cmd(version string) {
	var configFile string

	configSource := altsrc.NewStringPtrSourcer(&configFile)

	app := &cli.Command{
		Name:     "tmhi-cli",
		Usage:    "Utility to interact with T-Mobile Home Internet gateway",
		Version:  version,
		Flags:    cmdFlags(&configFile, configSource),
		Commands: cmdCommands(),
		Before:   setupColor,
		OnUsageError: func(_ context.Context, cmd *cli.Command, err error, _ bool) error {
			_, _ = fmt.Fprintf(cmd.ErrWriter, "error: %v\n", err)

			return err
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
