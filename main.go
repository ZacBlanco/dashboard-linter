package main

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/grafana/dashboard-linter/lint"
)

var lintStrictFlag bool
var lintVerboseFlag bool
var lintConfigFlag string

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint [dashboard.json]",
	Short: "Lint a dashboard",
	Long:  `Returns warnings or errors for dashboard which do not adhere to accepted standards`,
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = viper.BindPFlags(cmd.PersistentFlags())
	},
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		buf, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", filename, err)
		}

		dashboard, err := lint.NewDashboard(buf)
		if err != nil {
			return fmt.Errorf("failed to parse dashboard: %v", err)
		}

		// if no config flag was passed, set a default path of a .lint file in the dashboards directory
		if lintConfigFlag == "" {
			lintConfigFlag = path.Join(path.Dir(filename), ".lint")
		}

		config := lint.NewConfigurationFile()
		if err := config.Load(lintConfigFlag); err != nil {
			return fmt.Errorf("failed to load lint config: %v", err)
		}
		config.Verbose = lintVerboseFlag

		rules := lint.NewRuleSet()
		results, err := rules.Lint([]lint.Dashboard{dashboard})
		if err != nil {
			return fmt.Errorf("failed to lint dashboard: %v", err)
		}

		results.Configure(config)
		results.ReportByRule()

		if lintStrictFlag && results.MaximumSeverity() >= lint.Warning {
			return fmt.Errorf("there were linting errors, please see previous output")
		}
		return nil
	},
}

var rulesCmd = &cobra.Command{
	Use:          "rules",
	Short:        "Print documentation about each lint rule.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		rules := lint.NewRuleSet()
		for _, rule := range rules.Rules() {
			fmt.Fprintf(os.Stdout, "* `%s` - %s\n", rule.Name(), rule.Description())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(rulesCmd)
	lintCmd.Flags().BoolVar(
		&lintStrictFlag,
		"strict",
		false,
		"fail upon linting error or warning",
	)
	lintCmd.Flags().BoolVar(
		&lintVerboseFlag,
		"verbose",
		false,
		"show more information about linting",
	)
	lintCmd.Flags().StringVarP(
		&lintConfigFlag,
		"config",
		"c",
		"",
		"path to a configuration file",
	)
}

var rootCmd = &cobra.Command{
	Use:   "dashboard-linter",
	Short: "A command-line application to lint Grafana dashboards.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
