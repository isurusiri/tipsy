package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isurusiri/tipsy/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MisrouteService manipulates service endpoints to simulate misrouting
func MisrouteService(client *kubernetes.Clientset, svcName, namespace, replaceSelector string, removeAll, dryRun bool) error {
	utils.Info(fmt.Sprintf("Starting misroute operation for service '%s' in namespace '%s'", svcName, namespace))

	// Fetch the Service
	service, err := client.CoreV1().Services(namespace).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get service '%s': %w", svcName, err)
	}

	utils.Info(fmt.Sprintf("Found service '%s' with %d ports", svcName, len(service.Spec.Ports)))

	// Fetch the current Endpoints
	endpoints, err := client.CoreV1().Endpoints(namespace).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get endpoints for service '%s': %w", svcName, err)
	}

	// Save original endpoints for rollback
	if !dryRun {
		err = saveOriginalEndpoints(endpoints, svcName, namespace)
		if err != nil {
			utils.Warn(fmt.Sprintf("Failed to save original endpoints for rollback: %v", err))
		}
	}

	// Create a copy of the endpoints to modify
	modifiedEndpoints := endpoints.DeepCopy()

	if removeAll {
		// Remove all endpoint subsets
		utils.Info("Removing all endpoint subsets")
		modifiedEndpoints.Subsets = []corev1.EndpointSubset{}
	} else if replaceSelector != "" {
		// Replace with pods matching the selector
		utils.Info(fmt.Sprintf("Replacing endpoints with pods matching selector '%s'", replaceSelector))
		
		// List pods matching the selector
		pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: replaceSelector,
		})
		if err != nil {
			return fmt.Errorf("failed to list pods with selector '%s': %w", replaceSelector, err)
		}

		if len(pods.Items) == 0 {
			utils.Warn(fmt.Sprintf("No pods found matching selector '%s' in namespace '%s'", replaceSelector, namespace))
			// Set empty subsets to effectively remove all endpoints
			modifiedEndpoints.Subsets = []corev1.EndpointSubset{}
		} else {
			utils.Info(fmt.Sprintf("Found %d pod(s) matching selector", len(pods.Items)))
			
			// Build new endpoint subsets from the pods
			newSubsets, err := buildEndpointSubsetsFromPods(pods.Items, service.Spec.Ports)
			if err != nil {
				return fmt.Errorf("failed to build endpoint subsets from pods: %w", err)
			}
			
			modifiedEndpoints.Subsets = newSubsets
		}
	}

	// Log what we're about to do
	if dryRun {
		utils.DryRun(fmt.Sprintf("Would update endpoints for service '%s':", svcName))
		if len(modifiedEndpoints.Subsets) == 0 {
			utils.DryRun("  - Remove all endpoint subsets (no traffic routing)")
		} else {
			utils.DryRun(fmt.Sprintf("  - Replace with %d endpoint subset(s)", len(modifiedEndpoints.Subsets)))
			for i, subset := range modifiedEndpoints.Subsets {
				utils.DryRun(fmt.Sprintf("    Subset %d: %d addresses, %d ports", i+1, len(subset.Addresses), len(subset.Ports)))
			}
		}
		return nil
	}

	// Update the endpoints
	_, err = client.CoreV1().Endpoints(namespace).Update(context.TODO(), modifiedEndpoints, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update endpoints: %w", err)
	}

	utils.Info("Successfully updated service endpoints")
	return nil
}

// buildEndpointSubsetsFromPods creates endpoint subsets from pod information
func buildEndpointSubsetsFromPods(pods []corev1.Pod, servicePorts []corev1.ServicePort) ([]corev1.EndpointSubset, error) {
	if len(servicePorts) == 0 {
		return []corev1.EndpointSubset{}, nil
	}

	// Group pods by their readiness
	readyAddresses := []corev1.EndpointAddress{}
	notReadyAddresses := []corev1.EndpointAddress{}

	for _, pod := range pods {
		// Skip pods that are not running or have no IP
		if pod.Status.Phase != corev1.PodRunning || pod.Status.PodIP == "" {
			continue
		}

		address := corev1.EndpointAddress{
			IP: pod.Status.PodIP,
			TargetRef: &corev1.ObjectReference{
				Kind:      "Pod",
				Namespace: pod.Namespace,
				Name:      pod.Name,
				UID:       pod.UID,
			},
		}

		// Check if pod is ready
		isReady := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				isReady = true
				break
			}
		}

		if isReady {
			readyAddresses = append(readyAddresses, address)
		} else {
			notReadyAddresses = append(notReadyAddresses, address)
		}
	}

	// Build ports from service ports
	ports := []corev1.EndpointPort{}
	for _, servicePort := range servicePorts {
		port := corev1.EndpointPort{
			Name:     servicePort.Name,
			Port:     servicePort.Port,
			Protocol: servicePort.Protocol,
		}
		ports = append(ports, port)
	}

	// Create subsets
	subsets := []corev1.EndpointSubset{}

	// Add ready addresses subset if any
	if len(readyAddresses) > 0 {
		subsets = append(subsets, corev1.EndpointSubset{
			Addresses: readyAddresses,
			Ports:     ports,
		})
	}

	// Add not ready addresses subset if any
	if len(notReadyAddresses) > 0 {
		subsets = append(subsets, corev1.EndpointSubset{
			NotReadyAddresses: notReadyAddresses,
			Ports:             ports,
		})
	}

	return subsets, nil
}

// saveOriginalEndpoints saves the original endpoints to disk for rollback
func saveOriginalEndpoints(endpoints *corev1.Endpoints, svcName, namespace string) error {
	// Create rollback directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	rollbackDir := filepath.Join(homeDir, ".tipsy", "rollback")
	err = os.MkdirAll(rollbackDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create rollback directory: %w", err)
	}

	// Create filename
	filename := fmt.Sprintf("%s_%s.json", svcName, namespace)
	filepath := filepath.Join(rollbackDir, filename)

	// Marshal endpoints to JSON
	data, err := json.MarshalIndent(endpoints, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal endpoints to JSON: %w", err)
	}

	// Write to file
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write endpoints to file: %w", err)
	}

	utils.Info(fmt.Sprintf("Saved original endpoints to %s for rollback", filepath))
	return nil
}
