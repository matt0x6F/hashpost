package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// PermissionDAO provides data access operations for permission checking
type PermissionDAO struct {
	db         bob.Executor
	roleKeyDAO *RoleKeyDAO
}

// NewPermissionDAO creates a new PermissionDAO
func NewPermissionDAO(db bob.Executor) *PermissionDAO {
	return &PermissionDAO{
		db:         db,
		roleKeyDAO: NewRoleKeyDAO(db),
	}
}

// CanAccessPrivateSubforum checks if a user can access a private subforum
// This method checks ALL of the user's pseudonyms (legacy behavior)
func (dao *PermissionDAO) CanAccessPrivateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	// Check if user has global roles that grant access via role keys
	// Get all pseudonyms for this user via identity mappings
	mappings, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)

	if err == nil && len(mappings) > 0 {
		// Check if any of the user's pseudonyms have global role keys (no subforum_id)
		for _, mapping := range mappings {
			roleKeys, err := models.RoleKeys.Query(
				models.SelectWhere.RoleKeys.PseudonymID.EQ(mapping.PseudonymID),
				models.SelectWhere.RoleKeys.IsActive.EQ(true),
			).All(ctx, dao.db)

			if err == nil && len(roleKeys) > 0 {
				for _, roleKey := range roleKeys {
					// Check if this is a global role key (no subforum_id)
					if !roleKey.SubforumID.Valid {
						// Check if this role has access to private subforums
						// Platform admin roles have access to all subforums
						if roleKey.RoleName == constants.RolePlatformAdmin ||
							roleKey.RoleName == constants.RoleTrustSafety ||
							roleKey.RoleName == constants.RoleLegalTeam {
							log.Debug().
								Int64("user_id", userID).
								Int32("subforum_id", subforumID).
								Str("role", roleKey.RoleName).
								Str("pseudonym_id", mapping.PseudonymID).
								Msg("User has global role with access to private subforums")
							return true, nil
						}
					}
				}
			}
		}
	}

	log.Debug().
		Int64("user_id", userID).
		Int32("subforum_id", subforumID).
		Msg("User does not have access to private subforum (platform-wide roles not checked)")
	return false, nil
}

// CanAccessPrivateSubforumWithActivePseudonym checks if a user can access a private subforum
// This method checks ONLY the active pseudonym (secure behavior)
func (dao *PermissionDAO) CanAccessPrivateSubforumWithActivePseudonym(ctx context.Context, userID int64, subforumID int32, activePseudonymID string) (bool, error) {
	// Check if user has global roles that grant access via role keys
	// Get all pseudonyms for this user via identity mappings
	mappings, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)

	if err == nil && len(mappings) > 0 {
		// Check if any of the user's pseudonyms have global role keys (no subforum_id)
		for _, mapping := range mappings {
			roleKeys, err := models.RoleKeys.Query(
				models.SelectWhere.RoleKeys.PseudonymID.EQ(mapping.PseudonymID),
				models.SelectWhere.RoleKeys.IsActive.EQ(true),
			).All(ctx, dao.db)

			if err == nil && len(roleKeys) > 0 {
				for _, roleKey := range roleKeys {
					// Check if this is a global role key (no subforum_id)
					if !roleKey.SubforumID.Valid {
						// Platform admin roles have access to all subforums
						if roleKey.RoleName == constants.RolePlatformAdmin ||
							roleKey.RoleName == constants.RoleTrustSafety ||
							roleKey.RoleName == constants.RoleLegalTeam {
							log.Debug().
								Int64("user_id", userID).
								Int32("subforum_id", subforumID).
								Str("role", roleKey.RoleName).
								Str("pseudonym_id", mapping.PseudonymID).
								Msg("User has global role with access to private subforums")
							return true, nil
						}
					}
				}
			}
		}
	}

	// Check if the active pseudonym has moderator role keys for this subforum
	moderatorKey, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.SubforumID.EQ(subforumID),
		models.SelectWhere.RoleKeys.PseudonymID.EQ(activePseudonymID),
		models.SelectWhere.RoleKeys.RoleName.EQ("moderator"),
		models.SelectWhere.RoleKeys.IsActive.EQ(true),
	).One(ctx, dao.db)

	if err == nil && moderatorKey != nil {
		log.Debug().
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Str("role", moderatorKey.RoleName).
			Msg("Active pseudonym is moderator of private subforum")
		return true, nil
	}

	// Check if user has platform-wide roles that grant access via role keys
	// Get all pseudonyms for this user via identity mappings
	mappings, err = models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)

	if err == nil && len(mappings) > 0 {
		// Check if any of the user's pseudonyms have platform-wide role keys
		platformWideRoles := []string{"platform_admin", "trust_safety", "legal_team"}
		for _, mapping := range mappings {
			for _, platformRole := range platformWideRoles {
				roleKey, err := models.RoleKeys.Query(
					models.SelectWhere.RoleKeys.PseudonymID.EQ(mapping.PseudonymID),
					models.SelectWhere.RoleKeys.RoleName.EQ(platformRole),
					models.SelectWhere.RoleKeys.IsActive.EQ(true),
				).One(ctx, dao.db)

				if err == nil && roleKey != nil {
					log.Debug().
						Int64("user_id", userID).
						Int32("subforum_id", subforumID).
						Str("role", platformRole).
						Str("pseudonym_id", mapping.PseudonymID).
						Msg("User has platform-wide role for private subforum access")
					return true, nil
				}
			}
		}
	}

	log.Debug().
		Int64("user_id", userID).
		Int32("subforum_id", subforumID).
		Str("pseudonym_id", activePseudonymID).
		Msg("User does not have access to private subforum")
	return false, nil
}

// HasSubforumCapability checks if a user has a specific capability for a subforum
// This method checks ALL of the user's pseudonyms (legacy behavior)
func (dao *PermissionDAO) HasSubforumCapability(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
	// Get user's pseudonyms and check if any have moderator permissions
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	// Get all pseudonyms for this user via identity mappings
	mappings, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return false, fmt.Errorf("failed to get user pseudonyms: %w", err)
	}

	// Check if any of the user's pseudonyms have role keys with the required capability
	for _, mapping := range mappings {
		roleKeys, err := models.RoleKeys.Query(
			models.SelectWhere.RoleKeys.SubforumID.EQ(subforumID),
			models.SelectWhere.RoleKeys.PseudonymID.EQ(mapping.PseudonymID),
			models.SelectWhere.RoleKeys.IsActive.EQ(true),
		).All(ctx, dao.db)

		if err == nil && len(roleKeys) > 0 {
			// Check if any role key has the specific capability
			for _, roleKey := range roleKeys {
				// Get capabilities from the role key
				capabilitiesBytes, err := roleKey.Capabilities.Value()
				if err != nil {
					continue
				}
				var capabilities []string
				if capabilitiesBytes != nil {
					if bytes, ok := capabilitiesBytes.([]byte); ok {
						if err := json.Unmarshal(bytes, &capabilities); err == nil {
							for _, cap := range capabilities {
								if cap == capability {
									log.Debug().
										Int64("user_id", userID).
										Int32("subforum_id", subforumID).
										Str("capability", capability).
										Str("role", roleKey.RoleName).
										Str("pseudonym_id", mapping.PseudonymID).
										Msg("User has subforum-specific capability")
									return true, nil
								}
							}
						}
					}
				}

				// Check role-based capabilities
				roleCapabilities := dao.getRoleCapabilities(roleKey.RoleName)
				for _, cap := range roleCapabilities {
					if cap == capability {
						log.Debug().
							Int64("user_id", userID).
							Int32("subforum_id", subforumID).
							Str("capability", capability).
							Str("role", roleKey.RoleName).
							Str("pseudonym_id", mapping.PseudonymID).
							Msg("User has role-based capability")
						return true, nil
					}
				}
			}
		}
	}

	log.Debug().
		Int64("user_id", userID).
		Int32("subforum_id", subforumID).
		Str("capability", capability).
		Msg("User does not have required capability")
	return false, nil
}

// HasSubforumCapabilityWithActivePseudonym checks if the active pseudonym has a specific capability for a subforum
func (dao *PermissionDAO) HasSubforumCapabilityWithActivePseudonym(ctx context.Context, userID int64, subforumID int32, capability string, activePseudonymID string) (bool, error) {
	log.Debug().
		Int64("user_id", userID).
		Int32("subforum_id", subforumID).
		Str("pseudonym_id", activePseudonymID).
		Str("capability", capability).
		Msg("Checking subforum capability with active pseudonym")

	// First check if the active pseudonym has moderator role keys for this subforum
	roleKey, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.SubforumID.EQ(subforumID),
		models.SelectWhere.RoleKeys.PseudonymID.EQ(activePseudonymID),
		models.SelectWhere.RoleKeys.RoleName.EQ("moderator"),
		models.SelectWhere.RoleKeys.IsActive.EQ(true),
	).One(ctx, dao.db)

	if err != nil {
		log.Debug().
			Err(err).
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Msg("Error querying role key record")
	} else if roleKey != nil {
		log.Debug().
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Str("role", roleKey.RoleName).
			Msg("Found moderator role key")

		// Check if the moderator role has this capability
		roleCaps := dao.getRoleCapabilities(roleKey.RoleName)
		log.Debug().
			Str("role", roleKey.RoleName).
			Interface("role_capabilities", roleCaps).
			Msg("Role capabilities")

		for _, cap := range roleCaps {
			if cap == capability {
				log.Debug().
					Int64("user_id", userID).
					Int32("subforum_id", subforumID).
					Str("pseudonym_id", activePseudonymID).
					Str("capability", capability).
					Str("role", roleKey.RoleName).
					Msg("Active pseudonym has capability through moderator role")
				return true, nil
			}
		}

		// Check specific capabilities from JSON
		capabilitiesBytes, err := roleKey.Capabilities.Value()
		if err == nil && capabilitiesBytes != nil {
			if bytes, ok := capabilitiesBytes.([]byte); ok {
				var capabilities []string
				if err := json.Unmarshal(bytes, &capabilities); err == nil {
					for _, cap := range capabilities {
						if cap == capability {
							log.Debug().
								Int64("user_id", userID).
								Int32("subforum_id", subforumID).
								Str("pseudonym_id", activePseudonymID).
								Str("capability", capability).
								Msg("Active pseudonym has capability through specific capability")
							return true, nil
						}
					}
				}
			}
		}
	} else {
		log.Debug().
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Msg("No moderator role key found")
	}

	// Check if the active pseudonym has platform-wide capabilities through role keys
	roleKeys, err := dao.roleKeyDAO.ListRoleKeysByPseudonym(ctx, activePseudonymID)
	if err != nil {
		return false, fmt.Errorf("failed to get role keys for pseudonym: %w", err)
	}

	for _, roleKey := range roleKeys {
		// Skip subforum-specific keys (we already checked those above)
		if roleKey.SubforumID.Valid {
			continue
		}

		// Check capabilities from this role key
		capabilitiesBytes, err := roleKey.Capabilities.Value()
		if err == nil && capabilitiesBytes != nil {
			if bytes, ok := capabilitiesBytes.([]byte); ok {
				var capabilities []string
				if err := json.Unmarshal(bytes, &capabilities); err == nil {
					for _, cap := range capabilities {
						if cap == capability {
							log.Debug().
								Int64("user_id", userID).
								Int32("subforum_id", subforumID).
								Str("pseudonym_id", activePseudonymID).
								Str("capability", capability).
								Str("role", roleKey.RoleName).
								Msg("Active pseudonym has platform-wide capability through role key")
							return true, nil
						}
					}
				}
			}
		}
	}

	log.Debug().
		Int64("user_id", userID).
		Int32("subforum_id", subforumID).
		Str("pseudonym_id", activePseudonymID).
		Str("capability", capability).
		Msg("Active pseudonym does not have capability")
	return false, nil
}

// GetUserSubforumRoles returns the roles a user has for a specific subforum
func (dao *PermissionDAO) GetUserSubforumRoles(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	var roles []string

	// Get user's pseudonyms and check their moderator roles
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get all pseudonyms for this user via identity mappings
	mappings, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pseudonyms: %w", err)
	}

	// Get subforum moderator role keys for each pseudonym
	for _, mapping := range mappings {
		roleKey, err := models.RoleKeys.Query(
			models.SelectWhere.RoleKeys.SubforumID.EQ(subforumID),
			models.SelectWhere.RoleKeys.PseudonymID.EQ(mapping.PseudonymID),
			models.SelectWhere.RoleKeys.IsActive.EQ(true),
		).One(ctx, dao.db)

		if err == nil && roleKey != nil {
			roles = append(roles, roleKey.RoleName)
		}
	}

	// Get platform-wide roles from role keys instead of user.roles
	// Since users don't have direct roles anymore, we'll only return subforum-specific roles
	// Platform-wide roles are managed through role keys for specific pseudonyms
	return roles, nil
}

// GetUserSubforumCapabilities returns all capabilities for a user in a specific subforum
func (dao *PermissionDAO) GetUserSubforumCapabilities(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	var capabilities []string

	// Get all pseudonyms for this user via identity mappings
	mappings, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pseudonyms: %w", err)
	}

	// Check if any of the user's pseudonyms have role keys for this subforum
	for _, mapping := range mappings {
		roleKey, err := models.RoleKeys.Query(
			models.SelectWhere.RoleKeys.SubforumID.EQ(subforumID),
			models.SelectWhere.RoleKeys.PseudonymID.EQ(mapping.PseudonymID),
			models.SelectWhere.RoleKeys.IsActive.EQ(true),
		).One(ctx, dao.db)

		if err == nil && roleKey != nil {
			// Add role-based capabilities
			roleCaps := dao.getRoleCapabilities(roleKey.RoleName)
			capabilities = append(capabilities, roleCaps...)

			// Add specific capabilities from JSON
			capabilitiesBytes, err := roleKey.Capabilities.Value()
			if err == nil && capabilitiesBytes != nil {
				if bytes, ok := capabilitiesBytes.([]byte); ok {
					var capabilitiesList []string
					if err := json.Unmarshal(bytes, &capabilitiesList); err == nil {
						capabilities = append(capabilities, capabilitiesList...)
					}
				}
			}
		}
	}

	return capabilities, nil
}

// CanModerateSubforum checks if a user can moderate a specific subforum
func (dao *PermissionDAO) CanModerateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, constants.CapabilityModerateContent)
}

// CanBanUsers checks if a user can ban users in a specific subforum
func (dao *PermissionDAO) CanBanUsers(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, constants.CapabilityBanUsers)
}

// CanRemoveContent checks if a user can remove content in a specific subforum
func (dao *PermissionDAO) CanRemoveContent(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, constants.CapabilityRemoveContent)
}

// CanManageModerators checks if a user can manage moderators in a specific subforum
func (dao *PermissionDAO) CanManageModerators(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, constants.CapabilityManageModerators)
}

// getRoleCapabilities returns the capabilities associated with a specific role
func (dao *PermissionDAO) getRoleCapabilities(role string) []string {
	// Use the constants package to get role capabilities
	return constants.GetRoleCapabilities(role)
}

// GetActivePseudonymRolesAndCapabilities returns the roles and capabilities for a specific active pseudonym
func (dao *PermissionDAO) GetActivePseudonymRolesAndCapabilities(ctx context.Context, userID int64, activePseudonymID string) ([]string, []string, error) {
	// Get role keys for the pseudonym
	roleKeys, err := dao.roleKeyDAO.ListRoleKeysByPseudonym(ctx, activePseudonymID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get role keys: %w", err)
	}

	// Extract unique roles and capabilities from role keys
	roleSet := make(map[string]bool)
	capabilitySet := make(map[string]bool)

	for _, roleKey := range roleKeys {
		// Add role name
		roleSet[roleKey.RoleName] = true

		// Extract capabilities from JSON
		var capabilities []string
		rawValue, err := roleKey.Capabilities.Value()
		if err != nil {
			continue
		}
		if err := json.Unmarshal(rawValue.([]byte), &capabilities); err != nil {
			continue
		}
		for _, capability := range capabilities {
			capabilitySet[capability] = true
		}
	}

	// Convert sets to slices
	var roles []string
	for role := range roleSet {
		roles = append(roles, role)
	}

	var capabilities []string
	for capability := range capabilitySet {
		capabilities = append(capabilities, capability)
	}

	// If no roles found, default to "user"
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	// If no capabilities found, provide default capabilities
	if len(capabilities) == 0 {
		capabilities = []string{"create_content", "vote", "message", "report"}
	}

	return roles, capabilities, nil
}

// GetUnifiedActivePseudonymRolesAndCapabilities returns the unified roles and capabilities for a specific active pseudonym
// This combines both global pseudonym capabilities and subforum-specific moderator capabilities
func (dao *PermissionDAO) GetUnifiedActivePseudonymRolesAndCapabilities(ctx context.Context, userID int64, activePseudonymID string, subforumID *int32) ([]string, []string, error) {
	// Get role keys for the pseudonym
	roleKeys, err := dao.roleKeyDAO.ListRoleKeysByPseudonym(ctx, activePseudonymID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get role keys: %w", err)
	}

	// Extract unique roles and capabilities from role keys
	roleSet := make(map[string]bool)
	capabilitySet := make(map[string]bool)

	for _, roleKey := range roleKeys {
		// Add role name
		roleSet[roleKey.RoleName] = true

		// Extract capabilities from JSON
		var capabilities []string
		rawValue, err := roleKey.Capabilities.Value()
		if err != nil {
			continue
		}
		if err := json.Unmarshal(rawValue.([]byte), &capabilities); err != nil {
			continue
		}
		for _, capability := range capabilities {
			capabilitySet[capability] = true
		}
	}

	// Convert sets to slices
	var roles []string
	for role := range roleSet {
		roles = append(roles, role)
	}

	var capabilities []string
	for capability := range capabilitySet {
		capabilities = append(capabilities, capability)
	}

	// If no roles found, default to "user"
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	// If no capabilities found, provide default capabilities
	if len(capabilities) == 0 {
		capabilities = []string{"create_content", "vote", "message", "report"}
	}

	// If subforumID is provided, add subforum-specific capabilities
	if subforumID != nil {
		subforumCapabilities, err := dao.getSubforumCapabilitiesForPseudonym(ctx, *subforumID, activePseudonymID)
		if err != nil {
			log.Warn().Err(err).Int32("subforum_id", *subforumID).Str("pseudonym_id", activePseudonymID).Msg("Failed to get subforum capabilities")
		} else {
			// Add subforum-specific capabilities
			capabilities = append(capabilities, subforumCapabilities...)

			// Add "moderator" role if the pseudonym has subforum-specific capabilities
			if len(subforumCapabilities) > 0 {
				// Check if "moderator" role is not already present
				hasModeratorRole := false
				for _, role := range roles {
					if role == "moderator" {
						hasModeratorRole = true
						break
					}
				}
				if !hasModeratorRole {
					roles = append(roles, "moderator")
				}
			}
		}
	}

	// Remove duplicates from capabilities
	capabilities = dao.removeDuplicateCapabilities(capabilities)

	return roles, capabilities, nil
}

// HasUnifiedCapability checks if the active pseudonym has a specific capability
// This combines both global pseudonym capabilities and subforum-specific moderator capabilities
func (dao *PermissionDAO) HasUnifiedCapability(ctx context.Context, userID int64, activePseudonymID string, capability string, subforumID *int32) (bool, error) {
	log.Debug().
		Int64("user_id", userID).
		Str("pseudonym_id", activePseudonymID).
		Str("capability", capability).
		Interface("subforum_id", subforumID).
		Msg("Checking unified capability")

	// Get unified roles and capabilities
	_, capabilities, err := dao.GetUnifiedActivePseudonymRolesAndCapabilities(ctx, userID, activePseudonymID, subforumID)
	if err != nil {
		return false, fmt.Errorf("failed to get unified roles and capabilities: %w", err)
	}

	// Check if the capability is present
	for _, cap := range capabilities {
		if cap == capability {
			log.Debug().
				Int64("user_id", userID).
				Str("pseudonym_id", activePseudonymID).
				Str("capability", capability).
				Interface("subforum_id", subforumID).
				Msg("Active pseudonym has unified capability")
			return true, nil
		}
	}

	log.Debug().
		Int64("user_id", userID).
		Str("pseudonym_id", activePseudonymID).
		Str("capability", capability).
		Interface("subforum_id", subforumID).
		Msg("Active pseudonym does not have unified capability")
	return false, nil
}

// getSubforumCapabilitiesForPseudonym gets subforum-specific capabilities for a specific pseudonym
func (dao *PermissionDAO) getSubforumCapabilitiesForPseudonym(ctx context.Context, subforumID int32, pseudonymID string) ([]string, error) {
	var capabilities []string

	// Check if database is available
	if dao.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Check if the pseudonym has role keys for this subforum
	roleKey, err := models.RoleKeys.Query(
		models.SelectWhere.RoleKeys.SubforumID.EQ(subforumID),
		models.SelectWhere.RoleKeys.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.RoleKeys.IsActive.EQ(true),
	).One(ctx, dao.db)

	if err != nil {
		return nil, fmt.Errorf("failed to query role key record: %w", err)
	}

	if roleKey != nil {
		log.Debug().
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", pseudonymID).
			Str("role", roleKey.RoleName).
			Msg("Found role key record for subforum capabilities")

		// Add role-based capabilities
		roleCaps := dao.getRoleCapabilities(roleKey.RoleName)
		capabilities = append(capabilities, roleCaps...)

		// Add specific capabilities from JSON
		capabilitiesBytes, err := roleKey.Capabilities.Value()
		if err == nil && capabilitiesBytes != nil {
			if bytes, ok := capabilitiesBytes.([]byte); ok {
				var capabilitiesList []string
				if err := json.Unmarshal(bytes, &capabilitiesList); err == nil {
					capabilities = append(capabilities, capabilitiesList...)
				}
			}
		}
	}

	return capabilities, nil
}

// removeDuplicateCapabilities removes duplicate capabilities from a slice
func (dao *PermissionDAO) removeDuplicateCapabilities(capabilities []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, cap := range capabilities {
		if !seen[cap] {
			seen[cap] = true
			result = append(result, cap)
		}
	}

	return result
}
