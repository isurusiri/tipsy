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
	packetLossSelector  string
	packetLossNamespace string
	loss                string
	packetLossDuration  string
)

// packetLossCmd represents the packetloss command
var packetLossCmd = &cobra.Command{
	Use:   "packetloss",
	Short: "Inject network packet loss using tc netem via ephemeral containers",
	Long: `Inject network packet loss into pods using tc netem via ephemeral containers.

This command will:
1. List pods matching the provided label selector
2. Add ephemeral containers to inject network packet loss using tc netem
3. The packet loss will be applied for the specified duration

Examples:
  tipsy packetloss --selector "app=nginx" --loss "30%" --duration "30s"
  tipsy packetloss --selector "environment=staging" --namespace production --loss "50%" --duration "1m"
  tipsy packetloss --selector "tier=frontend" --dry-run --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print configuration if verbose mode is enabled
		PrintConfig()

		// Validate required flags
		if packetLossSelector == "" {
			utils.Error("--selector flag is required")
			cmd.Help()
			return
		}

		// Use global namespace if not specified locally
		targetNamespace := packetLossNamespace
		if targetNamespace == "" {
			targetNamespace = config.GlobalConfig.Namespace
		}
		if targetNamespace == "" {
			targetNamespace = "default"
		}

		// Parse duration
		durationParsed, err := time.ParseDuration(packetLossDuration)
		if err != nil {
			utils.Error(fmt.Sprintf("Invalid duration format '%s': %v", packetLossDuration, err))
			return
		}

		// Create Kubernetes client
		client, err := k8s.NewClient(config.GlobalConfig.Kubeconfig)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to create Kubernetes client: %v", err))
			return
		}

		// Execute the packet loss injection
		affectedPods, err := chaos.InjectPacketLoss(client, targetNamespace, packetLossSelector, loss, durationParsed, config.GlobalConfig.DryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to inject packet loss: %v", err))
			return
		}

		// Save state for each affected pod
		if !config.GlobalConfig.DryRun {
			for _, podName := range affectedPods {
				action := state.ChaosAction{
					Type:      "packetloss",
					TargetPod: podName,
					Namespace: targetNamespace,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Metadata: map[string]string{
						"loss":     loss,
						"duration": packetLossDuration,
						"selector": packetLossSelector,
					},
				}
				if err := state.SaveAction(action); err != nil {
					utils.Warn(fmt.Sprintf("Failed to save state for pod '%s': %v", podName, err))
				}
			}
		}

		utils.Info("Packet loss injection operation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(packetLossCmd)

	// Local flags for the packetloss command
	packetLossCmd.Flags().StringVar(&packetLossSelector, "selector", "", "Kubernetes label selector (required)")
	packetLossCmd.Flags().StringVar(&packetLossNamespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	packetLossCmd.Flags().StringVar(&loss, "loss", "30%", "Network packet loss percentage to inject (e.g., '30%', '50%', '10%')")
	packetLossCmd.Flags().StringVar(&packetLossDuration, "duration", "30s", "How long to keep packet loss active (e.g., '30s', '1m', '5m')")

	// Mark selector as required
	packetLossCmd.MarkFlagRequired("selector")
}
