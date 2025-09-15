package rollback

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/isurusiri/tipsy/internal/state"
	"github.com/isurusiri/tipsy/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// RollbackAll rolls back all chaos actions or filtered actions
func RollbackAll(client kubernetes.Interface, dryRun bool, filterType, filterPod string) error {
	utils.Info("Starting rollback operation")

	// Load all actions from state
	actions, err := state.LoadActions()
	if err != nil {
		return fmt.Errorf("failed to load actions from state: %w", err)
	}

	if len(actions) == 0 {
		utils.Info("No actions found to rollback")
		return nil
	}

	// Filter actions if filters are provided
	filteredActions := filterActions(actions, filterType, filterPod)
	if len(filteredActions) == 0 {
		utils.Info("No actions match the specified filters")
		return nil
	}

	utils.Info(fmt.Sprintf("Found %d action(s) to rollback", len(filteredActions)))

	var successCount, failureCount int
	var failedActions []state.ChaosAction

	// Process each action
	for _, action := range filteredActions {
		utils.Info(fmt.Sprintf("Rolling back %s action for pod '%s' in namespace '%s'", 
			action.Type, action.TargetPod, action.Namespace))

		err := rollbackAction(client, action, dryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to rollback action: %v", err))
			failedActions = append(failedActions, action)
			failureCount++
		} else {
			// Only remove from state if not dry run and rollback was successful
			if !dryRun {
				err = state.DeleteAction(action)
				if err != nil {
					utils.Warn(fmt.Sprintf("Failed to remove action from state: %v", err))
				}
			}
			successCount++
		}
	}

	// Print summary
	if dryRun {
		utils.Info(fmt.Sprintf("Dry run completed: %d action(s) would be rolled back", len(filteredActions)))
	} else {
		utils.Info(fmt.Sprintf("Rollback completed: %d successful, %d failed", successCount, failureCount))
		if len(failedActions) > 0 {
			utils.Warn(fmt.Sprintf("Failed to rollback %d action(s)", len(failedActions)))
		}
	}

	return nil
}

// filterActions filters actions based on type and pod filters
func filterActions(actions []state.ChaosAction, filterType, filterPod string) []state.ChaosAction {
	var filtered []state.ChaosAction

	for _, action := range actions {
		// Check type filter
		if filterType != "" && action.Type != filterType {
			continue
		}

		// Check pod filter
		if filterPod != "" && action.TargetPod != filterPod {
			continue
		}

		filtered = append(filtered, action)
	}

	return filtered
}

// rollbackAction rolls back a specific action based on its type
func rollbackAction(client kubernetes.Interface, action state.ChaosAction, dryRun bool) error {
	switch action.Type {
	case "latency", "packetloss":
		return RevertTC(client, action, dryRun)
	case "cpustress":
		return RemoveEphemeral(client, action, dryRun)
	case "misroute":
		return RestoreEndpoints(client, action, dryRun)
	case "kill":
		return handleKillAction(client, action, dryRun)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// RevertTC reverts tc netem changes by removing ephemeral containers
func RevertTC(client kubernetes.Interface, action state.ChaosAction, dryRun bool) error {
	utils.Info(fmt.Sprintf("Reverting tc netem for pod '%s' in namespace '%s'", 
		action.TargetPod, action.Namespace))

	if dryRun {
		utils.DryRun(fmt.Sprintf("Would remove ephemeral containers from pod '%s' to revert tc netem", action.TargetPod))
		return nil
	}

	// Get the pod to find ephemeral containers
	pod, err := client.CoreV1().Pods(action.Namespace).Get(context.TODO(), action.TargetPod, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Find and remove ephemeral containers that match our pattern
	var containersToRemove []string
	for _, container := range pod.Spec.EphemeralContainers {
		if strings.Contains(container.Name, "latency-injector") || 
		   strings.Contains(container.Name, "packetloss-injector") ||
		   strings.Contains(container.Name, "tipsy-") {
			containersToRemove = append(containersToRemove, container.Name)
		}
	}

	if len(containersToRemove) == 0 {
		utils.Info("No ephemeral containers found to remove")
		return nil
	}

	// Remove ephemeral containers by patching the pod spec
	for _, containerName := range containersToRemove {
		err := removeEphemeralContainer(client, action.Namespace, action.TargetPod, containerName)
		if err != nil {
			utils.Warn(fmt.Sprintf("Failed to remove ephemeral container '%s': %v", containerName, err))
		}
	}

	utils.Info(fmt.Sprintf("Successfully reverted tc netem for pod '%s'", action.TargetPod))
	return nil
}

// RemoveEphemeral removes ephemeral containers from a pod
func RemoveEphemeral(client kubernetes.Interface, action state.ChaosAction, dryRun bool) error {
	utils.Info(fmt.Sprintf("Removing ephemeral containers from pod '%s' in namespace '%s'", 
		action.TargetPod, action.Namespace))

	if dryRun {
		utils.DryRun(fmt.Sprintf("Would remove ephemeral containers from pod '%s'", action.TargetPod))
		return nil
	}

	// Get the pod to find ephemeral containers
	pod, err := client.CoreV1().Pods(action.Namespace).Get(context.TODO(), action.TargetPod, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Find ephemeral containers that match our pattern
	var containersToRemove []string
	for _, container := range pod.Spec.EphemeralContainers {
		if strings.Contains(container.Name, "tipsy-cpu-stress") || 
		   strings.Contains(container.Name, "tipsy-") {
			containersToRemove = append(containersToRemove, container.Name)
		}
	}

	if len(containersToRemove) == 0 {
		utils.Info("No ephemeral containers found to remove")
		return nil
	}

	// Remove ephemeral containers
	for _, containerName := range containersToRemove {
		err := removeEphemeralContainer(client, action.Namespace, action.TargetPod, containerName)
		if err != nil {
			utils.Warn(fmt.Sprintf("Failed to remove ephemeral container '%s': %v", containerName, err))
		}
	}

	utils.Info(fmt.Sprintf("Successfully removed ephemeral containers from pod '%s'", action.TargetPod))
	return nil
}

// RestoreEndpoints restores original service endpoints from backup
func RestoreEndpoints(client kubernetes.Interface, action state.ChaosAction, dryRun bool) error {
	utils.Info(fmt.Sprintf("Restoring endpoints for service '%s' in namespace '%s'", 
		action.TargetPod, action.Namespace))

	// Get backup path from metadata
	backupPath, exists := action.Metadata["backupPath"]
	if !exists {
		// Try to construct the backup path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		backupPath = filepath.Join(homeDir, ".tipsy", "rollback", fmt.Sprintf("%s_%s.json", action.TargetPod, action.Namespace))
	}

	if dryRun {
		utils.DryRun(fmt.Sprintf("Would restore endpoints for service '%s' from backup: %s", action.TargetPod, backupPath))
		return nil
	}

	// Read the backup file
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file %s: %w", backupPath, err)
	}

	// Unmarshal the backup endpoints
	var backupEndpoints corev1.Endpoints
	err = json.Unmarshal(backupData, &backupEndpoints)
	if err != nil {
		return fmt.Errorf("failed to unmarshal backup endpoints: %w", err)
	}

	// Update the endpoints
	_, err = client.CoreV1().Endpoints(action.Namespace).Update(context.TODO(), &backupEndpoints, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restore endpoints: %w", err)
	}

	// Clean up the backup file
	err = os.Remove(backupPath)
	if err != nil {
		utils.Warn(fmt.Sprintf("Failed to remove backup file %s: %v", backupPath, err))
	}

	utils.Info(fmt.Sprintf("Successfully restored endpoints for service '%s'", action.TargetPod))
	return nil
}

// removeEphemeralContainer removes a specific ephemeral container from a pod
func removeEphemeralContainer(client kubernetes.Interface, namespace, podName, containerName string) error {
	// Get current pod
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Create new ephemeral containers list without the target container
	var newEphemeralContainers []corev1.EphemeralContainer
	for _, container := range pod.Spec.EphemeralContainers {
		if container.Name != containerName {
			newEphemeralContainers = append(newEphemeralContainers, container)
		}
	}

	// Patch the pod to remove the ephemeral container
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"ephemeralContainers": newEphemeralContainers,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	_, err = client.CoreV1().Pods(namespace).Patch(
		context.TODO(),
		podName,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
		"ephemeralContainers",
	)

	if err != nil {
		return fmt.Errorf("failed to remove ephemeral container: %w", err)
	}

	utils.Info(fmt.Sprintf("Successfully removed ephemeral container '%s' from pod '%s'", containerName, podName))
	return nil
}

// handleKillAction handles kill actions - these cannot be rolled back as pods are permanently deleted
func handleKillAction(client kubernetes.Interface, action state.ChaosAction, dryRun bool) error {
	utils.Warn(fmt.Sprintf("Cannot rollback kill action for pod '%s' in namespace '%s' - pod was permanently deleted", 
		action.TargetPod, action.Namespace))
	
	if dryRun {
		utils.DryRun(fmt.Sprintf("Would skip kill action for pod '%s' (cannot be rolled back)", action.TargetPod))
	}
	
	// For kill actions, we consider them "successfully" handled (even though we can't rollback)
	// This allows them to be removed from state without causing rollback failures
	return nil
}
