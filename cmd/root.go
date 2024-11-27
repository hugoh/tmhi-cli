package cmd

import (
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

func getGateway() pkg.GatewayI { //nolint:ireturn //FIXME:
	conf := internal.ReadConf(cfgFile)
	internal.LogSetup(debug)
	gateway, err := pkg.NewGateway(conf.Gateway.Model, conf.Login.Username,
		conf.Login.Password, conf.Gateway.IP, dryRun)
	if err != nil {
		logrus.Fatal("Error getting gateway interface")
	}
	return gateway
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
			gateway := getGateway()
			if err := gateway.Login(); err != nil {
				logrus.WithError(err).Fatal("could not log in")
			}
			logrus.Info("successfully logged in")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "reboot",
		Short: "Reboot the router",
		Args:  cobra.ExactArgs(0),
		Run: func(_ *cobra.Command, _ []string) {
			gateway := getGateway()
			if err := gateway.Reboot(); err != nil {
				logrus.WithError(err).Fatal("Could not reboot gateway")
			}
		},
	})

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
