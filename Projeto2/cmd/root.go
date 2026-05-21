package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/forge/internal/config"
	"github.com/user/forge/internal/history"
)

var (
	cfgFile string
	verbose bool
	noColor bool

	Cfg  *config.Config
	Hist *history.History
)

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "forge — Powerful deployment CLI",
	Long: `forge is a CLI tool for managing deployments across environments.

It supports rolling, blue-green, and recreate strategies with
persistent history, health checks, and per-environment config.

Commands:
  forge init                           Initialize .forge.yaml
  forge deploy [service] --env <env>   Deploy to an environment
  forge rollback [service] --env <env> Roll back to a previous version
  forge status [--env <env>]           Check service health
  forge logs [service] --env <env>     View service logs
  forge env list --env <env>           Manage environment variables
  forge history [--env <env>]          View deployment history

Run 'forge <command> --help' for more information.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if noColor {
			color.NoColor = true
		}

		if cmd.Annotations["skip-config"] == "true" {
			return nil
		}

		var err error
		Cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("config not loaded: %w\n\nRun 'forge init' to create a .forge.yaml", err)
		}

		Hist, err = history.Load()
		if err != nil {
			Hist = &history.History{}
		}
		return nil
	},
}

func Execute() {
	cobra.EnableTraverseRunHooks = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s %s\n\n", color.New(color.FgRed, color.Bold).Sprint("Error:"), err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .forge.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
}
