package cmd

import (
	"testing"

	"github.com/isurusiri/tipsy/internal/config"
	"github.com/spf13/cobra"
)

func TestRootCommand_Flags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new root command for testing
	testRootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	// Apply the same flag setup as the real root command
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Namespace, "namespace", "", "Kubernetes namespace to target")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.DryRun, "dry-run", false, "if true, simulate actions without taking effect")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.Verbose, "verbose", false, "enable verbose logging")

	// Test flag parsing
	testCases := []struct {
		name           string
		args           []string
		expectedConfig config.Config
	}{
		{
			name: "all flags set",
			args: []string{
				"--kubeconfig=/path/to/kubeconfig",
				"--namespace=default",
				"--dry-run=true",
				"--verbose=true",
			},
			expectedConfig: config.Config{
				Kubeconfig: "/path/to/kubeconfig",
				Namespace:  "default",
				DryRun:     true,
				Verbose:    true,
			},
		},
		{
			name: "partial flags set",
			args: []string{
				"--namespace=test-namespace",
				"--dry-run=true",
			},
			expectedConfig: config.Config{
				Kubeconfig: "",
				Namespace:  "test-namespace",
				DryRun:     true,
				Verbose:    false,
			},
		},
		{
			name: "no flags set",
			args: []string{},
			expectedConfig: config.Config{
				Kubeconfig: "",
				Namespace:  "",
				DryRun:     false,
				Verbose:    false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset config before each test
			config.GlobalConfig = config.Config{}

			// Parse flags
			testRootCmd.SetArgs(tc.args)
			err := testRootCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}

			// Check if config matches expected
			cfg := config.Get()
			if cfg.Kubeconfig != tc.expectedConfig.Kubeconfig {
				t.Errorf("Expected Kubeconfig '%s', got '%s'", tc.expectedConfig.Kubeconfig, cfg.Kubeconfig)
			}
			if cfg.Namespace != tc.expectedConfig.Namespace {
				t.Errorf("Expected Namespace '%s', got '%s'", tc.expectedConfig.Namespace, cfg.Namespace)
			}
			if cfg.DryRun != tc.expectedConfig.DryRun {
				t.Errorf("Expected DryRun %t, got %t", tc.expectedConfig.DryRun, cfg.DryRun)
			}
			if cfg.Verbose != tc.expectedConfig.Verbose {
				t.Errorf("Expected Verbose %t, got %t", tc.expectedConfig.Verbose, cfg.Verbose)
			}
		})
	}
}

func TestRootCommand_DefaultValues(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new root command for testing
	testRootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	// Apply the same flag setup as the real root command
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Namespace, "namespace", "", "Kubernetes namespace to target")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.DryRun, "dry-run", false, "if true, simulate actions without taking effect")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.Verbose, "verbose", false, "enable verbose logging")

	// Execute without any flags
	testRootCmd.SetArgs([]string{})
	err := testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	// Check default values
	cfg := config.Get()
	if cfg.Kubeconfig != "" {
		t.Errorf("Expected empty Kubeconfig by default, got '%s'", cfg.Kubeconfig)
	}
	if cfg.Namespace != "" {
		t.Errorf("Expected empty Namespace by default, got '%s'", cfg.Namespace)
	}
	if cfg.DryRun {
		t.Error("Expected DryRun to be false by default")
	}
	if cfg.Verbose {
		t.Error("Expected Verbose to be false by default")
	}
}

func TestPrintConfig(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{
		Kubeconfig: "/test/kubeconfig",
		Namespace:  "test-namespace",
		DryRun:     true,
		Verbose:    true,
	}

	// Test that PrintConfig doesn't panic when verbose is true
	// Note: We can't easily capture stdout in tests, so we just ensure it doesn't crash
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintConfig panicked: %v", r)
		}
	}()

	PrintConfig()
}
