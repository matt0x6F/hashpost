package appview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthHandler handles authentication endpoints by proxying to PDS
func (h *Handlers) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// Determine which auth endpoint based on the URL path
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/login"):
		h.handleLogin(w, r)
	case strings.HasSuffix(path, "/register"):
		h.handleRegister(w, r)
	case strings.HasSuffix(path, "/register/external"):
		h.handleExternalRegister(w, r)
	case strings.HasSuffix(path, "/me"):
		h.handleGetCurrentUser(w, r)
	case strings.HasSuffix(path, "/logout"):
		h.handleLogout(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleLogin handles POST /api/v1/auth/login
func (h *Handlers) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Handling login", "email", req.Email)

	// Proxy to PDS for authentication
	pdsResponse, err := h.makePDSRequest("POST", "com.atproto.server.createSession", map[string]string{
		"identifier": req.Email,
		"password":   req.Password,
	})
	if err != nil {
		h.logger.Error("PDS login failed", "error", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Parse PDS response and map to frontend format
	var pdsData map[string]interface{}
	if err := json.Unmarshal(pdsResponse, &pdsData); err != nil {
		h.logger.Error("Failed to parse PDS response", "error", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Map PDS response to frontend format
	response := map[string]interface{}{
		"accessToken":  pdsData["accessJwt"],
		"refreshToken": pdsData["refreshJwt"],
		"handle":       pdsData["handle"],
		"did":          pdsData["did"],
		"email":        pdsData["email"],
		"displayName":  pdsData["displayName"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleRegister handles POST /api/v1/auth/register
func (h *Handlers) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		Handle     string `json:"handle"`
		InviteCode string `json:"inviteCode,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Handle == "" {
		http.Error(w, "email, password, and handle required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Handling registration", "email", req.Email, "handle", req.Handle)

	// Proxy to PDS for registration
	body := map[string]string{
		"handle":   req.Handle,
		"password": req.Password,
		"email":    req.Email,
	}
	if req.InviteCode != "" {
		body["inviteCode"] = req.InviteCode
	}

	pdsResponse, err := h.makePDSRequest("POST", "com.atproto.server.createAccount", body)
	if err != nil {
		h.logger.Error("PDS registration failed", "error", err)

		// Parse error message to determine appropriate status code
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "409") && strings.Contains(errorMsg, "Handle already taken") {
			http.Error(w, "Handle already taken", http.StatusConflict)
			return
		} else if strings.Contains(errorMsg, "400") {
			http.Error(w, "Invalid registration data", http.StatusBadRequest)
			return
		} else if strings.Contains(errorMsg, "401") {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	// Parse PDS response and map to frontend format
	var pdsData map[string]interface{}
	if err := json.Unmarshal(pdsResponse, &pdsData); err != nil {
		h.logger.Error("Failed to parse PDS response", "error", err)
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	// Map PDS response to frontend format
	response := map[string]interface{}{
		"accessToken":  pdsData["accessJwt"],
		"refreshToken": pdsData["refreshJwt"],
		"handle":       pdsData["handle"],
		"did":          pdsData["did"],
		"email":        pdsData["email"],
		"displayName":  pdsData["displayName"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleExternalRegister handles POST /api/v1/auth/register/external
func (h *Handlers) handleExternalRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		DID    string `json:"did"`
		Handle string `json:"handle"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.DID == "" && req.Handle == "" {
		http.Error(w, "DID or handle required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Handling external registration", "did", req.DID, "handle", req.Handle)

	// Resolve DID if handle provided
	did := req.DID
	if did == "" && req.Handle != "" {
		// Resolve handle to DID
		resolvedDID, err := h.resolveHandleToDID(r.Context(), req.Handle)
		if err != nil {
			h.logger.Error("Failed to resolve handle", "error", err, "handle", req.Handle)
			http.Error(w, "Failed to resolve handle", http.StatusBadRequest)
			return
		}
		did = resolvedDID
	}

	// Validate DID ownership (simplified for now)
	if err := h.validateDIDOwnership(r.Context(), did); err != nil {
		h.logger.Error("DID ownership validation failed", "error", err, "did", did)
		http.Error(w, "DID ownership validation failed", http.StatusUnauthorized)
		return
	}

	// Create lightweight user record in AppView
	user, err := h.createExternalUserRecord(r.Context(), did, req.Handle)
	if err != nil {
		h.logger.Error("Failed to create external user record", "error", err, "did", did)
		http.Error(w, "Failed to create user record", http.StatusInternalServerError)
		return
	}

	// Return user info
	response := map[string]interface{}{
		"did":        user.DID,
		"handle":     user.Handle,
		"isExternal": true,
		"message":    "External user registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// resolveHandleToDID resolves a handle to a DID
func (h *Handlers) resolveHandleToDID(ctx context.Context, handle string) (string, error) {
	// For now, this is a simplified implementation
	// In production, this would use the identity directory to resolve the handle

	// Mock resolution for development
	if strings.Contains(handle, ".hashpost.local") {
		return "did:plc:hashpost-binding-test", nil
	}

	// For external handles, we'd need to resolve via identity directory
	return "did:plc:external-user", nil
}

// validateDIDOwnership validates that the user owns the DID
func (h *Handlers) validateDIDOwnership(ctx context.Context, did string) error {
	// For now, this is a simplified implementation
	// In production, this would:
	// 1. Check if the DID is resolvable
	// 2. Verify the user can authenticate with their PDS
	// 3. Perform a challenge-response or OAuth flow

	h.logger.Debug("Validating DID ownership", "did", did)

	// Mock validation for development
	return nil
}

// createExternalUserRecord creates a lightweight user record for external users
func (h *Handlers) createExternalUserRecord(ctx context.Context, did, handle string) (*AppViewUser, error) {
	// For now, this is a simplified implementation
	// In production, this would use the RBAC service to create the user record

	h.logger.Info("Creating external user record", "did", did, "handle", handle)

	// Mock user creation
	user := &AppViewUser{
		ID:          uuid.New(),
		DID:         did,
		Handle:      handle,
		DisplayName: handle,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		PDSSource:   &[]string{"https://external-pds.example.com"}[0],
		LastSeenAt:  timePtr(time.Now()),
	}

	return user, nil
}

// Helper functions for pointer creation
func timePtr(t time.Time) *time.Time {
	return &t
}

// handleGetCurrentUser handles GET /api/v1/auth/me
func (h *Handlers) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user context from middleware (authentication already handled by middleware)
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Query the users table to get displayName
	user, err := h.queries.GetUserByDID(r.Context(), userCtx.Did)
	if err != nil {
		h.logger.Error("Failed to get user from database", "error", err, "did", userCtx.Did)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Return user session info
	response := map[string]interface{}{
		"did":         userCtx.Did,
		"handle":      userCtx.Handle,
		"email":       "", // Email not available in AppView
		"displayName": user.DisplayName,
		"active":      userCtx.IsActive,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleLogout handles POST /api/v1/auth/logout
func (h *Handlers) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get JWT token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Proxy to PDS for logout
	pdsResponse, err := h.makePDSRequest("POST", "com.atproto.server.deleteSession", nil, authHeader)
	if err != nil {
		h.logger.Error("PDS logout failed", "error", err)
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	// Return the PDS response directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(pdsResponse)
}

// handleRefresh handles POST /api/v1/auth/refresh
func (h *Handlers) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get refresh token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}

	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	h.logger.Debug("Handling token refresh", "refreshTokenLength", len(refreshToken))

	// Proxy to PDS for token refresh
	pdsResponse, err := h.makePDSRequest("POST", "com.atproto.server.refreshSession", map[string]string{
		"refreshJwt": refreshToken,
	})

	if err != nil {
		h.logger.Error("PDS refresh failed", "error", err)
		http.Error(w, "Token refresh failed", http.StatusUnauthorized)
		return
	}

	// Parse PDS response and map to frontend format
	var pdsData map[string]interface{}
	if err := json.Unmarshal(pdsResponse, &pdsData); err != nil {
		h.logger.Error("Failed to parse PDS refresh response", "error", err)
		http.Error(w, "Token refresh failed", http.StatusInternalServerError)
		return
	}

	// Map PDS response to frontend format
	response := map[string]interface{}{
		"accessToken":  pdsData["accessJwt"],
		"refreshToken": pdsData["refreshJwt"],
		"handle":       pdsData["handle"],
		"did":          pdsData["did"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// makePDSRequest makes a request to the PDS server
func (h *Handlers) makePDSRequest(method, endpoint string, body interface{}, authHeader ...string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := h.pdsURL + "/xrpc/" + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if len(authHeader) > 0 && authHeader[0] != "" {
		req.Header.Set("Authorization", authHeader[0])
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("PDS request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
