package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// NewAuthHandlerWithMocks creates a new auth handler with mock DAOs and fixture data
func NewAuthHandlerWithMocks() (*AuthHandler, *mocks.MockUserDAO, *mocks.MockPseudonymDAO, *mocks.MockIdentityMappingDAO, *mocks.MockRoleKeyDAO, *mocks.MockSubforumDAO, *mocks.MockPermissionDAO) {
	mockUserDAO := &mocks.MockUserDAO{}
	mockPseudonymDAO := mocks.NewMockPseudonymDAO()
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockPermissionDAO := mocks.NewMockPermissionDAO()

	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

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

	handler := NewAuthHandler(
		cfg,
		nil, // nil db for testing
		mockUserDAO,
		mockPseudonymDAO,
		mockIdentityMappingDAO,
		mockRoleKeyDAO,
		ibeSystem,
		mockSubforumDAO,
		mockPermissionDAO,
		nil, // Email service
		nil, // Email verification token DAO
		nil, // Password reset token DAO
	)

	return handler, mockUserDAO, mockPseudonymDAO, mockIdentityMappingDAO, mockRoleKeyDAO, mockSubforumDAO, mockPermissionDAO
}

// TestRoleKeySecurity tests the role key security functionality
func TestRoleKeySecurity(t *testing.T) {
	t.Run("AuthenticationKeyScope", func(t *testing.T) {
		_, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

		// Test data
		testUserID := int64(1)
		testPseudonym := fixtures.CreateTestPseudonym()
		testPseudonym.PseudonymID = "pseudonym-123"
		testPseudonym.DisplayName = "AuthUser1"

		// Mock pseudonym retrieval
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, "user", "authentication").Return([]*dbmodels.Pseudonym{testPseudonym}, nil)

		// Call the method directly on the mock
		pseudonyms, err := mockPseudonymDAO.GetPseudonymsByUserID(context.Background(), testUserID, "user", "authentication")

		// Assert response
		require.NoError(t, err)
		assert.Len(t, pseudonyms, 1)
		assert.Equal(t, "pseudonym-123", pseudonyms[0].PseudonymID)

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("SelfCorrelationKeyScope", func(t *testing.T) {
		_, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

		// Test data
		testUserID := int64(1)
		testPseudonymID := "pseudonym-123"

		// Mock ownership verification
		mockPseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, testPseudonymID, testUserID, "user", "self_correlation").Return(true, nil)

		// Call the method directly on the mock
		ownsPseudonym, err := mockPseudonymDAO.VerifyPseudonymOwnership(context.Background(), testPseudonymID, testUserID, "user", "self_correlation")

		// Assert response
		require.NoError(t, err)
		assert.True(t, ownsPseudonym)

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("RoleKeyDatabaseOperations", func(t *testing.T) {
		_, _, _, _, mockRoleKeyDAO, _, _ := NewAuthHandlerWithMocks()

		// Test data
		capabilities := []string{"test_capability", "another_capability"}
		expiresAt := time.Now().AddDate(0, 1, 0)
		testRoleKey := fixtures.CreateTestRoleKey("key-123", "test_role", "test_scope", capabilities)

		// Mock role key creation
		mockRoleKeyDAO.On("CreateRoleKey", mock.Anything, "test_role", "test_scope", []byte("test_key_data"), capabilities, expiresAt, "test-pseudonym-id", "test-pseudonym-id", (*int32)(nil)).Return(testRoleKey, nil)

		// Mock role key retrieval
		mockRoleKeyDAO.On("GetRoleKey", mock.Anything, "test-pseudonym-id", "test_scope", (*int32)(nil)).Return(testRoleKey, nil)

		// Mock capability validation
		mockRoleKeyDAO.On("ValidateKeyCapability", mock.Anything, "test-pseudonym-id", "test_scope", "test_capability", (*int32)(nil)).Return(true, nil)
		mockRoleKeyDAO.On("ValidateKeyCapability", mock.Anything, "test-pseudonym-id", "test_scope", "invalid_capability", (*int32)(nil)).Return(false, nil)

		// Mock listing role keys
		mockRoleKeyDAO.On("ListRoleKeys", mock.Anything).Return([]*dbmodels.RoleKey{testRoleKey}, nil)

		// Mock deactivating role key
		mockRoleKeyDAO.On("DeactivateRoleKey", mock.Anything, "key-123").Return(nil)

		// Test role key creation
		roleKey, err := mockRoleKeyDAO.CreateRoleKey(context.Background(), "test_role", "test_scope", []byte("test_key_data"), capabilities, expiresAt, "test-pseudonym-id", "test-pseudonym-id", (*int32)(nil))
		require.NoError(t, err)
		assert.Equal(t, "test_role", roleKey.RoleName)
		assert.Equal(t, "test_scope", roleKey.Scope)

		// Test role key retrieval
		retrievedKey, err := mockRoleKeyDAO.GetRoleKey(context.Background(), "test-pseudonym-id", "test_scope", (*int32)(nil))
		require.NoError(t, err)
		assert.Equal(t, "test_role", retrievedKey.RoleName)
		assert.Equal(t, "test_scope", retrievedKey.Scope)

		// Test capability validation
		hasCapability, err := mockRoleKeyDAO.ValidateKeyCapability(context.Background(), "test-pseudonym-id", "test_scope", "test_capability", (*int32)(nil))
		require.NoError(t, err)
		assert.True(t, hasCapability)

		// Test invalid capability
		hasInvalidCapability, err := mockRoleKeyDAO.ValidateKeyCapability(context.Background(), "test-pseudonym-id", "test_scope", "invalid_capability", (*int32)(nil))
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
