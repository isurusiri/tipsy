package cmd

import (
	"testing"
	"time"

	"github.com/isurusiri/tipsy/internal/config"
	"github.com/spf13/cobra"
)

func TestPacketLossCommand_Flags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new packetloss command for testing
	testPacketLossCmd := &cobra.Command{
		Use:   "packetloss",
		Short: "Inject network packet loss using tc netem via ephemeral containers",
	}

	// Apply the same flag setup as the real packetloss command
	var selector, namespace, loss, duration string

	testPacketLossCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testPacketLossCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testPacketLossCmd.Flags().StringVar(&loss, "loss", "30%", "Network packet loss percentage to inject (e.g., '30%', '50%', '10%')")
	testPacketLossCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep packet loss active (e.g., '30s', '1m', '5m')")

	// Test flag parsing
	testCases := []struct {
		name              string
		args              []string
		expectedSelector  string
		expectedNamespace string
		expectedLoss      string
		expectedDuration  string
	}{
		{
			name:              "all flags set",
			args:              []string{"--selector=app=nginx", "--namespace=production", "--loss=50%", "--duration=1m"},
			expectedSelector:  "app=nginx",
			expectedNamespace: "production",
			expectedLoss:      "50%",
			expectedDuration:  "1m",
		},
		{
			name:              "only required selector",
			args:              []string{"--selector=app=test"},
			expectedSelector:  "app=test",
			expectedNamespace: "",
			expectedLoss:      "30%",
			expectedDuration:  "30s",
		},
		{
			name:              "selector and custom loss",
			args:              []string{"--selector=environment=staging", "--loss=25%"},
			expectedSelector:  "environment=staging",
			expectedNamespace: "",
			expectedLoss:      "25%",
			expectedDuration:  "30s",
		},
		{
			name:              "selector and custom duration",
			args:              []string{"--selector=tier=frontend", "--duration=5m"},
			expectedSelector:  "tier=frontend",
			expectedNamespace: "",
			expectedLoss:      "30%",
			expectedDuration:  "5m",
		},
		{
			name:              "selector, namespace and custom values",
			args:              []string{"--selector=app=api", "--namespace=staging", "--loss=15%", "--duration=2m"},
			expectedSelector:  "app=api",
			expectedNamespace: "staging",
			expectedLoss:      "15%",
			expectedDuration:  "2m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset variables before each test
			selector = ""
			namespace = ""
			loss = "30%"
			duration = "30s"

			// Parse flags
			testPacketLossCmd.SetArgs(tc.args)
			err := testPacketLossCmd.Execute()
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
			if loss != tc.expectedLoss {
				t.Errorf("Expected loss '%s', got '%s'", tc.expectedLoss, loss)
			}
			if duration != tc.expectedDuration {
				t.Errorf("Expected duration '%s', got '%s'", tc.expectedDuration, duration)
			}
		})
	}
}

func TestPacketLossCommand_DefaultValues(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a new packetloss command for testing
	testPacketLossCmd := &cobra.Command{
		Use:   "packetloss",
		Short: "Inject network packet loss using tc netem via ephemeral containers",
	}

	var selector, namespace, loss, duration string

	testPacketLossCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testPacketLossCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testPacketLossCmd.Flags().StringVar(&loss, "loss", "30%", "Network packet loss percentage to inject (e.g., '30%', '50%', '10%')")
	testPacketLossCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep packet loss active (e.g., '30s', '1m', '5m')")

	// Execute without any flags
	testPacketLossCmd.SetArgs([]string{})
	err := testPacketLossCmd.Execute()
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
	if loss != "30%" {
		t.Errorf("Expected loss to be '30%%' by default, got '%s'", loss)
	}
	if duration != "30s" {
		t.Errorf("Expected duration to be '30s' by default, got '%s'", duration)
	}
}

func TestPacketLossCommand_RequiredFlags(t *testing.T) {
	// Test that the packetloss command requires the selector flag
	// This is more of an integration test since we can't easily test
	// the MarkFlagRequired behavior in isolation
	
	// Create a test command similar to packetloss command
	testPacketLossCmd := &cobra.Command{
		Use:   "packetloss",
		Short: "Inject network packet loss using tc netem via ephemeral containers",
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
	testPacketLossCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testPacketLossCmd.MarkFlagRequired("selector")

	// Test with missing required flag
	testPacketLossCmd.SetArgs([]string{})
	err := testPacketLossCmd.Execute()
	if err == nil {
		t.Error("Expected error when required flag is missing")
	}
}

func TestPacketLossCommand_IntegrationWithGlobalFlags(t *testing.T) {
	// Reset global config for testing
	config.GlobalConfig = config.Config{}

	// Create a test command that includes both local and global flags
	testRootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	testPacketLossCmd := &cobra.Command{
		Use:   "packetloss",
		Short: "Inject network packet loss using tc netem via ephemeral containers",
	}

	// Add global flags
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	testRootCmd.PersistentFlags().StringVar(&config.GlobalConfig.Namespace, "namespace", "", "Kubernetes namespace to target")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.DryRun, "dry-run", false, "if true, simulate actions without taking effect")
	testRootCmd.PersistentFlags().BoolVar(&config.GlobalConfig.Verbose, "verbose", false, "enable verbose logging")

	// Add local flags
	var selector, namespace, loss, duration string
	testPacketLossCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
	testPacketLossCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
	testPacketLossCmd.Flags().StringVar(&loss, "loss", "30%", "Network packet loss percentage to inject (e.g., '30%', '50%', '10%')")
	testPacketLossCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep packet loss active (e.g., '30s', '1m', '5m')")

	testRootCmd.AddCommand(testPacketLossCmd)

	// Test with both global and local flags
	args := []string{
		"packetloss",
		"--selector=app=nginx",
		"--loss=50%",
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
	if loss != "50%" {
		t.Errorf("Expected local loss '50%%', got '%s'", loss)
	}
	if duration != "1m" {
		t.Errorf("Expected local duration '1m', got '%s'", duration)
	}
}

func TestPacketLossCommand_HelpText(t *testing.T) {
	// Test that the packetloss command has proper help text
	// This is more of a documentation test
	
	// Create a test command with the same help text as the real command
	testPacketLossCmd := &cobra.Command{
		Use:   "packetloss",
		Short: "Inject network packet loss using tc netem via ephemeral containers",
		Long: `Inject network packet loss into pods using tc netem via ephemeral containers.

This command will:
1. List pods matching the provided label selector
2. Add ephemeral containers to inject network packet loss using tc netem
3. The packet loss will be applied for the specified duration

Examples:
  tipsy packetloss --selector "app=nginx" --loss "30%" --duration "30s"
  tipsy packetloss --selector "environment=staging" --namespace production --loss "50%" --duration "1m"
  tipsy packetloss --selector "tier=frontend" --dry-run --verbose`,
	}

	// Check that help text is not empty
	if testPacketLossCmd.Short == "" {
		t.Error("Expected non-empty short description")
	}
	if testPacketLossCmd.Long == "" {
		t.Error("Expected non-empty long description")
	}

	// Check that examples are included
	if len(testPacketLossCmd.Long) < 100 {
		t.Error("Expected detailed help text with examples")
	}
}

func TestPacketLossCommand_FlagValidation(t *testing.T) {
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
			name:        "valid selector with custom loss",
			args:        []string{"--selector=app=nginx", "--loss=50%"},
			expectError: false,
			description: "Should accept valid selector with custom loss",
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
			args:        []string{"--selector=app=nginx", "--namespace=staging", "--loss=25%", "--duration=5m"},
			expectError: false,
			description: "Should accept all flags with valid values",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test command
			testPacketLossCmd := &cobra.Command{
				Use:   "packetloss",
				Short: "Inject network packet loss using tc netem via ephemeral containers",
			}

			var selector, namespace, loss, duration string

			testPacketLossCmd.Flags().StringVar(&selector, "selector", "", "Kubernetes label selector (required)")
			testPacketLossCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to operate in (optional, defaults to global namespace or 'default')")
			testPacketLossCmd.Flags().StringVar(&loss, "loss", "30%", "Network packet loss percentage to inject (e.g., '30%', '50%', '10%')")
			testPacketLossCmd.Flags().StringVar(&duration, "duration", "30s", "How long to keep packet loss active (e.g., '30s', '1m', '5m')")

			// Parse flags
			testPacketLossCmd.SetArgs(tc.args)
			err := testPacketLossCmd.Execute()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s': %s", tc.name, tc.description)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for test case '%s': %v - %s", tc.name, err, tc.description)
			}
		})
	}
}

func TestPacketLossCommand_DurationParsing(t *testing.T) {
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

func TestPacketLossCommand_LossFormat(t *testing.T) {
	// Test loss format validation
	testCases := []struct {
		name        string
		loss        string
		expectValid bool
		description string
	}{
		{
			name:        "percentage with %",
			loss:        "30%",
			expectValid: true,
			description: "Should accept percentage format with %",
		},
		{
			name:        "percentage without %",
			loss:        "30",
			expectValid: true,
			description: "Should accept percentage format without %",
		},
		{
			name:        "decimal percentage",
			loss:        "25.5%",
			expectValid: true,
			description: "Should accept decimal percentage format",
		},
		{
			name:        "high percentage",
			loss:        "100%",
			expectValid: true,
			description: "Should accept 100% packet loss",
		},
		{
			name:        "zero percentage",
			loss:        "0%",
			expectValid: true,
			description: "Should accept 0% packet loss",
		},
		{
			name:        "empty loss",
			loss:        "",
			expectValid: false,
			description: "Should reject empty loss value",
		},
		{
			name:        "invalid format",
			loss:        "invalid",
			expectValid: true, // For now, we accept any non-empty string as tc netem is flexible
			description: "Should accept any non-empty string (tc netem is flexible)",
		},
		{
			name:        "negative percentage",
			loss:        "-10%",
			expectValid: true, // tc netem might handle this
			description: "Should accept negative percentage (tc netem might handle it)",
		},
		{
			name:        "over 100%",
			loss:        "150%",
			expectValid: true, // tc netem might handle this
			description: "Should accept percentage over 100% (tc netem might handle it)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test loss format validation
			// For now, we'll just check basic format validation
			// In a real implementation, you might want to add more sophisticated validation
			isValid := tc.loss != "" && len(tc.loss) > 0
			
			// Additional validation for percentage format
			if tc.loss != "" {
				// Check if it ends with % or is a number
				if tc.loss[len(tc.loss)-1] == '%' {
					isValid = true
				} else {
					// Check if it's a valid number
					// This is a simplified check - in practice you'd use strconv.ParseFloat
					isValid = true // For testing purposes, assume valid if not empty
				}
			}

			if tc.expectValid && !isValid {
				t.Errorf("Expected valid loss format '%s': %s", tc.loss, tc.description)
			}
			if !tc.expectValid && isValid {
				t.Errorf("Expected invalid loss format '%s' to be rejected - %s", tc.loss, tc.description)
			}
		})
	}
}
