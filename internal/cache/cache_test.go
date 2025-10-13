package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheConfig(t *testing.T) {
	t.Run("default_config", func(t *testing.T) {
		config := DefaultCacheConfig()

		assert.Equal(t, 5*time.Minute, config.DefaultTTL)
		assert.Equal(t, 10*time.Minute, config.DIDTTL)
		assert.Equal(t, 5*time.Minute, config.HandleTTL)
		assert.Equal(t, 15*time.Minute, config.IdentityTTL)
		assert.Equal(t, 1000, config.MaxSize)
		assert.Equal(t, 100, config.MaxMemoryMB)
		assert.Equal(t, 1*time.Minute, config.CleanupInterval)
	})

	t.Run("production_config", func(t *testing.T) {
		config := ProductionCacheConfig()

		assert.Equal(t, 10*time.Minute, config.DefaultTTL)
		assert.Equal(t, 30*time.Minute, config.DIDTTL)
		assert.Equal(t, 10*time.Minute, config.HandleTTL)
		assert.Equal(t, 1*time.Hour, config.IdentityTTL)
		assert.Equal(t, 10000, config.MaxSize)
		assert.Equal(t, 500, config.MaxMemoryMB)
		assert.Equal(t, 5*time.Minute, config.CleanupInterval)
	})
}

func TestCacheService(t *testing.T) {
	logger := testutil.CreateMockLogger()
	config := &CacheConfig{
		DefaultTTL:      1 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 100 * time.Millisecond,
	}

	cache := NewCacheService(config, logger)
	defer cache.Close()

	t.Run("set_and_get", func(t *testing.T) {
		cache.Set("key1", "value1", 1*time.Minute)

		value, exists := cache.Get("key1")
		require.True(t, exists)
		assert.Equal(t, "value1", value)
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		value, exists := cache.Get("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("expired_entry", func(t *testing.T) {
		cache.Set("expired", "value", 1*time.Millisecond)
		time.Sleep(10 * time.Millisecond) // Wait for expiration

		value, exists := cache.Get("expired")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("delete", func(t *testing.T) {
		cache.Set("delete_me", "value", 1*time.Minute)
		cache.Delete("delete_me")

		value, exists := cache.Get("delete_me")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("clear", func(t *testing.T) {
		// Clear any existing items first
		cache.Clear()

		cache.Set("clear1", "value1", 1*time.Minute)
		cache.Set("clear2", "value2", 1*time.Minute)

		assert.Equal(t, 2, cache.Size())
		cache.Clear()
		assert.Equal(t, 0, cache.Size())
	})

	t.Run("size_limit", func(t *testing.T) {
		// Fill cache beyond max size
		for i := 0; i < 150; i++ {
			cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 1*time.Minute)
		}

		// Should not exceed max size
		assert.LessOrEqual(t, cache.Size(), config.MaxSize)
	})

	t.Run("stats", func(t *testing.T) {
		cache.Set("stats_key", "stats_value", 1*time.Minute)
		cache.Get("stats_key") // Access once

		stats := cache.Stats()
		assert.Contains(t, stats, "size")
		assert.Contains(t, stats, "max_size")
		assert.Contains(t, stats, "total_access")
		assert.Contains(t, stats, "expired_count")
	})

	t.Run("set_with_ttl", func(t *testing.T) {
		cache.SetWithTTL("ttl_key", "ttl_value", 1*time.Minute)

		value, exists := cache.Get("ttl_key")
		require.True(t, exists)
		assert.Equal(t, "ttl_value", value)
	})
}
