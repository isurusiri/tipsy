package config

import (
	"testing"
)

func TestConfig_Get(t *testing.T) {
	// Reset global config for testing
	GlobalConfig = Config{}

	// Test default values
	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() returned nil")
	}

	// Test that we get a pointer to the global config
	if cfg != &GlobalConfig {
		t.Error("Get() should return pointer to GlobalConfig")
	}
}

func TestConfig_Fields(t *testing.T) {
	// Reset global config for testing
	GlobalConfig = Config{
		Kubeconfig: "/path/to/kubeconfig",
		Namespace:  "default",
		DryRun:     true,
		Verbose:    false,
	}

	cfg := Get()

	// Test field values
	if cfg.Kubeconfig != "/path/to/kubeconfig" {
		t.Errorf("Expected Kubeconfig to be '/path/to/kubeconfig', got '%s'", cfg.Kubeconfig)
	}

	if cfg.Namespace != "default" {
		t.Errorf("Expected Namespace to be 'default', got '%s'", cfg.Namespace)
	}

	if !cfg.DryRun {
		t.Error("Expected DryRun to be true")
	}

	if cfg.Verbose {
		t.Error("Expected Verbose to be false")
	}
}

func TestConfig_EmptyValues(t *testing.T) {
	// Reset global config for testing
	GlobalConfig = Config{}

	cfg := Get()

	// Test empty string values
	if cfg.Kubeconfig != "" {
		t.Errorf("Expected empty Kubeconfig, got '%s'", cfg.Kubeconfig)
	}

	if cfg.Namespace != "" {
		t.Errorf("Expected empty Namespace, got '%s'", cfg.Namespace)
	}

	// Test boolean defaults
	if cfg.DryRun {
		t.Error("Expected DryRun to be false by default")
	}

	if cfg.Verbose {
		t.Error("Expected Verbose to be false by default")
	}
}
