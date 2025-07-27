package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/services"
	"github.com/matt0x6f/hashpost/internal/ibe"
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

// Command returns the key rotation command
func (cmd *KeyRotationCommand) Command() *cobra.Command {
	keyRotationCmd := &cobra.Command{
		Use:   "key-rotation",
		Short: "Manage key rotation migrations",
		Long:  "Commands for managing IBE key rotation migrations with resumable functionality",
	}

	keyRotationCmd.AddCommand(
		cmd.startCommand(),
		cmd.statusCommand(),
		cmd.pauseCommand(),
		cmd.resumeCommand(),
		cmd.recoverCommand(),
		cmd.listCommand(),
	)

	return keyRotationCmd
}

// startCommand creates the start migration command
func (cmd *KeyRotationCommand) startCommand() *cobra.Command {
	var domain string
	var oldKeyVersion int
	var newKeyVersion int
	var createdBy int64

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new key rotation migration",
		Long:  "Start a new key rotation migration for a specific domain and key versions",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := context.Background()

			log.Info().
				Str("domain", domain).
				Int("old_version", oldKeyVersion).
				Int("new_version", newKeyVersion).
				Int64("created_by", createdBy).
				Msg("Starting key rotation migration")

			err := cmd.migrationService.StartOrResumeMigration(ctx, domain, oldKeyVersion, newKeyVersion, createdBy)
			if err != nil {
				return fmt.Errorf("failed to start migration: %w", err)
			}

			log.Info().Msg("Key rotation migration started successfully")
			return nil
		},
	}

	startCmd.Flags().StringVar(&domain, "domain", "", "Domain to migrate (e.g., user_correlation, admin_correlation)")
	startCmd.Flags().IntVar(&oldKeyVersion, "old-version", 0, "Old key version to migrate from")
	startCmd.Flags().IntVar(&newKeyVersion, "new-version", 0, "New key version to migrate to")
	startCmd.Flags().Int64Var(&createdBy, "created-by", 1, "User ID who created the migration")

	startCmd.MarkFlagRequired("domain")
	startCmd.MarkFlagRequired("old-version")
	startCmd.MarkFlagRequired("new-version")

	return startCmd
}

// statusCommand creates the status command
func (cmd *KeyRotationCommand) statusCommand() *cobra.Command {
	var migrationID string

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get migration status",
		Long:  "Get the current status and progress of a key rotation migration",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := context.Background()

			if migrationID == "" {
				return fmt.Errorf("migration ID is required")
			}

			progress, err := cmd.migrationService.GetMigrationProgress(ctx, migrationID)
			if err != nil {
				return fmt.Errorf("failed to get migration progress: %w", err)
			}

			fmt.Printf("Migration Status: %s\n", progress.Status)
			fmt.Printf("Domain: %s\n", progress.Domain)
			fmt.Printf("Progress: %d/%d records (%.2f%%)\n",
				progress.ProcessedRecords, progress.TotalRecords, progress.Percentage)
			fmt.Printf("Failed Records: %d\n", progress.FailedRecords)
			fmt.Printf("Started At: %s\n", progress.StartedAt.Format(time.RFC3339))

			if progress.EstimatedCompletion != nil {
				fmt.Printf("Estimated Completion: %s\n", progress.EstimatedCompletion.Format(time.RFC3339))
			}

			return nil
		},
	}

	statusCmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to check status")
	statusCmd.MarkFlagRequired("id")

	return statusCmd
}

// pauseCommand creates the pause command
func (cmd *KeyRotationCommand) pauseCommand() *cobra.Command {
	var migrationID string

	pauseCmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause a running migration",
		Long:  "Pause a running key rotation migration",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := context.Background()

			if migrationID == "" {
				return fmt.Errorf("migration ID is required")
			}

			log.Info().Str("migration_id", migrationID).Msg("Pausing migration")

			err := cmd.migrationService.PauseMigration(ctx, migrationID)
			if err != nil {
				return fmt.Errorf("failed to pause migration: %w", err)
			}

			log.Info().Msg("Migration paused successfully")
			return nil
		},
	}

	pauseCmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to pause")
	pauseCmd.MarkFlagRequired("id")

	return pauseCmd
}

// resumeCommand creates the resume command
func (cmd *KeyRotationCommand) resumeCommand() *cobra.Command {
	var migrationID string

	resumeCmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a paused migration",
		Long:  "Resume a paused key rotation migration",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := context.Background()

			if migrationID == "" {
				return fmt.Errorf("migration ID is required")
			}

			log.Info().Str("migration_id", migrationID).Msg("Resuming migration")

			err := cmd.migrationService.ResumeMigration(ctx, migrationID)
			if err != nil {
				return fmt.Errorf("failed to resume migration: %w", err)
			}

			log.Info().Msg("Migration resumed successfully")
			return nil
		},
	}

	resumeCmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to resume")
	resumeCmd.MarkFlagRequired("id")

	return resumeCmd
}

// recoverCommand creates the recover command
func (cmd *KeyRotationCommand) recoverCommand() *cobra.Command {
	var migrationID string

	recoverCmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover a failed migration",
		Long:  "Recover a failed key rotation migration by resetting stuck records",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := context.Background()

			if migrationID == "" {
				return fmt.Errorf("migration ID is required")
			}

			log.Info().Str("migration_id", migrationID).Msg("Recovering migration")

			err := cmd.migrationService.RecoverFromFailure(ctx, migrationID)
			if err != nil {
				return fmt.Errorf("failed to recover migration: %w", err)
			}

			log.Info().Msg("Migration recovered successfully")
			return nil
		},
	}

	recoverCmd.Flags().StringVar(&migrationID, "id", "", "Migration ID to recover")
	recoverCmd.MarkFlagRequired("id")

	return recoverCmd
}

// listCommand creates the list command
func (cmd *KeyRotationCommand) listCommand() *cobra.Command {
	var domain string
	var status string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List migrations",
		Long:  "List key rotation migrations with optional filtering",
		RunE: func(c *cobra.Command, args []string) error {
			// For now, we'll just show a simple message
			// In a full implementation, you'd query the database for migrations
			fmt.Println("Migration listing functionality to be implemented")
			fmt.Printf("Domain filter: %s\n", domain)
			fmt.Printf("Status filter: %s\n", status)

			return nil
		},
	}

	listCmd.Flags().StringVar(&domain, "domain", "", "Filter by domain")
	listCmd.Flags().StringVar(&status, "status", "", "Filter by status (pending, in_progress, paused, completed, failed)")

	return listCmd
}
