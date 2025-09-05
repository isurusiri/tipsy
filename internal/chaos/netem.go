package chaos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/isurusiri/tipsy/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// InjectLatency injects network latency using tc netem via ephemeral containers
func InjectLatency(client kubernetes.Interface, namespace, selector, delay string, duration time.Duration, dryRun bool) error {
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

		err := injectLatencyToPod(client, namespace, pod.Name, delay, duration, dryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to inject latency to pod '%s': %v", pod.Name, err))
			// Continue with other pods even if one fails
		}
	}

	return nil
}

// injectLatencyToPod injects latency to a specific pod using an ephemeral container
func injectLatencyToPod(client kubernetes.Interface, namespace, podName, delay string, duration time.Duration, dryRun bool) error {
	utils.Info(fmt.Sprintf("Injecting latency to pod '%s' with delay '%s' for duration '%s'", podName, delay, duration))

	if dryRun {
		utils.DryRun(fmt.Sprintf("Would inject latency to pod '%s':", podName))
		utils.DryRun(fmt.Sprintf("  - Add ephemeral container with image: ghcr.io/chaos-tools/netem:latest"))
		utils.DryRun(fmt.Sprintf("  - Command: nsenter -t <pid> -n tc qdisc add dev eth0 root netem delay %s", delay))
		utils.DryRun(fmt.Sprintf("  - Duration: %s", duration))
		return nil
	}

	// Create ephemeral container spec
	ephemeralContainer := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  fmt.Sprintf("latency-injector-%d", time.Now().Unix()),
			Image: "ghcr.io/chaos-tools/netem:latest",
			Command: []string{
				"sh",
				"-c",
				fmt.Sprintf(`
					# Find the main container's PID
					MAIN_PID=$(ps -o pid= -C %s 2>/dev/null | head -1)
					if [ -z "$MAIN_PID" ]; then
						# Fallback: find any process in the container
						MAIN_PID=1
					fi
					
					# Apply network delay using tc netem
					nsenter -t $MAIN_PID -n tc qdisc add dev eth0 root netem delay %s
					
					# Wait for the specified duration
					sleep %d
					
					# Clean up: remove the netem rule
					nsenter -t $MAIN_PID -n tc qdisc del dev eth0 root netem delay %s 2>/dev/null || true
				`, getMainProcessName(), delay, int(duration.Seconds()), delay),
			},
			SecurityContext: &corev1.SecurityContext{
				Privileged: &[]bool{true}[0],
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"NET_ADMIN", "SYS_PTRACE"},
				},
			},
		},
	}

	// Apply the patch using the ephemeralContainers subresource
	_, err := client.CoreV1().Pods(namespace).Patch(
		context.TODO(),
		podName,
		types.StrategicMergePatchType,
		[]byte(fmt.Sprintf(`{"spec":{"ephemeralContainers":[%s]}}`, marshalEphemeralContainer(ephemeralContainer))),
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
		utils.Info(fmt.Sprintf("Latency injection completed for pod '%s'", podName))
	}()

	return nil
}

// getMainProcessName returns a common process name to look for in the container
// This is a simple heuristic - in practice, you might want to make this configurable
func getMainProcessName() string {
	// Common process names in containers
	commonProcesses := []string{"nginx", "apache2", "httpd", "node", "python", "java", "go", "main"}
	return commonProcesses[0] // Default to nginx, but the script will fallback to PID 1
}

// InjectPacketLoss injects network packet loss using tc netem via ephemeral containers
func InjectPacketLoss(client kubernetes.Interface, namespace, selector, loss string, duration time.Duration, dryRun bool) error {
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

		err := injectPacketLossToPod(client, namespace, pod.Name, loss, duration, dryRun)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to inject packet loss to pod '%s': %v", pod.Name, err))
			// Continue with other pods even if one fails
		}
	}

	return nil
}

// injectPacketLossToPod injects packet loss to a specific pod using an ephemeral container
func injectPacketLossToPod(client kubernetes.Interface, namespace, podName, loss string, duration time.Duration, dryRun bool) error {
	utils.Info(fmt.Sprintf("Injecting packet loss to pod '%s' with loss '%s' for duration '%s'", podName, loss, duration))

	if dryRun {
		utils.DryRun(fmt.Sprintf("Would inject packet loss to pod '%s':", podName))
		utils.DryRun(fmt.Sprintf("  - Add ephemeral container with image: ghcr.io/chaos-tools/netem:latest"))
		utils.DryRun(fmt.Sprintf("  - Command: nsenter -t <pid> -n tc qdisc add dev eth0 root netem loss %s", loss))
		utils.DryRun(fmt.Sprintf("  - Duration: %s", duration))
		return nil
	}

	// Create ephemeral container spec
	ephemeralContainer := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  fmt.Sprintf("packetloss-injector-%d", time.Now().Unix()),
			Image: "ghcr.io/chaos-tools/netem:latest",
			Command: []string{
				"sh",
				"-c",
				fmt.Sprintf(`
					# Find the main container's PID
					MAIN_PID=$(ps -o pid= -C %s 2>/dev/null | head -1)
					if [ -z "$MAIN_PID" ]; then
						# Fallback: find any process in the container
						MAIN_PID=1
					fi
					
					# Apply network packet loss using tc netem
					nsenter -t $MAIN_PID -n tc qdisc add dev eth0 root netem loss %s
					
					# Wait for the specified duration
					sleep %d
					
					# Clean up: remove the netem rule
					nsenter -t $MAIN_PID -n tc qdisc del dev eth0 root netem loss %s 2>/dev/null || true
				`, getMainProcessName(), loss, int(duration.Seconds()), loss),
			},
			SecurityContext: &corev1.SecurityContext{
				Privileged: &[]bool{true}[0],
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"NET_ADMIN", "SYS_PTRACE"},
				},
			},
		},
	}

	// Apply the patch using the ephemeralContainers subresource
	_, err := client.CoreV1().Pods(namespace).Patch(
		context.TODO(),
		podName,
		types.StrategicMergePatchType,
		[]byte(fmt.Sprintf(`{"spec":{"ephemeralContainers":[%s]}}`, marshalEphemeralContainer(ephemeralContainer))),
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
		utils.Info(fmt.Sprintf("Packet loss injection completed for pod '%s'", podName))
	}()

	return nil
}

// marshalEphemeralContainer converts an EphemeralContainer to JSON for patching
func marshalEphemeralContainer(container corev1.EphemeralContainer) string {
	// Escape the command string for JSON
	commandStr := container.Command[2]
	// Basic JSON escaping for the command string
	commandStr = strings.ReplaceAll(commandStr, `"`, `\"`)
	commandStr = strings.ReplaceAll(commandStr, "\n", "\\n")
	commandStr = strings.ReplaceAll(commandStr, "\r", "\\r")
	commandStr = strings.ReplaceAll(commandStr, "\t", "\\t")
	
	return fmt.Sprintf(`{
		"name": "%s",
		"image": "%s",
		"command": ["sh", "-c", "%s"],
		"securityContext": {
			"privileged": true,
			"capabilities": {
				"add": ["NET_ADMIN", "SYS_PTRACE"]
			}
		}
	}`, container.Name, container.Image, commandStr)
}
