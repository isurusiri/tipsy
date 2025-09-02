package utils

import (
	"time"

	"github.com/fatih/color"
	"github.com/isurusiri/tipsy/internal/config"
)

// Info prints an informational message with blue color and [INFO] prefix
func Info(msg string) {
	timestamp := time.Now().Format("15:04:05")
	infoColor := color.New(color.FgBlue, color.Bold)
	infoColor.Printf("[%s] [INFO] %s\n", timestamp, msg)
}

// Warn prints a warning message with yellow color and [WARN] prefix
func Warn(msg string) {
	timestamp := time.Now().Format("15:04:05")
	warnColor := color.New(color.FgYellow, color.Bold)
	warnColor.Printf("[%s] [WARN] %s\n", timestamp, msg)
}

// Error prints an error message with red color and [ERROR] prefix
func Error(msg string) {
	timestamp := time.Now().Format("15:04:05")
	errorColor := color.New(color.FgRed, color.Bold)
	errorColor.Printf("[%s] [ERROR] %s\n", timestamp, msg)
}

// DryRun prints a dry-run message with cyan color and [DRY-RUN] prefix
// Only prints if config.GlobalConfig.DryRun is true
func DryRun(msg string) {
	if config.GlobalConfig.DryRun {
		timestamp := time.Now().Format("15:04:05")
		dryRunColor := color.New(color.FgCyan, color.Bold)
		dryRunColor.Printf("[%s] [DRY-RUN] %s\n", timestamp, msg)
	}
}
