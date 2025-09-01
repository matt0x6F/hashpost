package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
)

// TestNewCorrelationHandler tests the correlation handler constructor
func TestNewCorrelationHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockIdentityMappingDAO := dao.NewMockIdentityMappingDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockCorrelationAuditDAO := dao.NewMockCorrelationAuditDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	// Create a mock bob.Executor
	var mockDB bob.Executor = nil // We'll use nil for testing

	// Create IBE system
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	handler := handlers.NewCorrelationHandler(
		mockDB,
		ibeSystem,
		mockPseudonymDAO,
		mockIdentityMappingDAO,
		mockPostDAO,
		mockCommentDAO,
		mockSubforumDAO,
		mockCorrelationAuditDAO,
		mockPermissionDAO,
	)

	assert.NotNil(t, handler)
}

// TestCorrelationHandler_RequestFingerprintCorrelation tests the RequestFingerprintCorrelation method
func TestCorrelationHandler_RequestFingerprintCorrelation(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockIdentityMappingDAO := dao.NewMockIdentityMappingDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockCorrelationAuditDAO := dao.NewMockCorrelationAuditDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	var mockDB bob.Executor = nil
	mockIBESystem := ibe.NewMockIBESystemInterface(ctrl)

	handler := handlers.NewCorrelationHandler(
		mockDB,
		mockIBESystem,
		mockPseudonymDAO,
		mockIdentityMappingDAO,
		mockPostDAO,
		mockCommentDAO,
		mockSubforumDAO,
		mockCorrelationAuditDAO,
		mockPermissionDAO,
	)

	tests := []struct {
		name           string
		input          *models.FingerprintCorrelationInput
		setupMocks     func()
		expectedError  bool
		expectedResult *models.FingerprintCorrelationResponse
	}{
		{
			name: "Success - Valid fingerprint correlation request",
			input: &models.FingerprintCorrelationInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				Body: models.FingerprintCorrelationInputBody{
					RequestedPseudonym:   "pseudonym-123",
					RequestedFingerprint: "fingerprint-456",
					Justification:        "Investigation of ban evasion",
					SubforumID:           1,
					IncidentID:           "incident-789",
				},
			},
			setupMocks: func() {
				// Mock permission check
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "correlate_fingerprints", nil).
					Return(true, nil)

				// Mock identity mapping retrieval
				mockIdentityMappingDAO.EXPECT().
					GetIdentityMappingByPseudonymID(gomock.Any(), "pseudonym-123").
					Return(&dbmodels.IdentityMapping{
						PseudonymID:           "pseudonym-123",
						EncryptedRealIdentity: []byte("encrypted-fingerprint:pseudonym-123"),
					}, nil)

				// Mock pseudonym retrieval (first call - initial lookup)
				mockPseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "pseudonym-123").
					Return(&dbmodels.Pseudonym{
						PseudonymID: "pseudonym-123",
						DisplayName: "test_user",
					}, nil)

				// Mock pseudonym retrieval (second call - for related mappings)
				mockPseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "pseudonym-123").
					Return(&dbmodels.Pseudonym{
						PseudonymID: "pseudonym-123",
						DisplayName: "test_user",
					}, nil)

				// Mock IBE system role key generation
				mockIBESystem.EXPECT().
					GenerateRoleKey("moderator", "correlation", gomock.Any()).
					Return([]byte("mock-admin-key"))

				// Mock IBE system decryption
				mockIBESystem.EXPECT().
					DecryptIdentity(gomock.Any(), []byte("mock-admin-key")).
					Return("fingerprint-456:pseudonym-123", "decrypted-pseudonym", nil)

				// Mock related identity mappings retrieval
				mockIdentityMappingDAO.EXPECT().
					GetIdentityMappingsByFingerprint(gomock.Any(), "fingerprint-456").
					Return(dbmodels.IdentityMappingSlice{
						&dbmodels.IdentityMapping{
							PseudonymID:           "pseudonym-123",
							EncryptedRealIdentity: []byte("encrypted-fingerprint:pseudonym-123"),
						},
					}, nil)

				// Mock post count in subforum
				mockPostDAO.EXPECT().
					CountPostsByPseudonymInSubforum(gomock.Any(), "pseudonym-123", int32(1)).
					Return(int64(5), nil)

				// Mock comment count in subforum
				mockCommentDAO.EXPECT().
					CountCommentsByPseudonymInSubforum(gomock.Any(), "pseudonym-123", int32(1)).
					Return(int64(10), nil)

				// Mock correlation audit creation
				mockCorrelationAuditDAO.EXPECT().
					CreateCorrelationAudit(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: false,
			expectedResult: &models.FingerprintCorrelationResponse{
				Status: 200,
				Body: models.FingerprintCorrelationResponseBody{
					CorrelationType: "fingerprint",
					Scope:           "subforum_specific",
					Status:          "pending",
				},
			},
		},
		{
			name: "Error - Authentication required",
			input: &models.FingerprintCorrelationInput{
				Body: models.FingerprintCorrelationInputBody{
					RequestedPseudonym:   "pseudonym-123",
					RequestedFingerprint: "fingerprint-456",
					Justification:        "Investigation of ban evasion",
					SubforumID:           1,
					IncidentID:           "incident-789",
				},
			},
			setupMocks:     func() {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Insufficient permissions",
			input: &models.FingerprintCorrelationInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				Body: models.FingerprintCorrelationInputBody{
					RequestedPseudonym:   "pseudonym-123",
					RequestedFingerprint: "fingerprint-456",
					Justification:        "Investigation of ban evasion",
					SubforumID:           1,
					IncidentID:           "incident-789",
				},
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "correlate_fingerprints", nil).
					Return(false, nil)
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Pseudonym not found",
			input: &models.FingerprintCorrelationInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				Body: models.FingerprintCorrelationInputBody{
					RequestedPseudonym:   "pseudonym-123",
					RequestedFingerprint: "fingerprint-456",
					Justification:        "Investigation of ban evasion",
					SubforumID:           1,
					IncidentID:           "incident-789",
				},
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "correlate_fingerprints", nil).
					Return(true, nil)

				mockPseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "pseudonym-123").
					Return(nil, errors.New("pseudonym not found"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test case
			tt.setupMocks()

			// Execute the method
			result, err := handler.RequestFingerprintCorrelation(context.Background(), tt.input)

			// Assertions
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.Equal(t, tt.expectedResult.Body.CorrelationType, result.Body.CorrelationType)
				assert.Equal(t, tt.expectedResult.Body.Scope, result.Body.Scope)
			}
		})
	}
}

// TestCorrelationHandler_RequestIdentityCorrelation tests the RequestIdentityCorrelation method
func TestCorrelationHandler_RequestIdentityCorrelation(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockIdentityMappingDAO := dao.NewMockIdentityMappingDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockCorrelationAuditDAO := dao.NewMockCorrelationAuditDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	var mockDB bob.Executor = nil
	mockIBESystem := ibe.NewMockIBESystemInterface(ctrl)

	handler := handlers.NewCorrelationHandler(
		mockDB,
		mockIBESystem,
		mockPseudonymDAO,
		mockIdentityMappingDAO,
		mockPostDAO,
		mockCommentDAO,
		mockSubforumDAO,
		mockCorrelationAuditDAO,
		mockPermissionDAO,
	)

	tests := []struct {
		name           string
		input          *models.IdentityCorrelationInput
		setupMocks     func()
		expectedError  bool
		expectedResult *models.IdentityCorrelationResponse
	}{
		{
			name: "Success - Valid identity correlation request",
			input: &models.IdentityCorrelationInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				Body: models.IdentityCorrelationInputBody{
					RequestedPseudonym:   "pseudonym-123",
					RequestedFingerprint: "fingerprint-456",
					Justification:        "Investigation of harassment",
					LegalBasis:           "Platform Terms of Service",
					IncidentID:           "incident-789",
					Scope:                "platform_wide",
				},
			},
			setupMocks: func() {
				// Mock permission check
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "correlate_identities", nil).
					Return(true, nil)

				// Mock identity mapping retrieval
				mockIdentityMappingDAO.EXPECT().
					GetIdentityMappingByPseudonymID(gomock.Any(), "pseudonym-123").
					Return(&dbmodels.IdentityMapping{
						PseudonymID:           "pseudonym-123",
						EncryptedRealIdentity: []byte("encrypted-fingerprint:pseudonym-123"),
					}, nil)

				// Mock pseudonym retrieval (first call - initial lookup)
				mockPseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "pseudonym-123").
					Return(&dbmodels.Pseudonym{
						PseudonymID: "pseudonym-123",
						DisplayName: "test_user",
					}, nil)

				// Mock pseudonym retrieval (second call - for related mappings)
				mockPseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "pseudonym-123").
					Return(&dbmodels.Pseudonym{
						PseudonymID: "pseudonym-123",
						DisplayName: "test_user",
					}, nil)

				// Mock IBE system role key generation
				mockIBESystem.EXPECT().
					GenerateRoleKey("site_admin", "full_correlation", gomock.Any()).
					Return([]byte("mock-admin-key"))

				// Mock IBE system decryption
				mockIBESystem.EXPECT().
					DecryptIdentity(gomock.Any(), []byte("mock-admin-key")).
					Return("fingerprint-456:pseudonym-123", "decrypted-pseudonym", nil)

				// Mock related identity mappings retrieval
				mockIdentityMappingDAO.EXPECT().
					GetIdentityMappingsByFingerprint(gomock.Any(), "fingerprint-456").
					Return(dbmodels.IdentityMappingSlice{
						&dbmodels.IdentityMapping{
							PseudonymID:           "pseudonym-123",
							EncryptedRealIdentity: []byte("encrypted-fingerprint:pseudonym-123"),
						},
					}, nil)

				// Mock total post count
				mockPostDAO.EXPECT().
					CountPostsByPseudonym(gomock.Any(), "pseudonym-123").
					Return(int64(15), nil)

				// Mock total comment count
				mockCommentDAO.EXPECT().
					CountCommentsByPseudonym(gomock.Any(), "pseudonym-123").
					Return(int64(25), nil)

				// Mock subforums by posts
				mockPostDAO.EXPECT().
					GetSubforumsByPseudonym(gomock.Any(), "pseudonym-123").
					Return([]int32{1, 2}, nil)

				// Mock subforums by comments
				mockCommentDAO.EXPECT().
					GetSubforumsByPseudonymComments(gomock.Any(), "pseudonym-123").
					Return([]int32{1, 3}, nil)

				// Mock subforum details
				mockSubforumDAO.EXPECT().
					GetSubforumByID(gomock.Any(), int32(1)).
					Return(&dbmodels.Subforum{
						SubforumID:  1,
						Name:        "test-subforum-1",
						DisplayName: "Test Subforum 1",
					}, nil)
				mockSubforumDAO.EXPECT().
					GetSubforumByID(gomock.Any(), int32(2)).
					Return(&dbmodels.Subforum{
						SubforumID:  2,
						Name:        "test-subforum-2",
						DisplayName: "Test Subforum 2",
					}, nil)
				mockSubforumDAO.EXPECT().
					GetSubforumByID(gomock.Any(), int32(3)).
					Return(&dbmodels.Subforum{
						SubforumID:  3,
						Name:        "test-subforum-3",
						DisplayName: "Test Subforum 3",
					}, nil)

				// Mock correlation audit creation
				mockCorrelationAuditDAO.EXPECT().
					CreateCorrelationAudit(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: false,
			expectedResult: &models.IdentityCorrelationResponse{
				Status: 200,
				Body: models.IdentityCorrelationResponseBody{
					CorrelationType: "identity",
					Scope:           "platform_wide",
					Status:          "pending",
				},
			},
		},
		{
			name: "Error - Authentication required",
			input: &models.IdentityCorrelationInput{
				Body: models.IdentityCorrelationInputBody{
					RequestedPseudonym:   "pseudonym-123",
					RequestedFingerprint: "fingerprint-456",
					Justification:        "Investigation of harassment",
					LegalBasis:           "Platform Terms of Service",
					IncidentID:           "incident-789",
					Scope:                "platform_wide",
				},
			},
			setupMocks:     func() {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Insufficient permissions",
			input: &models.IdentityCorrelationInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				Body: models.IdentityCorrelationInputBody{
					RequestedPseudonym:   "pseudonym-123",
					RequestedFingerprint: "fingerprint-456",
					Justification:        "Investigation of harassment",
					LegalBasis:           "Platform Terms of Service",
					IncidentID:           "incident-789",
					Scope:                "platform_wide",
				},
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "correlate_identities", nil).
					Return(false, nil)
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test case
			tt.setupMocks()

			// Execute the method
			result, err := handler.RequestIdentityCorrelation(context.Background(), tt.input)

			// Assertions
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.Equal(t, tt.expectedResult.Body.CorrelationType, result.Body.CorrelationType)
				assert.Equal(t, tt.expectedResult.Body.Scope, result.Body.Scope)
			}
		})
	}
}

// TestCorrelationHandler_GetCorrelationHistory tests the GetCorrelationHistory method
func TestCorrelationHandler_GetCorrelationHistory(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockIdentityMappingDAO := dao.NewMockIdentityMappingDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockCorrelationAuditDAO := dao.NewMockCorrelationAuditDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	var mockDB bob.Executor = nil
	mockIBESystem := ibe.NewMockIBESystemInterface(ctrl)

	handler := handlers.NewCorrelationHandler(
		mockDB,
		mockIBESystem,
		mockPseudonymDAO,
		mockIdentityMappingDAO,
		mockPostDAO,
		mockCommentDAO,
		mockSubforumDAO,
		mockCorrelationAuditDAO,
		mockPermissionDAO,
	)

	tests := []struct {
		name           string
		input          *models.CorrelationHistoryInput
		setupMocks     func()
		expectedError  bool
		expectedResult *models.CorrelationHistoryResponse
	}{
		{
			name: "Success - Valid correlation history retrieval",
			input: &models.CorrelationHistoryInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				CorrelationType: "fingerprint",
				Page:            1,
				Limit:           25,
			},
			setupMocks: func() {
				// Mock permission check
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "view_correlation_history", nil).
					Return(true, nil)

				// Mock correlation history retrieval
				expectedHistory := dbmodels.CorrelationAuditSlice{
					&dbmodels.CorrelationAudit{
						AuditID:         uuid.Must(uuid.NewV4()),
						CorrelationType: "fingerprint",
						Timestamp:       sql.Null[time.Time]{Valid: true, V: time.Now()},
					},
				}
				mockCorrelationAuditDAO.EXPECT().
					GetCorrelationHistory(gomock.Any(), "fingerprint", 1, 25).
					Return(expectedHistory, nil)
			},
			expectedError: false,
			expectedResult: &models.CorrelationHistoryResponse{
				Status: 200,
				Body: models.CorrelationHistoryResponseBody{
					Correlations: []models.Correlation{
						{
							CorrelationID:   "correlation-123",
							CorrelationType: "fingerprint",
							Status:          "completed",
						},
					},
				},
			},
		},
		{
			name: "Error - Authentication required",
			input: &models.CorrelationHistoryInput{
				CorrelationType: "fingerprint",
				Page:            1,
				Limit:           25,
			},
			setupMocks:     func() {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Insufficient permissions",
			input: &models.CorrelationHistoryInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				CorrelationType: "fingerprint",
				Page:            1,
				Limit:           25,
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "view_correlation_history", nil).
					Return(false, nil)
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Database failure",
			input: &models.CorrelationHistoryInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(456, "user-pseudonym-789"),
				},
				CorrelationType: "fingerprint",
				Page:            1,
				Limit:           25,
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(456), "user-pseudonym-789", "view_correlation_history", nil).
					Return(true, nil)

				mockCorrelationAuditDAO.EXPECT().
					GetCorrelationHistory(gomock.Any(), "fingerprint", 1, 25).
					Return(nil, errors.New("database connection failed"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test case
			tt.setupMocks()

			// Execute the method
			result, err := handler.GetCorrelationHistory(context.Background(), tt.input)

			// Assertions
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.Len(t, result.Body.Correlations, len(tt.expectedResult.Body.Correlations))
			}
		})
	}
}
