package services

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyManagementService_EnsureMessagingKeys(t *testing.T) {
	// Create mock dependencies
	mockRoleKeyDAO := &MockRoleKeyDAO{}
	mockIBESystem := createTestIBESystem()
	encryptionService := NewEncryptionService()

	service := NewKeyManagementService(encryptionService, mockRoleKeyDAO, mockIBESystem)

	ctx := context.Background()
	userID := int64(123)
	pseudonymID := "test-pseudonym-123"

	// Test key generation
	err := service.EnsureMessagingKeys(ctx, userID, pseudonymID)
	require.NoError(t, err)

	// Verify that CreateRoleKey was called with correct parameters
	assert.True(t, mockRoleKeyDAO.CreateRoleKeyCalled)
	assert.Equal(t, "user", mockRoleKeyDAO.LastRoleName)
	assert.Equal(t, constants.ScopeMessaging, mockRoleKeyDAO.LastScope)
	assert.Equal(t, pseudonymID, mockRoleKeyDAO.LastPseudonymID)
	assert.Nil(t, mockRoleKeyDAO.LastSubforumID)
}

func TestKeyManagementService_EnsureMessagingKeys_Error(t *testing.T) {
	// Create mock dependencies with error
	mockRoleKeyDAO := &MockRoleKeyDAO{ShouldError: true}
	mockIBESystem := createTestIBESystem()
	encryptionService := NewEncryptionService()

	service := NewKeyManagementService(encryptionService, mockRoleKeyDAO, mockIBESystem)

	ctx := context.Background()
	userID := int64(123)
	pseudonymID := "test-pseudonym-123"

	// Test key generation with error
	err := service.EnsureMessagingKeys(ctx, userID, pseudonymID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create messaging role key")
}

// MockRoleKeyDAO is a mock implementation for testing
type MockRoleKeyDAO struct {
	CreateRoleKeyCalled bool
	LastRoleName        string
	LastScope           string
	LastPseudonymID     string
	LastSubforumID      *int32
	ShouldError         bool
}

func (m *MockRoleKeyDAO) CreateRoleKey(ctx context.Context, roleName, scope string, keyData []byte, capabilities []string, expiresAt time.Time, createdByPseudonymID string, pseudonymID string, subforumID *int32) (*models.RoleKey, error) {
	if m.ShouldError {
		return nil, assert.AnError
	}

	m.CreateRoleKeyCalled = true
	m.LastRoleName = roleName
	m.LastScope = scope
	m.LastPseudonymID = pseudonymID
	m.LastSubforumID = subforumID

	return &models.RoleKey{}, nil
}

// Implement all required methods from RoleKeyDAOInterface
func (m *MockRoleKeyDAO) GetRoleKey(ctx context.Context, pseudonymID string, scope string, subforumID *int32) (*models.RoleKey, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) GetRoleKeyByID(ctx context.Context, keyID string) (*models.RoleKey, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) ListRoleKeys(ctx context.Context) ([]*models.RoleKey, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) ListRoleKeysByPseudonym(ctx context.Context, pseudonymID string) ([]*models.RoleKey, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) GetModeratorsForSubforum(ctx context.Context, subforumID int32) ([]*models.RoleKey, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) DeactivateRoleKey(ctx context.Context, keyID string) error {
	return nil
}

func (m *MockRoleKeyDAO) ValidateKeyCapability(ctx context.Context, pseudonymID string, scope, requiredCapability string, subforumID *int32) (bool, error) {
	return false, nil
}

func (m *MockRoleKeyDAO) GetKeyData(ctx context.Context, pseudonymID string, scope string, subforumID *int32) ([]byte, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) GetPlatformKeyData(ctx context.Context, roleName, scope string) ([]byte, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) ValidatePlatformKeyCapability(ctx context.Context, roleName, scope, requiredCapability string) (bool, error) {
	return false, nil
}

func (m *MockRoleKeyDAO) CreateRoleKeyWithIBE(ctx context.Context, roleName, scope string, capabilities []string, expiresAt time.Time, createdByPseudonymID string, pseudonymID string, subforumID *int32) (*models.RoleKey, error) {
	return nil, nil
}

func (m *MockRoleKeyDAO) EnsureDefaultKeys(ctx context.Context, ibeSystem interface{}, pseudonymID string, userRoles []string) error {
	return nil
}

func (m *MockRoleKeyDAO) DeleteByPseudonymID(ctx context.Context, pseudonymID string) error {
	return nil
}

// Helper function to create a test IBE system
func createTestIBESystem() *ibe.IBESystem {
	return ibe.NewIBESystemWithOptions(ibe.IBEOptions{
		DomainMasters: map[string][]byte{
			ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
			ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
			ibe.DOMAIN_USER_MESSAGING:    []byte("0123456789abcdef0123456789abcdef"),
			ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
			ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})
}
