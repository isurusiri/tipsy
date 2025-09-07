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
