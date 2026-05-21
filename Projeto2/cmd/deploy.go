package cmd

import (
	"github.com/spf13/cobra"
	"github.com/user/forge/internal/deployer"
)

var (
	deployEnv      string
	deployVersion  string
	deployStrategy string
	deployDryRun   bool
	deployTimeout  int
)

var deployCmd = &cobra.Command{
	Use:   "deploy [service]",
	Short: "Deploy a service to an environment",
	Long: `Deploy a service (or all services) to the specified environment.

Strategies:
  rolling      Replace replicas one-by-one (zero-downtime, default)
  blue-green   Deploy to parallel environment, then switch traffic
  recreate     Stop all, then start all (expects brief downtime)

Examples:
  forge deploy --env production
  forge deploy api --env staging --version v1.4.2
  forge deploy api --env production --strategy blue-green
  forge deploy api --env staging --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := ""
		if len(args) > 0 {
			service = args[0]
		}

		return deployer.Deploy(deployer.DeployOptions{
			Env:      deployEnv,
			Service:  service,
			Version:  deployVersion,
			Strategy: deployStrategy,
			DryRun:   deployDryRun,
			Timeout:  deployTimeout,
			Verbose:  verbose,
		}, Cfg, Hist)
	},
}

func init() {
	deployCmd.Flags().StringVarP(&deployEnv, "env", "e", "", "target environment (required)")
	deployCmd.Flags().StringVar(&deployVersion, "version", "", "image tag to deploy (default: latest)")
	deployCmd.Flags().StringVar(&deployStrategy, "strategy", "", "deploy strategy: rolling, blue-green, recreate")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "show what would be deployed without doing it")
	deployCmd.Flags().IntVar(&deployTimeout, "timeout", 0, "timeout in seconds (overrides config)")
	deployCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(deployCmd)
}
