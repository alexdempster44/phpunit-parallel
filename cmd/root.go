package cmd

import (
	"fmt"
	"os"

	"github.com/alexdempster44/phpunit-parallel/internal/config"
	"github.com/spf13/cobra"
)

const defaultRunnerConfigFile = "phpunit-parallel.xml"

var (
	configFile       string
	runnerConfigFile string
	teamcity         bool
	runnerConfig     = config.DefaultRunner()
)

var rootCmd = &cobra.Command{
	Use:   "phpunit-parallel",
	Short: "Run PHPUnit tests in parallel",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		configToLoad := runnerConfigFile
		if configToLoad == "" {
			if _, err := os.Stat(defaultRunnerConfigFile); err == nil {
				configToLoad = defaultRunnerConfigFile
			}
		}

		if configToLoad != "" {
			cfg, err := config.ParseRunner(configToLoad)
			if err != nil {
				return fmt.Errorf("failed to parse runner config: %w", err)
			}
			runnerConfig = cfg
		}

		if cmd.Flags().Changed("workers") {
			runnerConfig.Workers, _ = cmd.Flags().GetInt("workers")
		}
		if cmd.Flags().Changed("config-build-dir") {
			runnerConfig.ConfigBuildDir, _ = cmd.Flags().GetString("config-build-dir")
		}
		if cmd.Flags().Changed("run-command") {
			runnerConfig.RunCommand, _ = cmd.Flags().GetString("run-command")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ParsePHPUnit(configFile)
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		fmt.Printf("\n")
		fmt.Printf("Runner Config:\n")
		fmt.Printf("  Workers: %d\n", runnerConfig.Workers)
		fmt.Printf("  Config build dir: %s\n", runnerConfig.ConfigBuildDir)
		fmt.Printf("  Run command: %s\n", runnerConfig.RunCommand)
		fmt.Printf("\n")
		fmt.Printf("Bootstrap: %s\n", cfg.Bootstrap)
		fmt.Printf("Test Suites:\n")
		for _, suite := range cfg.TestSuites.TestSuites {
			fmt.Printf("  - %s\n", suite.Name)
			for _, dir := range suite.Directories {
				fmt.Printf("      path: %s\n", dir)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "configuration", "c", "phpunit.xml", "PHPUnit configuration file")
	rootCmd.Flags().BoolVar(&teamcity, "teamcity", false, "Output in TeamCity format")

	rootCmd.Flags().StringVar(&runnerConfigFile, "runner-config", "", "Runner configuration file")
	rootCmd.Flags().IntVarP(&runnerConfig.Workers, "workers", "w", runnerConfig.Workers, "Number of parallel workers")
	rootCmd.Flags().StringVar(&runnerConfig.ConfigBuildDir, "config-build-dir", runnerConfig.ConfigBuildDir, "Directory for generated config files")
	rootCmd.Flags().StringVar(&runnerConfig.RunCommand, "run-command", runnerConfig.RunCommand, "Command to run PHPUnit")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
