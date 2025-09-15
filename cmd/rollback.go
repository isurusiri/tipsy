package cmd

import (
	"fmt"

	"github.com/isurusiri/tipsy/internal/config"
	"github.com/isurusiri/tipsy/internal/k8s"
	"github.com/isurusiri/tipsy/internal/rollback"
	"github.com/isurusiri/tipsy/internal/utils"
	"github.com/spf13/cobra"
)

var (
	rollbackDryRun bool
	rollbackType   string
	rollbackPod    string
)

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback all chaos actions or specific actions",
	Long: `Rollback chaos actions that were previously applied.

This command will:
1. Read all actions from ~/.tipsy/state.json
2. Revert each action based on its type:
   - latency/packetloss: Remove ephemeral containers and clean up tc netem
   - cpustress: Remove ephemeral containers
   - misroute: Restore original service endpoints from backup
3. Remove successfully rolled back actions from state.json

Examples:
  tipsy rollback                    # Rollback all actions
  tipsy rollback --type latency     # Rollback only latency actions
  tipsy rollback --pod my-pod       # Rollback actions for specific pod
  tipsy rollback --dry-run          # Show what would be rolled back without executing`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print configuration if verbose mode is enabled
		PrintConfig()

		// Create Kubernetes client
		client, err := k8s.NewClient(config.GlobalConfig.Kubeconfig)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to create Kubernetes client: %v", err))
			return
		}

		// Use global dry-run flag if not specified locally
		dryRun := rollbackDryRun
		if !dryRun {
			dryRun = config.GlobalConfig.DryRun
		}

		// Execute rollback
		err = rollback.RollbackAll(client, dryRun, rollbackType, rollbackPod)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to rollback actions: %v", err))
			return
		}

		utils.Info("Rollback operation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)

	// Local flags for the rollback command
	rollbackCmd.Flags().BoolVar(&rollbackDryRun, "dry-run", false, "Show what would be rolled back without executing")
	rollbackCmd.Flags().StringVar(&rollbackType, "type", "", "Rollback only actions of specific type (latency, packetloss, cpustress, misroute)")
	rollbackCmd.Flags().StringVar(&rollbackPod, "pod", "", "Rollback only actions for specific pod")
}
