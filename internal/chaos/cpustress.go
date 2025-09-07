package chaos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/isurusiri/tipsy/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// InjectCPUStress injects CPU load into pods using ephemeral containers
func InjectCPUStress(client kubernetes.Interface, namespace, selector, method string, duration time.Duration, dryRun bool) error {
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

	// Process each pod
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			utils.Warn(fmt.Sprintf("Skipping pod '%s' - not in Running state (current: %s)", pod.Name, pod.Status.Phase))
			continue
		}

		err := injectCPUStressToPod(client, namespace, pod.Name, method, duration, dryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to inject CPU stress to pod '%s': %v", pod.Name, err))
			// Continue with other pods even if one fails
		}
	}

	return nil
}

// injectCPUStressToPod injects CPU stress to a specific pod using an ephemeral container
func injectCPUStressToPod(client kubernetes.Interface, namespace, podName, method string, duration time.Duration, dryRun bool) error {
	utils.Info(fmt.Sprintf("Injecting CPU stress to pod '%s' using method '%s' for duration '%s'", podName, method, duration))

	// Determine image and command based on method
	var image string
	var command []string

	switch method {
	case "stress-ng":
		image = "ghcr.io/chaos-tools/stress-ng:latest"
		command = []string{
			"stress-ng",
			"--cpu", "1",
			"--timeout", fmt.Sprintf("%.0fs", duration.Seconds()),
		}
	case "yes":
		image = "alpine"
		command = []string{
			"sh",
			"-c",
			fmt.Sprintf("yes > /dev/null & sleep %d", int(duration.Seconds())),
		}
	default:
		return fmt.Errorf("unsupported method: %s", method)
	}

	if dryRun {
		utils.DryRun(fmt.Sprintf("Would inject CPU stress to pod '%s':", podName))
		utils.DryRun(fmt.Sprintf("  - Add ephemeral container with image: %s", image))
		utils.DryRun(fmt.Sprintf("  - Command: %s", strings.Join(command, " ")))
		utils.DryRun(fmt.Sprintf("  - Duration: %s", duration))
		return nil
	}

	// Create ephemeral container spec
	ephemeralContainer := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    fmt.Sprintf("tipsy-cpu-stress-%d", time.Now().Unix()),
			Image:   image,
			Command: command,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
	}

	// Apply the patch using the ephemeralContainers subresource
	_, err := client.CoreV1().Pods(namespace).Patch(
		context.TODO(),
		podName,
		types.StrategicMergePatchType,
		[]byte(fmt.Sprintf(`{"spec":{"ephemeralContainers":[%s]}}`, marshalCPUStressEphemeralContainer(ephemeralContainer))),
		metav1.PatchOptions{},
		"ephemeralContainers",
	)

	if err != nil {
		return fmt.Errorf("failed to add ephemeral container: %w", err)
	}

	utils.Info(fmt.Sprintf("Successfully added ephemeral container to pod '%s'", podName))

	// Start a goroutine to monitor and clean up after duration
	go func() {
		time.Sleep(duration)
		utils.Info(fmt.Sprintf("CPU stress injection completed for pod '%s'", podName))
	}()

	return nil
}

// marshalCPUStressEphemeralContainer converts an EphemeralContainer to JSON for patching
func marshalCPUStressEphemeralContainer(container corev1.EphemeralContainer) string {
	// Escape the command strings for JSON
	commandStr := `["` + strings.Join(container.Command, `", "`) + `"]`
	
	// Build resources JSON
	resourcesJSON := `,
		"resources": {
			"requests": {
				"cpu": "100m",
				"memory": "64Mi"
			},
			"limits": {
				"cpu": "500m",
				"memory": "128Mi"
			}
		}`
	
	return fmt.Sprintf(`{
		"name": "%s",
		"image": "%s",
		"command": %s%s
	}`, container.Name, container.Image, commandStr, resourcesJSON)
}
