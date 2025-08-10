package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/mock"
)

// MockKeyRotationMigrationDAO is a mock implementation of KeyRotationMigrationDAOInterface
type MockKeyRotationMigrationDAO struct {
	mock.Mock
}

// NewMockKeyRotationMigrationDAO creates a new mock KeyRotationMigrationDAO
func NewMockKeyRotationMigrationDAO() *MockKeyRotationMigrationDAO {
	return &MockKeyRotationMigrationDAO{}
}

func (m *MockKeyRotationMigrationDAO) GetDB() bob.Executor {
	args := m.Called()
	return args.Get(0).(bob.Executor)
}

func (m *MockKeyRotationMigrationDAO) CreateMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int32, createdBy int64) (*dao.MigrationState, error) {
	args := m.Called(ctx, domain, oldKeyVersion, newKeyVersion, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dao.MigrationState), args.Error(1)
}

func (m *MockKeyRotationMigrationDAO) GetMigrationByDomain(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int32) (*dao.MigrationState, error) {
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
