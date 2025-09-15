package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSaveAndLoadActions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "test_state.json")

	// Test saving a single action
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata: map[string]string{
			"delay": "200ms",
		},
	}

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action: %v", err)
	}

	// Test loading actions
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != action.Type {
		t.Errorf("Expected type %s, got %s", action.Type, actions[0].Type)
	}

	if actions[0].TargetPod != action.TargetPod {
		t.Errorf("Expected target pod %s, got %s", action.TargetPod, actions[0].TargetPod)
	}

	// Test saving multiple actions
	action2 := ChaosAction{
		Type:      "misroute",
		TargetPod: "service-pod",
		Namespace: "production",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata: map[string]string{
			"service": "my-service",
		},
	}

	err = SaveAction(action2)
	if err != nil {
		t.Fatalf("Failed to save second action: %v", err)
	}

	// Load all actions
	actions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions after second save: %v", err)
	}

	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(actions))
	}
}

func TestLoadActionsEmptyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "empty_state.json")

	// Test loading from non-existent file
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions from non-existent file: %v", err)
	}

	if len(actions) != 0 {
		t.Fatalf("Expected 0 actions from non-existent file, got %d", len(actions))
	}

	// Create empty file
	err = os.WriteFile(stateFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Test loading from empty file
	actions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions from empty file: %v", err)
	}

	if len(actions) != 0 {
		t.Fatalf("Expected 0 actions from empty file, got %d", len(actions))
	}
}

func TestClearActions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "clear_test_state.json")

	// Save some actions first
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"delay": "200ms"},
	}

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action: %v", err)
	}

	// Verify action was saved
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action before clear, got %d", len(actions))
	}

	// Clear actions
	err = ClearActions()
	if err != nil {
		t.Fatalf("Failed to clear actions: %v", err)
	}

	// Verify actions were cleared
	actions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions after clear: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("Expected 0 actions after clear, got %d", len(actions))
	}
}

func TestSaveActionWithTimestamp(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "timestamp_test_state.json")

	// Test action without timestamp (should be auto-generated)
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Metadata:  map[string]string{"delay": "200ms"},
	}

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action: %v", err)
	}

	// Load and verify timestamp was added
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	if actions[0].Timestamp == "" {
		t.Fatalf("Expected timestamp to be auto-generated, got empty string")
	}

	// Verify timestamp is valid RFC3339 format
	_, err = time.Parse(time.RFC3339, actions[0].Timestamp)
	if err != nil {
		t.Fatalf("Invalid timestamp format: %v", err)
	}
}

func TestConcurrentSaveActions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "concurrent_test_state.json")

	// Test concurrent saves
	numGoroutines := 10
	actionsPerGoroutine := 5
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < actionsPerGoroutine; j++ {
				action := ChaosAction{
					Type:      "latency",
					TargetPod: fmt.Sprintf("pod-%d-%d", goroutineID, j),
					Namespace: "default",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Metadata: map[string]string{
						"goroutine": fmt.Sprintf("%d", goroutineID),
						"action":    fmt.Sprintf("%d", j),
					},
				}
				err := SaveAction(action)
				if err != nil {
					t.Errorf("Failed to save action in goroutine %d: %v", goroutineID, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all actions were saved
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}

	expectedCount := numGoroutines * actionsPerGoroutine
	if len(actions) != expectedCount {
		t.Fatalf("Expected %d actions, got %d", expectedCount, len(actions))
	}
}

func TestLoadActionsWithInvalidJSON(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "invalid_json_test_state.json")

	// Create a file with invalid JSON
	err := os.WriteFile(stateFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	// Try to load actions - should return an error
	_, err = LoadActions()
	if err == nil {
		t.Fatalf("Expected error when loading invalid JSON, got nil")
	}
}

func TestGetStateFilePath(t *testing.T) {
	// Test that GetStateFilePath returns a valid path
	path := GetStateFilePath()
	if path == "" {
		t.Fatalf("Expected non-empty state file path, got empty string")
	}

	// Verify it's a valid file path
	dir := filepath.Dir(path)
	if dir == "" {
		t.Fatalf("Expected valid directory in state file path: %s", path)
	}
}

func TestSaveActionWithComplexMetadata(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "complex_metadata_test_state.json")

	// Test action with complex metadata
	action := ChaosAction{
		Type:      "misroute",
		TargetPod: "my-service",
		Namespace: "production",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata: map[string]string{
			"service":                "my-service",
			"remove_all":            "true",
			"replace_with_selector": "app=nginx",
			"complex_value":         "value with spaces and special chars: !@#$%^&*()",
			"empty_value":           "",
			"numeric_value":         "123",
		},
	}

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action with complex metadata: %v", err)
	}

	// Load and verify
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	loadedAction := actions[0]
	if loadedAction.Type != action.Type {
		t.Errorf("Expected type %s, got %s", action.Type, loadedAction.Type)
	}

	if len(loadedAction.Metadata) != len(action.Metadata) {
		t.Errorf("Expected %d metadata entries, got %d", len(action.Metadata), len(loadedAction.Metadata))
	}

	// Verify specific metadata values
	for key, expectedValue := range action.Metadata {
		if actualValue, exists := loadedAction.Metadata[key]; !exists {
			t.Errorf("Missing metadata key: %s", key)
		} else if actualValue != expectedValue {
			t.Errorf("Metadata key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestStateFilePrettyPrinting(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "pretty_print_test_state.json")

	// Save an action
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"delay": "200ms"},
	}

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action: %v", err)
	}

	// Read the file and verify it's pretty-printed
	content, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	// Verify the JSON is pretty-printed (contains newlines and indentation)
	if !contains(string(content), "\n  ") {
		t.Errorf("Expected pretty-printed JSON with indentation, got: %s", string(content))
	}

	// Verify it's valid JSON
	var actions []ChaosAction
	err = json.Unmarshal(content, &actions)
	if err != nil {
		t.Fatalf("State file contains invalid JSON: %v", err)
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

func TestDeleteAction(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "delete_test_state.json")

	// Create test actions
	action1 := ChaosAction{
		Type:      "latency",
		TargetPod: "pod-1",
		Namespace: "default",
		Timestamp: "2023-01-01T00:00:00Z",
		Metadata:  map[string]string{"delay": "200ms"},
	}

	action2 := ChaosAction{
		Type:      "cpustress",
		TargetPod: "pod-2",
		Namespace: "default",
		Timestamp: "2023-01-01T00:01:00Z",
		Metadata:  map[string]string{"method": "stress-ng"},
	}

	action3 := ChaosAction{
		Type:      "misroute",
		TargetPod: "service-1",
		Namespace: "production",
		Timestamp: "2023-01-01T00:02:00Z",
		Metadata:  map[string]string{"service": "my-service"},
	}

	// Save all actions
	err := SaveAction(action1)
	if err != nil {
		t.Fatalf("Failed to save action1: %v", err)
	}

	err = SaveAction(action2)
	if err != nil {
		t.Fatalf("Failed to save action2: %v", err)
	}

	err = SaveAction(action3)
	if err != nil {
		t.Fatalf("Failed to save action3: %v", err)
	}

	// Verify all actions are saved
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}
	if len(actions) != 3 {
		t.Fatalf("Expected 3 actions, got %d", len(actions))
	}

	// Test deleting action2 (middle action)
	err = DeleteAction(action2)
	if err != nil {
		t.Fatalf("Failed to delete action2: %v", err)
	}

	// Verify action2 was deleted
	actions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions after delete: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions after delete, got %d", len(actions))
	}

	// Verify the remaining actions are correct
	remainingTypes := make([]string, len(actions))
	remainingPods := make([]string, len(actions))
	for i, action := range actions {
		remainingTypes[i] = action.Type
		remainingPods[i] = action.TargetPod
	}

	expectedTypes := []string{"latency", "misroute"}
	expectedPods := []string{"pod-1", "service-1"}

	if !slicesEqual(remainingTypes, expectedTypes) {
		t.Errorf("Expected types %v, got %v", expectedTypes, remainingTypes)
	}
	if !slicesEqual(remainingPods, expectedPods) {
		t.Errorf("Expected pods %v, got %v", expectedPods, remainingPods)
	}

	// Test deleting action1 (first action)
	err = DeleteAction(action1)
	if err != nil {
		t.Fatalf("Failed to delete action1: %v", err)
	}

	// Verify action1 was deleted
	actions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions after second delete: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action after second delete, got %d", len(actions))
	}

	if actions[0].Type != "misroute" || actions[0].TargetPod != "service-1" {
		t.Errorf("Expected remaining action to be misroute/service-1, got %s/%s", actions[0].Type, actions[0].TargetPod)
	}

	// Test deleting action3 (last action)
	err = DeleteAction(action3)
	if err != nil {
		t.Fatalf("Failed to delete action3: %v", err)
	}

	// Verify action3 was deleted
	actions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions after third delete: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("Expected 0 actions after third delete, got %d", len(actions))
	}
}

func TestDeleteActionNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "delete_not_found_test_state.json")

	// Create a non-existent action
	nonExistentAction := ChaosAction{
		Type:      "latency",
		TargetPod: "non-existent-pod",
		Namespace: "default",
		Timestamp: "2023-01-01T00:00:00Z",
		Metadata:  map[string]string{"delay": "200ms"},
	}

	// Try to delete non-existent action
	err := DeleteAction(nonExistentAction)
	if err == nil {
		t.Fatalf("Expected error when deleting non-existent action, got nil")
	}

	// Verify error message contains "not found"
	if !contains(err.Error(), "not found") {
		t.Errorf("Expected error message to contain 'not found', got: %v", err)
	}
}

func TestDeleteActionWithExactMatch(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "delete_exact_match_test_state.json")

	// Create two actions with same type and pod but different timestamps
	action1 := ChaosAction{
		Type:      "latency",
		TargetPod: "pod-1",
		Namespace: "default",
		Timestamp: "2023-01-01T00:00:00Z",
		Metadata:  map[string]string{"delay": "200ms"},
	}

	action2 := ChaosAction{
		Type:      "latency",
		TargetPod: "pod-1",
		Namespace: "default",
		Timestamp: "2023-01-01T00:01:00Z",
		Metadata:  map[string]string{"delay": "300ms"},
	}

	// Save both actions
	err := SaveAction(action1)
	if err != nil {
		t.Fatalf("Failed to save action1: %v", err)
	}

	err = SaveAction(action2)
	if err != nil {
		t.Fatalf("Failed to save action2: %v", err)
	}

	// Verify both actions are saved
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(actions))
	}

	// Delete action1 (should only delete the exact match)
	err = DeleteAction(action1)
	if err != nil {
		t.Fatalf("Failed to delete action1: %v", err)
	}

	// Verify only action1 was deleted
	actions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions after delete: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action after delete, got %d", len(actions))
	}

	// Verify the remaining action is action2
	if actions[0].Timestamp != action2.Timestamp {
		t.Errorf("Expected remaining action to have timestamp %s, got %s", action2.Timestamp, actions[0].Timestamp)
	}
}

func TestDeleteActionConcurrent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	defer func() {
		stateFile = originalStateFile
	}()

	// Set state file to temp directory
	stateFile = filepath.Join(tempDir, "delete_concurrent_test_state.json")

	// Create multiple actions
	numActions := 10
	actions := make([]ChaosAction, numActions)
	for i := 0; i < numActions; i++ {
		actions[i] = ChaosAction{
			Type:      "latency",
			TargetPod: fmt.Sprintf("pod-%d", i),
			Namespace: "default",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Metadata:  map[string]string{"index": fmt.Sprintf("%d", i)},
		}
		err := SaveAction(actions[i])
		if err != nil {
			t.Fatalf("Failed to save action %d: %v", i, err)
		}
	}

	// Verify all actions are saved
	loadedActions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}
	if len(loadedActions) != numActions {
		t.Fatalf("Expected %d actions, got %d", numActions, len(loadedActions))
	}

	// Concurrently delete half of the actions
	var wg sync.WaitGroup
	half := numActions / 2
	for i := 0; i < half; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			err := DeleteAction(actions[index])
			if err != nil {
				t.Errorf("Failed to delete action %d: %v", index, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify half of the actions were deleted
	loadedActions, err = LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions after concurrent delete: %v", err)
	}
	if len(loadedActions) != half {
		t.Fatalf("Expected %d actions after concurrent delete, got %d", half, len(loadedActions))
	}
}

// Helper function to check if two string slices are equal
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
