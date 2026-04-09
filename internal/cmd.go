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
)

type contextKey string

const (
	gatewayContextKey contextKey = "gateway"
)

// Gateway model constants.
const (
	ARCADYAN       string        = "ARCADYAN"
	NOK5G21        string        = "NOK5G21"
	DefaultTimeout time.Duration = 5 * time.Second
)

// Configuration flag names.
const (
	ConfigDryRun   string = "dry-run"
	ConfigConfig   string = "config"
	ConfigDebug    string = "debug"
	ConfigLogin    string = "login."
	ConfigUsername string = ConfigLogin + "username"
	ConfigPassword string = ConfigLogin + "password"
	ConfigGateway  string = "gateway."
	ConfigModel    string = ConfigGateway + "model"
	ConfigIP       string = ConfigGateway + "ip"
	ConfigTimeout  string = "timeout"
	ConfigRetries  string = "retries"
	// ConfigAutoConfirm enables auto-confirm prompts.
	ConfigAutoConfirm string = "yes"
)

// (No global LogSetup; switched to pterm-based output and explicit checks)

func commonContext(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	gateway, err := getGateway(cmd.Version,
		cmd.String(ConfigModel),
		cmd.String(ConfigUsername),
		cmd.String(ConfigPassword),
		cmd.String(ConfigIP),
		cmd.Duration(ConfigTimeout),
		cmd.Int(ConfigRetries),
		cmd.Bool(ConfigDebug),
	)
	if err != nil {
		pterm.Error.Println("could not instantiate gateway:", err)

		return nil, err
	}

	newCtx := context.WithValue(ctx, gatewayContextKey, gateway)

	return newCtx, nil
}

// Login handles the login CLI command.
func Login(ctx context.Context, _ *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.Gateway)

	err := gateway.Login()
	if err != nil {
		pterm.Error.Println("could not log in:", err)
	} else {
		pterm.Success.Println("successfully logged in")
	}

	return nil
}

// Req handles the req CLI command for custom HTTP requests.
func Req(ctx context.Context, cmd *cli.Command) error {
	const requiredArgsCount = 2
	if cmd.NArg() != requiredArgsCount {
		return cli.Exit("exactly 2 arguments required (HTTP method and path)", 1)
	}

	method := cmd.Args().Get(0)
	path := cmd.Args().Get(1)
	loginFirst := cmd.Bool("login")

	gateway, _ := ctx.Value(gatewayContextKey).(pkg.Gateway)
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
func Info(ctx context.Context, _ *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.Gateway)
	if err := gateway.Info(); err != nil {
		return fmt.Errorf("info command failed: %w", err)
	}

	return nil
}

// Status handles the status CLI command.
func Status(ctx context.Context, _ *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.Gateway)
	if err := gateway.Status(); err != nil {
		pterm.Error.Println("status check failed:", err)

		return fmt.Errorf("status check failed: %w", err)
	}

	return nil
}

// Signal handles the signal CLI command.
func Signal(ctx context.Context, _ *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.Gateway)
	if err := gateway.Signal(); err != nil {
		return fmt.Errorf("signal command failed: %w", err)
	}

	return nil
}

// Reboot handles the reboot CLI command.
func Reboot(ctx context.Context, cmd *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.Gateway)

	dryRun := cmd.Bool(ConfigDryRun)
	if !dryRun && !cmd.Bool(ConfigAutoConfirm) {
		confirm, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultValue(false).
			Show("Are you sure you want to reboot the gateway?")
		if !confirm {
			pterm.Warning.Println("Reboot cancelled")

			return nil
		}
	}

	err := gateway.Reboot(dryRun)
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

func buildFlags(configFile *string, configSource altsrc.Sourcer) []cli.Flag {
	return buildFlagsBaseCore(configFile, configSource)
}

func buildCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "login",
			Usage:  "Verify that the credentials can log the tool in",
			Action: Login,
		},
		{
			Name:   "reboot",
			Usage:  "Reboot the router",
			Action: Reboot,
		},
		{
			Name:   "info",
			Usage:  "Get gateway information",
			Action: Info,
		},
		{
			Name:   "status",
			Usage:  "Check gateway status",
			Action: Status,
		},
		{
			Name:   "signal",
			Usage:  "Display signal strength information",
			Action: Signal,
		},
		{
			Name:      "req",
			Usage:     "Make a custom HTTP request to the gateway",
			ArgsUsage: "<HTTP method> <path>",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "login",
					Aliases: []string{"l"},
					Value:   false,
					Usage:   "login before making request",
				},
			},
			Action: Req,
		},
	}
}

// Cmd is the main entry point for the CLI application.
func Cmd(version string) {
	var configFile string

	configSource := altsrc.NewStringPtrSourcer(&configFile)

	app := &cli.Command{
		Name:     "tmhi-cli",
		Usage:    "Utility to interact with T-Mobile Home Internet gateway",
		Version:  version,
		Flags:    buildFlags(&configFile, configSource),
		Commands: buildCommands(),
		Before:   commonContext,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		pterm.Fatal.Println("application error:", err)
		os.Exit(1)
	}
}
