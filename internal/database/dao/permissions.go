package dao

import (
	"context"
	"encoding/json"
	"fmt"

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
func (dao *PermissionDAO) CanAccessPrivateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	// First, check if user is a moderator of this subforum
	moderator, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
		models.SelectWhere.SubforumModerators.UserID.EQ(userID),
	).One(ctx, dao.db)

	if err == nil && moderator != nil {
		log.Debug().
			Int64("user_id", userID).
			Int32("subforum_id", subforumID).
			Str("role", moderator.Role).
			Msg("User is moderator of private subforum")
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

	// Check for platform-wide roles that grant access to private subforums
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

	// Check if user has specific capabilities that grant access
	accessCapabilities := []string{"access_private_subforums", "system_admin", "cross_platform_access"}
	if user.Capabilities.Valid {
		rawValue, err := user.Capabilities.V.Value()
		if err != nil {
			return false, fmt.Errorf("failed to get user capabilities value: %w", err)
		}
		var capabilities []string
		if err := json.Unmarshal(rawValue.([]byte), &capabilities); err == nil {
			for _, capability := range capabilities {
				for _, accessCap := range accessCapabilities {
					if capability == accessCap {
						log.Debug().
							Int64("user_id", userID).
							Int32("subforum_id", subforumID).
							Str("capability", capability).
							Msg("User has capability for private subforum access")
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

// HasSubforumCapability checks if a user has a specific capability for a subforum
func (dao *PermissionDAO) HasSubforumCapability(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
	// First, check subforum-specific moderator permissions
	moderator, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
		models.SelectWhere.SubforumModerators.UserID.EQ(userID),
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
					Msg("User has role-based capability")
				return true, nil
			}
		}
	}

	// Check platform-wide user capabilities
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	if user.Capabilities.Valid {
		rawValue, err := user.Capabilities.V.Value()
		if err != nil {
			return false, fmt.Errorf("failed to get user capabilities value: %w", err)
		}
		var capabilities []string
		if err := json.Unmarshal(rawValue.([]byte), &capabilities); err == nil {
			for _, userCap := range capabilities {
				if userCap == capability {
					log.Debug().
						Int64("user_id", userID).
						Int32("subforum_id", subforumID).
						Str("capability", capability).
						Msg("User has platform-wide capability")
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

// GetUserSubforumRoles returns the roles a user has for a specific subforum
func (dao *PermissionDAO) GetUserSubforumRoles(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	var roles []string

	// Get subforum moderator role
	moderator, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
		models.SelectWhere.SubforumModerators.UserID.EQ(userID),
	).One(ctx, dao.db)

	if err == nil && moderator != nil {
		roles = append(roles, moderator.Role)
	}

	// Get platform-wide roles
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user != nil && user.Roles.Valid {
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

// GetUserSubforumCapabilities returns all capabilities a user has for a specific subforum
func (dao *PermissionDAO) GetUserSubforumCapabilities(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	var capabilities []string

	// Get subforum moderator capabilities
	moderator, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
		models.SelectWhere.SubforumModerators.UserID.EQ(userID),
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

	// Get platform-wide capabilities
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user != nil && user.Capabilities.Valid {
		rawValue, err := user.Capabilities.V.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get user capabilities value: %w", err)
		}
		var userCapabilities []string
		if err := json.Unmarshal(rawValue.([]byte), &userCapabilities); err == nil {
			capabilities = append(capabilities, userCapabilities...)
		}
	}

	return capabilities, nil
}

// CanModerateSubforum checks if a user can moderate a specific subforum
func (dao *PermissionDAO) CanModerateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, "moderate_content")
}

// CanBanUsers checks if a user can ban users in a specific subforum
func (dao *PermissionDAO) CanBanUsers(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, "ban_users")
}

// CanRemoveContent checks if a user can remove content in a specific subforum
func (dao *PermissionDAO) CanRemoveContent(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, "remove_content")
}

// CanManageModerators checks if a user can manage moderators in a specific subforum
func (dao *PermissionDAO) CanManageModerators(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return dao.HasSubforumCapability(ctx, userID, subforumID, "manage_moderators")
}

// getRoleCapabilities returns the capabilities associated with a specific role
func (dao *PermissionDAO) getRoleCapabilities(role string) []string {
	roleCapabilities := map[string][]string{
		"owner": {
			"moderate_content",
			"ban_users",
			"remove_content",
			"correlate_fingerprints",
			"manage_moderators",
			"access_private_subforums",
		},
		"moderator": {
			"moderate_content",
			"ban_users",
			"remove_content",
			"correlate_fingerprints",
		},
		"junior_moderator": {
			"moderate_content",
			"remove_content",
		},
	}

	if caps, exists := roleCapabilities[role]; exists {
		return caps
	}
	return []string{}
}
