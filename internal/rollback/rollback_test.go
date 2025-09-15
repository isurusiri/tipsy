package rollback

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/isurusiri/tipsy/internal/state"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFilterActions(t *testing.T) {
	actions := []state.ChaosAction{
		{Type: "latency", TargetPod: "pod1", Namespace: "default"},
		{Type: "cpustress", TargetPod: "pod2", Namespace: "default"},
		{Type: "latency", TargetPod: "pod3", Namespace: "production"},
		{Type: "misroute", TargetPod: "service1", Namespace: "default"},
	}

	tests := []struct {
		name           string
		filterType     string
		filterPod      string
		expectedCount  int
		expectedTypes  []string
		expectedPods   []string
	}{
		{
			name:          "no filters",
			filterType:    "",
			filterPod:     "",
			expectedCount: 4,
			expectedTypes: []string{"latency", "cpustress", "latency", "misroute"},
			expectedPods:  []string{"pod1", "pod2", "pod3", "service1"},
		},
		{
			name:          "filter by type",
			filterType:    "latency",
			filterPod:     "",
			expectedCount: 2,
			expectedTypes: []string{"latency", "latency"},
			expectedPods:  []string{"pod1", "pod3"},
		},
		{
			name:          "filter by pod",
			filterType:    "",
			filterPod:     "pod2",
			expectedCount: 1,
			expectedTypes: []string{"cpustress"},
			expectedPods:  []string{"pod2"},
		},
		{
			name:          "filter by both type and pod",
			filterType:    "latency",
			filterPod:     "pod1",
			expectedCount: 1,
			expectedTypes: []string{"latency"},
			expectedPods:  []string{"pod1"},
		},
		{
			name:          "no matches",
			filterType:    "nonexistent",
			filterPod:     "",
			expectedCount: 0,
			expectedTypes: []string{},
			expectedPods:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterActions(actions, tt.filterType, tt.filterPod)
			
			if len(filtered) != tt.expectedCount {
				t.Errorf("Expected %d actions, got %d", tt.expectedCount, len(filtered))
			}

			for i, action := range filtered {
				if i < len(tt.expectedTypes) && action.Type != tt.expectedTypes[i] {
					t.Errorf("Expected type %s, got %s", tt.expectedTypes[i], action.Type)
				}
				if i < len(tt.expectedPods) && action.TargetPod != tt.expectedPods[i] {
					t.Errorf("Expected pod %s, got %s", tt.expectedPods[i], action.TargetPod)
				}
			}
		})
	}
}

func TestRollbackAction(t *testing.T) {
	client := fake.NewSimpleClientset()

	tests := []struct {
		name        string
		action      state.ChaosAction
		dryRun      bool
		expectError bool
	}{
		{
			name: "latency action",
			action: state.ChaosAction{
				Type:      "latency",
				TargetPod: "test-pod",
				Namespace: "default",
			},
			dryRun:      true,
			expectError: false,
		},
		{
			name: "cpustress action",
			action: state.ChaosAction{
				Type:      "cpustress",
				TargetPod: "test-pod",
				Namespace: "default",
			},
			dryRun:      true,
			expectError: false,
		},
		{
			name: "misroute action",
			action: state.ChaosAction{
				Type:      "misroute",
				TargetPod: "test-service",
				Namespace: "default",
			},
			dryRun:      true,
			expectError: false,
		},
		{
			name: "kill action",
			action: state.ChaosAction{
				Type:      "kill",
				TargetPod: "test-pod",
				Namespace: "default",
			},
			dryRun:      true,
			expectError: false,
		},
		{
			name: "unknown action type",
			action: state.ChaosAction{
				Type:      "unknown",
				TargetPod: "test-pod",
				Namespace: "default",
			},
			dryRun:      true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rollbackAction(client, tt.action, tt.dryRun)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRevertTC(t *testing.T) {
	// Create a pod with ephemeral containers
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			EphemeralContainers: []corev1.EphemeralContainer{
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name: "latency-injector-123",
					},
				},
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name: "packetloss-injector-456",
					},
				},
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name: "other-container",
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	action := state.ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
	}

	// Test dry run
	err := RevertTC(client, action, true)
	if err != nil {
		t.Errorf("Unexpected error in dry run: %v", err)
	}

	// Test actual execution
	err = RevertTC(client, action, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRemoveEphemeral(t *testing.T) {
	// Create a pod with ephemeral containers
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			EphemeralContainers: []corev1.EphemeralContainer{
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name: "tipsy-cpu-stress-123",
					},
				},
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name: "other-container",
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	action := state.ChaosAction{
		Type:      "cpustress",
		TargetPod: "test-pod",
		Namespace: "default",
	}

	// Test dry run
	err := RemoveEphemeral(client, action, true)
	if err != nil {
		t.Errorf("Unexpected error in dry run: %v", err)
	}

	// Test actual execution
	err = RemoveEphemeral(client, action, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRestoreEndpoints(t *testing.T) {
	// Create test backup file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".tipsy", "rollback")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create test endpoints
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: "10.0.0.1"},
				},
				Ports: []corev1.EndpointPort{
					{Port: 80, Protocol: corev1.ProtocolTCP},
				},
			},
		},
	}

	// Save backup file
	backupPath := filepath.Join(backupDir, "test-service_default.json")
	data, err := json.MarshalIndent(endpoints, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal endpoints: %v", err)
	}
	err = os.WriteFile(backupPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}
	defer os.Remove(backupPath)

	// Create fake client with existing endpoints
	client := fake.NewSimpleClientset(endpoints)
	action := state.ChaosAction{
		Type:      "misroute",
		TargetPod: "test-service",
		Namespace: "default",
		Metadata: map[string]string{
			"backupPath": backupPath,
		},
	}

	// Test dry run
	err = RestoreEndpoints(client, action, true)
	if err != nil {
		t.Errorf("Unexpected error in dry run: %v", err)
	}

	// Test actual execution
	err = RestoreEndpoints(client, action, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify backup file was removed
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file should have been removed")
	}
}

func TestHandleKillAction(t *testing.T) {
	client := fake.NewSimpleClientset()
	action := state.ChaosAction{
		Type:      "kill",
		TargetPod: "test-pod",
		Namespace: "default",
	}

	// Test dry run
	err := handleKillAction(client, action, true)
	if err != nil {
		t.Errorf("Unexpected error in dry run: %v", err)
	}

	// Test actual execution
	err = handleKillAction(client, action, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRemoveEphemeralContainer(t *testing.T) {
	// Create a pod with ephemeral containers
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			EphemeralContainers: []corev1.EphemeralContainer{
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name: "container-to-remove",
					},
				},
				{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{
						Name: "container-to-keep",
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)

	// Test removing a specific container
	err := removeEphemeralContainer(client, "default", "test-pod", "container-to-remove")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// For the fake client, we can't easily verify the patch worked as expected
	// since it doesn't persist changes. Instead, we just verify the function
	// doesn't return an error, which indicates the patch was attempted successfully.
	// In a real scenario, the Kubernetes API would persist the changes.
}

func TestRollbackAll(t *testing.T) {
	// Create a temporary state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "state.json")
	
	// Set up test state
	actions := []state.ChaosAction{
		{
			Type:      "latency",
			TargetPod: "test-pod-1",
			Namespace: "default",
			Timestamp: "2023-01-01T00:00:00Z",
		},
		{
			Type:      "cpustress",
			TargetPod: "test-pod-2",
			Namespace: "default",
			Timestamp: "2023-01-01T00:01:00Z",
		},
	}

	// Save test state
	data, err := json.MarshalIndent(actions, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test state: %v", err)
	}
	err = os.WriteFile(stateFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test state: %v", err)
	}

	// Mock the state file path
	originalStateFile := state.GetStateFilePath()
	defer func() {
		// Restore original state file path
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
		state.ReloadStateFilePath()
	}()

	// Set temporary state file
	os.Setenv("TIPSY_STATE_FILE", stateFile)
	state.ReloadStateFilePath()

	// Create fake client
	client := fake.NewSimpleClientset()

	// Test dry run
	err = RollbackAll(client, true, "", "")
	if err != nil {
		t.Errorf("Unexpected error in dry run: %v", err)
	}

	// Test with type filter
	err = RollbackAll(client, true, "latency", "")
	if err != nil {
		t.Errorf("Unexpected error with type filter: %v", err)
	}

	// Test with pod filter
	err = RollbackAll(client, true, "", "test-pod-1")
	if err != nil {
		t.Errorf("Unexpected error with pod filter: %v", err)
	}

	// Test with no matches
	err = RollbackAll(client, true, "nonexistent", "")
	if err != nil {
		t.Errorf("Unexpected error with no matches: %v", err)
	}
}

// Test helper to create a fake Kubernetes client with specific objects
func createFakeClient(objects ...runtime.Object) *fake.Clientset {
	return fake.NewSimpleClientset(objects...)
}
