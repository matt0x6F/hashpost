package appview

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/bluesky-social/indigo/atproto/auth" // Register JWT signing methods
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	appview "github.com/matt0x6f/hashpost/internal/database/generated/appview"
	"github.com/matt0x6f/hashpost/internal/pds"
)

// RBACService handles role-based access control for AppView
type RBACService struct {
	db             *pgxpool.Pool
	queries        *appview.Queries
	directory      identity.Directory
	logger         *slog.Logger
	tokenValidator *pds.MultiPDSTokenValidator
}

// NewRBACService creates a new RBAC service
func NewRBACService(db *pgxpool.Pool, logger *slog.Logger) *RBACService {
	// Use the same identity directory as PDS for consistency
	directory := identity.NewMockDirectory()

	// Add the same test identities as PDS
	testUser := identity.Identity{
		DID:    syntax.DID("did:plc:hashpost-binding-test"),
		Handle: syntax.Handle("testuser.hashpost.local"),
	}
	adminUser := identity.Identity{
		DID:    syntax.DID("did:plc:hashpost-admin-test"),
		Handle: syntax.Handle("admin.hashpost.local"),
	}

	directory.Insert(testUser)
	directory.Insert(adminUser)

	// Create queries instance
	queries := appview.New(db)

	// Create multi-PDS token validator
	tokenValidator := pds.NewMultiPDSTokenValidator(logger)

	return &RBACService{
		db:             db,
		queries:        queries,
		directory:      &directory,
		logger:         logger,
		tokenValidator: tokenValidator,
	}
}

// UserContext represents the authenticated user context
type UserContext struct {
	Did      string     `json:"did"`
	Handle   string     `json:"handle"`
	Roles    []UserRole `json:"roles"`
	IsActive bool       `json:"is_active"`
}

// ValidateToken validates a JWT token from any PDS and returns user context
func (r *RBACService) ValidateToken(ctx context.Context, tokenString string) (*UserContext, error) {
	r.logger.Debug("Validating JWT token", "token_length", len(tokenString))

	// Use multi-PDS token validator
	validationResult, err := r.tokenValidator.ValidateTokenFromAnyPDS(ctx, tokenString)
	if err != nil {
		r.logger.Error("Token validation failed", "error", err)
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Ensure user exists in AppView database (create lightweight record if needed)
	_, err = r.ensureUserExists(ctx, validationResult.DID, validationResult.Handle, validationResult.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user exists: %w", err)
	}

	// Get user roles from database
	roles, err := r.getUserRoles(ctx, validationResult.DID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return &UserContext{
		Did:      validationResult.DID,
		Handle:   validationResult.Handle,
		Roles:    roles,
		IsActive: true,
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (r *RBACService) CheckPermission(ctx context.Context, userDID, permission string, subforumID *string) (bool, error) {
	var subforumUUID pgtype.UUID
	if subforumID != nil {
		if uuid, err := uuid.Parse(*subforumID); err == nil {
			subforumUUID = pgtype.UUID{Bytes: uuid, Valid: true}
		}
	}

	hasPermission, err := r.queries.CheckUserPermission(ctx, &appview.CheckUserPermissionParams{
		UserDid:    userDID,
		Name:       permission,
		SubforumID: subforumUUID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}

// getUserRoles retrieves all active roles for a user
func (r *RBACService) getUserRoles(ctx context.Context, userDID string) ([]UserRole, error) {
	rows, err := r.queries.GetUserRoles(ctx, userDID)
	if err != nil {
		return nil, err
	}

	var roles []UserRole
	for _, row := range rows {
		role := UserRole{
			RoleName: row.RoleName,
		}

		// Handle optional fields
		if row.IsPlatformRole != nil {
			role.IsPlatformRole = *row.IsPlatformRole
		}
		if row.SubforumID.Valid {
			uuid := uuid.UUID(row.SubforumID.Bytes)
			role.SubforumId = &uuid
		}
		if row.SubforumSlug != nil {
			role.SubforumSlug = row.SubforumSlug
		}
		if row.SubforumName != nil {
			role.SubforumName = row.SubforumName
		}
		if row.ExpiresAt.Valid {
			role.ExpiresAt = &row.ExpiresAt.Time
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// AssignRole assigns a role to a user
func (r *RBACService) AssignRole(ctx context.Context, userDID, roleName string, subforumID *string, grantedBy string, expiresAt *string) error {
	var subforumUUID pgtype.UUID
	if subforumID != nil {
		if uuid, err := uuid.Parse(*subforumID); err == nil {
			subforumUUID = pgtype.UUID{Bytes: uuid, Valid: true}
		}
	}

	var expiresAtTime pgtype.Timestamptz
	if expiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *expiresAt); err == nil {
			expiresAtTime = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	_, err := r.queries.AssignUserRole(ctx, &appview.AssignUserRoleParams{
		UserDid:    userDID,
		Name:       roleName,
		SubforumID: subforumUUID,
		GrantedBy:  grantedBy,
		ExpiresAt:  expiresAtTime,
	})

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	r.logger.Info("Role assigned", "user_did", userDID, "role", roleName, "subforum_id", subforumID, "granted_by", grantedBy)
	return nil
}

// RevokeRole revokes a role from a user
func (r *RBACService) RevokeRole(ctx context.Context, userDID, roleName string, subforumID *string) error {
	var subforumUUID pgtype.UUID
	if subforumID != nil {
		if uuid, err := uuid.Parse(*subforumID); err == nil {
			subforumUUID = pgtype.UUID{Bytes: uuid, Valid: true}
		}
	}

	err := r.queries.RevokeUserRole(ctx, &appview.RevokeUserRoleParams{
		UserDid:    userDID,
		Name:       roleName,
		SubforumID: subforumUUID,
	})

	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	r.logger.Info("Role revoked", "user_did", userDID, "role", roleName, "subforum_id", subforumID)
	return nil
}

// HasRole checks if a user has a specific role
func (r *RBACService) HasRole(ctx context.Context, userDID, roleName string, subforumID *string) (bool, error) {
	var subforumUUID pgtype.UUID
	if subforumID != nil {
		if uuid, err := uuid.Parse(*subforumID); err == nil {
			subforumUUID = pgtype.UUID{Bytes: uuid, Valid: true}
		}
	}

	hasRole, err := r.queries.HasUserRole(ctx, &appview.HasUserRoleParams{
		UserDid:    userDID,
		Name:       roleName,
		SubforumID: subforumUUID,
	})
	if err != nil {
		return false, err
	}

	return hasRole, nil
}

// IsPlatformAdmin checks if a user is a platform admin
func (r *RBACService) IsPlatformAdmin(ctx context.Context, userDID string) (bool, error) {
	return r.HasRole(ctx, userDID, "platform_admin", nil)
}

// IsSubforumOwner checks if a user is the owner of a specific subforum
func (r *RBACService) IsSubforumOwner(ctx context.Context, userDID string, subforumID string) (bool, error) {
	return r.HasRole(ctx, userDID, "subforum_owner", &subforumID)
}

// IsSubforumModerator checks if a user is a moderator of a specific subforum
func (r *RBACService) IsSubforumModerator(ctx context.Context, userDID string, subforumID string) (bool, error) {
	return r.HasRole(ctx, userDID, "subforum_moderator", &subforumID)
}

// GetUsersWithRoles retrieves users with their roles
func (r *RBACService) GetUsersWithRoles(ctx context.Context, limit, offset int, subforumID string) ([]UserWithRoles, error) {
	var subforumIDStr string
	if subforumID != "" {
		subforumIDStr = subforumID
	}

	r.logger.Debug("Getting users with roles", "subforumID", subforumIDStr, "limit", limit, "offset", offset)

	rows, err := r.queries.GetUsersWithRoles(ctx, &appview.GetUsersWithRolesParams{
		Column1: subforumIDStr,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		r.logger.Error("Failed to get users with roles", "error", err)
		return nil, err
	}

	r.logger.Debug("Query returned users", "count", len(rows))

	var users []UserWithRoles
	for _, row := range rows {
		r.logger.Debug("Processing user", "user_did", row.UserDid, "handle", row.Handle, "role_count", row.RoleCount)
		user := UserWithRoles{
			UserDid:   row.UserDid,
			Handle:    row.Handle,
			RoleCount: int(row.RoleCount),
		}

		users = append(users, user)
	}

	r.logger.Debug("Returning users", "count", len(users))
	return users, nil
}

// GetSubforumMembers retrieves members of a specific subforum with their roles
func (r *RBACService) GetSubforumMembers(ctx context.Context, slug string, limit, offset int) ([]map[string]interface{}, error) {
	rows, err := r.queries.GetSubforumMembers(ctx, &appview.GetSubforumMembersParams{
		Slug:   slug,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subforum members: %w", err)
	}

	var members []map[string]interface{}
	for _, row := range rows {
		members = append(members, map[string]interface{}{
			"user_did":   row.UserDid,
			"role_count": row.RoleCount,
		})
	}

	return members, nil
}

// AssignSubforumRole assigns a role to a user in a specific subforum
func (r *RBACService) AssignSubforumRole(ctx context.Context, slug, userDid, roleName, assignedBy string, expiresAt *string) error {
	var expiresAtTime pgtype.Timestamptz
	if expiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *expiresAt); err == nil {
			expiresAtTime = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	_, err := r.queries.AssignSubforumRole(ctx, &appview.AssignSubforumRoleParams{
		UserDid:   userDid,
		Name:      roleName,
		Slug:      slug,
		GrantedBy: assignedBy,
		ExpiresAt: expiresAtTime,
	})

	if err != nil {
		return fmt.Errorf("failed to assign subforum role: %w", err)
	}

	return nil
}

// RevokeSubforumRole revokes a role from a user in a specific subforum
func (r *RBACService) RevokeSubforumRole(ctx context.Context, slug, userDid, roleName, revokedBy string) error {
	err := r.queries.RevokeSubforumRole(ctx, &appview.RevokeSubforumRoleParams{
		UserDid: userDid,
		Name:    roleName,
		Slug:    slug,
	})

	if err != nil {
		return fmt.Errorf("failed to revoke subforum role: %w", err)
	}

	return nil
}

// GetUserPermissions retrieves all permissions for a user
func (r *RBACService) GetUserPermissions(ctx context.Context, userDid string, subforumID *string) ([]string, error) {
	var subforumIDStr string
	if subforumID != nil {
		subforumIDStr = *subforumID
	}

	permissions, err := r.queries.GetUserPermissions(ctx, &appview.GetUserPermissionsParams{
		UserDid: userDid,
		Column2: subforumIDStr,
	})
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetAllRoles retrieves all available roles
func (r *RBACService) GetAllRoles(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.queries.GetAllRoles(ctx)
	if err != nil {
		return nil, err
	}

	var roles []map[string]interface{}
	for _, row := range rows {
		roles = append(roles, map[string]interface{}{
			"id":               row.ID,
			"name":             row.Name,
			"description":      row.Description,
			"is_platform_role": row.IsPlatformRole,
			"created_at":       row.CreatedAt,
		})
	}

	return roles, nil
}

// GetAllPermissions retrieves all available permissions
func (r *RBACService) GetAllPermissions(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.queries.GetAllPermissions(ctx)
	if err != nil {
		return nil, err
	}

	var permissions []map[string]interface{}
	for _, row := range rows {
		permissions = append(permissions, map[string]interface{}{
			"id":            row.ID,
			"name":          row.Name,
			"description":   row.Description,
			"resource_type": row.ResourceType,
			"created_at":    row.CreatedAt,
		})
	}

	return permissions, nil
}

// GetUserRoles retrieves roles for a specific user
func (r *RBACService) GetUserRoles(ctx context.Context, userDID string, subforumID *string) ([]UserRole, error) {
	// For now, we'll use the existing getUserRoles method since it doesn't filter by subforum
	// In the future, we could create a separate SQLC query for this
	return r.getUserRoles(ctx, userDID)
}

// ensureUserExists ensures a user exists in the AppView database, creating a lightweight record if needed
func (r *RBACService) ensureUserExists(ctx context.Context, did, handle, issuer string) (*AppViewUser, error) {
	// Check if user already exists
	existingUser, err := r.queries.GetUserByDID(ctx, did)
	if err == nil && existingUser != nil {
		// User exists, update last seen time if external
		if existingUser.PdsSource != nil && *existingUser.PdsSource != "http://hashpost-pds:8080" {
			// For now, just log that we would update last seen time
			// In production, we'd need to add this method to SQLC
			r.logger.Debug("Would update last seen time for external user", "did", did)
		}

		// Convert to AppViewUser type
		user := &AppViewUser{
			ID:          existingUser.ID,
			DID:         existingUser.Did,
			Handle:      existingUser.Handle,
			DisplayName: getStringValue(existingUser.DisplayName),
			CreatedAt:   existingUser.CreatedAt.Time,
			UpdatedAt:   existingUser.UpdatedAt.Time,
			PDSSource:   existingUser.PdsSource,
			LastSeenAt:  getTimeValue(existingUser.LastSeenAt),
		}
		return user, nil
	}

	// User doesn't exist, create lightweight record
	r.logger.Info("Creating lightweight user record", "did", did, "handle", handle, "issuer", issuer)

	// Determine PDS source
	pdsSource := r.determinePDSSource(issuer)

	// For now, we'll use a simplified approach since the SQLC methods might not exist yet
	// In production, we'd need to add the proper SQLC queries for user creation
	r.logger.Info("User creation not yet implemented in SQLC", "did", did, "handle", handle, "pds_source", pdsSource)

	// Return a mock user for now
	user := &AppViewUser{
		ID:          uuid.New(),
		DID:         did,
		Handle:      handle,
		DisplayName: handle,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		PDSSource:   pdsSource,
	}

	// Set last seen time for external users
	if pdsSource != nil && *pdsSource != "http://hashpost-pds:8080" {
		now := time.Now()
		user.LastSeenAt = &now
	}

	return user, nil
}

// determinePDSSource determines the PDS source for a given issuer
func (r *RBACService) determinePDSSource(issuer string) *string {
	// For now, we'll use a simple heuristic
	// In production, this would check against our PDS server endpoint
	localIdentifiers := []string{
		"localhost",
		"127.0.0.1",
		"hashpost",
	}

	for _, identifier := range localIdentifiers {
		if strings.Contains(issuer, identifier) {
			// This is our local PDS
			localPDS := "http://hashpost-pds:8080"
			return &localPDS
		}
	}

	// This is an external PDS
	return &issuer
}

// getStringValue safely extracts string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getTimeValue safely extracts time value from pgtype.Timestamptz
func getTimeValue(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}
