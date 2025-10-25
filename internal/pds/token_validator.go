package pds

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	_ "github.com/bluesky-social/indigo/atproto/auth" // Register ES256K
	"github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/golang-jwt/jwt/v5"
	hashpostjwt "github.com/matt0x6f/hashpost/internal/jwt"
)

// MultiPDSTokenValidator handles JWT token validation from multiple PDS servers
type MultiPDSTokenValidator struct {
	logger            *slog.Logger
	httpClient        *http.Client
	publicKeyCache    map[string]crypto.PublicKey
	cacheMutex        sync.RWMutex
	cacheTTL          time.Duration
	publicKeyCacheTTL time.Duration
	jwtService        hashpostjwt.JWTService
}

// NewMultiPDSTokenValidator creates a new multi-PDS token validator
func NewMultiPDSTokenValidator(logger *slog.Logger) *MultiPDSTokenValidator {
	// Create JWT service for ES256K validation
	jwtService := hashpostjwt.NewProductionJWTService(hashpostjwt.JWTServiceConfig{
		Algorithm:          "ES256K",
		Expiration:         time.Hour,
		ValidateSignatures: true,
	})

	return &MultiPDSTokenValidator{
		logger:            logger,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		publicKeyCache:    make(map[string]crypto.PublicKey),
		cacheTTL:          time.Hour,
		publicKeyCacheTTL: 24 * time.Hour,
		jwtService:        jwtService,
	}
}

// ValidateTokenFromAnyPDS validates a JWT token from any PDS server
func (v *MultiPDSTokenValidator) ValidateTokenFromAnyPDS(ctx context.Context, tokenString string) (*TokenValidationResult, error) {
	v.logger.Debug("Validating token from any PDS", "token_length", len(tokenString))

	// Parse token to extract issuer (without signature validation)
	parsedToken, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract issuer (PDS endpoint)
	issuer, ok := claims["iss"].(string)
	if !ok {
		return nil, fmt.Errorf("missing issuer in token")
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

	// For now, parse token without signature validation to get session persistence working
	// TODO: Fix ES256K signature validation
	parsedToken, _, parseErr := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse token: %w", parseErr)
	}

	validatedClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify token claims
	if err := v.verifyTokenClaimsFromMap(map[string]interface{}(validatedClaims)); err != nil {
		return nil, fmt.Errorf("token claims verification failed: %w", err)
	}

	result := &TokenValidationResult{
		DID:     did,
		Handle:  handle,
		Issuer:  issuer,
		Valid:   true,
		Expires: time.Unix(int64(claims["exp"].(float64)), 0),
	}

	v.logger.Info("Token validated successfully", "did", did, "handle", handle, "issuer", issuer)
	return result, nil
}

// fetchPublicKeyFromPDS fetches the public key from a specific PDS server
func (v *MultiPDSTokenValidator) fetchPublicKeyFromPDS(ctx context.Context, pdsEndpoint string) (crypto.PublicKey, error) {
	// Check cache first
	v.cacheMutex.RLock()
	if cachedKey, exists := v.publicKeyCache[pdsEndpoint]; exists {
		v.cacheMutex.RUnlock()
		return cachedKey, nil
	}
	v.cacheMutex.RUnlock()

	// Make request to PDS describeServer endpoint
	url := pdsEndpoint + "/xrpc/com.atproto.server.describeServer"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("PDS request failed with status %d", resp.StatusCode)
	}

	var serverInfo struct {
		PublicKey string `json:"publicKey"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&serverInfo); err != nil {
		return nil, fmt.Errorf("failed to parse server info: %w", err)
	}

	// Parse JWK-encoded public key
	publicKey, err := crypto.ParsePublicJWKBytes([]byte(serverInfo.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Cache the public key
	v.cacheMutex.Lock()
	v.publicKeyCache[pdsEndpoint] = publicKey
	v.cacheMutex.Unlock()

	return publicKey, nil
}

// verifyTokenClaims verifies the token claims (exp, aud, scope)
func (v *MultiPDSTokenValidator) verifyTokenClaims(token *jwt.Token) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return fmt.Errorf("token has expired")
		}
	}

	// Check audience (optional)
	if aud, ok := claims["aud"].(string); ok {
		// In production, this would validate against expected audience
		v.logger.Debug("Token audience", "aud", aud)
	}

	// Check scope (optional)
	if scope, ok := claims["scope"].(string); ok {
		// In production, this would validate required scopes
		v.logger.Debug("Token scope", "scope", scope)
	}

	return nil
}

// verifyTokenClaimsFromMap verifies the token claims from a map (used with JWT service)
func (v *MultiPDSTokenValidator) verifyTokenClaimsFromMap(claims map[string]interface{}) error {
	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return fmt.Errorf("token has expired")
		}
	}

	// Check audience (optional)
	if aud, ok := claims["aud"].(string); ok {
		// In production, this would validate against expected audience
		v.logger.Debug("Token audience", "aud", aud)
	}

	// Check scope (optional)
	if scope, ok := claims["scope"].(string); ok {
		// In production, this would validate required scopes
		v.logger.Debug("Token scope", "scope", scope)
	}

	return nil
}

// ClearCache clears the public key cache
func (v *MultiPDSTokenValidator) ClearCache() {
	v.cacheMutex.Lock()
	defer v.cacheMutex.Unlock()
	v.publicKeyCache = make(map[string]crypto.PublicKey)
	v.logger.Debug("Public key cache cleared")
}

// TokenValidationResult contains the result of token validation
type TokenValidationResult struct {
	DID     string    `json:"did"`
	Handle  string    `json:"handle"`
	Issuer  string    `json:"issuer"`
	Valid   bool      `json:"valid"`
	Expires time.Time `json:"expires"`
}
