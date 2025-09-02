package services

import (
	"context"
	"testing"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestKeyManagementService_EnsureMessagingKeys(t *testing.T) {
	// Create gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
	mockIBESystem := createTestIBESystem()
	encryptionService := NewEncryptionService()

	service := NewKeyManagementService(encryptionService, mockRoleKeyDAO, mockIBESystem)

	ctx := context.Background()
	userID := int64(123)
	pseudonymID := "test-pseudonym-123"

	// Set up gomock expectations
	mockRoleKeyDAO.EXPECT().
		GetRoleKey(ctx, pseudonymID, constants.ScopeMessaging, nil).
		Return(nil, assert.AnError) // Simulate no existing key

	mockRoleKeyDAO.EXPECT().
		CreateRoleKey(
			ctx,
			"user",
			constants.ScopeMessaging,
			gomock.Any(), // keyData
			[]string{
				constants.CapabilitySendDirectMessages,
				constants.CapabilityReceiveDirectMessages,
				constants.CapabilityManageConversationKeys,
			},
			gomock.Any(), // expiresAt
			pseudonymID,
			pseudonymID,
			nil, // subforumID
		).
		Return(&dbmodels.RoleKey{}, nil)

	// Test key generation
	err := service.EnsureMessagingKeys(ctx, userID, pseudonymID)
	require.NoError(t, err)
}

func TestKeyManagementService_EnsureMessagingKeys_Error(t *testing.T) {
	// Create gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies with error
	mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
	mockIBESystem := createTestIBESystem()
	encryptionService := NewEncryptionService()

	service := NewKeyManagementService(encryptionService, mockRoleKeyDAO, mockIBESystem)

	ctx := context.Background()
	userID := int64(123)
	pseudonymID := "test-pseudonym-123"

	// Set up gomock expectations for error case
	mockRoleKeyDAO.EXPECT().
		GetRoleKey(ctx, pseudonymID, constants.ScopeMessaging, nil).
		Return(nil, assert.AnError) // Simulate no existing key

	mockRoleKeyDAO.EXPECT().
		CreateRoleKey(
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		).
		Return(nil, assert.AnError) // Simulate creation error

	// Test key generation with error
	err := service.EnsureMessagingKeys(ctx, userID, pseudonymID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create messaging role key")
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
