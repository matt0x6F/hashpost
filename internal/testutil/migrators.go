package testutil

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/peterldowns/pgtestdb"
)

// PDSMigrator returns a migrator for PDS database
func PDSMigrator() pgtestdb.Migrator {
	return &golangMigrateMigrator{
		migrationsPath: getMigrationsPath("internal/database/migrations/pds"),
	}
}

// AppViewMigrator returns a migrator for AppView database
func AppViewMigrator() pgtestdb.Migrator {
	return &golangMigrateMigrator{
		migrationsPath: getMigrationsPath("internal/database/migrations/appview"),
	}
}

// getMigrationsPath returns the absolute path to migrations directory
func getMigrationsPath(relativePath string) string {
	// Try to find the project root by looking for go.mod
	wd, err := os.Getwd()
	if err != nil {
		return relativePath
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return filepath.Join(wd, relativePath)
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	return relativePath
}

// golangMigrateMigrator implements pgtestdb.Migrator using golang-migrate
type golangMigrateMigrator struct {
	migrationsPath string
}

// Hash returns a hash of all migration files to identify template databases
func (m *golangMigrateMigrator) Hash() (string, error) {
	var files []string
	err := filepath.WalkDir(m.migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".up.sql") || strings.HasSuffix(path, ".sql")) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(files)
	content := strings.Join(files, "\n")

	// Add file contents to hash
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		content += "\n" + string(data)
	}

	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash), nil
}

// Migrate runs the migrations using golang-migrate
func (m *golangMigrateMigrator) Migrate(ctx context.Context, db *sql.DB, config pgtestdb.Config) error {
	// Get database URL from the connection
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.Database)

	// Create migrate instance
	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", m.migrationsPath),
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	// Run migrations
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
