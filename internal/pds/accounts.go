package pds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
)

// handleCreateAccount handles account creation
func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Handle     string `json:"handle"`
		Password   string `json:"password"`
		Email      string `json:"email,omitempty"`
		InviteCode string `json:"inviteCode,omitempty"`
		DID        string `json:"did,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Handle == "" || req.Password == "" {
		http.Error(w, "handle and password required", http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := s.authService.ValidatePasswordStrength(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.logger.Debug("Creating account", "handle", req.Handle)

	// Check if handle is already taken
	existingUser, err := s.db.GetUserByHandle(context.Background(), req.Handle)
	if err == nil && existingUser != nil {
		http.Error(w, "Handle already taken", http.StatusConflict)
		return
	}

	// Validate invite code if required
	if req.InviteCode != "" {
		if err := s.validateInviteCode(r.Context(), req.InviteCode); err != nil {
			s.logger.Error("Invalid invite code", "error", err, "invite_code", req.InviteCode)
			http.Error(w, "Invalid invite code", http.StatusBadRequest)
			return
		}
	}

	// Generate proper DID (use provided DID if available, otherwise generate one)
	var did string
	if req.DID != "" {
		did = req.DID
	} else {
		did = fmt.Sprintf("did:plc:hashpost-%s", uuid.New().String())
	}

	// Hash the password
	passwordHash, err := s.authService.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create user record in database with password hash
	var email *string
	if req.Email != "" {
		email = &req.Email
	}
	user, err := s.db.CreateUserWithPassword(context.Background(), &generated.CreateUserWithPasswordParams{
		Handle:       req.Handle,
		Did:          did,
		Email:        email,
		PasswordHash: &passwordHash,
	})
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	// Create session for the new user
	session := &Session{
		ID:        uuid.New().String(),
		DID:       user.Did,
		Handle:    user.Handle,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Store session in database
	_, err = s.db.CreateUserSession(context.Background(), &generated.CreateUserSessionParams{
		SessionID: session.ID,
		UserDid:   session.DID,
		Handle:    session.Handle,
		ExpiresAt: session.ExpiresAt,
	})
	if err != nil {
		s.logger.Error("Failed to store session", "error", err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	// Add user to mock identity directory (development only)
	// Use proper handle format for identity directory
	handleWithDomain := user.Handle + ".hashpost.local"
	if err := s.authService.AddUserToMockDirectory(r.Context(), user.Did, handleWithDomain); err != nil {
		s.logger.Warn("Failed to add user to mock directory", "error", err)
		// Don't fail registration, just log the warning
	}

	// Generate tokens
	accessToken, refreshToken, err := s.authService.GenerateTokens(session)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"accessJwt":  accessToken,
		"refreshJwt": refreshToken,
		"handle":     user.Handle,
		"did":        user.Did,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
