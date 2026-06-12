// Package internal provides CLI command handling for tmhi-cli.
package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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
}

// spinnerFunc creates a new spinner. Overridable for testing.
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
	if len(message) == 0 {
		_ = w.spinnerPrinter.WithRemoveWhenDone().Stop()

		return
	}

	w.spinnerPrinter.Success(message...)
}

// confirmDialog prompts the user for confirmation. Overridable for testing.
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

	appName      = "tmhi-cli"
	defaultIP    = "192.168.12.1"
	defaultUser  = "admin"
	autoValue    = "auto"
	cmdLogin     = "login"
	cmdReq       = "req"
	cmdStatus    = "status"
	cmdSignal    = "signal"
	testAPN      = "test.apn"
	testRegState = "registered"
)

//nolint:gochecknoglobals
var initGatewayFunc = initGateway

// fetchWithFeedback runs an operation with a spinner, handling success/failure.
// It starts a spinner with the given message, executes the fetch function,
// displays the result using the display function, and properly stops the spinner.
//
//nolint:ireturn
func fetchWithFeedback[T any](
	ctx context.Context,
	message string,
	fetch func(context.Context) (T, error),
	display func(T),
	successMessage ...any,
) (T, error) {
	spinnerInstance, err := spinnerFunc(message)
	if err != nil {
		var zero T

		return zero, err
	}

	result, opErr := fetch(ctx)
	if opErr != nil {
		spinnerInstance.Fail(fmt.Sprintf("%s: %v", message, opErr))

		return result, displayed(fmt.Errorf("%s: %w", message, opErr))
	}

	spinnerInstance.Success(successMessage...)

	if display != nil {
		display(result)
	}

	return result, nil
}

// runWithFeedback runs an operation with a spinner when there is no result
// to display.
func runWithFeedback(
	ctx context.Context,
	message string,
	run func(context.Context) error,
	successMessage ...any,
) error {
	_, err := fetchWithFeedback(
		ctx,
		message,
		func(ctx context.Context) (struct{}, error) {
			return struct{}{}, run(ctx)
		},
		nil,
		successMessage...,
	)

	return err
}

//nolint:ireturn
func initGateway(cfg *Config) (tmhi.Gateway, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return getGateway(cfg)
}

func login(ctx context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	return runWithFeedback(ctx, "Logging in...", gateway.Login, "Successfully logged in")
}

func req(ctx context.Context, cmd *cli.Command) error {
	const requiredArgsCount = 2
	if cmd.NArg() != requiredArgsCount {
		return cli.Exit("exactly 2 arguments required (HTTP method and path)", 1)
	}

	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	method := cmd.Args().Get(0)
	path := cmd.Args().Get(1)
	loginFirst := cmd.Bool("login")

	if loginFirst {
		if err := gateway.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	result, err := gateway.Request(ctx, method, path)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	displayInfoResult(result)

	return nil
}

func info(ctx context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	_, err = fetchWithFeedback(ctx, "Fetching gateway info...", gateway.Info, displayInfoResult)

	return err
}

func status(ctx context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	_, err = fetchWithFeedback(
		ctx,
		"Checking gateway status...",
		gateway.Status,
		displayStatusResult,
	)

	return err
}

func signalCmd(ctx context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	_, err = fetchWithFeedback(
		ctx,
		"Fetching signal information...",
		gateway.Signal,
		displaySignalResult,
	)

	return err
}

func reboot(ctx context.Context, cmd *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	if appConfig.DryRun {
		pterm.Info.Println("Dry run - would send reboot request")

		return nil
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

	return runWithFeedback(
		ctx,
		"Rebooting gateway...",
		gateway.Reboot,
		"Reboot command sent successfully",
	)
}

func defaultConfigPath() string {
	const configFileName = ".tmhi-cli.toml"

	home, err := os.UserHomeDir()
	if err != nil {
		return configFileName
	}

	return filepath.Join(home, configFileName)
}

func setupColor(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	colorValue := cmd.String(ConfigColor)
	switch colorValue {
	case "always":
	case autoValue:
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			pterm.DisableStyling()
		}
	case "never":
		pterm.DisableStyling()
	}

	return ctx, nil
}

// Cmd runs the CLI application with the given version string.
func Cmd(version string) error {
	var configFile string

	configSource := altsrc.NewStringPtrSourcer(&configFile)

	app := &cli.Command{
		Name:     appName,
		Usage:    "Utility to interact with T-Mobile Home Internet gateway",
		Version:  version,
		Flags:    cmdFlags(&configFile, configSource),
		Commands: cmdCommands(),
		Before:   setupColor,
		OnUsageError: func(_ context.Context, cmd *cli.Command, err error, _ bool) error {
			_, _ = fmt.Fprintf(cmd.ErrWriter, "error: %v\n", err)

			return displayed(err)
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := app.Run(ctx, os.Args)
	if err != nil {
		if _, ok := errors.AsType[*displayedError](err); !ok {
			pterm.Error.Println(err)
		}
	}

	return err //nolint:wrapcheck
}
