package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ChaosAction represents a single chaos engineering action
type ChaosAction struct {
	Type      string            `json:"type"`
	TargetPod string            `json:"targetPod"`
	Namespace string            `json:"namespace"`
	Timestamp string            `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
}

var (
	stateFile string
	mu        sync.Mutex
)

// init sets up the state file path
func init() {
	// Check for environment variable override first
	if envFile := os.Getenv("TIPSY_STATE_FILE"); envFile != "" {
		stateFile = envFile
		return
	}
	
	// Default to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home directory, use current directory
		stateFile = ".tipsy/state.json"
		return
	}
	stateFile = filepath.Join(homeDir, ".tipsy", "state.json")
}

// SaveAction saves a chaos action to the state file
func SaveAction(action ChaosAction) error {
	mu.Lock()
	defer mu.Unlock()

	// Ensure the action has a timestamp if not set
	if action.Timestamp == "" {
		action.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// Load existing actions
	actions, err := loadActions()
	if err != nil {
		return fmt.Errorf("failed to load existing actions: %w", err)
	}

	// Append the new action
	actions = append(actions, action)

	// Ensure the directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal to pretty-printed JSON
	data, err := json.MarshalIndent(actions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal actions to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadActions loads all chaos actions from the state file
func LoadActions() ([]ChaosAction, error) {
	mu.Lock()
	defer mu.Unlock()

	return loadActions()
}

// loadActions is the internal function that loads actions without locking
// (used by SaveAction to avoid double-locking)
func loadActions() ([]ChaosAction, error) {
	// Check if file exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		// File doesn't exist, return empty slice
		return []ChaosAction{}, nil
	}

	// Read the file
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal JSON
	var actions []ChaosAction
	if len(data) == 0 {
		// Empty file, return empty slice
		return []ChaosAction{}, nil
	}

	if err := json.Unmarshal(data, &actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state file: %w", err)
	}

	return actions, nil
}

// GetStateFilePath returns the path to the state file
func GetStateFilePath() string {
	return stateFile
}

// ReloadStateFilePath reloads the state file path from environment variable
func ReloadStateFilePath() {
	// Check for environment variable override first
	if envFile := os.Getenv("TIPSY_STATE_FILE"); envFile != "" {
		stateFile = envFile
		return
	}
	
	// Default to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home directory, use current directory
		stateFile = ".tipsy/state.json"
		return
	}
	stateFile = filepath.Join(homeDir, ".tipsy", "state.json")
}

// ClearActions removes all actions from the state file
func ClearActions() error {
	mu.Lock()
	defer mu.Unlock()

	// Ensure the directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Write empty array to file
	data, err := json.MarshalIndent([]ChaosAction{}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal empty actions: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write empty state file: %w", err)
	}

	return nil
}
