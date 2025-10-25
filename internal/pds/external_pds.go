package pds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/golang-jwt/jwt/v5"
)

// ExternalPDSClient handles authentication and communication with external PDS servers
type ExternalPDSClient struct {
	directory  identity.Directory
	logger     *slog.Logger
	httpClient *http.Client
	// mockPDSUrl is used for testing to override PDS endpoint resolution
	mockPDSUrl string
	// mockPDSPublicKey is used for testing to override public key fetching
	mockPDSPublicKey crypto.PublicKey
}

// NewExternalPDSClient creates a new external PDS client
func NewExternalPDSClient(directory identity.Directory, logger *slog.Logger) *ExternalPDSClient {
	return &ExternalPDSClient{
		directory:  directory,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewExternalPDSClientForTesting creates a new external PDS client with a mock PDS URL for testing
func NewExternalPDSClientForTesting(directory identity.Directory, logger *slog.Logger, mockPDSUrl string, mockPDSPublicKey crypto.PublicKey) *ExternalPDSClient {
	return &ExternalPDSClient{
		directory:        directory,
		logger:           logger,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		mockPDSUrl:       mockPDSUrl,
		mockPDSPublicKey: mockPDSPublicKey,
	}
}

// AuthenticateUser authenticates a user against their home PDS
func (c *ExternalPDSClient) AuthenticateUser(ctx context.Context, did, identifier, password string) (*Session, error) {
	c.logger.Debug("Authenticating user with external PDS", "did", did, "identifier", identifier)

	// Resolve PDS endpoint from DID
	pdsEndpoint, err := c.ResolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	// Create session request to external PDS
	sessionReq := map[string]string{
		"identifier": identifier,
		"password":   password,
	}

	// Make request to external PDS
	resp, err := c.makeRequest(ctx, "POST", pdsEndpoint+"/xrpc/com.atproto.server.createSession", sessionReq)
	if err != nil {
		return nil, fmt.Errorf("external PDS authentication failed: %w", err)
	}

	// Parse response
	var sessionResp struct {
		AccessJwt  string `json:"accessJwt"`
		RefreshJwt string `json:"refreshJwt"`
		Handle     string `json:"handle"`
		DID        string `json:"did"`
		Email      string `json:"email,omitempty"`
	}

	if err := json.Unmarshal(resp, &sessionResp); err != nil {
		return nil, fmt.Errorf("failed to parse external PDS response: %w", err)
	}

	// Create session object
	session := &Session{
		ID:        generateSessionID(),
		DID:       sessionResp.DID,
		Handle:    sessionResp.Handle,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	c.logger.Info("External PDS authentication successful", "did", did, "handle", sessionResp.Handle)
	return session, nil
}

// ResolvePDSEndpoint resolves the PDS endpoint from a DID document
func (c *ExternalPDSClient) ResolvePDSEndpoint(ctx context.Context, did string) (string, error) {
	c.logger.Debug("Resolving PDS endpoint", "did", did)

	// Look up DID in identity directory
	identity, err := c.directory.LookupDID(ctx, syntax.DID(did))
	if err != nil {
		return "", fmt.Errorf("failed to resolve DID: %w", err)
	}

	// Extract PDS endpoint from DID document
	// For now, we'll use a simple heuristic - look for atproto service endpoints
	// In a full implementation, this would parse the DID document properly
	pdsEndpoint := c.extractPDSEndpointFromIdentity(identity)
	if pdsEndpoint == "" {
		return "", fmt.Errorf("no PDS endpoint found in DID document")
	}

	c.logger.Debug("PDS endpoint resolved", "did", did, "endpoint", pdsEndpoint)
	return pdsEndpoint, nil
}

// ValidateSessionToken validates a JWT token issued by an external PDS
func (c *ExternalPDSClient) ValidateSessionToken(ctx context.Context, token string) (*Session, error) {
	c.logger.Debug("Validating external PDS session token")

	// Parse token without validation to extract claims
	// Use ParseUnverified to skip signature validation temporarily
	tempToken, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := tempToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract user information
	did, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing subject (DID) in token")
	}

	handle, ok := claims["handle"].(string)
	if !ok {
		return nil, fmt.Errorf("missing handle in token")
	}

	issuer, ok := claims["iss"].(string)
	if !ok {
		return nil, fmt.Errorf("missing issuer in token")
	}

	// Resolve PDS endpoint and fetch public key
	pdsEndpoint, err := c.ResolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint: %w", err)
	}

	// Fetch and validate with PDS public key
	publicKey, err := c.fetchPDSPublicKey(ctx, pdsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PDS public key: %w", err)
	}

	c.logger.Debug("About to validate JWT token", "publicKeyType", fmt.Sprintf("%T", publicKey), "publicKeyNil", publicKey == nil)

	// Validate token signature with ECDSA public key
	validatedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		// Verify ES256 signing method
		if t.Method.Alg() != "ES256" {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		c.logger.Debug("Inside JWT validation callback", "tokenMethodType", fmt.Sprintf("%T", t.Method), "publicKeyType", fmt.Sprintf("%T", publicKey))
		// Return the public key directly - the JWT library should handle the type conversion
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !validatedToken.Valid {
		return nil, fmt.Errorf("invalid token signature")
	}

	// Create session object
	session := &Session{
		ID:        generateSessionID(),
		DID:       did,
		Handle:    handle,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	c.logger.Info("External PDS token validated", "did", did, "handle", handle, "issuer", issuer)
	return session, nil
}

// fetchPDSPublicKey fetches the public key from a PDS server
func (c *ExternalPDSClient) fetchPDSPublicKey(ctx context.Context, pdsEndpoint string) (crypto.PublicKey, error) {
	// For testing, use the mock public key if available
	if c.mockPDSPublicKey != nil {
		return c.mockPDSPublicKey, nil
	}

	// Make request to PDS server's public key endpoint
	// This is a simplified implementation - in practice, this would use
	// the proper atproto server discovery mechanisms
	resp, err := c.makeRequest(ctx, "GET", pdsEndpoint+"/xrpc/com.atproto.server.describeServer", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PDS server info: %w", err)
	}

	// Parse server response to extract public key
	var serverInfo struct {
		PublicKey string `json:"publicKey"`
	}

	if err := json.Unmarshal(resp, &serverInfo); err != nil {
		return nil, fmt.Errorf("failed to parse server info: %w", err)
	}

	// Parse the public key (this is simplified - would need proper key parsing)
	// For now, return a mock key - in production this would parse the actual key
	return nil, fmt.Errorf("public key parsing not implemented")
}

// extractPDSEndpointFromIdentity extracts PDS endpoint from identity information
func (c *ExternalPDSClient) extractPDSEndpointFromIdentity(identity *identity.Identity) string {
	// This is a simplified implementation
	// In practice, this would parse the DID document to find atproto service endpoints

	// For development/testing, we'll use a mock endpoint
	// In production, this would extract from the actual DID document
	handle := identity.Handle.String()
	if strings.Contains(handle, ".hashpost.local") {
		// This is a local user, return our PDS endpoint
		return "http://localhost:8080"
	}

	// For external users in tests, use the mock PDS URL if available
	if c.mockPDSUrl != "" && strings.Contains(handle, ".example.com") {
		// This is a test external user, return the mock external PDS endpoint
		// In a real implementation, this would be extracted from the DID document
		return c.mockPDSUrl
	}

	// For external users, we'd need to parse the DID document
	// For now, return empty to indicate no PDS found
	return ""
}

// makeRequest makes an HTTP request to an external PDS
func (c *ExternalPDSClient) makeRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	// Simple implementation - in production, use crypto/rand
	return fmt.Sprintf("ext_%d", time.Now().UnixNano())
}
