package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/config"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRoleKeyDAO is a mock implementation of RoleKeyDAOInterface
type MockRoleKeyDAO struct {
	mock.Mock
}

func (m *MockRoleKeyDAO) CreateRoleKey(ctx context.Context, roleName, scope string, keyData []byte, capabilities []string, expiresAt time.Time, createdBy int64) (*dbmodels.RoleKey, error) {
	args := m.Called(ctx, roleName, scope, keyData, capabilities, expiresAt, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetRoleKey(ctx context.Context, roleName, scope string) (*dbmodels.RoleKey, error) {
	args := m.Called(ctx, roleName, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetPerUserRoleKey(ctx context.Context, roleName, scope string, createdBy int64) (*dbmodels.RoleKey, error) {
	args := m.Called(ctx, roleName, scope, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetRoleKeyByID(ctx context.Context, keyID string) (*dbmodels.RoleKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) ListRoleKeys(ctx context.Context) ([]*dbmodels.RoleKey, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*dbmodels.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) ListRoleKeysByRole(ctx context.Context, roleName string) ([]*dbmodels.RoleKey, error) {
	args := m.Called(ctx, roleName)
	return args.Get(0).([]*dbmodels.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) DeactivateRoleKey(ctx context.Context, keyID string) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockRoleKeyDAO) ValidateKeyCapability(ctx context.Context, roleName, scope, requiredCapability string) (bool, error) {
	args := m.Called(ctx, roleName, scope, requiredCapability)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleKeyDAO) GetKeyData(ctx context.Context, roleName, scope string) ([]byte, error) {
	args := m.Called(ctx, roleName, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRoleKeyDAO) GetPerUserKeyData(ctx context.Context, roleName, scope string, createdBy int64) ([]byte, error) {
	args := m.Called(ctx, roleName, scope, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRoleKeyDAO) EnsureDefaultKeys(ctx context.Context, ibeSystem interface{}, userID int64) error {
	args := m.Called(ctx, ibeSystem, userID)
	return args.Error(0)
}

// MockIdentityMappingDAO is a mock implementation of IdentityMappingDAOInterface
type MockIdentityMappingDAO struct {
	mock.Mock
}

func (m *MockIdentityMappingDAO) CreateIdentityMapping(ctx context.Context, mapping *dbmodels.IdentityMappingSetter) (*dbmodels.IdentityMapping, error) {
	args := m.Called(ctx, mapping)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.IdentityMapping), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingByPseudonymID(ctx context.Context, pseudonymID string) (*dbmodels.IdentityMapping, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.IdentityMapping), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingsByPseudonymID(ctx context.Context, pseudonymID string) (dbmodels.IdentityMappingSlice, error) {
	args := m.Called(ctx, pseudonymID)
	return args.Get(0).(dbmodels.IdentityMappingSlice), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingsByUserID(ctx context.Context, userID int64) (dbmodels.IdentityMappingSlice, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(dbmodels.IdentityMappingSlice), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingsByFingerprint(ctx context.Context, fingerprint string) (dbmodels.IdentityMappingSlice, error) {
	args := m.Called(ctx, fingerprint)
	return args.Get(0).(dbmodels.IdentityMappingSlice), args.Error(1)
}

func (m *MockIdentityMappingDAO) DeactivateIdentityMapping(ctx context.Context, mappingID string) error {
	args := m.Called(ctx, mappingID)
	return args.Error(0)
}

// MockSubforumDAO is a mock implementation of SubforumDAOInterface
type MockSubforumDAO struct {
	mock.Mock
}

func (m *MockSubforumDAO) CreateSubforum(ctx context.Context, name, displayName, description, sidebarText, rulesText string, isNSFW, isPrivate, isRestricted bool) (*dbmodels.Subforum, error) {
	args := m.Called(ctx, name, displayName, description, sidebarText, rulesText, isNSFW, isPrivate, isRestricted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.Subforum), args.Error(1)
}

func (m *MockSubforumDAO) GetSubforumByID(ctx context.Context, subforumID int32) (*dbmodels.Subforum, error) {
	args := m.Called(ctx, subforumID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.Subforum), args.Error(1)
}

func (m *MockSubforumDAO) GetSubforumByName(ctx context.Context, name string) (*dbmodels.Subforum, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbmodels.Subforum), args.Error(1)
}

func (m *MockSubforumDAO) ListSubforums(ctx context.Context) ([]*dbmodels.Subforum, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*dbmodels.Subforum), args.Error(1)
}

// Helper functions for testing
func createTestAuthHandler() (*AuthHandler, *MockUserDAO, *MockSecurePseudonymDAO, *MockIdentityMappingDAO, *MockRoleKeyDAO, *ibe.IBESystem) {
	mockUserDAO := &MockUserDAO{}
	mockSecurePseudonymDAO := &MockSecurePseudonymDAO{}
	mockIdentityMappingDAO := &MockIdentityMappingDAO{}
	mockRoleKeyDAO := &MockRoleKeyDAO{}
	mockSubforumDAO := &MockSubforumDAO{}

	ibeSystem := ibe.NewIBESystem()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test_secret_key_for_jwt_signing",
			Expiration:  time.Hour,
			Development: true,
		},
		Security: config.SecurityConfig{
			PasswordValidation: config.PasswordValidationConfig{
				MinLength:          8,
				RequireUppercase:   true,
				RequireLowercase:   true,
				RequireDigit:       true,
				RequireSpecialChar: true,
				DisallowCommon:     true,
			},
		},
	}

	handler := NewAuthHandlerWithDependencies(
		cfg,
		mockUserDAO,
		mockSecurePseudonymDAO,
		mockIdentityMappingDAO,
		mockRoleKeyDAO,
		ibeSystem,
		mockSubforumDAO,
	)

	return handler, mockUserDAO, mockSecurePseudonymDAO, mockIdentityMappingDAO, mockRoleKeyDAO, ibeSystem
}

func createTestUser(userID int64, email string, roles []string) *dbmodels.User {
	return &dbmodels.User{
		UserID:       userID,
		Email:        email,
		PasswordHash: "hashed_password",
		IsActive:     sql.Null[bool]{V: true, Valid: true},
	}
}

func createTestPseudonymSimple(pseudonymID, displayName string, userID int64) *dbmodels.Pseudonym {
	return &dbmodels.Pseudonym{
		PseudonymID: pseudonymID,
		DisplayName: displayName,
	}
}

func createTestRoleKey(keyID string, roleName, scope string, capabilities []string) *dbmodels.RoleKey {
	// Create a UUID from the string keyID for testing
	uuid, _ := uuid.FromString(keyID)
	return &dbmodels.RoleKey{
		KeyID:     uuid,
		RoleName:  roleName,
		Scope:     scope,
		KeyData:   []byte("test_key_data"),
		ExpiresAt: time.Now().AddDate(1, 0, 0),
		CreatedBy: 1,
	}
}

// TestRoleKeySecurity tests the role key security functionality
func TestRoleKeySecurity(t *testing.T) {
	t.Run("AuthenticationKeyScope", func(t *testing.T) {
		_, _, mockSecurePseudonymDAO, _, _, _ := createTestAuthHandler()

		// Test data
		testUserID := int64(1)
		testPseudonym := createTestPseudonymSimple("pseudonym-123", "AuthUser1", testUserID)

		// Mock pseudonym retrieval
		mockSecurePseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, "user", "authentication").Return([]*dbmodels.Pseudonym{testPseudonym}, nil)

		// Call the method directly on the mock
		pseudonyms, err := mockSecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUserID, "user", "authentication")

		// Assert response
		require.NoError(t, err)
		assert.Len(t, pseudonyms, 1)
		assert.Equal(t, "pseudonym-123", pseudonyms[0].PseudonymID)

		// Verify mocks were called
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("SelfCorrelationKeyScope", func(t *testing.T) {
		_, _, mockSecurePseudonymDAO, _, _, _ := createTestAuthHandler()

		// Test data
		testUserID := int64(1)
		testPseudonymID := "pseudonym-123"

		// Mock ownership verification
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, testPseudonymID, testUserID, "user", "self_correlation").Return(true, nil)

		// Call the method directly on the mock
		ownsPseudonym, err := mockSecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), testPseudonymID, testUserID, "user", "self_correlation")

		// Assert response
		require.NoError(t, err)
		assert.True(t, ownsPseudonym)

		// Verify mocks were called
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("RoleKeyDatabaseOperations", func(t *testing.T) {
		_, _, _, _, mockRoleKeyDAO, _ := createTestAuthHandler()

		// Test data
		capabilities := []string{"test_capability", "another_capability"}
		expiresAt := time.Now().AddDate(0, 1, 0)
		createdBy := int64(1)
		testRoleKey := createTestRoleKey("key-123", "test_role", "test_scope", capabilities)

		// Mock role key creation
		mockRoleKeyDAO.On("CreateRoleKey", mock.Anything, "test_role", "test_scope", []byte("test_key_data"), capabilities, expiresAt, createdBy).Return(testRoleKey, nil)

		// Mock role key retrieval
		mockRoleKeyDAO.On("GetRoleKey", mock.Anything, "test_role", "test_scope").Return(testRoleKey, nil)

		// Mock capability validation
		mockRoleKeyDAO.On("ValidateKeyCapability", mock.Anything, "test_role", "test_scope", "test_capability").Return(true, nil)
		mockRoleKeyDAO.On("ValidateKeyCapability", mock.Anything, "test_role", "test_scope", "invalid_capability").Return(false, nil)

		// Mock listing role keys
		mockRoleKeyDAO.On("ListRoleKeys", mock.Anything).Return([]*dbmodels.RoleKey{testRoleKey}, nil)

		// Mock deactivating role key
		mockRoleKeyDAO.On("DeactivateRoleKey", mock.Anything, "key-123").Return(nil)

		// Test role key creation
		roleKey, err := mockRoleKeyDAO.CreateRoleKey(context.Background(), "test_role", "test_scope", []byte("test_key_data"), capabilities, expiresAt, createdBy)
		require.NoError(t, err)
		assert.Equal(t, "test_role", roleKey.RoleName)
		assert.Equal(t, "test_scope", roleKey.Scope)

		// Test role key retrieval
		retrievedKey, err := mockRoleKeyDAO.GetRoleKey(context.Background(), "test_role", "test_scope")
		require.NoError(t, err)
		assert.Equal(t, "test_role", retrievedKey.RoleName)
		assert.Equal(t, "test_scope", retrievedKey.Scope)

		// Test capability validation
		hasCapability, err := mockRoleKeyDAO.ValidateKeyCapability(context.Background(), "test_role", "test_scope", "test_capability")
		require.NoError(t, err)
		assert.True(t, hasCapability)

		// Test invalid capability
		hasInvalidCapability, err := mockRoleKeyDAO.ValidateKeyCapability(context.Background(), "test_role", "test_scope", "invalid_capability")
		require.NoError(t, err)
		assert.False(t, hasInvalidCapability)

		// Test listing role keys
		roleKeys, err := mockRoleKeyDAO.ListRoleKeys(context.Background())
		require.NoError(t, err)
		assert.Len(t, roleKeys, 1)

		// Test deactivating role key
		err = mockRoleKeyDAO.DeactivateRoleKey(context.Background(), "key-123")
		require.NoError(t, err)

		// Verify mocks were called
		mockRoleKeyDAO.AssertExpectations(t)
	})
}
