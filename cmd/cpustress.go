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
	cpuStressSelector  string
	cpuStressNamespace string
	cpuStressDuration  string
	cpuStressMethod    string
)

// cpustressCmd represents the cpustress command
var cpustressCmd = &cobra.Command{
	Use:   "cpustress",
	Short: "Inject CPU load into pods using ephemeral containers",
	Long: `Inject CPU load into pods using ephemeral containers.

This command will:
1. List pods matching the provided label selector
2. Add ephemeral containers to inject CPU stress using either stress-ng or yes command
3. The CPU stress will be applied for the specified duration

Examples:
  tipsy cpustress --selector "app=nginx" --duration "60s"
  tipsy cpustress --selector "environment=staging" --namespace production --method "yes" --duration "2m"
  tipsy cpustress --selector "tier=frontend" --method "stress-ng" --duration "30s" --dry-run --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print configuration if verbose mode is enabled
		PrintConfig()

		// Validate required flags
		if cpuStressSelector == "" {
			utils.Error("--selector flag is required")
			cmd.Help()
			return
		}

		// Validate method
		if cpuStressMethod != "stress-ng" && cpuStressMethod != "yes" {
			utils.Error("--method must be either 'stress-ng' or 'yes'")
			cmd.Help()
			return
		}

		// Use global namespace if not specified locally
		targetNamespace := cpuStressNamespace
		if targetNamespace == "" {
			targetNamespace = config.GlobalConfig.Namespace
		}
		if targetNamespace == "" {
			targetNamespace = "default"
		}

		// Parse duration
		durationParsed, err := time.ParseDuration(cpuStressDuration)
		if err != nil {
			utils.Error(fmt.Sprintf("Invalid duration format '%s': %v", cpuStressDuration, err))
			return
		}

		// Create Kubernetes client
		client, err := k8s.NewClient(config.GlobalConfig.Kubeconfig)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to create Kubernetes client: %v", err))
			return
		}

		// Execute the CPU stress injection
		err = chaos.InjectCPUStress(client, targetNamespace, cpuStressSelector, cpuStressMethod, durationParsed, config.GlobalConfig.DryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to inject CPU stress: %v", err))
			return
		}

		utils.Info("CPU stress injection operation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(cpustressCmd)

	// Local flags for the cpustress command
	cpustressCmd.Flags().StringVar(&cpuStressSelector, "selector", "", "Kubernetes label selector (required)")
	cpustressCmd.Flags().StringVar(&cpuStressNamespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	cpustressCmd.Flags().StringVar(&cpuStressDuration, "duration", "60s", "How long to run CPU stress (e.g., '30s', '1m', '5m')")
	cpustressCmd.Flags().StringVar(&cpuStressMethod, "method", "stress-ng", "CPU stress method: 'stress-ng' or 'yes'")

	// Mark selector as required
	cpustressCmd.MarkFlagRequired("selector")
}
