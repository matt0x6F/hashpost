package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// generateJWTWithCapabilities generates a JWT token with specific capabilities
func generateJWTWithCapabilities(userID int64, activePseudonymID string, capabilities []string) string {
	claims := &middleware.JWTClaims{
		UserID:            userID,
		Email:             "test@example.com",
		Roles:             []string{"user"},
		Capabilities:      capabilities,
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
		panic(fmt.Sprintf("Failed to generate test JWT token: %v", err))
	}

	return tokenString
}

func TestSubforumHandler_GetSubforums(t *testing.T) {
	tests := []struct {
		name          string
		input         *models.SubforumListInput
		userCtx       *middleware.UserContext
		wantErr       bool
		expectedCount int
	}{
		{
			name: "GetPublicSubforums",
			input: &models.SubforumListInput{
				Page:  1,
				Limit: 10,
				Sort:  "name",
			},
			userCtx:       fixtures.CreateTestUserContext(),
			wantErr:       false,
			expectedCount: 2, // Including private subforum with access
		},
		{
			name: "GetSubforumsAnonymous",
			input: &models.SubforumListInput{
				Page:  1,
				Limit: 10,
				Sort:  "name",
			},
			userCtx:       nil,
			wantErr:       false,
			expectedCount: 1,
		},
		{
			name: "GetSubforumsWithPrivateAccess",
			input: &models.SubforumListInput{
				Page:  1,
				Limit: 10,
				Sort:  "name",
			},
			userCtx:       fixtures.CreateTestUserContext(),
			wantErr:       false,
			expectedCount: 2, // Including private subforum with access
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			mockSubforumDAO.InjectAllSubforums([]*dbmodels.Subforum{fixtures.CreateTestSubforum(), fixtures.CreateTestPrivateSubforum()})
			mockSubforumDAO.SetDefaultBehavior()

			mockSubforumSubscriptionDAO.InjectCount("subforum_1", 100)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			// Set up mock expectations for CanAccessPrivateSubforumWithActivePseudonym
			mockPermissionDAO.On("CanAccessPrivateSubforumWithActivePseudonym", mock.Anything, int64(1), int32(1), "test-pseudonym-id").Return(true, nil)
			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			response, err := handler.GetSubforums(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Body.Subforums, tt.expectedCount)
			}
		})
	}
}

func TestSubforumHandler_GetSubforumDetails(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *models.SubforumSubscriptionInput
		userCtx        *middleware.UserContext
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "GetPublicSubforumDetails",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "test-subforum",
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetPrivateSubforumDetailsWithAccess",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "private-test-subforum",
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetPrivateSubforumDetailsWithoutAccess",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "private-test-subforum",
			},
			userCtx:        nil,
			wantErr:        true,
			expectedStatus: 403,
		},
		{
			name: "GetNonExistentSubforum",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "non-existent",
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        true,
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			switch tt.input.SubforumName {
			case "test-subforum":
				mockSubforumDAO.InjectSubforumByCommunityTypeAndName("t", "test-subforum", fixtures.CreateTestSubforum())
			case "private-test-subforum":
				mockSubforumDAO.InjectSubforumByCommunityTypeAndName("t", "private-test-subforum", fixtures.CreateTestPrivateSubforum())
			}

			mockSubforumDAO.SetDefaultBehavior()

			mockSubforumSubscriptionDAO.InjectSubscribedStatus("moderator-pseudonym-123", 1, true)
			mockSubforumSubscriptionDAO.InjectFavoriteStatus("moderator-pseudonym-123", 1, false)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			// Set up mock expectations for CanAccessPrivateSubforumWithActivePseudonym
			mockPermissionDAO.On("CanAccessPrivateSubforumWithActivePseudonym", mock.Anything, int64(1), int32(1), "test-pseudonym-id").Return(true, nil)
			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Set up mock expectations for GetModeratorsForSubforum
			mockRoleKeyDAO.On("GetModeratorsForSubforum", mock.Anything, int32(1)).Return([]*dbmodels.RoleKey{}, nil)

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			input := &struct {
				middleware.AuthInput
				models.SubforumSubscriptionInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: func() string {
						if tt.userCtx != nil {
							return "Bearer " + generateJWTWithCapabilities(tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, tt.userCtx.Capabilities)
						}
						return ""
					}(),
				},
				SubforumSubscriptionInput: *tt.input,
			}
			response, err := handler.GetSubforumDetails(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.input.SubforumName, response.Body.Subforum.Name)
			}
		})
	}
}

func TestSubforumHandler_SubscribeToSubforum(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *models.SubforumSubscriptionInput
		userCtx        *middleware.UserContext
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "SubscribeToSubforum",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "test-subforum",
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "SubscribeToNonExistentSubforum",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "non-existent",
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        true,
			expectedStatus: 404,
		},
		{
			name: "SubscribeWithoutAuthentication",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "test-subforum",
			},
			userCtx:        nil,
			wantErr:        true,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			if tt.input.SubforumName == "test-subforum" {
				mockSubforumDAO.InjectSubforumByCommunityTypeAndName("t", "test-subforum", fixtures.CreateTestSubforum())
			}
			mockSubforumDAO.SetDefaultBehavior()

			mockSubforumSubscriptionDAO.InjectSubscribedStatus("moderator-pseudonym-123", 1, false) // Not subscribed initially
			mockSubforumSubscriptionDAO.InjectCount("subforum_1", 100)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			input := &struct {
				middleware.AuthInput
				models.SubforumSubscriptionInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: func() string {
						if tt.userCtx != nil {
							return "Bearer " + generateJWTWithCapabilities(tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, tt.userCtx.Capabilities)
						}
						return ""
					}(),
				},
				SubforumSubscriptionInput: *tt.input,
			}
			response, err := handler.SubscribeToSubforum(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Body.IsSubscribed)
			}
		})
	}
}

func TestSubforumHandler_UnsubscribeFromSubforum(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *models.SubforumSubscriptionInput
		userCtx        *middleware.UserContext
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "UnsubscribeFromSubforum",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "test-subforum",
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "UnsubscribeFromNonExistentSubforum",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "non-existent",
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        true,
			expectedStatus: 404,
		},
		{
			name: "UnsubscribeWithoutAuthentication",
			input: &models.SubforumSubscriptionInput{
				CommunityType: "t",
				SubforumName:  "test-subforum",
			},
			userCtx:        nil,
			wantErr:        true,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			if tt.input.SubforumName == "test-subforum" {
				mockSubforumDAO.InjectSubforumByCommunityTypeAndName("t", "test-subforum", fixtures.CreateTestSubforum())
			}
			mockSubforumDAO.SetDefaultBehavior()

			// Create a subscription record for the mock
			testSubscription := &dbmodels.SubforumSubscription{
				SubscriptionID: 123,
				PseudonymID:    "test-pseudonym-id",
				SubforumID:     1,
				IsFavorite:     sql.Null[bool]{V: false, Valid: true},
				SubscribedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
			}
			mockSubforumSubscriptionDAO.InjectSubscription(testSubscription)
			mockSubforumSubscriptionDAO.InjectSubscribedStatus("test-pseudonym-id", 1, true) // Subscribed initially
			mockSubforumSubscriptionDAO.InjectCount("subforum_1", 100)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			input := &struct {
				middleware.AuthInput
				models.SubforumSubscriptionInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: func() string {
						if tt.userCtx != nil {
							return "Bearer " + generateJWTWithCapabilities(tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, tt.userCtx.Capabilities)
						}
						return ""
					}(),
				},
				SubforumSubscriptionInput: *tt.input,
			}
			response, err := handler.UnsubscribeFromSubforum(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.False(t, response.Body.IsSubscribed)
			}
		})
	}
}

func TestSubforumHandler_CreateSubforum(t *testing.T) {
	tests := []struct {
		name               string
		input              *models.SubforumCreateInput
		userCtx            *middleware.UserContext
		wantErr            bool
		expectedStatus     int
		expectedGovernance string
	}{
		{
			name: "CreateTopicalCommunity",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "programming",
					Name:          "Programming",
					Description:   "Programming discussions",
					CommunityType: "t",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx:            fixtures.CreateTestModeratorContext(),
			wantErr:            false,
			expectedStatus:     200,
			expectedGovernance: "democratic",
		},
		{
			name: "CreateGeographicCommunity",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "seattle",
					Name:          "Seattle",
					Description:   "Seattle area discussions",
					CommunityType: "g",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx:            fixtures.CreateTestModeratorContext(),
			wantErr:            false,
			expectedStatus:     200,
			expectedGovernance: "democratic",
		},
		{
			name: "CreateBrandedCommunity",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "apple",
					Name:          "Apple",
					Description:   "Apple company discussions",
					CommunityType: "b",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx:            fixtures.CreateTestModeratorContext(),
			wantErr:            false,
			expectedStatus:     200,
			expectedGovernance: "owned",
		},
		{
			name: "CreateCreatorCommunity",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "creator",
					Name:          "Creator Community",
					Description:   "Creator community discussions",
					CommunityType: "c",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx:            fixtures.CreateTestModeratorContext(),
			wantErr:            false,
			expectedStatus:     200,
			expectedGovernance: "owned",
		},
		{
			name: "CreateSubforumWithInvalidCommunityType",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "invalid",
					Name:          "Invalid",
					Description:   "Invalid community type",
					CommunityType: "x", // Invalid type
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        true,
			expectedStatus: 400,
		},
		{
			name: "CreateSubforumWithoutPermission",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "new-subforum",
					Name:          "New Subforum",
					Description:   "A new test subforum",
					CommunityType: "t",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx: &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: "moderator-pseudonym-123",
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content"}, // Missing create_subforum capability
			},
			wantErr:        true,
			expectedStatus: 403,
		},
		{
			name: "CreateSubforumWithoutAuthentication",
			input: &models.SubforumCreateInput{
				Body: models.SubforumCreateBody{
					Slug:          "new-subforum",
					Name:          "New Subforum",
					Description:   "A new test subforum",
					CommunityType: "t",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx:        nil,
			wantErr:        true,
			expectedStatus: 401,
		},
		{
			name: "CreateSubforumWithoutRequiredFields",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "", // Missing slug
					Name:          "", // Missing name
					Description:   "", // Missing description
					CommunityType: "t",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        true,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			mockSubforumDAO.SetDefaultBehavior()
			mockSubforumSubscriptionDAO.SetDefaultBehavior()
			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			response, err := handler.CreateSubforum(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.input.Body.Name, response.Body.Subforum.DisplayName)
				assert.Equal(t, tt.input.Body.CommunityType, response.Body.Subforum.CommunityType)
				assert.Equal(t, tt.expectedGovernance, response.Body.Subforum.GovernanceStyle)
				assert.Equal(t, "test-pseudonym-id", response.Body.Subforum.OwnerPseudonymID)
			}
		})
	}
}

// TestGovernanceStyleEnforcement tests that governance styles are correctly enforced based on community type
func TestGovernanceStyleEnforcement(t *testing.T) {
	tests := []struct {
		name               string
		communityType      string
		expectedGovernance string
	}{
		{"TopicalCommunity", "t", "democratic"},
		{"GeographicCommunity", "g", "democratic"},
		{"BrandedCommunity", "b", "owned"},
		{"CreatorCommunity", "c", "owned"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			mockSubforumDAO.SetDefaultBehavior()
			mockSubforumSubscriptionDAO.SetDefaultBehavior()
			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create test input
			input := &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "test-" + tt.communityType,
					Name:          "Test " + tt.communityType,
					Description:   "Test community",
					CommunityType: tt.communityType,
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			}

			// Create context
			ctx := createTestContextWithUser(fixtures.CreateTestModeratorContext())

			// Call method
			response, err := handler.CreateSubforum(ctx, input)

			// Assertions
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, tt.expectedGovernance, response.Body.Subforum.GovernanceStyle,
				"Community type %s should have governance style %s", tt.communityType, tt.expectedGovernance)
		})
	}
}

// TestCommunityTypeValidation tests that invalid community types are rejected
func TestCommunityTypeValidation(t *testing.T) {
	invalidTypes := []string{"x", "h", "z", "invalid", ""}

	for _, invalidType := range invalidTypes {
		t.Run("InvalidType_"+invalidType, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			mockSubforumDAO.SetDefaultBehavior()
			mockSubforumSubscriptionDAO.SetDefaultBehavior()
			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create test input with invalid community type
			input := &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				Body: models.SubforumCreateBody{
					Slug:          "test-invalid",
					Name:          "Test Invalid",
					Description:   "Test invalid community type",
					CommunityType: invalidType,
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			}

			// Create context
			ctx := createTestContextWithUser(fixtures.CreateTestModeratorContext())

			// Call method
			response, err := handler.CreateSubforum(ctx, input)

			// Assertions
			assert.Error(t, err)
			assert.Nil(t, response)
		})
	}
}

// TestOwnerAssignment tests that the creating user's active pseudonym becomes the owner
func TestOwnerAssignment(t *testing.T) {
	tests := []struct {
		name              string
		activePseudonymID string
		expectedOwner     string
	}{
		{"OwnerAssignment", "owner-pseudonym-456", "owner-pseudonym-456"},
		{"DifferentPseudonym", "different-pseudonym-789", "different-pseudonym-789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			mockSubforumDAO.SetDefaultBehavior()
			mockSubforumSubscriptionDAO.SetDefaultBehavior()
			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create user context with specific active pseudonym
			userCtx := &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: tt.activePseudonymID,
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "create_subforum"},
			}

			// Create test input
			input := &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, tt.activePseudonymID),
				Body: models.SubforumCreateBody{
					Slug:          "test-owner",
					Name:          "Test Owner",
					Description:   "Test owner assignment",
					CommunityType: "t",
					IsNSFW:        false,
					IsPrivate:     false,
					IsRestricted:  false,
				},
			}

			// Create context
			ctx := createTestContextWithUser(userCtx)

			// Call method
			response, err := handler.CreateSubforum(ctx, input)

			// Assertions
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, tt.expectedOwner, response.Body.Subforum.OwnerPseudonymID)
		})
	}
}

func TestSubforumHandler_GetPseudonymSubscriptions(t *testing.T) {
	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			models.PseudonymSubscriptionsInput
		}
		userCtx        *middleware.UserContext
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "GetOwnPseudonymSubscriptions",
			input: &struct {
				middleware.AuthInput
				models.PseudonymSubscriptionsInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "moderator-pseudonym-123"),
				},
				PseudonymSubscriptionsInput: models.PseudonymSubscriptionsInput{
					PseudonymID: "moderator-pseudonym-123",
				},
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetOtherPseudonymSubscriptions",
			input: &struct {
				middleware.AuthInput
				models.PseudonymSubscriptionsInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "moderator-pseudonym-123"),
				},
				PseudonymSubscriptionsInput: models.PseudonymSubscriptionsInput{
					PseudonymID: "other-pseudonym-456",
				},
			},
			userCtx:        fixtures.CreateTestUserContext(),
			wantErr:        true,
			expectedStatus: 403,
		},
		{
			name: "GetSubscriptionsWithoutAuthentication",
			input: &struct {
				middleware.AuthInput
				models.PseudonymSubscriptionsInput
			}{
				PseudonymSubscriptionsInput: models.PseudonymSubscriptionsInput{
					PseudonymID: "moderator-pseudonym-123",
				},
			},
			userCtx:        nil,
			wantErr:        true,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock data
			mockSubforumDAO.InjectSubforum(fixtures.CreateTestSubforum())
			mockSubforumDAO.SetDefaultBehavior()
			mockSubforumSubscriptionDAO.InjectSubscriptionsByPseudonym("moderator-pseudonym-123", []*dbmodels.SubforumSubscription{fixtures.CreateTestSubforumSubscription()})
			mockSubforumSubscriptionDAO.SetDefaultBehavior()
			mockPermissionDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Set up mock identity mappings for the user
			mockIdentityMappingDAO.On("GetIdentityMappingsByUserID", mock.Anything, int64(1)).Return(
				[]*dbmodels.IdentityMapping{
					{
						PseudonymID: "moderator-pseudonym-123",
						UserID:      1,
						IsActive:    sql.Null[bool]{V: true, Valid: true},
					},
				}, nil)

			// Create mock post DAO
			mockPostDAO := mocks.NewMockPostDAO()

			// Set up default behavior for post DAO
			mockPostDAO.SetDefaultBehavior()

			// Create mock pseudonym DAO
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()

			// Create mock role key DAO
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Create handler with mocks
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			response, err := handler.GetPseudonymSubscriptions(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.Body.Subforums)
			}
		})
	}
}

// TestNewSubforumHandler tests the main constructor function
func TestNewSubforumHandler(t *testing.T) {
	t.Run("NewSubforumHandlerSuccess", func(t *testing.T) {
		// Create mock dependencies
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockSubforumSubscriptionDAO := &mocks.MockSubforumSubscriptionDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
		mockPostDAO := &mocks.MockPostDAO{}
		// Create mock pseudonym DAO
		mockPseudonymDAO := mocks.NewMockPseudonymDAO()

		// Create mock role key DAO
		mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

		// Create handler with dependencies
		handler := NewSubforumHandler(
			nil, // nil db for testing
			mockSubforumDAO,
			mockSubforumSubscriptionDAO,
			mockPermissionDAO,
			mockIdentityMappingDAO,
			mockPseudonymDAO,
			mockPostDAO,
			mockRoleKeyDAO,
		)

		// Verify handler is created
		assert.NotNil(t, handler)
	})
}

func TestSubforumHandler_GetSubforumSettings(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			models.SubforumSettingsGetInput
		}
		userCtx        *middleware.UserContext
		mockSubforum   *dbmodels.Subforum
		mockPermission bool
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "GetSettingsSuccess",
			input: &struct {
				middleware.AuthInput
				models.SubforumSettingsGetInput
			}{
				AuthInput: middleware.AuthInput{},
				SubforumSettingsGetInput: models.SubforumSettingsGetInput{
					Type: "b",
					Name: "hashpost",
				},
			},
			userCtx: fixtures.CreateTestUserContext(),
			mockSubforum: &dbmodels.Subforum{
				SubforumID:             1,
				Name:                   "hashpost",
				AllowImages:            sql.Null[bool]{V: true, Valid: true},
				AllowVideos:            sql.Null[bool]{V: false, Valid: true},
				AllowPolls:             sql.Null[bool]{V: true, Valid: true},
				RequireFlair:           sql.Null[bool]{V: false, Valid: true},
				MinimumAccountAgeHours: sql.Null[int32]{V: 24, Valid: true},
				MinimumKarmaRequired:   sql.Null[int32]{V: 10, Valid: true},
				IsPrivate:              sql.Null[bool]{V: false, Valid: true},
				IsRestricted:           sql.Null[bool]{V: false, Valid: true},
				IsNSFW:                 sql.Null[bool]{V: false, Valid: true},
				SidebarText:            sql.Null[string]{V: "Welcome to HashPost!", Valid: true},
				UpdatedAt:              sql.Null[time.Time]{V: time.Now(), Valid: true},
			},
			mockPermission: true,
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetSettingsUnauthorized",
			input: &struct {
				middleware.AuthInput
				models.SubforumSettingsGetInput
			}{
				AuthInput: middleware.AuthInput{},
				SubforumSettingsGetInput: models.SubforumSettingsGetInput{
					Type: "b",
					Name: "hashpost",
				},
			},
			userCtx:        nil,
			mockSubforum:   nil,
			mockPermission: false,
			wantErr:        true,
			expectedStatus: 401,
		},
		{
			name: "GetSettingsInsufficientPermissions",
			input: &struct {
				middleware.AuthInput
				models.SubforumSettingsGetInput
			}{
				AuthInput: middleware.AuthInput{},
				SubforumSettingsGetInput: models.SubforumSettingsGetInput{
					Type: "b",
					Name: "hashpost",
				},
			},
			userCtx:        fixtures.CreateTestUserContext(),
			mockSubforum:   fixtures.CreateTestSubforum(),
			mockPermission: false,
			wantErr:        true,
			expectedStatus: 403,
		},
		{
			name: "GetSettingsSubforumNotFound",
			input: &struct {
				middleware.AuthInput
				models.SubforumSettingsGetInput
			}{
				AuthInput: middleware.AuthInput{},
				SubforumSettingsGetInput: models.SubforumSettingsGetInput{
					Type: "b",
					Name: "nonexistent",
				},
			},
			userCtx:        fixtures.CreateTestUserContext(),
			mockSubforum:   nil,
			mockPermission: false,
			wantErr:        true,
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize global auth middleware for testing
			authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
			middleware.SetGlobalAuthMiddleware(authMiddleware)

			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock expectations only if authentication should succeed
			if tt.userCtx != nil {
				if tt.mockSubforum != nil {
					mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(tt.mockSubforum, nil)
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_subforum_settings", &tt.mockSubforum.SubforumID).Return(tt.mockPermission, nil)
				} else {
					mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(nil, nil)
				}
			}

			// Create handler
			handler := &SubforumHandler{
				subforumDAO:   mockSubforumDAO,
				permissionDAO: mockPermissionDAO,
			}

			// Set up the input with proper authentication context
			var authInput middleware.AuthInput
			if tt.userCtx != nil {
				// Create a valid JWT token for testing
				token, tokenErr := middleware.GenerateJWT(tt.userCtx, "test-secret", 24*time.Hour)
				require.NoError(t, tokenErr)
				authInput = middleware.AuthInput{
					Authorization: "Bearer " + token,
					AccessToken:   "",
				}
			} else {
				authInput = middleware.AuthInput{
					Authorization: "",
					AccessToken:   "",
				}
			}

			// Update the input with proper authentication
			inputWithAuth := &struct {
				middleware.AuthInput
				models.SubforumSettingsGetInput
			}{
				AuthInput:                authInput,
				SubforumSettingsGetInput: tt.input.SubforumSettingsGetInput,
			}

			// Execute test
			result, err := handler.GetSubforumSettings(context.Background(), inputWithAuth)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
				assert.Equal(t, tt.mockSubforum.SubforumID, result.Body.SubforumID)
				assert.Equal(t, tt.mockSubforum.Name, result.Body.Name)
			}

			// Verify mocks only if they were set up
			if tt.userCtx != nil {
				mockSubforumDAO.AssertExpectations(t)
				mockPermissionDAO.AssertExpectations(t)
			}
		})
	}
}

func TestSubforumHandler_UpdateSubforumSettings(t *testing.T) {
	t.Skip("TODO: Mock database operations for UpdateSubforumSettings test")
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			models.SubforumSettingsInput
		}
		userCtx        *middleware.UserContext
		mockSubforum   *dbmodels.Subforum
		mockPermission bool
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "UpdateSettingsSuccess",
			input: &struct {
				middleware.AuthInput
				models.SubforumSettingsInput
			}{
				AuthInput: middleware.AuthInput{},
				SubforumSettingsInput: models.SubforumSettingsInput{
					Type: "b",
					Name: "hashpost",
					Body: models.SubforumSettings{
						AllowImages:            true,
						AllowVideos:            false,
						AllowPolls:             true,
						RequireFlair:           false,
						MinimumAccountAgeHours: 24,
						MinimumKarmaRequired:   10,
						IsPrivate:              false,
						IsRestricted:           false,
						IsNSFW:                 false,
						SidebarText:            "Updated sidebar text",
					},
				},
			},
			userCtx: fixtures.CreateTestUserContext(),
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "hashpost",
			},
			mockPermission: true,
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "UpdateSettingsUnauthorized",
			input: &struct {
				middleware.AuthInput
				models.SubforumSettingsInput
			}{
				AuthInput: middleware.AuthInput{},
				SubforumSettingsInput: models.SubforumSettingsInput{
					Type: "b",
					Name: "hashpost",
					Body: models.SubforumSettings{},
				},
			},
			userCtx:        nil,
			mockSubforum:   nil,
			mockPermission: false,
			wantErr:        true,
			expectedStatus: 401,
		},
		{
			name: "UpdateSettingsInsufficientPermissions",
			input: &struct {
				middleware.AuthInput
				models.SubforumSettingsInput
			}{
				AuthInput: middleware.AuthInput{},
				SubforumSettingsInput: models.SubforumSettingsInput{
					Type: "b",
					Name: "hashpost",
					Body: models.SubforumSettings{},
				},
			},
			userCtx:        fixtures.CreateTestUserContext(),
			mockSubforum:   fixtures.CreateTestSubforum(),
			mockPermission: false,
			wantErr:        true,
			expectedStatus: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock expectations only if authentication should succeed
			if tt.userCtx != nil {
				if tt.mockSubforum != nil {
					mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(tt.mockSubforum, nil)
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_subforum_settings", &tt.mockSubforum.SubforumID).Return(tt.mockPermission, nil)
				} else {
					mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(nil, nil)
				}
			}

			// Create handler
			handler := &SubforumHandler{
				subforumDAO:   mockSubforumDAO,
				permissionDAO: mockPermissionDAO,
			}

			// Set up the input with proper authentication context
			var authInput middleware.AuthInput
			if tt.userCtx != nil {
				// Create a valid JWT token for testing
				token, tokenErr := middleware.GenerateJWT(tt.userCtx, "test-secret", 24*time.Hour)
				require.NoError(t, tokenErr)
				authInput = middleware.AuthInput{
					Authorization: "Bearer " + token,
					AccessToken:   "",
				}
			} else {
				authInput = middleware.AuthInput{
					Authorization: "",
					AccessToken:   "",
				}
			}

			// Update the input with proper authentication
			inputWithAuth := &struct {
				middleware.AuthInput
				models.SubforumSettingsInput
			}{
				AuthInput:             authInput,
				SubforumSettingsInput: tt.input.SubforumSettingsInput,
			}

			// Execute test
			result, err := handler.UpdateSubforumSettings(context.Background(), inputWithAuth)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
				assert.Equal(t, tt.mockSubforum.SubforumID, result.Body.SubforumID)
				assert.Equal(t, tt.mockSubforum.Name, result.Body.Name)
			}

			// Verify mocks only if they were set up
			if tt.userCtx != nil {
				mockSubforumDAO.AssertExpectations(t)
				mockPermissionDAO.AssertExpectations(t)
			}
		})
	}
}

func TestSubforumHandler_GetModeratorTeam(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			models.ModeratorTeamInput
		}
		userCtx        *middleware.UserContext
		mockSubforum   *dbmodels.Subforum
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "GetModeratorTeamSuccess",
			input: &struct {
				middleware.AuthInput
				models.ModeratorTeamInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + generateJWTWithCapabilities(1, "moderator-pseudonym-123", []string{"create_content", "vote", "message", "report", "create_subforum", "manage_moderators"}),
				},
				ModeratorTeamInput: models.ModeratorTeamInput{
					Type: "b",
					Name: "hashpost",
				},
			},
			userCtx: &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: "moderator-pseudonym-123",
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "vote", "message", "report", "create_subforum", "manage_moderators"},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID:       1,
				Name:             "hashpost",
				OwnerPseudonymID: sql.Null[string]{V: "owner-pseudonym-123", Valid: true},
			},
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetModeratorTeamUnauthorized",
			input: &struct {
				middleware.AuthInput
				models.ModeratorTeamInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "", // No JWT token
				},
				ModeratorTeamInput: models.ModeratorTeamInput{
					Type: "b",
					Name: "hashpost",
				},
			},
			userCtx:        nil,
			mockSubforum:   nil,
			wantErr:        true,
			expectedStatus: 401,
		},
		{
			name: "GetModeratorTeamInsufficientPermissions",
			input: &struct {
				middleware.AuthInput
				models.ModeratorTeamInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + generateJWTWithCapabilities(1, "moderator-pseudonym-123", []string{"create_content", "vote", "message", "report", "create_subforum"}),
				},
				ModeratorTeamInput: models.ModeratorTeamInput{
					Type: "b",
					Name: "hashpost",
				},
			},
			userCtx: &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: "moderator-pseudonym-123",
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "vote", "message", "report", "create_subforum"}, // No manage_moderators capability
			},
			mockSubforum:   fixtures.CreateTestSubforum(),
			wantErr:        true,
			expectedStatus: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()

			// Create mock DAOs for all dependencies
			mockPermissionDAO := mocks.NewMockPermissionDAO()
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPostDAO := mocks.NewMockPostDAO()
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Set up mock expectations for subforum lookup
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(tt.mockSubforum, nil)
			} else {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(nil, nil)
			}

			// Set up mock expectations for GetModeratorsForSubforum
			if tt.mockSubforum != nil {
				// Create a mock role key for the owner if the subforum has an owner
				var roleKeys []*dbmodels.RoleKey
				if tt.mockSubforum.OwnerPseudonymID.Valid {
					roleKey := &dbmodels.RoleKey{
						PseudonymID:  tt.mockSubforum.OwnerPseudonymID.V,
						RoleName:     "moderator",
						IsActive:     sql.Null[bool]{V: true, Valid: true},
						CreatedAt:    sql.Null[time.Time]{V: time.Now(), Valid: true},
						CreatedBy:    "system",
						Capabilities: types.JSON[json.RawMessage]{},
					}
					roleKeys = append(roleKeys, roleKey)

					// Set up mock expectation for GetPseudonymByID
					mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, tt.mockSubforum.OwnerPseudonymID.V).Return(&dbmodels.Pseudonym{
						PseudonymID: tt.mockSubforum.OwnerPseudonymID.V,
						DisplayName: "Owner User",
						IsDefault:   true,
					}, nil)
				}
				mockRoleKeyDAO.On("GetModeratorsForSubforum", mock.Anything, tt.mockSubforum.SubforumID).Return(roleKeys, nil)
			}

			// Set up mock expectations for permission check
			if tt.userCtx != nil && tt.mockSubforum != nil {
				if !tt.wantErr {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(true, nil)
				} else {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(false, nil)
				}
			}

			// Set up mock expectations for pseudonym lookups
			// Note: This functionality has been moved to role keys system

			// Create handler with all dependencies
			handler := NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockIdentityMappingDAO, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO)

			// Create context with user if provided
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Execute test
			result, err := handler.GetModeratorTeam(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockSubforum.SubforumID, result.Body.SubforumID)
				assert.Equal(t, tt.mockSubforum.Name, result.Body.Name)
				assert.NotEmpty(t, result.Body.Owner.PseudonymID)
			}

			// Verify mocks only for successful cases
			if !tt.wantErr {
				mockSubforumDAO.AssertExpectations(t)
			}
		})
	}
}

func TestSubforumHandler_AddModerator(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			models.AddModeratorInput
		}
		userCtx        *middleware.UserContext
		mockSubforum   *dbmodels.Subforum
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "AddModeratorSuccess",
			input: &struct {
				middleware.AuthInput
				models.AddModeratorInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + generateJWTWithCapabilities(1, "admin-pseudonym-123", []string{"manage_moderators"}),
				},
				AddModeratorInput: models.AddModeratorInput{
					Type: "b",
					Name: "hashpost",
					Body: struct {
						PseudonymID  string   `json:"pseudonym_id" example:"abc123"`
						Role         string   `json:"role" example:"moderator"`
						Capabilities []string `json:"capabilities" example:"moderate_content,ban_users"`
					}{
						PseudonymID:  "new-moderator-123",
						Role:         "moderator",
						Capabilities: []string{"moderate_content", "ban_users"},
					},
				},
			},
			userCtx: &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: "admin-pseudonym-123",
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"manage_moderators"},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "hashpost",
			},
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "AddModeratorUnauthorized",
			input: &struct {
				middleware.AuthInput
				models.AddModeratorInput
			}{
				AuthInput: middleware.AuthInput{},
				AddModeratorInput: models.AddModeratorInput{
					Type: "b",
					Name: "hashpost",
					Body: struct {
						PseudonymID  string   `json:"pseudonym_id" example:"abc123"`
						Role         string   `json:"role" example:"moderator"`
						Capabilities []string `json:"capabilities" example:"moderate_content,ban_users"`
					}{},
				},
			},
			userCtx:        nil,
			mockSubforum:   nil,
			wantErr:        true,
			expectedStatus: 401,
		},
		{
			name: "AddModeratorInsufficientPermissions",
			input: &struct {
				middleware.AuthInput
				models.AddModeratorInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + generateJWTWithCapabilities(1, "test-pseudonym-id", []string{"create_content", "vote"}),
				},
				AddModeratorInput: models.AddModeratorInput{
					Type: "b",
					Name: "hashpost",
					Body: struct {
						PseudonymID  string   `json:"pseudonym_id" example:"abc123"`
						Role         string   `json:"role" example:"moderator"`
						Capabilities []string `json:"capabilities" example:"moderate_content,ban_users"`
					}{},
				},
			},
			userCtx: &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: "test-pseudonym-id",
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "vote"}, // No manage_moderators capability
			},
			mockSubforum:   fixtures.CreateTestSubforum(),
			wantErr:        true,
			expectedStatus: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs for all dependencies
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPostDAO := mocks.NewMockPostDAO()
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Set up mock expectations for subforum lookup
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(tt.mockSubforum, nil)
			} else if tt.input.AuthInput.Authorization != "" {
				// Only set up mock if there's an authorization header (meaning the request will proceed past auth)
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(nil, nil)
			}

			// Set up mock expectations for permission check
			if tt.userCtx != nil && tt.mockSubforum != nil {
				if !tt.wantErr {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(true, nil)
				} else {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(false, nil)
				}
			}

			// Set up mock expectations for moderator creation
			if tt.userCtx != nil && tt.mockSubforum != nil && !tt.wantErr {
				// Mock role key check - no existing moderator
				mockRoleKeyDAO.On("GetRoleKey", mock.Anything, tt.input.Body.PseudonymID, "moderation", &tt.mockSubforum.SubforumID).Return(nil, nil)
				// Mock role key creation
				mockRoleKeyDAO.On("CreateRoleKey", mock.Anything, "moderator", "moderation", []byte{}, mock.Anything, mock.Anything, tt.userCtx.ActivePseudonymID, tt.input.Body.PseudonymID, &tt.mockSubforum.SubforumID).Return(&dbmodels.RoleKey{}, nil)
			}

			// Create handler with all dependencies
			handler := &SubforumHandler{
				subforumDAO:             mockSubforumDAO,
				permissionDAO:           mockPermissionDAO,
				pseudonymDAO:            mockPseudonymDAO,
				identityMappingDAO:      mockIdentityMappingDAO,
				subforumSubscriptionDAO: mockSubforumSubscriptionDAO,
				postDAO:                 mockPostDAO,
				roleKeyDAO:              mockRoleKeyDAO,
			}

			// Create context with user if provided
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Execute test
			result, err := handler.AddModerator(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Body.PseudonymID, result.PseudonymID)
				assert.Equal(t, tt.input.Body.Role, result.Role)
				assert.Equal(t, tt.input.Body.Capabilities, result.Capabilities)
			}

			// Verify mocks
			mockSubforumDAO.AssertExpectations(t)
			mockRoleKeyDAO.AssertExpectations(t)
		})
	}
}

func TestSubforumHandler_UpdateModerator(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			models.UpdateModeratorInput
		}
		userCtx        *middleware.UserContext
		mockSubforum   *dbmodels.Subforum
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "UpdateModeratorSuccess",
			input: &struct {
				middleware.AuthInput
				models.UpdateModeratorInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + generateJWTWithCapabilities(1, "admin-pseudonym-123", []string{"manage_moderators"}),
				},
				UpdateModeratorInput: models.UpdateModeratorInput{
					Type:        "b",
					Name:        "hashpost",
					PseudonymID: "moderator-123",
					Body: struct {
						Role         string   `json:"role" example:"moderator"`
						Capabilities []string `json:"capabilities" example:"moderate_content,ban_users"`
						IsActive     bool     `json:"is_active" example:"true"`
					}{
						Role:         "senior_moderator",
						Capabilities: []string{"moderate_content", "ban_users", "manage_moderators"},
						IsActive:     true,
					},
				},
			},
			userCtx: &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: "admin-pseudonym-123",
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"manage_moderators"},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "hashpost",
			},
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "UpdateModeratorUnauthorized",
			input: &struct {
				middleware.AuthInput
				models.UpdateModeratorInput
			}{
				AuthInput: middleware.AuthInput{},
				UpdateModeratorInput: models.UpdateModeratorInput{
					Type:        "b",
					Name:        "hashpost",
					PseudonymID: "moderator-123",
					Body: struct {
						Role         string   `json:"role" example:"moderator"`
						Capabilities []string `json:"capabilities" example:"moderate_content,ban_users"`
						IsActive     bool     `json:"is_active" example:"true"`
					}{},
				},
			},
			userCtx:        nil,
			mockSubforum:   nil,
			wantErr:        true,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs for all dependencies
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPostDAO := mocks.NewMockPostDAO()
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Set up mock expectations for subforum lookup
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(tt.mockSubforum, nil)
			} else if tt.input.AuthInput.Authorization != "" {
				// Only set up mock if there's an authorization header (meaning the request will proceed past auth)
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(nil, nil)
			}

			// Set up mock expectations for permission check
			if tt.userCtx != nil && tt.mockSubforum != nil {
				if !tt.wantErr {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(true, nil)
				} else {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(false, nil)
				}
			}

			// Set up mock expectations for moderator update
			if tt.userCtx != nil && tt.mockSubforum != nil && !tt.wantErr {
				// Mock role key check - existing moderator found
				mockRoleKeyDAO.On("GetRoleKey", mock.Anything, tt.input.PseudonymID, "moderation", &tt.mockSubforum.SubforumID).Return(&dbmodels.RoleKey{}, nil)
			}

			// Create handler with all dependencies
			handler := &SubforumHandler{
				subforumDAO:             mockSubforumDAO,
				permissionDAO:           mockPermissionDAO,
				pseudonymDAO:            mockPseudonymDAO,
				identityMappingDAO:      mockIdentityMappingDAO,
				subforumSubscriptionDAO: mockSubforumSubscriptionDAO,
				postDAO:                 mockPostDAO,
				roleKeyDAO:              mockRoleKeyDAO,
			}

			// Create context with user if provided
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Execute test
			result, err := handler.UpdateModerator(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.PseudonymID, result.PseudonymID)
				assert.Equal(t, tt.input.Body.Role, result.Role)
				assert.Equal(t, tt.input.Body.Capabilities, result.Capabilities)
			}

			// Verify mocks
			mockSubforumDAO.AssertExpectations(t)
			mockRoleKeyDAO.AssertExpectations(t)
		})
	}
}

func TestSubforumHandler_RemoveModerator(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			models.RemoveModeratorInput
		}
		userCtx        *middleware.UserContext
		mockSubforum   *dbmodels.Subforum
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "RemoveModeratorSuccess",
			input: &struct {
				middleware.AuthInput
				models.RemoveModeratorInput
			}{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + generateJWTWithCapabilities(1, "admin-pseudonym-123", []string{"manage_moderators"}),
				},
				RemoveModeratorInput: models.RemoveModeratorInput{
					Type:        "b",
					Name:        "hashpost",
					PseudonymID: "moderator-123",
				},
			},
			userCtx: &middleware.UserContext{
				UserID:            1,
				Email:             "user@example.com",
				ActivePseudonymID: "admin-pseudonym-123",
				DisplayName:       "TestUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"manage_moderators"},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "hashpost",
			},
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "RemoveModeratorUnauthorized",
			input: &struct {
				middleware.AuthInput
				models.RemoveModeratorInput
			}{
				AuthInput: middleware.AuthInput{},
				RemoveModeratorInput: models.RemoveModeratorInput{
					Type:        "b",
					Name:        "hashpost",
					PseudonymID: "moderator-123",
				},
			},
			userCtx:        nil,
			mockSubforum:   nil,
			wantErr:        true,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs for all dependencies
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()
			mockPseudonymDAO := mocks.NewMockPseudonymDAO()
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()
			mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
			mockPostDAO := mocks.NewMockPostDAO()
			mockRoleKeyDAO := mocks.NewMockRoleKeyDAO()

			// Set up mock expectations for subforum lookup
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(tt.mockSubforum, nil)
			} else if tt.input.AuthInput.Authorization != "" {
				// Only set up mock if there's an authorization header (meaning the request will proceed past auth)
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.Type, tt.input.Name).Return(nil, nil)
			}

			// Set up mock expectations for permission check
			if tt.userCtx != nil && tt.mockSubforum != nil {
				if !tt.wantErr {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(true, nil)
				} else {
					mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_moderators", &tt.mockSubforum.SubforumID).Return(false, nil)
				}
			}

			// Set up mock expectations for moderator removal
			if tt.userCtx != nil && tt.mockSubforum != nil && !tt.wantErr {
				// Mock role key check - existing moderator found
				mockRoleKeyDAO.On("GetRoleKey", mock.Anything, tt.input.PseudonymID, "moderation", &tt.mockSubforum.SubforumID).Return(&dbmodels.RoleKey{KeyID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}, nil)
				// Mock role key deactivation
				mockRoleKeyDAO.On("DeactivateRoleKey", mock.Anything, mock.Anything).Return(nil)
			}

			// Create handler with all dependencies
			handler := &SubforumHandler{
				subforumDAO:             mockSubforumDAO,
				permissionDAO:           mockPermissionDAO,
				pseudonymDAO:            mockPseudonymDAO,
				identityMappingDAO:      mockIdentityMappingDAO,
				subforumSubscriptionDAO: mockSubforumSubscriptionDAO,
				postDAO:                 mockPostDAO,
				roleKeyDAO:              mockRoleKeyDAO,
			}

			// Create context with user if provided
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Execute test
			result, err := handler.RemoveModerator(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.Equal(t, "Moderator removed successfully", result.Message)
			}

			// Verify mocks
			mockSubforumDAO.AssertExpectations(t)
			mockRoleKeyDAO.AssertExpectations(t)
		})
	}
}
