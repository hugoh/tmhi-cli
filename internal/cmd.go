package internal

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
)

type Globals struct {
	Config  string           `help:"config file"                             short:"c"      type:"existingfile"`
	Debug   bool             `help:"display debugging output in the console" short:"d"`
	DryRun  bool             `help:"dump parsed configuration and quit"      name:"dry-run" short:"D"`
	Version kong.VersionFlag `help:"print version information"`
}

type CLI struct {
	Globals

	Login  LoginCmd  `cmd:"" help:"Verify that the credentials can log the tool in"`
	Reboot RebootCmd `cmd:"" help:"Reboot the router"`
}

type LoginCmd struct{}

type RebootCmd struct{}

func getGateway(globals *Globals) pkg.GatewayI { //nolint:ireturn //FIXME:
	conf := ReadConf(globals.Config)
	LogSetup(globals.Debug)
	gateway, err := pkg.NewGateway(conf.Gateway.Model, conf.Login.Username,
		conf.Login.Password, conf.Gateway.IP, globals.DryRun)
	if err != nil {
		logrus.Fatal("Error getting gateway interface")
	}
	return gateway
}

func (l *LoginCmd) Run(globals *Globals) error {
	gateway := getGateway(globals)
	err := gateway.Login()
	if err != nil {
		logrus.WithError(err).Fatal("could not log in")
	} else {
		logrus.Info("successfully logged in")
	}
	return fmt.Errorf("login failed: %w", err)
}

func (r *RebootCmd) Run(globals *Globals) error {
	gateway := getGateway(globals)
	err := gateway.Reboot()
	if err != nil {
		logrus.WithError(err).Fatal("Could not reboot gateway")
	}
	return fmt.Errorf("reboot failed: %w", err)
}

func Cmd(version string) {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("tmhi-cli"),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		})

	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
