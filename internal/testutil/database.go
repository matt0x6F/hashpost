package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/stretchr/testify/require"
)

// TestDBConfig holds configuration for test database setup
type TestDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// DefaultTestDBConfig returns default test database configuration
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("TEST_DB_USER", "hashpost"),
		Password: getEnv("TEST_DB_PASSWORD", "password"),
		Database: getEnv("TEST_DB_NAME", "postgres"),
		SSLMode:  "disable",
	}
}

// SetupTestDB creates a test database connection using the existing Docker Compose database
func SetupTestDB(t *testing.T, config *TestDBConfig) (*sql.DB, *pgxpool.Pool, func()) {
	if config == nil {
		config = DefaultTestDBConfig()
	}

	// Create DSN for the test database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	// Create standard database/sql connection for testfixtures
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Create pgxpool connection for the application
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to create pgxpool connection: %v", err)
	}

	// Test pgxpool connection
	if err := pool.Ping(context.Background()); err != nil {
		sqlDB.Close()
		pool.Close()
		t.Fatalf("Failed to ping pgxpool connection: %v", err)
	}

	cleanup := func() {
		sqlDB.Close()
		pool.Close()
	}

	return sqlDB, pool, cleanup
}

// SetupTestDBWithFixtures creates a test database and loads fixtures
func SetupTestDBWithFixtures(t *testing.T, config *TestDBConfig, fixturePaths []string) (*sql.DB, *pgxpool.Pool, func()) {
	sqlDB, pool, cleanup := SetupTestDB(t, config)

	// Load fixtures if provided
	if len(fixturePaths) > 0 {
		if err := LoadFixtures(sqlDB, fixturePaths); err != nil {
			cleanup()
			t.Fatalf("Failed to load fixtures: %v", err)
		}
	}

	return sqlDB, pool, cleanup
}

// LoadFixtures loads test fixtures into the database
func LoadFixtures(sqlDB *sql.DB, fixturePaths []string) error {
	// Get absolute path to fixtures directory
	fixturesDir := getFixturesPath()

	// Create testfixtures instance
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory(fixturesDir),
	)
	if err != nil {
		return fmt.Errorf("failed to create fixtures: %w", err)
	}

	// Load fixtures
	if err := fixtures.Load(); err != nil {
		return fmt.Errorf("failed to load fixtures: %w", err)
	}

	return nil
}

// getFixturesPath returns the absolute path to the fixtures directory
func getFixturesPath() string {
	// Try to find the project root by looking for go.mod
	wd, err := os.Getwd()
	if err != nil {
		return "testdata/fixtures"
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return filepath.Join(wd, "testdata/fixtures")
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	return "testdata/fixtures"
}

// CreateTestLogger creates a test logger with reduced verbosity
func CreateTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))
}

// CreateDebugTestLogger creates a test logger with debug level
func CreateDebugTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CleanupTestDB is a helper to ensure test database cleanup
func CleanupTestDB(t *testing.T, cleanup func()) {
	if cleanup != nil {
		cleanup()
	}
}

// RunWithTestDB runs a test function with a test database
func RunWithTestDB(t *testing.T, testFunc func(t *testing.T, pool *pgxpool.Pool)) {
	_, pool, cleanup := SetupTestDB(t, nil)
	defer CleanupTestDB(t, cleanup)

	testFunc(t, pool)
}

// RunWithTestDBAndFixtures runs a test function with a test database and fixtures
func RunWithTestDBAndFixtures(t *testing.T, fixturePaths []string, testFunc func(t *testing.T, pool *pgxpool.Pool)) {
	_, pool, cleanup := SetupTestDBWithFixtures(t, nil, fixturePaths)
	defer CleanupTestDB(t, cleanup)

	testFunc(t, pool)
}

// SetupPDSTestDB creates an isolated PDS test database using pgtestdb
func SetupPDSTestDB(t *testing.T) *pgxpool.Pool {
	conf := pgtestdb.Config{
		DriverName: "pgx",
		User:       getEnv("TEST_DB_USER", "postgres"),
		Password:   getEnv("TEST_DB_PASSWORD", "password"),
		Host:       getEnv("TEST_DB_HOST", "localhost"),
		Port:       getEnv("TEST_DB_PORT", "5432"),
		Options:    "sslmode=disable",
	}

	sqlDB := pgtestdb.New(t, conf, PDSMigrator())

	// Get the actual database name from the connection
	var dbName string
	err := sqlDB.QueryRow("SELECT current_database()").Scan(&dbName)
	require.NoError(t, err)

	// Construct DSN with the actual database name
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.User, conf.Password, conf.Host, conf.Port, dbName)

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		pool.Close()
		sqlDB.Close()
	})
	return pool
}

// SetupAppViewTestDB creates an isolated AppView test database using pgtestdb
func SetupAppViewTestDB(t *testing.T) *pgxpool.Pool {
	conf := pgtestdb.Config{
		DriverName: "pgx",
		User:       getEnv("TEST_DB_USER", "postgres"),
		Password:   getEnv("TEST_DB_PASSWORD", "password"),
		Host:       getEnv("TEST_DB_HOST", "localhost"),
		Port:       getEnv("TEST_DB_PORT", "5432"),
		Options:    "sslmode=disable",
	}

	sqlDB := pgtestdb.New(t, conf, AppViewMigrator())

	// Get the actual database name from the connection
	var dbName string
	err := sqlDB.QueryRow("SELECT current_database()").Scan(&dbName)
	require.NoError(t, err)

	// Construct DSN with the actual database name
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.User, conf.Password, conf.Host, conf.Port, dbName)

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		pool.Close()
		sqlDB.Close()
	})
	return pool
}

// SetupPDSTestDBWithFixtures creates PDS test database and loads fixtures
func SetupPDSTestDBWithFixtures(t *testing.T, fixturePaths []string) *pgxpool.Pool {
	pool := SetupPDSTestDB(t)
	if len(fixturePaths) > 0 {
		sqlDB, err := sql.Open("pgx", pool.Config().ConnString())
		require.NoError(t, err)
		defer sqlDB.Close()
		require.NoError(t, LoadFixtures(sqlDB, fixturePaths))
	}
	return pool
}

// SetupAppViewTestDBWithFixtures creates AppView test database and loads fixtures
func SetupAppViewTestDBWithFixtures(t *testing.T, fixturePaths []string) *pgxpool.Pool {
	pool := SetupAppViewTestDB(t)
	if len(fixturePaths) > 0 {
		sqlDB, err := sql.Open("pgx", pool.Config().ConnString())
		require.NoError(t, err)
		defer sqlDB.Close()
		require.NoError(t, LoadFixtures(sqlDB, fixturePaths))
	}
	return pool
}
