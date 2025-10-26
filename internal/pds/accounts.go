package pds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

	// Ensure handle has proper domain format
	var handleWithDomain string
	if strings.Contains(req.Handle, ".") {
		// Handle already has domain
		handleWithDomain = req.Handle
	} else {
		// Use the handle base from configuration
		handleWithDomain = req.Handle + "." + s.config.PDS.Atproto.HandleBase
	}

	// Check if handle is already taken
	existingUser, err := s.db.GetUserByHandle(context.Background(), handleWithDomain)
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
		Handle:       handleWithDomain,
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
	// Handle already has domain from above
	if err := s.authService.AddUserToMockDirectory(r.Context(), user.Did, user.Handle); err != nil {
		s.logger.Warn("Failed to add user to mock directory", "error", err)
		// Don't fail registration, just log the warning
	}

	// Publish identity resolved event
	if err := s.eventStream.PublishIdentityEvent(r.Context(), EventTypeIdentityResolved, user.Handle, user.Did); err != nil {
		s.logger.Error("Failed to publish identity resolved event", "error", err)
		// Don't fail the request, just log the error
	}

	// Create app.bsky.actor.profile record for the new user
	// This ensures the profile fetcher can find display names
	profileRecord := map[string]interface{}{
		"$type":       "app.bsky.actor.profile",
		"displayName": user.Handle, // Use handle as initial display name
		"description": "New HashPost user",
		"createdAt":   time.Now().Format(time.RFC3339),
	}

	// Create the profile record in PDS
	profileURI, profileCID, err := s.createRecordInPDS(r.Context(), user.Did, "app.bsky.actor.profile", "self", profileRecord)
	if err != nil {
		s.logger.Warn("Failed to create profile record", "error", err, "did", user.Did)
		// Don't fail registration, just log the warning
	} else {
		s.logger.Info("Profile record created", "did", user.Did, "uri", profileURI, "cid", profileCID)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.authService.GenerateTokens(session)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	// For now, use handle as displayName (TODO: fetch from profile records)
	displayName := user.Handle
	// TODO: Fetch displayName from app.bsky.actor.profile record
	// For now, just use handle

	response := map[string]interface{}{
		"accessJwt":   accessToken,
		"refreshJwt":  refreshToken,
		"handle":      user.Handle,
		"did":         user.Did,
		"email":       user.Email,
		"displayName": displayName,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// createRecordInPDS creates a record in the PDS database
func (s *Server) createRecordInPDS(ctx context.Context, repo, collection, rkey string, record map[string]interface{}) (uri, cid string, err error) {
	// Generate record ID and URI
	uri = fmt.Sprintf("at://%s/%s/%s", repo, collection, rkey)

	// Compute proper content-addressed CID for the record
	computedCID, err := s.cidService.ComputeRecordCID(ctx, record)
	if err != nil {
		s.logger.Error("Failed to compute CID for record", "error", err)
		return "", "", fmt.Errorf("failed to compute record CID: %w", err)
	}
	cid = computedCID

	// Store record in database based on collection type
	switch collection {
	case "app.bsky.actor.profile":
		err = s.createGenericRecord(ctx, repo, collection, rkey, uri, cid, record)
	default:
		// For other collections, store as generic record
		err = s.createGenericRecord(ctx, repo, collection, rkey, uri, cid, record)
	}

	if err != nil {
		s.logger.Error("Failed to create record", "error", err, "collection", collection)
		return "", "", fmt.Errorf("failed to create record: %w", err)
	}

	s.logger.Debug("Record created in PDS", "uri", uri, "cid", cid)
	return uri, cid, nil
}
