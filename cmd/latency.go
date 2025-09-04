package cmd

import (
	"fmt"
	"time"

	"github.com/isurusiri/tipsy/internal/chaos"
	"github.com/isurusiri/tipsy/internal/config"
	"github.com/isurusiri/tipsy/internal/k8s"
	"github.com/isurusiri/tipsy/internal/utils"
	"github.com/spf13/cobra"
)

var (
	latencySelector  string
	latencyNamespace string
	delay            string
	duration         string
)

// latencyCmd represents the latency command
var latencyCmd = &cobra.Command{
	Use:   "latency",
	Short: "Inject network latency using tc netem via ephemeral containers",
	Long: `Inject network latency into pods using tc netem via ephemeral containers.

This command will:
1. List pods matching the provided label selector
2. Add ephemeral containers to inject network latency using tc netem
3. The latency will be applied for the specified duration

Examples:
  tipsy latency --selector "app=nginx" --delay "200ms" --duration "30s"
  tipsy latency --selector "environment=staging" --namespace production --delay "500ms" --duration "1m"
  tipsy latency --selector "tier=frontend" --dry-run --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print configuration if verbose mode is enabled
		PrintConfig()

		// Validate required flags
		if latencySelector == "" {
			utils.Error("--selector flag is required")
			cmd.Help()
			return
		}

		// Use global namespace if not specified locally
		targetNamespace := latencyNamespace
		if targetNamespace == "" {
			targetNamespace = config.GlobalConfig.Namespace
		}
		if targetNamespace == "" {
			targetNamespace = "default"
		}

		// Parse duration
		durationParsed, err := time.ParseDuration(duration)
		if err != nil {
			utils.Error(fmt.Sprintf("Invalid duration format '%s': %v", duration, err))
			return
		}

		// Create Kubernetes client
		client, err := k8s.NewClient(config.GlobalConfig.Kubeconfig)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to create Kubernetes client: %v", err))
			return
		}

		// Execute the latency injection
		err = chaos.InjectLatency(client, targetNamespace, latencySelector, delay, durationParsed, config.GlobalConfig.DryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to inject latency: %v", err))
			return
		}

		utils.Info("Latency injection operation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(latencyCmd)

	// Local flags for the latency command
	latencyCmd.Flags().StringVar(&latencySelector, "selector", "", "Kubernetes label selector (required)")
	latencyCmd.Flags().StringVar(&latencyNamespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	latencyCmd.Flags().StringVar(&delay, "delay", "200ms", "Network delay to inject (e.g., '200ms', '500ms', '1s')")
	latencyCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep latency active (e.g., '30s', '1m', '5m')")

	// Mark selector as required
	latencyCmd.MarkFlagRequired("selector")
}
