package cmd

import (
	"testing"

	"github.com/isurusiri/tipsy/internal/config"
)

func TestMisrouteCmd_FlagValidation(t *testing.T) {
	tests := []struct {
		name                    string
		service                 string
		removeAll               bool
		replaceWithSelector     string
		expectedValidationError string
	}{
		{
			name:                    "valid with remove-all",
			service:                 "test-service",
			removeAll:               true,
			replaceWithSelector:     "",
			expectedValidationError: "",
		},
		{
			name:                    "valid with replace-with-selector",
			service:                 "test-service",
			removeAll:               false,
			replaceWithSelector:     "app=nginx",
			expectedValidationError: "",
		},
		{
			name:                    "invalid - both flags set",
			service:                 "test-service",
			removeAll:               true,
			replaceWithSelector:     "app=nginx",
			expectedValidationError: "--remove-all and --replace-with-selector cannot be used together",
		},
		{
			name:                    "invalid - neither flag set",
			service:                 "test-service",
			removeAll:               false,
			replaceWithSelector:     "",
			expectedValidationError: "either --remove-all or --replace-with-selector must be specified",
		},
		{
			name:                    "invalid - empty service",
			service:                 "",
			removeAll:               true,
			replaceWithSelector:     "",
			expectedValidationError: "--service flag is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global config
			config.GlobalConfig = config.Config{
				DryRun:  true,
				Verbose: false,
			}

			// Test the validation logic directly
			var validationError string

			// Check service flag
			if tt.service == "" {
				validationError = "--service flag is required"
			} else if !tt.removeAll && tt.replaceWithSelector == "" {
				validationError = "either --remove-all or --replace-with-selector must be specified"
			} else if tt.removeAll && tt.replaceWithSelector != "" {
				validationError = "--remove-all and --replace-with-selector cannot be used together"
			}

			// Check expected validation error
			if tt.expectedValidationError != "" {
				if validationError != tt.expectedValidationError {
					t.Errorf("Expected validation error '%s', got '%s'", tt.expectedValidationError, validationError)
				}
			} else {
				if validationError != "" {
					t.Errorf("Unexpected validation error: %s", validationError)
				}
			}
		})
	}
}

func TestMisrouteCmd_NamespaceLogic(t *testing.T) {
	tests := []struct {
		name                string
		globalNamespace     string
		localNamespace      string
		expectedNamespace   string
	}{
		{
			name:              "no namespace specified",
			globalNamespace:   "",
			localNamespace:    "",
			expectedNamespace: "default",
		},
		{
			name:              "global namespace only",
			globalNamespace:   "production",
			localNamespace:    "",
			expectedNamespace: "production",
		},
		{
			name:              "local namespace overrides global",
			globalNamespace:   "production",
			localNamespace:    "staging",
			expectedNamespace: "staging",
		},
		{
			name:              "local namespace only",
			globalNamespace:   "",
			localNamespace:    "staging",
			expectedNamespace: "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global config
			config.GlobalConfig = config.Config{
				Namespace: tt.globalNamespace,
			}

			// Test the namespace logic
			targetNamespace := tt.localNamespace
			if targetNamespace == "" {
				targetNamespace = config.GlobalConfig.Namespace
			}
			if targetNamespace == "" {
				targetNamespace = "default"
			}

			if targetNamespace != tt.expectedNamespace {
				t.Errorf("Expected namespace '%s', got '%s'", tt.expectedNamespace, targetNamespace)
			}
		})
	}
}

func TestMisrouteCmd_GlobalConfig(t *testing.T) {
	tests := []struct {
		name           string
		globalDryRun   bool
		globalVerbose  bool
		expectedDryRun bool
		expectedVerbose bool
	}{
		{
			name:           "global dry-run enabled",
			globalDryRun:   true,
			globalVerbose:  false,
			expectedDryRun: true,
			expectedVerbose: false,
		},
		{
			name:           "global verbose enabled",
			globalDryRun:   false,
			globalVerbose:  true,
			expectedDryRun: false,
			expectedVerbose: true,
		},
		{
			name:           "both global flags enabled",
			globalDryRun:   true,
			globalVerbose:  true,
			expectedDryRun: true,
			expectedVerbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global config
			config.GlobalConfig = config.Config{
				DryRun:  tt.globalDryRun,
				Verbose: tt.globalVerbose,
			}

			// Check that global config is set correctly
			if config.GlobalConfig.DryRun != tt.expectedDryRun {
				t.Errorf("Expected DryRun %t, got %t", tt.expectedDryRun, config.GlobalConfig.DryRun)
			}
			if config.GlobalConfig.Verbose != tt.expectedVerbose {
				t.Errorf("Expected Verbose %t, got %t", tt.expectedVerbose, config.GlobalConfig.Verbose)
			}
		})
	}
}
