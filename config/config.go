package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for AutoAR
type Config struct {
	// Nuclei settings
	Nuclei NucleiConfig `yaml:"nuclei"`

	// Output settings
	Output OutputConfig `yaml:"output"`

	// Scope settings
	Scope ScopeConfig `yaml:"scope"`

	// Concurrency settings
	Concurrency int `yaml:"concurrency"`

	// Verbose mode
	Verbose bool `yaml:"verbose"`
}

// NucleiConfig holds nuclei-specific settings
type NucleiConfig struct {
	TemplatesPath string   `yaml:"templates_path"`
	Severity      []string `yaml:"severity"`
	RateLimit     int      `yaml:"rate_limit"`
	BulkSize      int      `yaml:"bulk_size"`
	Timeout       int      `yaml:"timeout"`
}

// OutputConfig holds output-related settings
type OutputConfig struct {
	Directory string `yaml:"directory"`
	Format    string `yaml:"format"` // json, text
	Silent    bool   `yaml:"silent"`
}

// ScopeConfig defines the recon scope
type ScopeConfig struct {
	Domains     []string `yaml:"domains"`
	Exclude     []string `yaml:"exclude"`
	Wildcard    bool     `yaml:"wildcard"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Concurrency: 10,
		Verbose:     false,
		Nuclei: NucleiConfig{
			Severity:  []string{"critical", "high", "medium"},
			RateLimit: 150,
			BulkSize:  25,
			Timeout:   5,
		},
		Output: OutputConfig{
			Directory: "./output",
			Format:    "text",
			Silent:    false,
		},
		Scope: ScopeConfig{
			Wildcard: true,
		},
	}
}

// LoadConfig reads a YAML config file and returns a Config.
// Falls back to defaults for missing values.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Allow env var overrides
	if v := os.Getenv("AUTOAR_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Concurrency = n
		}
	}
	if v := os.Getenv("AUTOAR_OUTPUT_DIR"); v != "" {
		cfg.Output.Directory = v
	}
	if v := os.Getenv("AUTOAR_VERBOSE"); v == "true" || v == "1" {
		cfg.Verbose = true
	}

	return cfg, nil
}
