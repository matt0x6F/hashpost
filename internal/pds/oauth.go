package pds

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
)

// OAuthService handles OAuth 2.0 authorization flow
type OAuthService struct {
	authService *AuthService
	logger      *slog.Logger
	db          *generated.Queries
}

// OAuthClient represents an OAuth client application
type OAuthClient struct {
	ClientID      string    `json:"client_id"`
	ClientName    string    `json:"client_name"`
	RedirectURIs  []string  `json:"redirect_uris"`
	Scopes        []string  `json:"scopes"`
	GrantTypes    []string  `json:"grant_types"`
	ResponseTypes []string  `json:"response_types"`
	CreatedAt     time.Time `json:"created_at"`
}

// AuthorizationRequest represents an OAuth authorization request
type AuthorizationRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	Nonce               string `json:"nonce"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

// AuthorizationCode represents an OAuth authorization code
type AuthorizationCode struct {
	Code        string    `json:"code"`
	ClientID    string    `json:"client_id"`
	UserDID     string    `json:"user_did"`
	RedirectURI string    `json:"redirect_uri"`
	Scope       string    `json:"scope"`
	ExpiresAt   time.Time `json:"expires_at"`
	Nonce       string    `json:"nonce"`
}

// TokenRequest represents an OAuth token request
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	CodeVerifier string `json:"code_verifier"`
}

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(authService *AuthService, db *generated.Queries, logger *slog.Logger) *OAuthService {
	return &OAuthService{
		authService: authService,
		logger:      logger,
		db:          db,
	}
}

// GetClientMetadata returns OAuth client metadata
func (o *OAuthService) GetClientMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		http.Error(w, "client_id parameter required", http.StatusBadRequest)
		return
	}

	client, err := o.db.GetOAuthClient(r.Context(), clientID)
	if err != nil {
		o.logger.Error("Failed to get OAuth client", "error", err, "client_id", clientID)
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Convert database client to response format
	response := OAuthClient{
		ClientID:      client.ClientID,
		ClientName:    client.ClientName,
		RedirectURIs:  client.RedirectUris,
		Scopes:        client.Scopes,
		GrantTypes:    client.GrantTypes,
		ResponseTypes: client.ResponseTypes,
		CreatedAt:     client.CreatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAuthorization handles OAuth authorization requests
func (o *OAuthService) HandleAuthorization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse authorization request
	req := &AuthorizationRequest{
		ClientID:            r.URL.Query().Get("client_id"),
		RedirectURI:         r.URL.Query().Get("redirect_uri"),
		ResponseType:        r.URL.Query().Get("response_type"),
		Scope:               r.URL.Query().Get("scope"),
		State:               r.URL.Query().Get("state"),
		Nonce:               r.URL.Query().Get("nonce"),
		CodeChallenge:       r.URL.Query().Get("code_challenge"),
		CodeChallengeMethod: r.URL.Query().Get("code_challenge_method"),
	}

	// Validate client
	client, err := o.db.GetOAuthClient(r.Context(), req.ClientID)
	if err != nil {
		o.logger.Error("Invalid OAuth client", "error", err, "client_id", req.ClientID)
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
		return
	}

	// Validate redirect URI
	if !o.isValidRedirectURI(client.RedirectUris, req.RedirectURI) {
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	// Validate response type
	if req.ResponseType != "code" {
		http.Error(w, "Unsupported response_type", http.StatusBadRequest)
		return
	}

	// Check if user is authenticated
	// For now, we'll require authentication via existing session
	// In a full implementation, this would redirect to login if not authenticated
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Redirect to login page
		loginURL := fmt.Sprintf("/login?redirect_uri=%s&client_id=%s&state=%s",
			url.QueryEscape(req.RedirectURI),
			url.QueryEscape(req.ClientID),
			url.QueryEscape(req.State))
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	// Extract and validate token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	session, err := o.authService.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Generate authorization code
	code, err := o.generateAuthorizationCode(req, session.DID)
	if err != nil {
		o.logger.Error("Failed to generate authorization code", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect back to client with authorization code
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s",
		req.RedirectURI,
		url.QueryEscape(code),
		url.QueryEscape(req.State))

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleToken handles OAuth token requests
func (o *OAuthService) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse token request
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate grant type
	if req.GrantType != "authorization_code" {
		http.Error(w, "Unsupported grant_type", http.StatusBadRequest)
		return
	}

	// Validate authorization code
	authCode, err := o.validateAuthorizationCode(req.Code, req.ClientID, req.RedirectURI)
	if err != nil {
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}

	// Resolve handle from DID
	handle, err := o.authService.ResolveHandle(context.Background(), authCode.UserDID)
	if err != nil {
		o.logger.Warn("Failed to resolve handle from DID", "error", err, "did", authCode.UserDID)
		handle = "user.hashpost.local" // Fallback handle
	}

	// Create session for the user
	session := &Session{
		ID:        generateRandomString(32),
		DID:       authCode.UserDID,
		Handle:    handle,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Generate tokens
	accessToken, refreshToken, err := o.authService.GenerateTokens(session)
	if err != nil {
		o.logger.Error("Failed to generate tokens", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return token response
	response := TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		RefreshToken: refreshToken,
		Scope:        authCode.Scope,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (o *OAuthService) isValidRedirectURI(redirectURIs []string, redirectURI string) bool {
	for _, uri := range redirectURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

func (o *OAuthService) generateAuthorizationCode(req *AuthorizationRequest, userDID string) (string, error) {
	// Generate random authorization code
	code := generateRandomString(32)

	// Store authorization code in database
	_, err := o.db.CreateAuthorizationCode(context.Background(), &generated.CreateAuthorizationCodeParams{
		Code:        code,
		ClientID:    req.ClientID,
		UserDid:     userDID,
		RedirectUri: req.RedirectURI,
		Scope:       req.Scope,
		Nonce:       &req.Nonce,
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minute expiry
	})
	if err != nil {
		o.logger.Error("Failed to store authorization code", "error", err)
		return "", fmt.Errorf("failed to store authorization code: %w", err)
	}

	o.logger.Info("Generated authorization code", "code", code, "client_id", req.ClientID, "user_did", userDID)
	return code, nil
}

func (o *OAuthService) validateAuthorizationCode(code, clientID, redirectURI string) (*AuthorizationCode, error) {
	if code == "" {
		return nil, fmt.Errorf("empty authorization code")
	}

	// Retrieve from database and validate
	dbAuthCode, err := o.db.GetAuthorizationCode(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("authorization code not found or expired")
	}

	// Validate client ID
	if dbAuthCode.ClientID != clientID {
		return nil, fmt.Errorf("client ID mismatch")
	}

	// Validate redirect URI
	if dbAuthCode.RedirectUri != redirectURI {
		return nil, fmt.Errorf("redirect URI mismatch")
	}

	// Mark as used
	err = o.db.MarkAuthorizationCodeUsed(context.Background(), code)
	if err != nil {
		o.logger.Error("Failed to mark authorization code as used", "error", err)
		// Don't fail the validation, just log the error
	}

	// Convert to response format
	authCode := &AuthorizationCode{
		Code:        dbAuthCode.Code,
		ClientID:    dbAuthCode.ClientID,
		UserDID:     dbAuthCode.UserDid,
		RedirectURI: dbAuthCode.RedirectUri,
		Scope:       dbAuthCode.Scope,
		ExpiresAt:   dbAuthCode.ExpiresAt,
		Nonce:       *dbAuthCode.Nonce,
	}

	return authCode, nil
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
