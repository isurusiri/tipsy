package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/isurusiri/tipsy/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KillPods deletes pods based on a label selector
func KillPods(client kubernetes.Interface, namespace, selector string, count int, dryRun bool) error {
	utils.Info(fmt.Sprintf("Searching for pods with selector '%s' in namespace '%s'", selector, namespace))

	// List pods matching the selector
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		utils.Warn(fmt.Sprintf("No pods found matching selector '%s' in namespace '%s'", selector, namespace))
		return nil
	}

	utils.Info(fmt.Sprintf("Found %d pod(s) matching selector", len(pods.Items)))

	// Determine how many pods to kill
	podsToKill := count
	if count <= 0 {
		podsToKill = 0
		utils.Warn(fmt.Sprintf("Invalid count %d, no pods will be deleted", count))
	} else if len(pods.Items) < count {
		podsToKill = len(pods.Items)
		utils.Warn(fmt.Sprintf("Only %d pods available, limiting kill count to %d", len(pods.Items), podsToKill))
	}

	// Randomly select pods to kill
	rand.Seed(time.Now().UnixNano())
	selectedPods := make([]string, 0, podsToKill)
	usedIndices := make(map[int]bool)

	for len(selectedPods) < podsToKill {
		index := rand.Intn(len(pods.Items))
		if !usedIndices[index] {
			usedIndices[index] = true
			selectedPods = append(selectedPods, pods.Items[index].Name)
		}
	}

	// Execute the kill operation
	if dryRun {
		utils.DryRun(fmt.Sprintf("Would delete %d pod(s):", len(selectedPods)))
		for _, podName := range selectedPods {
			utils.DryRun(fmt.Sprintf("  - %s", podName))
		}
	} else {
		utils.Info(fmt.Sprintf("Deleting %d pod(s):", len(selectedPods)))
		for _, podName := range selectedPods {
			utils.Info(fmt.Sprintf("  Deleting pod: %s", podName))
			
			err := client.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
			if err != nil {
				utils.Error(fmt.Sprintf("Failed to delete pod '%s': %v", podName, err))
				// Continue with other pods even if one fails
			} else {
				utils.Info(fmt.Sprintf("  Successfully deleted pod: %s", podName))
			}
		}
	}

	return nil
}
