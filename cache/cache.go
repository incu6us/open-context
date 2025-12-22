package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles caching operations with TTL support
type Manager struct {
	cacheDir string
	ttl      time.Duration
}

// NewManager creates a new cache manager
func NewManager(cacheDir string, ttl time.Duration) *Manager {
	return &Manager{
		cacheDir: cacheDir,
		ttl:      ttl,
	}
}

// GetCacheDir returns the cache directory path
func (m *Manager) GetCacheDir() string {
	return m.cacheDir
}

// GetTTL returns the cache TTL
func (m *Manager) GetTTL() time.Duration {
	return m.ttl
}

// IsExpired checks if a file at the given path has expired based on cache TTL
func (m *Manager) IsExpired(filePath string) (bool, error) {
	// If TTL is 0, cache never expires
	if m.ttl == 0 {
		return false, nil
	}

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil // File doesn't exist, treat as expired
		}
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if file is older than TTL
	age := time.Since(info.ModTime())
	return age > m.ttl, nil
}

// Load attempts to load data from cache. Returns true if loaded successfully, false if expired/not found
func (m *Manager) Load(filePath string, v interface{}) (bool, error) {
	// Check if cache exists and is not expired
	expired, err := m.IsExpired(filePath)
	if err != nil {
		return false, err
	}

	if expired {
		// Cache is expired, remove it
		if m.ttl > 0 {
			fmt.Printf("Cache expired (TTL: %v), removing: %s\n", m.ttl, filepath.Base(filePath))
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Warning: failed to remove expired cache file: %v\n", err)
			}
		}
		return false, nil
	}

	// Try to load from cache
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read cache file: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return true, nil
}

// Save saves data to cache
func (m *Manager) Save(filePath string, v interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Marshal data to JSON
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Remove removes a cache file
func (m *Manager) Remove(filePath string) error {
	return os.Remove(filePath)
}

// Clear removes all cache files in a directory
func (m *Manager) Clear(dirPath string) error {
	return os.RemoveAll(dirPath)
}

// GetFilePath builds a cache file path within the cache directory
func (m *Manager) GetFilePath(subpath ...string) string {
	parts := append([]string{m.cacheDir}, subpath...)
	return filepath.Join(parts...)
}
