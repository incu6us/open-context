package fetcher

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/incu6us/open-context/cache"
	"github.com/incu6us/open-context/config"
)

const (
	defaultHTTPTimeout = 30 * time.Second
)

// BaseFetcher provides common functionality for all fetchers
type BaseFetcher struct {
	client *http.Client
	cache  *cache.Manager
}

// NewBaseFetcher creates a new base fetcher with common configuration
func NewBaseFetcher(cacheDir string) *BaseFetcher {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config, using defaults: %v\n", err)
		cfg = &config.Config{
			CacheTTL: config.Duration{Duration: 0},
		}
	}

	// Create cache manager
	cacheManager := cache.NewManager(cacheDir, cfg.CacheTTL.Duration)

	return &BaseFetcher{
		client: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		cache: cacheManager,
	}
}

// getClient returns the HTTP client
func (b *BaseFetcher) getClient() *http.Client {
	return b.client
}

// getCache returns the cache manager
func (b *BaseFetcher) getCache() *cache.Manager {
	return b.cache
}
