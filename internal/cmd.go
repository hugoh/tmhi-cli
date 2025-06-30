package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
	altsrc "github.com/urfave/cli-altsrc/v3"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

const (
	ARCADYAN string = "ARCADYAN"
	NOK5G21  string = "NOK5G21"
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
)

func LogSetup(debugFlag bool) {
	if debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func getGatewayFromCtxOrFail(cCtx *cli.Command) pkg.GatewayI { //nolint:ireturn
	gateway, err := getGateway(cCtx.String(ConfigModel),
		cCtx.String(ConfigUsername),
		cCtx.String(ConfigPassword),
		cCtx.String(ConfigIP),
		cCtx.Bool(ConfigDebug))
	if err != nil {
		logrus.WithError(err).Fatal("unsupported gateway")
		// NOTREACHED
	}
	return gateway
}

func Login(_ context.Context, cCtx *cli.Command) error {
	gateway := getGatewayFromCtxOrFail(cCtx)
	err := gateway.Login()
	if err != nil {
		logrus.WithError(err).Fatal("could not log in")
	} else {
		logrus.Info("successfully logged in")
	}
	return nil
}

func Reboot(_ context.Context, cCtx *cli.Command) error {
	gateway := getGatewayFromCtxOrFail(cCtx)
	err := gateway.Reboot(cCtx.Bool(ConfigDryRun))
	if err != nil {
		logrus.WithError(err).Error("could not reboot gateway")
		return fmt.Errorf("could not reboot gateway: %w", err)
	}
	return nil
}

func Cmd(version string) { //nolint:funlen
	var configFile string
	configSource := altsrc.NewStringPtrSourcer(&configFile)

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        ConfigConfig,
			Aliases:     []string{"c"},
			Usage:       "use the specified TOML configuration file",
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
			Required: true,
			Usage:    "admin password",
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
	}

	app := &cli.Command{
		Name:     "tmhi-cli",
		Usage:    "Utility to interact with T-Mobile Home Internet gateway",
		Version:  version,
		Flags:    flags,
		Commands: commands,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		logrus.WithError(err).Fatal()
	}
}
