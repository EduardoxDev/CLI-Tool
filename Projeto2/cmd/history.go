package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/user/forge/internal/ui"
)

var (
	histEnv     string
	histService string
	histLimit   int
	histFormat  string
	histClear   bool
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View deployment history",
	Long: `View the history of all deployments.

Examples:
  forge history
  forge history --env production
  forge history --service api --limit 10
  forge history --format json
  forge history --env staging --clear`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if histClear {
			return clearHistory()
		}
		return showHistory()
	},
}

func showHistory() error {
	deploys := Hist.List(histEnv, histService, histLimit)

	if len(deploys) == 0 {
		ui.Warn("No deployments found")
		if histEnv != "" {
			ui.Dim(fmt.Sprintf("No deployments for environment %q", histEnv))
		}
		fmt.Println()
		return nil
	}

	if histFormat == "json" {
		data, err := json.MarshalIndent(deploys, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	title := "Deployment History"
	if histEnv != "" {
		title += " — " + histEnv
	}
	if histService != "" {
		title += " / " + histService
	}
	ui.Section(title)

	tbl := table.New("  ID", "SERVICE", "ENV", "VERSION", "STRATEGY", "STATUS", "DURATION", "DEPLOYED BY", "DATE")
	tbl.WithHeaderFormatter(color.New(color.FgBlue, color.Bold).SprintfFunc())
	tbl.WithFirstColumnFormatter(color.New(color.Faint).SprintfFunc())

	for _, d := range deploys {
		strategy := d.Strategy
		if strategy == "" {
			strategy = "rolling"
		}

		duration := ""
		if d.Duration > 0 {
			duration = ui.FormatDuration(d.Duration)
		}

		tbl.AddRow(
			"  "+truncate(d.ID, 16),
			d.Service,
			d.Env,
			color.CyanString(d.Version),
			strategy,
			ui.StatusBadge(d.Status),
			duration,
			d.DeployedBy,
			d.StartedAt.Format("01/02 15:04"),
		)
	}

	tbl.Print()
	fmt.Println()
	fmt.Printf("  Showing %s of %s total deployments\n",
		color.CyanString("%d", len(deploys)),
		color.CyanString("%d", len(Hist.List("", "", 0))),
	)
	fmt.Println()
	return nil
}

func clearHistory() error {
	count := len(Hist.List(histEnv, histService, 0))
	if count == 0 {
		ui.Warn("No deployments to clear")
		return nil
	}

	scope := "all deployments"
	if histEnv != "" || histService != "" {
		parts := []string{}
		if histEnv != "" {
			parts = append(parts, "env="+histEnv)
		}
		if histService != "" {
			parts = append(parts, "service="+histService)
		}
		scope = strings.Join(parts, ", ")
	}

	if err := Hist.Clear(histEnv, histService); err != nil {
		return fmt.Errorf("clearing history: %w", err)
	}

	ui.Success(fmt.Sprintf("Cleared %d deployments (%s)", count, scope))
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func init() {
	historyCmd.Flags().StringVarP(&histEnv, "env", "e", "", "filter by environment")
	historyCmd.Flags().StringVar(&histService, "service", "", "filter by service")
	historyCmd.Flags().IntVarP(&histLimit, "limit", "n", 20, "number of deployments to show")
	historyCmd.Flags().StringVar(&histFormat, "format", "table", "output format: table, json")
	historyCmd.Flags().BoolVar(&histClear, "clear", false, "clear deployment history")
	rootCmd.AddCommand(historyCmd)
}
