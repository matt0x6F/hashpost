package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAPIKeyDAO is a mock implementation of the API key DAO for testing
type MockAPIKeyDAO struct {
	mock.Mock
}

func (m *MockAPIKeyDAO) ValidateAPIKey(ctx context.Context, tokenString string) (*dao.APIKeyPermissions, string, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*dao.APIKeyPermissions), args.String(1), args.Error(2)
}

// Helper function to create test user handler for blocking tests
func createTestUserHandlerForBlocking() (*UserHandler, *MockSecurePseudonymDAO, *MockUserBlocksDAO) {
	mockSecurePseudonymDAO := &MockSecurePseudonymDAO{}
	mockUserBlocksDAO := &MockUserBlocksDAO{}

	handler := &UserHandler{
		securePseudonymDAO: mockSecurePseudonymDAO,
		userBlocksDAO:      mockUserBlocksDAO,
	}

	return handler, mockSecurePseudonymDAO, mockUserBlocksDAO
}

// Helper function to create test user context for blocking tests
func createTestUserContextForBlocking(userID int64, activePseudonymID string) *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		Roles:             []string{"user"},
		Capabilities:      []string{"create_content", "vote", "message", "report"},
		ActivePseudonymID: activePseudonymID,
		DisplayName:       "TestUser",
	}
}

// Helper function to create test pseudonym for blocking tests
func createTestPseudonymForBlocking(pseudonymID, displayName string) *dbmodels.Pseudonym {
	return &dbmodels.Pseudonym{
		PseudonymID: pseudonymID,
		DisplayName: displayName,
		IsActive:    sql.Null[bool]{V: true, Valid: true},
		CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// Helper function to create test user block
func createTestUserBlock(blockID int64, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) *dbmodels.UserBlock {
	block := &dbmodels.UserBlock{
		BlockID:            blockID,
		BlockerPseudonymID: blockerPseudonymID,
		CreatedAt:          sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	if blockedPseudonymID != "" {
		block.BlockedPseudonymID = sql.Null[string]{V: blockedPseudonymID, Valid: true}
		block.BlockedUserID = sql.Null[int64]{Valid: false}
	} else {
		block.BlockedPseudonymID = sql.Null[string]{Valid: false}
		block.BlockedUserID = sql.Null[int64]{V: blockedUserID, Valid: true}
	}

	return block
}

// Helper function to generate a valid JWT token for testing
func generateTestJWTToken(userID int64, activePseudonymID string) string {
	// Create a JWT token using the actual JWT generation logic
	// This ensures the token is valid for the test environment
	claims := &middleware.JWTClaims{
		UserID:            userID,
		Email:             "test@example.com",
		Roles:             []string{"user"},
		Capabilities:      []string{"create_content", "vote", "message", "report", "create_subforum"},
		MFAEnabled:        false,
		ActivePseudonymID: activePseudonymID,
		DisplayName:       "TestUser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		panic("Failed to generate test JWT token: " + err.Error())
	}

	return tokenString
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
		handler, mockSecurePseudonymDAO, mockUserBlocksDAO := createTestUserHandlerForBlocking()

		// Test data
		blockerUserID := int64(1)
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"

		// Mock blocked pseudonym retrieval
		mockBlockedPseudonym := createTestPseudonymForBlocking(blockedPseudonymID, "BlockedUser")
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, blockedPseudonymID).Return(mockBlockedPseudonym, nil)

		// Mock ownership verification (should return false since it's not the same user)
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, blockedPseudonymID, blockerUserID, "user", "self_correlation").Return(false, nil)

		// Mock user block creation
		mockUserBlock := createTestUserBlock(1, blockerPseudonymID, blockedPseudonymID, 0)
		mockUserBlocksDAO.On("CreateUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID, int64(0)).Return(mockUserBlock, nil)

		// Create input
		blockAllPersonas := false
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWTToken(blockerUserID, blockerPseudonymID),
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
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})

	t.Run("BlockFingerprintLevel", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, mockUserBlocksDAO := createTestUserHandlerForBlocking()

		// Test data
		blockerUserID := int64(1)
		blockedUserID := int64(2)
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"

		// Mock blocked pseudonym retrieval
		mockBlockedPseudonym := createTestPseudonymForBlocking(blockedPseudonymID, "BlockedUser")
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, blockedPseudonymID).Return(mockBlockedPseudonym, nil)

		// Mock ownership verification (should return false since it's not the same user)
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, blockedPseudonymID, blockerUserID, "user", "self_correlation").Return(false, nil)

		// Mock getting user ID by pseudonym
		mockSecurePseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, blockedPseudonymID, "user", "self_correlation").Return(blockedUserID, nil)

		// Mock user block creation (fingerprint level)
		mockUserBlock := createTestUserBlock(1, blockerPseudonymID, "", blockedUserID)
		mockUserBlocksDAO.On("CreateUserBlock", mock.Anything, blockerPseudonymID, "", blockedUserID).Return(mockUserBlock, nil)

		// Create input
		blockAllPersonas := true
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWTToken(blockerUserID, blockerPseudonymID),
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
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})

	t.Run("BlockSelfPrevention", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, _ := createTestUserHandlerForBlocking()

		// Test data
		userID := int64(1)
		pseudonymID := "self-pseudonym-123"

		// Mock pseudonym retrieval
		mockPseudonym := createTestPseudonymForBlocking(pseudonymID, "SelfUser")
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, pseudonymID).Return(mockPseudonym, nil)

		// Mock ownership verification (should return true since it's the same user)
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", mock.Anything, pseudonymID, userID, "user", "self_correlation").Return(true, nil)

		// Create input
		blockAllPersonas := true
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWTToken(userID, pseudonymID),
			},
			PseudonymIDPathParam: apimodels.PseudonymIDPathParam{
				PseudonymID: pseudonymID,
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
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Cannot block yourself")

		// Verify mocks were called
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("BlockNonExistentPseudonym", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, _ := createTestUserHandlerForBlocking()

		// Test data
		blockedPseudonymID := "nonexistent-pseudonym-123"

		// Mock pseudonym retrieval (should return nil)
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, blockedPseudonymID).Return(nil, nil)

		// Create input
		blockAllPersonas := false
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
			apimodels.BlockUserInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWTToken(1, "blocker-pseudonym-123"),
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
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Blocked pseudonym not found")

		// Verify mocks were called
		mockSecurePseudonymDAO.AssertExpectations(t)
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
		handler, _, mockUserBlocksDAO := createTestUserHandlerForBlocking()

		// Test data
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"

		// Mock existing direct block
		mockUserBlock := createTestUserBlock(1, blockerPseudonymID, blockedPseudonymID, 0)
		mockUserBlocksDAO.On("GetUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(mockUserBlock, nil)

		// Mock block deletion
		mockUserBlocksDAO.On("DeleteUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWTToken(1, blockerPseudonymID),
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
	})

	t.Run("UnblockFingerprintLevel", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, mockUserBlocksDAO := createTestUserHandlerForBlocking()

		// Test data
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"
		blockedUserID := int64(2)

		// Mock no direct block found
		mockUserBlocksDAO.On("GetUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(nil, nil)

		// Mock getting user ID by pseudonym
		mockSecurePseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, blockedPseudonymID, "user", "self_correlation").Return(blockedUserID, nil)

		// Mock fingerprint-level blocks
		mockUserBlock := createTestUserBlock(1, blockerPseudonymID, "", blockedUserID)
		mockUserBlocksDAO.On("GetFingerprintLevelBlocks", mock.Anything, blockedUserID).Return([]*dbmodels.UserBlock{mockUserBlock}, nil)

		// Mock block deletion by ID
		mockUserBlocksDAO.On("DeleteUserBlockByID", mock.Anything, int64(1)).Return(nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWTToken(1, blockerPseudonymID),
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
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})

	t.Run("UnblockNonExistentBlock", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, mockUserBlocksDAO := createTestUserHandlerForBlocking()

		// Test data
		blockerPseudonymID := "blocker-pseudonym-123"
		blockedPseudonymID := "blocked-pseudonym-456"
		blockedUserID := int64(2)

		// Mock no direct block found
		mockUserBlocksDAO.On("GetUserBlock", mock.Anything, blockerPseudonymID, blockedPseudonymID).Return(nil, nil)

		// Mock getting user ID by pseudonym
		mockSecurePseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, blockedPseudonymID, "user", "self_correlation").Return(blockedUserID, nil)

		// Mock no fingerprint-level blocks found
		mockUserBlocksDAO.On("GetFingerprintLevelBlocks", mock.Anything, blockedUserID).Return([]*dbmodels.UserBlock{}, nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.PseudonymIDPathParam
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWTToken(1, blockerPseudonymID),
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
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockUserBlocksDAO.AssertExpectations(t)
	})
}
