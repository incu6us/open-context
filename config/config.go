package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	globalConfig *Config
	configOnce   sync.Once
)

// Config represents the application configuration
type Config struct {
	CacheTTL Duration `yaml:"cache_ttl"`
}

// Duration is a custom type that supports parsing durations like "7d", "1w", etc.
type Duration struct {
	time.Duration
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	duration, err := ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	d.Duration = duration
	return nil
}

// ParseDuration parses duration strings with support for days (d) and weeks (w)
// Supported formats: "24h", "7d", "1w", "30m", "0"
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0, nil
	}

	// Match pattern like "7d", "1w", "24h", etc.
	re := regexp.MustCompile(`^(\d+)([dhmsw])$`)
	matches := re.FindStringSubmatch(s)

	if matches == nil {
		// Try standard time.ParseDuration for formats like "24h", "30m"
		return time.ParseDuration(s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %w", err)
	}

	unit := matches[2]
	switch unit {
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "s":
		return time.Duration(value) * time.Second, nil
	default:
		return 0, fmt.Errorf("unsupported time unit: %s", unit)
	}
}

// Load loads the configuration from config.yaml
func Load() (*Config, error) {
	var err error
	configOnce.Do(func() {
		globalConfig, err = loadConfig()
	})
	return globalConfig, err
}

// loadConfig reads and parses the config.yaml file
func loadConfig() (*Config, error) {
	// Default configuration
	cfg := &Config{
		CacheTTL: Duration{Duration: 7 * 24 * time.Hour}, // 7 days default
	}

	// Look for config.yaml in current directory first
	configPath := "config.yaml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Try ~/.open-context/config.yaml
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil {
			configPath = filepath.Join(homeDir, ".open-context", "config.yaml")
			data, err = os.ReadFile(configPath)

			// If not found in home directory, create a default one
			if err != nil && os.IsNotExist(err) {
				if createErr := createDefaultConfig(configPath); createErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create default config: %v\n", createErr)
					fmt.Fprintf(os.Stderr, "Info: Using default configuration (cache_ttl: %v)\n", cfg.CacheTTL.Duration)
					return cfg, nil
				}
				// Try reading the newly created config
				data, err = os.ReadFile(configPath)
			}
		}

		// If still not found, use defaults
		if err != nil {
			fmt.Fprintf(os.Stderr, "Info: Using default configuration (cache_ttl: %v)\n", cfg.CacheTTL.Duration)
			return cfg, nil
		}
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Info: Loaded configuration from %s (cache_ttl: %v)\n", configPath, cfg.CacheTTL.Duration)
	return cfg, nil
}

// createDefaultConfig creates a default config.yaml file at the specified path
func createDefaultConfig(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config already exists, don't overwrite
	}

	// Default config content
	defaultConfig := `# Open Context Configuration File
# Generated automatically on first run

# Cache TTL (Time To Live) - How long cached data should be kept
# Supported formats:
#   - "0" or empty: No expiration (cache never expires)
#   - "7d": 7 days (default)
#   - "24h": 24 hours
#   - "30m": 30 minutes
#   - "1w": 1 week (same as 7d)
#
# If a cached file is older than cache_ttl, it will be automatically
# removed and fresh data will be fetched from the source.
#
# Examples:
#   cache_ttl: 0      # Never expire
#   cache_ttl: 7d     # Expire after 7 days (recommended)
#   cache_ttl: 24h    # Expire after 24 hours
#   cache_ttl: 1w     # Expire after 1 week

cache_ttl: 7d
`

	// Write config file
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Info: Created default config at %s\n", configPath)
	return nil
}

// GetCacheDir returns the cache directory path for open-context.
// It creates the directory if it doesn't exist.
// The cache directory is located at ~/.open-context/cache on all platforms.
func GetCacheDir() (string, error) {
	// Get user's home directory (works on macOS, Windows, Linux)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Build cache directory path: ~/.open-context/cache
	cacheDir := filepath.Join(homeDir, ".open-context", "cache")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// GetDataDir returns the data directory path within the cache.
// This is an alias for GetCacheDir for backward compatibility.
func GetDataDir() (string, error) {
	return GetCacheDir()
}
