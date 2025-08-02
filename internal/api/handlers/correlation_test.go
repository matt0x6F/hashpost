package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create test correlation handler with mocks
func NewCorrelationHandlerWithMocks() (*CorrelationHandler, *mocks.MockSecurePseudonymDAO, *mocks.MockIdentityMappingDAO, *mocks.MockPostDAO, *mocks.MockCommentDAO, *mocks.MockSubforumDAO, *mocks.MockCorrelationAuditDAO) {
	mockSecurePseudonymDAO := &mocks.MockSecurePseudonymDAO{}
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockPostDAO := &mocks.MockPostDAO{}
	mockCommentDAO := &mocks.MockCommentDAO{}
	mockSubforumDAO := &mocks.MockSubforumDAO{}
	mockCorrelationAuditDAO := &mocks.MockCorrelationAuditDAO{}

	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	handler := NewCorrelationHandler(
		nil, // No direct DB needed since we use DAOs
		ibeSystem,
		mockSecurePseudonymDAO,
		mockIdentityMappingDAO,
		mockPostDAO,
		mockCommentDAO,
		mockSubforumDAO,
		mockCorrelationAuditDAO,
	)

	return handler, mockSecurePseudonymDAO, mockIdentityMappingDAO, mockPostDAO, mockCommentDAO, mockSubforumDAO, mockCorrelationAuditDAO
}

// Helper function to create test identity mapping
func createTestIdentityMapping(pseudonymID string, encryptedIdentity []byte) *dbmodels.IdentityMapping {
	mappingID := uuid.Must(uuid.NewV4())
	return &dbmodels.IdentityMapping{
		MappingID:             mappingID,
		PseudonymID:           pseudonymID,
		EncryptedRealIdentity: encryptedIdentity,
		KeyScope:              "user",
		IsActive:              sql.Null[bool]{V: true, Valid: true},
		CreatedAt:             sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// setupTestAuthMiddleware sets up the global auth middleware for testing
func setupTestAuthMiddleware() {
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
		Secret:      "test-secret",
		Expiration:  time.Hour,
		Development: true,
	}, &config.SecurityConfig{
		EnableMFA: false,
	})
	middleware.SetGlobalAuthMiddleware(authMiddleware)
}

// createAuthenticatedIdentityInput creates an input with a valid JWT token for identity correlation testing
func createAuthenticatedIdentityInput(userID int64, email string, capabilities []string, activePseudonymID string, displayName string) *apimodels.IdentityCorrelationInput {
	// Create a simple JWT token for testing
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sImNhcGFiaWxpdGllcyI6WyJjb3JyZWxhdGVfaWRlbnRpdGllcyJdLCJtZmFfZW5hYmxlZCI6ZmFsc2UsImFjdGl2ZV9wc2V1ZG9ueW1faWQiOiJ0ZXN0LXBzZXVkb255bS0xMjMiLCJkaXNwbGF5X25hbWUiOiJUZXN0VXNlciIsImV4cCI6MTc1NDEwNjA1OCwibmJmIjoxNzU0MTAyNDU4LCJpYXQiOjE3NTQxMDI0NTh9.test_signature"

	return &apimodels.IdentityCorrelationInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		Body: apimodels.IdentityCorrelationInputBody{
			RequestedPseudonym:   "target-pseudonym-123",
			RequestedFingerprint: "test-fingerprint-456",
			Justification:        "Investigation of identity correlation",
		},
	}
}

// createAuthenticatedHistoryInput creates an input with a valid JWT token for correlation history testing
func createAuthenticatedHistoryInput(userID int64, email string, capabilities []string, activePseudonymID string, displayName string) *apimodels.CorrelationHistoryInput {
	// Create a simple JWT token for testing
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sImNhcGFiaWxpdGllcyI6WyJ2aWV3X2NvcnJlbGF0aW9uX2hpc3RvcnkiXSwibWZhX2VuYWJsZWQiOmZhbHNlLCJhY3RpdmVfcHNldWRvbnltX2lkIjoidGVzdC1wc2V1ZG9ueW0tMTIzIiwiZGlzcGxheV9uYW1lIjoiVGVzdFVzZXIiLCJleHAiOjE3NTQxMDYwNTgsIm5iZiI6MTc1NDEwMjQ1OCwiaWF0IjoxNzU0MTAyNDU4fQ.test_signature"

	return &apimodels.CorrelationHistoryInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		CorrelationType: "fingerprint",
		Page:            1,
		Limit:           25,
	}
}

func TestCorrelationHandler_RequestFingerprintCorrelation(t *testing.T) {
	// Set up global auth middleware for testing
	setupTestAuthMiddleware()

	t.Run("Success", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, mockIdentityMappingDAO, mockPostDAO, mockCommentDAO, _, mockCorrelationAuditDAO := NewCorrelationHandlerWithMocks()

		// Test data
		adminUserID := int64(1)
		requestedPseudonymID := "target-pseudonym-123"
		subforumID := 1

		// Create test user context with correlation capability
		userCtx := fixtures.CreateTestUserContext()
		userCtx.UserID = adminUserID
		userCtx.Email = "admin@example.com"
		userCtx.Capabilities = []string{"correlate_fingerprints"}

		// Mock pseudonym retrieval
		testPseudonym := fixtures.CreateTestPseudonym()
		testPseudonym.PseudonymID = requestedPseudonymID
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, requestedPseudonymID).Return(testPseudonym, nil)

		// Use IBE system to encrypt the identity mapping
		ibeSystem := handler.ibeSystem
		adminKey := ibeSystem.GenerateRoleKey("moderator", "subforum_correlation", time.Now().AddDate(0, 1, 0))

		// Generate the actual fingerprint that will be used
		realIdentity := "test-fingerprint-456:target-pseudonym-123"
		fingerprint := ibeSystem.GenerateFingerprint(realIdentity)

		// Use the actual fingerprint for encryption
		plaintext := fingerprint + ":" + requestedPseudonymID
		encryptedBytes, err := ibeSystem.EncryptIdentity(plaintext, requestedPseudonymID, adminKey)
		require.NoError(t, err)
		testMapping := createTestIdentityMapping(requestedPseudonymID, encryptedBytes)
		mockIdentityMappingDAO.On("GetIdentityMappingByPseudonymID", mock.Anything, requestedPseudonymID).Return(testMapping, nil)

		// Mock related identity mappings (same fingerprint)
		relatedPseudonymID := "related-pseudonym-789"
		relatedPlaintext := fingerprint + ":" + relatedPseudonymID
		relatedEncryptedBytes, err := ibeSystem.EncryptIdentity(relatedPlaintext, relatedPseudonymID, adminKey)
		require.NoError(t, err)
		relatedMapping := createTestIdentityMapping(relatedPseudonymID, relatedEncryptedBytes)
		relatedMappings := dbmodels.IdentityMappingSlice{testMapping, relatedMapping}

		// The handler will call GenerateFingerprint on the decrypted mapping
		// which will be "fingerprint:pseudonymID", so we need to mock for that fingerprint
		decryptedMapping := fingerprint + ":" + requestedPseudonymID
		actualFingerprint := ibeSystem.GenerateFingerprint(decryptedMapping)
		mockIdentityMappingDAO.On("GetIdentityMappingsByFingerprint", mock.Anything, actualFingerprint).Return(relatedMappings, nil)

		// Mock pseudonym retrieval for related pseudonyms
		relatedPseudonym := fixtures.CreateTestPseudonym()
		relatedPseudonym.PseudonymID = relatedPseudonymID
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, relatedPseudonymID).Return(relatedPseudonym, nil)

		// Mock post and comment counts for subforum (using available methods)
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(5), nil).Maybe()
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(3), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(10), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(7), nil).Maybe()

		// Mock the specific subforum count methods that the handler calls
		mockPostDAO.On("CountPostsByPseudonymInSubforum", mock.Anything, requestedPseudonymID, int32(subforumID)).Return(int64(5), nil).Maybe()
		mockPostDAO.On("CountPostsByPseudonymInSubforum", mock.Anything, relatedPseudonymID, int32(subforumID)).Return(int64(3), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonymInSubforum", mock.Anything, requestedPseudonymID, int32(subforumID)).Return(int64(10), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonymInSubforum", mock.Anything, relatedPseudonymID, int32(subforumID)).Return(int64(7), nil).Maybe()

		// Mock pseudonym retrieval for each result
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, requestedPseudonymID).Return(testPseudonym, nil).Maybe()
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, relatedPseudonymID).Return(relatedPseudonym, nil).Maybe()

		// Mock additional calls that the handler makes for each pseudonym in results
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(5), nil).Maybe()
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(3), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(10), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(7), nil).Maybe()

		// Mock the final calls that the handler makes for each pseudonym in results
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(5), nil).Maybe()
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(3), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(10), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(7), nil).Maybe()

		// Mock the correlation audit DAO
		mockCorrelationAuditDAO.On("CreateCorrelationAudit", mock.Anything, mock.AnythingOfType("*models.CorrelationAuditSetter")).Return(nil).Maybe()

		// Create authenticated input
		// Create authenticated input with correlation capability
		token, err := fixtures.GenerateTestJWTToken(adminUserID, "test-pseudonym-123", "TestUser", "admin@example.com", []string{"user"}, []string{"correlate_fingerprints"})
		require.NoError(t, err)

		input := &apimodels.FingerprintCorrelationInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			Body: apimodels.FingerprintCorrelationInputBody{
				RequestedPseudonym: requestedPseudonymID,
				Justification:      "Investigation of ban evasion",
				SubforumID:         subforumID,
				IncidentID:         "ban_evasion_123",
			},
		}

		// Call the method
		result, err := handler.RequestFingerprintCorrelation(context.Background(), input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.Body.CorrelationID)
		assert.Len(t, result.Body.Results, 2)
		assert.NotEmpty(t, result.Body.AuditID)

		// Verify the results contain both pseudonyms
		pseudonymIDs := make(map[string]bool)
		for _, result := range result.Body.Results {
			pseudonymIDs[result.PseudonymID] = true
		}
		assert.True(t, pseudonymIDs[requestedPseudonymID])
		assert.True(t, pseudonymIDs[relatedPseudonymID])

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockIdentityMappingDAO.AssertExpectations(t)
		mockPostDAO.AssertExpectations(t)
		mockCommentDAO.AssertExpectations(t)
		mockCorrelationAuditDAO.AssertExpectations(t)
	})

	t.Run("InsufficientPermissions", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewCorrelationHandlerWithMocks()

		// Create input without correlation capability
		// Create input with insufficient capabilities
		token, err := fixtures.GenerateTestJWTToken(1, "test-pseudonym-123", "TestUser", "user@example.com", []string{"user"}, []string{"create_content"})
		require.NoError(t, err)

		input := &apimodels.FingerprintCorrelationInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			Body: apimodels.FingerprintCorrelationInputBody{
				RequestedPseudonym: "target-pseudonym-123",
				Justification:      "Investigation of ban evasion",
				SubforumID:         1,
				IncidentID:         "ban_evasion_123",
			},
		}

		// Call the method
		result, err := handler.RequestFingerprintCorrelation(context.Background(), input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "insufficient permissions")
		assert.Contains(t, err.Error(), "correlate_fingerprints capability required")
	})

	t.Run("PseudonymNotFound", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, _, _, _, _, _ := NewCorrelationHandlerWithMocks()

		// Mock pseudonym not found
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, "nonexistent-pseudonym").Return(nil, nil).Maybe()

		// Create authenticated input with correlation capability
		// Create input with correlation capability
		token, err := fixtures.GenerateTestJWTToken(1, "test-pseudonym-123", "TestUser", "admin@example.com", []string{"user"}, []string{"correlate_fingerprints"})
		require.NoError(t, err)

		input := &apimodels.FingerprintCorrelationInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			Body: apimodels.FingerprintCorrelationInputBody{
				RequestedPseudonym: "nonexistent-pseudonym",
				Justification:      "Investigation of ban evasion",
				SubforumID:         1,
				IncidentID:         "ban_evasion_123",
			},
		}

		// Call the method
		result, err := handler.RequestFingerprintCorrelation(context.Background(), input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "pseudonym not found")

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
	})
}

func TestCorrelationHandler_RequestIdentityCorrelation(t *testing.T) {
	// Set up global auth middleware for testing
	setupTestAuthMiddleware()

	t.Run("Success", func(t *testing.T) {
		handler, mockSecurePseudonymDAO, mockIdentityMappingDAO, mockPostDAO, mockCommentDAO, mockSubforumDAO, mockCorrelationAuditDAO := NewCorrelationHandlerWithMocks()

		// Test data
		adminUserID := int64(1)
		requestedPseudonymID := "target-pseudonym-123"

		// Mock pseudonym retrieval
		testPseudonym := fixtures.CreateTestPseudonym()
		testPseudonym.PseudonymID = requestedPseudonymID
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, requestedPseudonymID).Return(testPseudonym, nil)

		// Use IBE system to encrypt the identity mapping
		ibeSystem := handler.ibeSystem
		adminKey := ibeSystem.GenerateRoleKey("site_admin", "full_correlation", time.Now().AddDate(0, 1, 0))

		// Generate the actual fingerprint that will be used
		realIdentity := "test-fingerprint-456:target-pseudonym-123"
		fingerprint := ibeSystem.GenerateFingerprint(realIdentity)

		// Use the actual fingerprint for encryption
		plaintext := fingerprint + ":" + requestedPseudonymID
		encryptedBytes, err := ibeSystem.EncryptIdentity(plaintext, requestedPseudonymID, adminKey)
		require.NoError(t, err)
		testMapping := createTestIdentityMapping(requestedPseudonymID, encryptedBytes)
		mockIdentityMappingDAO.On("GetIdentityMappingByPseudonymID", mock.Anything, requestedPseudonymID).Return(testMapping, nil)

		// Mock related identity mappings (same fingerprint)
		relatedPseudonymID := "related-pseudonym-789"
		relatedPlaintext := fingerprint + ":" + relatedPseudonymID
		relatedEncryptedBytes, err := ibeSystem.EncryptIdentity(relatedPlaintext, relatedPseudonymID, adminKey)
		require.NoError(t, err)
		relatedMapping := createTestIdentityMapping(relatedPseudonymID, relatedEncryptedBytes)
		relatedMappings := dbmodels.IdentityMappingSlice{testMapping, relatedMapping}

		// The handler will call GenerateFingerprint on the decrypted mapping
		// which will be "fingerprint:pseudonymID", so we need to mock for that fingerprint
		decryptedMapping := fingerprint + ":" + requestedPseudonymID
		actualFingerprint := ibeSystem.GenerateFingerprint(decryptedMapping)
		mockIdentityMappingDAO.On("GetIdentityMappingsByFingerprint", mock.Anything, actualFingerprint).Return(relatedMappings, nil)

		// Mock pseudonym retrieval for related pseudonyms
		relatedPseudonym := fixtures.CreateTestPseudonym()
		relatedPseudonym.PseudonymID = relatedPseudonymID
		mockSecurePseudonymDAO.On("GetPseudonymByID", mock.Anything, relatedPseudonymID).Return(relatedPseudonym, nil)

		// Mock post and comment counts (platform-wide)
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(15), nil).Maybe()
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(8), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, requestedPseudonymID).Return(int64(25), nil).Maybe()
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, relatedPseudonymID).Return(int64(12), nil).Maybe()

		// Mock subforum activity (using available methods)
		mockPostDAO.On("GetSubforumsByPseudonym", mock.Anything, requestedPseudonymID).Return([]int32{1, 2}, nil).Maybe()
		mockPostDAO.On("GetSubforumsByPseudonym", mock.Anything, relatedPseudonymID).Return([]int32{1}, nil).Maybe()
		mockCommentDAO.On("GetSubforumsByPseudonymComments", mock.Anything, requestedPseudonymID).Return([]int32{1, 2}, nil).Maybe()
		mockCommentDAO.On("GetSubforumsByPseudonymComments", mock.Anything, relatedPseudonymID).Return([]int32{1}, nil).Maybe()

		// Mock subforum details
		subforum1 := fixtures.CreateTestSubforum()
		subforum1.SubforumID = 1
		subforum1.Name = "golang"
		subforum2 := fixtures.CreateTestSubforum()
		subforum2.SubforumID = 2
		subforum2.Name = "programming"

		mockSubforumDAO.On("GetSubforumByID", mock.Anything, int32(1)).Return(subforum1, nil).Maybe()
		mockSubforumDAO.On("GetSubforumByID", mock.Anything, int32(2)).Return(subforum2, nil).Maybe()

		// Mock the correlation audit DAO
		mockCorrelationAuditDAO.On("CreateCorrelationAudit", mock.Anything, mock.AnythingOfType("*models.CorrelationAuditSetter")).Return(nil).Maybe()

		// Create authenticated input with identity correlation capability
		token, err := fixtures.GenerateTestJWTToken(adminUserID, "test-pseudonym-123", "TestUser", "admin@example.com", []string{"user"}, []string{"correlate_identities"})
		require.NoError(t, err)

		input := &apimodels.IdentityCorrelationInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			Body: apimodels.IdentityCorrelationInputBody{
				RequestedPseudonym:   requestedPseudonymID,
				RequestedFingerprint: "test-fingerprint-456",
				Justification:        "Investigation of identity correlation",
			},
		}

		// Call the method
		result, err := handler.RequestIdentityCorrelation(context.Background(), input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.Body.CorrelationID)
		assert.Len(t, result.Body.Results, 2)
		assert.NotEmpty(t, result.Body.AuditID)

		// Verify the results contain both pseudonyms
		pseudonymIDs := make(map[string]bool)
		for _, result := range result.Body.Results {
			pseudonymIDs[result.PseudonymID] = true
		}
		assert.True(t, pseudonymIDs[requestedPseudonymID])
		assert.True(t, pseudonymIDs[relatedPseudonymID])

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockIdentityMappingDAO.AssertExpectations(t)
		mockPostDAO.AssertExpectations(t)
		mockCommentDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
		mockCorrelationAuditDAO.AssertExpectations(t)
	})

	t.Run("InsufficientPermissions", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewCorrelationHandlerWithMocks()

		// Create input with insufficient capabilities
		token, err := fixtures.GenerateTestJWTToken(1, "test-pseudonym-123", "TestUser", "user@example.com", []string{"user"}, []string{"correlate_fingerprints"})
		require.NoError(t, err)

		input := &apimodels.IdentityCorrelationInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			Body: apimodels.IdentityCorrelationInputBody{
				RequestedPseudonym:   "target-pseudonym-123",
				RequestedFingerprint: "test-fingerprint-456",
				Justification:        "Investigation of identity correlation",
			},
		}

		// Call the method
		result, err := handler.RequestIdentityCorrelation(context.Background(), input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "insufficient permissions")
		assert.Contains(t, err.Error(), "correlate_identities capability required")
	})
}

func TestCorrelationHandler_GetCorrelationHistory(t *testing.T) {
	// Set up global auth middleware for testing
	setupTestAuthMiddleware()

	t.Run("Success", func(t *testing.T) {
		handler, _, _, _, _, _, mockCorrelationAuditDAO := NewCorrelationHandlerWithMocks()

		// Mock the correlation audit DAO
		fakeAudits := dbmodels.CorrelationAuditSlice{
			&dbmodels.CorrelationAudit{
				AuditID:              uuid.Must(uuid.NewV4()),
				UserID:               1,
				PseudonymID:          "test-pseudonym",
				AdminUsername:        "admin@example.com",
				RoleUsed:             "moderator",
				RequestedPseudonym:   "target-pseudonym",
				RequestedFingerprint: sql.Null[string]{},
				Justification:        "Test justification",
				CorrelationType:      "fingerprint",
				CorrelationResult:    sql.Null[types.JSON[json.RawMessage]]{},
				Timestamp:            sql.Null[time.Time]{V: time.Now(), Valid: true},
				IncidentID:           sql.Null[string]{},
				RequestSource:        sql.Null[string]{},
			},
		}
		mockCorrelationAuditDAO.On("GetCorrelationHistory", mock.Anything, "fingerprint", 1, 25).Return(fakeAudits, nil).Maybe()

		// Create authenticated input with history viewing capability
		token, err := fixtures.GenerateTestJWTToken(1, "test-pseudonym-123", "TestUser", "admin@example.com", []string{"user"}, []string{"view_correlation_history"})
		require.NoError(t, err)

		input := &apimodels.CorrelationHistoryInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			CorrelationType: "fingerprint",
			Page:            1,
			Limit:           25,
		}

		// Call the method
		result, err := handler.GetCorrelationHistory(context.Background(), input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.Body.Correlations)
		assert.Equal(t, 1, result.Body.Pagination.Page)
		assert.Equal(t, 25, result.Body.Pagination.Limit)

		// Verify mock expectations
		mockCorrelationAuditDAO.AssertExpectations(t)
	})

	t.Run("InsufficientPermissions", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewCorrelationHandlerWithMocks()

		// Create input with insufficient capabilities
		token, err := fixtures.GenerateTestJWTToken(1, "test-pseudonym-123", "TestUser", "user@example.com", []string{"user"}, []string{"create_content"})
		require.NoError(t, err)

		input := &apimodels.CorrelationHistoryInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			CorrelationType: "fingerprint",
			Page:            1,
			Limit:           25,
		}

		// Call the method
		result, err := handler.GetCorrelationHistory(context.Background(), input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "insufficient permissions")
		assert.Contains(t, err.Error(), "view_correlation_history capability required")
	})
}

// TestCorrelationFixtures tests the fixture functions used in correlation tests
func TestCorrelationFixtures(t *testing.T) {
	t.Run("CreateTestIdentityMapping", func(t *testing.T) {
		pseudonymID := "test-pseudonym-123"
		encryptedIdentity := []byte("encrypted-data")

		mapping := createTestIdentityMapping(pseudonymID, encryptedIdentity)

		// Assertions
		assert.NotNil(t, mapping)
		assert.Equal(t, pseudonymID, mapping.PseudonymID)
		assert.Equal(t, encryptedIdentity, mapping.EncryptedRealIdentity)
		assert.Equal(t, "user", mapping.KeyScope)
		assert.True(t, mapping.IsActive.Valid)
		assert.True(t, mapping.IsActive.V)
		assert.True(t, mapping.CreatedAt.Valid)
	})

	t.Run("CreateTestPseudonymForCorrelation", func(t *testing.T) {
		pseudonymID := "test-pseudonym-123"
		displayName := "TestUser"

		pseudonym := fixtures.CreateTestPseudonym()
		pseudonym.PseudonymID = pseudonymID
		pseudonym.DisplayName = displayName

		// Assertions
		assert.NotNil(t, pseudonym)
		assert.Equal(t, pseudonymID, pseudonym.PseudonymID)
		assert.Equal(t, displayName, pseudonym.DisplayName)
		assert.True(t, pseudonym.IsActive.Valid)
		assert.True(t, pseudonym.IsActive.V)
		assert.True(t, pseudonym.CreatedAt.Valid)
	})
}

// TestCorrelationModels tests the correlation model structures
func TestCorrelationModels(t *testing.T) {
	t.Run("FingerprintCorrelationInput", func(t *testing.T) {
		input := &apimodels.FingerprintCorrelationInput{
			Body: apimodels.FingerprintCorrelationInputBody{
				RequestedPseudonym: "target-pseudonym-123",
				Justification:      "Investigation of ban evasion",
				SubforumID:         1,
				IncidentID:         "ban_evasion_123",
			},
		}

		// Assertions
		assert.NotNil(t, input)
		assert.Equal(t, "target-pseudonym-123", input.Body.RequestedPseudonym)
		assert.Equal(t, "Investigation of ban evasion", input.Body.Justification)
		assert.Equal(t, 1, input.Body.SubforumID)
		assert.Equal(t, "ban_evasion_123", input.Body.IncidentID)
	})

	t.Run("IdentityCorrelationInput", func(t *testing.T) {
		input := &apimodels.IdentityCorrelationInput{
			Body: apimodels.IdentityCorrelationInputBody{
				RequestedPseudonym: "target-pseudonym-123",
				Justification:      "Investigation of harassment",
				LegalBasis:         "Platform Terms of Service",
				IncidentID:         "harassment_case_123",
				Scope:              "platform_wide",
			},
		}

		// Assertions
		assert.NotNil(t, input)
		assert.Equal(t, "target-pseudonym-123", input.Body.RequestedPseudonym)
		assert.Equal(t, "Investigation of harassment", input.Body.Justification)
		assert.Equal(t, "Platform Terms of Service", input.Body.LegalBasis)
		assert.Equal(t, "harassment_case_123", input.Body.IncidentID)
		assert.Equal(t, "platform_wide", input.Body.Scope)
	})

	t.Run("CorrelationResult", func(t *testing.T) {
		postsInSubforum := 5
		commentsInSubforum := 10

		result := apimodels.CorrelationResult{
			PseudonymID:        "test-pseudonym-123",
			DisplayName:        "TestUser",
			CreatedAt:          time.Now().Format(time.RFC3339),
			PostsInSubforum:    &postsInSubforum,
			CommentsInSubforum: &commentsInSubforum,
		}

		// Assertions
		assert.NotNil(t, result)
		assert.Equal(t, "test-pseudonym-123", result.PseudonymID)
		assert.Equal(t, "TestUser", result.DisplayName)
		assert.Equal(t, 5, *result.PostsInSubforum)
		assert.Equal(t, 10, *result.CommentsInSubforum)
	})
}

// TestCorrelationResponseCreation tests the response creation functions
func TestCorrelationResponseCreation(t *testing.T) {
	t.Run("NewFingerprintCorrelationResponse", func(t *testing.T) {
		correlationID := "correlation-123"
		auditID := "audit-456"
		results := []apimodels.CorrelationResult{
			{
				PseudonymID: "pseudonym-1",
				DisplayName: "User1",
			},
			{
				PseudonymID: "pseudonym-2",
				DisplayName: "User2",
			},
		}

		response := apimodels.NewFingerprintCorrelationResponse(correlationID, results, auditID)

		// Assertions
		assert.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, "fingerprint", response.Body.CorrelationType)
		assert.Equal(t, "subforum_specific", response.Body.Scope)
		assert.Equal(t, "completed", response.Body.Status)
		assert.Equal(t, correlationID, response.Body.CorrelationID)
		assert.Equal(t, auditID, response.Body.AuditID)
		assert.Len(t, response.Body.Results, 2)
	})

	t.Run("NewIdentityCorrelationResponse", func(t *testing.T) {
		correlationID := "correlation-123"
		auditID := "audit-456"
		results := []apimodels.CorrelationResult{
			{
				PseudonymID: "pseudonym-1",
				DisplayName: "User1",
			},
		}

		response := apimodels.NewIdentityCorrelationResponse(correlationID, results, auditID)

		// Assertions
		assert.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, "identity", response.Body.CorrelationType)
		assert.Equal(t, "platform_wide", response.Body.Scope)
		assert.Equal(t, "completed", response.Body.Status)
		assert.Equal(t, correlationID, response.Body.CorrelationID)
		assert.Equal(t, auditID, response.Body.AuditID)
		assert.Len(t, response.Body.Results, 1)
	})
}

// TestNewCorrelationHandler tests the main constructor function
func TestNewCorrelationHandler(t *testing.T) {
	t.Run("NewCorrelationHandlerSuccess", func(t *testing.T) {
		// Create mock dependencies
		var mockDB bob.Executor = nil
		mockIBESystem := &ibe.IBESystem{}
		mockSecurePseudonymDAO := &mocks.MockSecurePseudonymDAO{}
		mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
		mockPostDAO := &mocks.MockPostDAO{}
		mockCommentDAO := &mocks.MockCommentDAO{}
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockCorrelationAuditDAO := &mocks.MockCorrelationAuditDAO{}

		// Create handler with dependencies
		handler := NewCorrelationHandler(
			mockDB,
			mockIBESystem,
			mockSecurePseudonymDAO,
			mockIdentityMappingDAO,
			mockPostDAO,
			mockCommentDAO,
			mockSubforumDAO,
			mockCorrelationAuditDAO,
		)

		// Verify handler is created
		assert.NotNil(t, handler)
	})
}
