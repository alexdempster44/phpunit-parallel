package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alexdempster44/phpunit-parallel/internal/config"
	"github.com/alexdempster44/phpunit-parallel/internal/output"
	"github.com/alexdempster44/phpunit-parallel/internal/output/tui"
	"github.com/alexdempster44/phpunit-parallel/internal/runner"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const defaultRunnerConfigFile = "phpunit-parallel.xml"

var (
	configFile       string
	runnerConfigFile string
	teamcity         bool
	shardFlag        string
	runnerConfig     = config.DefaultRunner()
	versionInfo      string
)

var rootCmd = &cobra.Command{
	Use:          "phpunit-parallel",
	Short:        "Run PHPUnit tests in parallel",
	SilenceUsage: true,
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

		if v, ok := os.LookupEnv("PHPUNIT_PARALLEL_WORKERS"); ok {
			w, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("invalid PHPUNIT_PARALLEL_WORKERS value: %w", err)
			}
			runnerConfig.Workers = w
		}

		if cmd.Flags().Changed("workers") {
			runnerConfig.Workers, _ = cmd.Flags().GetInt("workers")
		}
		if cmd.Flags().Changed("config-build-dir") {
			runnerConfig.ConfigBuildDir, _ = cmd.Flags().GetString("config-build-dir")
		}
		if cmd.Flags().Changed("before") {
			runnerConfig.Before, _ = cmd.Flags().GetString("before")
		}
		if cmd.Flags().Changed("before-worker") {
			runnerConfig.BeforeWorker, _ = cmd.Flags().GetString("before-worker")
		}
		if cmd.Flags().Changed("run-worker") {
			runnerConfig.RunWorker, _ = cmd.Flags().GetString("run-worker")
		}
		if cmd.Flags().Changed("after-worker") {
			runnerConfig.AfterWorker, _ = cmd.Flags().GetString("after-worker")
		}
		if cmd.Flags().Changed("after") {
			runnerConfig.After, _ = cmd.Flags().GetString("after")
		}
		if cmd.Flags().Changed("filter") {
			runnerConfig.Filter, _ = cmd.Flags().GetString("filter")
		}
		if cmd.Flags().Changed("test-suffix") {
			runnerConfig.TestSuffix, _ = cmd.Flags().GetString("test-suffix")
		}
		if cmd.Flags().Changed("group") {
			runnerConfig.Group, _ = cmd.Flags().GetString("group")
		}
		if cmd.Flags().Changed("exclude-group") {
			runnerConfig.ExcludeGroup, _ = cmd.Flags().GetString("exclude-group")
		}
		if cmd.Flags().Changed("shard") {
			idx, total, err := parseShard(shardFlag)
			if err != nil {
				return fmt.Errorf("invalid --shard value: %w", err)
			}
			runnerConfig.ShardIndex = idx
			runnerConfig.ShardTotal = total
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("configuration") {
			if runnerConfig.Configuration != "" {
				configFile = runnerConfig.Configuration
			} else if _, err := os.Stat("phpunit.xml"); err != nil {
				if _, err := os.Stat("phpunit.xml.dist"); err == nil {
					configFile = "phpunit.xml.dist"
				}
			}
		}

		cfg, err := config.ParsePHPUnit(configFile)
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		baseDir := filepath.Dir(configFile)
		if baseDir == "." {
			baseDir, _ = os.Getwd()
		}

		var out output.Output
		if teamcity {
			out = output.NewTeamCityOutput()
		} else if !term.IsTerminal(int(os.Stdout.Fd())) {
			out = output.NewTerminalOutput()
		} else {
			out = tui.New()
		}

		r := runner.New(cfg, runnerConfig, baseDir, out)
		r.Version = versionInfo
		return r.Run()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "configuration", "c", "phpunit.xml", "PHPUnit configuration file")
	rootCmd.Flags().BoolVar(&teamcity, "teamcity", false, "Output in TeamCity format")

	rootCmd.Flags().StringVar(&runnerConfigFile, "runner-config", "", "Runner configuration file")
	rootCmd.Flags().IntVarP(&runnerConfig.Workers, "workers", "w", runnerConfig.Workers, "Number of parallel workers")
	rootCmd.Flags().StringVar(&runnerConfig.ConfigBuildDir, "config-build-dir", runnerConfig.ConfigBuildDir, "Directory for generated config files")
	rootCmd.Flags().StringVar(&runnerConfig.Before, "before", "", "Command to run once before all workers start")
	rootCmd.Flags().StringVar(&runnerConfig.BeforeWorker, "before-worker", "", "Command to run before each worker starts")
	rootCmd.Flags().StringVar(&runnerConfig.RunWorker, "run-worker", runnerConfig.RunWorker, "Command to run PHPUnit for each worker")
	rootCmd.Flags().StringVar(&runnerConfig.AfterWorker, "after-worker", "", "Command to run after each worker completes")
	rootCmd.Flags().StringVar(&runnerConfig.After, "after", "", "Command to run once after all workers complete")
	rootCmd.Flags().StringVar(&runnerConfig.Filter, "filter", "", "Filter which tests to run (passed to PHPUnit --filter)")
	rootCmd.Flags().StringVar(&runnerConfig.TestSuffix, "test-suffix", runnerConfig.TestSuffix, "Suffix for test files")
	rootCmd.Flags().StringVar(&runnerConfig.Group, "group", "", "Only run tests from the specified group(s)")
	rootCmd.Flags().StringVar(&runnerConfig.ExcludeGroup, "exclude-group", "", "Exclude tests from the specified group(s)")
	rootCmd.Flags().StringVar(&shardFlag, "shard", "", "Run only this shard's slice of tests, in the format X/N (e.g. 2/4). Combine with a CI matrix to fan tests across runners.")
}

func parseShard(s string) (int, int, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format X/N, got %q", s)
	}
	idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid shard index: %w", err)
	}
	total, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid shard total: %w", err)
	}
	if total < 1 {
		return 0, 0, fmt.Errorf("shard total must be >= 1, got %d", total)
	}
	if idx < 1 || idx > total {
		return 0, 0, fmt.Errorf("shard index %d out of range [1..%d]", idx, total)
	}
	return idx, total, nil
}

func SetVersionInfo(version string) {
	versionInfo = version
	rootCmd.Version = version
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
