package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient_WithEmptyPath(t *testing.T) {
	// Test that NewClient handles empty kubeconfig path gracefully
	// This will try in-cluster config first, then fall back to default location
	client, err := NewClient("")
	
	// We expect an error in test environment since we're not in a cluster
	// and may not have a default kubeconfig
	if err == nil {
		t.Log("NewClient succeeded with empty path - this is unexpected in test environment")
		// If it succeeds, we should have a valid client
		if client == nil {
			t.Error("Expected non-nil client when no error occurred")
		}
	} else {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestNewClient_WithInvalidPath(t *testing.T) {
	// Test with a non-existent kubeconfig path
	invalidPath := "/non/existent/path/kubeconfig"
	client, err := NewClient(invalidPath)
	
	if err == nil {
		t.Error("Expected error with invalid kubeconfig path")
	}
	
	if client != nil {
		t.Error("Expected nil client when error occurred")
	}
}

func TestGetConfig_WithEmptyPath(t *testing.T) {
	// Test GetConfig with empty path
	config, err := GetConfig("")
	
	// Similar to NewClient test, we expect an error in test environment
	if err == nil {
		t.Log("GetConfig succeeded with empty path - this is unexpected in test environment")
		if config == nil {
			t.Error("Expected non-nil config when no error occurred")
		}
	} else {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestGetConfig_WithInvalidPath(t *testing.T) {
	// Test GetConfig with invalid path
	invalidPath := "/non/existent/path/kubeconfig"
	config, err := GetConfig(invalidPath)
	
	if err == nil {
		t.Error("Expected error with invalid kubeconfig path")
	}
	
	if config != nil {
		t.Error("Expected nil config when error occurred")
	}
}

func TestIsInCluster(t *testing.T) {
	// Test IsInCluster function
	inCluster := IsInCluster()
	
	// In test environment, we expect to be outside cluster
	if inCluster {
		t.Log("Running in-cluster - this is unexpected in test environment")
	} else {
		t.Log("Running outside cluster - expected in test environment")
	}
}

func TestNewClient_Integration(t *testing.T) {
	// Integration test that tries to create a client
	// This test will help catch any import or dependency issues
	
	// Test with empty path (should try in-cluster then default)
	_, err := NewClient("")
	if err != nil {
		t.Logf("NewClient with empty path failed as expected: %v", err)
	}
	
	// Test with a dummy path (should fail)
	_, err = NewClient("/dummy/path")
	if err == nil {
		t.Error("Expected error with dummy kubeconfig path")
	}
}

func TestGetConfig_Integration(t *testing.T) {
	// Integration test for GetConfig function
	
	// Test with empty path
	_, err := GetConfig("")
	if err != nil {
		t.Logf("GetConfig with empty path failed as expected: %v", err)
	}
	
	// Test with dummy path
	_, err = GetConfig("/dummy/path")
	if err == nil {
		t.Error("Expected error with dummy kubeconfig path")
	}
}

// Helper function to create a temporary kubeconfig for testing
func createTempKubeconfig(t *testing.T) string {
	t.Helper()
	
	// Create a minimal valid kubeconfig content
	kubeconfigContent := `apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: https://test-server:6443
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	
	// Create temporary file
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
	
	err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary kubeconfig: %v", err)
	}
	
	return kubeconfigPath
}

func TestNewClient_WithValidTempKubeconfig(t *testing.T) {
	// Test with a valid temporary kubeconfig
	kubeconfigPath := createTempKubeconfig(t)
	defer os.Remove(kubeconfigPath)
	
	// This should still fail because the server doesn't exist,
	// but it should fail at connection time, not config parsing time
	client, err := NewClient(kubeconfigPath)
	
	// The config should be parsed successfully, but connection will fail
	// We expect an error because the server doesn't exist
	if err == nil {
		t.Log("NewClient succeeded with temp kubeconfig - this means config parsing worked")
		// If it succeeds, we should have a valid client
		if client == nil {
			t.Error("Expected non-nil client when no error occurred")
		}
	} else {
		t.Logf("NewClient failed as expected: %v", err)
		// The error should be about connection, not config parsing
		if client != nil {
			t.Error("Expected nil client when error occurred")
		}
	}
}
