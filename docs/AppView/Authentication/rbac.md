# RBAC System

## Overview

The AppView implements a comprehensive Role-Based Access Control (RBAC) system using SQLC-generated queries. It provides hierarchical permissions with platform-level and subforum-specific roles.

## Implementation

### RBAC Service

**File**: `internal/appview/rbac.go`  
**Service**: `RBACService`

The RBAC service handles role and permission management:

```go
type RBACService struct {
    queries *generated.Queries
    logger  *slog.Logger
}

func NewRBACService(queries *generated.Queries, logger *slog.Logger) *RBACService {
    return &RBACService{
        queries: queries,
        logger:  logger,
    }
}
```

### Permission Checking

**Method**: `CheckUserPermission(ctx context.Context, userDID, permission string, subforumID *string) (bool, error)`

```go
func (r *RBACService) CheckUserPermission(ctx context.Context, userDID, permission string, subforumID *string) (bool, error) {
    result, err := r.queries.CheckUserPermission(ctx, generated.CheckUserPermissionParams{
        UserDid:      userDID,
        Permission:   permission,
        SubforumID:   subforumID,
    })
    if err != nil {
        r.logger.Error("Failed to check user permission", "error", err, "user_did", userDID, "permission", permission)
        return false, fmt.Errorf("failed to check permission: %w", err)
    }

    return result.HasPermission, nil
}
```

## Database Schema

### Roles Table

**Table**: `roles`  
**Purpose**: Define available roles in the system

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_platform_role BOOLEAN DEFAULT FALSE, -- true for platform admin, false for subforum-specific
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Permissions Table

**Table**: `permissions`  
**Purpose**: Define available permissions

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource_type VARCHAR(50) NOT NULL, -- 'platform', 'subforum', 'post', 'comment'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Role Permissions Table

**Table**: `role_permissions`  
**Purpose**: Define which permissions each role has

```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);
```

### User Roles Table

**Table**: `user_roles`  
**Purpose**: Assign roles to users

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_did VARCHAR(255) NOT NULL, -- Reference to user's DID
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    subforum_id UUID REFERENCES appview_subforums(id) ON DELETE CASCADE, -- NULL for platform roles
    granted_by VARCHAR(255) NOT NULL, -- DID of user who granted the role
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ, -- NULL for permanent roles
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_did, role_id, subforum_id)
);
```

## SQLC Queries

### Permission Checking

**File**: `internal/database/queries/appview/check_user_permission.sql`

```sql
-- name: CheckUserPermission :one
SELECT EXISTS(
    SELECT 1
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_did = $1
        AND p.name = $2
        AND ur.is_active = TRUE
        AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
        AND (
            -- Platform permissions (no subforum restriction)
            (p.resource_type = 'platform' AND ur.subforum_id IS NULL)
            OR
            -- Subforum permissions (specific subforum or any subforum)
            (p.resource_type IN ('subforum', 'post', 'comment', 'vote') AND (
                ur.subforum_id = $3 OR ur.subforum_id IS NULL
            ))
        )
) as has_permission;
```

### Role Management

**File**: `internal/database/queries/appview/roles.sql`

```sql
-- name: CreateRole :one
INSERT INTO roles (name, description, is_platform_role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: UpdateRole :one
UPDATE roles 
SET description = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;
```

### Permission Management

**File**: `internal/database/queries/appview/permissions.sql`

```sql
-- name: CreatePermission :one
INSERT INTO permissions (name, description, resource_type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPermissionByName :one
SELECT * FROM permissions WHERE name = $1;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY name;

-- name: UpdatePermission :one
UPDATE permissions 
SET description = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;
```

### User Role Management

**File**: `internal/database/queries/appview/user_roles.sql`

```sql
-- name: AssignRole :one
INSERT INTO user_roles (user_did, role_id, subforum_id, granted_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: RevokeRole :exec
DELETE FROM user_roles WHERE user_did = $1 AND role_id = $2 AND subforum_id = $3;

-- name: GetUserRoles :many
SELECT ur.*, r.name as role_name, r.description as role_description
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_did = $1 AND ur.is_active = TRUE;

-- name: GetUserRolesBySubforum :many
SELECT ur.*, r.name as role_name, r.description as role_description
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_did = $1 AND ur.subforum_id = $2 AND ur.is_active = TRUE;
```

## API Endpoints

### Role Management

**Endpoint**: `POST /api/v1/admin/assign-role`

```go
func (h *RBACHandlers) AssignRole(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserDID     string `json:"user_did"`
        RoleName    string `json:"role_name"`
        SubforumID  string `json:"subforum_id,omitempty"`
    }

    // Validate request
    if req.UserDID == "" || req.RoleName == "" {
        http.Error(w, "user_did and role_name required", http.StatusBadRequest)
        return
    }

    // Get role
    role, err := h.queries.GetRoleByName(r.Context(), req.RoleName)
    if err != nil {
        http.Error(w, "Role not found", http.StatusNotFound)
        return
    }

    // Assign role
    var subforumID *string
    if req.SubforumID != "" {
        subforumID = &req.SubforumID
    }

    _, err = h.queries.AssignRole(r.Context(), generated.AssignRoleParams{
        UserDid:    req.UserDID,
        RoleID:     role.ID,
        SubforumID: subforumID,
        GrantedBy:  "system", // TODO: Get from authenticated user
    })

    if err != nil {
        http.Error(w, "Failed to assign role", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

**Endpoint**: `POST /api/v1/admin/revoke-role`

```go
func (h *RBACHandlers) RevokeRole(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserDID     string `json:"user_did"`
        RoleName    string `json:"role_name"`
        SubforumID  string `json:"subforum_id,omitempty"`
    }

    // Get role
    role, err := h.queries.GetRoleByName(r.Context(), req.RoleName)
    if err != nil {
        http.Error(w, "Role not found", http.StatusNotFound)
        return
    }

    // Revoke role
    var subforumID *string
    if req.SubforumID != "" {
        subforumID = &req.SubforumID
    }

    err = h.queries.RevokeRole(r.Context(), generated.RevokeRoleParams{
        UserDid:    req.UserDID,
        RoleID:     role.ID,
        SubforumID: subforumID,
    })

    if err != nil {
        http.Error(w, "Failed to revoke role", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

### Permission Checking

**Endpoint**: `GET /api/v1/me/permissions`

```go
func (h *RBACHandlers) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
    // Get user DID from authenticated session
    userDID := r.Header.Get("X-User-DID") // TODO: Extract from JWT token

    // Get user permissions
    permissions, err := h.queries.GetUserPermissions(r.Context(), userDID)
    if err != nil {
        http.Error(w, "Failed to get permissions", http.StatusInternalServerError)
        return
    }

    // Format response
    response := map[string]interface{}{
        "permissions": permissions,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## Default Roles and Permissions

### Platform Roles

**Role**: `platform_admin`
- **Description**: Platform administrator with full system access
- **Type**: Platform role (applies globally)
- **Permissions**: All platform permissions

**Role**: `platform_moderator`
- **Description**: Platform moderator with moderation access
- **Type**: Platform role (applies globally)
- **Permissions**: Moderation permissions

### Subforum Roles

**Role**: `subforum_admin`
- **Description**: Subforum administrator
- **Type**: Subforum-specific role
- **Permissions**: Subforum management permissions

**Role**: `subforum_moderator`
- **Description**: Subforum moderator
- **Type**: Subforum-specific role
- **Permissions**: Subforum moderation permissions

**Role**: `subforum_member`
- **Description**: Subforum member
- **Type**: Subforum-specific role
- **Permissions**: Basic subforum permissions

### Permission Types

**Platform Permissions**:
- `platform.manage_users` - Manage all users
- `platform.manage_roles` - Manage roles and permissions
- `platform.moderate_content` - Moderate all content
- `platform.view_analytics` - View platform analytics

**Subforum Permissions**:
- `subforum.manage` - Manage subforum settings
- `subforum.moderate` - Moderate subforum content
- `subforum.post` - Create posts in subforum
- `subforum.comment` - Comment on posts
- `subforum.vote` - Vote on content

**Post Permissions**:
- `post.create` - Create posts
- `post.edit` - Edit posts
- `post.delete` - Delete posts
- `post.moderate` - Moderate posts

**Comment Permissions**:
- `comment.create` - Create comments
- `comment.edit` - Edit comments
- `comment.delete` - Delete comments
- `comment.moderate` - Moderate comments

## Middleware Integration

### Permission Middleware

**File**: `internal/appview/middleware.go`  
**Function**: `RequirePermission(permission string)`

```go
func RequirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Get user DID from authenticated session
            userDID := r.Header.Get("X-User-DID") // TODO: Extract from JWT token

            // Check permission
            hasPermission, err := rbacService.CheckUserPermission(r.Context(), userDID, permission, nil)
            if err != nil {
                http.Error(w, "Failed to check permission", http.StatusInternalServerError)
                return
            }

            if !hasPermission {
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Usage in Handlers

```go
// Protect admin endpoints
adminMux.HandleFunc("/assign-role", RequirePermission("platform.manage_roles")(rbacHandlers.AssignRole))

// Protect subforum endpoints
subforumMux.HandleFunc("/create", RequirePermission("subforum.manage")(subforumHandlers.CreateSubforum))
```

## Error Handling

### Permission Check Errors

```go
// Handle permission check failures
hasPermission, err := rbacService.CheckUserPermission(ctx, userDID, permission, subforumID)
if err != nil {
    r.logger.Error("Failed to check user permission", 
        "error", err, 
        "user_did", userDID, 
        "permission", permission,
    )
    return false, fmt.Errorf("failed to check permission: %w", err)
}
```

### Role Assignment Errors

```go
// Handle role assignment failures
_, err = h.queries.AssignRole(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        return fmt.Errorf("user already has this role")
    }
    return fmt.Errorf("failed to assign role: %w", err)
}
```

## Performance Considerations

### Query Optimization
- **Indexes**: Proper indexes on user_did, role_id, subforum_id
- **Joins**: Efficient joins for permission checking
- **Caching**: Consider caching user permissions

### Permission Caching
- **User Permissions**: Cache user permissions for session duration
- **Role Permissions**: Cache role-permission mappings
- **Invalidation**: Clear cache on role/permission changes

## References

- [RBAC Implementation](internal/appview/rbac.go)
- [RBAC Handlers](internal/appview/rbac_handlers.go)
- [Database Queries](internal/database/queries/appview/)
- [Database Schema](internal/database/migrations/appview/002_rbac_schema.up.sql)
