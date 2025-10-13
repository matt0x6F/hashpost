package pds

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	jwtservice "github.com/matt0x6f/hashpost/internal/jwt"
)

// DPoPService handles Demonstrating Proof of Possession
type DPoPService struct {
	authService *AuthService
	logger      *slog.Logger
	db          *generated.Queries
	jwtService  jwtservice.JWTService
}

// DPoPProof represents a DPoP proof JWT
type DPoPProof struct {
	Header DPoPHeader `json:"header"`
	Claims DPoPClaims `json:"claims"`
}

// DPoPHeader represents DPoP JWT header
type DPoPHeader struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
	Jwk string `json:"jwk"`
}

// DPoPClaims represents DPoP JWT claims
type DPoPClaims struct {
	JTI   string `json:"jti"`
	IAT   int64  `json:"iat"`
	HTTP  string `json:"http"`
	URI   string `json:"uri"`
	HTM   string `json:"htm"`
	Nonce string `json:"nonce,omitempty"`
}

// DPoPJWK represents a DPoP JWK
type DPoPJWK struct {
	KTY string `json:"kty"`
	CRV string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// DPoPNonce represents a DPoP nonce
type DPoPNonce struct {
	Nonce   string    `json:"nonce"`
	Expires time.Time `json:"expires"`
}

// NewDPoPService creates a new DPoP service
func NewDPoPService(authService *AuthService, db *generated.Queries, logger *slog.Logger, jwtService jwtservice.JWTService) *DPoPService {
	return &DPoPService{
		authService: authService,
		logger:      logger,
		db:          db,
		jwtService:  jwtService,
	}
}

// GenerateNonce generates a DPoP nonce
func (d *DPoPService) GenerateNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate random nonce
	nonce := d.generateRandomNonce()
	expires := time.Now().Add(5 * time.Minute) // 5 minute expiry

	// Store nonce in database
	_, err := d.db.CreateDPoPNonce(r.Context(), &generated.CreateDPoPNonceParams{
		Nonce:     nonce,
		ExpiresAt: expires,
	})
	if err != nil {
		d.logger.Error("Failed to store DPoP nonce", "error", err)
		http.Error(w, "Failed to generate nonce", http.StatusInternalServerError)
		return
	}

	// Clean up expired nonces
	err = d.db.CleanupExpiredDPoPNonces(r.Context())
	if err != nil {
		d.logger.Error("Failed to cleanup expired nonces", "error", err)
		// Don't fail the request, just log the error
	}

	response := DPoPNonce{
		Nonce:   nonce,
		Expires: expires,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateDPoPProof validates a DPoP proof
func (d *DPoPService) ValidateDPoPProof(proofJWT, httpMethod, httpURI string) (*DPoPProof, error) {
	// Use JWT service to validate and parse the token
	claims, err := d.jwtService.ValidateAndParse(proofJWT, func(token interface{}) (interface{}, error) {
		// This function is called by the JWT service to get the public key
		// We need to extract the JWK from the token header and convert it to a public key

		// Parse the JWT header to extract JWK
		parts := strings.Split(proofJWT, ".")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid JWT format")
		}

		headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, fmt.Errorf("failed to decode header: %w", err)
		}

		var header map[string]interface{}
		if err := json.Unmarshal(headerBytes, &header); err != nil {
			return nil, fmt.Errorf("failed to unmarshal header: %w", err)
		}

		// Extract JWK from header
		jwkStr, ok := header["jwk"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid jwk in header")
		}

		var jwk DPoPJWK
		if err := json.Unmarshal([]byte(jwkStr), &jwk); err != nil {
			return nil, fmt.Errorf("invalid jwk format: %w", err)
		}

		// Convert JWK to ECDSA public key using JWT service
		publicKey, err := d.jwtService.JWKToPublicKey(jwk)
		if err != nil {
			return nil, fmt.Errorf("failed to convert jwk to public key: %w", err)
		}

		// Debug: log the key type
		d.logger.Info("DPoP validation", "key_type", fmt.Sprintf("%T", publicKey))

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse DPoP proof: %w", err)
	}

	// Validate required claims
	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("missing jti in DPoP proof")
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing iat in DPoP proof")
	}

	httpClaim, ok := claims["http"].(string)
	if !ok {
		return nil, fmt.Errorf("missing http in DPoP proof")
	}

	uri, ok := claims["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("missing uri in DPoP proof")
	}

	htm, ok := claims["htm"].(string)
	if !ok {
		return nil, fmt.Errorf("missing htm in DPoP proof")
	}

	// Validate HTTP method and URI
	if htm != httpMethod {
		return nil, fmt.Errorf("DPoP proof method mismatch: expected %s, got %s", httpMethod, htm)
	}

	if uri != httpURI {
		return nil, fmt.Errorf("DPoP proof URI mismatch: expected %s, got %s", httpURI, uri)
	}

	// Validate timestamp (not too old, not in future)
	issuedAt := time.Unix(int64(iat), 0)
	now := time.Now()
	if now.Sub(issuedAt) > 5*time.Minute {
		return nil, fmt.Errorf("DPoP proof too old")
	}
	if issuedAt.After(now.Add(1 * time.Minute)) {
		return nil, fmt.Errorf("DPoP proof issued in future")
	}

	// Validate nonce if present
	if nonce, ok := claims["nonce"].(string); ok {
		if !d.validateNonce(nonce) {
			return nil, fmt.Errorf("invalid or expired nonce")
		}
	}

	// Extract JWK from header for the proof object
	parts := strings.Split(proofJWT, ".")
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	jwkStr, ok := header["jwk"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid jwk in header")
	}

	// Create DPoP proof object
	proof := &DPoPProof{
		Header: DPoPHeader{
			Typ: "dpop+jwt",
			Alg: "ES256",
			Jwk: jwkStr,
		},
		Claims: DPoPClaims{
			JTI:   jti,
			IAT:   int64(iat),
			HTTP:  httpClaim,
			URI:   uri,
			HTM:   htm,
			Nonce: getNonceFromClaims(claims),
		},
	}

	return proof, nil
}

// getNonceFromClaims safely extracts the nonce from claims, returning empty string if not present
func getNonceFromClaims(claims map[string]interface{}) string {
	if nonce, ok := claims["nonce"].(string); ok {
		return nonce
	}
	return ""
}

// GenerateDPoPProof generates a DPoP proof for a request
func (d *DPoPService) GenerateDPoPProof(httpMethod, httpURI, nonce string) (string, error) {
	// Generate ECDSA key pair using JWT service
	privateKey, err := d.jwtService.GenerateKeyPair()
	if err != nil {
		return "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create JWK from public key using JWT service
	jwk, err := d.jwtService.CreateJWKFromPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to create JWK: %w", err)
	}

	// Marshal JWK to JSON using JWT service
	jwkBytes, err := d.jwtService.MarshalJWK(jwk)
	if err != nil {
		return "", fmt.Errorf("failed to marshal jwk: %w", err)
	}

	// Create DPoP claims
	now := time.Now()
	claims := map[string]interface{}{
		"jti":  d.generateRandomString(32),
		"iat":  now.Unix(),
		"http": "POST", // HTTP method
		"uri":  httpURI,
		"htm":  httpMethod,
	}

	// Only include nonce if provided
	if nonce != "" {
		claims["nonce"] = nonce
	}

	// Create header with JWK
	header := map[string]interface{}{
		"typ": "dpop+jwt",
		"alg": "ES256",
		"jwk": jwkBytes,
	}

	// Generate signed token using JWT service
	tokenString, err := d.jwtService.GenerateSignedToken(claims, header, privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign DPoP proof: %w", err)
	}

	return tokenString, nil
}

// Helper methods

func (d *DPoPService) generateRandomNonce() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func (d *DPoPService) generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func (d *DPoPService) validateNonce(nonce string) bool {
	// Get nonce from database
	dbNonce, err := d.db.GetDPoPNonce(context.Background(), nonce)
	if err != nil {
		return false
	}

	// Mark as used
	err = d.db.MarkDPoPNonceUsed(context.Background(), nonce)
	if err != nil {
		d.logger.Error("Failed to mark nonce as used", "error", err)
		// Don't fail validation, just log the error
	}

	// Check if nonce exists and is not expired
	return dbNonce != nil && time.Now().Before(dbNonce.ExpiresAt)
}
