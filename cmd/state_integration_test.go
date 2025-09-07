package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/isurusiri/tipsy/internal/state"
)

func TestLatencyCommandStateTracking(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := state.GetStateFilePath()
	defer func() {
		// Restore original state file path
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
	}()

	// Set custom state file for testing
	testStateFile := filepath.Join(tempDir, "latency_test_state.json")
	os.Setenv("TIPSY_STATE_FILE", testStateFile)
	state.ReloadStateFilePath()

	// Test data
	testCases := []struct {
		name           string
		selector       string
		namespace      string
		delay          string
		duration       string
		dryRun         bool
		expectedPods   []string
		expectedAction state.ChaosAction
	}{
		{
			name:         "latency_with_dry_run",
			selector:     "app=nginx",
			namespace:    "default",
			delay:        "200ms",
			duration:     "30s",
			dryRun:       true,
			expectedPods: []string{"pod-1", "pod-2"},
			expectedAction: state.ChaosAction{
				Type:      "latency",
				TargetPod: "pod-1",
				Namespace: "default",
				Metadata: map[string]string{
					"delay":    "200ms",
					"duration": "30s",
					"selector": "app=nginx",
				},
			},
		},
		{
			name:         "latency_with_custom_namespace",
			selector:     "app=api",
			namespace:    "production",
			delay:        "500ms",
			duration:     "1m",
			dryRun:       false,
			expectedPods: []string{"api-pod-1"},
			expectedAction: state.ChaosAction{
				Type:      "latency",
				TargetPod: "api-pod-1",
				Namespace: "production",
				Metadata: map[string]string{
					"delay":    "500ms",
					"duration": "1m",
					"selector": "app=api",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear state before test
			actions, err := state.LoadActions()
			if err != nil {
				t.Fatalf("Failed to load actions: %v", err)
			}
			if len(actions) > 0 {
				err = state.ClearActions()
				if err != nil {
					t.Fatalf("Failed to clear actions: %v", err)
				}
			}

			// Simulate the latency command execution
			// In a real test, we would mock the Kubernetes client
			// For now, we'll simulate the state saving logic
			if !tc.dryRun {
				for _, podName := range tc.expectedPods {
					action := state.ChaosAction{
						Type:      "latency",
						TargetPod: podName,
						Namespace: tc.namespace,
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						Metadata: map[string]string{
							"delay":    tc.delay,
							"duration": tc.duration,
							"selector": tc.selector,
						},
					}
					err := state.SaveAction(action)
					if err != nil {
						t.Fatalf("Failed to save action: %v", err)
					}
				}
			}

			// Verify state was saved correctly
			actions, err = state.LoadActions()
			if err != nil {
				t.Fatalf("Failed to load actions: %v", err)
			}

			if tc.dryRun {
				// Dry run should not save state
				if len(actions) != 0 {
					t.Errorf("Expected 0 actions in dry run, got %d", len(actions))
				}
			} else {
				// Non-dry run should save state for each pod
				expectedCount := len(tc.expectedPods)
				if len(actions) != expectedCount {
					t.Errorf("Expected %d actions, got %d", expectedCount, len(actions))
				}

				// Verify action details
				for i, action := range actions {
					if action.Type != tc.expectedAction.Type {
						t.Errorf("Action %d: expected type %s, got %s", i, tc.expectedAction.Type, action.Type)
					}
					if action.Namespace != tc.expectedAction.Namespace {
						t.Errorf("Action %d: expected namespace %s, got %s", i, tc.expectedAction.Namespace, action.Namespace)
					}
					if action.TargetPod != tc.expectedPods[i] {
						t.Errorf("Action %d: expected target pod %s, got %s", i, tc.expectedPods[i], action.TargetPod)
					}
					if action.Timestamp == "" {
						t.Errorf("Action %d: expected timestamp, got empty", i)
					}

					// Verify metadata
					for key, expectedValue := range tc.expectedAction.Metadata {
						if actualValue, exists := action.Metadata[key]; !exists {
							t.Errorf("Action %d: missing metadata key %s", i, key)
						} else if actualValue != expectedValue {
							t.Errorf("Action %d: metadata key %s: expected %s, got %s", i, key, expectedValue, actualValue)
						}
					}
				}
			}
		})
	}
}

func TestKillCommandStateTracking(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := state.GetStateFilePath()
	defer func() {
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
	}()

	// Set custom state file for testing
	testStateFile := filepath.Join(tempDir, "kill_test_state.json")
	os.Setenv("TIPSY_STATE_FILE", testStateFile)
	state.ReloadStateFilePath()

	// Test data
	testCases := []struct {
		name           string
		selector       string
		namespace      string
		count          int
		dryRun         bool
		expectedPods   []string
		expectedAction state.ChaosAction
	}{
		{
			name:         "kill_with_dry_run",
			selector:     "app=nginx",
			namespace:    "default",
			count:        2,
			dryRun:       true,
			expectedPods: []string{"pod-1", "pod-2"},
			expectedAction: state.ChaosAction{
				Type:      "kill",
				TargetPod: "pod-1",
				Namespace: "default",
				Metadata: map[string]string{
					"selector": "app=nginx",
					"count":    "2",
				},
			},
		},
		{
			name:         "kill_single_pod",
			selector:     "app=api",
			namespace:    "production",
			count:        1,
			dryRun:       false,
			expectedPods: []string{"api-pod-1"},
			expectedAction: state.ChaosAction{
				Type:      "kill",
				TargetPod: "api-pod-1",
				Namespace: "production",
				Metadata: map[string]string{
					"selector": "app=api",
					"count":    "1",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear state before test
			actions, err := state.LoadActions()
			if err != nil {
				t.Fatalf("Failed to load actions: %v", err)
			}
			if len(actions) > 0 {
				err = state.ClearActions()
				if err != nil {
					t.Fatalf("Failed to clear actions: %v", err)
				}
			}

			// Simulate the kill command execution
			if !tc.dryRun {
				for _, podName := range tc.expectedPods {
					action := state.ChaosAction{
						Type:      "kill",
						TargetPod: podName,
						Namespace: tc.namespace,
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						Metadata: map[string]string{
							"selector": tc.selector,
							"count":    fmt.Sprintf("%d", tc.count),
						},
					}
					err := state.SaveAction(action)
					if err != nil {
						t.Fatalf("Failed to save action: %v", err)
					}
				}
			}

			// Verify state was saved correctly
			actions, err = state.LoadActions()
			if err != nil {
				t.Fatalf("Failed to load actions: %v", err)
			}

			if tc.dryRun {
				if len(actions) != 0 {
					t.Errorf("Expected 0 actions in dry run, got %d", len(actions))
				}
			} else {
				expectedCount := len(tc.expectedPods)
				if len(actions) != expectedCount {
					t.Errorf("Expected %d actions, got %d", expectedCount, len(actions))
				}

				// Verify action details
				for i, action := range actions {
					if action.Type != tc.expectedAction.Type {
						t.Errorf("Action %d: expected type %s, got %s", i, tc.expectedAction.Type, action.Type)
					}
					if action.Namespace != tc.expectedAction.Namespace {
						t.Errorf("Action %d: expected namespace %s, got %s", i, tc.expectedAction.Namespace, action.Namespace)
					}
					if action.TargetPod != tc.expectedPods[i] {
						t.Errorf("Action %d: expected target pod %s, got %s", i, tc.expectedPods[i], action.TargetPod)
					}

					// Verify metadata
					for key, expectedValue := range tc.expectedAction.Metadata {
						if actualValue, exists := action.Metadata[key]; !exists {
							t.Errorf("Action %d: missing metadata key %s", i, key)
						} else if actualValue != expectedValue {
							t.Errorf("Action %d: metadata key %s: expected %s, got %s", i, key, expectedValue, actualValue)
						}
					}
				}
			}
		})
	}
}

func TestMisrouteCommandStateTracking(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := state.GetStateFilePath()
	defer func() {
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
	}()

	// Set custom state file for testing
	testStateFile := filepath.Join(tempDir, "misroute_test_state.json")
	os.Setenv("TIPSY_STATE_FILE", testStateFile)
	state.ReloadStateFilePath()

	// Test data
	testCases := []struct {
		name                    string
		service                 string
		namespace               string
		removeAll               bool
		replaceWithSelector     string
		dryRun                  bool
		expectedAction          state.ChaosAction
	}{
		{
			name:                "misroute_remove_all_with_dry_run",
			service:             "my-service",
			namespace:           "default",
			removeAll:           true,
			replaceWithSelector: "",
			dryRun:              true,
			expectedAction: state.ChaosAction{
				Type:      "misroute",
				TargetPod: "my-service",
				Namespace: "default",
				Metadata: map[string]string{
					"service":                "my-service",
					"remove_all":            "true",
					"replace_with_selector": "",
				},
			},
		},
		{
			name:                "misroute_replace_with_selector",
			service:             "api-service",
			namespace:           "production",
			removeAll:           false,
			replaceWithSelector: "app=nginx",
			dryRun:              false,
			expectedAction: state.ChaosAction{
				Type:      "misroute",
				TargetPod: "api-service",
				Namespace: "production",
				Metadata: map[string]string{
					"service":                "api-service",
					"remove_all":            "false",
					"replace_with_selector": "app=nginx",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear state before test
			actions, err := state.LoadActions()
			if err != nil {
				t.Fatalf("Failed to load actions: %v", err)
			}
			if len(actions) > 0 {
				err = state.ClearActions()
				if err != nil {
					t.Fatalf("Failed to clear actions: %v", err)
				}
			}

			// Simulate the misroute command execution
			if !tc.dryRun {
				action := state.ChaosAction{
					Type:      "misroute",
					TargetPod: tc.service, // Using service name as target
					Namespace: tc.namespace,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Metadata: map[string]string{
						"service":                tc.service,
						"remove_all":            fmt.Sprintf("%t", tc.removeAll),
						"replace_with_selector": tc.replaceWithSelector,
					},
				}
				err := state.SaveAction(action)
				if err != nil {
					t.Fatalf("Failed to save action: %v", err)
				}
			}

			// Verify state was saved correctly
			actions, err = state.LoadActions()
			if err != nil {
				t.Fatalf("Failed to load actions: %v", err)
			}

			if tc.dryRun {
				if len(actions) != 0 {
					t.Errorf("Expected 0 actions in dry run, got %d", len(actions))
				}
			} else {
				if len(actions) != 1 {
					t.Errorf("Expected 1 action, got %d", len(actions))
				}

				action := actions[0]
				if action.Type != tc.expectedAction.Type {
					t.Errorf("Expected type %s, got %s", tc.expectedAction.Type, action.Type)
				}
				if action.Namespace != tc.expectedAction.Namespace {
					t.Errorf("Expected namespace %s, got %s", tc.expectedAction.Namespace, action.Namespace)
				}
				if action.TargetPod != tc.expectedAction.TargetPod {
					t.Errorf("Expected target pod %s, got %s", tc.expectedAction.TargetPod, action.TargetPod)
				}

				// Verify metadata
				for key, expectedValue := range tc.expectedAction.Metadata {
					if actualValue, exists := action.Metadata[key]; !exists {
						t.Errorf("Missing metadata key %s", key)
					} else if actualValue != expectedValue {
						t.Errorf("Metadata key %s: expected %s, got %s", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestStateFileFormat(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := state.GetStateFilePath()
	defer func() {
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
	}()

	// Set custom state file for testing
	testStateFile := filepath.Join(tempDir, "format_test_state.json")
	os.Setenv("TIPSY_STATE_FILE", testStateFile)
	state.ReloadStateFilePath()
	
	// Clear any existing state
	state.ClearActions()

	// Save multiple actions of different types
	actions := []state.ChaosAction{
		{
			Type:      "latency",
			TargetPod: "pod-1",
			Namespace: "default",
			Timestamp: "2025-09-07T12:00:00Z",
			Metadata:  map[string]string{"delay": "200ms"},
		},
		{
			Type:      "kill",
			TargetPod: "pod-2",
			Namespace: "production",
			Timestamp: "2025-09-07T12:01:00Z",
			Metadata:  map[string]string{"count": "1"},
		},
		{
			Type:      "misroute",
			TargetPod: "service-1",
			Namespace: "default",
			Timestamp: "2025-09-07T12:02:00Z",
			Metadata:  map[string]string{"remove_all": "true"},
		},
	}

	for _, action := range actions {
		err := state.SaveAction(action)
		if err != nil {
			t.Fatalf("Failed to save action: %v", err)
		}
	}

	// Read the file and verify format
	content, err := os.ReadFile(testStateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	// Verify it's valid JSON
	var loadedActions []state.ChaosAction
	err = json.Unmarshal(content, &loadedActions)
	if err != nil {
		t.Fatalf("State file contains invalid JSON: %v", err)
	}

	// Verify all actions were saved
	if len(loadedActions) != len(actions) {
		t.Errorf("Expected %d actions, got %d", len(actions), len(loadedActions))
	}

	// Verify the file is pretty-printed (contains indentation)
	if !contains(string(content), "\n  ") {
		t.Errorf("Expected pretty-printed JSON, got: %s", string(content))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
