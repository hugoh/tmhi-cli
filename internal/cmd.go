package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

const (
	NOK5G21 string = "NOK5G21"
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

func getGateway(cCtx *cli.Context) pkg.GatewayI { //nolint:ireturn
	LogSetup(cCtx.Bool(ConfigDebug))
	model := cCtx.String(ConfigModel)
	var gateway pkg.GatewayI
	switch model {
	case "NOK5G21":
		gateway = pkg.NewNokiaGateway(cCtx.String(ConfigUsername), cCtx.String(ConfigPassword),
			cCtx.String(ConfigIP))
	default:
		logrus.WithField("gateway", model).Fatal("unsupported gateway")
	}
	return gateway
}

func Login(cCtx *cli.Context) error {
	gateway := getGateway(cCtx)
	err := gateway.Login()
	if err != nil {
		logrus.WithError(err).Fatal("could not log in")
	} else {
		logrus.Info("successfully logged in")
	}
	return nil
}

func Reboot(cCtx *cli.Context) error {
	gateway := getGateway(cCtx)
	err := gateway.Reboot(cCtx.Bool(ConfigDryRun))
	if err != nil {
		logrus.WithError(err).Fatal("Could not reboot gateway")
	}
	return fmt.Errorf("reboot failed: %w", err)
}

func Cmd(version string) { //nolint:funlen
	// FIXME: altsrc and Required don't play well: https://github.com/urfave/cli/issues/1725
	flags := []cli.Flag{
		&cli.PathFlag{
			Name:      ConfigConfig,
			Aliases:   []string{"c"},
			Usage:     "use the specified TOML configuration file",
			TakesFile: true,
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
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: ConfigModel,
			// Required: true,
			Usage: "gateway model: options: " + NOK5G21,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  ConfigIP,
			Value: "192.168.12.1",
			// Required: true,
			Usage: "gateway IP",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  ConfigUsername,
			Value: "admin",
			// Required: true,
			Usage: "admin username",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: ConfigPassword,
			// Required: true,
			Usage: "admin password",
		}),
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

	app := &cli.App{
		Name:     "tmhi-cli",
		Usage:    "Utility to interact with T-Mobile Home Internet gateway",
		Version:  version,
		Flags:    flags,
		Commands: commands,
		Before: altsrc.InitInputSourceWithContext(flags,
			func(context *cli.Context) (altsrc.InputSourceContext, error) {
				if context.IsSet(ConfigConfig) {
					filePath := context.String(ConfigConfig)
					return altsrc.NewTomlSourceFromFile(filePath)
				}
				return &altsrc.MapInputSource{}, nil
			}),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
