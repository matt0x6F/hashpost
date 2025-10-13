package database

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionConfig(t *testing.T) {
	t.Run("default_config", func(t *testing.T) {
		config := DefaultConnectionConfig()

		assert.Equal(t, int32(25), config.MaxConns)
		assert.Equal(t, int32(5), config.MinConns)
		assert.Equal(t, 30*time.Minute, config.MaxConnLifetime)
		assert.Equal(t, 5*time.Minute, config.MaxConnIdleTime)
		assert.Equal(t, 30*time.Second, config.ConnectTimeout)
		assert.Equal(t, 10*time.Second, config.AcquireTimeout)
		assert.Equal(t, 30*time.Second, config.HealthCheckPeriod)
		assert.Equal(t, 3, config.RetryAttempts)
		assert.Equal(t, 1*time.Second, config.RetryDelay)
	})

	t.Run("production_config", func(t *testing.T) {
		config := ProductionConnectionConfig()

		assert.Equal(t, int32(100), config.MaxConns)
		assert.Equal(t, int32(10), config.MinConns)
		assert.Equal(t, 1*time.Hour, config.MaxConnLifetime)
		assert.Equal(t, 10*time.Minute, config.MaxConnIdleTime)
		assert.Equal(t, 30*time.Second, config.ConnectTimeout)
		assert.Equal(t, 5*time.Second, config.AcquireTimeout)
		assert.Equal(t, 30*time.Second, config.HealthCheckPeriod)
		assert.Equal(t, 5, config.RetryAttempts)
		assert.Equal(t, 2*time.Second, config.RetryDelay)
	})

	t.Run("development_config", func(t *testing.T) {
		config := DevelopmentConnectionConfig()

		assert.Equal(t, int32(10), config.MaxConns)
		assert.Equal(t, int32(2), config.MinConns)
		assert.Equal(t, 10*time.Minute, config.MaxConnLifetime)
		assert.Equal(t, 1*time.Minute, config.MaxConnIdleTime)
		assert.Equal(t, 10*time.Second, config.ConnectTimeout)
		assert.Equal(t, 5*time.Second, config.AcquireTimeout)
		assert.Equal(t, 1*time.Minute, config.HealthCheckPeriod)
		assert.Equal(t, 2, config.RetryAttempts)
		assert.Equal(t, 500*time.Millisecond, config.RetryDelay)
	})
}

func TestConnectionManager(t *testing.T) {
	t.Run("new_connection_manager", func(t *testing.T) {
		logger := testutil.CreateMockLogger()
		config := DefaultConnectionConfig()

		manager := NewConnectionManager(config, logger)

		assert.NotNil(t, manager)
		assert.Equal(t, config, manager.config)
		assert.Equal(t, logger, manager.logger)
	})

	t.Run("new_connection_manager_with_nil_config", func(t *testing.T) {
		logger := testutil.CreateMockLogger()

		manager := NewConnectionManager(nil, logger)

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.config)
		assert.Equal(t, logger, manager.logger)
	})
}

func TestConnectionManager_CreatePool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database using pgtestdb
	pool := testutil.SetupPDSTestDB(t)

	// Get database URL from the test pool
	databaseURL := pool.Config().ConnString()

	logger := testutil.CreateMockLogger()
	config := DevelopmentConnectionConfig() // Use development config for testing
	manager := NewConnectionManager(config, logger)

	t.Run("create_pool_success", func(t *testing.T) {
		ctx := context.Background()

		pool, err := manager.CreatePool(ctx, databaseURL)
		require.NoError(t, err)
		require.NotNil(t, pool)

		// Test that the pool is working
		err = pool.Ping(ctx)
		require.NoError(t, err)

		// Close the pool
		pool.Close()
	})

	t.Run("create_pool_invalid_url", func(t *testing.T) {
		ctx := context.Background()
		invalidURL := "postgres://invalid:invalid@invalid:5432/invalid"

		pool, err := manager.CreatePool(ctx, invalidURL)
		require.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "failed to ping database")
	})
}

func TestConnectionManager_CreatePoolWithRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database using pgtestdb
	pool := testutil.SetupPDSTestDB(t)

	// Get database URL from the test pool
	databaseURL := pool.Config().ConnString()

	logger := testutil.CreateMockLogger()
	config := DevelopmentConnectionConfig()
	config.RetryAttempts = 2
	config.RetryDelay = 100 * time.Millisecond
	manager := NewConnectionManager(config, logger)

	t.Run("create_pool_with_retry_success", func(t *testing.T) {
		ctx := context.Background()

		pool, err := manager.CreatePoolWithRetry(ctx, databaseURL)
		require.NoError(t, err)
		require.NotNil(t, pool)

		// Test that the pool is working
		err = pool.Ping(ctx)
		require.NoError(t, err)

		// Close the pool
		pool.Close()
	})

	t.Run("create_pool_with_retry_failure", func(t *testing.T) {
		ctx := context.Background()
		invalidURL := "postgres://invalid:invalid@invalid:5432/invalid"

		pool, err := manager.CreatePoolWithRetry(ctx, invalidURL)
		require.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "failed to create database connection pool after 2 attempts")
	})
}

func TestConnectionManager_GetPoolStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database using pgtestdb
	pool := testutil.SetupPDSTestDB(t)

	logger := testutil.CreateMockLogger()
	config := DevelopmentConnectionConfig()
	manager := NewConnectionManager(config, logger)

	t.Run("get_pool_stats", func(t *testing.T) {
		stats := manager.GetPoolStats(pool)

		assert.Contains(t, stats, "acquire_count")
		assert.Contains(t, stats, "acquire_duration")
		assert.Contains(t, stats, "acquired_conns")
		assert.Contains(t, stats, "canceled_acquire_count")
		assert.Contains(t, stats, "constructing_conns")
		assert.Contains(t, stats, "empty_acquire_count")
		assert.Contains(t, stats, "idle_conns")
		assert.Contains(t, stats, "max_conns")
		assert.Contains(t, stats, "total_conns")

		// Verify that stats are non-negative
		for key, value := range stats {
			if key == "acquire_duration" {
				// Duration can be 0
				continue
			}
			// Check if the value is a numeric type that can be compared
			switch v := value.(type) {
			case int:
				assert.GreaterOrEqual(t, v, 0, "stat %s should be non-negative", key)
			case int64:
				assert.GreaterOrEqual(t, v, int64(0), "stat %s should be non-negative", key)
			case float64:
				assert.GreaterOrEqual(t, v, 0.0, "stat %s should be non-negative", key)
			case time.Duration:
				assert.GreaterOrEqual(t, v, time.Duration(0), "stat %s should be non-negative", key)
			}
		}
	})
}

func TestConnectionManager_LogPoolStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database using pgtestdb
	pool := testutil.SetupPDSTestDB(t)

	logger := testutil.CreateMockLogger()
	config := DevelopmentConnectionConfig()
	manager := NewConnectionManager(config, logger)

	t.Run("log_pool_stats", func(t *testing.T) {
		// This test just ensures the method doesn't panic
		// In a real test, we would capture the log output and verify it
		assert.NotPanics(t, func() {
			manager.LogPoolStats(pool)
		})
	})
}

func TestConnectionManager_ClosePool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database using pgtestdb
	pool := testutil.SetupPDSTestDB(t)

	logger := testutil.CreateMockLogger()
	config := DevelopmentConnectionConfig()
	manager := NewConnectionManager(config, logger)

	t.Run("close_pool", func(t *testing.T) {
		// This test just ensures the method doesn't panic
		assert.NotPanics(t, func() {
			manager.ClosePool(pool)
		})
	})

	t.Run("close_nil_pool", func(t *testing.T) {
		// This test ensures the method handles nil pools gracefully
		assert.NotPanics(t, func() {
			manager.ClosePool(nil)
		})
	})
}
