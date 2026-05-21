package cmd

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/forge/internal/ui"
)

var (
	logsEnv    string
	logsTail   int
	logsFollow bool
	logsSince  string
	logsFilter string
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View logs for a service",
	Long: `View logs for a service in the given environment.

Examples:
  forge logs api --env dev
  forge logs api --env staging --tail 100
  forge logs api --env dev --follow
  forge logs api --env dev --filter "ERROR"
  forge logs api --env dev --since 1h`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := "api"
		if len(args) > 0 {
			service = args[0]
		}

		env, err := Cfg.GetEnvironment(logsEnv)
		if err != nil {
			return err
		}

		svc, err := env.GetService(service)
		if err != nil {
			return err
		}

		ui.Section(fmt.Sprintf("Logs: %s @ %s", color.New(color.FgCyan, color.Bold).Sprint(svc.Name), logsEnv))
		if logsSince != "" {
			ui.Dim(fmt.Sprintf("Showing logs since: %s", logsSince))
		}
		ui.Dim(fmt.Sprintf("Showing last %d lines", logsTail))
		fmt.Println()

		lines := generateLogs(svc.Name, logsEnv, logsTail)
		for _, line := range lines {
			if logsFilter != "" && !strings.Contains(line.msg, logsFilter) {
				continue
			}
			printLogLine(line)
		}

		if logsFollow {
			fmt.Println()
			color.New(color.Faint).Println("  Following logs — Ctrl+C to exit")
			fmt.Println()
			followLogs(svc.Name, logsEnv)
		}

		return nil
	},
}

type logLine struct {
	ts    time.Time
	level string
	msg   string
	extra map[string]string
}

func printLogLine(l logLine) {
	ts := color.New(color.Faint).Sprint(l.ts.Format("2006-01-02T15:04:05.000Z07:00"))
	var levelColor *color.Color
	switch l.level {
	case "ERROR":
		levelColor = color.New(color.FgRed, color.Bold)
	case "WARN":
		levelColor = color.New(color.FgYellow)
	case "DEBUG":
		levelColor = color.New(color.Faint)
	default:
		levelColor = color.New(color.FgCyan)
	}
	level := levelColor.Sprintf("%-5s", l.level)

	extras := ""
	for k, v := range l.extra {
		extras += color.New(color.Faint).Sprintf(" %s=", k) + color.New(color.FgGreen).Sprint(v)
	}

	fmt.Printf("  %s  %s  %s%s\n", ts, level, l.msg, extras)
}

func generateLogs(service, env string, n int) []logLine {
	messages := []struct {
		level string
		msg   string
		extra map[string]string
	}{
		{"INFO", "Server started", map[string]string{"port": "8080", "env": env}},
		{"INFO", "Connected to database", map[string]string{"host": "db:5432", "pool": "10"}},
		{"INFO", "Redis connection established", map[string]string{"host": "redis:6379"}},
		{"INFO", "GET /health", map[string]string{"status": "200", "latency": "1ms"}},
		{"INFO", "GET /api/users", map[string]string{"status": "200", "latency": "12ms"}},
		{"INFO", "POST /api/users", map[string]string{"status": "201", "latency": "45ms"}},
		{"INFO", "GET /api/orders", map[string]string{"status": "200", "latency": "8ms"}},
		{"DEBUG", "Cache hit", map[string]string{"key": "user:42", "ttl": "299s"}},
		{"DEBUG", "Cache miss", map[string]string{"key": "user:99"}},
		{"WARN", "Slow query detected", map[string]string{"query": "SELECT *", "duration": "234ms"}},
		{"WARN", "Rate limit approaching", map[string]string{"client": "192.168.1.5", "remaining": "12"}},
		{"INFO", "Background job started", map[string]string{"job": "email-digest", "id": "job-8f2a"}},
		{"INFO", "Background job completed", map[string]string{"job": "email-digest", "duration": "1.2s"}},
		{"ERROR", "Failed to send email", map[string]string{"to": "user@example.com", "error": "timeout"}},
		{"INFO", "GET /api/products", map[string]string{"status": "200", "latency": "22ms"}},
		{"INFO", "DELETE /api/sessions/abc123", map[string]string{"status": "204", "latency": "5ms"}},
		{"WARN", "Memory usage high", map[string]string{"used": "82%", "threshold": "80%"}},
		{"INFO", "Config reloaded", map[string]string{"source": ".env"}},
		{"DEBUG", "Middleware: auth", map[string]string{"user": "42", "role": "admin"}},
		{"INFO", "GET /metrics", map[string]string{"status": "200", "latency": "3ms"}},
	}

	t := time.Now().Add(-time.Duration(n) * 3 * time.Second)
	var lines []logLine
	for i := 0; i < n; i++ {
		m := messages[rand.Intn(len(messages))]
		lines = append(lines, logLine{
			ts:    t.Add(time.Duration(i*3+rand.Intn(3)) * time.Second),
			level: m.level,
			msg:   m.msg,
			extra: m.extra,
		})
	}
	return lines
}

func followLogs(service, env string) {
	for {
		time.Sleep(time.Duration(300+rand.Intn(1200)) * time.Millisecond)
		msgs := generateLogs(service, env, 1)
		for _, l := range msgs {
			l.ts = time.Now()
			printLogLine(l)
		}
	}
}

func init() {
	logsCmd.Flags().StringVarP(&logsEnv, "env", "e", "", "target environment (required)")
	logsCmd.Flags().IntVarP(&logsTail, "tail", "n", 50, "number of log lines to show")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "show logs since duration (e.g. 1h, 30m)")
	logsCmd.Flags().StringVar(&logsFilter, "filter", "", "filter log lines by string")
	logsCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(logsCmd)
}
