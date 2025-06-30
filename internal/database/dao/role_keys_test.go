//go:build integration

package dao

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureDefaultKeys_RegularUser(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user with "user" role
	testUser := suite.CreateTestUser(t, "testuser@example.com", "password123", []string{"user"})

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, testUser.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 2, "Should have 2 role keys for regular user")

	// Check authentication key
	authKey := findRoleKey(roleKeys, "user", "authentication")
	require.NotNil(t, authKey, "Authentication key should exist")
	assert.Contains(t, getCapabilities(t, authKey), "access_own_pseudonyms")
	assert.Contains(t, getCapabilities(t, authKey), "login")
	assert.Contains(t, getCapabilities(t, authKey), "session_management")

	// Check self-correlation key
	selfKey := findRoleKey(roleKeys, "user", "self_correlation")
	require.NotNil(t, selfKey, "Self-correlation key should exist")
	assert.Contains(t, getCapabilities(t, selfKey), "verify_own_pseudonym_ownership")
	assert.Contains(t, getCapabilities(t, selfKey), "manage_own_profile")

	// Verify no admin keys were created
	adminKey := findRoleKey(roleKeys, "user", "correlation")
	assert.Nil(t, adminKey, "Regular user should not have correlation key")
}

func TestEnsureDefaultKeys_PlatformAdmin(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user with "platform_admin" role
	testUser := suite.CreateTestUser(t, "admin@example.com", "password123", []string{"platform_admin"})

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, testUser.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 3, "Should have 3 role keys for platform admin")

	// Check authentication key
	authKey := findRoleKey(roleKeys, "platform_admin", "authentication")
	require.NotNil(t, authKey, "Authentication key should exist")

	// Check self-correlation key
	selfKey := findRoleKey(roleKeys, "platform_admin", "self_correlation")
	require.NotNil(t, selfKey, "Self-correlation key should exist")

	// Check correlation key with correct role name
	corrKey := findRoleKey(roleKeys, "platform_admin", "correlation")
	require.NotNil(t, corrKey, "Correlation key should exist for platform_admin")
	assert.Contains(t, getCapabilities(t, corrKey), "access_all_pseudonyms")
	assert.Contains(t, getCapabilities(t, corrKey), "cross_user_correlation")
	assert.Contains(t, getCapabilities(t, corrKey), "moderation")
	assert.Contains(t, getCapabilities(t, corrKey), "compliance")
	assert.Contains(t, getCapabilities(t, corrKey), "legal_requests")

	// Verify the role name is correct (not "admin")
	assert.Equal(t, "platform_admin", corrKey.RoleName, "Role name should be platform_admin, not admin")
}

func TestEnsureDefaultKeys_TrustSafety(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user with "trust_safety" role
	testUser := suite.CreateTestUser(t, "trustsafety@example.com", "password123", []string{"trust_safety"})

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, testUser.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 3, "Should have 3 role keys for trust_safety user")

	// Check correlation key with correct role name
	corrKey := findRoleKey(roleKeys, "trust_safety", "correlation")
	require.NotNil(t, corrKey, "Correlation key should exist for trust_safety")
	assert.Equal(t, "trust_safety", corrKey.RoleName, "Role name should be trust_safety")
}

func TestEnsureDefaultKeys_LegalTeam(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user with "legal_team" role
	testUser := suite.CreateTestUser(t, "legal@example.com", "password123", []string{"legal_team"})

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, testUser.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 3, "Should have 3 role keys for legal_team user")

	// Check correlation key with correct role name
	corrKey := findRoleKey(roleKeys, "legal_team", "correlation")
	require.NotNil(t, corrKey, "Correlation key should exist for legal_team")
	assert.Equal(t, "legal_team", corrKey.RoleName, "Role name should be legal_team")
}

func TestEnsureDefaultKeys_UserWithoutRoles(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user without roles
	testUser := suite.CreateTestUser(t, "noroles@example.com", "password123", nil)

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, testUser.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify only basic user keys were created
	roleKeys := getRoleKeysForUser(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 2, "Should have 2 role keys for user without roles")

	// Should not have any admin keys
	for _, key := range roleKeys {
		assert.NotEqual(t, "correlation", key.Scope, "User without roles should not have correlation key")
	}
}

func TestEnsureDefaultKeys_UserWithMultipleRoles(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user with multiple roles
	testUser := suite.CreateTestUser(t, "multi@example.com", "password123", []string{"user", "platform_admin"})

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, testUser.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 5, "Should have 5 role keys for user with multiple roles (2 for user + 3 for platform_admin)")

	// Check that correlation key exists for platform_admin
	corrKey := findRoleKey(roleKeys, "platform_admin", "correlation")
	require.NotNil(t, corrKey, "Correlation key should exist for platform_admin role")

	// Ensure only one correlation key is created for the admin role
	count := 0
	for _, key := range roleKeys {
		if key.Scope == "correlation" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Should only have one correlation key for admin role")
}

func TestEnsureDefaultKeys_KeyUpdate(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user
	testUser := suite.CreateTestUser(t, "update@example.com", "password123", []string{"user"})

	// Create a key with missing capabilities
	keyData := ibeSystem.GenerateTestRoleKey("user", "authentication")
	expiresAt := time.Now().AddDate(1, 0, 0)
	_, err := roleKeyDAO.CreateRoleKey(ctx, "user", "authentication", keyData, []string{"login"}, expiresAt, testUser.UserID)
	require.NoError(t, err, "Failed to create incomplete key")

	// Ensure default keys (should update existing key)
	err = roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, testUser.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify the key was updated with all required capabilities
	roleKeys := getRoleKeysForUser(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 2, "Should have 2 role keys")

	authKey := findRoleKey(roleKeys, "user", "authentication")
	require.NotNil(t, authKey, "Authentication key should exist")
	capabilities := getCapabilities(t, authKey)
	assert.Contains(t, capabilities, "access_own_pseudonyms")
	assert.Contains(t, capabilities, "login")
	assert.Contains(t, capabilities, "session_management")
}

func TestEnsureDefaultKeys_UserNotFound(t *testing.T) {
	suite := testutil.NewIntegrationTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.Cleanup()

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Try to ensure keys for non-existent user
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, 99999)
	assert.Error(t, err, "Should fail for non-existent user")
	assert.Contains(t, err.Error(), "user 99999 not found")
}

// Helper functions

func findRoleKey(roleKeys []*models.RoleKey, roleName, scope string) *models.RoleKey {
	for _, key := range roleKeys {
		if key.RoleName == roleName && key.Scope == scope {
			return key
		}
	}
	return nil
}

func getCapabilities(t *testing.T, roleKey *models.RoleKey) []string {
	capabilitiesBytes, err := roleKey.Capabilities.Value()
	require.NoError(t, err, "Failed to get capabilities value")

	var capabilities []string
	err = json.Unmarshal(capabilitiesBytes.([]byte), &capabilities)
	require.NoError(t, err, "Failed to unmarshal capabilities")

	return capabilities
}

func getRoleKeysForUser(t *testing.T, db bob.DB, userID int64) []*models.RoleKey {
	ctx := context.Background()

	roleKeys, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.CreatedBy.EQ(userID),
	).All(ctx, db)
	require.NoError(t, err, "Failed to get role keys for user")

	return roleKeys
}
