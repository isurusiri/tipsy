package cmd

import (
	"fmt"
	"os"

	"github.com/isurusiri/tipsy/internal/config"
	"github.com/spf13/cobra"
)

// rootCmd - the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tipsy",
	Short: "Chaos engineering CLI for Kubernetes.",
	Long: `Tipsy is a minimalist chaos engineering CLI for Kubernetes.

It lets you inject real-world failures—like pod kills, network latency, and CPU stress—
using ephemeral containers and native Kubernetes APIs. No sidecars. No CRDs.

Safe, scriptable, and built for CI/CD pipelines and production-grade testing.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Namespace, "namespace", "", "Kubernetes namespace to target")
	rootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.DryRun, "dry-run", false, "if true, simulate actions without taking effect")
	rootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.Verbose, "verbose", false, "enable verbose logging")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// PrintConfig prints the current configuration (useful for debugging)
func PrintConfig() {
	if config.GlobalConfig.Verbose {
		fmt.Printf("Configuration:\n")
		fmt.Printf("  Kubeconfig: %s\n", config.GlobalConfig.Kubeconfig)
		fmt.Printf("  Namespace: %s\n", config.GlobalConfig.Namespace)
		fmt.Printf("  Dry Run: %t\n", config.GlobalConfig.DryRun)
		fmt.Printf("  Verbose: %t\n", config.GlobalConfig.Verbose)
	}
} 