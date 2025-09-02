package utils

import (
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/isurusiri/tipsy/internal/config"
)

func TestInfo(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test that Info function doesn't panic and executes successfully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Info function panicked: %v", r)
		}
	}()

	Info("test info message")
}

func TestWarn(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test that Warn function doesn't panic and executes successfully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Warn function panicked: %v", r)
		}
	}()

	Warn("test warning message")
}

func TestError(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test that Error function doesn't panic and executes successfully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Error function panicked: %v", r)
		}
	}()

	Error("test error message")
}

func TestDryRun_WhenDryRunDisabled(t *testing.T) {
	// Save original state
	originalDryRun := config.GlobalConfig.DryRun
	defer func() {
		config.GlobalConfig.DryRun = originalDryRun
	}()

	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Disable dry run
	config.GlobalConfig.DryRun = false

	// Test that DryRun function doesn't panic when disabled
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DryRun function panicked when disabled: %v", r)
		}
	}()

	DryRun("test dry run message")
}

func TestDryRun_WhenDryRunEnabled(t *testing.T) {
	// Save original state
	originalDryRun := config.GlobalConfig.DryRun
	defer func() {
		config.GlobalConfig.DryRun = originalDryRun
	}()

	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Enable dry run
	config.GlobalConfig.DryRun = true

	// Test that DryRun function doesn't panic and executes successfully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DryRun function panicked: %v", r)
		}
	}()

	DryRun("test dry run message")
}

func TestTimestampFormat(t *testing.T) {
	// Test that the timestamp format is valid by checking if it can be parsed
	now := time.Now()
	formatted := now.Format("15:04:05")
	
	// Try to parse the formatted timestamp back
	_, err := time.Parse("15:04:05", formatted)
	if err != nil {
		t.Errorf("Invalid timestamp format '%s': %v", formatted, err)
	}
}

func TestMultipleLogCalls(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test that multiple log calls don't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Multiple log calls panicked: %v", r)
		}
	}()

	Info("first message")
	Warn("second message")
	Error("third message")
}

func TestEmptyMessage(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test that empty message doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Empty message test panicked: %v", r)
		}
	}()

	Info("")
}

func TestSpecialCharacters(t *testing.T) {
	// Disable color for testing
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()

	specialMsg := "test with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
	
	// Test that special characters don't cause issues
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Special characters test panicked: %v", r)
		}
	}()

	Info(specialMsg)
}

// Benchmark tests
func BenchmarkInfo(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()
	
	for i := 0; i < b.N; i++ {
		Info("benchmark test message")
	}
}

func BenchmarkWarn(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()
	
	for i := 0; i < b.N; i++ {
		Warn("benchmark test message")
	}
}

func BenchmarkError(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()
	
	for i := 0; i < b.N; i++ {
		Error("benchmark test message")
	}
}

func BenchmarkDryRun(b *testing.B) {
	// Disable color for benchmarking
	originalNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = originalNoColor
	}()
	
	// Enable dry run for benchmarking
	originalDryRun := config.GlobalConfig.DryRun
	config.GlobalConfig.DryRun = true
	defer func() {
		config.GlobalConfig.DryRun = originalDryRun
	}()
	
	for i := 0; i < b.N; i++ {
		DryRun("benchmark test message")
	}
}
