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
	db bob.Executor
}

// NewPermissionDAO creates a new PermissionDAO
func NewPermissionDAO(db bob.Executor) *PermissionDAO {
	return &PermissionDAO{
		db: db,
	}
}

// CanAccessPrivateSubforum checks if a user can access a private subforum
// This method checks ALL of the user's pseudonyms (legacy behavior)
func (dao *PermissionDAO) CanAccessPrivateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	// Check if user has platform-wide roles that grant access
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	platformWideRoles := []string{"platform_admin", "trust_safety", "legal_team"}
	if user.Roles.Valid {
		rawValue, err := user.Roles.V.Value()
		if err != nil {
			return false, fmt.Errorf("failed to get user roles value: %w", err)
		}
		var roles []string
		if err := json.Unmarshal(rawValue.([]byte), &roles); err == nil {
			for _, role := range roles {
				for _, platformRole := range platformWideRoles {
					if role == platformRole {
						log.Debug().
							Int64("user_id", userID).
							Int32("subforum_id", subforumID).
							Str("role", role).
							Msg("User has platform-wide role for private subforum access")
						return true, nil
					}
				}
			}
		}
	}

	log.Debug().
		Int64("user_id", userID).
		Int32("subforum_id", subforumID).
		Msg("User does not have access to private subforum")
	return false, nil
}

// CanAccessPrivateSubforumWithActivePseudonym checks if a user can access a private subforum
// This method checks ONLY the active pseudonym (secure behavior)
func (dao *PermissionDAO) CanAccessPrivateSubforumWithActivePseudonym(ctx context.Context, userID int64, subforumID int32, activePseudonymID string) (bool, error) {
	// Check if the active pseudonym is a moderator of this subforum
	moderator, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
		models.SelectWhere.SubforumModerators.PseudonymID.EQ(activePseudonymID),
	).One(ctx, dao.db)

	if err == nil && moderator != nil {
		log.Debug().
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Str("role", moderator.Role).
			Msg("Active pseudonym is moderator of private subforum")
		return true, nil
	}

	// Check if user has platform-wide roles that grant access
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	platformWideRoles := []string{"platform_admin", "trust_safety", "legal_team"}
	if user.Roles.Valid {
		rawValue, err := user.Roles.V.Value()
		if err != nil {
			return false, fmt.Errorf("failed to get user roles value: %w", err)
		}
		var roles []string
		if err := json.Unmarshal(rawValue.([]byte), &roles); err == nil {
			for _, role := range roles {
				for _, platformRole := range platformWideRoles {
					if role == platformRole {
						log.Debug().
							Int64("user_id", userID).
							Int32("subforum_id", subforumID).
							Str("role", role).
							Msg("User has platform-wide role for private subforum access")
						return true, nil
					}
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

	// Check if any of the user's pseudonyms are moderators with the required capability
	for _, mapping := range mappings {
		moderator, err := models.SubforumModerators.Query(
			models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
			models.SelectWhere.SubforumModerators.PseudonymID.EQ(mapping.PseudonymID),
		).One(ctx, dao.db)

		if err == nil && moderator != nil {
			// Check if moderator has the specific capability in their permissions
			if moderator.Permissions.Valid {
				rawValue, err := moderator.Permissions.V.Value()
				if err != nil {
					return false, fmt.Errorf("failed to get moderator permissions value: %w", err)
				}
				var permissions []string
				if err := json.Unmarshal(rawValue.([]byte), &permissions); err == nil {
					for _, perm := range permissions {
						if perm == capability {
							log.Debug().
								Int64("user_id", userID).
								Int32("subforum_id", subforumID).
								Str("capability", capability).
								Str("role", moderator.Role).
								Str("pseudonym_id", mapping.PseudonymID).
								Msg("User has subforum-specific capability")
							return true, nil
						}
					}
				}
			}

			// Check role-based capabilities
			roleCapabilities := dao.getRoleCapabilities(moderator.Role)
			for _, cap := range roleCapabilities {
				if cap == capability {
					log.Debug().
						Int64("user_id", userID).
						Int32("subforum_id", subforumID).
						Str("capability", capability).
						Str("role", moderator.Role).
						Str("pseudonym_id", mapping.PseudonymID).
						Msg("User has role-based capability")
					return true, nil
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

	// First check if the active pseudonym is a moderator of this subforum
	moderator, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
		models.SelectWhere.SubforumModerators.PseudonymID.EQ(activePseudonymID),
	).One(ctx, dao.db)

	if err != nil {
		log.Debug().
			Err(err).
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Msg("Error querying moderator record")
	} else if moderator != nil {
		log.Debug().
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Str("role", moderator.Role).
			Msg("Found moderator record")

		// Check if the moderator role has this capability
		roleCaps := dao.getRoleCapabilities(moderator.Role)
		log.Debug().
			Str("role", moderator.Role).
			Interface("role_capabilities", roleCaps).
			Msg("Role capabilities")

		for _, cap := range roleCaps {
			if cap == capability {
				log.Debug().
					Int64("user_id", userID).
					Int32("subforum_id", subforumID).
					Str("pseudonym_id", activePseudonymID).
					Str("capability", capability).
					Str("role", moderator.Role).
					Msg("Active pseudonym has capability through moderator role")
				return true, nil
			}
		}

		// Check specific permissions from JSON
		if moderator.Permissions.Valid {
			rawValue, err := moderator.Permissions.V.Value()
			if err != nil {
				return false, fmt.Errorf("failed to get moderator permissions value: %w", err)
			}
			var permissions []string
			if err := json.Unmarshal(rawValue.([]byte), &permissions); err == nil {
				for _, perm := range permissions {
					if perm == capability {
						log.Debug().
							Int64("user_id", userID).
							Int32("subforum_id", subforumID).
							Str("pseudonym_id", activePseudonymID).
							Str("capability", capability).
							Msg("Active pseudonym has capability through specific permission")
						return true, nil
					}
				}
			}
		}
	} else {
		log.Debug().
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("pseudonym_id", activePseudonymID).
			Msg("No moderator record found")
	}

	// Check if the active pseudonym has platform-wide capabilities
	pseudonym, err := models.FindPseudonym(ctx, dao.db, activePseudonymID)
	if err != nil {
		return false, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		return false, fmt.Errorf("pseudonym not found")
	}

	if pseudonym.Capabilities.Valid {
		rawValue, err := pseudonym.Capabilities.V.Value()
		if err != nil {
			return false, fmt.Errorf("failed to get pseudonym capabilities: %w", err)
		}
		var capabilities []string
		if err := json.Unmarshal(rawValue.([]byte), &capabilities); err == nil {
			for _, cap := range capabilities {
				if cap == capability {
					log.Debug().
						Int64("user_id", userID).
						Int32("subforum_id", subforumID).
						Str("pseudonym_id", activePseudonymID).
						Str("capability", capability).
						Msg("Active pseudonym has platform-wide capability")
					return true, nil
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

	// Get subforum moderator roles for each pseudonym
	for _, mapping := range mappings {
		moderator, err := models.SubforumModerators.Query(
			models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
			models.SelectWhere.SubforumModerators.PseudonymID.EQ(mapping.PseudonymID),
		).One(ctx, dao.db)

		if err == nil && moderator != nil {
			roles = append(roles, moderator.Role)
		}
	}

	// Get platform-wide roles
	if user.Roles.Valid {
		rawValue, err := user.Roles.V.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get user roles value: %w", err)
		}
		var userRoles []string
		if err := json.Unmarshal(rawValue.([]byte), &userRoles); err == nil {
			roles = append(roles, userRoles...)
		}
	}

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

	// Check if any of the user's pseudonyms are moderators of this subforum
	for _, mapping := range mappings {
		moderator, err := models.SubforumModerators.Query(
			models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
			models.SelectWhere.SubforumModerators.PseudonymID.EQ(mapping.PseudonymID),
		).One(ctx, dao.db)

		if err == nil && moderator != nil {
			// Add role-based capabilities
			roleCaps := dao.getRoleCapabilities(moderator.Role)
			capabilities = append(capabilities, roleCaps...)

			// Add specific permissions from JSON
			if moderator.Permissions.Valid {
				rawValue, err := moderator.Permissions.V.Value()
				if err != nil {
					return nil, fmt.Errorf("failed to get moderator permissions value: %w", err)
				}
				var permissions []string
				if err := json.Unmarshal(rawValue.([]byte), &permissions); err == nil {
					capabilities = append(capabilities, permissions...)
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
	var roles []string
	var capabilities []string

	// Get the active pseudonym
	pseudonym, err := models.FindPseudonym(ctx, dao.db, activePseudonymID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		return nil, nil, fmt.Errorf("pseudonym not found")
	}

	// Get pseudonym roles and capabilities
	if pseudonym.Roles.Valid {
		rawValue, err := pseudonym.Roles.V.Value()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get pseudonym roles: %w", err)
		}
		if err := json.Unmarshal(rawValue.([]byte), &roles); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal pseudonym roles: %w", err)
		}
	} else {
		// Default role for all pseudonyms
		roles = []string{"user"}
	}

	if pseudonym.Capabilities.Valid {
		rawValue, err := pseudonym.Capabilities.V.Value()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get pseudonym capabilities: %w", err)
		}
		if err := json.Unmarshal(rawValue.([]byte), &capabilities); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal pseudonym capabilities: %w", err)
		}
	} else {
		// Default capabilities for all pseudonyms
		capabilities = []string{"create_content", "vote", "message", "report"}
	}

	return roles, capabilities, nil
}

// removeDuplicates removes duplicate strings from a slice
func (dao *PermissionDAO) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
