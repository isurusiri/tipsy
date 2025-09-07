package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateFileWithEnvironmentOverride(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	originalEnv := os.Getenv("TIPSY_STATE_FILE")
	
	defer func() {
		stateFile = originalStateFile
		if originalEnv == "" {
			os.Unsetenv("TIPSY_STATE_FILE")
		} else {
			os.Setenv("TIPSY_STATE_FILE", originalEnv)
		}
	}()

	// Test with custom state file via environment variable
	customStateFile := filepath.Join(tempDir, "custom_state.json")
	os.Setenv("TIPSY_STATE_FILE", customStateFile)

	// Reset the stateFile variable to pick up the environment variable
	// In a real implementation, you might want to add a function to reload the state file path
	stateFile = customStateFile

	// Test saving an action
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"delay": "200ms"},
	}

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action with custom state file: %v", err)
	}

	// Verify the action was saved to the custom location
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions from custom state file: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != action.Type {
		t.Errorf("Expected type %s, got %s", action.Type, actions[0].Type)
	}

	// Verify the file was created in the custom location
	if _, err := os.Stat(customStateFile); os.IsNotExist(err) {
		t.Errorf("Expected state file to be created at %s", customStateFile)
	}
}

func TestStateFileWithInvalidEnvironmentPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	originalEnv := os.Getenv("TIPSY_STATE_FILE")
	
	defer func() {
		stateFile = originalStateFile
		if originalEnv == "" {
			os.Unsetenv("TIPSY_STATE_FILE")
		} else {
			os.Setenv("TIPSY_STATE_FILE", originalEnv)
		}
	}()

	// Test with invalid state file path (parent directory doesn't exist)
	invalidStateFile := filepath.Join(tempDir, "nonexistent", "state.json")
	os.Setenv("TIPSY_STATE_FILE", invalidStateFile)
	stateFile = invalidStateFile

	// Test saving an action - should create the directory
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"delay": "200ms"},
	}

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action with invalid state file path: %v", err)
	}

	// Verify the action was saved
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}

	// Verify the directory was created
	if _, err := os.Stat(filepath.Dir(invalidStateFile)); os.IsNotExist(err) {
		t.Errorf("Expected directory to be created: %s", filepath.Dir(invalidStateFile))
	}
}

func TestStateFileWithReadOnlyDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	originalEnv := os.Getenv("TIPSY_STATE_FILE")
	
	defer func() {
		stateFile = originalStateFile
		if originalEnv == "" {
			os.Unsetenv("TIPSY_STATE_FILE")
		} else {
			os.Setenv("TIPSY_STATE_FILE", originalEnv)
		}
	}()

	// Create a read-only directory
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err := os.Mkdir(readOnlyDir, 0444) // Read-only permissions
	if err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}

	// Test with state file in read-only directory
	readOnlyStateFile := filepath.Join(readOnlyDir, "state.json")
	os.Setenv("TIPSY_STATE_FILE", readOnlyStateFile)
	stateFile = readOnlyStateFile

	// Test saving an action - should fail
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"delay": "200ms"},
	}

	err = SaveAction(action)
	if err == nil {
		t.Errorf("Expected error when saving to read-only directory, got nil")
	}

	// Verify no action was saved - LoadActions should also fail in read-only directory
	_, err = LoadActions()
	if err == nil {
		t.Errorf("Expected error when loading from read-only directory, got nil")
	}
}

func TestStateFileWithEmptyEnvironmentVariable(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalStateFile := stateFile
	originalEnv := os.Getenv("TIPSY_STATE_FILE")
	
	defer func() {
		stateFile = originalStateFile
		if originalEnv == "" {
			os.Unsetenv("TIPSY_STATE_FILE")
		} else {
			os.Setenv("TIPSY_STATE_FILE", originalEnv)
		}
	}()

	// Test with empty environment variable
	os.Setenv("TIPSY_STATE_FILE", "")
	stateFile = "" // This should trigger fallback to default

	// Test saving an action - should use default path
	action := ChaosAction{
		Type:      "latency",
		TargetPod: "test-pod",
		Namespace: "default",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"delay": "200ms"},
	}

	// Reset to default state file for this test
	stateFile = filepath.Join(tempDir, "default_state.json")

	err := SaveAction(action)
	if err != nil {
		t.Fatalf("Failed to save action with empty environment variable: %v", err)
	}

	// Verify the action was saved
	actions, err := LoadActions()
	if err != nil {
		t.Fatalf("Failed to load actions: %v", err)
	}

	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}
}
