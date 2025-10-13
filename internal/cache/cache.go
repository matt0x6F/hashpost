package cache

import (
	"log/slog"
	"time"
)

// CacheConfig holds configuration for the cache service
type CacheConfig struct {
	// Cache TTL settings
	DefaultTTL  time.Duration `yaml:"default_ttl"`  // Default TTL for cache entries
	DIDTTL      time.Duration `yaml:"did_ttl"`      // TTL for DID resolution cache
	HandleTTL   time.Duration `yaml:"handle_ttl"`   // TTL for handle resolution cache
	IdentityTTL time.Duration `yaml:"identity_ttl"` // TTL for identity document cache

	// Cache size limits
	MaxSize     int `yaml:"max_size"`      // Maximum number of entries in cache
	MaxMemoryMB int `yaml:"max_memory_mb"` // Maximum memory usage in MB

	// Cleanup settings
	CleanupInterval time.Duration `yaml:"cleanup_interval"` // How often to clean up expired entries
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:      5 * time.Minute,  // 5 minutes
		DIDTTL:          10 * time.Minute, // 10 minutes
		HandleTTL:       5 * time.Minute,  // 5 minutes
		IdentityTTL:     15 * time.Minute, // 15 minutes
		MaxSize:         1000,             // 1000 entries
		MaxMemoryMB:     100,              // 100 MB
		CleanupInterval: 1 * time.Minute,  // 1 minute
	}
}

// ProductionCacheConfig returns production-optimized cache configuration
func ProductionCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:      10 * time.Minute, // 10 minutes
		DIDTTL:          30 * time.Minute, // 30 minutes
		HandleTTL:       10 * time.Minute, // 10 minutes
		IdentityTTL:     1 * time.Hour,    // 1 hour
		MaxSize:         10000,            // 10000 entries
		MaxMemoryMB:     500,              // 500 MB
		CleanupInterval: 5 * time.Minute,  // 5 minutes
	}
}

// CacheEntry represents a cached entry with metadata
type CacheEntry struct {
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int         `json:"access_count"`
}

// IsExpired checks if the cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// CacheService provides caching functionality for DID resolution and other operations
type CacheService struct {
	config        *CacheConfig
	logger        *slog.Logger
	cache         map[string]*CacheEntry
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// NewCacheService creates a new cache service
func NewCacheService(config *CacheConfig, logger *slog.Logger) *CacheService {
	if config == nil {
		config = DefaultCacheConfig()
	}

	service := &CacheService{
		config:   config,
		logger:   logger,
		cache:    make(map[string]*CacheEntry),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	service.startCleanup()

	return service
}

// startCleanup starts the cleanup goroutine
func (cs *CacheService) startCleanup() {
	cs.cleanupTicker = time.NewTicker(cs.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-cs.cleanupTicker.C:
				cs.cleanup()
			case <-cs.stopChan:
				cs.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup removes expired entries from the cache
func (cs *CacheService) cleanup() {
	expiredKeys := make([]string, 0)

	for key, entry := range cs.cache {
		if entry.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	if len(expiredKeys) > 0 {
		for _, key := range expiredKeys {
			delete(cs.cache, key)
		}

		cs.logger.Debug("Cleaned up expired cache entries",
			"expired_count", len(expiredKeys),
			"remaining_count", len(cs.cache),
		)
	}
}

// Set stores a value in the cache with the specified TTL
func (cs *CacheService) Set(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = cs.config.DefaultTTL
	}

	entry := &CacheEntry{
		Value:       value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	cs.cache[key] = entry

	// Check if we need to evict entries due to size limit
	if len(cs.cache) > cs.config.MaxSize {
		cs.evictLRU()
	}
}

// Get retrieves a value from the cache
func (cs *CacheService) Get(key string) (interface{}, bool) {
	entry, exists := cs.cache[key]
	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		delete(cs.cache, key)
		return nil, false
	}

	// Update access count
	entry.AccessCount++

	return entry.Value, true
}

// evictLRU evicts the least recently used entry
func (cs *CacheService) evictLRU() {
	if len(cs.cache) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	for key, entry := range cs.cache {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(cs.cache, oldestKey)
		cs.logger.Debug("Evicted LRU cache entry", "key", oldestKey)
	}
}

// Delete removes a key from the cache
func (cs *CacheService) Delete(key string) {
	delete(cs.cache, key)
}

// Clear removes all entries from the cache
func (cs *CacheService) Clear() {
	cs.cache = make(map[string]*CacheEntry)
	cs.logger.Debug("Cleared all cache entries")
}

// Size returns the number of entries in the cache
func (cs *CacheService) Size() int {
	return len(cs.cache)
}

// Stats returns cache statistics
func (cs *CacheService) Stats() map[string]interface{} {
	totalAccess := 0
	expiredCount := 0

	for _, entry := range cs.cache {
		totalAccess += entry.AccessCount
		if entry.IsExpired() {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"size":          len(cs.cache),
		"max_size":      cs.config.MaxSize,
		"total_access":  totalAccess,
		"expired_count": expiredCount,
	}
}

// Close stops the cache service and cleans up resources
func (cs *CacheService) Close() {
	if cs.cleanupTicker != nil {
		cs.cleanupTicker.Stop()
	}
	close(cs.stopChan)
	cs.Clear()
}

// SetWithTTL stores a value in the cache with the specified TTL
func (cs *CacheService) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	cs.Set(key, value, ttl)
}
