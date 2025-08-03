package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Domain constants for key rotation migration
const (
	DOMAIN_USER_CORRELATION = "user_correlation"
)

// ResumableMigrationService handles key rotation migrations with fault tolerance
type ResumableMigrationService struct {
	migrationDAO *dao.KeyRotationMigrationDAO
	ibeSystem    *ibe.IBESystem
	batchSize    int
	maxRetries   int
	rateLimit    time.Duration
	logger       zerolog.Logger
}

// NewResumableMigrationService creates a new resumable migration service
func NewResumableMigrationService(migrationDAO *dao.KeyRotationMigrationDAO, ibeSystem *ibe.IBESystem) *ResumableMigrationService {
	return &ResumableMigrationService{
		migrationDAO: migrationDAO,
		ibeSystem:    ibeSystem,
		batchSize:    100, // Process 100 records at a time
		maxRetries:   3,
		rateLimit:    100 * time.Millisecond, // 100ms between batches
		logger:       log.With().Str("component", "resumable_migration").Logger(),
	}
}

// StartOrResumeMigration starts a new migration or resumes an existing one
func (s *ResumableMigrationService) StartOrResumeMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int, createdBy int64) error {
	// 1. Check for existing migration
	existingMigration, err := s.migrationDAO.GetMigrationByDomain(ctx, domain, oldKeyVersion, newKeyVersion)
	if err != nil {
		return fmt.Errorf("failed to check existing migration: %w", err)
	}

	if existingMigration != nil {
		// Resume existing migration
		s.logger.Info().
			Str("migration_id", existingMigration.MigrationID).
			Str("domain", existingMigration.Domain).
			Int64("processed", existingMigration.ProcessedRecords).
			Int64("total", existingMigration.TotalRecords).
			Msg("Resuming existing migration from checkpoint")

		return s.resumeMigration(ctx, existingMigration)
	} else {
		// Start new migration
		s.logger.Info().
			Str("domain", domain).
			Int("old_version", oldKeyVersion).
			Int("new_version", newKeyVersion).
			Msg("Starting new key rotation migration")

		return s.startNewMigration(ctx, domain, oldKeyVersion, newKeyVersion, createdBy)
	}
}

// startNewMigration creates and starts a new migration
func (s *ResumableMigrationService) startNewMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int, createdBy int64) error {
	// 1. Create migration record
	migration, err := s.migrationDAO.CreateMigration(ctx, domain, oldKeyVersion, newKeyVersion, createdBy)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	// 2. Update status to in_progress
	err = s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "in_progress")
	if err != nil {
		return fmt.Errorf("failed to update migration status: %w", err)
	}

	// 3. Start processing
	return s.processMigration(ctx, migration)
}

// resumeMigration resumes an existing migration from its checkpoint
func (s *ResumableMigrationService) resumeMigration(ctx context.Context, migration *dao.MigrationState) error {
	// 1. Update status to in_progress
	err := s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "in_progress")
	if err != nil {
		return fmt.Errorf("failed to update migration status: %w", err)
	}

	// 2. Resume processing from checkpoint
	return s.processMigration(ctx, migration)
}

// processMigration handles the actual migration processing
func (s *ResumableMigrationService) processMigration(ctx context.Context, migration *dao.MigrationState) error {
	offset := int(migration.ProcessedRecords)
	lastProcessedID := migration.LastProcessedID

	for {
		// Check if migration is complete
		if offset >= int(migration.TotalRecords) {
			s.logger.Info().
				Str("migration_id", migration.MigrationID).
				Msg("Migration completed successfully")
			return s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "completed")
		}

		// Get batch of unmigrated records
		mappings, err := s.migrationDAO.GetUnmigratedBatch(ctx, migration.MigrationID, migration.Domain, offset, s.batchSize, lastProcessedID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get unmigrated batch")
			return fmt.Errorf("failed to get unmigrated batch: %w", err)
		}

		if len(mappings) == 0 {
			// No more records to migrate
			s.logger.Info().
				Str("migration_id", migration.MigrationID).
				Msg("No more records to migrate")
			return s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "completed")
		}

		// Process batch
		processedCount := 0
		failedCount := 0
		var lastID *string

		for _, mapping := range mappings {
			err := s.migrateSingleRecordWithCheckpoint(ctx, migration.MigrationID, mapping)
			if err != nil {
				s.logger.Error().Err(err).Str("mapping_id", mapping.MappingID.String()).Msg("Failed to migrate record")
				failedCount++
				continue
			}

			processedCount++
			mappingID := mapping.MappingID.String()
			lastID = &mappingID
		}

		// Update progress
		offset += len(mappings)
		err = s.migrationDAO.UpdateMigrationProgress(ctx, migration.MigrationID, int64(offset), int64(failedCount), lastID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to update migration progress")
		}

		s.logger.Info().
			Str("migration_id", migration.MigrationID).
			Int("processed", processedCount).
			Int("failed", failedCount).
			Int("total_processed", offset).
			Int64("total_records", migration.TotalRecords).
			Msg("Batch migration progress")

		// Rate limiting
		time.Sleep(s.rateLimit)
	}
}

// migrateSingleRecordWithCheckpoint migrates a single record with checkpointing
func (s *ResumableMigrationService) migrateSingleRecordWithCheckpoint(ctx context.Context, migrationID string, mapping *models.IdentityMapping) error {
	// 1. Mark as processing
	err := s.migrationDAO.MarkRecordProcessing(ctx, migrationID, mapping.MappingID.String())
	if err != nil {
		return fmt.Errorf("failed to mark record as processing: %w", err)
	}

	// 2. Check if already migrated
	alreadyMigrated, err := s.migrationDAO.IsRecordAlreadyMigrated(ctx, migrationID, mapping.MappingID.String())
	if err != nil {
		return fmt.Errorf("failed to check if record already migrated: %w", err)
	}

	if alreadyMigrated {
		return s.migrationDAO.MarkRecordCompleted(ctx, migrationID, mapping.MappingID.String())
	}

	// 3. Attempt migration with retries
	var lastErr error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		err := s.attemptRecordMigration(ctx, mapping)
		if err == nil {
			// Success
			return s.migrationDAO.MarkRecordCompleted(ctx, migrationID, mapping.MappingID.String())
		}

		lastErr = err
		s.logger.Warn().
			Err(err).
			Str("mapping_id", mapping.MappingID.String()).
			Int("attempt", attempt+1).
			Msg("Migration attempt failed, retrying")

		// Exponential backoff
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	// 4. Mark as failed after all retries
	return s.migrationDAO.MarkRecordFailed(ctx, migrationID, mapping.MappingID.String(), lastErr.Error())
}

// attemptRecordMigration attempts to migrate a single record
func (s *ResumableMigrationService) attemptRecordMigration(ctx context.Context, mapping *models.IdentityMapping) error {
	// 1. Decrypt with old key version
	decrypted, err := s.decryptWithOldKey(mapping.EncryptedRealIdentity, int(mapping.KeyVersion))
	if err != nil {
		return fmt.Errorf("failed to decrypt with old key: %w", err)
	}

	// 2. Re-encrypt with new key version
	reEncrypted, err := s.encryptWithNewKey(decrypted, mapping.KeyScope)
	if err != nil {
		return fmt.Errorf("failed to re-encrypt mapping %s with new key: %w", mapping.MappingID.String(), err)
	}

	// 3. Update the record with new key version
	newVersion := int32(s.ibeSystem.GetKeyVersion())
	setter := &models.IdentityMappingSetter{
		EncryptedRealIdentity:     &reEncrypted,
		EncryptedPseudonymMapping: &reEncrypted,
		KeyVersion:                &newVersion,
	}

	// Get the database connection from the DAO
	err = mapping.Update(ctx, s.migrationDAO.GetDB(), setter)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}

	return nil
}

// decryptWithOldKey decrypts data using the old key version
func (s *ResumableMigrationService) decryptWithOldKey(encryptedData []byte, keyVersion int) ([]byte, error) {
	// Use the multi-version IBE system to decrypt with the old key version
	// Parse the encrypted data to extract fingerprint and pseudonym ID
	fingerprint, pseudonymID, err := s.ibeSystem.DecryptIdentityWithVersion(encryptedData, DOMAIN_USER_CORRELATION, keyVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with version %d: %w", keyVersion, err)
	}

	// Return the decrypted data as a mapping string (fingerprint:pseudonym_id)
	mappingData := fmt.Sprintf("%s:%s", fingerprint, pseudonymID)
	return []byte(mappingData), nil
}

// encryptWithNewKey encrypts data using the new key version
func (s *ResumableMigrationService) encryptWithNewKey(data []byte, keyScope string) ([]byte, error) {
	// Parse the mapping data (contains fingerprint:pseudonym_id)
	mappingStr := string(data)
	parts := strings.Split(mappingStr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid mapping data format")
	}

	fingerprint := parts[0]
	pseudonymID := parts[1]

	// Map scope to domain
	var domain string
	switch keyScope {
	case "authentication":
		domain = ibe.DOMAIN_USER_CORRELATION
	case "self_correlation":
		domain = ibe.DOMAIN_USER_CORRELATION
	case "correlation":
		domain = ibe.DOMAIN_ADMIN_CORRELATION
	default:
		domain = ibe.DOMAIN_USER_CORRELATION
	}

	// Use the multi-version IBE system to encrypt with the new key version
	// Use the fingerprint mapping method to avoid double-hashing
	encrypted, err := s.ibeSystem.EncryptFingerprintMapping(fingerprint, pseudonymID, domain, s.ibeSystem.GetKeyVersion())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with new key version: %w", err)
	}

	return encrypted, nil
}

// RecoverFromFailure recovers a migration from failure
func (s *ResumableMigrationService) RecoverFromFailure(ctx context.Context, migrationID string) error {
	// 1. Get migration state
	migration, err := s.migrationDAO.GetMigrationByID(ctx, migrationID)
	if err != nil {
		return fmt.Errorf("failed to get migration: %w", err)
	}

	if migration == nil {
		return fmt.Errorf("migration not found")
	}

	// 2. Check for stuck records
	stuckRecords, err := s.migrationDAO.GetStuckRecords(ctx, migrationID, 10) // 10 minute timeout
	if err != nil {
		return fmt.Errorf("failed to get stuck records: %w", err)
	}

	// 3. Reset stuck records to pending
	for _, record := range stuckRecords {
		err := s.migrationDAO.ResetRecordStatus(ctx, migrationID, record.MappingID.String(), "pending")
		if err != nil {
			s.logger.Error().Err(err).Str("mapping_id", record.MappingID.String()).Msg("Failed to reset record")
		}
	}

	s.logger.Info().
		Str("migration_id", migrationID).
		Int("stuck_records", len(stuckRecords)).
		Msg("Recovered migration from failure")

	// 4. Resume migration
	return s.resumeMigration(ctx, migration)
}

// GetMigrationProgress gets the progress of a migration
func (s *ResumableMigrationService) GetMigrationProgress(ctx context.Context, migrationID string) (*dao.MigrationProgress, error) {
	return s.migrationDAO.GetMigrationProgress(ctx, migrationID)
}

// PauseMigration pauses a running migration
func (s *ResumableMigrationService) PauseMigration(ctx context.Context, migrationID string) error {
	return s.migrationDAO.UpdateMigrationStatus(ctx, migrationID, "paused")
}

// ResumeMigration resumes a paused migration
func (s *ResumableMigrationService) ResumeMigration(ctx context.Context, migrationID string) error {
	migration, err := s.migrationDAO.GetMigrationByID(ctx, migrationID)
	if err != nil {
		return fmt.Errorf("failed to get migration: %w", err)
	}

	if migration == nil {
		return fmt.Errorf("migration not found")
	}

	return s.resumeMigration(ctx, migration)
}
