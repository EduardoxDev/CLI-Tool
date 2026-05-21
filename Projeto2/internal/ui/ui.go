package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	cSuccess = color.New(color.FgGreen, color.Bold)
	cError   = color.New(color.FgRed, color.Bold)
	cWarning = color.New(color.FgYellow, color.Bold)
	cInfo    = color.New(color.FgCyan)
	cBold    = color.New(color.Bold)
	cDim     = color.New(color.Faint)
	cHeader  = color.New(color.FgBlue, color.Bold)
	cPurple  = color.New(color.FgMagenta, color.Bold)
)

func Banner(title string) {
	width := len(title) + 4
	line := strings.Repeat("─", width)
	fmt.Println()
	cHeader.Printf("  ┌%s┐\n", line)
	cHeader.Printf("  │  ")
	cBold.Printf("%s", title)
	cHeader.Printf("  │\n")
	cHeader.Printf("  └%s┘\n", line)
	fmt.Println()
}

func Section(title string) {
	fmt.Println()
	cPurple.Printf("  ── %s ──\n", title)
	fmt.Println()
}

func Step(n, total int, msg string) {
	cBold.Printf("  [%d/%d] ", n, total)
	fmt.Println(msg)
}

func Success(msg string) {
	cSuccess.Printf("  ✓ ")
	fmt.Println(msg)
}

func Fail(msg string) {
	cError.Printf("  ✗ ")
	fmt.Println(msg)
}

func Warn(msg string) {
	cWarning.Printf("  ⚠ ")
	fmt.Println(msg)
}

func Info(indent int, msg string) {
	pad := strings.Repeat("  ", indent)
	cInfo.Printf("%s  → ", pad)
	fmt.Println(msg)
}

func Dim(msg string) {
	cDim.Println("    " + msg)
}

func RunStep(msg string, work func() error) error {
	cInfo.Printf("  → %s", msg)
	start := time.Now()
	err := work()
	elapsed := time.Since(start)
	if err != nil {
		cError.Printf(" ✗ (%s)\n", FormatDuration(elapsed))
		return err
	}
	cSuccess.Printf(" ✓ (%s)\n", FormatDuration(elapsed))
	return nil
}

func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func StatusBadge(status string) string {
	switch strings.ToLower(status) {
	case "healthy", "success", "running":
		return cSuccess.Sprint("● " + status)
	case "unhealthy", "failed", "stopped", "error":
		return cError.Sprint("● " + status)
	case "degraded", "warning", "rolled-back":
		return cWarning.Sprint("● " + status)
	default:
		return cDim.Sprint("○ " + status)
	}
}

func Header(msg string) {
	cHeader.Println(msg)
}

func Bold(msg string) {
	cBold.Println(msg)
}

func KeyValue(key, value string) {
	cDim.Printf("  %-20s", key)
	fmt.Println(value)
}

func Separator() {
	cDim.Println("  " + strings.Repeat("·", 50))
}

func ErrorBox(title string, err error) {
	fmt.Println()
	cError.Printf("  ╔═ %s ═╗\n", title)
	cError.Printf("  ║ ")
	fmt.Printf("%s\n", err)
	cError.Printf("  ╚%s╝\n", strings.Repeat("═", len(title)+4))
	fmt.Println()
}

func SuccessBox(msg string) {
	fmt.Println()
	cSuccess.Println("  ╔══════════════════════════════════════╗")
	cSuccess.Printf("  ║  ✓ ")
	cBold.Printf("%-34s", msg)
	cSuccess.Println("║")
	cSuccess.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
}
