package cmd

import (
	"fmt"
	"os"

	"github.com/hugoh/tmhi-cli/internal"
	"github.com/hugoh/tmhi-cli/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug   bool   //nolint:gochecknoglobals
	dryRun  bool   //nolint:gochecknoglobals
	cfgFile string //nolint:gochecknoglobals
)

func getGateway() (pkg.GatewayI, error) { //nolint:ireturn //FIXME:
	conf, err := internal.ReadConf(cfgFile)
	internal.LogSetup(debug)
	internal.FatalIfError(err)
	gateway, errNew := pkg.NewGateway(conf.Gateway.Model, conf.Login.Username,
		conf.Login.Password, conf.Gateway.IP, dryRun)
	if errNew != nil {
		return gateway, fmt.Errorf("error getting gateway interface: %w", errNew)
	}
	return gateway, nil
}

func Execute(version string) {
	rootCmd := &cobra.Command{
		Use:     "tmhi-cli",
		Version: version,
		Short:   "A brief description of your application",
	}
	rootCmd.PersistentFlags().
		StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.tmhi-cli.yaml)")
	rootCmd.PersistentFlags().
		BoolVarP(&dryRun, "dry-run", "D", false, "don't perform any change to the gateway")
	rootCmd.PersistentFlags().
		BoolVarP(&debug, "debug", "d", false, "display debugging output in the console")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "Verify that the credentials can log the tool in",
		Args:  cobra.ExactArgs(0),
		Run: func(_ *cobra.Command, _ []string) {
			gateway, err := getGateway()
			internal.FatalIfError(err)
			loginErr := gateway.Login()
			internal.FatalIfError(loginErr)
			logrus.Info("Successfully logged in")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "reboot",
		Short: "Reboot the router",
		Args:  cobra.ExactArgs(0),
		Run: func(_ *cobra.Command, _ []string) {
			gateway, err := getGateway()
			internal.FatalIfError(err)
			rebootErr := gateway.Reboot()
			internal.FatalIfError(rebootErr)
		},
	})

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
