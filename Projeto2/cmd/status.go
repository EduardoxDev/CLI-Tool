package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/user/forge/internal/config"
	"github.com/user/forge/internal/ui"
)

var (
	statusEnv     string
	statusService string
	statusWatch   bool
	statusFormat  string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show health status of services",
	Long: `Show health status of services across environments.

Examples:
  forge status
  forge status --env production
  forge status --env staging --service api
  forge status --watch`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if statusWatch {
			return watchStatus()
		}
		return showStatus()
	},
}

func showStatus() error {
	envNames := Cfg.EnvNames()
	if statusEnv != "" {
		if _, err := Cfg.GetEnvironment(statusEnv); err != nil {
			return err
		}
		envNames = []string{statusEnv}
	}

	for _, envName := range envNames {
		env := Cfg.Environments[envName]
		printEnvStatus(envName, env)
	}
	return nil
}

func printEnvStatus(envName string, env config.Environment) {
	ui.Section(fmt.Sprintf("Environment: %s", color.New(color.FgCyan, color.Bold).Sprint(envName)))
	ui.KeyValue("Host:", env.Host)
	ui.KeyValue("Deploy path:", env.DeployPath)
	fmt.Println()

	tbl := table.New("  SERVICE", "REPLICAS", "PORT", "STATUS", "HEALTH URL")
	tbl.WithHeaderFormatter(color.New(color.FgBlue, color.Bold).SprintfFunc())
	tbl.WithFirstColumnFormatter(color.New(color.Bold).SprintfFunc())

	for _, svc := range env.Services {
		if statusService != "" && svc.Name != statusService {
			continue
		}

		status := checkHealth(svc.HealthURL)
		healthDisplay := svc.HealthURL
		if healthDisplay == "" {
			healthDisplay = color.New(color.Faint).Sprint("—")
		}
		port := fmt.Sprintf("%d", svc.Port)
		if svc.Port == 0 {
			port = color.New(color.Faint).Sprint("—")
		}

		tbl.AddRow(
			"  "+svc.Name,
			fmt.Sprintf("%d", svc.Replicas),
			port,
			ui.StatusBadge(status),
			healthDisplay,
		)
	}

	tbl.Print()
	fmt.Println()
}

func checkHealth(healthURL string) string {
	if healthURL == "" {
		return "unknown"
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "healthy"
	}
	return "unhealthy"
}

func watchStatus() error {
	ui.Warn("Watch mode: refreshing every 5s — Ctrl+C to exit")
	for {
		fmt.Print("\033[H\033[2J") // clear terminal
		ui.Banner("forge status — " + time.Now().Format("15:04:05"))
		if err := showStatus(); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}

func printStatusHeader() {
	fmt.Println()
	parts := []string{"SERVICE", "REPLICAS", "PORT", "STATUS"}
	header := "  " + strings.Join(parts, "   ")
	color.New(color.FgBlue, color.Bold).Println(header)
	color.New(color.Faint).Println("  " + strings.Repeat("─", 60))
}

func init() {
	statusCmd.Flags().StringVarP(&statusEnv, "env", "e", "", "filter by environment")
	statusCmd.Flags().StringVar(&statusService, "service", "", "filter by service name")
	statusCmd.Flags().BoolVarP(&statusWatch, "watch", "w", false, "watch mode (refresh every 5s)")
	statusCmd.Flags().StringVar(&statusFormat, "format", "table", "output format: table, json")
	rootCmd.AddCommand(statusCmd)
}
