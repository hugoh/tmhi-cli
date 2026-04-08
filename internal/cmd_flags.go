package internal

import (
	"fmt"

	altsrc "github.com/urfave/cli-altsrc/v3"
	toml "github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
)

func buildFlagsBaseCore(configFile *string, configSource altsrc.Sourcer) []cli.Flag {
	partA := buildFlagsBaseCorePartA(configFile, configSource)
	partB := buildFlagsBaseCorePartB(configFile, configSource)

	return append(partA, partB...)
}

func buildFlagsBaseCorePartA(configFile *string, configSource altsrc.Sourcer) []cli.Flag {
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
			Name:    ConfigAutoConfirm,
			Aliases: []string{"y"},
			Value:   false,
			Usage:   "Skip confirmation prompts",
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
	}
}

func buildFlagsBaseCorePartB(_ *string, configSource altsrc.Sourcer) []cli.Flag {
	return []cli.Flag{
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
