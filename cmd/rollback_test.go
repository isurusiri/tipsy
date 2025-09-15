package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/isurusiri/tipsy/internal/state"
	"github.com/spf13/cobra"
)

func TestRollbackCmd(t *testing.T) {
	// Test that rollback command is properly registered
	rollbackCmd := rootCmd.Commands()
	found := false
	for _, cmd := range rollbackCmd {
		if cmd.Name() == "rollback" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("rollback command not found in root command")
	}
}

func TestRollbackCmdFlags(t *testing.T) {
	// Test that rollback command has the expected flags
	rollbackCmd := getRollbackCmd()
	
	// Test dry-run flag
	if rollbackCmd.Flag("dry-run") == nil {
		t.Error("rollback command missing --dry-run flag")
	}

	// Test type flag
	if rollbackCmd.Flag("type") == nil {
		t.Error("rollback command missing --type flag")
	}

	// Test pod flag
	if rollbackCmd.Flag("pod") == nil {
		t.Error("rollback command missing --pod flag")
	}
}

func TestRollbackCmdWithEmptyState(t *testing.T) {
	// Create a temporary state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "empty_state.json")
	
	// Set up empty state
	err := os.WriteFile(stateFile, []byte("[]"), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty state file: %v", err)
	}

	// Mock the state file path
	originalStateFile := state.GetStateFilePath()
	defer func() {
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
		state.ReloadStateFilePath()
	}()

	// Set temporary state file
	os.Setenv("TIPSY_STATE_FILE", stateFile)
	state.ReloadStateFilePath()

	// Test rollback command with empty state
	rollbackCmd := getRollbackCmd()
	rollbackCmd.SetArgs([]string{"--dry-run"})
	
	// Capture output
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err = rollbackCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	if err != nil {
		t.Errorf("Unexpected error with empty state: %v", err)
	}
}

func TestRollbackCmdWithTestState(t *testing.T) {
	// Create a temporary state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "test_state.json")
	
	// Set up test state
	actions := []state.ChaosAction{
		{
			Type:      "latency",
			TargetPod: "test-pod-1",
			Namespace: "default",
			Timestamp: "2023-01-01T00:00:00Z",
			Metadata:  map[string]string{"delay": "200ms"},
		},
		{
			Type:      "cpustress",
			TargetPod: "test-pod-2",
			Namespace: "default",
			Timestamp: "2023-01-01T00:01:00Z",
			Metadata:  map[string]string{"method": "stress-ng"},
		},
		{
			Type:      "misroute",
			TargetPod: "test-service",
			Namespace: "default",
			Timestamp: "2023-01-01T00:02:00Z",
			Metadata:  map[string]string{"service": "test-service"},
		},
	}

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
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
		state.ReloadStateFilePath()
	}()

	// Set temporary state file
	os.Setenv("TIPSY_STATE_FILE", stateFile)
	state.ReloadStateFilePath()

	// Test rollback command with test state
	rollbackCmd := getRollbackCmd()
	rollbackCmd.SetArgs([]string{"--dry-run"})
	
	// Capture output
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err = rollbackCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	if err != nil {
		t.Errorf("Unexpected error with test state: %v", err)
	}
}

func TestRollbackCmdWithTypeFilter(t *testing.T) {
	// Create a temporary state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "filtered_state.json")
	
	// Set up test state with multiple types
	actions := []state.ChaosAction{
		{
			Type:      "latency",
			TargetPod: "test-pod-1",
			Namespace: "default",
			Timestamp: "2023-01-01T00:00:00Z",
			Metadata:  map[string]string{"delay": "200ms"},
		},
		{
			Type:      "cpustress",
			TargetPod: "test-pod-2",
			Namespace: "default",
			Timestamp: "2023-01-01T00:01:00Z",
			Metadata:  map[string]string{"method": "stress-ng"},
		},
		{
			Type:      "latency",
			TargetPod: "test-pod-3",
			Namespace: "default",
			Timestamp: "2023-01-01T00:02:00Z",
			Metadata:  map[string]string{"delay": "300ms"},
		},
	}

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
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
		state.ReloadStateFilePath()
	}()

	// Set temporary state file
	os.Setenv("TIPSY_STATE_FILE", stateFile)
	state.ReloadStateFilePath()

	// Test rollback command with type filter
	rollbackCmd := getRollbackCmd()
	rollbackCmd.SetArgs([]string{"--type", "latency", "--dry-run"})
	
	// Capture output
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err = rollbackCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	if err != nil {
		t.Errorf("Unexpected error with type filter: %v", err)
	}
}

func TestRollbackCmdWithPodFilter(t *testing.T) {
	// Create a temporary state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "pod_filtered_state.json")
	
	// Set up test state
	actions := []state.ChaosAction{
		{
			Type:      "latency",
			TargetPod: "test-pod-1",
			Namespace: "default",
			Timestamp: "2023-01-01T00:00:00Z",
			Metadata:  map[string]string{"delay": "200ms"},
		},
		{
			Type:      "cpustress",
			TargetPod: "test-pod-2",
			Namespace: "default",
			Timestamp: "2023-01-01T00:01:00Z",
			Metadata:  map[string]string{"method": "stress-ng"},
		},
	}

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
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
		state.ReloadStateFilePath()
	}()

	// Set temporary state file
	os.Setenv("TIPSY_STATE_FILE", stateFile)
	state.ReloadStateFilePath()

	// Test rollback command with pod filter
	rollbackCmd := getRollbackCmd()
	rollbackCmd.SetArgs([]string{"--pod", "test-pod-1", "--dry-run"})
	
	// Capture output
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err = rollbackCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	if err != nil {
		t.Errorf("Unexpected error with pod filter: %v", err)
	}
}

func TestRollbackCmdWithNoMatches(t *testing.T) {
	// Create a temporary state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "no_matches_state.json")
	
	// Set up test state
	actions := []state.ChaosAction{
		{
			Type:      "latency",
			TargetPod: "test-pod-1",
			Namespace: "default",
			Timestamp: "2023-01-01T00:00:00Z",
			Metadata:  map[string]string{"delay": "200ms"},
		},
	}

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
		os.Setenv("TIPSY_STATE_FILE", originalStateFile)
		state.ReloadStateFilePath()
	}()

	// Set temporary state file
	os.Setenv("TIPSY_STATE_FILE", stateFile)
	state.ReloadStateFilePath()

	// Test rollback command with no matches
	rollbackCmd := getRollbackCmd()
	rollbackCmd.SetArgs([]string{"--type", "nonexistent", "--dry-run"})
	
	// Capture output
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err = rollbackCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	if err != nil {
		t.Errorf("Unexpected error with no matches: %v", err)
	}
}

func TestRollbackCmdHelp(t *testing.T) {
	// Test that help text is properly formatted
	rollbackCmd := getRollbackCmd()
	rollbackCmd.SetArgs([]string{"--help"})
	
	// Capture output
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := rollbackCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Help command should not return an error
	if err != nil {
		t.Errorf("Unexpected error with help command: %v", err)
	}
}

// Helper function to get the rollback command for testing
func getRollbackCmd() *cobra.Command {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "rollback" {
			return cmd
		}
	}
	return nil
}
