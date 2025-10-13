package appview

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// AuthMiddleware handles authentication for AppView endpoints
type AuthMiddleware struct {
	rbacService *RBACService
	logger      *slog.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(rbacService *RBACService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		rbacService: rbacService,
		logger:      logger,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token and get user context
		userCtx, err := m.rbacService.ValidateToken(r.Context(), token)
		if err != nil {
			m.logger.Error("Token validation failed", "error", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user context to request
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequirePermission middleware that requires a specific permission
func (m *AuthMiddleware) RequirePermission(permission string, getSubforumID func(*http.Request) *string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user context from request
			userCtx, ok := r.Context().Value(UserContextKey).(*UserContext)
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Get subforum ID if needed
			var subforumID *string
			if getSubforumID != nil {
				subforumID = getSubforumID(r)
			}

			// Check permission
			hasPermission, err := m.rbacService.CheckPermission(r.Context(), userCtx.Did, permission, subforumID)
			if err != nil {
				m.logger.Error("Permission check failed", "error", err, "user", userCtx.Did, "permission", permission)
				http.Error(w, "Permission check failed", http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				m.logger.Warn("Permission denied", "user", userCtx.Did, "permission", permission, "subforum_id", subforumID)
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireRole middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(roleName string, getSubforumID func(*http.Request) *string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user context from request
			userCtx, ok := r.Context().Value(UserContextKey).(*UserContext)
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Get subforum ID if needed
			var subforumID *string
			if getSubforumID != nil {
				subforumID = getSubforumID(r)
			}

			// Check role
			hasRole, err := m.rbacService.HasRole(r.Context(), userCtx.Did, roleName, subforumID)
			if err != nil {
				m.logger.Error("Role check failed", "error", err, "user", userCtx.Did, "role", roleName)
				http.Error(w, "Role check failed", http.StatusInternalServerError)
				return
			}

			if !hasRole {
				m.logger.Warn("Role required", "user", userCtx.Did, "role", roleName, "subforum_id", subforumID)
				http.Error(w, "Insufficient role", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// GetUserContext extracts user context from request
func GetUserContext(r *http.Request) *UserContext {
	userCtx, ok := r.Context().Value(UserContextKey).(*UserContext)
	if !ok {
		return nil
	}
	return userCtx
}

// Helper functions for common permission checks
func (m *AuthMiddleware) RequirePlatformAdmin(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole("platform_admin", nil)(next)
}

func (m *AuthMiddleware) RequireSubforumOwner(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole("subforum_owner", func(r *http.Request) *string {
		// Extract subforum ID from URL path
		// For routes like /api/v1/subforums/{slug}/posts, extract the slug
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/subforums/")
		if path == "" {
			return nil
		}

		// Split by '/' to get the first part (slug)
		parts := strings.Split(path, "/")
		if len(parts) > 0 && parts[0] != "" {
			// Convert slug to subforum ID by looking it up in the database
			// This is a simplified approach - in production, you'd want to cache this
			subforumID := m.getSubforumIDBySlug(r.Context(), parts[0])
			return subforumID
		}

		return nil
	})(next)
}

func (m *AuthMiddleware) RequireSubforumModerator(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole("subforum_moderator", func(r *http.Request) *string {
		// Extract subforum ID from URL path
		// For routes like /api/v1/subforums/{slug}/posts, extract the slug
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/subforums/")
		if path == "" {
			return nil
		}

		// Split by '/' to get the first part (slug)
		parts := strings.Split(path, "/")
		if len(parts) > 0 && parts[0] != "" {
			// Convert slug to subforum ID by looking it up in the database
			// This is a simplified approach - in production, you'd want to cache this
			subforumID := m.getSubforumIDBySlug(r.Context(), parts[0])
			return subforumID
		}

		return nil
	})(next)
}
