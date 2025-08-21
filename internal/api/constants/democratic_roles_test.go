package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDemocraticRoles_Constants(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{
			name:     "elected moderator role constant",
			role:     RoleElectedModerator,
			expected: "elected_moderator",
		},
		{
			name:     "appointed moderator role constant",
			role:     RoleAppointedModerator,
			expected: "appointed_moderator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role)
		})
	}
}

func TestDemocraticRoles_ValidRoles(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		isValid bool
	}{
		{
			name:    "elected moderator is valid",
			role:    RoleElectedModerator,
			isValid: true,
		},
		{
			name:    "appointed moderator is valid",
			role:    RoleAppointedModerator,
			isValid: true,
		},
		{
			name:    "invalid role is not valid",
			role:    "invalid_role",
			isValid: false,
		},
		{
			name:    "empty role is not valid",
			role:    "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRole(tt.role)
			assert.Equal(t, tt.isValid, result)
		})
	}
}

func TestDemocraticRoles_RoleDefinitions(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		expectNil      bool
		expectedScopes []string
	}{
		{
			name:      "elected moderator has definition",
			role:      RoleElectedModerator,
			expectNil: false,
			expectedScopes: []string{
				ScopeAuthentication,
				ScopeSelfCorrelation,
				ScopeModeration,
				ScopeCorrelation,
			},
		},
		{
			name:      "appointed moderator has definition",
			role:      RoleAppointedModerator,
			expectNil: false,
			expectedScopes: []string{
				ScopeAuthentication,
				ScopeSelfCorrelation,
				ScopeModeration,
				ScopeCorrelation,
			},
		},
		{
			name:      "invalid role returns nil",
			role:      "invalid_role",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition := GetRoleDefinition(tt.role)

			if tt.expectNil {
				assert.Nil(t, definition)
			} else {
				require.NotNil(t, definition)
				assert.Equal(t, tt.role, definition.RoleName)
				assert.Equal(t, tt.expectedScopes, definition.Scopes)
			}
		})
	}
}

func TestDemocraticRoles_Capabilities(t *testing.T) {
	tests := []struct {
		name                 string
		role                 string
		expectedCapabilities []string
	}{
		{
			name: "elected moderator has full moderation capabilities",
			role: RoleElectedModerator,
			expectedCapabilities: []string{
				// Authentication scope
				CapabilityAccessOwnPseudonyms,
				CapabilityLogin,
				CapabilitySessionManagement,
				// Self correlation scope
				CapabilityVerifyOwnPseudonymOwnership,
				CapabilityManageOwnProfile,
				// Moderation scope (same as subforum_owner)
				CapabilityModerateContent,
				CapabilityBanUsers,
				CapabilityRemoveContent,
				CapabilityManageModerators,
				CapabilityReviewReports,
				CapabilityForwardReports,
				CapabilityManageSubforumRules,
				CapabilityManageSubforumSettings,
				CapabilityStickyPost,
				CapabilityLockPost,
				// Correlation scope
				CapabilityAccessSubforumPseudonyms,
				CapabilityCorrelateFingerprints,
			},
		},
		{
			name: "appointed moderator has same capabilities as elected",
			role: RoleAppointedModerator,
			expectedCapabilities: []string{
				// Should be identical to elected_moderator
				CapabilityAccessOwnPseudonyms,
				CapabilityLogin,
				CapabilitySessionManagement,
				CapabilityVerifyOwnPseudonymOwnership,
				CapabilityManageOwnProfile,
				CapabilityModerateContent,
				CapabilityBanUsers,
				CapabilityRemoveContent,
				CapabilityManageModerators,
				CapabilityReviewReports,
				CapabilityForwardReports,
				CapabilityManageSubforumRules,
				CapabilityManageSubforumSettings,
				CapabilityStickyPost,
				CapabilityLockPost,
				CapabilityAccessSubforumPseudonyms,
				CapabilityCorrelateFingerprints,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capabilities := GetRoleCapabilities(tt.role)

			// Check that all expected capabilities are present
			for _, expectedCap := range tt.expectedCapabilities {
				assert.Contains(t, capabilities, expectedCap,
					"Role %s should have capability %s", tt.role, expectedCap)
			}

			// Check that we have the expected number of capabilities
			assert.Len(t, capabilities, len(tt.expectedCapabilities),
				"Role %s should have exactly %d capabilities", tt.role, len(tt.expectedCapabilities))
		})
	}
}

func TestDemocraticRoles_CapabilityParity(t *testing.T) {
	// Ensure elected and appointed moderators have the same capabilities as subforum_owner
	subforumOwnerCaps := GetRoleCapabilities(RoleSubforumOwner)
	electedModCaps := GetRoleCapabilities(RoleElectedModerator)
	appointedModCaps := GetRoleCapabilities(RoleAppointedModerator)

	t.Run("elected moderator has same capabilities as subforum owner", func(t *testing.T) {
		assert.ElementsMatch(t, subforumOwnerCaps, electedModCaps,
			"elected_moderator should have identical capabilities to subforum_owner")
	})

	t.Run("appointed moderator has same capabilities as subforum owner", func(t *testing.T) {
		assert.ElementsMatch(t, subforumOwnerCaps, appointedModCaps,
			"appointed_moderator should have identical capabilities to subforum_owner")
	})

	t.Run("elected and appointed moderators have identical capabilities", func(t *testing.T) {
		assert.ElementsMatch(t, electedModCaps, appointedModCaps,
			"elected_moderator and appointed_moderator should have identical capabilities")
	})
}

func TestDemocraticRoles_ScopeParity(t *testing.T) {
	// Ensure elected and appointed moderators have the same scopes as subforum_owner
	subforumOwnerScopes := GetRoleScopes(RoleSubforumOwner)
	electedModScopes := GetRoleScopes(RoleElectedModerator)
	appointedModScopes := GetRoleScopes(RoleAppointedModerator)

	t.Run("elected moderator has same scopes as subforum owner", func(t *testing.T) {
		assert.ElementsMatch(t, subforumOwnerScopes, electedModScopes,
			"elected_moderator should have identical scopes to subforum_owner")
	})

	t.Run("appointed moderator has same scopes as subforum owner", func(t *testing.T) {
		assert.ElementsMatch(t, subforumOwnerScopes, appointedModScopes,
			"appointed_moderator should have identical scopes to subforum_owner")
	})
}

func TestDemocraticRoles_GetAllRoles(t *testing.T) {
	allRoles := GetAllRoles()

	t.Run("includes new democratic roles", func(t *testing.T) {
		assert.Contains(t, allRoles, RoleElectedModerator,
			"GetAllRoles should include elected_moderator")
		assert.Contains(t, allRoles, RoleAppointedModerator,
			"GetAllRoles should include appointed_moderator")
	})

	t.Run("includes existing roles", func(t *testing.T) {
		expectedRoles := []string{
			RoleUser,
			RoleModerator,
			RoleSubforumOwner,
			RoleElectedModerator,
			RoleAppointedModerator,
			RolePlatformAdmin,
			RoleTrustSafety,
			RoleLegalTeam,
		}

		for _, expectedRole := range expectedRoles {
			assert.Contains(t, allRoles, expectedRole,
				"GetAllRoles should include %s", expectedRole)
		}

		assert.Len(t, allRoles, len(expectedRoles),
			"GetAllRoles should return exactly %d roles", len(expectedRoles))
	})
}
