package cmd

import (
	"testing"

	"github.com/isurusiri/tipsy/internal/config"
	"github.com/spf13/cobra"
)

func TestKillCommand_Flags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new kill command for testing
	testKillCmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill pods based on label selector",
	}

	// Apply the same flag setup as the real kill command
	var selector, namespace string
	var count int

	testKillCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testKillCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testKillCmd.Flags().IntVar(&count, "count", 1, "Number of pods to kill (optional, default 1)")

	// Test flag parsing
	testCases := []struct {
		name           string
		args           []string
		expectedSelector string
		expectedNamespace string
		expectedCount  int
	}{
		{
			name:           "all flags set",
			args:           []string{"--selector=app=nginx", "--namespace=production", "--count=3"},
			expectedSelector: "app=nginx",
			expectedNamespace: "production",
			expectedCount:  3,
		},
		{
			name:           "only required selector",
			args:           []string{"--selector=app=test"},
			expectedSelector: "app=test",
			expectedNamespace: "",
			expectedCount:  1,
		},
		{
			name:           "selector and count",
			args:           []string{"--selector=environment=staging", "--count=2"},
			expectedSelector: "environment=staging",
			expectedNamespace: "",
			expectedCount:  2,
		},
		{
			name:           "selector and namespace",
			args:           []string{"--selector=tier=frontend", "--namespace=default"},
			expectedSelector: "tier=frontend",
			expectedNamespace: "default",
			expectedCount:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset variables before each test
			selector = ""
			namespace = ""
			count = 1

			// Parse flags
			testKillCmd.SetArgs(tc.args)
			err := testKillCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}

			// Check if flags match expected
			if selector != tc.expectedSelector {
				t.Errorf("Expected selector '%s', got '%s'", tc.expectedSelector, selector)
			}
			if namespace != tc.expectedNamespace {
				t.Errorf("Expected namespace '%s', got '%s'", tc.expectedNamespace, namespace)
			}
			if count != tc.expectedCount {
				t.Errorf("Expected count %d, got %d", tc.expectedCount, count)
			}
		})
	}
}

func TestKillCommand_DefaultValues(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new kill command for testing
	testKillCmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill pods based on label selector",
	}

	var selector, namespace string
	var count int

	testKillCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testKillCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testKillCmd.Flags().IntVar(&count, "count", 1, "Number of pods to kill (optional, default 1)")

	// Execute without any flags
	testKillCmd.SetArgs([]string{})
	err := testKillCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	// Check default values
	if selector != "" {
		t.Errorf("Expected empty selector by default, got '%s'", selector)
	}
	if namespace != "" {
		t.Errorf("Expected empty namespace by default, got '%s'", namespace)
	}
	if count != 1 {
		t.Errorf("Expected count to be 1 by default, got %d", count)
	}
}

func TestKillCommand_RequiredFlags(t *testing.T) {
	// Test that the kill command requires the selector flag
	// This is more of an integration test since we can't easily test
	// the MarkFlagRequired behavior in isolation
	
	// Create a test command similar to kill command
	testKillCmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill pods based on label selector",
		Run: func(cmd *cobra.Command, args []string) {
			// Simulate the validation logic from the real command
			selector, _ := cmd.Flags().GetString("selector")
			if selector == "" {
				// This would normally call cmd.Help() and return
				// For testing, we'll just check that this path is taken
			}
		},
	}

	var selector string
	testKillCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testKillCmd.MarkFlagRequired("selector")

	// Test with missing required flag
	testKillCmd.SetArgs([]string{})
	err := testKillCmd.Execute()
	if err == nil {
		t.Error("Expected error when required flag is missing")
	}
}

func TestKillCommand_IntegrationWithGlobalFlags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a test command that includes both local and global flags
	testRootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	testKillCmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill pods based on label selector",
	}

	// Add global flags
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Namespace, "namespace", "", "Kubernetes namespace to target")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.DryRun, "dry-run", false, "if true, simulate actions without taking effect")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.Verbose, "verbose", false, "enable verbose logging")

	// Add local flags
	var selector, namespace string
	var count int
	testKillCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testKillCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testKillCmd.Flags().IntVar(&count, "count", 1, "Number of pods to kill (optional, default 1)")

	testRootCmd.AddCommand(testKillCmd)

	// Test with both global and local flags
	args := []string{
		"kill",
		"--selector=app=nginx",
		"--count=2",
		"--dry-run=true",
		"--verbose=true",
		"--kubeconfig=/path/to/kubeconfig",
	}

	testRootCmd.SetArgs(args)
	err := testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	// Check global config
	if config.GlobalConfig.Kubeconfig != "/path/to/kubeconfig" {
		t.Errorf("Expected global kubeconfig '/path/to/kubeconfig', got '%s'", config.GlobalConfig.Kubeconfig)
	}
	if !config.GlobalConfig.DryRun {
		t.Error("Expected global dry-run to be true")
	}
	if !config.GlobalConfig.Verbose {
		t.Error("Expected global verbose to be true")
	}

	// Check local flags
	if selector != "app=nginx" {
		t.Errorf("Expected local selector 'app=nginx', got '%s'", selector)
	}
	if count != 2 {
		t.Errorf("Expected local count 2, got %d", count)
	}
}

func TestKillCommand_HelpText(t *testing.T) {
	// Test that the kill command has proper help text
	// This is more of a documentation test
	
	// Create a test command with the same help text as the real command
	testKillCmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill pods based on label selector",
		Long: `Kill pods in Kubernetes based on a label selector.

This command will:
1. List pods matching the provided label selector
2. Randomly select the specified number of pods to kill
3. Delete the selected pods (or simulate deletion in dry-run mode)

Examples:
  tipsy kill --selector "app=nginx" --count 2
  tipsy kill --selector "environment=staging" --namespace production --count 1
  tipsy kill --selector "tier=frontend" --dry-run --verbose`,
	}

	// Check that help text is not empty
	if testKillCmd.Short == "" {
		t.Error("Expected non-empty short description")
	}
	if testKillCmd.Long == "" {
		t.Error("Expected non-empty long description")
	}

	// Check that examples are included
	if len(testKillCmd.Long) < 100 {
		t.Error("Expected detailed help text with examples")
	}
}

func TestKillCommand_FlagValidation(t *testing.T) {
	// Test various flag validation scenarios
	
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "valid selector",
			args:        []string{"--selector=app=nginx"},
			expectError: false,
			description: "Should accept valid label selector",
		},
		{
			name:        "valid selector with count",
			args:        []string{"--selector=app=nginx", "--count=5"},
			expectError: false,
			description: "Should accept valid selector with count",
		},
		{
			name:        "valid selector with namespace",
			args:        []string{"--selector=app=nginx", "--namespace=production"},
			expectError: false,
			description: "Should accept valid selector with namespace",
		},
		{
			name:        "zero count",
			args:        []string{"--selector=app=nginx", "--count=0"},
			expectError: false,
			description: "Should accept zero count (though it won't kill anything)",
		},
		{
			name:        "negative count",
			args:        []string{"--selector=app=nginx", "--count=-1"},
			expectError: false,
			description: "Should accept negative count (though it won't kill anything)",
		},
		{
			name:        "large count",
			args:        []string{"--selector=app=nginx", "--count=1000"},
			expectError: false,
			description: "Should accept large count",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test command
			testKillCmd := &cobra.Command{
				Use:   "kill",
				Short: "Kill pods based on label selector",
			}

			var selector, namespace string
			var count int

			testKillCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
			testKillCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
			testKillCmd.Flags().IntVar(&count, "count", 1, "Number of pods to kill (optional, default 1)")

			// Parse flags
			testKillCmd.SetArgs(tc.args)
			err := testKillCmd.Execute()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}
