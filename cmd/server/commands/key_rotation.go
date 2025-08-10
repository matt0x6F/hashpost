package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/matt0x6f/hashpost/internal/services"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// KeyRotationCommand provides CLI commands for key rotation management
type KeyRotationCommand struct {
	migrationDAO     *dao.KeyRotationMigrationDAO
	migrationService *services.ResumableMigrationService
	ibeSystem        *ibe.IBESystem
}

// NewKeyRotationCommand creates a new key rotation command
func NewKeyRotationCommand(migrationDAO *dao.KeyRotationMigrationDAO, migrationService *services.ResumableMigrationService, ibeSystem *ibe.IBESystem) *KeyRotationCommand {
	return &KeyRotationCommand{
		migrationDAO:     migrationDAO,
		migrationService: migrationService,
		ibeSystem:        ibeSystem,
	}
}

// GetCommand returns the cobra command for key rotation operations
func (k *KeyRotationCommand) GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-rotation",
		Short: "Manage key rotation migrations",
		Long:  "Commands for managing IBE key rotation migrations across domains",
	}

	cmd.AddCommand(
		k.getCreateCommand(),
		k.getStartCommand(),
		k.getPauseCommand(),
		k.getResumeCommand(),
		k.getStatusCommand(),
		k.getListCommand(),
		k.getCancelCommand(),
	)

	return cmd
}

// getCreateCommand returns the create migration command
func (k *KeyRotationCommand) getCreateCommand() *cobra.Command {
	var domain string
	var oldKeyVersion int32
	var newKeyVersion int32

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new key rotation migration",
		Long:  "Create a new key rotation migration for a specific domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.createMigration(cmd.Context(), domain, oldKeyVersion, newKeyVersion)
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Domain to migrate (required)")
	cmd.Flags().Int32Var(&oldKeyVersion, "old-version", 0, "Old key version (required)")
	cmd.Flags().Int32Var(&newKeyVersion, "new-version", 0, "New key version (required)")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("old-version")
	cmd.MarkFlagRequired("new-version")

	return cmd
}

// getStartCommand returns the start migration command
func (k *KeyRotationCommand) getStartCommand() *cobra.Command {
	var migrationID string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a key rotation migration",
		Long:  "Start processing a pending key rotation migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.startMigration(cmd.Context(), migrationID)
		},
	}

	cmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to start (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// getPauseCommand returns the pause migration command
func (k *KeyRotationCommand) getPauseCommand() *cobra.Command {
	var migrationID string

	cmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause a running key rotation migration",
		Long:  "Pause a currently running key rotation migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.pauseMigration(cmd.Context(), migrationID)
		},
	}

	cmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to pause (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// getResumeCommand returns the resume migration command
func (k *KeyRotationCommand) getResumeCommand() *cobra.Command {
	var migrationID string

	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a paused key rotation migration",
		Long:  "Resume a previously paused key rotation migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.resumeMigration(cmd.Context(), migrationID)
		},
	}

	cmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to resume (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// getStatusCommand returns the status command
func (k *KeyRotationCommand) getStatusCommand() *cobra.Command {
	var migrationID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get migration status",
		Long:  "Get the current status and progress of a migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.getMigrationStatus(cmd.Context(), migrationID)
		},
	}

	cmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to check (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// getListCommand returns the list migrations command
func (k *KeyRotationCommand) getListCommand() *cobra.Command {
	var domain string
	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List key rotation migrations",
		Long:  "List all key rotation migrations with optional filtering",
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.listMigrations(cmd.Context(), domain, status)
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Filter by domain")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")

	return cmd
}

// getCancelCommand returns the cancel migration command
func (k *KeyRotationCommand) getCancelCommand() *cobra.Command {
	var migrationID string

	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a key rotation migration",
		Long:  "Cancel a pending or paused key rotation migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return k.cancelMigration(cmd.Context(), migrationID)
		},
	}

	cmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to cancel (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// createMigration creates a new key rotation migration
func (k *KeyRotationCommand) createMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int32) error {
	log.Info().
		Str("domain", domain).
		Int32("old_version", oldKeyVersion).
		Int32("new_version", newKeyVersion).
		Msg("Creating key rotation migration")

	// Create migration using the service
	err := k.migrationService.StartOrResumeMigration(ctx, domain, oldKeyVersion, newKeyVersion, 1)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create migration")
		return fmt.Errorf("failed to create migration: %w", err)
	}

	log.Info().
		Str("domain", domain).
		Int32("old_version", oldKeyVersion).
		Int32("new_version", newKeyVersion).
		Msg("Migration created and started successfully")

	fmt.Printf("Migration created and started successfully:\n")
	fmt.Printf("  Domain: %s\n", domain)
	fmt.Printf("  Old Key Version: %d\n", oldKeyVersion)
	fmt.Printf("  New Key Version: %d\n", newKeyVersion)

	return nil
}

// startMigration starts a key rotation migration
func (k *KeyRotationCommand) startMigration(ctx context.Context, migrationID string) error {
	log.Info().Str("migration_id", migrationID).Msg("Starting key rotation migration")

	// Get migration details first
	migration, err := k.migrationDAO.GetMigrationByID(ctx, migrationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get migration")
		return fmt.Errorf("failed to get migration: %w", err)
	}

	if migration == nil {
		return fmt.Errorf("migration not found: %s", migrationID)
	}

	// Resume migration using the service
	err = k.migrationService.ResumeMigration(ctx, migrationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start migration")
		return fmt.Errorf("failed to start migration: %w", err)
	}

	log.Info().Str("migration_id", migrationID).Msg("Migration started successfully")
	fmt.Printf("Migration %s started successfully\n", migrationID)

	return nil
}

// pauseMigration pauses a running key rotation migration
func (k *KeyRotationCommand) pauseMigration(ctx context.Context, migrationID string) error {
	log.Info().Str("migration_id", migrationID).Msg("Pausing key rotation migration")

	// Pause migration using the service
	err := k.migrationService.PauseMigration(ctx, migrationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to pause migration")
		return fmt.Errorf("failed to pause migration: %w", err)
	}

	log.Info().Str("migration_id", migrationID).Msg("Migration paused successfully")
	fmt.Printf("Migration %s paused successfully\n", migrationID)

	return nil
}

// resumeMigration resumes a paused key rotation migration
func (k *KeyRotationCommand) resumeMigration(ctx context.Context, migrationID string) error {
	log.Info().Str("migration_id", migrationID).Msg("Resuming key rotation migration")

	// Resume migration using the service
	err := k.migrationService.ResumeMigration(ctx, migrationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to resume migration")
		return fmt.Errorf("failed to resume migration: %w", err)
	}

	log.Info().Str("migration_id", migrationID).Msg("Migration resumed successfully")
	fmt.Printf("Migration %s resumed successfully\n", migrationID)

	return nil
}

// getMigrationStatus gets the status of a migration
func (k *KeyRotationCommand) getMigrationStatus(ctx context.Context, migrationID string) error {
	log.Info().Str("migration_id", migrationID).Msg("Getting migration status")

	// Get migration status using the service
	progress, err := k.migrationService.GetMigrationProgress(ctx, migrationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get migration status")
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	if progress == nil {
		return fmt.Errorf("migration not found: %s", migrationID)
	}

	// Display migration status
	fmt.Printf("Migration Status: %s\n", progress.MigrationID)
	fmt.Printf("  Domain: %s\n", progress.Domain)
	fmt.Printf("  Status: %s\n", progress.Status)
	fmt.Printf("  Progress: %d/%d (%.1f%%)\n",
		progress.ProcessedRecords,
		progress.TotalRecords,
		progress.Percentage)
	fmt.Printf("  Failed Records: %d\n", progress.FailedRecords)
	fmt.Printf("  Started: %s\n", progress.StartedAt.Format(time.RFC3339))

	if progress.EstimatedCompletion != nil {
		fmt.Printf("  Estimated Completion: %s\n", progress.EstimatedCompletion.Format(time.RFC3339))
	}

	return nil
}

// listMigrations lists all migrations with optional filtering
func (k *KeyRotationCommand) listMigrations(ctx context.Context, domain, status string) error {
	log.Info().
		Str("domain", domain).
		Str("status", status).
		Msg("Listing key rotation migrations")

	// List migrations using the DAO
	migrations, err := k.migrationDAO.ListMigrations(ctx, domain, status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list migrations")
		return fmt.Errorf("failed to list migrations: %w", err)
	}

	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	// Display migrations
	fmt.Printf("Found %d migration(s):\n\n", len(migrations))
	for _, migration := range migrations {
		fmt.Printf("ID: %s\n", migration.MigrationID)
		fmt.Printf("  Domain: %s\n", migration.Domain)
		fmt.Printf("  Status: %s\n", migration.Status)
		fmt.Printf("  Progress: %d/%d\n", migration.ProcessedRecords, migration.TotalRecords)
		fmt.Printf("  Started: %s\n", migration.StartedAt.Format(time.RFC3339))
		if migration.CompletedAt != nil {
			fmt.Printf("  Completed: %s\n", migration.CompletedAt.Format(time.RFC3339))
		}
		fmt.Println()
	}

	return nil
}

// cancelMigration cancels a migration
func (k *KeyRotationCommand) cancelMigration(ctx context.Context, migrationID string) error {
	log.Info().Str("migration_id", migrationID).Msg("Canceling key rotation migration")

	// Update migration status to canceled
	err := k.migrationDAO.UpdateMigrationStatus(ctx, migrationID, "canceled")
	if err != nil {
		log.Error().Err(err).Msg("Failed to cancel migration")
		return fmt.Errorf("failed to cancel migration: %w", err)
	}

	log.Info().Str("migration_id", migrationID).Msg("Migration canceled successfully")
	fmt.Printf("Migration %s canceled successfully\n", migrationID)

	return nil
}

// NewKeyRotationCommands returns all key rotation-related commands
func NewKeyRotationCommands(cfg *config.Config) []*cobra.Command {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Initialize IBE system using configuration instead of hardcoded defaults
	ibeSystem, err := ibe.NewIBESystemFromConfig(cfg.IBE.DomainKeysDir, cfg.IBE.KeyVersion, cfg.IBE.Salt)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize IBE system")
	}

	// Create DAOs and services
	migrationDAO := dao.NewKeyRotationMigrationDAO(db)
	migrationService := services.NewResumableMigrationService(migrationDAO, ibeSystem)

	// Create key rotation command
	keyRotationCmd := NewKeyRotationCommand(migrationDAO, migrationService, ibeSystem)

	return []*cobra.Command{keyRotationCmd.GetCommand()}
}
