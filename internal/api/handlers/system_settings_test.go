package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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
)

// TestNewSystemSettingsHandler tests the system settings handler constructor
func TestNewSystemSettingsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSystemSettingsDAO := dao.NewMockSystemSettingsDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	handler := handlers.NewSystemSettingsHandler(
		mockSystemSettingsDAO,
		mockPermissionDAO,
		mockPseudonymDAO,
		ibeSystem,
	)

	assert.NotNil(t, handler)
}

// TestSystemSettingsHandler_GetSystemSetting tests the GetSystemSetting method
func TestSystemSettingsHandler_GetSystemSetting(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSystemSettingsDAO := dao.NewMockSystemSettingsDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	handler := handlers.NewSystemSettingsHandler(
		mockSystemSettingsDAO,
		mockPermissionDAO,
		mockPseudonymDAO,
		ibeSystem,
	)

	tests := []struct {
		name           string
		input          *models.SystemSettingGetInput
		setupMocks     func()
		expectedError  bool
		expectedResult *models.SystemSettingResponse
	}{
		{
			name: "Success - Valid system setting retrieval",
			input: &models.SystemSettingGetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				SettingKey: "site_name",
			},
			setupMocks: func() {
				// Mock permission check
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(true, nil)

				// Mock system setting retrieval
				expectedSetting := &dbmodels.SystemSetting{
					SettingKey:   "site_name",
					SettingValue: "HashPost",
					SettingType:  "string",
					UpdatedAt:    sql.Null[time.Time]{Valid: true, V: time.Now()},
				}
				mockSystemSettingsDAO.EXPECT().
					GetSetting(gomock.Any(), "site_name").
					Return(expectedSetting, nil)
			},
			expectedError: false,
			expectedResult: &models.SystemSettingResponse{
				Status: 200,
				Body: models.SystemSettingResponseBody{
					SettingKey:   "site_name",
					SettingValue: "HashPost",
					SettingType:  "string",
					Message:      "Setting retrieved successfully",
				},
			},
		},
		{
			name: "Error - Authentication required",
			input: &models.SystemSettingGetInput{
				SettingKey: "site_name",
			},
			setupMocks:     func() {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Insufficient permissions",
			input: &models.SystemSettingGetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				SettingKey: "site_name",
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(false, nil)
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Setting not found",
			input: &models.SystemSettingGetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				SettingKey: "nonexistent_setting",
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(true, nil)

				mockSystemSettingsDAO.EXPECT().
					GetSetting(gomock.Any(), "nonexistent_setting").
					Return(nil, nil)
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Database failure",
			input: &models.SystemSettingGetInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				SettingKey: "site_name",
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(true, nil)

				mockSystemSettingsDAO.EXPECT().
					GetSetting(gomock.Any(), "site_name").
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
			result, err := handler.GetSystemSetting(context.Background(), tt.input)

			// Assertions
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.Equal(t, tt.expectedResult.Body.SettingKey, result.Body.SettingKey)
				assert.Equal(t, tt.expectedResult.Body.SettingValue, result.Body.SettingValue)
				assert.Equal(t, tt.expectedResult.Body.SettingType, result.Body.SettingType)
			}
		})
	}
}

// TestSystemSettingsHandler_UpdateSystemSetting tests the UpdateSystemSetting method
func TestSystemSettingsHandler_UpdateSystemSetting(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSystemSettingsDAO := dao.NewMockSystemSettingsDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	handler := handlers.NewSystemSettingsHandler(
		mockSystemSettingsDAO,
		mockPermissionDAO,
		mockPseudonymDAO,
		ibeSystem,
	)

	tests := []struct {
		name           string
		input          *models.SystemSettingUpdateInput
		setupMocks     func()
		expectedError  bool
		expectedResult *models.SystemSettingResponse
	}{
		{
			name: "Success - Valid system setting update",
			input: &models.SystemSettingUpdateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				Body: models.SystemSettingUpdateInputBody{
					SettingKey:   "site_name",
					SettingValue: "NewHashPost",
					SettingType:  "string",
				},
			},
			setupMocks: func() {
				// Mock permission check
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(true, nil)

				// Mock system setting update
				mockSystemSettingsDAO.EXPECT().
					SetSetting(gomock.Any(), "site_name", "NewHashPost", "string", int64(123)).
					Return(nil)

				// Mock pseudonym update
				mockPseudonymDAO.EXPECT().
					UpdateLastActive(gomock.Any(), "admin-pseudonym-456").
					Return(nil)
			},
			expectedError: false,
			expectedResult: &models.SystemSettingResponse{
				Status: 200,
				Body: models.SystemSettingResponseBody{
					SettingKey:   "site_name",
					SettingValue: "NewHashPost",
					SettingType:  "string",
					Message:      "Setting updated successfully",
				},
			},
		},
		{
			name: "Error - Authentication required",
			input: &models.SystemSettingUpdateInput{
				Body: models.SystemSettingUpdateInputBody{
					SettingKey:   "site_name",
					SettingValue: "NewHashPost",
					SettingType:  "string",
				},
			},
			setupMocks:     func() {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Insufficient permissions",
			input: &models.SystemSettingUpdateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				Body: models.SystemSettingUpdateInputBody{
					SettingKey:   "site_name",
					SettingValue: "NewHashPost",
					SettingType:  "string",
				},
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(false, nil)
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Database failure on setting update",
			input: &models.SystemSettingUpdateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				Body: models.SystemSettingUpdateInputBody{
					SettingKey:   "site_name",
					SettingValue: "NewHashPost",
					SettingType:  "string",
				},
			},
			setupMocks: func() {
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(true, nil)

				mockSystemSettingsDAO.EXPECT().
					SetSetting(gomock.Any(), "site_name", "NewHashPost", "string", int64(123)).
					Return(errors.New("database connection failed"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name: "Error - Pseudonym update failure (non-critical)",
			input: &models.SystemSettingUpdateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(123, "admin-pseudonym-456"),
				},
				Body: models.SystemSettingUpdateInputBody{
					SettingKey:   "site_name",
					SettingValue: "NewHashPost",
					SettingType:  "string",
				},
			},
			setupMocks: func() {
				// Mock permission check
				mockPermissionDAO.EXPECT().
					HasUnifiedCapability(gomock.Any(), int64(123), "admin-pseudonym-456", "system_admin", nil).
					Return(true, nil)

				// Mock successful setting update
				mockSystemSettingsDAO.EXPECT().
					SetSetting(gomock.Any(), "site_name", "NewHashPost", "string", int64(123)).
					Return(nil)

				// Mock failed pseudonym update (should not fail the request)
				mockPseudonymDAO.EXPECT().
					UpdateLastActive(gomock.Any(), "admin-pseudonym-456").
					Return(errors.New("pseudonym update failed"))
			},
			expectedError: false,
			expectedResult: &models.SystemSettingResponse{
				Status: 200,
				Body: models.SystemSettingResponseBody{
					SettingKey:   "site_name",
					SettingValue: "NewHashPost",
					SettingType:  "string",
					Message:      "Setting updated successfully",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test case
			tt.setupMocks()

			// Execute the method
			result, err := handler.UpdateSystemSetting(context.Background(), tt.input)

			// Assertions
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.Equal(t, tt.expectedResult.Body.SettingKey, result.Body.SettingKey)
				assert.Equal(t, tt.expectedResult.Body.SettingValue, result.Body.SettingValue)
				assert.Equal(t, tt.expectedResult.Body.SettingType, result.Body.SettingType)
				assert.Equal(t, tt.expectedResult.Body.Message, result.Body.Message)
			}
		})
	}
}
