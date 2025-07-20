package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// createTestSubforumUserContext creates a test user context with subforum capabilities
func createTestSubforumUserContext() *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            1,
		Email:             "user@example.com",
		ActivePseudonymID: "moderator-pseudonym-123",
		DisplayName:       "TestUser",
		Roles:             []string{"user"},
		Capabilities:      []string{"create_content", "vote", "create_subforum"},
	}
}

// createTestSubforumContextWithUser creates a context with user information
func createTestSubforumContextWithUser(userCtx *middleware.UserContext) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, middleware.UserContextKeyValue, userCtx)
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
			userCtx:       createTestSubforumUserContext(),
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
			userCtx:       createTestSubforumUserContext(),
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
			mockSubforumModeratorDAO := mocks.NewMockSubforumModeratorDAO()

			// Set up mock data
			mockSubforumDAO.InjectAllSubforums([]*dbmodels.Subforum{fixtures.CreateTestSubforum(), fixtures.CreateTestPrivateSubforum()})
			mockSubforumDAO.SetDefaultBehavior()

			mockSubforumSubscriptionDAO.InjectCount("subforum_1", 100)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			mockPermissionDAO.InjectAccessPermission(1, 1, true) // Allow access to private subforum
			mockPermissionDAO.SetDefaultBehavior()

			mockSubforumModeratorDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create handler with mocks
			handler := NewSubforumHandlerWithMocks(mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockSubforumModeratorDAO, mockIdentityMappingDAO, nil)

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
				SubforumName: "test-subforum",
			},
			userCtx:        createTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetPrivateSubforumDetailsWithAccess",
			input: &models.SubforumSubscriptionInput{
				SubforumName: "private-test-subforum",
			},
			userCtx:        createTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetPrivateSubforumDetailsWithoutAccess",
			input: &models.SubforumSubscriptionInput{
				SubforumName: "private-test-subforum",
			},
			userCtx:        nil,
			wantErr:        true,
			expectedStatus: 403,
		},
		{
			name: "GetNonExistentSubforum",
			input: &models.SubforumSubscriptionInput{
				SubforumName: "non-existent",
			},
			userCtx:        createTestUserContext(),
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
			mockSubforumModeratorDAO := mocks.NewMockSubforumModeratorDAO()

			// Set up mock data
			if tt.input.SubforumName == "test-subforum" {
				mockSubforumDAO.InjectSubforumByName("test-subforum", fixtures.CreateTestSubforum())
			} else if tt.input.SubforumName == "private-test-subforum" {
				mockSubforumDAO.InjectSubforumByName("private-test-subforum", fixtures.CreateTestPrivateSubforum())
			}
			mockSubforumDAO.SetDefaultBehavior()

			mockSubforumSubscriptionDAO.InjectSubscribedStatus("moderator-pseudonym-123", 1, true)
			mockSubforumSubscriptionDAO.InjectFavoriteStatus("moderator-pseudonym-123", 1, false)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			mockPermissionDAO.InjectAccessPermission(1, 1, true)
			mockPermissionDAO.SetDefaultBehavior()

			mockSubforumModeratorDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create handler with mocks
			handler := NewSubforumHandlerWithMocks(mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockSubforumModeratorDAO, mockIdentityMappingDAO, nil)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			response, err := handler.GetSubforumDetails(ctx, tt.input)

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
				SubforumName: "test-subforum",
			},
			userCtx:        createTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "SubscribeToNonExistentSubforum",
			input: &models.SubforumSubscriptionInput{
				SubforumName: "non-existent",
			},
			userCtx:        createTestUserContext(),
			wantErr:        true,
			expectedStatus: 404,
		},
		{
			name: "SubscribeWithoutAuthentication",
			input: &models.SubforumSubscriptionInput{
				SubforumName: "test-subforum",
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
			mockSubforumModeratorDAO := mocks.NewMockSubforumModeratorDAO()

			// Set up mock data
			if tt.input.SubforumName == "test-subforum" {
				mockSubforumDAO.InjectSubforumByName("test-subforum", fixtures.CreateTestSubforum())
			}
			mockSubforumDAO.SetDefaultBehavior()

			mockSubforumSubscriptionDAO.InjectSubscribedStatus("moderator-pseudonym-123", 1, false) // Not subscribed initially
			mockSubforumSubscriptionDAO.InjectCount("subforum_1", 100)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			mockPermissionDAO.SetDefaultBehavior()
			mockSubforumModeratorDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create handler with mocks
			handler := NewSubforumHandlerWithMocks(mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockSubforumModeratorDAO, mockIdentityMappingDAO, nil)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			response, err := handler.SubscribeToSubforum(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Body.Subscribed)
			}
		})
	}
}

func TestSubforumHandler_UnsubscribeFromSubforum(t *testing.T) {
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
				SubforumName: "test-subforum",
			},
			userCtx:        createTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "UnsubscribeFromNonExistentSubforum",
			input: &models.SubforumSubscriptionInput{
				SubforumName: "non-existent",
			},
			userCtx:        createTestUserContext(),
			wantErr:        true,
			expectedStatus: 404,
		},
		{
			name: "UnsubscribeWithoutAuthentication",
			input: &models.SubforumSubscriptionInput{
				SubforumName: "test-subforum",
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
			mockSubforumModeratorDAO := mocks.NewMockSubforumModeratorDAO()

			// Set up mock data
			if tt.input.SubforumName == "test-subforum" {
				mockSubforumDAO.InjectSubforumByName("test-subforum", fixtures.CreateTestSubforum())
			}
			mockSubforumDAO.SetDefaultBehavior()

			// Create a subscription record for the mock
			testSubscription := &dbmodels.SubforumSubscription{
				SubscriptionID: 123,
				PseudonymID:    "moderator-pseudonym-123",
				SubforumID:     1,
				IsFavorite:     sql.Null[bool]{V: false, Valid: true},
				SubscribedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
			}
			mockSubforumSubscriptionDAO.InjectSubscription(testSubscription)
			mockSubforumSubscriptionDAO.InjectSubscribedStatus("moderator-pseudonym-123", 1, true) // Subscribed initially
			mockSubforumSubscriptionDAO.InjectCount("subforum_1", 100)
			mockSubforumSubscriptionDAO.SetDefaultBehavior()

			mockPermissionDAO.SetDefaultBehavior()
			mockSubforumModeratorDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create handler with mocks
			handler := NewSubforumHandlerWithMocks(mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockSubforumModeratorDAO, mockIdentityMappingDAO, nil)

			// Create context
			ctx := context.Background()
			if tt.userCtx != nil {
				ctx = createTestContextWithUser(tt.userCtx)
			}

			// Call method
			response, err := handler.UnsubscribeFromSubforum(ctx, tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.False(t, response.Body.Subscribed)
			}
		})
	}
}

func TestSubforumHandler_CreateSubforum(t *testing.T) {
	tests := []struct {
		name           string
		input          *models.SubforumCreateInput
		userCtx        *middleware.UserContext
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "CreateSubforumWithPermission",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "moderator-pseudonym-123"),
				Body: models.SubforumCreateBody{
					Slug:         "new-subforum",
					Name:         "New Subforum",
					Description:  "A new test subforum",
					IsNSFW:       false,
					IsPrivate:    false,
					IsRestricted: false,
				},
			},
			userCtx:        createTestUserContext(),
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "CreateSubforumWithoutPermission",
			input: &models.SubforumCreateInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "moderator-pseudonym-123"),
				Body: models.SubforumCreateBody{
					Slug:         "new-subforum",
					Name:         "New Subforum",
					Description:  "A new test subforum",
					IsNSFW:       false,
					IsPrivate:    false,
					IsRestricted: false,
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
					Slug:         "new-subforum",
					Name:         "New Subforum",
					Description:  "A new test subforum",
					IsNSFW:       false,
					IsPrivate:    false,
					IsRestricted: false,
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
			mockSubforumModeratorDAO := mocks.NewMockSubforumModeratorDAO()

			// Set up mock data
			mockSubforumDAO.SetDefaultBehavior()
			mockSubforumSubscriptionDAO.SetDefaultBehavior()
			mockPermissionDAO.SetDefaultBehavior()
			mockSubforumModeratorDAO.SetDefaultBehavior()

			// Create mock identity mapping DAO
			mockIdentityMappingDAO := mocks.NewMockIdentityMappingDAO()

			// Create handler with mocks
			handler := NewSubforumHandlerWithMocks(mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockSubforumModeratorDAO, mockIdentityMappingDAO, nil)

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
			}
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
			userCtx:        createTestUserContext(),
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
			userCtx:        createTestUserContext(),
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
			mockSubforumModeratorDAO := mocks.NewMockSubforumModeratorDAO()

			// Set up mock data
			mockSubforumDAO.InjectSubforum(fixtures.CreateTestSubforum())
			mockSubforumDAO.SetDefaultBehavior()
			mockSubforumSubscriptionDAO.InjectSubscriptionsByPseudonym("moderator-pseudonym-123", []*dbmodels.SubforumSubscription{fixtures.CreateTestSubforumSubscription()})
			mockSubforumSubscriptionDAO.SetDefaultBehavior()
			mockPermissionDAO.SetDefaultBehavior()
			mockSubforumModeratorDAO.SetDefaultBehavior()

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

			// Create handler with mocks
			handler := NewSubforumHandlerWithMocks(mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, mockSubforumModeratorDAO, mockIdentityMappingDAO, nil)

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
