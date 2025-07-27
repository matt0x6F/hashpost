package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// KeyRotationMigrationDAO provides database operations for key rotation migrations
type KeyRotationMigrationDAO struct {
	db bob.Executor
}

// NewKeyRotationMigrationDAO creates a new key rotation migration DAO
func NewKeyRotationMigrationDAO(db bob.Executor) *KeyRotationMigrationDAO {
	return &KeyRotationMigrationDAO{
		db: db,
	}
}

// GetDB returns the database connection
func (dao *KeyRotationMigrationDAO) GetDB() bob.Executor {
	return dao.db
}

// MigrationState represents the state of a key rotation migration
type MigrationState struct {
	MigrationID      string     `json:"migration_id"`
	Domain           string     `json:"domain"`
	OldKeyVersion    int        `json:"old_key_version"`
	NewKeyVersion    int        `json:"new_key_version"`
	Status           string     `json:"status"`
	StartedAt        time.Time  `json:"started_at"`
	PausedAt         *time.Time `json:"paused_at,omitempty"`
	ResumedAt        *time.Time `json:"resumed_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	TotalRecords     int64      `json:"total_records"`
	ProcessedRecords int64      `json:"processed_records"`
	FailedRecords    int64      `json:"failed_records"`
	LastProcessedID  *string    `json:"last_processed_id,omitempty"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	RetryCount       int        `json:"retry_count"`
	MaxRetries       int        `json:"max_retries"`
	CreatedBy        int64      `json:"created_by"`
}

// CreateMigration creates a new key rotation migration
func (dao *KeyRotationMigrationDAO) CreateMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int, createdBy int64) (*MigrationState, error) {
	// Check if migration already exists
	existing, err := dao.GetMigrationByDomain(ctx, domain, oldKeyVersion, newKeyVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing migration: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Get total records to migrate
	totalRecords, err := dao.getTotalRecordsToMigrate(ctx, domain, oldKeyVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get total records: %w", err)
	}

	// Create migration record
	totalRecordsNull := sql.Null[int64]{}
	totalRecordsNull.Scan(totalRecords)

	processedRecordsNull := sql.Null[int64]{}
	processedRecordsNull.Scan(int64(0))

	failedRecordsNull := sql.Null[int64]{}
	failedRecordsNull.Scan(int64(0))

	retryCountNull := sql.Null[int32]{}
	retryCountNull.Scan(int32(0))

	maxRetriesNull := sql.Null[int32]{}
	maxRetriesNull.Scan(int32(3))

	migration := &models.KeyRotationMigrationSetter{
		Domain:           &domain,
		OldKeyVersion:    &[]int32{int32(oldKeyVersion)}[0],
		NewKeyVersion:    &[]int32{int32(newKeyVersion)}[0],
		Status:           &[]string{"pending"}[0],
		TotalRecords:     &totalRecordsNull,
		ProcessedRecords: &processedRecordsNull,
		FailedRecords:    &failedRecordsNull,
		RetryCount:       &retryCountNull,
		MaxRetries:       &maxRetriesNull,
		CreatedBy:        &createdBy,
	}

	createdMigration, err := models.KeyRotationMigrations.Insert(migration).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration: %w", err)
	}

	return dao.convertToMigrationState(createdMigration), nil
}

// GetMigrationByDomain retrieves a migration by domain and key versions
func (dao *KeyRotationMigrationDAO) GetMigrationByDomain(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int) (*MigrationState, error) {
	migration, err := models.KeyRotationMigrations.Query(
		models.SelectWhere.KeyRotationMigrations.Domain.EQ(domain),
		models.SelectWhere.KeyRotationMigrations.OldKeyVersion.EQ(int32(oldKeyVersion)),
		models.SelectWhere.KeyRotationMigrations.NewKeyVersion.EQ(int32(newKeyVersion)),
	).One(ctx, dao.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get migration: %w", err)
	}

	return dao.convertToMigrationState(migration), nil
}

// GetMigrationByID retrieves a migration by ID
func (dao *KeyRotationMigrationDAO) GetMigrationByID(ctx context.Context, migrationID string) (*MigrationState, error) {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return nil, fmt.Errorf("invalid migration ID: %w", err)
	}

	migration, err := models.KeyRotationMigrations.Query(
		models.SelectWhere.KeyRotationMigrations.MigrationID.EQ(migrationUUID),
	).One(ctx, dao.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get migration: %w", err)
	}

	return dao.convertToMigrationState(migration), nil
}

// UpdateMigrationStatus updates the status of a migration
func (dao *KeyRotationMigrationDAO) UpdateMigrationStatus(ctx context.Context, migrationID, status string) error {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return fmt.Errorf("invalid migration ID: %w", err)
	}

	setter := &models.KeyRotationMigrationSetter{
		Status: &status,
	}

	if status == "paused" {
		now := time.Now()
		setter.PausedAt = &sql.Null[time.Time]{}
		setter.PausedAt.Scan(now)
	} else if status == "in_progress" {
		now := time.Now()
		setter.ResumedAt = &sql.Null[time.Time]{}
		setter.ResumedAt.Scan(now)
	} else if status == "completed" {
		now := time.Now()
		setter.CompletedAt = &sql.Null[time.Time]{}
		setter.CompletedAt.Scan(now)
	}

	_, err = models.KeyRotationMigrations.Update(
		setter.UpdateMod(),
		models.UpdateWhere.KeyRotationMigrations.MigrationID.EQ(migrationUUID),
	).One(ctx, dao.db)

	if err != nil {
		return fmt.Errorf("failed to update migration status: %w", err)
	}

	return nil
}

// UpdateMigrationProgress updates the progress of a migration
func (dao *KeyRotationMigrationDAO) UpdateMigrationProgress(ctx context.Context, migrationID string, processedRecords, failedRecords int64, lastProcessedID *string) error {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return fmt.Errorf("invalid migration ID: %w", err)
	}

	processedRecordsNull := sql.Null[int64]{}
	processedRecordsNull.Scan(processedRecords)

	failedRecordsNull := sql.Null[int64]{}
	failedRecordsNull.Scan(failedRecords)

	setter := &models.KeyRotationMigrationSetter{
		ProcessedRecords: &processedRecordsNull,
		FailedRecords:    &failedRecordsNull,
	}

	if lastProcessedID != nil {
		lastIDUUID, err := uuid.FromString(*lastProcessedID)
		if err != nil {
			return fmt.Errorf("invalid last processed ID: %w", err)
		}
		lastIDNull := sql.Null[uuid.UUID]{}
		lastIDNull.Scan(lastIDUUID)
		setter.LastProcessedID = &lastIDNull
	}

	_, err = models.KeyRotationMigrations.Update(
		setter.UpdateMod(),
		models.UpdateWhere.KeyRotationMigrations.MigrationID.EQ(migrationUUID),
	).One(ctx, dao.db)

	if err != nil {
		return fmt.Errorf("failed to update migration progress: %w", err)
	}

	return nil
}

// GetUnmigratedBatch retrieves a batch of unmigrated records
func (dao *KeyRotationMigrationDAO) GetUnmigratedBatch(ctx context.Context, migrationID, domain string, offset int, batchSize int, lastProcessedID *string) ([]*models.IdentityMapping, error) {
	// Build the base query with conditions
	var query models.IdentityMappingsQuery

	if lastProcessedID != nil {
		lastIDUUID, err := uuid.FromString(*lastProcessedID)
		if err != nil {
			return nil, fmt.Errorf("invalid last processed ID: %w", err)
		}

		query = models.IdentityMappings.Query(
			models.SelectWhere.IdentityMappings.KeyScope.EQ(domain),
			models.SelectWhere.IdentityMappings.IsActive.EQ(true),
			models.SelectWhere.IdentityMappings.MappingID.GT(lastIDUUID),
			sm.Limit(batchSize),
			sm.Offset(offset),
		)
	} else {
		query = models.IdentityMappings.Query(
			models.SelectWhere.IdentityMappings.KeyScope.EQ(domain),
			models.SelectWhere.IdentityMappings.IsActive.EQ(true),
			sm.Limit(batchSize),
			sm.Offset(offset),
		)
	}

	return query.All(ctx, dao.db)
}

// MarkRecordProcessing marks a record as being processed
func (dao *KeyRotationMigrationDAO) MarkRecordProcessing(ctx context.Context, migrationID, mappingID string) error {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return fmt.Errorf("invalid migration ID: %w", err)
	}

	mappingUUID, err := uuid.FromString(mappingID)
	if err != nil {
		return fmt.Errorf("invalid mapping ID: %w", err)
	}

	now := time.Now()
	progress := &models.MigrationProgressSetter{
		MigrationID: &migrationUUID,
		MappingID:   &mappingUUID,
		Status:      &[]string{"processing"}[0],
		StartedAt:   &sql.Null[time.Time]{},
	}
	progress.StartedAt.Scan(now)

	_, err = models.MigrationProgresses.Insert(progress).One(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to mark record as processing: %w", err)
	}

	return nil
}

// MarkRecordCompleted marks a record as completed
func (dao *KeyRotationMigrationDAO) MarkRecordCompleted(ctx context.Context, migrationID, mappingID string) error {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return fmt.Errorf("invalid migration ID: %w", err)
	}

	mappingUUID, err := uuid.FromString(mappingID)
	if err != nil {
		return fmt.Errorf("invalid mapping ID: %w", err)
	}

	now := time.Now()
	setter := &models.MigrationProgressSetter{
		Status:      &[]string{"completed"}[0],
		CompletedAt: &sql.Null[time.Time]{},
	}
	setter.CompletedAt.Scan(now)

	_, err = models.MigrationProgresses.Update(
		setter.UpdateMod(),
		models.UpdateWhere.MigrationProgresses.MigrationID.EQ(migrationUUID),
		models.UpdateWhere.MigrationProgresses.MappingID.EQ(mappingUUID),
	).One(ctx, dao.db)

	if err != nil {
		return fmt.Errorf("failed to mark record as completed: %w", err)
	}

	return nil
}

// MarkRecordFailed marks a record as failed
func (dao *KeyRotationMigrationDAO) MarkRecordFailed(ctx context.Context, migrationID, mappingID, errorMessage string) error {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return fmt.Errorf("invalid migration ID: %w", err)
	}

	mappingUUID, err := uuid.FromString(mappingID)
	if err != nil {
		return fmt.Errorf("invalid mapping ID: %w", err)
	}

	now := time.Now()
	errorMessageNull := sql.Null[string]{}
	errorMessageNull.Scan(errorMessage)

	setter := &models.MigrationProgressSetter{
		Status:       &[]string{"failed"}[0],
		CompletedAt:  &sql.Null[time.Time]{},
		ErrorMessage: &errorMessageNull,
	}
	setter.CompletedAt.Scan(now)

	_, err = models.MigrationProgresses.Update(
		setter.UpdateMod(),
		models.UpdateWhere.MigrationProgresses.MigrationID.EQ(migrationUUID),
		models.UpdateWhere.MigrationProgresses.MappingID.EQ(mappingUUID),
	).One(ctx, dao.db)

	if err != nil {
		return fmt.Errorf("failed to mark record as failed: %w", err)
	}

	return nil
}

// IsRecordAlreadyMigrated checks if a record has already been migrated
func (dao *KeyRotationMigrationDAO) IsRecordAlreadyMigrated(ctx context.Context, migrationID, mappingID string) (bool, error) {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return false, fmt.Errorf("invalid migration ID: %w", err)
	}

	mappingUUID, err := uuid.FromString(mappingID)
	if err != nil {
		return false, fmt.Errorf("invalid mapping ID: %w", err)
	}

	progress, err := models.MigrationProgresses.Query(
		models.SelectWhere.MigrationProgresses.MigrationID.EQ(migrationUUID),
		models.SelectWhere.MigrationProgresses.MappingID.EQ(mappingUUID),
		models.SelectWhere.MigrationProgresses.Status.EQ("completed"),
	).One(ctx, dao.db)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check record migration status: %w", err)
	}

	return progress != nil, nil
}

// GetStuckRecords retrieves records that have been processing for too long
func (dao *KeyRotationMigrationDAO) GetStuckRecords(ctx context.Context, migrationID string, timeoutMinutes int) ([]*models.MigrationProgress, error) {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return nil, fmt.Errorf("invalid migration ID: %w", err)
	}

	timeout := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)

	return models.MigrationProgresses.Query(
		models.SelectWhere.MigrationProgresses.MigrationID.EQ(migrationUUID),
		models.SelectWhere.MigrationProgresses.Status.EQ("processing"),
		models.SelectWhere.MigrationProgresses.StartedAt.LT(timeout),
	).All(ctx, dao.db)
}

// ResetRecordStatus resets a record's status to pending
func (dao *KeyRotationMigrationDAO) ResetRecordStatus(ctx context.Context, migrationID, mappingID, status string) error {
	migrationUUID, err := uuid.FromString(migrationID)
	if err != nil {
		return fmt.Errorf("invalid migration ID: %w", err)
	}

	mappingUUID, err := uuid.FromString(mappingID)
	if err != nil {
		return fmt.Errorf("invalid mapping ID: %w", err)
	}

	setter := &models.MigrationProgressSetter{
		Status: &status,
	}

	_, err = models.MigrationProgresses.Update(
		setter.UpdateMod(),
		models.UpdateWhere.MigrationProgresses.MigrationID.EQ(migrationUUID),
		models.UpdateWhere.MigrationProgresses.MappingID.EQ(mappingUUID),
	).One(ctx, dao.db)

	if err != nil {
		return fmt.Errorf("failed to reset record status: %w", err)
	}

	return nil
}

// GetMigrationProgress retrieves the progress of a migration
func (dao *KeyRotationMigrationDAO) GetMigrationProgress(ctx context.Context, migrationID string) (*MigrationProgress, error) {
	migration, err := dao.GetMigrationByID(ctx, migrationID)
	if err != nil {
		return nil, err
	}

	if migration == nil {
		return nil, fmt.Errorf("migration not found")
	}

	percentage := float64(0)
	if migration.TotalRecords > 0 {
		percentage = float64(migration.ProcessedRecords) / float64(migration.TotalRecords) * 100
	}

	return &MigrationProgress{
		MigrationID:         migration.MigrationID,
		Domain:              migration.Domain,
		Status:              migration.Status,
		TotalRecords:        migration.TotalRecords,
		ProcessedRecords:    migration.ProcessedRecords,
		FailedRecords:       migration.FailedRecords,
		Percentage:          percentage,
		StartedAt:           migration.StartedAt,
		EstimatedCompletion: dao.estimateCompletion(migration),
	}, nil
}

// MigrationProgress represents the progress of a migration
type MigrationProgress struct {
	MigrationID         string     `json:"migration_id"`
	Domain              string     `json:"domain"`
	Status              string     `json:"status"`
	TotalRecords        int64      `json:"total_records"`
	ProcessedRecords    int64      `json:"processed_records"`
	FailedRecords       int64      `json:"failed_records"`
	Percentage          float64    `json:"percentage"`
	StartedAt           time.Time  `json:"started_at"`
	EstimatedCompletion *time.Time `json:"estimated_completion,omitempty"`
}

// Helper methods

func (dao *KeyRotationMigrationDAO) getTotalRecordsToMigrate(ctx context.Context, domain string, oldKeyVersion int) (int64, error) {
	count, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.KeyScope.EQ(domain),
		models.SelectWhere.IdentityMappings.KeyVersion.EQ(int32(oldKeyVersion)),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).Count(ctx, dao.db)

	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

func (dao *KeyRotationMigrationDAO) convertToMigrationState(migration *models.KeyRotationMigration) *MigrationState {
	state := &MigrationState{
		MigrationID:   migration.MigrationID.String(),
		Domain:        migration.Domain,
		OldKeyVersion: int(migration.OldKeyVersion),
		NewKeyVersion: int(migration.NewKeyVersion),
		Status:        migration.Status,
		CreatedBy:     migration.CreatedBy,
	}

	// Handle nullable fields
	if migration.StartedAt.Valid {
		state.StartedAt = migration.StartedAt.V
	}
	if migration.PausedAt.Valid {
		state.PausedAt = &migration.PausedAt.V
	}
	if migration.ResumedAt.Valid {
		state.ResumedAt = &migration.ResumedAt.V
	}
	if migration.CompletedAt.Valid {
		state.CompletedAt = &migration.CompletedAt.V
	}
	if migration.TotalRecords.Valid {
		state.TotalRecords = migration.TotalRecords.V
	}
	if migration.ProcessedRecords.Valid {
		state.ProcessedRecords = migration.ProcessedRecords.V
	}
	if migration.FailedRecords.Valid {
		state.FailedRecords = migration.FailedRecords.V
	}
	if migration.LastProcessedID.Valid {
		lastID := migration.LastProcessedID.V.String()
		state.LastProcessedID = &lastID
	}
	if migration.ErrorMessage.Valid {
		state.ErrorMessage = &migration.ErrorMessage.V
	}
	if migration.RetryCount.Valid {
		state.RetryCount = int(migration.RetryCount.V)
	}
	if migration.MaxRetries.Valid {
		state.MaxRetries = int(migration.MaxRetries.V)
	}

	return state
}

func (dao *KeyRotationMigrationDAO) estimateCompletion(migration *MigrationState) *time.Time {
	if migration.ProcessedRecords == 0 {
		return nil
	}

	elapsed := time.Since(migration.StartedAt)
	recordsPerSecond := float64(migration.ProcessedRecords) / elapsed.Seconds()

	if recordsPerSecond <= 0 {
		return nil
	}

	remainingRecords := migration.TotalRecords - migration.ProcessedRecords
	estimatedSeconds := float64(remainingRecords) / recordsPerSecond

	estimatedCompletion := time.Now().Add(time.Duration(estimatedSeconds) * time.Second)
	return &estimatedCompletion
}
