package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
)

// RoleKeyDAO handles database operations for role keys
type RoleKeyDAO struct {
	db bob.Executor
}

// NewRoleKeyDAO creates a new RoleKeyDAO
func NewRoleKeyDAO(db bob.Executor) *RoleKeyDAO {
	return &RoleKeyDAO{
		db: db,
	}
}

// CreateRoleKey creates a new role key in the database
func (dao *RoleKeyDAO) CreateRoleKey(ctx context.Context, roleName, scope string, keyData []byte, capabilities []string, expiresAt time.Time, createdBy int64) (*models.RoleKey, error) {
	// Convert capabilities to JSON
	capabilitiesJSON, err := json.Marshal(capabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	now := time.Now()
	keyVersion := int32(1) // Start with version 1
	isActive := sql.Null[bool]{}
	isActive.Scan(true)

	// Convert capabilities to the correct type
	var capabilitiesType types.JSON[json.RawMessage]
	capabilitiesType.Scan(capabilitiesJSON)

	setter := &models.RoleKeySetter{
		RoleName:     &roleName,
		Scope:        &scope,
		KeyData:      &keyData,
		KeyVersion:   &keyVersion,
		Capabilities: &capabilitiesType,
		ExpiresAt:    &expiresAt,
		IsActive:     &isActive,
		CreatedBy:    &createdBy,
	}

	// Set created_at
	createdAt := sql.Null[time.Time]{}
	createdAt.Scan(now)
	setter.CreatedAt = &createdAt

	// Insert the role key
	roleKey, err := models.RoleKeys.Insert(setter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create role key: %w", err)
	}

	return roleKey, nil
}

// GetRoleKey retrieves a role key by role name and scope
func (dao *RoleKeyDAO) GetRoleKey(ctx context.Context, roleName, scope string) (*models.RoleKey, error) {
	roleKey, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.RoleName.EQ(roleName),
		models.SelectWhere.RoleKeys.Scope.EQ(scope),
		models.SelectWhere.RoleKeys.IsActive.EQ(true),
		models.SelectWhere.RoleKeys.ExpiresAt.GT(time.Now()),
	).One(ctx, dao.db)

	if err != nil {
		return nil, fmt.Errorf("failed to get role key for role=%s scope=%s: %w", roleName, scope, err)
	}

	return roleKey, nil
}

// GetRoleKeyByID retrieves a role key by its ID
func (dao *RoleKeyDAO) GetRoleKeyByID(ctx context.Context, keyID string) (*models.RoleKey, error) {
	// Convert string to UUID
	uuid, err := uuid.FromString(keyID)
	if err != nil {
		return nil, fmt.Errorf("invalid key ID format: %w", err)
	}

	roleKey, err := models.FindRoleKey(ctx, dao.db, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key by ID %s: %w", keyID, err)
	}

	return roleKey, nil
}

// ListRoleKeys retrieves all active role keys
func (dao *RoleKeyDAO) ListRoleKeys(ctx context.Context) ([]*models.RoleKey, error) {
	roleKeys, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.IsActive.EQ(true),
		models.SelectWhere.RoleKeys.ExpiresAt.GT(time.Now()),
	).All(ctx, dao.db)

	if err != nil {
		return nil, fmt.Errorf("failed to list role keys: %w", err)
	}

	return roleKeys, nil
}

// ListRoleKeysByRole retrieves all active role keys for a specific role
func (dao *RoleKeyDAO) ListRoleKeysByRole(ctx context.Context, roleName string) ([]*models.RoleKey, error) {
	roleKeys, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.RoleName.EQ(roleName),
		models.SelectWhere.RoleKeys.IsActive.EQ(true),
		models.SelectWhere.RoleKeys.ExpiresAt.GT(time.Now()),
	).All(ctx, dao.db)

	if err != nil {
		return nil, fmt.Errorf("failed to list role keys for role %s: %w", roleName, err)
	}

	return roleKeys, nil
}

// DeactivateRoleKey deactivates a role key
func (dao *RoleKeyDAO) DeactivateRoleKey(ctx context.Context, keyID string) error {
	// Convert string to UUID
	uuid, err := uuid.FromString(keyID)
	if err != nil {
		return fmt.Errorf("invalid key ID format: %w", err)
	}

	roleKey, err := models.FindRoleKey(ctx, dao.db, uuid)
	if err != nil {
		return fmt.Errorf("failed to find role key %s: %w", keyID, err)
	}

	isActive := sql.Null[bool]{}
	isActive.Scan(false)
	setter := &models.RoleKeySetter{
		IsActive: &isActive,
	}

	err = roleKey.Update(ctx, dao.db, setter)
	if err != nil {
		return fmt.Errorf("failed to deactivate role key %s: %w", keyID, err)
	}

	return nil
}

// ValidateKeyCapability checks if a role key has a specific capability
func (dao *RoleKeyDAO) ValidateKeyCapability(ctx context.Context, roleName, scope, requiredCapability string) (bool, error) {
	roleKey, err := dao.GetRoleKey(ctx, roleName, scope)
	if err != nil {
		return false, fmt.Errorf("failed to get role key for validation: %w", err)
	}

	// Parse capabilities from JSON
	var capabilities []string
	capabilitiesBytes, err := roleKey.Capabilities.Value()
	if err != nil {
		return false, fmt.Errorf("failed to get capabilities value: %w", err)
	}

	if err := json.Unmarshal(capabilitiesBytes.([]byte), &capabilities); err != nil {
		return false, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	// Check if the required capability is present
	for _, capability := range capabilities {
		if capability == requiredCapability {
			return true, nil
		}
	}

	return false, nil
}

// GetKeyData retrieves the key data for a role key
func (dao *RoleKeyDAO) GetKeyData(ctx context.Context, roleName, scope string) ([]byte, error) {
	roleKey, err := dao.GetRoleKey(ctx, roleName, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key: %w", err)
	}

	return roleKey.KeyData, nil
}

// EnsureDefaultKeys creates default role keys if they don't exist
func (dao *RoleKeyDAO) EnsureDefaultKeys(ctx context.Context, ibeSystem interface{}, userID int64) error {
	// Type assert to get the IBE system
	ibe, ok := ibeSystem.(*ibe.IBESystem)
	if !ok {
		return fmt.Errorf("invalid IBE system type")
	}

	// Get the user's actual role from the database
	userDAO := NewUserDAO(dao.db)
	user, err := userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", userID, err)
	}
	if user == nil {
		return fmt.Errorf("user %d not found", userID)
	}

	// Parse user's roles from JSON
	var userRoles []string
	if user.Roles.Valid {
		rolesBytes, err := user.Roles.V.Value()
		if err != nil {
			return fmt.Errorf("failed to get user roles value: %w", err)
		}
		if err := json.Unmarshal(rolesBytes.([]byte), &userRoles); err != nil {
			return fmt.Errorf("failed to unmarshal user roles: %w", err)
		}
	}

	// If no roles found, default to "user"
	if len(userRoles) == 0 {
		userRoles = []string{"user"}
	}

	log.Debug().Int64("user_id", userID).Strs("user_roles", userRoles).Msg("Provisioning role keys for user")

	// Define default keys for each user role
	defaultKeys := []struct {
		roleName     string
		scope        string
		capabilities []string
	}{}

	// Add authentication and self_correlation keys for each user role
	for _, userRole := range userRoles {
		defaultKeys = append(defaultKeys, struct {
			roleName     string
			scope        string
			capabilities []string
		}{
			roleName: userRole,
			scope:    "authentication",
			capabilities: []string{
				"access_own_pseudonyms",
				"login",
				"session_management",
			},
		})
		defaultKeys = append(defaultKeys, struct {
			roleName     string
			scope        string
			capabilities []string
		}{
			roleName: userRole,
			scope:    "self_correlation",
			capabilities: []string{
				"verify_own_pseudonym_ownership",
				"manage_own_profile",
			},
		})
	}

	// Add admin-specific correlation keys for admin roles
	adminRoles := []string{"platform_admin", "trust_safety", "legal_team"}
	for _, userRole := range userRoles {
		for _, adminRole := range adminRoles {
			if userRole == adminRole {
				defaultKeys = append(defaultKeys, struct {
					roleName     string
					scope        string
					capabilities []string
				}{
					roleName: userRole, // Use the actual user role, not hardcoded "admin"
					scope:    "correlation",
					capabilities: []string{
						"access_all_pseudonyms",
						"cross_user_correlation",
						"moderation",
						"compliance",
						"legal_requests",
					},
				})
				break
			}
		}
	}

	log.Debug().Int64("user_id", userID).Msgf("defaultKeys to provision: %+v", defaultKeys)

	// Check if each default key exists, create if not
	for _, keyDef := range defaultKeys {
		existingKey, err := dao.GetRoleKey(ctx, keyDef.roleName, keyDef.scope)
		if err != nil {
			// Key doesn't exist, create it with proper IBE key
			expiresAt := time.Now().AddDate(1, 0, 0) // Expire in 1 year

			// Generate key data using the actual role name and scope
			keyData := ibe.GenerateTestRoleKey(keyDef.roleName, keyDef.scope)

			_, err = dao.CreateRoleKey(ctx, keyDef.roleName, keyDef.scope, keyData, keyDef.capabilities, expiresAt, userID)
			if err != nil {
				log.Error().Str("role", keyDef.roleName).Str("scope", keyDef.scope).Err(err).Msg("Failed to create role key")
				return fmt.Errorf("failed to create default key for role=%s scope=%s: %w", keyDef.roleName, keyDef.scope, err)
			}
		} else {
			// Key exists, check if it needs updating
			capabilitiesBytes, err := existingKey.Capabilities.Value()
			if err != nil {
				return fmt.Errorf("failed to get capabilities value: %w", err)
			}

			var capabilities []string
			if err := json.Unmarshal(capabilitiesBytes.([]byte), &capabilities); err != nil {
				return fmt.Errorf("failed to unmarshal existing capabilities: %w", err)
			}

			// Check if all required capabilities are present
			capabilityMap := make(map[string]bool)
			for _, cap := range capabilities {
				capabilityMap[cap] = true
			}

			needsUpdate := false
			for _, requiredCap := range keyDef.capabilities {
				if !capabilityMap[requiredCap] {
					needsUpdate = true
					break
				}
			}

			if needsUpdate {
				// Update the key with new capabilities
				capabilitiesJSON, _ := json.Marshal(keyDef.capabilities)
				var capabilitiesType types.JSON[json.RawMessage]
				capabilitiesType.Scan(capabilitiesJSON)
				setter := &models.RoleKeySetter{
					Capabilities: &capabilitiesType,
				}

				err = existingKey.Update(ctx, dao.db, setter)
				if err != nil {
					log.Error().Str("role", keyDef.roleName).Str("scope", keyDef.scope).Err(err).Msg("Failed to update role key capabilities")
					return fmt.Errorf("failed to update key capabilities: %w", err)
				}
			}
		}
	}

	// After provisioning, log all role keys for the user
	roleKeys, err := models.RoleKeys.Query(models.SelectWhere.RoleKeys.CreatedBy.EQ(userID)).All(ctx, dao.db)
	if err == nil {
		for _, k := range roleKeys {
			log.Debug().Str("role", k.RoleName).Str("scope", k.Scope).Msg("Role key present in DB after provisioning")
		}
	} else {
		log.Error().Int64("user_id", userID).Err(err).Msg("Failed to query role keys after provisioning")
	}

	return nil
}
