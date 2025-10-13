package appview

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// TestRBACHandlers contains test handlers for RBAC functionality
type TestRBACHandlers struct {
	rbacService *RBACService
	logger      *slog.Logger
}

// NewTestRBACHandlers creates a new test RBAC handlers instance
func NewTestRBACHandlers(rbacService *RBACService, logger *slog.Logger) *TestRBACHandlers {
	return &TestRBACHandlers{
		rbacService: rbacService,
		logger:      logger,
	}
}

// TestAuthWithoutJWT tests authentication without JWT validation
func (h *TestRBACHandlers) TestAuthWithoutJWT(w http.ResponseWriter, r *http.Request) {
	// Create a mock user context for testing
	userCtx := &UserContext{
		Did:      "did:plc:hashpost-binding-test",
		Handle:   "testuser.hashpost.local",
		Roles:    []UserRole{},
		IsActive: true,
	}

	// Get user roles from database
	roles, err := h.rbacService.getUserRoles(r.Context(), userCtx.Did)
	if err != nil {
		h.logger.Error("Failed to get user roles", "error", err)
		http.Error(w, "Failed to get user roles", http.StatusInternalServerError)
		return
	}

	userCtx.Roles = roles

	// Return user context
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"did":    userCtx.Did,
			"handle": userCtx.Handle,
			"roles":  userCtx.Roles,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestPermissionCheck tests permission checking without authentication
func (h *TestRBACHandlers) TestPermissionCheck(w http.ResponseWriter, r *http.Request) {
	// Get parameters from query
	userDID := r.URL.Query().Get("user_did")
	permission := r.URL.Query().Get("permission")
	subforumID := r.URL.Query().Get("subforum_id")

	if userDID == "" {
		userDID = "did:plc:hashpost-binding-test" // Default test user
	}

	if permission == "" {
		permission = "post.create" // Default permission
	}

	var subforumIDPtr *string
	if subforumID != "" {
		subforumIDPtr = &subforumID
	}

	// Check permission
	hasPermission, err := h.rbacService.CheckPermission(r.Context(), userDID, permission, subforumIDPtr)
	if err != nil {
		h.logger.Error("Permission check failed", "error", err)
		http.Error(w, "Permission check failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_did":       userDID,
		"permission":     permission,
		"subforum_id":    subforumID,
		"has_permission": hasPermission,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestRoleAssignment tests role assignment without authentication
func (h *TestRBACHandlers) TestRoleAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		UserDID    string  `json:"user_did"`
		RoleName   string  `json:"role_name"`
		SubforumID *string `json:"subforum_id,omitempty"`
		ExpiresAt  *string `json:"expires_at,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use default values if not provided
	if req.UserDID == "" {
		req.UserDID = "did:plc:hashpost-binding-test"
	}
	if req.RoleName == "" {
		req.RoleName = "user"
	}

	// Assign the role
	err := h.rbacService.AssignRole(r.Context(), req.UserDID, req.RoleName, req.SubforumID, "did:plc:hashpost-admin-test", req.ExpiresAt)
	if err != nil {
		h.logger.Error("Failed to assign role", "error", err)
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "Role assigned successfully",
		"user_did": req.UserDID,
		"role":     req.RoleName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestRoleCheck tests role checking without authentication
func (h *TestRBACHandlers) TestRoleCheck(w http.ResponseWriter, r *http.Request) {
	// Get parameters from query
	userDID := r.URL.Query().Get("user_did")
	roleName := r.URL.Query().Get("role_name")
	subforumID := r.URL.Query().Get("subforum_id")

	if userDID == "" {
		userDID = "did:plc:hashpost-binding-test" // Default test user
	}

	if roleName == "" {
		roleName = "user" // Default role
	}

	var subforumIDPtr *string
	if subforumID != "" {
		subforumIDPtr = &subforumID
	}

	// Check role
	hasRole, err := h.rbacService.HasRole(r.Context(), userDID, roleName, subforumIDPtr)
	if err != nil {
		h.logger.Error("Role check failed", "error", err)
		http.Error(w, "Role check failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_did":    userDID,
		"role_name":   roleName,
		"subforum_id": subforumID,
		"has_role":    hasRole,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
