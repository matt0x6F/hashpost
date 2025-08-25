package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// NewUserHandlerWithMocksForBlockingForBlocking creates a new user handler with mock DAOs for blocking tests
func NewUserHandlerWithMocksForBlocking() (*UserHandler, *mocks.MockPseudonymDAO, *mocks.MockUserBlocksDAO) {
	mockPseudonymDAO := mocks.NewMockPseudonymDAO()
	mockUserBlocksDAO := &mocks.MockUserBlocksDAO{}

	handler := &UserHandler{
		pseudonymDAO:  mockPseudonymDAO,
		userBlocksDAO: mockUserBlocksDAO,
	}

	return handler, mockPseudonymDAO, mockUserBlocksDAO
}

// TestUserHandler_BlockUser tests the user blocking functionality
func TestUserHandler_BlockUser(t *testing.T) {
	// Initialize global auth middleware for testing with nil API key DAO since tests use JWT
	middleware.SetGlobalAuthMiddleware(middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
		Secret:      "test-secret",
		Expiration:  time.Hour,
		Development: true,
	}, &config.SecurityConfig{}))

	t.Run("BlockPseudonymLevel", func(t *testing.T) {
		handler, mockPseudonymDAO, mockUserBlocksDAO := NewUserHandlerWithMocksForBlocking()

		// Test data
		blockerUserID := int64(1)
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"

		// Mock blocked pseudonym retrieval
		mockBlockedPseudonym := fixtures.CreateTestPseudonymForBlocking(blockedPseudonymID, "BlockedUser")
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, blockedPseudonymID).Return(mockBlockedPseudonym, nil)

		// Set up mock expectations for pseudonym ownership verification
		mockPseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, blockedPseudonymID, blockerUserID, "blocker-pseudonym-123", "user", "self_correlation").Return(false, nil)

		// Mock user block creation
		mockUserBlock := fixtures.CreateTestUserBlock(1, blockerPseudonymID, blockedPseudonymID, 0)
		mockUserBlocksDAO.On("CreateUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID, int64(0)).Return(mockUserBlock, nil)

		// Set up mock expectation for UpdateLastActive
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, blockerPseudonymID).Return(nil)

		// Create input
		blockAllPersonas := false
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(blockerUserID, blockerPseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: blockedPseudonymID,
			},
			BlockUserInput: apimodels.BlockUserInput{
				Body: apimodels.BlockUserBody{
					BlockAllPersonas: &blockAllPersonas,
				},
			},
		}

		// Call the handler
		response, err := handler.BlockUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, blockedPseudonymID, response.Body.BlockedPseudonymID)

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})

	t.Run("BlockFingerprintLevel", func(t *testing.T) {
		handler, mockPseudonymDAO, mockUserBlocksDAO := NewUserHandlerWithMocksForBlocking()

		// Test data
		blockerUserID := int64(1)
		blockedUserID := int64(2)
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"

		// Mock blocked pseudonym retrieval
		mockBlockedPseudonym := fixtures.CreateTestPseudonymForBlocking(blockedPseudonymID, "BlockedUser")
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, blockedPseudonymID).Return(mockBlockedPseudonym, nil)

		// Set up mock expectations for pseudonym ownership verification
		mockPseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, blockedPseudonymID, blockerUserID, "blocker-pseudonym-123", "user", "self_correlation").Return(false, nil)

		// Mock getting user ID by pseudonym
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, blockedPseudonymID, "user", "self_correlation").Return(blockedUserID, nil)

		// Mock user block creation (fingerprint level)
		mockUserBlock := fixtures.CreateTestUserBlock(1, blockerPseudonymID, "", blockedUserID)
		mockUserBlocksDAO.On("CreateUserBlock", mock.Anything, blockerPseudonymID, "", blockedUserID).Return(mockUserBlock, nil)

		// Set up mock expectation for UpdateLastActive
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, blockerPseudonymID).Return(nil)

		// Create input
		blockAllPersonas := true
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(blockerUserID, blockerPseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: blockedPseudonymID,
			},
			BlockUserInput: apimodels.BlockUserInput{
				Body: apimodels.BlockUserBody{
					BlockAllPersonas: &blockAllPersonas,
				},
			},
		}

		// Call the handler
		response, err := handler.BlockUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, blockedPseudonymID, response.Body.BlockedPseudonymID)

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})

	t.Run("BlockSelfPrevention", func(t *testing.T) {
		handler, mockPseudonymDAO, mockUserBlocksDAO := NewUserHandlerWithMocksForBlocking()

		// Test data
		blockerUserID := int64(1)
		blockerPseudonymID := "self-pseudonym-123"
		blockedPseudonymID := "self-pseudonym-123" // Same as blocker

		// Mock blocked pseudonym retrieval
		mockBlockedPseudonym := fixtures.CreateTestPseudonymForBlocking(blockedPseudonymID, "SelfUser")
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, blockedPseudonymID).Return(mockBlockedPseudonym, nil)

		// Set up mock expectations for pseudonym ownership verification
		mockPseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, blockedPseudonymID, blockerUserID, "self-pseudonym-123", "user", "self_correlation").Return(true, nil)

		// Create input
		blockAllPersonas := false
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(blockerUserID, blockerPseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: blockedPseudonymID,
			},
			BlockUserInput: apimodels.BlockUserInput{
				Body: apimodels.BlockUserBody{
					BlockAllPersonas: &blockAllPersonas,
				},
			},
		}

		// Call the handler
		_, err := handler.BlockUser(context.Background(), input)

		// Should return an error for self-blocking
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot block yourself")

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})

	t.Run("BlockNonExistentPseudonym", func(t *testing.T) {
		handler, mockPseudonymDAO, mockUserBlocksDAO := NewUserHandlerWithMocksForBlocking()

		// Test data
		blockerUserID := int64(1)
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "nonexistent-pseudonym-123"

		// Mock blocked pseudonym retrieval (should return nil, not found)
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, blockedPseudonymID).Return(nil, nil)

		// Create input
		blockAllPersonas := false
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(blockerUserID, blockerPseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: blockedPseudonymID,
			},
			BlockUserInput: apimodels.BlockUserInput{
				Body: apimodels.BlockUserBody{
					BlockAllPersonas: &blockAllPersonas,
				},
			},
		}

		// Call the handler
		_, err := handler.BlockUser(context.Background(), input)

		// Should return an error for non-existent pseudonym
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})
}

// TestUserHandler_UnblockUser tests the user unblocking functionality
func TestUserHandler_UnblockUser(t *testing.T) {
	// Initialize global auth middleware for testing
	middleware.SetGlobalAuthMiddleware(middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
		Secret:      "test-secret",
		Expiration:  time.Hour,
		Development: true,
	}, &config.SecurityConfig{}))

	t.Run("UnblockPseudonymLevel", func(t *testing.T) {
		handler, mockPseudonymDAO, mockUserBlocksDAO := NewUserHandlerWithMocksForBlocking()

		// Test data
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"

		// Mock existing direct block
		mockUserBlock := fixtures.CreateTestUserBlock(1, blockerPseudonymID, blockedPseudonymID, 0)
		mockUserBlocksDAO.On("GetUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(mockUserBlock, nil)

		// Mock block deletion
		mockUserBlocksDAO.On("DeleteUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(nil)

		// Set up mock expectation for UpdateLastActive
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, blockerPseudonymID).Return(nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, blockerPseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: blockedPseudonymID,
			},
		}

		// Call the handler
		response, err := handler.UnblockUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, blockedPseudonymID, response.Body.BlockedPseudonymID)

		// Verify mocks were called
		mockUserBlocksDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("UnblockFingerprintLevel", func(t *testing.T) {
		handler, mockPseudonymDAO, mockUserBlocksDAO := NewUserHandlerWithMocksForBlocking()

		// Test data
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"
		blockedUserID := int64(2)

		// Mock no direct block found
		mockUserBlocksDAO.On("GetUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(nil, nil)

		// Mock getting user ID by pseudonym
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, blockedPseudonymID, "user", "self_correlation").Return(blockedUserID, nil)

		// Mock fingerprint-level blocks
		mockUserBlock := fixtures.CreateTestUserBlock(1, blockerPseudonymID, "", blockedUserID)
		mockUserBlocksDAO.On("GetFingerprintLevelBlocks", mock.Anything, blockedUserID).Return([]*dbmodels.UserBlock{mockUserBlock}, nil)

		// Mock block deletion by ID
		mockUserBlocksDAO.On("DeleteUserBlockByID", mock.Anything, int64(1)).Return(nil)

		// Set up mock expectation for UpdateLastActive
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, blockerPseudonymID).Return(nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, blockerPseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: blockedPseudonymID,
			},
		}

		// Call the handler
		response, err := handler.UnblockUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, blockedPseudonymID, response.Body.BlockedPseudonymID)

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})

	t.Run("UnblockNonExistentBlock", func(t *testing.T) {
		handler, mockPseudonymDAO, mockUserBlocksDAO := NewUserHandlerWithMocksForBlocking()

		// Test data
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"
		blockedUserID := int64(2)

		// Mock no direct block found
		mockUserBlocksDAO.On("GetUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(nil, nil)

		// Mock getting user ID by pseudonym
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, blockedPseudonymID, "user", "self_correlation").Return(blockedUserID, nil)

		// Mock no fingerprint-level blocks found
		mockUserBlocksDAO.On("GetFingerprintLevelBlocks", mock.Anything, blockedUserID).Return([]*dbmodels.UserBlock{}, nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, blockerPseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: blockedPseudonymID,
			},
		}

		// Call the handler
		response, err := handler.UnblockUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Block not found")

		// Verify mocks were called
		mockPseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})
}
