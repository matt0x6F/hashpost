package appview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	// Return the PDS response directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(pdsResponse)
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
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	// Return the PDS response directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(pdsResponse)
}

// handleGetCurrentUser handles GET /api/v1/auth/me
func (h *Handlers) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get JWT token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Proxy to PDS for session info
	pdsResponse, err := h.makePDSRequest("GET", "com.atproto.server.getSession", nil, authHeader)
	if err != nil {
		h.logger.Error("PDS session lookup failed", "error", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Return the PDS response directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(pdsResponse)
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
