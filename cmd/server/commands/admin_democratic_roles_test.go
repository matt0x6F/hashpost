package commands

import (
	"testing"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/stretchr/testify/assert"
)

func TestIsAdminRole_DemocraticRoles(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		// Positive cases - should be admin roles
		{
			name:     "moderator is admin role",
			role:     constants.RoleModerator,
			expected: true,
		},
		{
			name:     "subforum owner is admin role",
			role:     constants.RoleSubforumOwner,
			expected: true,
		},
		{
			name:     "elected moderator is admin role",
			role:     constants.RoleElectedModerator,
			expected: true,
		},
		{
			name:     "appointed moderator is admin role",
			role:     constants.RoleAppointedModerator,
			expected: true,
		},
		{
			name:     "platform admin is admin role",
			role:     constants.RolePlatformAdmin,
			expected: true,
		},
		// Negative cases - should not be admin roles
		{
			name:     "user is not admin role",
			role:     constants.RoleUser,
			expected: false,
		},
		{
			name:     "trust safety is not admin role",
			role:     constants.RoleTrustSafety,
			expected: false,
		},
		{
			name:     "legal team is not admin role",
			role:     constants.RoleLegalTeam,
			expected: false,
		},
		{
			name:     "invalid role is not admin role",
			role:     "invalid_role",
			expected: false,
		},
		{
			name:     "empty role is not admin role",
			role:     "",
			expected: false,
		},
		{
			name:     "random string is not admin role",
			role:     "random_string",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAdminRole(tt.role)
			assert.Equal(t, tt.expected, result,
				"isAdminRole(%s) should return %t", tt.role, tt.expected)
		})
	}
}

func TestIsAdminRole_AllModeratorTypes(t *testing.T) {
	// Test that all moderator-level roles are considered admin roles
	moderatorRoles := []string{
		constants.RoleModerator,
		constants.RoleSubforumOwner,
		constants.RoleElectedModerator,
		constants.RoleAppointedModerator,
	}

	for _, role := range moderatorRoles {
		t.Run("moderator role "+role+" is admin", func(t *testing.T) {
			assert.True(t, isAdminRole(role),
				"All moderator-level roles should be considered admin roles: %s", role)
		})
	}
}

func TestIsAdminRole_NonModeratorTypes(t *testing.T) {
	// Test that non-moderator roles are handled correctly
	nonModeratorRoles := []string{
		constants.RoleUser,
		constants.RoleTrustSafety,
		constants.RoleLegalTeam,
	}

	for _, role := range nonModeratorRoles {
		t.Run("non-moderator role "+role, func(t *testing.T) {
			// Only platform_admin should be true among non-moderator roles
			expected := role == constants.RolePlatformAdmin
			assert.Equal(t, expected, isAdminRole(role),
				"Role %s admin status should be %t", role, expected)
		})
	}
}
