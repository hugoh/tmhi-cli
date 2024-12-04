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

func getGateway(cCtx *cli.Context) pkg.GatewayI { //nolint:ireturn //FIXME:
	LogSetup(cCtx.Bool("debug"))
	gateway, err := pkg.NewGateway(cCtx.String("gateway.model"), cCtx.String("login.username"),
		cCtx.String("login.password"), cCtx.String("gateway.ip"), cCtx.Bool("dry-run"))
	if err != nil {
		logrus.Fatal("Error getting gateway interface")
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
	return fmt.Errorf("login failed: %w", err)
}

func Reboot(cCtx *cli.Context) error {
	gateway := getGateway(cCtx)
	err := gateway.Reboot()
	if err != nil {
		logrus.WithError(err).Fatal("Could not reboot gateway")
	}
	return fmt.Errorf("reboot failed: %w", err)
}

func Cmd(version string) { //nolint:funlen
	// FIXME: altsrc and Required don't play well: https://github.com/urfave/cli/issues/1725
	flags := []cli.Flag{
		&cli.PathFlag{
			Name:      "config",
			Aliases:   []string{"c"},
			Usage:     "use the specified YAML configuration file",
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "display debugging output in the console",
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"D"},
			Value:   false,
			Usage:   "do not perform any change to the gateway",
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: "gateway.model",
			// Required: true,
			Usage: "gateway model",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "gateway.ip",
			Value: "192.168.12.1",
			// Required: true,
			Usage: "gateway IP",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "login.username",
			Value: "admin",
			// Required: true,
			Usage: "admin username",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name: "login.password",
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
		Version:  version,
		Flags:    flags,
		Commands: commands,
		Before: altsrc.InitInputSourceWithContext(flags,
			func(context *cli.Context) (altsrc.InputSourceContext, error) {
				if context.IsSet("config") {
					filePath := context.String("config")
					return altsrc.NewYamlSourceFromFile(filePath)
				}
				return &altsrc.MapInputSource{}, nil
			}),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
