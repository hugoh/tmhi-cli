package internal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
	altsrc "github.com/urfave/cli-altsrc/v3"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

type contextKey string

const (
	gatewayContextKey contextKey = "gateway"
)

const (
	ARCADYAN       string        = "ARCADYAN"
	NOK5G21        string        = "NOK5G21"
	DefaultTimeout time.Duration = 5 * time.Second
)

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
)

func LogSetup(debugFlag bool) {
	if debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

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
		logrus.WithError(err).Fatal("could not instantiate gateway")
		// NOTREACHED
	}
	newCtx := context.WithValue(ctx, gatewayContextKey, gateway)
	return newCtx, nil
}

func Login(ctx context.Context, _ *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.GatewayI)
	err := gateway.Login()
	if err != nil {
		logrus.WithError(err).Fatal("could not log in")
	} else {
		logrus.Info("successfully logged in")
	}
	return nil
}

func Req(ctx context.Context, cmd *cli.Command) error {
	const requiredArgsCount = 2
	if cmd.NArg() != requiredArgsCount {
		return cli.Exit("exactly 2 arguments required (HTTP method and path)", 1)
	}
	method := cmd.Args().Get(0)
	path := cmd.Args().Get(1)
	loginFirst := cmd.Bool("login")

	gateway, _ := ctx.Value(gatewayContextKey).(pkg.GatewayI)
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

func Info(ctx context.Context, _ *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.GatewayI)
	if err := gateway.Info(); err != nil {
		return fmt.Errorf("info command failed: %w", err)
	}
	return nil
}

func Status(ctx context.Context, _ *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.GatewayI)
	if err := gateway.Status(); err != nil {
		logrus.WithError(err).Error("status check failed")
		return fmt.Errorf("status check failed: %w", err)
	}
	return nil
}

func Reboot(ctx context.Context, cmd *cli.Command) error {
	gateway, _ := ctx.Value(gatewayContextKey).(pkg.GatewayI)
	err := gateway.Reboot(cmd.Bool(ConfigDryRun))
	if err != nil {
		logrus.WithError(err).Error("could not reboot gateway")
		return fmt.Errorf("could not reboot gateway: %w", err)
	}
	return nil
}

func Cmd(version string) { //nolint:funlen
	var configFile string

	home, err := os.UserHomeDir()
	defaultConfig := ".tmhi-cli.toml"
	if err == nil {
		defaultConfig = home + "/" + defaultConfig
	}
	configSource := altsrc.NewStringPtrSourcer(&configFile)

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        ConfigConfig,
			Aliases:     []string{"c"},
			Usage:       "use the specified TOML configuration file",
			Value:       defaultConfig,
			Destination: &configFile,
			TakesFile:   true,
		},
		&cli.BoolFlag{
			Name:    ConfigDebug,
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "display debugging output in the console",
		},
		&cli.BoolFlag{
			Name:    ConfigDryRun,
			Aliases: []string{"D"},
			Value:   false,
			Usage:   "do not perform any change to the gateway",
		},
		&cli.StringFlag{
			Name:     ConfigModel,
			Sources:  cli.NewValueSourceChain(toml.TOML(ConfigModel, configSource)),
			Required: true,
			Usage:    fmt.Sprintf("gateway model: options: %s, %s", ARCADYAN, NOK5G21),
		},
		&cli.StringFlag{
			Name:    ConfigIP,
			Sources: cli.NewValueSourceChain(toml.TOML(ConfigIP, configSource)),
			Value:   "192.168.12.1",
			Usage:   "gateway IP",
		},
		&cli.StringFlag{
			Name:    ConfigUsername,
			Sources: cli.NewValueSourceChain(toml.TOML(ConfigUsername, configSource)),
			Value:   "admin",
			Usage:   "admin username",
		},
		&cli.StringFlag{
			Name:     ConfigPassword,
			Sources:  cli.NewValueSourceChain(toml.TOML(ConfigPassword, configSource)),
			Required: false,
			Usage:    "admin password",
		},
		&cli.IntFlag{
			Name:    ConfigRetries,
			Sources: cli.NewValueSourceChain(toml.TOML(ConfigRetries, configSource)),
			Value:   0,
			Usage:   "number of retries",
		},
		&cli.DurationFlag{
			Name:    ConfigTimeout,
			Sources: cli.NewValueSourceChain(toml.TOML(ConfigRetries, configSource)),
			Value:   DefaultTimeout,
			Usage:   "request timeout in seconds",
		},
	}
	commands := []*cli.Command{
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

	app := &cli.Command{
		Name:     "tmhi-cli",
		Usage:    "Utility to interact with T-Mobile Home Internet gateway",
		Version:  version,
		Flags:    flags,
		Commands: commands,
		Before:   commonContext,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		logrus.WithError(err).Fatal()
	}
}
