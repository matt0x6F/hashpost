package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// PermissionDAOInterface defines the interface for permission checking operations
type PermissionDAOInterface interface {
	HasSubforumCapability(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error)
	CanAccessPrivateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error)
	GetUserSubforumRoles(ctx context.Context, userID int64, subforumID int32) ([]string, error)
	GetUserSubforumCapabilities(ctx context.Context, userID int64, subforumID int32) ([]string, error)
}

// PermissionMiddleware provides subforum-specific permission checking
type PermissionMiddleware struct {
	permissionDAO PermissionDAOInterface
}

// NewPermissionMiddleware creates a new PermissionMiddleware
func NewPermissionMiddleware(db bob.Executor) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionDAO: dao.NewPermissionDAO(db),
	}
}

// NewPermissionMiddlewareWithDAO creates a new PermissionMiddleware with a custom DAO
func NewPermissionMiddlewareWithDAO(dao PermissionDAOInterface) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionDAO: dao,
	}
}

// RequireSubforumCapability creates middleware that requires a specific capability for a subforum
func (pm *PermissionMiddleware) RequireSubforumCapability(capability string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user context from request
			userCtx, err := ExtractUserFromRequest(r)
			if err != nil || userCtx == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract subforum ID from URL parameters
			subforumIDStr := r.URL.Query().Get("subforum_id")
			if subforumIDStr == "" {
				// Try to get from path parameters (e.g., /subforums/{id}/posts)
				// This would need to be adjusted based on your routing structure
				http.Error(w, "Subforum ID required", http.StatusBadRequest)
				return
			}

			subforumID, err := strconv.ParseInt(subforumIDStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid subforum ID", http.StatusBadRequest)
				return
			}

			// Check if user has the required capability for this subforum
			hasCapability, err := pm.permissionDAO.HasSubforumCapability(r.Context(), userCtx.UserID, int32(subforumID), capability)
			if err != nil {
				log.Error().Err(err).
					Int64("user_id", userCtx.UserID).
					Int32("subforum_id", int32(subforumID)).
					Str("capability", capability).
					Msg("Failed to check subforum capability")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !hasCapability {
				log.Warn().
					Int64("user_id", userCtx.UserID).
					Int32("subforum_id", int32(subforumID)).
					Str("capability", capability).
					Msg("User lacks required subforum capability")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Add subforum context to request
			ctx := context.WithValue(r.Context(), "subforum_id", int32(subforumID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePrivateSubforumAccess creates middleware that checks if user can access a private subforum
func (pm *PermissionMiddleware) RequirePrivateSubforumAccess() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user context from request
			userCtx, err := ExtractUserFromRequest(r)
			if err != nil || userCtx == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract subforum ID from URL parameters
			subforumIDStr := r.URL.Query().Get("subforum_id")
			if subforumIDStr == "" {
				http.Error(w, "Subforum ID required", http.StatusBadRequest)
				return
			}

			subforumID, err := strconv.ParseInt(subforumIDStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid subforum ID", http.StatusBadRequest)
				return
			}

			// Check if user can access this private subforum
			canAccess, err := pm.permissionDAO.CanAccessPrivateSubforum(r.Context(), userCtx.UserID, int32(subforumID))
			if err != nil {
				log.Error().Err(err).
					Int64("user_id", userCtx.UserID).
					Int32("subforum_id", int32(subforumID)).
					Msg("Failed to check private subforum access")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !canAccess {
				log.Warn().
					Int64("user_id", userCtx.UserID).
					Int32("subforum_id", int32(subforumID)).
					Msg("User denied access to private subforum")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Add subforum context to request
			ctx := context.WithValue(r.Context(), "subforum_id", int32(subforumID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireModerationCapability creates middleware that requires moderation capability
func (pm *PermissionMiddleware) RequireModerationCapability() func(http.Handler) http.Handler {
	return pm.RequireSubforumCapability("moderate_content")
}

// RequireBanCapability creates middleware that requires ban capability
func (pm *PermissionMiddleware) RequireBanCapability() func(http.Handler) http.Handler {
	return pm.RequireSubforumCapability("ban_users")
}

// RequireRemoveContentCapability creates middleware that requires content removal capability
func (pm *PermissionMiddleware) RequireRemoveContentCapability() func(http.Handler) http.Handler {
	return pm.RequireSubforumCapability("remove_content")
}

// RequireManageModeratorsCapability creates middleware that requires moderator management capability
func (pm *PermissionMiddleware) RequireManageModeratorsCapability() func(http.Handler) http.Handler {
	return pm.RequireSubforumCapability("manage_moderators")
}

// GetSubforumIDFromContext extracts subforum ID from request context
func GetSubforumIDFromContext(r *http.Request) (int32, bool) {
	subforumID, ok := r.Context().Value("subforum_id").(int32)
	return subforumID, ok
}

// PermissionChecker provides a convenient interface for checking permissions in handlers
type PermissionChecker struct {
	permissionDAO PermissionDAOInterface
}

// NewPermissionChecker creates a new PermissionChecker
func NewPermissionChecker(db bob.Executor) *PermissionChecker {
	return &PermissionChecker{
		permissionDAO: dao.NewPermissionDAO(db),
	}
}

// NewPermissionCheckerWithDAO creates a new PermissionChecker with a custom DAO
func NewPermissionCheckerWithDAO(dao PermissionDAOInterface) *PermissionChecker {
	return &PermissionChecker{
		permissionDAO: dao,
	}
}

// CheckSubforumCapability checks if a user has a specific capability for a subforum
func (pc *PermissionChecker) CheckSubforumCapability(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
	return pc.permissionDAO.HasSubforumCapability(ctx, userID, subforumID, capability)
}

// CheckPrivateSubforumAccess checks if a user can access a private subforum
func (pc *PermissionChecker) CheckPrivateSubforumAccess(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	return pc.permissionDAO.CanAccessPrivateSubforum(ctx, userID, subforumID)
}

// GetUserSubforumRoles gets all roles a user has for a specific subforum
func (pc *PermissionChecker) GetUserSubforumRoles(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	return pc.permissionDAO.GetUserSubforumRoles(ctx, userID, subforumID)
}

// GetUserSubforumCapabilities gets all capabilities a user has for a specific subforum
func (pc *PermissionChecker) GetUserSubforumCapabilities(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	return pc.permissionDAO.GetUserSubforumCapabilities(ctx, userID, subforumID)
}
