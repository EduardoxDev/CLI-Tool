package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	Version   = "0.1.0"
	BuildDate = "2026-05-20"
)

var versionCmd = &cobra.Command{
	Use:         "version",
	Short:       "Print forge version",
	Annotations: map[string]string{"skip-config": "true"},
	Run: func(cmd *cobra.Command, args []string) {
		bold := color.New(color.FgCyan, color.Bold)
		bold.Printf("forge")
		fmt.Printf(" version %s (built %s)\n", Version, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
