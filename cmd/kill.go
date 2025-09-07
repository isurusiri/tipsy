package cmd

import (
	"fmt"
	"time"

	"github.com/isurusiri/tipsy/internal/chaos"
	"github.com/isurusiri/tipsy/internal/config"
	"github.com/isurusiri/tipsy/internal/k8s"
	"github.com/isurusiri/tipsy/internal/state"
	"github.com/isurusiri/tipsy/internal/utils"
	"github.com/spf13/cobra"
)

var (
	selector  string
	namespace string
	count     int
)

// killCmd represents the kill command
var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill pods based on label selector",
	Long: `Kill pods in Kubernetes based on a label selector.

This command will:
1. List pods matching the provided label selector
2. Randomly select the specified number of pods to kill
3. Delete the selected pods (or simulate deletion in dry-run mode)

Examples:
  tipsy kill --selector "app=nginx" --count 2
  tipsy kill --selector "environment=staging" --namespace production --count 1
  tipsy kill --selector "tier=frontend" --dry-run --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print configuration if verbose mode is enabled
		PrintConfig()

		// Validate required flags
		if selector == "" {
			utils.Error("--selector flag is required")
			cmd.Help()
			return
		}

		// Use global namespace if not specified locally
		targetNamespace := namespace
		if targetNamespace == "" {
			targetNamespace = config.GlobalConfig.Namespace
		}
		if targetNamespace == "" {
			targetNamespace = "default"
		}

		// Create Kubernetes client
		client, err := k8s.NewClient(config.GlobalConfig.Kubeconfig)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to create Kubernetes client: %v", err))
			return
		}

		// Execute the kill operation
		killedPods, err := chaos.KillPods(client, targetNamespace, selector, count, config.GlobalConfig.DryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to kill pods: %v", err))
			return
		}

		// Save state for each killed pod
		if !config.GlobalConfig.DryRun {
			for _, podName := range killedPods {
				action := state.ChaosAction{
					Type:      "kill",
					TargetPod: podName,
					Namespace: targetNamespace,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Metadata: map[string]string{
						"selector": selector,
						"count":    fmt.Sprintf("%d", count),
					},
				}
				if err := state.SaveAction(action); err != nil {
					utils.Warn(fmt.Sprintf("Failed to save state for pod '%s': %v", podName, err))
				}
			}
		}

		utils.Info("Kill operation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(killCmd)

	// Local flags for the kill command
	killCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	killCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	killCmd.Flags().IntVar(&count, "count", 1, "Number of pods to kill (optional, default 1)")

	// Mark selector as required
	killCmd.MarkFlagRequired("selector")
}
