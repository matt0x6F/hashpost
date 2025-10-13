package appview

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

// RBACHandlers contains handlers for RBAC-related endpoints
type RBACHandlers struct {
	rbacService *RBACService
	logger      *slog.Logger
}

// NewRBACHandlers creates a new RBAC handlers instance
func NewRBACHandlers(rbacService *RBACService, logger *slog.Logger) *RBACHandlers {
	return &RBACHandlers{
		rbacService: rbacService,
		logger:      logger,
	}
}

// TestAuth tests authentication and returns user context
func (h *RBACHandlers) TestAuth(w http.ResponseWriter, r *http.Request) {
	// Get user context from request
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

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

// AssignRole assigns a role to a user (platform admin only)
func (h *RBACHandlers) AssignRole(w http.ResponseWriter, r *http.Request) {
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

	// Get current user (who is assigning the role)
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Assign the role
	err := h.rbacService.AssignRole(r.Context(), req.UserDID, req.RoleName, req.SubforumID, userCtx.Did, req.ExpiresAt)
	if err != nil {
		h.logger.Error("Failed to assign role", "error", err)
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Role assigned successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RevokeRole revokes a role from a user (platform admin only)
func (h *RBACHandlers) RevokeRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		UserDID    string  `json:"user_did"`
		RoleName   string  `json:"role_name"`
		SubforumID *string `json:"subforum_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get current user (who is revoking the role)
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Revoke the role
	err := h.rbacService.RevokeRole(r.Context(), req.UserDID, req.RoleName, req.SubforumID)
	if err != nil {
		h.logger.Error("Failed to revoke role", "error", err)
		http.Error(w, "Failed to revoke role", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Role revoked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckPermission checks if a user has a specific permission
func (h *RBACHandlers) CheckPermission(w http.ResponseWriter, r *http.Request) {
	// Get user context from request
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get permission from query parameter
	permission := r.URL.Query().Get("permission")
	if permission == "" {
		http.Error(w, "Permission parameter required", http.StatusBadRequest)
		return
	}

	// Get subforum ID from query parameter (optional)
	subforumID := r.URL.Query().Get("subforum_id")
	var subforumIDPtr *string
	if subforumID != "" {
		subforumIDPtr = &subforumID
	}

	// Check permission
	hasPermission, err := h.rbacService.CheckPermission(r.Context(), userCtx.Did, permission, subforumIDPtr)
	if err != nil {
		h.logger.Error("Permission check failed", "error", err)
		http.Error(w, "Permission check failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_did":       userCtx.Did,
		"permission":     permission,
		"subforum_id":    subforumID,
		"has_permission": hasPermission,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListUsers lists all users with their roles (platform admin only)
func (h *RBACHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	limit := 20
	offset := 0
	subforumID := r.URL.Query().Get("subforum_id")

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get users with roles
	users, err := h.rbacService.GetUsersWithRoles(r.Context(), limit, offset, subforumID)
	if err != nil {
		h.logger.Error("Failed to get users", "error", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListRoles lists all available roles (platform admin only)
func (h *RBACHandlers) ListRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roles, err := h.rbacService.GetAllRoles(r.Context())
	if err != nil {
		h.logger.Error("Failed to get roles", "error", err)
		http.Error(w, "Failed to get roles", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"roles": roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListPermissions lists all available permissions (platform admin only)
func (h *RBACHandlers) ListPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	permissions, err := h.rbacService.GetAllPermissions(r.Context())
	if err != nil {
		h.logger.Error("Failed to get permissions", "error", err)
		http.Error(w, "Failed to get permissions", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"permissions": permissions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserRoles gets all roles for a specific user
func (h *RBACHandlers) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user DID from query parameter
	userDID := r.URL.Query().Get("user_did")
	if userDID == "" {
		http.Error(w, "user_did parameter required", http.StatusBadRequest)
		return
	}

	// Get subforum ID from query parameter (optional)
	subforumID := r.URL.Query().Get("subforum_id")
	var subforumIDPtr *string
	if subforumID != "" {
		subforumIDPtr = &subforumID
	}

	roles, err := h.rbacService.GetUserRoles(r.Context(), userDID, subforumIDPtr)
	if err != nil {
		h.logger.Error("Failed to get user roles", "error", err)
		http.Error(w, "Failed to get user roles", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_did": userDID,
		"roles":    roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
