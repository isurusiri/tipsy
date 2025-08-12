package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient creates a new Kubernetes client with support for both in-cluster and out-of-cluster configuration.
// It follows this priority order:
// 1. If kubeconfigPath is provided, use that specific path
// 2. Try in-cluster configuration (when running inside a pod)
// 3. Fall back to default kubeconfig location (~/.kube/config)
func NewClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Priority 1: Use provided kubeconfig path
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig path %s: %w", kubeconfigPath, err)
		}
	} else {
		// Priority 2: Try in-cluster configuration
		config, err = rest.InClusterConfig()
		if err == nil {
			// Successfully got in-cluster config
		} else {
			// Priority 3: Fall back to default kubeconfig location
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %w", err)
			}

			defaultKubeconfig := filepath.Join(homeDir, ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
			if err != nil {
				return nil, fmt.Errorf("failed to build config from default kubeconfig location %s: %w", defaultKubeconfig, err)
			}
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// GetConfig returns the Kubernetes REST config using the same logic as NewClient.
// This is useful when you need the config object directly rather than a clientset.
func GetConfig(kubeconfigPath string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	// Priority 1: Use provided kubeconfig path
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig path %s: %w", kubeconfigPath, err)
		}
		return config, nil
	}

	// Priority 2: Try in-cluster configuration
	config, err = rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Priority 3: Fall back to default kubeconfig location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	defaultKubeconfig := filepath.Join(homeDir, ".kube", "config")
	config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from default kubeconfig location %s: %w", defaultKubeconfig, err)
	}

	return config, nil
}

// IsInCluster checks if the application is running inside a Kubernetes cluster.
func IsInCluster() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}
