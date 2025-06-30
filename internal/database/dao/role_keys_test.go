package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) bob.DB {
	// Use test database configuration
	config := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5433, // Test database port
		User:     "hashpost",
		Password: "hashpost_test",
		Database: "hashpost_test",
		SSLMode:  "disable",
	}

	db, err := database.NewConnection(config)
	require.NoError(t, err, "Failed to connect to test database")

	return db
}

// createTestUser creates a test user with specified roles
func createTestUser(t *testing.T, db bob.DB, email string, roles []string) int64 {
	userDAO := NewUserDAO(db)
	ctx := context.Background()

	// Add a timestamp to the email to ensure uniqueness
	email = fmt.Sprintf("%s_%d", email, time.Now().UnixNano())

	// Create user
	user, err := userDAO.CreateUser(ctx, email, "hashed_password")
	require.NoError(t, err, "Failed to create test user")

	// Set roles if provided
	if len(roles) > 0 {
		rolesJSON, err := json.Marshal(roles)
		require.NoError(t, err, "Failed to marshal roles")

		rolesNull := sql.Null[types.JSON[json.RawMessage]]{}
		rolesNull.Scan(rolesJSON)

		updates := &models.UserSetter{
			Roles: &rolesNull,
		}

		err = userDAO.UpdateUser(ctx, user.UserID, updates)
		require.NoError(t, err, "Failed to update user roles")
	}

	return user.UserID
}

// cleanupTestUser removes test user and associated data
func cleanupTestUser(t *testing.T, db bob.DB, userID int64) {
	ctx := context.Background()

	// Delete role keys
	_, err := db.ExecContext(ctx, "DELETE FROM role_keys WHERE created_by = $1", userID)
	require.NoError(t, err, "Failed to cleanup role keys")

	// Delete user
	_, err = db.ExecContext(ctx, "DELETE FROM users WHERE user_id = $1", userID)
	require.NoError(t, err, "Failed to cleanup user")
}

// getRoleKeysForUser retrieves all role keys for a user
func getRoleKeysForUser(t *testing.T, db bob.DB, userID int64) []*models.RoleKey {
	ctx := context.Background()

	roleKeys, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.CreatedBy.EQ(userID),
	).All(ctx, db)
	require.NoError(t, err, "Failed to get role keys for user")

	return roleKeys
}

func TestEnsureDefaultKeys_RegularUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

	// Create test user with "user" role
	userID := createTestUser(t, db, "testuser@example.com", []string{"user"})
	defer cleanupTestUser(t, db, userID)

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, userID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, db, userID)
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
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

	// Create test user with "platform_admin" role
	userID := createTestUser(t, db, "admin@example.com", []string{"platform_admin"})
	defer cleanupTestUser(t, db, userID)

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, userID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, db, userID)
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
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

	// Create test user with "trust_safety" role
	userID := createTestUser(t, db, "trustsafety@example.com", []string{"trust_safety"})
	defer cleanupTestUser(t, db, userID)

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, userID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, db, userID)
	assert.Len(t, roleKeys, 3, "Should have 3 role keys for trust_safety user")

	// Check correlation key with correct role name
	corrKey := findRoleKey(roleKeys, "trust_safety", "correlation")
	require.NotNil(t, corrKey, "Correlation key should exist for trust_safety")
	assert.Equal(t, "trust_safety", corrKey.RoleName, "Role name should be trust_safety")
}

func TestEnsureDefaultKeys_LegalTeam(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

	// Create test user with "legal_team" role
	userID := createTestUser(t, db, "legal@example.com", []string{"legal_team"})
	defer cleanupTestUser(t, db, userID)

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, userID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, db, userID)
	assert.Len(t, roleKeys, 3, "Should have 3 role keys for legal_team user")

	// Check correlation key with correct role name
	corrKey := findRoleKey(roleKeys, "legal_team", "correlation")
	require.NotNil(t, corrKey, "Correlation key should exist for legal_team")
	assert.Equal(t, "legal_team", corrKey.RoleName, "Role name should be legal_team")
}

func TestEnsureDefaultKeys_UserWithoutRoles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

	// Create test user without roles
	userID := createTestUser(t, db, "noroles@example.com", nil)
	defer cleanupTestUser(t, db, userID)

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, userID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify only basic user keys were created
	roleKeys := getRoleKeysForUser(t, db, userID)
	assert.Len(t, roleKeys, 2, "Should have 2 role keys for user without roles")

	// Should not have any admin keys
	for _, key := range roleKeys {
		assert.NotEqual(t, "correlation", key.Scope, "User without roles should not have correlation key")
	}
}

func TestEnsureDefaultKeys_UserWithMultipleRoles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

	// Create test user with multiple roles
	userID := createTestUser(t, db, "multi@example.com", []string{"user", "platform_admin"})
	defer cleanupTestUser(t, db, userID)

	// Ensure default keys
	err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, userID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify role keys were created
	roleKeys := getRoleKeysForUser(t, db, userID)
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
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

	// Create test user
	userID := createTestUser(t, db, "update@example.com", []string{"user"})
	defer cleanupTestUser(t, db, userID)

	// Create a key with missing capabilities
	keyData := ibeSystem.GenerateTestRoleKey("user", "authentication")
	expiresAt := time.Now().AddDate(1, 0, 0)
	_, err := roleKeyDAO.CreateRoleKey(ctx, "user", "authentication", keyData, []string{"login"}, expiresAt, userID)
	require.NoError(t, err, "Failed to create incomplete key")

	// Ensure default keys (should update existing key)
	err = roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, userID)
	require.NoError(t, err, "Failed to ensure default keys")

	// Verify the key was updated with all required capabilities
	roleKeys := getRoleKeysForUser(t, db, userID)
	assert.Len(t, roleKeys, 2, "Should have 2 role keys")

	authKey := findRoleKey(roleKeys, "user", "authentication")
	require.NotNil(t, authKey, "Authentication key should exist")
	capabilities := getCapabilities(t, authKey)
	assert.Contains(t, capabilities, "access_own_pseudonyms")
	assert.Contains(t, capabilities, "login")
	assert.Contains(t, capabilities, "session_management")
}

func TestEnsureDefaultKeys_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	roleKeyDAO := NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystem()

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
