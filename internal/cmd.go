// Package internal provides CLI command handling for tmhi-cli.
package internal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/pterm/pterm"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

// Gateway model constants.
const (
	ARCADYAN       string        = "ARCADYAN"
	NOK5G21        string        = "NOK5G21"
	DefaultTimeout time.Duration = 5 * time.Second
)

//nolint:gochecknoglobals
var initGatewayFunc = initGateway

//nolint:ireturn
func initGateway(_ *Config) (pkg.Gateway, error) {
	gateway, err := getGateway(appConfig)
	if err != nil {
		pterm.Error.Println("could not instantiate gateway:", err)

		return nil, err
	}

	return gateway, nil
}

// Login handles the login CLI command.
func Login(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	err = gateway.Login()
	if err != nil {
		pterm.Error.Println("could not log in:", err)

		return cli.Exit("login failed", 1)
	}

	pterm.Success.Println("successfully logged in")

	return nil
}

// Req handles the req CLI command for custom HTTP requests.
func Req(_ context.Context, cmd *cli.Command) error {
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
		if err := gateway.Login(); err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
	}

	if err := gateway.Request(method, path); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}

// Info handles the info CLI command.
func Info(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	if err := gateway.Info(); err != nil {
		return fmt.Errorf("info command failed: %w", err)
	}

	return nil
}

// Status handles the status CLI command.
func Status(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	if err := gateway.Status(); err != nil {
		pterm.Error.Println("status check failed:", err)

		return fmt.Errorf("status check failed: %w", err)
	}

	return nil
}

// Signal handles the signal CLI command.
func Signal(_ context.Context, _ *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	if err := gateway.Signal(); err != nil {
		return fmt.Errorf("signal command failed: %w", err)
	}

	return nil
}

// Reboot handles the reboot CLI command.
func Reboot(_ context.Context, cmd *cli.Command) error {
	gateway, err := initGatewayFunc(appConfig)
	if err != nil {
		return err
	}

	if !cmd.Bool(ConfigAutoConfirm) {
		confirm, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultValue(false).
			Show("Are you sure you want to reboot the gateway?")
		if !confirm {
			pterm.Warning.Println("Reboot cancelled")

			return nil
		}
	}

	err = gateway.Reboot(appConfig.DryRun)
	if err != nil {
		pterm.Error.Println("could not reboot gateway:", err)

		return fmt.Errorf("could not reboot gateway: %w", err)
	}

	return nil
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".tmhi-cli.toml"
	}

	return home + "/.tmhi-cli.toml"
}

// setupColor handles the --color flag logic before command execution.
func setupColor(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	colorValue := cmd.String(ConfigColor)
	switch colorValue {
	case "always":
		// pterm default - colors enabled
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

// Cmd is the main entry point for the CLI application.
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
