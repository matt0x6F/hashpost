//go:build integration

package integration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
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

	// Verify per-user role keys were created
	roleKeys := getPerUserRoleKeys(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 2, "Should have 2 per-user role keys for regular user")

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

	// Verify global role keys exist for platform_admin
	roleKeys := getGlobalRoleKeysForRoles(t, suite.DB, []string{"platform_admin"})
	assert.Len(t, roleKeys, 3, "Should have 3 global role keys for platform admin")

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
	roleKeys := getPerUserRoleKeys(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 3, "Should have 3 per-user role keys for trust_safety user")

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
	roleKeys := getPerUserRoleKeys(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 3, "Should have 3 per-user role keys for legal_team user")

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
	roleKeys := getPerUserRoleKeys(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 2, "Should have 2 per-user role keys for user without roles")

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
	roleKeys := getPerUserRoleKeys(t, suite.DB, testUser.UserID)
	assert.Len(t, roleKeys, 5, "Should have 5 per-user role keys for user with multiple roles (2 for user + 3 for platform_admin)")

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

	ctx := context.Background()
	roleKeyDAO := suite.RoleKeyDAO
	ibeSystem := suite.IBESystem

	// Create test user without calling EnsureDefaultKeys first
	// This simulates a user that was created before the role key system was implemented
	passwordHash := hashPassword("password123")
	user, err := suite.UserDAO.CreateUser(ctx, "update@example.com", passwordHash)
	require.NoError(t, err, "Failed to create test user")

	// Set user roles
	rolesJSON, _ := json.Marshal([]string{"user"})
	capabilities := getCapabilitiesForRoles([]string{"user"})
	capabilitiesJSON, _ := json.Marshal(capabilities)

	rolesNull := sql.Null[types.JSON[json.RawMessage]]{}
	rolesNull.Scan(rolesJSON)

	capabilitiesNull := sql.Null[types.JSON[json.RawMessage]]{}
	capabilitiesNull.Scan(capabilitiesJSON)

	updates := &dbmodels.UserSetter{
		Roles:        &rolesNull,
		Capabilities: &capabilitiesNull,
	}

	err = suite.UserDAO.UpdateUser(ctx, user.UserID, updates)
	require.NoError(t, err, "Failed to update test user roles")

	// Track user for cleanup
	suite.Tracker.TrackUser(user.UserID)

	// Create a key with missing capabilities (simulating an incomplete key)
	keyData := ibeSystem.GenerateTestRoleKey("user", "authentication")
	expiresAt := time.Now().AddDate(1, 0, 0)
	createdKey, err := roleKeyDAO.CreateRoleKey(ctx, "user", "authentication", keyData, []string{"login"}, expiresAt, user.UserID)
	require.NoError(t, err, "Failed to create incomplete key")

	// Track the created role key for cleanup
	suite.Tracker.TrackRoleKey(createdKey.KeyID.String())

	// Verify we have only 1 key initially
	roleKeys := getPerUserRoleKeys(t, suite.DB, user.UserID)
	assert.Len(t, roleKeys, 1, "Should have 1 per-user role key initially")

	// Ensure default keys (should update existing key and create missing ones)
	err = roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, user.UserID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify the keys were properly created/updated
	roleKeys = getPerUserRoleKeys(t, suite.DB, user.UserID)
	assert.Len(t, roleKeys, 2, "Should have 2 per-user role keys after ensuring defaults")

	// Check that the authentication key was updated with all required capabilities
	authKey := findRoleKey(roleKeys, "user", "authentication")
	require.NotNil(t, authKey, "Authentication key should exist")
	capabilities = getCapabilities(t, authKey)
	assert.Contains(t, capabilities, "access_own_pseudonyms")
	assert.Contains(t, capabilities, "login")
	assert.Contains(t, capabilities, "session_management")

	// Check that the self-correlation key was created
	selfKey := findRoleKey(roleKeys, "user", "self_correlation")
	require.NotNil(t, selfKey, "Self-correlation key should exist")
	capabilities = getCapabilities(t, selfKey)
	assert.Contains(t, capabilities, "verify_own_pseudonym_ownership")
	assert.Contains(t, capabilities, "manage_own_profile")

	defer suite.Cleanup()
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

func getPerUserRoleKeys(t *testing.T, db bob.DB, userID int64) []*models.RoleKey {
	ctx := context.Background()
	roleKeys, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.CreatedBy.EQ(userID),
	).All(ctx, db)
	require.NoError(t, err, "Failed to get per-user role keys for user")
	return roleKeys
}

func getGlobalRoleKeysForRoles(t *testing.T, db bob.DB, roles []string) []*models.RoleKey {
	ctx := context.Background()
	var allRoleKeys []*models.RoleKey
	for _, roleName := range roles {
		roleKeys, err := models.RoleKeys.Query(
			models.SelectWhere.RoleKeys.RoleName.EQ(roleName),
		).All(ctx, db)
		require.NoError(t, err, "Failed to get global role keys for role %s", roleName)
		allRoleKeys = append(allRoleKeys, roleKeys...)
	}
	return allRoleKeys
}

// Helper functions for user creation
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func getCapabilitiesForRoles(roles []string) []string {
	capabilities := []string{}

	for _, role := range roles {
		switch role {
		case "user":
			capabilities = append(capabilities, "create_content", "vote", "message", "report")
		case "platform_admin":
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "create_subforum", "moderation", "compliance", "legal_requests")
		case "trust_safety":
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "moderation", "compliance")
		case "legal_team":
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "compliance", "legal_requests")
		}
	}

	return capabilities
}
