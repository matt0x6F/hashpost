package handlers

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestAdminHandler_Constructor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock DAOs using gomock
	mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
	mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

	// Create admin handler
	handler := NewAdminHandler(
		mockUserDAO,
		mockPseudonymDAO,
		mockPermissionDAO,
		nil, // passwordResetTokenDAO - not available as mock
		nil, // emailService
		nil, // config
		mockPostDAO,
		mockCommentDAO,
		mockRoleKeyDAO,
		mockSubforumDAO,
	)

	// Test that the handler was created successfully
	require.NotNil(t, handler)
	assert.NotNil(t, handler)
}

func TestAdminHandler_MockSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test that we can create mocks successfully
	mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

	// Verify mocks are created
	require.NotNil(t, mockRoleKeyDAO)
	require.NotNil(t, mockSubforumDAO)

	// Test that we can set expectations on the mocks
	// This verifies the mock generation is working correctly
	mockRoleKeyDAO.EXPECT().ListRoleKeysByPseudonym(gomock.Any(), gomock.Any()).Return(nil, nil)
	mockSubforumDAO.EXPECT().GetSubforumByID(gomock.Any(), gomock.Any()).Return(nil, nil)

	// Actually call the methods to satisfy the expectations
	ctx := context.Background()
	_, _ = mockRoleKeyDAO.ListRoleKeysByPseudonym(ctx, "test-pseudonym")
	_, _ = mockSubforumDAO.GetSubforumByID(ctx, 1)
}

func TestAdminHandler_ListUsers(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *ListUsersInput
		setupMocks     func(*dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockPermissionDAOInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful user list",
			input: &ListUsersInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user has system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user count
				userDAO.EXPECT().
					CountUsers(gomock.Any()).
					Return(int64(2), nil)

				// Mock user list
				testUsers := []*dbmodels.User{
					{
						UserID:        1,
						Email:         "user1@example.com",
						CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
						IsActive:      sql.Null[bool]{V: true, Valid: true},
						IsSuspended:   sql.Null[bool]{V: false, Valid: true},
						EmailVerified: sql.Null[bool]{V: true, Valid: true},
					},
					{
						UserID:        2,
						Email:         "user2@example.com",
						CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
						IsActive:      sql.Null[bool]{V: true, Valid: true},
						IsSuspended:   sql.Null[bool]{V: false, Valid: true},
						EmailVerified: sql.Null[bool]{V: false, Valid: true},
					},
				}
				userDAO.EXPECT().
					ListUsers(gomock.Any(), 25, 0).
					Return(testUsers, nil)

				// Mock pseudonym counts
				pseudonymDAO.EXPECT().
					GetPseudonymsByRealIdentityDirect(gomock.Any(), "user1@example.com").
					Return([]*dbmodels.Pseudonym{{PseudonymID: "pseudo1"}}, nil)
				pseudonymDAO.EXPECT().
					GetPseudonymsByRealIdentityDirect(gomock.Any(), "user2@example.com").
					Return([]*dbmodels.Pseudonym{{PseudonymID: "pseudo2"}, {PseudonymID: "pseudo3"}}, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "unauthorized - no token",
			input: &ListUsersInput{
				AuthInput: middleware.AuthInput{},
				Page:      1,
				Limit:     25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// No mocks needed - should fail at authentication
			},
			expectedStatus: 401,
			expectedError:  "authentication required",
		},
		{
			name: "forbidden - insufficient permissions",
			input: &ListUsersInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "user-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user lacks system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "user-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(false, nil)
			},
			expectedStatus: 403,
			expectedError:  "insufficient permissions: platform admin capability required",
		},
		{
			name: "database error on user count",
			input: &ListUsersInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user count error
				userDAO.EXPECT().
					CountUsers(gomock.Any()).
					Return(int64(0), errors.New("database error"))
			},
			expectedError: "failed to count users: database error",
		},
		{
			name: "database error on user list",
			input: &ListUsersInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user count
				userDAO.EXPECT().
					CountUsers(gomock.Any()).
					Return(int64(1), nil)

				// Mock user list error
				userDAO.EXPECT().
					ListUsers(gomock.Any(), 25, 0).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get users: database error",
		},
		{
			name: "search query filters results",
			input: &ListUsersInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  1,
				Limit: 25,
				Query: "user1",
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user count
				userDAO.EXPECT().
					CountUsers(gomock.Any()).
					Return(int64(2), nil)

				// Mock user list
				testUsers := []*dbmodels.User{
					{
						UserID:        1,
						Email:         "user1@example.com",
						CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
						IsActive:      sql.Null[bool]{V: true, Valid: true},
						IsSuspended:   sql.Null[bool]{V: false, Valid: true},
						EmailVerified: sql.Null[bool]{V: true, Valid: true},
					},
					{
						UserID:        2,
						Email:         "user2@example.com",
						CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
						IsActive:      sql.Null[bool]{V: true, Valid: true},
						IsSuspended:   sql.Null[bool]{V: false, Valid: true},
						EmailVerified: sql.Null[bool]{V: false, Valid: true},
					},
				}
				userDAO.EXPECT().
					ListUsers(gomock.Any(), 25, 0).
					Return(testUsers, nil)

				// Mock pseudonym count for filtered user
				pseudonymDAO.EXPECT().
					GetPseudonymsByRealIdentityDirect(gomock.Any(), "user1@example.com").
					Return([]*dbmodels.Pseudonym{{PseudonymID: "pseudo1"}}, nil)
			},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
			mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
			mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
			mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
			mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
			mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
			mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

			// Setup mocks
			tt.setupMocks(mockUserDAO, mockPseudonymDAO, mockPermissionDAO)

			// Create handler
			handler := NewAdminHandler(
				mockUserDAO,
				mockPseudonymDAO,
				mockPermissionDAO,
				nil, // passwordResetTokenDAO
				nil, // emailService
				nil, // config
				mockPostDAO,
				mockCommentDAO,
				mockRoleKeyDAO,
				mockSubforumDAO,
			)

			// Call method
			response, err := handler.ListUsers(context.Background(), tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)

			if tt.expectedStatus == 200 {
				assert.NotNil(t, response.Body)
				assert.NotNil(t, response.Body.Users)
				assert.NotNil(t, response.Body.Pagination)
			}
		})
	}
}

func TestAdminHandler_GetUser(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *GetUserInput
		setupMocks     func(*dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockPermissionDAOInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful user get",
			input: &GetUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user has system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user get
				testUser := &dbmodels.User{
					UserID:        123,
					Email:         "testuser@example.com",
					CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
					IsActive:      sql.Null[bool]{V: true, Valid: true},
					IsSuspended:   sql.Null[bool]{V: false, Valid: true},
					EmailVerified: sql.Null[bool]{V: true, Valid: true},
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(testUser, nil)

				// Mock pseudonym count
				pseudonymDAO.EXPECT().
					GetPseudonymsByRealIdentityDirect(gomock.Any(), "testuser@example.com").
					Return([]*dbmodels.Pseudonym{{PseudonymID: "pseudo1"}, {PseudonymID: "pseudo2"}}, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "unauthorized - no token",
			input: &GetUserInput{
				AuthInput: middleware.AuthInput{},
				UserID:    123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// No mocks needed - should fail at authentication
			},
			expectedStatus: 401,
			expectedError:  "authentication required",
		},
		{
			name: "forbidden - insufficient permissions",
			input: &GetUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "user-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user lacks system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "user-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(false, nil)
			},
			expectedStatus: 403,
			expectedError:  "insufficient permissions: platform admin capability required",
		},
		{
			name: "user not found",
			input: &GetUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 999,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user not found
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(999)).
					Return(nil, nil)
			},
			expectedStatus: 404,
			expectedError:  "user not found",
		},
		{
			name: "database error on user get",
			input: &GetUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user get error
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get user: database error",
		},
		{
			name: "pseudonym count error - should still succeed",
			input: &GetUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user get
				testUser := &dbmodels.User{
					UserID:        123,
					Email:         "testuser@example.com",
					CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
					IsActive:      sql.Null[bool]{V: true, Valid: true},
					IsSuspended:   sql.Null[bool]{V: false, Valid: true},
					EmailVerified: sql.Null[bool]{V: true, Valid: true},
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(testUser, nil)

				// Mock pseudonym count error - should not fail the request
				pseudonymDAO.EXPECT().
					GetPseudonymsByRealIdentityDirect(gomock.Any(), "testuser@example.com").
					Return(nil, errors.New("pseudonym error"))
			},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
			mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
			mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
			mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
			mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
			mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
			mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

			// Setup mocks
			tt.setupMocks(mockUserDAO, mockPseudonymDAO, mockPermissionDAO)

			// Create handler
			handler := NewAdminHandler(
				mockUserDAO,
				mockPseudonymDAO,
				mockPermissionDAO,
				nil, // passwordResetTokenDAO
				nil, // emailService
				nil, // config
				mockPostDAO,
				mockCommentDAO,
				mockRoleKeyDAO,
				mockSubforumDAO,
			)

			// Call method
			response, err := handler.GetUser(context.Background(), tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)

			if tt.expectedStatus == 200 {
				assert.NotNil(t, response.Body)
				assert.Equal(t, int64(123), response.Body.UserID)
				assert.Equal(t, "testuser@example.com", response.Body.Email)
			}
		})
	}
}

func TestAdminHandler_ListPseudonyms(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *ListPseudonymsInput
		setupMocks     func(*dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockPermissionDAOInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful pseudonym list",
			input: &ListPseudonymsInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user has system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock pseudonym count
				pseudonymDAO.EXPECT().
					CountAllPseudonyms(gomock.Any()).
					Return(int64(2), nil)

				// Mock pseudonym list
				testPseudonyms := []*dbmodels.Pseudonym{
					{
						PseudonymID: "pseudo1",
						DisplayName: "TestUser1",
						Slug:        sql.Null[string]{V: "test-user-1", Valid: true},
						KarmaScore:  sql.Null[int32]{V: 100, Valid: true},
						IsActive:    sql.Null[bool]{V: true, Valid: true},
						CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
					},
					{
						PseudonymID: "pseudo2",
						DisplayName: "TestUser2",
						Slug:        sql.Null[string]{Valid: false},
						KarmaScore:  sql.Null[int32]{V: 50, Valid: true},
						IsActive:    sql.Null[bool]{V: false, Valid: true},
						CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
					},
				}
				pseudonymDAO.EXPECT().
					GetAllPseudonyms(gomock.Any()).
					Return(testPseudonyms, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "unauthorized - no token",
			input: &ListPseudonymsInput{
				AuthInput: middleware.AuthInput{},
				Page:      1,
				Limit:     25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// No mocks needed - should fail at authentication
			},
			expectedStatus: 401,
			expectedError:  "authentication required",
		},
		{
			name: "forbidden - insufficient permissions",
			input: &ListPseudonymsInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "user-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user lacks system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "user-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(false, nil)
			},
			expectedStatus: 403,
			expectedError:  "insufficient permissions: platform admin capability required",
		},
		{
			name: "database error on pseudonym count",
			input: &ListPseudonymsInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock pseudonym count error
				pseudonymDAO.EXPECT().
					CountAllPseudonyms(gomock.Any()).
					Return(int64(0), errors.New("database error"))
			},
			expectedError: "failed to count pseudonyms: database error",
		},
		{
			name: "database error on pseudonym list",
			input: &ListPseudonymsInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  1,
				Limit: 25,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock pseudonym count
				pseudonymDAO.EXPECT().
					CountAllPseudonyms(gomock.Any()).
					Return(int64(1), nil)

				// Mock pseudonym list error
				pseudonymDAO.EXPECT().
					GetAllPseudonyms(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get pseudonyms: database error",
		},
		{
			name: "pagination with empty results",
			input: &ListPseudonymsInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				Page:  2,
				Limit: 10,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock pseudonym count
				pseudonymDAO.EXPECT().
					CountAllPseudonyms(gomock.Any()).
					Return(int64(5), nil)

				// Mock empty pseudonym list (page 2 with only 5 total items)
				pseudonymDAO.EXPECT().
					GetAllPseudonyms(gomock.Any()).
					Return([]*dbmodels.Pseudonym{}, nil)
			},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
			mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
			mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
			mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
			mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
			mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
			mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

			// Setup mocks
			tt.setupMocks(mockUserDAO, mockPseudonymDAO, mockPermissionDAO)

			// Create handler
			handler := NewAdminHandler(
				mockUserDAO,
				mockPseudonymDAO,
				mockPermissionDAO,
				nil, // passwordResetTokenDAO
				nil, // emailService
				nil, // config
				mockPostDAO,
				mockCommentDAO,
				mockRoleKeyDAO,
				mockSubforumDAO,
			)

			// Call method
			response, err := handler.ListPseudonyms(context.Background(), tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)

			if tt.expectedStatus == 200 {
				assert.NotNil(t, response.Body)
				assert.NotNil(t, response.Body.Pseudonyms)
				assert.NotNil(t, response.Body.Pagination)
			}
		})
	}
}

func TestAdminHandler_UpdateUser(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *UpdateUserInput
		setupMocks     func(*dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockPermissionDAOInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful user update - email only",
			input: &UpdateUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
				Body: struct {
					Email         *string `json:"email,omitempty" example:"newemail@example.com"`
					IsActive      *bool   `json:"is_active,omitempty" example:"true"`
					IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
					EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
				}{
					Email: stringPtr("newemail@example.com"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing user
				existingUser := &dbmodels.User{
					UserID:        123,
					Email:         "oldemail@example.com",
					CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
					IsActive:      sql.Null[bool]{V: true, Valid: true},
					IsSuspended:   sql.Null[bool]{V: false, Valid: true},
					EmailVerified: sql.Null[bool]{V: true, Valid: true},
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(existingUser, nil)

				// Mock user update
				userDAO.EXPECT().
					UpdateUser(gomock.Any(), int64(123), gomock.Any()).
					Return(nil)

				// Mock get updated user
				updatedUser := &dbmodels.User{
					UserID:        123,
					Email:         "newemail@example.com",
					CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
					IsActive:      sql.Null[bool]{V: true, Valid: true},
					IsSuspended:   sql.Null[bool]{V: false, Valid: true},
					EmailVerified: sql.Null[bool]{V: true, Valid: true},
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(updatedUser, nil)

				// Mock pseudonym count
				pseudonymDAO.EXPECT().
					GetPseudonymsByRealIdentityDirect(gomock.Any(), "newemail@example.com").
					Return([]*dbmodels.Pseudonym{{PseudonymID: "pseudo1"}}, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "successful user update - multiple fields",
			input: &UpdateUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
				Body: struct {
					Email         *string `json:"email,omitempty" example:"newemail@example.com"`
					IsActive      *bool   `json:"is_active,omitempty" example:"true"`
					IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
					EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
				}{
					Email:         stringPtr("newemail@example.com"),
					IsActive:      boolPtr(false),
					IsSuspended:   boolPtr(true),
					EmailVerified: boolPtr(false),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing user
				existingUser := &dbmodels.User{
					UserID:        123,
					Email:         "oldemail@example.com",
					CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
					IsActive:      sql.Null[bool]{V: true, Valid: true},
					IsSuspended:   sql.Null[bool]{V: false, Valid: true},
					EmailVerified: sql.Null[bool]{V: true, Valid: true},
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(existingUser, nil)

				// Mock user update
				userDAO.EXPECT().
					UpdateUser(gomock.Any(), int64(123), gomock.Any()).
					Return(nil)

				// Mock get updated user
				updatedUser := &dbmodels.User{
					UserID:        123,
					Email:         "newemail@example.com",
					CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
					IsActive:      sql.Null[bool]{V: false, Valid: true},
					IsSuspended:   sql.Null[bool]{V: true, Valid: true},
					EmailVerified: sql.Null[bool]{V: false, Valid: true},
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(updatedUser, nil)

				// Mock pseudonym count
				pseudonymDAO.EXPECT().
					GetPseudonymsByRealIdentityDirect(gomock.Any(), "newemail@example.com").
					Return([]*dbmodels.Pseudonym{{PseudonymID: "pseudo1"}}, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "unauthorized - no token",
			input: &UpdateUserInput{
				AuthInput: middleware.AuthInput{},
				UserID:    123,
				Body: struct {
					Email         *string `json:"email,omitempty" example:"newemail@example.com"`
					IsActive      *bool   `json:"is_active,omitempty" example:"true"`
					IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
					EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
				}{},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// No mocks needed - should fail at authentication
			},
			expectedStatus: 401,
			expectedError:  "authentication required",
		},
		{
			name: "forbidden - insufficient permissions",
			input: &UpdateUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "user-pseudonym"),
				},
				UserID: 123,
				Body: struct {
					Email         *string `json:"email,omitempty" example:"newemail@example.com"`
					IsActive      *bool   `json:"is_active,omitempty" example:"true"`
					IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
					EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
				}{},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user lacks system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "user-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(false, nil)
			},
			expectedStatus: 403,
			expectedError:  "insufficient permissions: platform admin capability required",
		},
		{
			name: "user not found",
			input: &UpdateUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 999,
				Body: struct {
					Email         *string `json:"email,omitempty" example:"newemail@example.com"`
					IsActive      *bool   `json:"is_active,omitempty" example:"true"`
					IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
					EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
				}{},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user not found
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(999)).
					Return(nil, nil)
			},
			expectedStatus: 404,
			expectedError:  "user not found",
		},
		{
			name: "database error on user get",
			input: &UpdateUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
				Body: struct {
					Email         *string `json:"email,omitempty" example:"newemail@example.com"`
					IsActive      *bool   `json:"is_active,omitempty" example:"true"`
					IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
					EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
				}{},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user get error
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get user: database error",
		},
		{
			name: "database error on user update",
			input: &UpdateUserInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
				Body: struct {
					Email         *string `json:"email,omitempty" example:"newemail@example.com"`
					IsActive      *bool   `json:"is_active,omitempty" example:"true"`
					IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
					EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
				}{
					Email: stringPtr("newemail@example.com"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing user
				existingUser := &dbmodels.User{
					UserID:        123,
					Email:         "oldemail@example.com",
					CreatedAt:     sql.Null[time.Time]{V: time.Now(), Valid: true},
					IsActive:      sql.Null[bool]{V: true, Valid: true},
					IsSuspended:   sql.Null[bool]{V: false, Valid: true},
					EmailVerified: sql.Null[bool]{V: true, Valid: true},
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(existingUser, nil)

				// Mock user update error
				userDAO.EXPECT().
					UpdateUser(gomock.Any(), int64(123), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "failed to update user: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
			mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
			mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
			mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
			mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
			mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
			mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

			// Setup mocks
			tt.setupMocks(mockUserDAO, mockPseudonymDAO, mockPermissionDAO)

			// Create handler
			handler := NewAdminHandler(
				mockUserDAO,
				mockPseudonymDAO,
				mockPermissionDAO,
				nil, // passwordResetTokenDAO
				nil, // emailService
				nil, // config
				mockPostDAO,
				mockCommentDAO,
				mockRoleKeyDAO,
				mockSubforumDAO,
			)

			// Call method
			response, err := handler.UpdateUser(context.Background(), tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)

			if tt.expectedStatus == 200 {
				assert.NotNil(t, response.Body)
				assert.Equal(t, int64(123), response.Body.UserID)
			}
		})
	}
}

func TestAdminHandler_TriggerPasswordReset(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *TriggerPasswordResetInput
		setupMocks     func(*dao.MockUserDAOInterface, *dao.MockPasswordResetTokenDAOInterface, *dao.MockPermissionDAOInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful password reset trigger",
			input: &TriggerPasswordResetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, passwordResetTokenDAO *dao.MockPasswordResetTokenDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get target user
				targetUser := &dbmodels.User{
					UserID: 123,
					Email:  "target@example.com",
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(targetUser, nil)

				// Mock create token
				passwordResetTokenDAO.EXPECT().
					CreateToken(gomock.Any(), int64(123), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: 200,
		},
		{
			name: "unauthorized - no token",
			input: &TriggerPasswordResetInput{
				AuthInput: middleware.AuthInput{},
				UserID:    123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, passwordResetTokenDAO *dao.MockPasswordResetTokenDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// No mocks needed - should fail at authentication
			},
			expectedStatus: 401,
			expectedError:  "authentication required",
		},
		{
			name: "forbidden - insufficient permissions",
			input: &TriggerPasswordResetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "user-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, passwordResetTokenDAO *dao.MockPasswordResetTokenDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user lacks system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "user-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(false, nil)
			},
			expectedStatus: 403,
			expectedError:  "insufficient permissions: platform admin capability required",
		},
		{
			name: "target user not found",
			input: &TriggerPasswordResetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 999,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, passwordResetTokenDAO *dao.MockPasswordResetTokenDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user not found
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(999)).
					Return(nil, nil)
			},
			expectedStatus: 404,
			expectedError:  "user not found",
		},
		{
			name: "database error on user get",
			input: &TriggerPasswordResetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, passwordResetTokenDAO *dao.MockPasswordResetTokenDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock user get error
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get target user: database error",
		},
		{
			name: "database error on token creation",
			input: &TriggerPasswordResetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				UserID: 123,
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, passwordResetTokenDAO *dao.MockPasswordResetTokenDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get target user
				targetUser := &dbmodels.User{
					UserID: 123,
					Email:  "target@example.com",
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(targetUser, nil)

				// Mock create token error
				passwordResetTokenDAO.EXPECT().
					CreateToken(gomock.Any(), int64(123), gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "failed to store reset token: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
			mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
			mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
			mockPasswordResetTokenDAO := dao.NewMockPasswordResetTokenDAOInterface(ctrl)
			mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
			mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
			mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
			mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

			// Setup mocks
			tt.setupMocks(mockUserDAO, mockPasswordResetTokenDAO, mockPermissionDAO)

			// Create handler
			handler := NewAdminHandler(
				mockUserDAO,
				mockPseudonymDAO,
				mockPermissionDAO,
				mockPasswordResetTokenDAO,
				nil, // emailService - nil for these tests
				nil, // config
				mockPostDAO,
				mockCommentDAO,
				mockRoleKeyDAO,
				mockSubforumDAO,
			)

			// Call method
			response, err := handler.TriggerPasswordReset(context.Background(), tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)

			if tt.expectedStatus == 200 {
				assert.NotNil(t, response.Body)
				assert.Equal(t, "Password reset email sent", response.Body.Message)
			}
		})
	}
}

func TestAdminHandler_GetPseudonym(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *GetPseudonymInput
		setupMocks     func(*dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockPermissionDAOInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful pseudonym get",
			input: &GetPseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get pseudonym
				pseudonym := &dbmodels.Pseudonym{
					PseudonymID: "test-pseudonym",
					DisplayName: "Test User",
					CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
				}
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(pseudonym, nil)

				// Mock get user ID from pseudonym
				pseudonymDAO.EXPECT().
					GetUserIDByPseudonym(gomock.Any(), "test-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).
					Return(int64(123), nil)

				// Mock get user info for real identity
				userInfo := &dbmodels.User{
					UserID: 123,
					Email:  "testuser@example.com",
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(userInfo, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "unauthorized - no token",
			input: &GetPseudonymInput{
				AuthInput:   middleware.AuthInput{},
				PseudonymID: "test-pseudonym",
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// No mocks needed - should fail at authentication
			},
			expectedStatus: 401,
			expectedError:  "authentication required",
		},
		{
			name: "forbidden - insufficient permissions",
			input: &GetPseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "user-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user lacks system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "user-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(false, nil)
			},
			expectedStatus: 403,
			expectedError:  "insufficient permissions: platform admin capability required",
		},
		{
			name: "pseudonym not found",
			input: &GetPseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "nonexistent-pseudonym",
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock pseudonym not found
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "nonexistent-pseudonym").
					Return(nil, nil)
			},
			expectedStatus: 404,
			expectedError:  "pseudonym not found",
		},
		{
			name: "database error",
			input: &GetPseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock database error
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get pseudonym: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
			mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
			mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
			mockPasswordResetTokenDAO := dao.NewMockPasswordResetTokenDAOInterface(ctrl)
			mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
			mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
			mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
			mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

			// Setup mocks
			tt.setupMocks(mockUserDAO, mockPseudonymDAO, mockPermissionDAO)

			// Add additional mocks for GetPseudonym
			if tt.name == "successful pseudonym get" {
				mockPostDAO.EXPECT().CountPostsByPseudonym(gomock.Any(), "test-pseudonym").Return(int64(5), nil)
				mockCommentDAO.EXPECT().CountCommentsByPseudonym(gomock.Any(), "test-pseudonym").Return(int64(10), nil)
				mockRoleKeyDAO.EXPECT().ListRoleKeysByPseudonym(gomock.Any(), "test-pseudonym").Return([]*dbmodels.RoleKey{}, nil)
			}

			// Create handler
			handler := NewAdminHandler(
				mockUserDAO,
				mockPseudonymDAO,
				mockPermissionDAO,
				mockPasswordResetTokenDAO,
				nil, // emailService
				nil, // config
				mockPostDAO,
				mockCommentDAO,
				mockRoleKeyDAO,
				mockSubforumDAO,
			)

			// Call method
			response, err := handler.GetPseudonym(context.Background(), tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)

			if tt.expectedStatus == 200 {
				assert.NotNil(t, response.Body)
				assert.Equal(t, "test-pseudonym", response.Body.PseudonymID)
				assert.Equal(t, "Test User", response.Body.DisplayName)
			}
		})
	}
}

func TestAdminHandler_UpdatePseudonym(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *UpdatePseudonymInput
		setupMocks     func(*dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockPermissionDAOInterface)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful pseudonym update - name only",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
				Body: UpdatePseudonymBody{
					DisplayName: stringPtr("Updated Name"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing pseudonym
				existingPseudonym := &dbmodels.Pseudonym{
					PseudonymID: "test-pseudonym",
					DisplayName: "Original Name",
					CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
				}
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(existingPseudonym, nil)

				// Mock check if display name is already taken
				pseudonymDAO.EXPECT().
					GetPseudonymByDisplayName(gomock.Any(), "Updated Name").
					Return(nil, nil)

				// Mock get user ID from pseudonym
				pseudonymDAO.EXPECT().
					GetUserIDByPseudonym(gomock.Any(), "test-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).
					Return(int64(123), nil)

				// Mock get user info for real identity
				userInfo := &dbmodels.User{
					UserID: 123,
					Email:  "testuser@example.com",
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(userInfo, nil)

				// Mock update pseudonym
				pseudonymDAO.EXPECT().
					UpdatePseudonym(gomock.Any(), "test-pseudonym", gomock.Any()).
					Return(nil)
			},
			expectedStatus: 200,
		},
		{
			name: "successful pseudonym update - active status only",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
				Body: UpdatePseudonymBody{
					IsActive: boolPtr(false),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing pseudonym
				existingPseudonym := &dbmodels.Pseudonym{
					PseudonymID: "test-pseudonym",
					DisplayName: "Test Name",
					IsActive:    sql.Null[bool]{V: true, Valid: true},
					CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
				}
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(existingPseudonym, nil)

				// Mock get user ID from pseudonym
				pseudonymDAO.EXPECT().
					GetUserIDByPseudonym(gomock.Any(), "test-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).
					Return(int64(123), nil)

				// Mock get user info for real identity
				userInfo := &dbmodels.User{
					UserID: 123,
					Email:  "testuser@example.com",
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(userInfo, nil)

				// Mock update pseudonym
				pseudonymDAO.EXPECT().
					UpdatePseudonym(gomock.Any(), "test-pseudonym", gomock.Any()).
					Return(nil)
			},
			expectedStatus: 200,
		},
		{
			name: "successful pseudonym update - both name and active",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
				Body: UpdatePseudonymBody{
					DisplayName: stringPtr("Updated Name"),
					IsActive:    boolPtr(false),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing pseudonym
				existingPseudonym := &dbmodels.Pseudonym{
					PseudonymID: "test-pseudonym",
					DisplayName: "Original Name",
					IsActive:    sql.Null[bool]{V: true, Valid: true},
					CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
				}
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(existingPseudonym, nil)

				// Mock check if display name is already taken
				pseudonymDAO.EXPECT().
					GetPseudonymByDisplayName(gomock.Any(), "Updated Name").
					Return(nil, nil)

				// Mock get user ID from pseudonym
				pseudonymDAO.EXPECT().
					GetUserIDByPseudonym(gomock.Any(), "test-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).
					Return(int64(123), nil)

				// Mock get user info for real identity
				userInfo := &dbmodels.User{
					UserID: 123,
					Email:  "testuser@example.com",
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(userInfo, nil)

				// Mock update pseudonym
				pseudonymDAO.EXPECT().
					UpdatePseudonym(gomock.Any(), "test-pseudonym", gomock.Any()).
					Return(nil)
			},
			expectedStatus: 200,
		},
		{
			name: "unauthorized - no token",
			input: &UpdatePseudonymInput{
				AuthInput:   middleware.AuthInput{},
				PseudonymID: "test-pseudonym",
				Body: UpdatePseudonymBody{
					DisplayName: stringPtr("Updated Name"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// No mocks needed - should fail at authentication
			},
			expectedStatus: 401,
			expectedError:  "authentication required",
		},
		{
			name: "forbidden - insufficient permissions",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "user-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
				Body: UpdatePseudonymBody{
					DisplayName: stringPtr("Updated Name"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check - user lacks system admin capability
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "user-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(false, nil)
			},
			expectedStatus: 403,
			expectedError:  "insufficient permissions: platform admin capability required",
		},
		{
			name: "pseudonym not found",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "nonexistent-pseudonym",
				Body: UpdatePseudonymBody{
					DisplayName: stringPtr("Updated Name"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock pseudonym not found
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "nonexistent-pseudonym").
					Return(nil, nil)
			},
			expectedStatus: 404,
			expectedError:  "pseudonym not found",
		},
		{
			name: "database error on get",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
				Body: UpdatePseudonymBody{
					DisplayName: stringPtr("Updated Name"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock database error on get
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get pseudonym: database error",
		},
		{
			name: "database error on update",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
				Body: UpdatePseudonymBody{
					DisplayName: stringPtr("Updated Name"),
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing pseudonym
				existingPseudonym := &dbmodels.Pseudonym{
					PseudonymID: "test-pseudonym",
					DisplayName: "Original Name",
					CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
				}
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(existingPseudonym, nil)

				// Mock check if display name is already taken
				pseudonymDAO.EXPECT().
					GetPseudonymByDisplayName(gomock.Any(), "Updated Name").
					Return(nil, nil)

				// Mock update error - handler returns early, so GetUserIDByPseudonym won't be called
				pseudonymDAO.EXPECT().
					UpdatePseudonym(gomock.Any(), "test-pseudonym", gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: "failed to update pseudonym: database error",
		},
		{
			name: "no changes provided",
			input: &UpdatePseudonymInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "admin-pseudonym"),
				},
				PseudonymID: "test-pseudonym",
				Body:        UpdatePseudonymBody{
					// No fields provided
				},
			},
			setupMocks: func(userDAO *dao.MockUserDAOInterface, pseudonymDAO *dao.MockPseudonymDAOInterface, permissionDAO *dao.MockPermissionDAOInterface) {
				// Mock permission check
				permissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(1), "admin-pseudonym", constants.CapabilitySystemAdmin, nil).
					Return(true, nil)

				// Mock get existing pseudonym
				existingPseudonym := &dbmodels.Pseudonym{
					PseudonymID: "test-pseudonym",
					DisplayName: "Original Name",
					CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
				}
				pseudonymDAO.EXPECT().
					GetPseudonymByID(gomock.Any(), "test-pseudonym").
					Return(existingPseudonym, nil)

				// Mock get user ID from pseudonym
				pseudonymDAO.EXPECT().
					GetUserIDByPseudonym(gomock.Any(), "test-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).
					Return(int64(123), nil)

				// Mock get user info for real identity
				userInfo := &dbmodels.User{
					UserID: 123,
					Email:  "testuser@example.com",
				}
				userDAO.EXPECT().
					GetUserByID(gomock.Any(), int64(123)).
					Return(userInfo, nil)

				// No update call expected since no changes
			},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
			mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
			mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
			mockPasswordResetTokenDAO := dao.NewMockPasswordResetTokenDAOInterface(ctrl)
			mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
			mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
			mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
			mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

			// Setup mocks
			tt.setupMocks(mockUserDAO, mockPseudonymDAO, mockPermissionDAO)

			// Add additional mocks for UpdatePseudonym - it calls GetPseudonymByID twice
			if tt.name == "successful pseudonym update - name only" ||
				tt.name == "successful pseudonym update - active status only" ||
				tt.name == "successful pseudonym update - both name and active" ||
				tt.name == "no changes provided" ||
				tt.name == "database error on update" {
				// Second call to GetPseudonymByID at the end of the handler
				mockPseudonymDAO.EXPECT().GetPseudonymByID(gomock.Any(), "test-pseudonym").Return(&dbmodels.Pseudonym{
					PseudonymID:         "test-pseudonym",
					DisplayName:         "UpdatedUser",
					Slug:                sql.Null[string]{V: "updateduser", Valid: true},
					IsActive:            sql.Null[bool]{V: true, Valid: true},
					Bio:                 sql.Null[string]{V: "Updated bio", Valid: true},
					WebsiteURL:          sql.Null[string]{V: "https://updated.com", Valid: true},
					ShowKarma:           sql.Null[bool]{V: true, Valid: true},
					AllowDirectMessages: sql.Null[bool]{V: true, Valid: true},
					CreatedAt:           sql.Null[time.Time]{V: time.Now(), Valid: true},
				}, nil).AnyTimes()

				// Second call to GetUserIDByPseudonym at the end
				mockPseudonymDAO.EXPECT().GetUserIDByPseudonym(gomock.Any(), "test-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).Return(int64(123), nil).AnyTimes()

				// Second call to GetUserByID at the end
				mockUserDAO.EXPECT().GetUserByID(gomock.Any(), int64(123)).Return(&dbmodels.User{
					UserID: 123,
					Email:  "testuser@example.com",
				}, nil).AnyTimes()

				// Additional mocks for post/comment counts and role keys
				mockPostDAO.EXPECT().CountPostsByPseudonym(gomock.Any(), "test-pseudonym").Return(int64(5), nil).AnyTimes()
				mockCommentDAO.EXPECT().CountCommentsByPseudonym(gomock.Any(), "test-pseudonym").Return(int64(10), nil).AnyTimes()
				mockRoleKeyDAO.EXPECT().ListRoleKeysByPseudonym(gomock.Any(), "test-pseudonym").Return([]*dbmodels.RoleKey{}, nil).AnyTimes()
			}

			// Create handler
			handler := NewAdminHandler(
				mockUserDAO,
				mockPseudonymDAO,
				mockPermissionDAO,
				mockPasswordResetTokenDAO,
				nil, // emailService
				nil, // config
				mockPostDAO,
				mockCommentDAO,
				mockRoleKeyDAO,
				mockSubforumDAO,
			)

			// Call method
			response, err := handler.UpdatePseudonym(context.Background(), tt.input)

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)

			if tt.expectedStatus == 200 {
				assert.NotNil(t, response.Body)
				assert.Equal(t, "test-pseudonym", response.Body.PseudonymID)
				assert.Equal(t, "UpdatedUser", response.Body.DisplayName)
			}
		})
	}
}

// Type alias for the UpdatePseudonym body to avoid struct literal issues
type UpdatePseudonymBody = struct {
	DisplayName         *string          `json:"display_name,omitempty" example:"JohnDoe"`
	Slug                *string          `json:"slug,omitempty" example:"john-doe"`
	IsActive            *bool            `json:"is_active,omitempty" example:"true"`
	Bio                 *string          `json:"bio,omitempty" example:"A brief description of the user."`
	WebsiteURL          *string          `json:"website_url,omitempty" example:"https://example.com"`
	ShowKarma           *bool            `json:"show_karma,omitempty" example:"true"`
	AllowDirectMessages *bool            `json:"allow_direct_messages,omitempty" example:"true"`
	RoleKeys            *[]RoleKeyUpdate `json:"role_keys,omitempty"`
}

// Helper functions for test data
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
