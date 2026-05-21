package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/user/forge/internal/ui"
)

var envCmdEnv string

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables per environment",
	Long: `Manage environment variables stored in .forge.yaml.

Examples:
  forge env list --env dev
  forge env get DATABASE_URL --env dev
  forge env set LOG_LEVEL debug --env staging
  forge env unset OLD_KEY --env dev`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, err := Cfg.GetEnvironment(envCmdEnv)
		if err != nil {
			return err
		}

		ui.Section(fmt.Sprintf("Environment variables: %s", color.New(color.FgCyan, color.Bold).Sprint(envCmdEnv)))

		if len(env.EnvVars) == 0 {
			ui.Dim("No variables defined")
			fmt.Println()
			return nil
		}

		tbl := table.New("  KEY", "VALUE")
		tbl.WithHeaderFormatter(color.New(color.FgBlue, color.Bold).SprintfFunc())
		tbl.WithFirstColumnFormatter(color.New(color.Bold).SprintfFunc())

		for k, v := range env.EnvVars {
			masked := maskSecret(k, v)
			tbl.AddRow("  "+k, masked)
		}
		tbl.Print()
		fmt.Println()
		fmt.Printf("  %s variables in %s\n", color.CyanString("%d", len(env.EnvVars)), envCmdEnv)
		fmt.Println()
		return nil
	},
}

var envGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get a specific environment variable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		env, err := Cfg.GetEnvironment(envCmdEnv)
		if err != nil {
			return err
		}

		val, ok := env.EnvVars[key]
		if !ok {
			return fmt.Errorf("variable %q not found in environment %q", args[0], envCmdEnv)
		}

		fmt.Println(val)
		return nil
	},
}

var envSetCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Set an environment variable",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		viperKey := fmt.Sprintf("environments.%s.env_vars.%s", envCmdEnv, key)
		viper.Set(viperKey, value)

		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		ui.Success(fmt.Sprintf("Set %s=%s in %s", color.CyanString(key), color.GreenString(value), envCmdEnv))
		return nil
	},
}

var envUnsetCmd = &cobra.Command{
	Use:   "unset KEY",
	Short: "Remove an environment variable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		env, err := Cfg.GetEnvironment(envCmdEnv)
		if err != nil {
			return err
		}

		if _, ok := env.EnvVars[key]; !ok {
			return fmt.Errorf("variable %q not found in environment %q", key, envCmdEnv)
		}

		delete(env.EnvVars, key)
		Cfg.Environments[envCmdEnv] = env

		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		ui.Success(fmt.Sprintf("Removed %s from %s", color.CyanString(key), envCmdEnv))
		return nil
	},
}

var envExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export variables as shell export statements",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, err := Cfg.GetEnvironment(envCmdEnv)
		if err != nil {
			return err
		}

		for k, v := range env.EnvVars {
			fmt.Fprintf(os.Stdout, "export %s=%q\n", k, v)
		}
		return nil
	},
}

func maskSecret(key, value string) string {
	secretKeys := []string{"PASSWORD", "SECRET", "KEY", "TOKEN", "PRIVATE", "CREDENTIALS"}
	for _, s := range secretKeys {
		if contains(key, s) {
			if len(value) <= 4 {
				return "****"
			}
			return value[:4] + "****"
		}
	}
	return value
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && containsStr(s, substr)))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func init() {
	envCmd.PersistentFlags().StringVarP(&envCmdEnv, "env", "e", "", "target environment (required)")
	envCmd.MarkPersistentFlagRequired("env")

	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(envExportCmd)
	rootCmd.AddCommand(envCmd)
}
