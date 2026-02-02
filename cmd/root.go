package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	teamcity   bool
)

var rootCmd = &cobra.Command{
	Use:   "phpunit-parallel",
	Short: "Run PHPUnit tests in parallel",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: implement
	},
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "configuration", "c", "phpunit.xml", "PHPUnit configuration file")
	rootCmd.Flags().BoolVar(&teamcity, "teamcity", false, "Output in TeamCity format")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
