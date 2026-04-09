package internal

import (
	"fmt"

	altsrc "github.com/urfave/cli-altsrc/v3"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

// Configuration flag names.
const (
	ConfigAutoConfirm string = "yes"
	ConfigConfig      string = "config"
	ConfigDebug       string = "debug"
	ConfigDryRun      string = "dry-run"
	ConfigGateway     string = "gateway."
	ConfigIP          string = ConfigGateway + "ip"
	ConfigLogin       string = "login."
	ConfigModel       string = ConfigGateway + "model"
	ConfigNoColor     string = "no-color"
	ConfigPassword    string = ConfigLogin + "password"
	ConfigQuiet       string = "quiet"
	ConfigRetries     string = "retries"
	ConfigTimeout     string = "timeout"
	ConfigUsername    string = ConfigLogin + "username"
)

func cmdCommands() []*cli.Command {
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

func cmdFlags(configFile *string, configSource altsrc.Sourcer) []cli.Flag { //nolint:funlen
	return []cli.Flag{
		&cli.StringFlag{
			Name:        ConfigConfig,
			Aliases:     []string{"c"},
			Usage:       "use the specified TOML configuration file",
			Value:       defaultConfigPath(),
			Destination: configFile,
			TakesFile:   true,
		},
		&cli.BoolFlag{
			Name:    ConfigDebug,
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "display debugging output in the console",
		},
		&cli.BoolFlag{
			Name:  ConfigNoColor,
			Value: false,
			Usage: "disable colored output",
		},
		&cli.BoolFlag{
			Name:    ConfigQuiet,
			Aliases: []string{"q"},
			Value:   false,
			Usage:   "quiet mode, suppresses output",
		},
		&cli.BoolFlag{
			Name:    ConfigAutoConfirm,
			Aliases: []string{"y"},
			Value:   false,
			Usage:   "skip confirmation prompts",
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
}
