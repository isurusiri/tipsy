package config

// Config holds the global configuration for the CLI
type Config struct {
	Kubeconfig string
	Namespace  string
	DryRun     bool
	Verbose    bool
}

// Global config instance
var GlobalConfig Config

// Get returns the global configuration
func Get() *Config {
	return &GlobalConfig
} 