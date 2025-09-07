package cmd

import (
	"fmt"

	"github.com/isurusiri/tipsy/internal/chaos"
	"github.com/isurusiri/tipsy/internal/config"
	"github.com/isurusiri/tipsy/internal/k8s"
	"github.com/isurusiri/tipsy/internal/utils"
	"github.com/spf13/cobra"
)

var (
	misrouteService           string
	misrouteNamespace         string
	misrouteRemoveAll         bool
	misrouteReplaceWithSelector string
)

// misrouteCmd represents the misroute command
var misrouteCmd = &cobra.Command{
	Use:   "misroute",
	Short: "Misroute traffic by manipulating service endpoints",
	Long: `Misroute traffic by removing or replacing endpoints in a Kubernetes Service.

This command will:
1. Fetch the target service and its current endpoints
2. Save the original endpoints for rollback purposes
3. Either remove all endpoints or replace them with pods matching a selector
4. Update the service endpoints to simulate misrouting

Examples:
  tipsy misroute --service my-service --remove-all
  tipsy misroute --service my-service --replace-with-selector "app=nginx"
  tipsy misroute --service my-service --namespace production --remove-all --dry-run
  tipsy misroute --service my-service --replace-with-selector "tier=backend" --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		// Print configuration if verbose mode is enabled
		PrintConfig()

		// Validate required flags
		if misrouteService == "" {
			utils.Error("--service flag is required")
			cmd.Help()
			return
		}

		// Validate that either remove-all or replace-with-selector is specified, but not both
		if !misrouteRemoveAll && misrouteReplaceWithSelector == "" {
			utils.Error("either --remove-all or --replace-with-selector must be specified")
			cmd.Help()
			return
		}

		if misrouteRemoveAll && misrouteReplaceWithSelector != "" {
			utils.Error("--remove-all and --replace-with-selector cannot be used together")
			cmd.Help()
			return
		}

		// Use global namespace if not specified locally
		targetNamespace := misrouteNamespace
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

		// Execute the misroute operation
		err = chaos.MisrouteService(client, misrouteService, targetNamespace, misrouteReplaceWithSelector, misrouteRemoveAll, config.GlobalConfig.DryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to misroute service: %v", err))
			return
		}

		utils.Info("Service misrouting operation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(misrouteCmd)

	// Local flags for the misroute command
	misrouteCmd.Flags().StringVar(&misrouteService, "service", "", "Target service name (required)")
	misrouteCmd.Flags().StringVar(&misrouteNamespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	misrouteCmd.Flags().BoolVar(&misrouteRemoveAll, "remove-all", false, "Remove all endpoints from the service")
	misrouteCmd.Flags().StringVar(&misrouteReplaceWithSelector, "replace-with-selector", "", "Replace endpoints with those from pods matching this selector")

	// Mark service as required
	misrouteCmd.MarkFlagRequired("service")
}
