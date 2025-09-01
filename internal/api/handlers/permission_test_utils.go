package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/database/dao"
)

// PermissionTestScenario defines a complete permission testing scenario
type PermissionTestScenario struct {
	Name               string
	UserID             int64
	ActivePseudonymID  string
	SubforumID         *int32
	RequiredCapability string
	Roles              []string
	Capabilities       []string
	ShouldHaveAccess   bool
	ErrorExpected      bool
	Description        string
}

// PermissionTestSuite provides comprehensive permission testing utilities
type PermissionTestSuite struct {
	MockPermissionDAO dao.PermissionDAOInterface
	scenarios         []PermissionTestScenario
}

// NewPermissionTestSuite creates a new permission test suite
func NewPermissionTestSuite(ctrl *gomock.Controller) *PermissionTestSuite {
	return &PermissionTestSuite{
		MockPermissionDAO: dao.NewMockPermissionDAOInterface(ctrl),
		scenarios:         GetStandardPermissionScenarios(),
	}
}

// SetupMockExpectations configures all mock expectations for the scenarios
func (pts *PermissionTestSuite) SetupMockExpectations() {
	// Set up expectations for HasUnifiedCapability
	pts.MockPermissionDAO.(*dao.MockPermissionDAOInterface).EXPECT().
		HasUnifiedCapability(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID int64, activePseudonymID string, capability string, subforumID *int32) (bool, error) {
			// Find matching scenario
			for _, scenario := range pts.scenarios {
				if scenario.UserID == userID &&
					scenario.ActivePseudonymID == activePseudonymID &&
					scenario.RequiredCapability == capability &&
					((scenario.SubforumID == nil && subforumID == nil) ||
						(scenario.SubforumID != nil && subforumID != nil && *scenario.SubforumID == *subforumID)) {

					if scenario.ErrorExpected {
						return false, errors.New("test error")
					}
					return scenario.ShouldHaveAccess, nil
				}
			}
			// Default: no access
			return false, nil
		}).AnyTimes()

	// Set up expectations for GetUnifiedActivePseudonymRolesAndCapabilities
	pts.MockPermissionDAO.(*dao.MockPermissionDAOInterface).EXPECT().
		GetUnifiedActivePseudonymRolesAndCapabilities(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID int64, activePseudonymID string, subforumID *int32) ([]string, []string, error) {
			// Find matching scenario
			for _, scenario := range pts.scenarios {
				if scenario.UserID == userID &&
					scenario.ActivePseudonymID == activePseudonymID &&
					((scenario.SubforumID == nil && subforumID == nil) ||
						(scenario.SubforumID != nil && subforumID != nil && *scenario.SubforumID == *subforumID)) {

					if scenario.ErrorExpected {
						return nil, nil, errors.New("test error")
					}
					return scenario.Roles, scenario.Capabilities, nil
				}
			}
			// Default: basic user
			return []string{constants.RoleUser}, []string{constants.CapabilityCreateContent, constants.CapabilityVote}, nil
		}).AnyTimes()
}

// AddScenario adds a custom scenario to the test suite
func (pts *PermissionTestSuite) AddScenario(scenario PermissionTestScenario) {
	pts.scenarios = append(pts.scenarios, scenario)
}

// TestScenario tests a specific permission scenario
func (pts *PermissionTestSuite) TestScenario(t *testing.T, scenario PermissionTestScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		ctx := context.Background()

		hasAccess, err := pts.MockPermissionDAO.HasUnifiedCapability(
			ctx,
			scenario.UserID,
			scenario.ActivePseudonymID,
			scenario.RequiredCapability,
			scenario.SubforumID,
		)

		if scenario.ErrorExpected {
			assert.Error(t, err, "Expected error for scenario: %s", scenario.Description)
		} else {
			assert.NoError(t, err, "Unexpected error for scenario: %s", scenario.Description)
			assert.Equal(t, scenario.ShouldHaveAccess, hasAccess,
				"Permission mismatch for scenario: %s", scenario.Description)
		}

		// Also test role and capability retrieval
		if !scenario.ErrorExpected {
			roles, capabilities, err := pts.MockPermissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
				ctx,
				scenario.UserID,
				scenario.ActivePseudonymID,
				scenario.SubforumID,
			)

			assert.NoError(t, err, "Unexpected error getting roles/capabilities for scenario: %s", scenario.Description)
			assert.Equal(t, scenario.Roles, roles, "Role mismatch for scenario: %s", scenario.Description)
			assert.Equal(t, scenario.Capabilities, capabilities, "Capability mismatch for scenario: %s", scenario.Description)
		}
	})
}

// TestAllScenarios runs all configured scenarios
func (pts *PermissionTestSuite) TestAllScenarios(t *testing.T) {
	pts.SetupMockExpectations()

	for _, scenario := range pts.scenarios {
		pts.TestScenario(t, scenario)
	}
}

// GetStandardPermissionScenarios returns comprehensive test scenarios covering all user types and permissions
func GetStandardPermissionScenarios() []PermissionTestScenario {
	subforumID := int32(1)

	return []PermissionTestScenario{
		// Regular User Scenarios
		{
			Name:               "Regular_User_Create_Content_Global",
			UserID:             1,
			ActivePseudonymID:  "user-pseudo-1",
			SubforumID:         nil,
			RequiredCapability: constants.CapabilityCreateContent,
			Roles:              []string{constants.RoleUser},
			Capabilities:       []string{constants.CapabilityCreateContent, constants.CapabilityVote, constants.CapabilityMessage, constants.CapabilityReport},
			ShouldHaveAccess:   true,
			ErrorExpected:      false,
			Description:        "Regular user should be able to create content globally",
		},
		{
			Name:               "Regular_User_Cannot_Moderate_Global",
			UserID:             1,
			ActivePseudonymID:  "user-pseudo-1",
			SubforumID:         nil,
			RequiredCapability: constants.CapabilityModerateContent,
			Roles:              []string{constants.RoleUser},
			Capabilities:       []string{constants.CapabilityCreateContent, constants.CapabilityVote, constants.CapabilityMessage, constants.CapabilityReport},
			ShouldHaveAccess:   false,
			ErrorExpected:      false,
			Description:        "Regular user should not be able to moderate content globally",
		},
		{
			Name:               "Subforum_Moderator_Can_Moderate_In_Subforum",
			UserID:             2,
			ActivePseudonymID:  "mod-pseudo-2",
			SubforumID:         &subforumID,
			RequiredCapability: constants.CapabilityModerateContent,
			Roles:              []string{constants.RoleUser, constants.RoleModerator},
			Capabilities:       []string{constants.CapabilityCreateContent, constants.CapabilityVote, constants.CapabilityMessage, constants.CapabilityReport, constants.CapabilityModerateContent, constants.CapabilityBanUsers, constants.CapabilityRemoveContent},
			ShouldHaveAccess:   true,
			ErrorExpected:      false,
			Description:        "Subforum moderator should be able to moderate content in their subforum",
		},
		{
			Name:               "Subforum_Owner_Can_Manage_Moderators",
			UserID:             3,
			ActivePseudonymID:  "owner-pseudo-3",
			SubforumID:         &subforumID,
			RequiredCapability: constants.CapabilityManageModerators,
			Roles:              []string{constants.RoleUser, constants.RoleSubforumOwner},
			Capabilities:       []string{constants.CapabilityCreateContent, constants.CapabilityVote, constants.CapabilityMessage, constants.CapabilityReport, constants.CapabilityModerateContent, constants.CapabilityBanUsers, constants.CapabilityRemoveContent, constants.CapabilityManageModerators, constants.CapabilityManageSubforumSettings},
			ShouldHaveAccess:   true,
			ErrorExpected:      false,
			Description:        "Subforum owner should be able to manage moderators in their subforum",
		},
		{
			Name:               "Platform_Admin_System_Admin_Global",
			UserID:             4,
			ActivePseudonymID:  "admin-pseudo-4",
			SubforumID:         nil,
			RequiredCapability: constants.CapabilitySystemAdmin,
			Roles:              []string{constants.RoleUser, constants.RolePlatformAdmin},
			Capabilities:       []string{constants.CapabilityCreateContent, constants.CapabilityVote, constants.CapabilityMessage, constants.CapabilityReport, constants.CapabilitySystemAdmin, constants.CapabilityUserManagement, constants.CapabilityModeration, constants.CapabilityCompliance},
			ShouldHaveAccess:   true,
			ErrorExpected:      false,
			Description:        "Platform admin should have system admin capabilities globally",
		},
	}
}

// CreateTestUserContext creates a UserContext for testing with the given parameters
func CreateTestUserContext(userID int64, pseudonymID string, email string, displayName string) *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            userID,
		Email:             email,
		ActivePseudonymID: pseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false,
	}
}

// CreateTestUserContextFromScenario creates a UserContext from a permission scenario
func CreateTestUserContextFromScenario(scenario PermissionTestScenario) *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            scenario.UserID,
		Email:             "test@example.com",
		ActivePseudonymID: scenario.ActivePseudonymID,
		DisplayName:       "Test User",
		MFAEnabled:        false,
	}
}
