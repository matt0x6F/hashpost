package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// NewConnection creates a new database connection
func NewConnection(cfg *config.DatabaseConfig) (bob.DB, error) {
	dsn := cfg.GetDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return bob.DB{}, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return bob.DB{}, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Database).
		Int("max_open_conns", cfg.MaxOpenConns).
		Int("max_idle_conns", cfg.MaxIdleConns).
		Msg("Database connection established")

	// Convert to bob.DB which implements bob.Executor
	bobDB := bob.NewDB(db)
	return bobDB, nil
}

// HealthCheck performs a health check on the database
func HealthCheck(ctx context.Context, db bob.DB) error {
	return db.PingContext(ctx)
}
