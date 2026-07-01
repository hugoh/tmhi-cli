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

	tmhi "github.com/hugoh/tmhi-gateway/v2"
	"github.com/muesli/termenv"
	"github.com/pterm/pterm"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

type spinner interface {
	Fail(message ...any)
	Success(message ...any)
}

// app carries the configuration and the dependencies command actions use,
// so tests can swap them without mutating package state.
type app struct {
	config      *Config
	initGateway func(*Config) (tmhi.Gateway, error)
	newSpinner  func(message string) (spinner, error)
	confirm     func(ctx context.Context, msg string, defaultVal bool) (bool, error)
}

func newApp() *app {
	return &app{
		config:      &Config{},
		initGateway: initGateway,
		newSpinner:  newPtermSpinner,
		confirm:     ptermConfirm,
	}
}

//nolint:ireturn
func newPtermSpinner(message string) (spinner, error) {
	sp, err := pterm.DefaultSpinner.Start(message)
	if err != nil {
		return nil, fmt.Errorf("failed to start spinner: %w", err)
	}

	return &spinnerWrapper{spinnerPrinter: sp}, nil
}

// ptermConfirm shows an interactive confirmation prompt, returning early with
// ctx.Err() if ctx is cancelled (e.g. by SIGINT) before the user answers.
// The prompt itself keeps blocking on stdin in the background since pterm
// offers no way to interrupt it.
func ptermConfirm(ctx context.Context, msg string, defaultVal bool) (bool, error) {
	type confirmResult struct {
		confirmed bool
		err       error
	}

	resultCh := make(chan confirmResult, 1)

	go func() {
		confirmed, err := pterm.DefaultInteractiveConfirm.
			WithDefaultValue(defaultVal).
			Show(msg)
		if err != nil {
			resultCh <- confirmResult{err: fmt.Errorf("confirmation failed: %w", err)}

			return
		}

		resultCh <- confirmResult{confirmed: confirmed}
	}()

	select {
	case <-ctx.Done():
		return false, ctx.Err() //nolint:wrapcheck
	case r := <-resultCh:
		return r.confirmed, r.err
	}
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
		// WithRemoveWhenDone has a value receiver and returns a new copy, so
		// calling Stop() on that copy never reaches the goroutine's IsActive.
		// Set the field on the original pointer and stop it directly.
		w.spinnerPrinter.RemoveWhenDone = true
		_ = w.spinnerPrinter.Stop()

		return
	}

	w.spinnerPrinter.Success(message...)
}

// Gateway model constants.
const (
	ARCADYAN       string        = "ARCADYAN"
	NOK5G21        string        = "NOK5G21"
	DefaultTimeout time.Duration = 5 * time.Second

	appName     = "tmhi-cli"
	defaultIP   = "192.168.12.1"
	defaultUser = "admin"
	autoValue   = "auto"
	cmdLogin    = "login"
	cmdReq      = "req"
	cmdStatus   = "status"
	cmdSignal   = "signal"
)

// ErrReqArgs is returned when req is not given exactly an HTTP method and a path.
var ErrReqArgs = errors.New("exactly 2 arguments required (HTTP method and path)")

// ErrReqMethod is returned when req is given an empty HTTP method.
var ErrReqMethod = errors.New("HTTP method must not be empty")

// fetchWithFeedback runs an operation with a spinner, handling success/failure.
// It starts a spinner with the given message, executes the fetch function,
// displays the result using the display function, and properly stops the spinner.
//
//nolint:ireturn
func fetchWithFeedback[T any](
	ctx context.Context,
	newSpinner func(string) (spinner, error),
	message string,
	fetch func(context.Context) (T, error),
	display func(T),
	successMessage ...any,
) (T, error) {
	spinnerInstance, err := newSpinner(message)
	if err != nil {
		var zero T

		return zero, err
	}

	result, opErr := fetch(ctx)
	if opErr != nil {
		spinnerInstance.Fail(fmt.Sprintf("%s: %v", message, opErr))

		var zero T

		return zero, displayed(fmt.Errorf("%s: %w", message, opErr))
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
	newSpinner func(string) (spinner, error),
	message string,
	run func(context.Context) error,
	successMessage ...any,
) error {
	_, err := fetchWithFeedback(
		ctx,
		newSpinner,
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

	return getGateway(cfg, "")
}

func (a *app) login(ctx context.Context, _ *cli.Command) error {
	gateway, err := a.initGateway(a.config)
	if err != nil {
		return err
	}

	return runWithFeedback(
		ctx,
		a.newSpinner,
		"Logging in...",
		gateway.Login,
		"Successfully logged in",
	)
}

func (a *app) req(ctx context.Context, cmd *cli.Command) error {
	const requiredArgsCount = 2
	if cmd.NArg() != requiredArgsCount {
		return ErrReqArgs
	}

	method := cmd.Args().Get(0)
	path := cmd.Args().Get(1)

	if method == "" {
		return ErrReqMethod
	}

	gateway, err := a.initGateway(a.config)
	if err != nil {
		return err
	}

	if a.config.DryRun {
		pterm.Info.Printfln("Dry run - would send %s %s request", method, path)

		return nil
	}

	if cmd.Bool("login") {
		if err := runWithFeedback(
			ctx,
			a.newSpinner,
			"Logging in...",
			gateway.Login,
			"Successfully logged in",
		); err != nil {
			return err
		}
	}

	_, err = fetchWithFeedback(
		ctx,
		a.newSpinner,
		fmt.Sprintf("%s %s...", method, path),
		func(ctx context.Context) (*tmhi.InfoResult, error) {
			return gateway.Request(ctx, method, path)
		},
		displayInfoResult,
	)

	return err
}

func (a *app) info(ctx context.Context, _ *cli.Command) error {
	gateway, err := a.initGateway(a.config)
	if err != nil {
		return err
	}

	_, err = fetchWithFeedback(
		ctx,
		a.newSpinner,
		"Fetching gateway info...",
		gateway.Info,
		displayInfoResult,
	)

	return err
}

func (a *app) status(ctx context.Context, _ *cli.Command) error {
	gateway, err := a.initGateway(a.config)
	if err != nil {
		return err
	}

	_, err = fetchWithFeedback(
		ctx,
		a.newSpinner,
		"Checking gateway status...",
		gateway.Status,
		displayStatusResult,
	)

	return err
}

func (a *app) signal(ctx context.Context, _ *cli.Command) error {
	gateway, err := a.initGateway(a.config)
	if err != nil {
		return err
	}

	_, err = fetchWithFeedback(
		ctx,
		a.newSpinner,
		"Fetching signal information...",
		gateway.Signal,
		displaySignalResult,
	)

	return err
}

func (a *app) reboot(ctx context.Context, cmd *cli.Command) error {
	gateway, err := a.initGateway(a.config)
	if err != nil {
		return err
	}

	if a.config.DryRun {
		pterm.Info.Println("Dry run - would send reboot request")

		return nil
	}

	if !cmd.Bool(ConfigAutoConfirm) {
		confirmed, confirmErr := a.confirm(
			ctx,
			"Are you sure you want to reboot the gateway?",
			false,
		)
		if confirmErr != nil {
			return confirmErr
		}

		if !confirmed {
			pterm.Warning.Println("Reboot cancelled")

			return nil
		}
	}

	return runWithFeedback(
		ctx,
		a.newSpinner,
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

func applyLightBgTheme() {
	// FgLight* colors are near-white and invisible on light backgrounds.
	pterm.ThemeDefault.PrimaryStyle = pterm.Style{pterm.FgCyan}
	pterm.ThemeDefault.SecondaryStyle = pterm.Style{pterm.FgMagenta}
	pterm.ThemeDefault.InfoMessageStyle = pterm.Style{pterm.FgCyan}
	pterm.ThemeDefault.WarningMessageStyle = pterm.Style{pterm.FgYellow, pterm.Bold}
	pterm.ThemeDefault.ErrorMessageStyle = pterm.Style{pterm.FgRed}
	pterm.ThemeDefault.FatalMessageStyle = pterm.Style{pterm.FgRed}
	pterm.ThemeDefault.SpinnerStyle = pterm.Style{pterm.FgCyan}
	pterm.ThemeDefault.SpinnerTextStyle = pterm.Style{pterm.FgDefault}
	pterm.ThemeDefault.TableHeaderStyle = pterm.Style{pterm.FgCyan, pterm.Bold}
	pterm.ThemeDefault.ProgressbarTitleStyle = pterm.Style{pterm.FgCyan}
	pterm.ThemeDefault.HeatmapHeaderStyle = pterm.Style{pterm.FgCyan}
	pterm.ThemeDefault.BarLabelStyle = pterm.Style{pterm.FgCyan}
	// For elements with their own background, FgLightWhite is readable
	// regardless of the terminal background — the badge/header bg provides contrast.
	pterm.ThemeDefault.InfoPrefixStyle = pterm.Style{pterm.FgLightWhite, pterm.BgCyan}
	pterm.ThemeDefault.SuccessPrefixStyle = pterm.Style{pterm.FgLightWhite, pterm.BgGreen}
	pterm.ThemeDefault.WarningPrefixStyle = pterm.Style{pterm.FgLightWhite, pterm.BgYellow}
	pterm.ThemeDefault.ErrorPrefixStyle = pterm.Style{pterm.FgLightWhite, pterm.BgRed}
	pterm.ThemeDefault.FatalPrefixStyle = pterm.Style{pterm.FgLightWhite, pterm.BgRed}
}

func setupColor(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	colorValue := cmd.String(ConfigColor)
	switch colorValue {
	case "always":
		if !termenv.HasDarkBackground() {
			applyLightBgTheme()
		}
	case autoValue:
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			pterm.DisableStyling()
		} else if !termenv.HasDarkBackground() {
			applyLightBgTheme()
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
	cliApp := newApp()
	cliApp.initGateway = func(cfg *Config) (tmhi.Gateway, error) {
		if err := cfg.Validate(); err != nil {
			return nil, err
		}

		return getGateway(cfg, appName+"/"+version)
	}

	root := &cli.Command{
		Name:     appName,
		Usage:    "Utility to interact with T-Mobile Home Internet gateway",
		Version:  version,
		Flags:    cliApp.flags(&configFile, configSource),
		Commands: cliApp.commands(),
		Before:   setupColor,
		OnUsageError: func(_ context.Context, cmd *cli.Command, err error, _ bool) error {
			_, _ = fmt.Fprintf(cmd.ErrWriter, "error: %v\n", err)

			return displayed(err)
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := root.Run(ctx, os.Args)
	if err != nil {
		if _, ok := errors.AsType[*displayedError](err); !ok {
			pterm.Error.Println(err)
		}
	}

	return err //nolint:wrapcheck
}
