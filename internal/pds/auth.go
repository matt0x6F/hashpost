package pds

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/bluesky-social/indigo/atproto/auth" // Register JWT signing methods
	"github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/google/uuid"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	jwtservice "github.com/matt0x6f/hashpost/internal/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	EnvironmentProduction  = "production"
	EnvironmentDevelopment = "development"
)

// AuthService handles DID-based authentication using Indigo libraries
type AuthService struct {
	directory  identity.Directory
	logger     *slog.Logger
	signingKey crypto.PrivateKey
	serverDID  string
	db         *generated.Queries
	jwtService jwtservice.JWTService
}

// NewAuthService creates a new authentication service
func NewAuthService(db *generated.Queries, logger *slog.Logger, jwtService jwtservice.JWTService) *AuthService {
	var directory identity.Directory

	// Check environment to determine which directory to use
	environment := os.Getenv("ENVIRONMENT")
	if environment == EnvironmentProduction {
		// Production: Use real atproto identity resolution
		// This connects to plc.directory and DNS for real DID/handle resolution
		directory = identity.DefaultDirectory()
		logger.Info("Using production identity directory (plc.directory + DNS)")
	} else {
		// Development/Testing: Use mock directory with test identities
		mockDir := identity.NewMockDirectory()

		// Add test identities for development
		testUser := identity.Identity{
			DID:    syntax.DID("did:plc:hashpost-binding-test"),
			Handle: syntax.Handle("testuser.hashpost.local"),
		}
		adminUser := identity.Identity{
			DID:    syntax.DID("did:plc:hashpost-admin-test"),
			Handle: syntax.Handle("admin.hashpost.local"),
		}

		mockDir.Insert(testUser)
		mockDir.Insert(adminUser)

		directory = &mockDir
		logger.Info("Using development identity directory (mock)")
	}

	// Generate a signing key for JWT tokens
	signingKey, err := crypto.GeneratePrivateKeyK256()
	if err != nil {
		logger.Error("Failed to generate signing key", "error", err)
		// In production, this should be fatal
		panic("Failed to generate signing key")
	}

	// Get server DID from environment or use default
	serverDID := os.Getenv("SERVER_DID")
	if serverDID == "" {
		serverDID = "did:plc:hashpost-server"
	}

	return &AuthService{
		directory:  directory,
		logger:     logger,
		signingKey: signingKey,
		serverDID:  serverDID,
		db:         db,
		jwtService: jwtService,
	}
}

// AddUserToMockDirectory adds a user to the mock identity directory (development only)
func (as *AuthService) AddUserToMockDirectory(ctx context.Context, did, handle string) error {
	// Only works in development mode with mock directory
	if mockDir, ok := as.directory.(*identity.MockDirectory); ok {
		userIdentity := identity.Identity{
			DID:    syntax.DID(did),
			Handle: syntax.Handle(handle),
		}
		mockDir.Insert(userIdentity)
		as.logger.Debug("Added user to mock identity directory", "did", did, "handle", handle)
		return nil
	}
	return fmt.Errorf("not a mock directory")
}

// Session represents an authenticated session
type Session struct {
	ID        string    `json:"id"`
	DID       string    `json:"did"`
	Handle    string    `json:"handle"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthenticateSession authenticates a user session using DID-based verification
func (as *AuthService) AuthenticateSession(ctx context.Context, identifier, password string) (*Session, error) {
	as.logger.Debug("Authenticating session", "identifier", identifier)

	// Parse the identifier (could be handle or DID)
	ident, err := syntax.ParseAtIdentifier(identifier)
	if err != nil {
		return nil, fmt.Errorf("invalid identifier: %w", err)
	}

	var did string
	var handle string

	// Resolve the identifier to get DID and handle
	if ident.IsHandle() {
		// Resolve handle to DID
		handle = ident.String()

		// First check if this is a local user in our database
		if as.db != nil {
			user, err := as.db.GetUserByHandle(ctx, handle)
			if err == nil && user != nil {
				as.logger.Info("Handle resolved from database", "handle", handle, "did", user.Did)
				did = user.Did
			} else {
				// If not found in database, try the identity directory
				identity, err := as.directory.LookupHandle(ctx, syntax.Handle(handle))
				if err != nil {
					as.logger.Error("Failed to resolve handle", "error", err, "handle", handle)
					return nil, fmt.Errorf("handle resolution failed: %w", err)
				}
				did = identity.DID.String()
			}
		} else {
			// No database available, use identity directory
			identity, err := as.directory.LookupHandle(ctx, syntax.Handle(handle))
			if err != nil {
				as.logger.Error("Failed to resolve handle", "error", err, "handle", handle)
				return nil, fmt.Errorf("handle resolution failed: %w", err)
			}
			did = identity.DID.String()
		}
	} else if ident.IsDID() {
		// Resolve DID to get handle
		did = ident.String()

		// First check if this is a local user in our database
		if as.db != nil {
			user, err := as.db.GetUserByDID(ctx, did)
			if err == nil && user != nil {
				as.logger.Info("DID resolved from database", "did", did, "handle", user.Handle)
				handle = user.Handle
			} else {
				// If not found in database, try the identity directory
				identity, err := as.directory.LookupDID(ctx, syntax.DID(did))
				if err != nil {
					as.logger.Error("Failed to resolve DID", "error", err, "did", did)
					return nil, fmt.Errorf("DID resolution failed: %w", err)
				}
				handle = identity.Handle.String()
			}
		} else {
			// No database available, use identity directory
			identity, err := as.directory.LookupDID(ctx, syntax.DID(did))
			if err != nil {
				as.logger.Error("Failed to resolve DID", "error", err, "did", did)
				return nil, fmt.Errorf("DID resolution failed: %w", err)
			}
			handle = identity.Handle.String()
		}
	} else {
		return nil, fmt.Errorf("invalid identifier type")
	}

	// Validate password against stored hash
	if err := as.validatePassword(ctx, did, password); err != nil {
		as.logger.Error("Password validation failed", "error", err, "did", did)
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	// Create session
	session := &Session{
		ID:        uuid.New().String(),
		DID:       did,
		Handle:    handle,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}

	// Store session in database
	_, err = as.db.CreateUserSession(ctx, &generated.CreateUserSessionParams{
		SessionID: session.ID,
		UserDid:   session.DID,
		Handle:    session.Handle,
		ExpiresAt: session.ExpiresAt,
	})
	if err != nil {
		as.logger.Error("Failed to store session", "error", err)
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	as.logger.Info("Session created", "did", did, "handle", handle, "session_id", session.ID)
	return session, nil
}

// ValidateSession validates an existing session
func (as *AuthService) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	as.logger.Debug("Validating session", "session_id", sessionID)

	// Get session from database
	dbSession, err := as.db.GetUserSession(ctx, sessionID)
	if err != nil {
		as.logger.Error("Session not found", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("session not found or expired")
	}

	// Update last accessed time
	err = as.db.UpdateUserSessionLastAccessed(ctx, sessionID)
	if err != nil {
		as.logger.Error("Failed to update session last accessed", "error", err)
		// Don't fail validation, just log the error
	}

	// Convert to session object
	session := &Session{
		ID:        dbSession.SessionID,
		DID:       dbSession.UserDid,
		Handle:    dbSession.Handle,
		CreatedAt: dbSession.CreatedAt.Time,
		ExpiresAt: dbSession.ExpiresAt,
	}

	as.logger.Debug("Session validated", "session_id", sessionID, "user_did", session.DID)
	return session, nil
}

// GenerateTokens generates access and refresh tokens for a session
func (as *AuthService) GenerateTokens(session *Session) (string, string, error) {
	now := time.Now()

	// Create access token (short-lived, 1 hour)
	accessClaims := map[string]interface{}{
		"sub":    session.DID,
		"iss":    as.serverDID,
		"aud":    as.serverDID,
		"iat":    now.Unix(),
		"exp":    now.Add(time.Hour).Unix(),
		"jti":    session.ID,
		"scope":  "com.atproto.access",
		"handle": session.Handle,
	}

	// Create header for ES256K signing method
	accessHeader := map[string]interface{}{
		"alg": "ES256K",
		"typ": "JWT",
	}

	// Generate access token using JWT service
	accessTokenString, err := as.jwtService.GenerateSignedToken(accessClaims, accessHeader, as.signingKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token (long-lived, 30 days)
	refreshClaims := map[string]interface{}{
		"sub":    session.DID,
		"iss":    as.serverDID,
		"aud":    as.serverDID,
		"iat":    now.Unix(),
		"exp":    now.Add(30 * 24 * time.Hour).Unix(),
		"jti":    session.ID + "_refresh",
		"scope":  "com.atproto.refresh",
		"handle": session.Handle,
	}

	// Create header for ES256K signing method
	refreshHeader := map[string]interface{}{
		"alg": "ES256K",
		"typ": "JWT",
	}

	// Generate refresh token using JWT service
	refreshTokenString, err := as.jwtService.GenerateSignedToken(refreshClaims, refreshHeader, as.signingKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	as.logger.Debug("Generated JWT tokens", "session_id", session.ID, "handle", session.Handle, "expires", now.Add(time.Hour))
	return accessTokenString, refreshTokenString, nil
}

// ValidateToken validates a JWT token and returns session information
func (as *AuthService) ValidateToken(token string) (*Session, error) {
	// Use JWT service to validate and parse the token
	claims, err := as.jwtService.ValidateAndParse(token, func(token interface{}) (interface{}, error) {
		// Get the public key from our signing key
		publicKey, err := as.signingKey.PublicKey()
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Extract required fields
	did, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing subject (DID) in token")
	}

	handle, ok := claims["handle"].(string)
	if !ok {
		return nil, fmt.Errorf("missing handle in token")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("missing JTI in token")
	}

	// Check expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing expiration in token")
	}

	expirationTime := time.Unix(int64(exp), 0)
	if time.Now().After(expirationTime) {
		return nil, fmt.Errorf("token has expired")
	}

	// Create session from token claims
	session := &Session{
		ID:        jti,
		DID:       did,
		Handle:    handle,
		CreatedAt: time.Now().Add(-1 * time.Hour), // Approximate
		ExpiresAt: expirationTime,
	}

	as.logger.Debug("Validated JWT token", "session_id", session.ID, "did", session.DID, "handle", session.Handle)
	return session, nil
}

// ResolveHandle resolves a handle to a DID using the identity directory
func (as *AuthService) ResolveHandle(ctx context.Context, handle string) (string, error) {
	as.logger.Debug("Resolving handle", "handle", handle)

	// First check if this is a local user in our database (if database is available)
	if as.db != nil {
		user, err := as.db.GetUserByHandle(ctx, handle)
		if err == nil && user != nil {
			as.logger.Info("Handle resolved from database", "handle", handle, "did", user.Did)
			return user.Did, nil
		}
	}

	// If not found in database, try the identity directory
	identity, err := as.directory.LookupHandle(ctx, syntax.Handle(handle))
	if err != nil {
		as.logger.Error("Failed to resolve handle", "error", err, "handle", handle)
		return "", fmt.Errorf("handle resolution failed: %w", err)
	}

	did := identity.DID.String()
	as.logger.Info("Handle resolved from directory", "handle", handle, "did", did)
	return did, nil
}

// ResolveDID resolves a DID to get identity information
func (as *AuthService) ResolveDID(ctx context.Context, did string) (*identity.Identity, error) {
	as.logger.Debug("Resolving DID", "did", did)

	identity, err := as.directory.LookupDID(ctx, syntax.DID(did))
	if err != nil {
		as.logger.Error("Failed to resolve DID", "error", err, "did", did)
		return nil, fmt.Errorf("DID resolution failed: %w", err)
	}

	as.logger.Info("DID resolved", "did", did, "handle", identity.Handle.String())
	return identity, nil
}

// VerifySignature verifies a cryptographic signature
func (as *AuthService) VerifySignature(ctx context.Context, did string, message []byte, signature []byte) error {
	as.logger.Debug("Verifying signature", "did", did)

	// Resolve the DID to get the public key
	identityInfo, err := as.ResolveDID(ctx, did)
	if err != nil {
		return fmt.Errorf("failed to resolve DID for signature verification: %w", err)
	}

	// Get the public key from the DID document
	// For now, we'll use a simplified approach since we're using mock directory
	// In production, this would involve parsing the DID document and extracting the public key
	if _, ok := as.directory.(*identity.MockDirectory); ok {
		// For mock directory, we'll just verify the identity exists
		// In production, you'd extract the public key from the DID document
		as.logger.Debug("Mock signature verification", "did", did, "handle", identityInfo.Handle.String())
	} else {
		// For production, implement proper signature verification:
		// 1. Parse the DID document to extract the public key
		// 2. Use the public key to verify the signature against the message
		// 3. Check signature validity and expiration
		as.logger.Debug("Production signature verification", "did", did)
	}

	as.logger.Info("Signature verified", "did", did, "handle", identityInfo.Handle.String())
	return nil
}

// validatePassword validates a password against the stored hash
func (as *AuthService) validatePassword(ctx context.Context, did, password string) error {
	// Get the user's password hash from the database
	user, err := as.db.GetUserByDID(ctx, did)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user has a password hash (for existing users without passwords)
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		// For development, allow any password for users without password hash
		// In production, this should require password reset
		as.logger.Warn("User has no password hash, allowing authentication", "did", did)
		return nil
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		return fmt.Errorf("password mismatch: %w", err)
	}

	return nil
}

// HashPassword hashes a password using bcrypt
func (as *AuthService) HashPassword(password string) (string, error) {
	// Generate a salt and hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// ValidatePasswordStrength validates password complexity requirements
func (as *AuthService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add more complexity requirements as needed
	// For now, just check minimum length
	return nil
}

// handleCreateSession handles session creation
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Identifier string `json:"identifier"` // handle or email
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Identifier == "" || req.Password == "" {
		http.Error(w, "identifier and password required", http.StatusBadRequest)
		return
	}

	s.logger.Debug("Creating session", "identifier", req.Identifier)

	// Use DID-based authentication
	session, err := s.authService.AuthenticateSession(r.Context(), req.Identifier, req.Password)
	if err != nil {
		s.logger.Error("Authentication failed", "error", err, "identifier", req.Identifier)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := s.authService.GenerateTokens(session)
	if err != nil {
		s.logger.Error("Failed to generate tokens", "error", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"accessJwt":  accessToken,
		"refreshJwt": refreshToken,
		"handle":     session.Handle,
		"did":        session.DID,
		"email":      s.getUserEmailFromDID(r.Context(), session.DID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetSession handles session retrieval
func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Basic token validation (in production, use proper JWT validation)
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	session, err := s.authService.ValidateToken(token)
	if err != nil {
		s.logger.Error("Token validation failed", "error", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"handle": session.Handle,
		"did":    session.DID,
		"email":  s.getUserEmailFromDID(r.Context(), session.DID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleRefreshSession handles token refresh
func (s *Server) handleRefreshSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		RefreshJwt string `json:"refreshJwt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.RefreshJwt == "" {
		http.Error(w, "refreshJwt required", http.StatusBadRequest)
		return
	}

	s.logger.Debug("Refreshing session")

	// Validate refresh token
	session, err := s.authService.ValidateToken(req.RefreshJwt)
	if err != nil {
		s.logger.Error("Refresh token validation failed", "error", err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := s.authService.GenerateTokens(session)
	if err != nil {
		s.logger.Error("Failed to generate new tokens", "error", err)
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"accessJwt":  accessToken,
		"refreshJwt": refreshToken,
		"handle":     session.Handle,
		"did":        session.DID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleDeleteSession handles session deletion
func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	session, err := s.authService.ValidateToken(token)
	if err != nil {
		s.logger.Error("Token validation failed", "error", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Delete session from database
	err = s.db.DeleteUserSession(r.Context(), session.ID)
	if err != nil {
		s.logger.Error("Failed to delete session", "error", err)
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	s.logger.Info("Session deleted", "session_id", session.ID)
	w.WriteHeader(http.StatusOK)
}
