package services

import (
	"context"
	"testing"

	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
)

// TestResumableMigrationService tests the resumable migration service with mocks
func TestResumableMigrationService(t *testing.T) {
	t.Run("StartNewMigration", func(t *testing.T) {
		// Setup mocks
		mockDAO := mocks.NewMockKeyRotationMigrationDAO()
		mockIBE := ibe.NewMockIBESystem()

		service := NewResumableMigrationService(mockDAO, mockIBE)
		ctx := context.Background()

		// Mock expectations for starting a new migration
		mockDAO.On("GetMigrationByDomain", ctx, "user_correlation", int32(1), int32(2)).Return(nil, nil)

		expectedMigration := &dao.MigrationState{
			MigrationID:      "test-migration-id",
			Domain:           "user_correlation",
			OldKeyVersion:    int32(1),
			NewKeyVersion:    int32(2),
			Status:           "pending",
			TotalRecords:     100,
			ProcessedRecords: 0,
			FailedRecords:    0,
			CreatedBy:        1,
		}
		mockDAO.On("CreateMigration", ctx, "user_correlation", int32(1), int32(2), int64(1)).Return(expectedMigration, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "test-migration-id", "in_progress").Return(nil)

		// Mock expectations for processing (simplified)
		mockDAO.On("GetUnmigratedBatch", ctx, "test-migration-id", "user_correlation", 0, 100, (*string)(nil)).Return([]*models.IdentityMapping{}, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "test-migration-id", "completed").Return(nil)

		// Execute test
		err := service.StartOrResumeMigration(ctx, "user_correlation", int32(1), int32(2), 1)

		// Verify
		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
		mockIBE.AssertExpectations(t)
	})

	t.Run("ResumeExistingMigration", func(t *testing.T) {
		// Setup mocks
		mockDAO := mocks.NewMockKeyRotationMigrationDAO()
		mockIBE := ibe.NewMockIBESystem()

		service := NewResumableMigrationService(mockDAO, mockIBE)
		ctx := context.Background()

		existingMigration := &dao.MigrationState{
			MigrationID:      "existing-migration-id",
			Domain:           "user_correlation",
			OldKeyVersion:    int32(1),
			NewKeyVersion:    int32(2),
			Status:           "paused",
			TotalRecords:     100,
			ProcessedRecords: 50,
			FailedRecords:    0,
			CreatedBy:        1,
		}

		// Mock expectations for resuming existing migration
		mockDAO.On("GetMigrationByDomain", ctx, "user_correlation", int32(1), int32(2)).Return(existingMigration, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "existing-migration-id", "in_progress").Return(nil)
		// Note: offset starts from ProcessedRecords (50), not 0, when resuming
		mockDAO.On("GetUnmigratedBatch", ctx, "existing-migration-id", "user_correlation", 50, 100, (*string)(nil)).Return([]*models.IdentityMapping{}, nil)
		mockDAO.On("UpdateMigrationStatus", ctx, "existing-migration-id", "completed").Return(nil)

		// Execute test
		err := service.StartOrResumeMigration(ctx, "user_correlation", int32(1), int32(2), 1)

		// Verify
		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
		mockIBE.AssertExpectations(t)
	})
}

// TestMigrationState tests the migration state struct
func TestMigrationState(t *testing.T) {
	state := &dao.MigrationState{
		MigrationID:      "test-migration",
		Domain:           "user_correlation",
		OldKeyVersion:    int32(1),
		NewKeyVersion:    int32(2),
		Status:           "pending",
		TotalRecords:     100,
		ProcessedRecords: 0,
		FailedRecords:    0,
		CreatedBy:        1,
	}

	assert.Equal(t, "test-migration", state.MigrationID)
	assert.Equal(t, "user_correlation", state.Domain)
	assert.Equal(t, int32(1), state.OldKeyVersion)
	assert.Equal(t, int32(2), state.NewKeyVersion)
	assert.Equal(t, "pending", state.Status)
}
