package cmd

import (
	"github.com/spf13/cobra"
	"github.com/user/forge/internal/deployer"
)

var (
	rollbackEnv     string
	rollbackVersion string
	rollbackSteps   int
	rollbackDryRun  bool
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [service]",
	Short: "Roll back a service to a previous version",
	Long: `Roll back a service to a previous version.

If --version is not specified, forge looks up the deployment history
and rolls back to the previous successful deployment.

Examples:
  forge rollback api --env production
  forge rollback api --env production --version v1.3.0
  forge rollback api --env staging --steps 2
  forge rollback --env production --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := ""
		if len(args) > 0 {
			service = args[0]
		}

		return deployer.Rollback(deployer.RollbackOptions{
			Env:     rollbackEnv,
			Service: service,
			Version: rollbackVersion,
			Steps:   rollbackSteps,
			DryRun:  rollbackDryRun,
		}, Cfg, Hist)
	},
}

func init() {
	rollbackCmd.Flags().StringVarP(&rollbackEnv, "env", "e", "", "target environment (required)")
	rollbackCmd.Flags().StringVar(&rollbackVersion, "version", "", "roll back to this specific version")
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 1, "number of versions to roll back")
	rollbackCmd.Flags().BoolVar(&rollbackDryRun, "dry-run", false, "show what would be rolled back")
	rollbackCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(rollbackCmd)
}
