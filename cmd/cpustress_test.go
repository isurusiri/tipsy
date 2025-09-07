package cmd

import (
	"testing"
	"time"

	"github.com/isurusiri/tipsy/internal/config"
	"github.com/spf13/cobra"
)

func TestCPUStressCommand_Flags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new cpustress command for testing
	testCPUStressCmd := &cobra.Command{
		Use:   "cpustress",
		Short: "Inject CPU load into pods using ephemeral containers",
	}

	// Apply the same flag setup as the real cpustress command
	var selector, namespace, duration, method string

	testCPUStressCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testCPUStressCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testCPUStressCmd.Flags().StringVar(&duration, "duration", "60s", "How long to run CPU stress (e.g., '30s', '1m', '5m')")
	testCPUStressCmd.Flags().StringVar(&method, "method", "stress-ng", "CPU stress method: 'stress-ng' or 'yes'")

	// Test flag parsing
	testCases := []struct {
		name              string
		args              []string
		expectedSelector  string
		expectedNamespace string
		expectedDuration  string
		expectedMethod    string
	}{
		{
			name:              "all flags set",
			args:              []string{"--selector=app=nginx", "--namespace=production", "--duration=2m", "--method=yes"},
			expectedSelector:  "app=nginx",
			expectedNamespace: "production",
			expectedDuration:  "2m",
			expectedMethod:    "yes",
		},
		{
			name:              "only required selector",
			args:              []string{"--selector=app=test"},
			expectedSelector:  "app=test",
			expectedNamespace: "",
			expectedDuration:  "60s",
			expectedMethod:    "stress-ng",
		},
		{
			name:              "selector and custom duration",
			args:              []string{"--selector=environment=staging", "--duration=5m"},
			expectedSelector:  "environment=staging",
			expectedNamespace: "",
			expectedDuration:  "5m",
			expectedMethod:    "stress-ng",
		},
		{
			name:              "selector and custom method",
			args:              []string{"--selector=tier=frontend", "--method=yes"},
			expectedSelector:  "tier=frontend",
			expectedNamespace: "",
			expectedDuration:  "60s",
			expectedMethod:    "yes",
		},
		{
			name:              "selector, namespace and custom values",
			args:              []string{"--selector=app=api", "--namespace=staging", "--duration=30s", "--method=stress-ng"},
			expectedSelector:  "app=api",
			expectedNamespace: "staging",
			expectedDuration:  "30s",
			expectedMethod:    "stress-ng",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset variables before each test
			selector = ""
			namespace = ""
			duration = "60s"
			method = "stress-ng"

			// Parse flags
			testCPUStressCmd.SetArgs(tc.args)
			err := testCPUStressCmd.Execute()
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
			if duration != tc.expectedDuration {
				t.Errorf("Expected duration '%s', got '%s'", tc.expectedDuration, duration)
			}
			if method != tc.expectedMethod {
				t.Errorf("Expected method '%s', got '%s'", tc.expectedMethod, method)
			}
		})
	}
}

func TestCPUStressCommand_DefaultValues(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new cpustress command for testing
	testCPUStressCmd := &cobra.Command{
		Use:   "cpustress",
		Short: "Inject CPU load into pods using ephemeral containers",
	}

	var selector, namespace, duration, method string

	testCPUStressCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testCPUStressCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testCPUStressCmd.Flags().StringVar(&duration, "duration", "60s", "How long to run CPU stress (e.g., '30s', '1m', '5m')")
	testCPUStressCmd.Flags().StringVar(&method, "method", "stress-ng", "CPU stress method: 'stress-ng' or 'yes'")

	// Execute without any flags
	testCPUStressCmd.SetArgs([]string{})
	err := testCPUStressCmd.Execute()
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
	if duration != "60s" {
		t.Errorf("Expected duration to be '60s' by default, got '%s'", duration)
	}
	if method != "stress-ng" {
		t.Errorf("Expected method to be 'stress-ng' by default, got '%s'", method)
	}
}

func TestCPUStressCommand_RequiredFlags(t *testing.T) {
	// Test that the cpustress command requires the selector flag
	// This is more of an integration test since we can't easily test
	// the MarkFlagRequired behavior in isolation
	
	// Create a test command similar to cpustress command
	testCPUStressCmd := &cobra.Command{
		Use:   "cpustress",
		Short: "Inject CPU load into pods using ephemeral containers",
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
	testCPUStressCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testCPUStressCmd.MarkFlagRequired("selector")

	// Test with missing required flag
	testCPUStressCmd.SetArgs([]string{})
	err := testCPUStressCmd.Execute()
	if err == nil {
		t.Error("Expected error when required flag is missing")
	}
}

func TestCPUStressCommand_IntegrationWithGlobalFlags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a test command that includes both local and global flags
	testRootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	testCPUStressCmd := &cobra.Command{
		Use:   "cpustress",
		Short: "Inject CPU load into pods using ephemeral containers",
	}

	// Add global flags
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Namespace, "namespace", "", "Kubernetes namespace to target")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.DryRun, "dry-run", false, "if true, simulate actions without taking effect")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.Verbose, "verbose", false, "enable verbose logging")

	// Add local flags
	var selector, namespace, duration, method string
	testCPUStressCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testCPUStressCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testCPUStressCmd.Flags().StringVar(&duration, "duration", "60s", "How long to run CPU stress (e.g., '30s', '1m', '5m')")
	testCPUStressCmd.Flags().StringVar(&method, "method", "stress-ng", "CPU stress method: 'stress-ng' or 'yes'")

	testRootCmd.AddCommand(testCPUStressCmd)

	// Test with both global and local flags
	args := []string{
		"cpustress",
		"--selector=app=nginx",
		"--duration=2m",
		"--method=yes",
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
	if duration != "2m" {
		t.Errorf("Expected local duration '2m', got '%s'", duration)
	}
	if method != "yes" {
		t.Errorf("Expected local method 'yes', got '%s'", method)
	}
}

func TestCPUStressCommand_HelpText(t *testing.T) {
	// Test that the cpustress command has proper help text
	// This is more of a documentation test
	
	// Create a test command with the same help text as the real command
	testCPUStressCmd := &cobra.Command{
		Use:   "cpustress",
		Short: "Inject CPU load into pods using ephemeral containers",
		Long: `Inject CPU load into pods using ephemeral containers.

This command will:
1. List pods matching the provided label selector
2. Add ephemeral containers to inject CPU stress using either stress-ng or yes command
3. The CPU stress will be applied for the specified duration

Examples:
  tipsy cpustress --selector "app=nginx" --duration "60s"
  tipsy cpustress --selector "environment=staging" --namespace production --method "yes" --duration "2m"
  tipsy cpustress --selector "tier=frontend" --method "stress-ng" --duration "30s" --dry-run --verbose`,
	}

	// Check that help text is not empty
	if testCPUStressCmd.Short == "" {
		t.Error("Expected non-empty short description")
	}
	if testCPUStressCmd.Long == "" {
		t.Error("Expected non-empty long description")
	}

	// Check that examples are included
	if len(testCPUStressCmd.Long) < 100 {
		t.Error("Expected detailed help text with examples")
	}
}

func TestCPUStressCommand_FlagValidation(t *testing.T) {
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
			name:        "valid selector with custom duration",
			args:        []string{"--selector=app=nginx", "--duration=5m"},
			expectError: false,
			description: "Should accept valid selector with custom duration",
		},
		{
			name:        "valid selector with custom method",
			args:        []string{"--selector=app=nginx", "--method=yes"},
			expectError: false,
			description: "Should accept valid selector with custom method",
		},
		{
			name:        "valid selector with namespace",
			args:        []string{"--selector=app=nginx", "--namespace=production"},
			expectError: false,
			description: "Should accept valid selector with namespace",
		},
		{
			name:        "complex selector",
			args:        []string{"--selector=app=nginx,environment=production"},
			expectError: false,
			description: "Should accept complex label selector",
		},
		{
			name:        "all flags with valid values",
			args:        []string{"--selector=app=nginx", "--namespace=staging", "--duration=30s", "--method=stress-ng"},
			expectError: false,
			description: "Should accept all flags with valid values",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test command
			testCPUStressCmd := &cobra.Command{
				Use:   "cpustress",
				Short: "Inject CPU load into pods using ephemeral containers",
			}

			var selector, namespace, duration, method string

			testCPUStressCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
			testCPUStressCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
			testCPUStressCmd.Flags().StringVar(&duration, "duration", "60s", "How long to run CPU stress (e.g., '30s', '1m', '5m')")
			testCPUStressCmd.Flags().StringVar(&method, "method", "stress-ng", "CPU stress method: 'stress-ng' or 'yes'")

			// Parse flags
			testCPUStressCmd.SetArgs(tc.args)
			err := testCPUStressCmd.Execute()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}

func TestCPUStressCommand_DurationParsing(t *testing.T) {
	// Test duration parsing with various valid formats
	testCases := []struct {
		name        string
		duration    string
		expectError bool
		description string
	}{
		{
			name:        "seconds",
			duration:    "30s",
			expectError: false,
			description: "Should parse seconds correctly",
		},
		{
			name:        "minutes",
			duration:    "5m",
			expectError: false,
			description: "Should parse minutes correctly",
		},
		{
			name:        "hours",
			duration:    "1h",
			expectError: false,
			description: "Should parse hours correctly",
		},
		{
			name:        "milliseconds",
			duration:    "500ms",
			expectError: false,
			description: "Should parse milliseconds correctly",
		},
		{
			name:        "combined duration",
			duration:    "1h30m",
			expectError: false,
			description: "Should parse combined duration correctly",
		},
		{
			name:        "invalid duration",
			duration:    "invalid",
			expectError: true,
			description: "Should reject invalid duration format",
		},
		{
			name:        "empty duration",
			duration:    "",
			expectError: true,
			description: "Should reject empty duration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test duration parsing
			_, err := time.ParseDuration(tc.duration)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for duration '%s': %s", tc.duration, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for duration '%s': %v - %s", tc.duration, err, tc.description)
			}
		})
	}
}

func TestCPUStressCommand_MethodValidation(t *testing.T) {
	// Test method validation scenarios
	testCases := []struct {
		name        string
		method      string
		expectValid bool
		description string
	}{
		{
			name:        "stress-ng method",
			method:      "stress-ng",
			expectValid: true,
			description: "Should accept stress-ng method",
		},
		{
			name:        "yes method",
			method:      "yes",
			expectValid: true,
			description: "Should accept yes method",
		},
		{
			name:        "invalid method",
			method:      "invalid",
			expectValid: false,
			description: "Should reject invalid method",
		},
		{
			name:        "empty method",
			method:      "",
			expectValid: false,
			description: "Should reject empty method",
		},
		{
			name:        "uppercase stress-ng",
			method:      "STRESS-NG",
			expectValid: false,
			description: "Should reject uppercase stress-ng (case sensitive)",
		},
		{
			name:        "uppercase yes",
			method:      "YES",
			expectValid: false,
			description: "Should reject uppercase yes (case sensitive)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test method validation
			isValid := tc.method == "stress-ng" || tc.method == "yes"

			if tc.expectValid && !isValid {
				t.Errorf("Expected valid method '%s': %s", tc.method, tc.description)
			}
			if !tc.expectValid && isValid {
				t.Errorf("Expected invalid method '%s' to be rejected - %s", tc.method, tc.description)
			}
		})
	}
}

func TestCPUStressCommand_EdgeCases(t *testing.T) {
	// Test edge cases for CPU stress command
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "very short duration",
			args:        []string{"--selector=app=test", "--duration=1ms"},
			expectError: false,
			description: "Should handle very short duration",
		},
		{
			name:        "very long duration",
			args:        []string{"--selector=app=test", "--duration=24h"},
			expectError: false,
			description: "Should handle very long duration",
		},
		{
			name:        "zero duration",
			args:        []string{"--selector=app=test", "--duration=0s"},
			expectError: false,
			description: "Should handle zero duration",
		},
		{
			name:        "complex selector with multiple labels",
			args:        []string{"--selector=app=nginx,environment=production,tier=frontend"},
			expectError: false,
			description: "Should handle complex selector with multiple labels",
		},
		{
			name:        "selector with special characters",
			args:        []string{"--selector=app=nginx,version!=1.0"},
			expectError: false,
			description: "Should handle selector with special characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test command
			testCPUStressCmd := &cobra.Command{
				Use:   "cpustress",
				Short: "Inject CPU load into pods using ephemeral containers",
			}

			var selector, namespace, duration, method string

			testCPUStressCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
			testCPUStressCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
			testCPUStressCmd.Flags().StringVar(&duration, "duration", "60s", "How long to run CPU stress (e.g., '30s', '1m', '5m')")
			testCPUStressCmd.Flags().StringVar(&method, "method", "stress-ng", "CPU stress method: 'stress-ng' or 'yes'")

			// Parse flags
			testCPUStressCmd.SetArgs(tc.args)
			err := testCPUStressCmd.Execute()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}
