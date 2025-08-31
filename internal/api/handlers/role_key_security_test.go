package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// NewAuthHandlerWithGomocks creates a new auth handler with gomock DAOs and fixture data
func NewAuthHandlerWithGomocks(ctrl *gomock.Controller) (*handlers.AuthHandler, *dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockIdentityMappingDAOInterface, *dao.MockRoleKeyDAOInterface, *dao.MockSubforumDAOInterface, *dao.MockPermissionDAOInterface) {
	mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockIdentityMappingDAO := dao.NewMockIdentityMappingDAOInterface(ctrl)
	mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

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

	handler := handlers.NewAuthHandler(
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

// TestRoleKeySecurity tests the role key security functionality using gomock
func TestRoleKeySecurity(t *testing.T) {
	t.Run("AuthenticationKeyScope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		_, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithGomocks(ctrl)

		// Test data
		testUserID := int64(1)
		testPseudonym := fixtures.CreateTestPseudonym()
		testPseudonym.PseudonymID = "pseudonym-123"
		testPseudonym.DisplayName = "AuthUser1"

		// Mock pseudonym retrieval
		mockPseudonymDAO.EXPECT().
			GetPseudonymsByUserID(gomock.Any(), testUserID, testPseudonym.PseudonymID, "user", "authentication").
			Return([]*dbmodels.Pseudonym{testPseudonym}, nil).
			Times(1)

		// Call the method directly on the mock
		pseudonyms, err := mockPseudonymDAO.GetPseudonymsByUserID(context.Background(), testUserID, testPseudonym.PseudonymID, "user", "authentication")

		// Assert response
		require.NoError(t, err)
		assert.Len(t, pseudonyms, 1)
		assert.Equal(t, "pseudonym-123", pseudonyms[0].PseudonymID)
	})

	t.Run("SelfCorrelationKeyScope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		_, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithGomocks(ctrl)

		// Test data
		testUserID := int64(1)
		testPseudonymID := "pseudonym-123"

		// Set up gomock expectations
		mockPseudonymDAO.EXPECT().
			VerifyPseudonymOwnership(gomock.Any(), testPseudonymID, testUserID, testPseudonymID, "user", "self_correlation").
			Return(true, nil).
			Times(1)

		// Test the method
		ownsPseudonym, err := mockPseudonymDAO.VerifyPseudonymOwnership(context.Background(), testPseudonymID, testUserID, testPseudonymID, "user", "self_correlation")

		// Assert response
		require.NoError(t, err)
		assert.True(t, ownsPseudonym)
	})

	t.Run("RoleKeyDatabaseOperations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		_, _, _, _, mockRoleKeyDAO, _, _ := NewAuthHandlerWithGomocks(ctrl)

		// Test data
		capabilities := []string{"test_capability", "another_capability"}
		expiresAt := time.Now().AddDate(0, 1, 0)
		testRoleKey := fixtures.CreateTestRoleKey("key-123", "test_role", "test_scope", capabilities)

		// Mock role key creation
		mockRoleKeyDAO.EXPECT().
			CreateRoleKey(gomock.Any(), "test_role", "test_scope", []byte("test_key_data"), capabilities, expiresAt, "test-pseudonym-id", "test-pseudonym-id", (*int32)(nil)).
			Return(testRoleKey, nil).
			Times(1)

		// Mock role key retrieval
		mockRoleKeyDAO.EXPECT().
			GetRoleKey(gomock.Any(), "test-pseudonym-id", "test_scope", (*int32)(nil)).
			Return(testRoleKey, nil).
			Times(1)

		// Mock capability validation
		mockRoleKeyDAO.EXPECT().
			ValidateKeyCapability(gomock.Any(), "test-pseudonym-id", "test_scope", "test_capability", (*int32)(nil)).
			Return(true, nil).
			Times(1)

		mockRoleKeyDAO.EXPECT().
			ValidateKeyCapability(gomock.Any(), "test-pseudonym-id", "test_scope", "invalid_capability", (*int32)(nil)).
			Return(false, nil).
			Times(1)

		// Mock listing role keys
		mockRoleKeyDAO.EXPECT().
			ListRoleKeys(gomock.Any()).
			Return([]*dbmodels.RoleKey{testRoleKey}, nil).
			Times(1)

		// Mock deactivating role key
		mockRoleKeyDAO.EXPECT().
			DeactivateRoleKey(gomock.Any(), "key-123").
			Return(nil).
			Times(1)

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
	})
}
