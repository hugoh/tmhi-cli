/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/hugoh/thmi-cli/internal"
	"github.com/hugoh/thmi-cli/pkg"
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

// rootCmd represents the base command when called without any subcommands
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "thmi-cli",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
    examples and usage of using your application. For example:

    Cobra is a CLI library for Go that empowers applications.
    This application is a tool to generate the needed files
    to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
	rootCmd.PersistentFlags().
		StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.thmi-cli.yaml)")
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

	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
