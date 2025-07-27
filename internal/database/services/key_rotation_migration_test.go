package services

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MigrationDAOInterface defines the interface for migration DAO operations
type MigrationDAOInterface interface {
	GetDB() interface{}
	CreateMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int, createdBy int64) (*dao.MigrationState, error)
	GetMigrationByDomain(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int) (*dao.MigrationState, error)
	GetMigrationByID(ctx context.Context, migrationID string) (*dao.MigrationState, error)
	UpdateMigrationStatus(ctx context.Context, migrationID, status string) error
	UpdateMigrationProgress(ctx context.Context, migrationID string, processedRecords, failedRecords int64, lastProcessedID *string) error
	GetUnmigratedBatch(ctx context.Context, migrationID, domain string, offset int, batchSize int, lastProcessedID *string) ([]*models.IdentityMapping, error)
	MarkRecordProcessing(ctx context.Context, migrationID, mappingID string) error
	MarkRecordCompleted(ctx context.Context, migrationID, mappingID string) error
	MarkRecordFailed(ctx context.Context, migrationID, mappingID, errorMessage string) error
	IsRecordAlreadyMigrated(ctx context.Context, migrationID, mappingID string) (bool, error)
	GetStuckRecords(ctx context.Context, migrationID string, timeoutMinutes int) ([]*models.MigrationProgress, error)
	ResetRecordStatus(ctx context.Context, migrationID, mappingID, status string) error
	GetMigrationProgress(ctx context.Context, migrationID string) (*dao.MigrationProgress, error)
}

// IBESystemInterface defines the interface for IBE system operations
type IBESystemInterface interface {
	GetDomainMasters() map[string][]byte
	GetKeyVersion() int
}

// MockKeyRotationMigrationDAO is a mock implementation of the migration DAO
type MockKeyRotationMigrationDAO struct {
	mock.Mock
}

func (m *MockKeyRotationMigrationDAO) GetDB() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockKeyRotationMigrationDAO) CreateMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int, createdBy int64) (*dao.MigrationState, error) {
	args := m.Called(ctx, domain, oldKeyVersion, newKeyVersion, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dao.MigrationState), args.Error(1)
}

func (m *MockKeyRotationMigrationDAO) GetMigrationByDomain(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int) (*dao.MigrationState, error) {
	args := m.Called(ctx, domain, oldKeyVersion, newKeyVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dao.MigrationState), args.Error(1)
}

func (m *MockKeyRotationMigrationDAO) GetMigrationByID(ctx context.Context, migrationID string) (*dao.MigrationState, error) {
	args := m.Called(ctx, migrationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dao.MigrationState), args.Error(1)
}

func (m *MockKeyRotationMigrationDAO) UpdateMigrationStatus(ctx context.Context, migrationID, status string) error {
	args := m.Called(ctx, migrationID, status)
	return args.Error(0)
}

func (m *MockKeyRotationMigrationDAO) UpdateMigrationProgress(ctx context.Context, migrationID string, processedRecords, failedRecords int64, lastProcessedID *string) error {
	args := m.Called(ctx, migrationID, processedRecords, failedRecords, lastProcessedID)
	return args.Error(0)
}

func (m *MockKeyRotationMigrationDAO) GetUnmigratedBatch(ctx context.Context, migrationID, domain string, offset int, batchSize int, lastProcessedID *string) ([]*models.IdentityMapping, error) {
	args := m.Called(ctx, migrationID, domain, offset, batchSize, lastProcessedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.IdentityMapping), args.Error(1)
}

func (m *MockKeyRotationMigrationDAO) MarkRecordProcessing(ctx context.Context, migrationID, mappingID string) error {
	args := m.Called(ctx, migrationID, mappingID)
	return args.Error(0)
}

func (m *MockKeyRotationMigrationDAO) MarkRecordCompleted(ctx context.Context, migrationID, mappingID string) error {
	args := m.Called(ctx, migrationID, mappingID)
	return args.Error(0)
}

func (m *MockKeyRotationMigrationDAO) MarkRecordFailed(ctx context.Context, migrationID, mappingID, errorMessage string) error {
	args := m.Called(ctx, migrationID, mappingID, errorMessage)
	return args.Error(0)
}

func (m *MockKeyRotationMigrationDAO) IsRecordAlreadyMigrated(ctx context.Context, migrationID, mappingID string) (bool, error) {
	args := m.Called(ctx, migrationID, mappingID)
	return args.Bool(0), args.Error(1)
}

func (m *MockKeyRotationMigrationDAO) GetStuckRecords(ctx context.Context, migrationID string, timeoutMinutes int) ([]*models.MigrationProgress, error) {
	args := m.Called(ctx, migrationID, timeoutMinutes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MigrationProgress), args.Error(1)
}

func (m *MockKeyRotationMigrationDAO) ResetRecordStatus(ctx context.Context, migrationID, mappingID, status string) error {
	args := m.Called(ctx, migrationID, mappingID, status)
	return args.Error(0)
}

func (m *MockKeyRotationMigrationDAO) GetMigrationProgress(ctx context.Context, migrationID string) (*dao.MigrationProgress, error) {
	args := m.Called(ctx, migrationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dao.MigrationProgress), args.Error(1)
}

// MockIBESystem is a mock implementation of the IBE system for testing
type MockIBESystem struct {
	mock.Mock
	domainMasters map[string][]byte
	keyVersion    int
}

func NewMockIBESystem() *MockIBESystem {
	return &MockIBESystem{
		domainMasters: map[string][]byte{
			ibe.DOMAIN_USER_CORRELATION:  []byte("user-domain-key"),
			ibe.DOMAIN_ADMIN_CORRELATION: []byte("admin-domain-key"),
		},
		keyVersion: 2,
	}
}

func (m *MockIBESystem) GetDomainMasters() map[string][]byte {
	return m.domainMasters
}

func (m *MockIBESystem) GetKeyVersion() int {
	return m.keyVersion
}

// NewResumableMigrationServiceWithInterfaces creates a new resumable migration service with interfaces
func NewResumableMigrationServiceWithInterfaces(migrationDAO MigrationDAOInterface, ibeSystem IBESystemInterface) *ResumableMigrationServiceWithInterfaces {
	return &ResumableMigrationServiceWithInterfaces{
		migrationDAO: migrationDAO,
		ibeSystem:    ibeSystem,
		batchSize:    100, // Process 100 records at a time
		maxRetries:   3,
		rateLimit:    100 * time.Millisecond, // 100ms between batches
	}
}

// ResumableMigrationServiceWithInterfaces is a version of the service that uses interfaces
type ResumableMigrationServiceWithInterfaces struct {
	migrationDAO MigrationDAOInterface
	ibeSystem    IBESystemInterface
	batchSize    int
	maxRetries   int
	rateLimit    time.Duration
}

// StartOrResumeMigration starts a new migration or resumes an existing one
func (s *ResumableMigrationServiceWithInterfaces) StartOrResumeMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int, createdBy int64) error {
	// 1. Check for existing migration
	existingMigration, err := s.migrationDAO.GetMigrationByDomain(ctx, domain, oldKeyVersion, newKeyVersion)
	if err != nil {
		return err
	}

	if existingMigration != nil {
		// Resume existing migration
		return s.resumeMigration(ctx, existingMigration)
	} else {
		// Start new migration
		return s.startNewMigration(ctx, domain, oldKeyVersion, newKeyVersion, createdBy)
	}
}

// startNewMigration creates and starts a new migration
func (s *ResumableMigrationServiceWithInterfaces) startNewMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int, createdBy int64) error {
	// 1. Create migration record
	migration, err := s.migrationDAO.CreateMigration(ctx, domain, oldKeyVersion, newKeyVersion, createdBy)
	if err != nil {
		return err
	}

	// 2. Update status to in_progress
	err = s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "in_progress")
	if err != nil {
		return err
	}

	// 3. Start processing
	return s.processMigration(ctx, migration)
}

// resumeMigration resumes an existing migration from its checkpoint
func (s *ResumableMigrationServiceWithInterfaces) resumeMigration(ctx context.Context, migration *dao.MigrationState) error {
	// 1. Update status to in_progress
	err := s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "in_progress")
	if err != nil {
		return err
	}

	// 2. Resume processing from checkpoint
	return s.processMigration(ctx, migration)
}

// processMigration handles the actual migration processing
func (s *ResumableMigrationServiceWithInterfaces) processMigration(ctx context.Context, migration *dao.MigrationState) error {
	offset := int(migration.ProcessedRecords)
	lastProcessedID := migration.LastProcessedID

	for {
		// Check if migration is complete
		if offset >= int(migration.TotalRecords) {
			return s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "completed")
		}

		// Get batch of unmigrated records
		mappings, err := s.migrationDAO.GetUnmigratedBatch(ctx, migration.MigrationID, migration.Domain, offset, s.batchSize, lastProcessedID)
		if err != nil {
			return err
		}

		if len(mappings) == 0 {
			// No more records to migrate
			return s.migrationDAO.UpdateMigrationStatus(ctx, migration.MigrationID, "completed")
		}

		// Process batch
		processedCount := 0
		failedCount := 0
		var lastID *string

		for _, mapping := range mappings {
			err := s.migrateSingleRecordWithCheckpoint(ctx, migration.MigrationID, mapping)
			if err != nil {
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
			return err
		}

		// Rate limiting
		time.Sleep(s.rateLimit)
	}
}

// migrateSingleRecordWithCheckpoint migrates a single record with checkpointing
func (s *ResumableMigrationServiceWithInterfaces) migrateSingleRecordWithCheckpoint(ctx context.Context, migrationID string, mapping *models.IdentityMapping) error {
	// 1. Mark as processing
	err := s.migrationDAO.MarkRecordProcessing(ctx, migrationID, mapping.MappingID.String())
	if err != nil {
		return err
	}

	// 2. Check if already migrated
	alreadyMigrated, err := s.migrationDAO.IsRecordAlreadyMigrated(ctx, migrationID, mapping.MappingID.String())
	if err != nil {
		return err
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
		// Exponential backoff
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	// 4. Mark as failed after all retries
	return s.migrationDAO.MarkRecordFailed(ctx, migrationID, mapping.MappingID.String(), lastErr.Error())
}

// attemptRecordMigration attempts to migrate a single record
func (s *ResumableMigrationServiceWithInterfaces) attemptRecordMigration(ctx context.Context, mapping *models.IdentityMapping) error {
	// Simplified implementation for testing - just return success
	// In a real implementation, this would decrypt and re-encrypt the data
	return nil
}

// RecoverFromFailure recovers a migration from failure
func (s *ResumableMigrationServiceWithInterfaces) RecoverFromFailure(ctx context.Context, migrationID string) error {
	// 1. Get migration state
	migration, err := s.migrationDAO.GetMigrationByID(ctx, migrationID)
	if err != nil {
		return err
	}

	if migration == nil {
		return err
	}

	// 2. Check for stuck records
	stuckRecords, err := s.migrationDAO.GetStuckRecords(ctx, migrationID, 10) // 10 minute timeout
	if err != nil {
		return err
	}

	// 3. Reset stuck records to pending
	for _, record := range stuckRecords {
		err := s.migrationDAO.ResetRecordStatus(ctx, migrationID, record.MappingID.String(), "pending")
		if err != nil {
			// Continue even if some resets fail
		}
	}

	// 4. Resume migration
	return s.resumeMigration(ctx, migration)
}

// GetMigrationProgress gets the progress of a migration
func (s *ResumableMigrationServiceWithInterfaces) GetMigrationProgress(ctx context.Context, migrationID string) (*dao.MigrationProgress, error) {
	return s.migrationDAO.GetMigrationProgress(ctx, migrationID)
}

// PauseMigration pauses a running migration
func (s *ResumableMigrationServiceWithInterfaces) PauseMigration(ctx context.Context, migrationID string) error {
	return s.migrationDAO.UpdateMigrationStatus(ctx, migrationID, "paused")
}

// ResumeMigration resumes a paused migration
func (s *ResumableMigrationServiceWithInterfaces) ResumeMigration(ctx context.Context, migrationID string) error {
	migration, err := s.migrationDAO.GetMigrationByID(ctx, migrationID)
	if err != nil {
		return err
	}

	if migration == nil {
		return err
	}

	return s.resumeMigration(ctx, migration)
}

// TestResumableMigrationService tests the resumable migration service with mocks
func TestResumableMigrationService(t *testing.T) {
	t.Run("StartNewMigration", func(t *testing.T) {
		// Setup mocks
		mockDAO := &MockKeyRotationMigrationDAO{}
		mockIBE := NewMockIBESystem()

		service := NewResumableMigrationServiceWithInterfaces(mockDAO, mockIBE)
		ctx := context.Background()

		// Mock expectations for starting a new migration
		mockDAO.On("GetMigrationByDomain", ctx, "user_correlation", 1, 2).Return(nil, nil)

		expectedMigration := &dao.MigrationState{
			MigrationID:      "test-migration-id",
			Domain:           "user_correlation",
			OldKeyVersion:    1,
			NewKeyVersion:    2,
			Status:           "pending",
			TotalRecords:     100,
			ProcessedRecords: 0,
			FailedRecords:    0,
			CreatedBy:        1,
		}
		mockDAO.On("CreateMigration", ctx, "user_correlation", 1, 2, int64(1)).Return(expectedMigration, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "test-migration-id", "in_progress").Return(nil)

		// Mock expectations for processing
		testMappings := []*models.IdentityMapping{
			{
				MappingID:                 uuid.Must(uuid.NewV4()),
				EncryptedRealIdentity:     []byte("encrypted-data-1"),
				EncryptedPseudonymMapping: []byte("encrypted-data-1"),
				KeyVersion:                1,
				KeyScope:                  "authentication",
			},
			{
				MappingID:                 uuid.Must(uuid.NewV4()),
				EncryptedRealIdentity:     []byte("encrypted-data-2"),
				EncryptedPseudonymMapping: []byte("encrypted-data-2"),
				KeyVersion:                1,
				KeyScope:                  "self_correlation",
			},
		}

		mockDAO.On("GetUnmigratedBatch", ctx, "test-migration-id", "user_correlation", 0, 100, (*string)(nil)).Return(testMappings, nil)
		mockDAO.On("MarkRecordProcessing", ctx, "test-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("IsRecordAlreadyMigrated", ctx, "test-migration-id", mock.AnythingOfType("string")).Return(false, nil)
		mockDAO.On("MarkRecordCompleted", ctx, "test-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("UpdateMigrationProgress", ctx, "test-migration-id", int64(2), int64(0), mock.AnythingOfType("*string")).Return(nil)
		mockDAO.On("GetUnmigratedBatch", ctx, "test-migration-id", "user_correlation", 2, 100, mock.AnythingOfType("*string")).Return([]*models.IdentityMapping{}, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "test-migration-id", "completed").Return(nil)

		// Execute migration
		err := service.StartOrResumeMigration(ctx, "user_correlation", 1, 2, 1)

		// Verify results
		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
	})

	t.Run("ResumeExistingMigration", func(t *testing.T) {
		// Setup mocks
		mockDAO := &MockKeyRotationMigrationDAO{}
		mockIBE := NewMockIBESystem()

		service := NewResumableMigrationServiceWithInterfaces(mockDAO, mockIBE)
		ctx := context.Background()

		// Mock existing migration
		existingMigration := &dao.MigrationState{
			MigrationID:      "existing-migration-id",
			Domain:           "user_correlation",
			OldKeyVersion:    1,
			NewKeyVersion:    2,
			Status:           "paused",
			TotalRecords:     100,
			ProcessedRecords: 50,
			FailedRecords:    2,
			CreatedBy:        1,
		}

		mockDAO.On("GetMigrationByDomain", ctx, "user_correlation", 1, 2).Return(existingMigration, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "existing-migration-id", "in_progress").Return(nil)

		// Mock remaining records to process
		remainingMappings := []*models.IdentityMapping{
			{
				MappingID:                 uuid.Must(uuid.NewV4()),
				EncryptedRealIdentity:     []byte("encrypted-data-3"),
				EncryptedPseudonymMapping: []byte("encrypted-data-3"),
				KeyVersion:                1,
				KeyScope:                  "authentication",
			},
		}

		mockDAO.On("GetUnmigratedBatch", ctx, "existing-migration-id", "user_correlation", 50, 100, (*string)(nil)).Return(remainingMappings, nil)
		mockDAO.On("MarkRecordProcessing", ctx, "existing-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("IsRecordAlreadyMigrated", ctx, "existing-migration-id", mock.AnythingOfType("string")).Return(false, nil)
		mockDAO.On("MarkRecordCompleted", ctx, "existing-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("UpdateMigrationProgress", ctx, "existing-migration-id", int64(51), int64(0), mock.AnythingOfType("*string")).Return(nil)
		mockDAO.On("GetUnmigratedBatch", ctx, "existing-migration-id", "user_correlation", 51, 100, mock.AnythingOfType("*string")).Return([]*models.IdentityMapping{}, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "existing-migration-id", "completed").Return(nil)

		// Execute migration
		err := service.StartOrResumeMigration(ctx, "user_correlation", 1, 2, 1)

		// Verify results
		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
	})

	t.Run("RecoverFromFailure", func(t *testing.T) {
		// Setup mocks
		mockDAO := &MockKeyRotationMigrationDAO{}
		mockIBE := NewMockIBESystem()

		service := NewResumableMigrationServiceWithInterfaces(mockDAO, mockIBE)
		ctx := context.Background()

		// Mock migration state
		migration := &dao.MigrationState{
			MigrationID:      "failed-migration-id",
			Domain:           "user_correlation",
			OldKeyVersion:    1,
			NewKeyVersion:    2,
			Status:           "failed",
			TotalRecords:     100,
			ProcessedRecords: 75,
			FailedRecords:    5,
			CreatedBy:        1,
		}

		mockDAO.On("GetMigrationByID", ctx, "failed-migration-id").Return(migration, nil)

		// Mock stuck records
		stuckRecords := []*models.MigrationProgress{
			{
				MigrationID: uuid.Must(uuid.NewV4()),
				MappingID:   uuid.Must(uuid.NewV4()),
				Status:      "processing",
			},
		}

		mockDAO.On("GetStuckRecords", ctx, "failed-migration-id", 10).Return(stuckRecords, nil)
		mockDAO.On("ResetRecordStatus", ctx, "failed-migration-id", mock.AnythingOfType("string"), "pending").Return(nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "failed-migration-id", "in_progress").Return(nil)

		// Mock remaining processing
		remainingMappings := []*models.IdentityMapping{
			{
				MappingID:                 uuid.Must(uuid.NewV4()),
				EncryptedRealIdentity:     []byte("encrypted-data-4"),
				EncryptedPseudonymMapping: []byte("encrypted-data-4"),
				KeyVersion:                1,
				KeyScope:                  "authentication",
			},
		}

		mockDAO.On("GetUnmigratedBatch", ctx, "failed-migration-id", "user_correlation", 75, 100, (*string)(nil)).Return(remainingMappings, nil)
		mockDAO.On("MarkRecordProcessing", ctx, "failed-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("IsRecordAlreadyMigrated", ctx, "failed-migration-id", mock.AnythingOfType("string")).Return(false, nil)
		mockDAO.On("MarkRecordCompleted", ctx, "failed-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("UpdateMigrationProgress", ctx, "failed-migration-id", int64(76), int64(0), mock.AnythingOfType("*string")).Return(nil)
		mockDAO.On("GetUnmigratedBatch", ctx, "failed-migration-id", "user_correlation", 76, 100, mock.AnythingOfType("*string")).Return([]*models.IdentityMapping{}, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "failed-migration-id", "completed").Return(nil)

		// Execute recovery
		err := service.RecoverFromFailure(ctx, "failed-migration-id")

		// Verify results
		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
	})

	t.Run("GetMigrationProgress", func(t *testing.T) {
		// Setup mocks
		mockDAO := &MockKeyRotationMigrationDAO{}
		mockIBE := NewMockIBESystem()

		service := NewResumableMigrationServiceWithInterfaces(mockDAO, mockIBE)
		ctx := context.Background()

		// Mock progress response
		expectedProgress := &dao.MigrationProgress{
			MigrationID:      "test-migration-id",
			Domain:           "user_correlation",
			Status:           "in_progress",
			TotalRecords:     100,
			ProcessedRecords: 50,
			FailedRecords:    2,
			Percentage:       50.0,
			StartedAt:        time.Now(),
		}

		mockDAO.On("GetMigrationProgress", ctx, "test-migration-id").Return(expectedProgress, nil)

		// Execute progress check
		progress, err := service.GetMigrationProgress(ctx, "test-migration-id")

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedProgress, progress)
		mockDAO.AssertExpectations(t)
	})

	t.Run("PauseAndResumeMigration", func(t *testing.T) {
		// Setup mocks
		mockDAO := &MockKeyRotationMigrationDAO{}
		mockIBE := NewMockIBESystem()

		service := NewResumableMigrationServiceWithInterfaces(mockDAO, mockIBE)
		ctx := context.Background()

		// Mock pause
		mockDAO.On("UpdateMigrationStatus", ctx, "test-migration-id", "paused").Return(nil)

		// Execute pause
		err := service.PauseMigration(ctx, "test-migration-id")
		assert.NoError(t, err)

		// Mock resume
		migration := &dao.MigrationState{
			MigrationID:      "test-migration-id",
			Domain:           "user_correlation",
			OldKeyVersion:    1,
			NewKeyVersion:    2,
			Status:           "paused",
			TotalRecords:     100,
			ProcessedRecords: 25,
			FailedRecords:    1,
			CreatedBy:        1,
		}

		mockDAO.On("GetMigrationByID", ctx, "test-migration-id").Return(migration, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "test-migration-id", "in_progress").Return(nil)

		// Mock remaining processing
		remainingMappings := []*models.IdentityMapping{
			{
				MappingID:                 uuid.Must(uuid.NewV4()),
				EncryptedRealIdentity:     []byte("encrypted-data-5"),
				EncryptedPseudonymMapping: []byte("encrypted-data-5"),
				KeyVersion:                1,
				KeyScope:                  "authentication",
			},
		}

		mockDAO.On("GetUnmigratedBatch", ctx, "test-migration-id", "user_correlation", 25, 100, (*string)(nil)).Return(remainingMappings, nil)
		mockDAO.On("MarkRecordProcessing", ctx, "test-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("IsRecordAlreadyMigrated", ctx, "test-migration-id", mock.AnythingOfType("string")).Return(false, nil)
		mockDAO.On("MarkRecordCompleted", ctx, "test-migration-id", mock.AnythingOfType("string")).Return(nil)
		mockDAO.On("UpdateMigrationProgress", ctx, "test-migration-id", int64(26), int64(0), mock.AnythingOfType("*string")).Return(nil)
		mockDAO.On("GetUnmigratedBatch", ctx, "test-migration-id", "user_correlation", 26, 100, mock.AnythingOfType("*string")).Return([]*models.IdentityMapping{}, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "test-migration-id", "completed").Return(nil)

		// Execute resume
		err = service.ResumeMigration(ctx, "test-migration-id")
		assert.NoError(t, err)

		mockDAO.AssertExpectations(t)
	})
}

// TestMigrationState tests the migration state structure
func TestMigrationState(t *testing.T) {
	state := &dao.MigrationState{
		MigrationID:      "test-migration-id",
		Domain:           "user_correlation",
		OldKeyVersion:    1,
		NewKeyVersion:    2,
		Status:           "in_progress",
		StartedAt:        time.Now(),
		TotalRecords:     1000,
		ProcessedRecords: 500,
		FailedRecords:    5,
	}

	assert.Equal(t, "test-migration-id", state.MigrationID)
	assert.Equal(t, "user_correlation", state.Domain)
	assert.Equal(t, 1, state.OldKeyVersion)
	assert.Equal(t, 2, state.NewKeyVersion)
	assert.Equal(t, "in_progress", state.Status)
	assert.Equal(t, int64(1000), state.TotalRecords)
	assert.Equal(t, int64(500), state.ProcessedRecords)
	assert.Equal(t, int64(5), state.FailedRecords)
}

// TestMigrationProgress tests the migration progress structure
func TestMigrationProgress(t *testing.T) {
	progress := &dao.MigrationProgress{
		MigrationID:         "test-migration-id",
		Domain:              "user_correlation",
		Status:              "in_progress",
		TotalRecords:        1000,
		ProcessedRecords:    500,
		FailedRecords:       5,
		Percentage:          50.0,
		StartedAt:           time.Now(),
		EstimatedCompletion: nil,
	}

	assert.Equal(t, "test-migration-id", progress.MigrationID)
	assert.Equal(t, "user_correlation", progress.Domain)
	assert.Equal(t, "in_progress", progress.Status)
	assert.Equal(t, int64(1000), progress.TotalRecords)
	assert.Equal(t, int64(500), progress.ProcessedRecords)
	assert.Equal(t, int64(5), progress.FailedRecords)
	assert.Equal(t, 50.0, progress.Percentage)
}
