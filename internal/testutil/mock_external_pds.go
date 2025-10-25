package testutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/golang-jwt/jwt/v5"
)

// MockExternalPDS represents a mock external PDS server for testing
type MockExternalPDS struct {
	server          *httptest.Server
	users           map[string]*MockUser
	privateKeyECDSA *ecdsa.PrivateKey // For JWT signing
	publicKeyIndigo crypto.PublicKey  // For validation with Indigo
	logger          *slog.Logger
}

// MockUser represents a user in the mock external PDS
type MockUser struct {
	DID      string
	Handle   string
	Password string
	Email    string
	Active   bool
}

// NewMockExternalPDS creates a new mock external PDS server
func NewMockExternalPDS(logger *slog.Logger) *MockExternalPDS {
	// Generate ECDSA key pair for JWT signing
	privateKeyECDSA, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate ECDSA key pair: %v", err))
	}

	// Convert to Indigo format for validation
	// Get the compressed public key bytes
	pubBytes := elliptic.MarshalCompressed(privateKeyECDSA.Curve, privateKeyECDSA.X, privateKeyECDSA.Y)
	publicKeyIndigo, err := crypto.ParsePublicBytesP256(pubBytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert to Indigo public key: %v", err))
	}

	mock := &MockExternalPDS{
		users:           make(map[string]*MockUser),
		privateKeyECDSA: privateKeyECDSA,
		publicKeyIndigo: publicKeyIndigo,
		logger:          logger,
	}

	// Create test server
	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	// Add some test users
	mock.addTestUsers()

	return mock
}

// Close shuts down the mock external PDS server
func (m *MockExternalPDS) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// URL returns the base URL of the mock external PDS server
func (m *MockExternalPDS) URL() string {
	return m.server.URL
}

// PublicKey returns the public key of the mock external PDS server in Indigo format
func (m *MockExternalPDS) PublicKey() crypto.PublicKey {
	return m.publicKeyIndigo
}

// AddUser adds a user to the mock external PDS
func (m *MockExternalPDS) AddUser(did, handle, password, email string) {
	m.users[did] = &MockUser{
		DID:      did,
		Handle:   handle,
		Password: password,
		Email:    email,
		Active:   true,
	}
}

// GetUser retrieves a user from the mock external PDS
func (m *MockExternalPDS) GetUser(did string) (*MockUser, bool) {
	user, exists := m.users[did]
	return user, exists
}

// GenerateAccessToken generates an access token for a user (for testing)
func (m *MockExternalPDS) GenerateAccessToken(user *MockUser) (string, error) {
	return m.generateAccessToken(user)
}

// handleRequest handles HTTP requests to the mock external PDS
func (m *MockExternalPDS) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/xrpc/com.atproto.server.createSession":
		m.handleCreateSession(w, r)
	case path == "/xrpc/com.atproto.server.refreshSession":
		m.handleRefreshSession(w, r)
	case path == "/xrpc/com.atproto.identity.resolveHandle":
		m.handleResolveHandle(w, r)
	case path == "/xrpc/com.atproto.server.describeServer":
		m.handleDescribeServer(w, r)
	case path == "/.well-known/did.json":
		m.handleDIDDocument(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleCreateSession handles com.atproto.server.createSession requests
func (m *MockExternalPDS) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find user by identifier
	var user *MockUser
	for _, u := range m.users {
		if u.Handle == req.Identifier || u.DID == req.Identifier {
			user = u
			break
		}
	}

	if user == nil || user.Password != req.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT tokens
	accessToken, err := m.generateAccessToken(user)
	if err != nil {
		m.logger.Error("Failed to generate access token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := m.generateRefreshToken(user)
	if err != nil {
		m.logger.Error("Failed to generate refresh token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return session response
	response := map[string]interface{}{
		"accessJwt":  accessToken,
		"refreshJwt": refreshToken,
		"handle":     user.Handle,
		"did":        user.DID,
		"email":      user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRefreshSession handles com.atproto.server.refreshSession requests
func (m *MockExternalPDS) handleRefreshSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// For simplicity, we'll just generate new tokens
	// In a real implementation, we'd validate the refresh token
	response := map[string]interface{}{
		"accessJwt":  "new_access_token",
		"refreshJwt": "new_refresh_token",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleResolveHandle handles com.atproto.identity.resolveHandle requests
func (m *MockExternalPDS) handleResolveHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handle := r.URL.Query().Get("handle")
	if handle == "" {
		http.Error(w, "handle parameter required", http.StatusBadRequest)
		return
	}

	// Find user by handle
	var user *MockUser
	for _, u := range m.users {
		if u.Handle == handle {
			user = u
			break
		}
	}

	if user == nil {
		http.Error(w, "Handle not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"did": user.DID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDIDDocument handles DID document requests
func (m *MockExternalPDS) handleDIDDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return a mock DID document
	didDoc := map[string]interface{}{
		"@context": []string{"https://www.w3.org/ns/did/v1"},
		"id":       "did:plc:mock-external-pds",
		"service": []map[string]interface{}{
			{
				"id":              "atproto_pds",
				"type":            "AtprotoPersonalDataServer",
				"serviceEndpoint": m.URL(),
			},
		},
		"verificationMethod": []map[string]interface{}{
			{
				"id":         "did:plc:mock-external-pds#key-1",
				"type":       "EcdsaSecp256k1VerificationKey2019",
				"controller": "did:plc:mock-external-pds",
				"publicKeyJwk": map[string]interface{}{
					"kty": "EC",
					"crv": "secp256k1",
					"x":   "mock_x_coordinate",
					"y":   "mock_y_coordinate",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(didDoc)
}

// handleDescribeServer handles com.atproto.server.describeServer requests
func (m *MockExternalPDS) handleDescribeServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return server information including public key
	// For testing, we'll return the public key in a format that can be parsed
	serverInfo := map[string]interface{}{
		"publicKey": "mock-public-key-for-testing",
		"version":   "1.0.0",
		"name":      "Mock External PDS",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serverInfo)
}

// generateAccessToken generates a JWT access token for a user
func (m *MockExternalPDS) generateAccessToken(user *MockUser) (string, error) {
	claims := jwt.MapClaims{
		"iss":    m.URL(),
		"sub":    user.DID,
		"aud":    "com.atproto.access",
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
		"handle": user.Handle,
		"scope":  "com.atproto.access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(m.privateKeyECDSA)
}

// generateRefreshToken generates a JWT refresh token for a user
func (m *MockExternalPDS) generateRefreshToken(user *MockUser) (string, error) {
	claims := jwt.MapClaims{
		"iss":   m.URL(),
		"sub":   user.DID,
		"aud":   "com.atproto.refresh",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
		"scope": "com.atproto.refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(m.privateKeyECDSA)
}

// addTestUsers adds some test users to the mock external PDS
func (m *MockExternalPDS) addTestUsers() {
	testUsers := []struct {
		did      string
		handle   string
		password string
		email    string
	}{
		{
			did:      "did:plc:test-user-1",
			handle:   "testuser1.example.com",
			password: "testpassword1",
			email:    "testuser1@example.com",
		},
		{
			did:      "did:plc:test-user-2",
			handle:   "testuser2.example.com",
			password: "testpassword2",
			email:    "testuser2@example.com",
		},
		{
			did:      "did:plc:test-user-3",
			handle:   "testuser3.example.com",
			password: "testpassword3",
			email:    "testuser3@example.com",
		},
		{
			did:      "did:plc:external-user-test",
			handle:   "externaluser.example.com",
			password: "testpassword",
			email:    "externaluser@example.com",
		},
		{
			did:      "did:plc:integration-test",
			handle:   "integrationuser.example.com",
			password: "testpassword",
			email:    "integrationuser@example.com",
		},
	}

	for _, user := range testUsers {
		m.AddUser(user.did, user.handle, user.password, user.email)
	}
}

// CreateMockExternalPDSIdentity creates a mock identity for the external PDS
func CreateMockExternalPDSIdentity(pdsURL string) identity.Identity {
	return identity.Identity{
		DID:    syntax.DID("did:plc:mock-external-pds"),
		Handle: syntax.Handle("mock-external-pds.example.com"),
		// Add other fields as needed
	}
}
