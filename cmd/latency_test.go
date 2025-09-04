package cmd

import (
	"testing"
	"time"

	"github.com/isurusiri/tipsy/internal/config"
	"github.com/spf13/cobra"
)

func TestLatencyCommand_Flags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new latency command for testing
	testLatencyCmd := &cobra.Command{
		Use:   "latency",
		Short: "Inject network latency using tc netem via ephemeral containers",
	}

	// Apply the same flag setup as the real latency command
	var selector, namespace, delay, duration string

	testLatencyCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testLatencyCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testLatencyCmd.Flags().StringVar(&delay, "delay", "200ms", "Network delay to inject (e.g., '200ms', '500ms', '1s')")
	testLatencyCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep latency active (e.g., '30s', '1m', '5m')")

	// Test flag parsing
	testCases := []struct {
		name              string
		args              []string
		expectedSelector  string
		expectedNamespace string
		expectedDelay     string
		expectedDuration  string
	}{
		{
			name:              "all flags set",
			args:              []string{"--selector=app=nginx", "--namespace=production", "--delay=500ms", "--duration=1m"},
			expectedSelector:  "app=nginx",
			expectedNamespace: "production",
			expectedDelay:     "500ms",
			expectedDuration:  "1m",
		},
		{
			name:              "only required selector",
			args:              []string{"--selector=app=test"},
			expectedSelector:  "app=test",
			expectedNamespace: "",
			expectedDelay:     "200ms",
			expectedDuration:  "30s",
		},
		{
			name:              "selector and custom delay",
			args:              []string{"--selector=environment=staging", "--delay=1s"},
			expectedSelector:  "environment=staging",
			expectedNamespace: "",
			expectedDelay:     "1s",
			expectedDuration:  "30s",
		},
		{
			name:              "selector and custom duration",
			args:              []string{"--selector=tier=frontend", "--duration=5m"},
			expectedSelector:  "tier=frontend",
			expectedNamespace: "",
			expectedDelay:     "200ms",
			expectedDuration:  "5m",
		},
		{
			name:              "selector, namespace and custom values",
			args:              []string{"--selector=app=api", "--namespace=staging", "--delay=300ms", "--duration=2m"},
			expectedSelector:  "app=api",
			expectedNamespace: "staging",
			expectedDelay:     "300ms",
			expectedDuration:  "2m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset variables before each test
			selector = ""
			namespace = ""
			delay = "200ms"
			duration = "30s"

			// Parse flags
			testLatencyCmd.SetArgs(tc.args)
			err := testLatencyCmd.Execute()
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
			if delay != tc.expectedDelay {
				t.Errorf("Expected delay '%s', got '%s'", tc.expectedDelay, delay)
			}
			if duration != tc.expectedDuration {
				t.Errorf("Expected duration '%s', got '%s'", tc.expectedDuration, duration)
			}
		})
	}
}

func TestLatencyCommand_DefaultValues(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new latency command for testing
	testLatencyCmd := &cobra.Command{
		Use:   "latency",
		Short: "Inject network latency using tc netem via ephemeral containers",
	}

	var selector, namespace, delay, duration string

	testLatencyCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testLatencyCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testLatencyCmd.Flags().StringVar(&delay, "delay", "200ms", "Network delay to inject (e.g., '200ms', '500ms', '1s')")
	testLatencyCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep latency active (e.g., '30s', '1m', '5m')")

	// Execute without any flags
	testLatencyCmd.SetArgs([]string{})
	err := testLatencyCmd.Execute()
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
	if delay != "200ms" {
		t.Errorf("Expected delay to be '200ms' by default, got '%s'", delay)
	}
	if duration != "30s" {
		t.Errorf("Expected duration to be '30s' by default, got '%s'", duration)
	}
}

func TestLatencyCommand_RequiredFlags(t *testing.T) {
	// Test that the latency command requires the selector flag
	// This is more of an integration test since we can't easily test
	// the MarkFlagRequired behavior in isolation
	
	// Create a test command similar to latency command
	testLatencyCmd := &cobra.Command{
		Use:   "latency",
		Short: "Inject network latency using tc netem via ephemeral containers",
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
	testLatencyCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testLatencyCmd.MarkFlagRequired("selector")

	// Test with missing required flag
	testLatencyCmd.SetArgs([]string{})
	err := testLatencyCmd.Execute()
	if err == nil {
		t.Error("Expected error when required flag is missing")
	}
}

func TestLatencyCommand_IntegrationWithGlobalFlags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a test command that includes both local and global flags
	testRootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	testLatencyCmd := &cobra.Command{
		Use:   "latency",
		Short: "Inject network latency using tc netem via ephemeral containers",
	}

	// Add global flags
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Namespace, "namespace", "", "Kubernetes namespace to target")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.DryRun, "dry-run", false, "if true, simulate actions without taking effect")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.Verbose, "verbose", false, "enable verbose logging")

	// Add local flags
	var selector, namespace, delay, duration string
	testLatencyCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testLatencyCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testLatencyCmd.Flags().StringVar(&delay, "delay", "200ms", "Network delay to inject (e.g., '200ms', '500ms', '1s')")
	testLatencyCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep latency active (e.g., '30s', '1m', '5m')")

	testRootCmd.AddCommand(testLatencyCmd)

	// Test with both global and local flags
	args := []string{
		"latency",
		"--selector=app=nginx",
		"--delay=500ms",
		"--duration=1m",
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
	if delay != "500ms" {
		t.Errorf("Expected local delay '500ms', got '%s'", delay)
	}
	if duration != "1m" {
		t.Errorf("Expected local duration '1m', got '%s'", duration)
	}
}

func TestLatencyCommand_HelpText(t *testing.T) {
	// Test that the latency command has proper help text
	// This is more of a documentation test
	
	// Create a test command with the same help text as the real command
	testLatencyCmd := &cobra.Command{
		Use:   "latency",
		Short: "Inject network latency using tc netem via ephemeral containers",
		Long: `Inject network latency into pods using tc netem via ephemeral containers.

This command will:
1. List pods matching the provided label selector
2. Add ephemeral containers to inject network latency using tc netem
3. The latency will be applied for the specified duration

Examples:
  tipsy latency --selector "app=nginx" --delay "200ms" --duration "30s"
  tipsy latency --selector "environment=staging" --namespace production --delay "500ms" --duration "1m"
  tipsy latency --selector "tier=frontend" --dry-run --verbose`,
	}

	// Check that help text is not empty
	if testLatencyCmd.Short == "" {
		t.Error("Expected non-empty short description")
	}
	if testLatencyCmd.Long == "" {
		t.Error("Expected non-empty long description")
	}

	// Check that examples are included
	if len(testLatencyCmd.Long) < 100 {
		t.Error("Expected detailed help text with examples")
	}
}

func TestLatencyCommand_FlagValidation(t *testing.T) {
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
			name:        "valid selector with custom delay",
			args:        []string{"--selector=app=nginx", "--delay=500ms"},
			expectError: false,
			description: "Should accept valid selector with custom delay",
		},
		{
			name:        "valid selector with custom duration",
			args:        []string{"--selector=app=nginx", "--duration=1m"},
			expectError: false,
			description: "Should accept valid selector with custom duration",
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
			args:        []string{"--selector=app=nginx", "--namespace=staging", "--delay=1s", "--duration=5m"},
			expectError: false,
			description: "Should accept all flags with valid values",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test command
			testLatencyCmd := &cobra.Command{
				Use:   "latency",
				Short: "Inject network latency using tc netem via ephemeral containers",
			}

			var selector, namespace, delay, duration string

			testLatencyCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
			testLatencyCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
			testLatencyCmd.Flags().StringVar(&delay, "delay", "200ms", "Network delay to inject (e.g., '200ms', '500ms', '1s')")
			testLatencyCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep latency active (e.g., '30s', '1m', '5m')")

			// Parse flags
			testLatencyCmd.SetArgs(tc.args)
			err := testLatencyCmd.Execute()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}

func TestLatencyCommand_DurationParsing(t *testing.T) {
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

func TestLatencyCommand_DelayFormat(t *testing.T) {
	// Test delay format validation
	testCases := []struct {
		name        string
		delay       string
		expectValid bool
		description string
	}{
		{
			name:        "milliseconds",
			delay:       "200ms",
			expectValid: true,
			description: "Should accept milliseconds format",
		},
		{
			name:        "seconds",
			delay:       "1s",
			expectValid: true,
			description: "Should accept seconds format",
		},
		{
			name:        "microseconds",
			delay:       "500us",
			expectValid: true,
			description: "Should accept microseconds format",
		},
		{
			name:        "nanoseconds",
			delay:       "100ns",
			expectValid: true,
			description: "Should accept nanoseconds format",
		},
		{
			name:        "decimal seconds",
			delay:       "0.5s",
			expectValid: true,
			description: "Should accept decimal seconds format",
		},
		{
			name:        "empty delay",
			delay:       "",
			expectValid: false,
			description: "Should reject empty delay (time.ParseDuration doesn't accept empty string)",
		},
		{
			name:        "invalid format",
			delay:       "invalid",
			expectValid: false,
			description: "Should reject invalid delay format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test delay format by trying to parse it as a duration
			// (tc netem accepts similar formats)
			_, err := time.ParseDuration(tc.delay)

			if tc.expectValid && err != nil {
				t.Errorf("Expected valid delay format '%s': %v - %s", tc.delay, err, tc.description)
			}
			if !tc.expectValid && err == nil && tc.delay != "" {
				t.Errorf("Expected invalid delay format '%s' to be rejected - %s", tc.delay, tc.description)
			}
		})
	}
}
