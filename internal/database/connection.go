package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectionConfig holds configuration for database connection pooling
type ConnectionConfig struct {
	// Connection pool settings
	MaxConns        int32         `yaml:"max_conns"`          // Maximum number of connections in the pool
	MinConns        int32         `yaml:"min_conns"`          // Minimum number of connections in the pool
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`  // Maximum lifetime of a connection
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"` // Maximum idle time of a connection

	// Connection timeout settings
	ConnectTimeout time.Duration `yaml:"connect_timeout"` // Connection timeout
	AcquireTimeout time.Duration `yaml:"acquire_timeout"` // Acquire connection timeout

	// Health check settings
	HealthCheckPeriod time.Duration `yaml:"health_check_period"` // Health check period

	// Retry settings
	RetryAttempts int           `yaml:"retry_attempts"` // Number of retry attempts
	RetryDelay    time.Duration `yaml:"retry_delay"`    // Delay between retries
}

// DefaultConnectionConfig returns default connection pool configuration
func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		MaxConns:          25,               // Maximum 25 connections
		MinConns:          5,                // Minimum 5 connections
		MaxConnLifetime:   30 * time.Minute, // 30 minutes
		MaxConnIdleTime:   5 * time.Minute,  // 5 minutes
		ConnectTimeout:    30 * time.Second, // 30 seconds
		AcquireTimeout:    10 * time.Second, // 10 seconds
		HealthCheckPeriod: 30 * time.Second, // 30 seconds
		RetryAttempts:     3,                // 3 retry attempts
		RetryDelay:        1 * time.Second,  // 1 second delay
	}
}

// ProductionConnectionConfig returns production-optimized connection pool configuration
func ProductionConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		MaxConns:          100,              // Maximum 100 connections
		MinConns:          10,               // Minimum 10 connections
		MaxConnLifetime:   1 * time.Hour,    // 1 hour
		MaxConnIdleTime:   10 * time.Minute, // 10 minutes
		ConnectTimeout:    30 * time.Second, // 30 seconds
		AcquireTimeout:    5 * time.Second,  // 5 seconds
		HealthCheckPeriod: 30 * time.Second, // 30 seconds
		RetryAttempts:     5,                // 5 retry attempts
		RetryDelay:        2 * time.Second,  // 2 second delay
	}
}

// DevelopmentConnectionConfig returns development-optimized connection pool configuration
func DevelopmentConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		MaxConns:          10,                     // Maximum 10 connections
		MinConns:          2,                      // Minimum 2 connections
		MaxConnLifetime:   10 * time.Minute,       // 10 minutes
		MaxConnIdleTime:   1 * time.Minute,        // 1 minute
		ConnectTimeout:    10 * time.Second,       // 10 seconds
		AcquireTimeout:    5 * time.Second,        // 5 seconds
		HealthCheckPeriod: 1 * time.Minute,        // 1 minute
		RetryAttempts:     2,                      // 2 retry attempts
		RetryDelay:        500 * time.Millisecond, // 500ms delay
	}
}

// ConnectionManager manages database connections with proper pooling configuration
type ConnectionManager struct {
	config *ConnectionConfig
	logger *slog.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(config *ConnectionConfig, logger *slog.Logger) *ConnectionManager {
	if config == nil {
		config = DefaultConnectionConfig()
	}

	return &ConnectionManager{
		config: config,
		logger: logger,
	}
}

// CreatePool creates a new database connection pool with the configured settings
func (cm *ConnectionManager) CreatePool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	// Parse the database URL to get connection configuration
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Apply connection pool configuration
	config.MaxConns = cm.config.MaxConns
	config.MinConns = cm.config.MinConns
	config.MaxConnLifetime = cm.config.MaxConnLifetime
	config.MaxConnIdleTime = cm.config.MaxConnIdleTime
	config.ConnConfig.ConnectTimeout = cm.config.ConnectTimeout
	config.HealthCheckPeriod = cm.config.HealthCheckPeriod

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	cm.logger.Info("Database connection pool created",
		"max_conns", cm.config.MaxConns,
		"min_conns", cm.config.MinConns,
		"max_conn_lifetime", cm.config.MaxConnLifetime,
		"max_conn_idle_time", cm.config.MaxConnIdleTime,
	)

	return pool, nil
}

// CreatePoolWithRetry creates a new database connection pool with retry logic
func (cm *ConnectionManager) CreatePoolWithRetry(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	for attempt := 1; attempt <= cm.config.RetryAttempts; attempt++ {
		cm.logger.Debug("Attempting to create database connection pool",
			"attempt", attempt,
			"max_attempts", cm.config.RetryAttempts,
		)

		pool, err = cm.CreatePool(ctx, databaseURL)
		if err == nil {
			cm.logger.Info("Database connection pool created successfully",
				"attempt", attempt,
			)
			return pool, nil
		}

		cm.logger.Warn("Failed to create database connection pool",
			"attempt", attempt,
			"error", err,
		)

		if attempt < cm.config.RetryAttempts {
			cm.logger.Debug("Retrying database connection",
				"delay", cm.config.RetryDelay,
			)
			time.Sleep(cm.config.RetryDelay)
		}
	}

	return nil, fmt.Errorf("failed to create database connection pool after %d attempts: %w", cm.config.RetryAttempts, err)
}

// GetPoolStats returns statistics about the connection pool
func (cm *ConnectionManager) GetPoolStats(pool *pgxpool.Pool) map[string]interface{} {
	stats := pool.Stat()
	return map[string]interface{}{
		"acquire_count":          stats.AcquireCount(),
		"acquire_duration":       stats.AcquireDuration(),
		"acquired_conns":         stats.AcquiredConns(),
		"canceled_acquire_count": stats.CanceledAcquireCount(),
		"constructing_conns":     stats.ConstructingConns(),
		"empty_acquire_count":    stats.EmptyAcquireCount(),
		"idle_conns":             stats.IdleConns(),
		"max_conns":              stats.MaxConns(),
		"total_conns":            stats.TotalConns(),
	}
}

// LogPoolStats logs connection pool statistics
func (cm *ConnectionManager) LogPoolStats(pool *pgxpool.Pool) {
	stats := cm.GetPoolStats(pool)
	cm.logger.Info("Database connection pool statistics",
		"acquire_count", stats["acquire_count"],
		"acquire_duration", stats["acquire_duration"],
		"acquired_conns", stats["acquired_conns"],
		"canceled_acquire_count", stats["canceled_acquire_count"],
		"constructing_conns", stats["constructing_conns"],
		"empty_acquire_count", stats["empty_acquire_count"],
		"idle_conns", stats["idle_conns"],
		"max_conns", stats["max_conns"],
		"total_conns", stats["total_conns"],
	)
}

// ClosePool closes the connection pool
func (cm *ConnectionManager) ClosePool(pool *pgxpool.Pool) {
	if pool != nil {
		cm.logger.Info("Closing database connection pool")
		pool.Close()
	}
}
