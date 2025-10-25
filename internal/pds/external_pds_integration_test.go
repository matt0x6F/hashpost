//go:build integration

package pds

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/matt0x6f/hashpost/internal/testutil"
)

func TestExternalPDSIntegration(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	// Create identity directory with mock external PDS
	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Add test user identities to the directory with PDS endpoint information
	testUsers := []identity.Identity{
		{
			DID:    syntax.DID("did:plc:test-user-1"),
			Handle: syntax.Handle("testuser1.example.com"),
		},
		{
			DID:    syntax.DID("did:plc:test-user-2"),
			Handle: syntax.Handle("testuser2.example.com"),
		},
		{
			DID:    syntax.DID("did:plc:test-user-3"),
			Handle: syntax.Handle("testuser3.example.com"),
		},
	}

	for _, user := range testUsers {
		mockDir.Insert(user)
	}

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	t.Run("complete external PDS authentication flow", func(t *testing.T) {
		ctx := context.Background()

		// Test user credentials
		did := "did:plc:test-user-1"
		handle := "testuser1.example.com"
		password := "testpassword1"

		// Step 1: Resolve PDS endpoint
		endpoint, err := client.ResolvePDSEndpoint(ctx, did)
		if err != nil {
			t.Fatalf("ResolvePDSEndpoint() error = %v", err)
		}

		if endpoint == "" {
			t.Fatal("ResolvePDSEndpoint() returned empty endpoint")
		}

		t.Logf("Resolved PDS endpoint: %s", endpoint)

		// Step 2: Authenticate user
		session, err := client.AuthenticateUser(ctx, did, handle, password)
		if err != nil {
			t.Fatalf("AuthenticateUser() error = %v", err)
		}

		if session == nil {
			t.Fatal("AuthenticateUser() returned nil session")
		}

		if session.DID != did {
			t.Errorf("AuthenticateUser() DID = %v, want %v", session.DID, did)
		}

		if session.Handle != handle {
			t.Errorf("AuthenticateUser() Handle = %v, want %v", session.Handle, handle)
		}

		t.Logf("Authentication successful: DID=%s, Handle=%s", session.DID, session.Handle)

		// Step 3: Validate session token
		// For testing, we'll generate a valid token directly from the mock PDS
		// In a real scenario, this token would come from the authentication response
		mockUser, exists := mockPDS.GetUser(did)
		if !exists {
			t.Fatalf("Mock user not found: %s", did)
		}

		// Generate a valid token from the mock PDS
		token, err := mockPDS.GenerateAccessToken(mockUser)
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		validatedSession, err := client.ValidateSessionToken(ctx, token)
		if err != nil {
			t.Fatalf("ValidateSessionToken() error = %v", err)
		}

		if validatedSession == nil {
			t.Fatal("ValidateSessionToken() returned nil session")
		}

		t.Logf("Token validation successful")
	})

	t.Run("external PDS error handling", func(t *testing.T) {
		ctx := context.Background()

		// Test invalid credentials
		_, err := client.AuthenticateUser(ctx, "did:plc:test-user-1", "testuser1.example.com", "wrongpassword")
		if err == nil {
			t.Errorf("AuthenticateUser() expected error for invalid credentials")
		}

		// Test non-existent user
		_, err = client.AuthenticateUser(ctx, "did:plc:non-existent", "nonexistent.example.com", "password")
		if err == nil {
			t.Errorf("AuthenticateUser() expected error for non-existent user")
		}

		t.Logf("Error handling tests passed")
	})

	t.Run("external PDS performance", func(t *testing.T) {
		ctx := context.Background()
		concurrency := 5
		results := make(chan error, concurrency)

		// Run concurrent authentication requests
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				did := "did:plc:test-user-1"
				handle := "testuser1.example.com"
				password := "testpassword1"

				_, err := client.AuthenticateUser(ctx, did, handle, password)
				results <- err
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < concurrency; i++ {
			err := <-results
			if err == nil {
				successCount++
			}
		}

		t.Logf("Concurrent authentication: %d/%d successful", successCount, concurrency)
	})
}

func TestExternalPDSMultiServerIntegration(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create multiple mock external PDS servers
	mockPDS1 := testutil.NewMockExternalPDS(logger)
	defer mockPDS1.Close()

	mockPDS2 := testutil.NewMockExternalPDS(logger)
	defer mockPDS2.Close()

	// Create identity directory with multiple external PDS servers
	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS1.URL()))
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS2.URL()))

	// Add test user identities to the directory
	testUsers := []identity.Identity{
		{
			DID:    syntax.DID("did:plc:test-user-1"),
			Handle: syntax.Handle("testuser1.example.com"),
		},
		{
			DID:    syntax.DID("did:plc:test-user-2"),
			Handle: syntax.Handle("testuser2.example.com"),
		},
	}

	for _, user := range testUsers {
		mockDir.Insert(user)
	}

	// Create external PDS client with mock PDS URLs and public keys
	// For multi-server test, we'll use PDS1 as the primary mock
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS1.URL(), mockPDS1.PublicKey())

	t.Run("multi-server authentication", func(t *testing.T) {
		ctx := context.Background()

		// Test authentication with first PDS
		session1, err := client.AuthenticateUser(ctx, "did:plc:test-user-1", "testuser1.example.com", "testpassword1")
		if err != nil {
			t.Fatalf("AuthenticateUser() with PDS1 error = %v", err)
		}

		if session1 == nil {
			t.Fatal("AuthenticateUser() with PDS1 returned nil session")
		}

		t.Logf("PDS1 authentication successful: %s", session1.DID)

		// Test authentication with second PDS
		session2, err := client.AuthenticateUser(ctx, "did:plc:test-user-2", "testuser2.example.com", "testpassword2")
		if err != nil {
			t.Fatalf("AuthenticateUser() with PDS2 error = %v", err)
		}

		if session2 == nil {
			t.Fatal("AuthenticateUser() with PDS2 returned nil session")
		}

		t.Logf("PDS2 authentication successful: %s", session2.DID)

		// Verify different sessions
		if session1.DID == session2.DID {
			t.Errorf("Sessions from different PDS servers should have different DIDs")
		}
	})
}

func TestExternalPDSTokenValidationIntegration(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	// Create identity directory
	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Add test user identity
	testUser := identity.Identity{
		DID:    syntax.DID("did:plc:test-user-1"),
		Handle: syntax.Handle("testuser1.example.com"),
	}
	mockDir.Insert(testUser)

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	t.Run("token validation with public key", func(t *testing.T) {
		ctx := context.Background()

		// Get public key from mock PDS
		publicKey := mockPDS.PublicKey()
		if publicKey == nil {
			t.Fatal("Mock PDS public key is nil")
		}

		t.Logf("Mock PDS public key: %+v", publicKey)

		// Generate a real token from the mock PDS
		did := "did:plc:test-user-1"
		mockUser, exists := mockPDS.GetUser(did)
		if !exists {
			t.Fatalf("Mock user not found: %s", did)
		}

		// Generate a valid token from the mock PDS
		token, err := mockPDS.GenerateAccessToken(mockUser)
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Test token validation
		session, err := client.ValidateSessionToken(ctx, token)
		if err != nil {
			t.Fatalf("ValidateSessionToken() error = %v", err)
		}

		if session == nil {
			t.Fatal("ValidateSessionToken() returned nil session")
		}

		t.Logf("Token validation successful with public key")
	})
}

func TestExternalPDSEndToEndIntegration(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	// Create identity directory
	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Add test user identity
	testUser := identity.Identity{
		DID:    syntax.DID("did:plc:external-user-test"),
		Handle: syntax.Handle("externaluser.example.com"),
	}
	mockDir.Insert(testUser)

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	t.Run("end-to-end external PDS flow", func(t *testing.T) {
		ctx := context.Background()

		// Step 1: User registration (simulated)
		did := "did:plc:external-user-test"
		handle := "externaluser.example.com"
		password := "testpassword"

		t.Logf("Step 1: User registration simulated")

		// Step 2: PDS discovery
		endpoint, err := client.ResolvePDSEndpoint(ctx, did)
		if err != nil {
			t.Fatalf("ResolvePDSEndpoint() error = %v", err)
		}

		t.Logf("Step 2: PDS discovery successful: %s", endpoint)

		// Step 3: User authentication
		session, err := client.AuthenticateUser(ctx, did, handle, password)
		if err != nil {
			t.Fatalf("AuthenticateUser() error = %v", err)
		}

		t.Logf("Step 3: User authentication successful: %s", session.DID)

		// Step 4: Token validation
		// Generate a real token from the mock PDS
		mockUser, exists := mockPDS.GetUser(did)
		if !exists {
			t.Fatalf("Mock user not found: %s", did)
		}

		token, err := mockPDS.GenerateAccessToken(mockUser)
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		validatedSession, err := client.ValidateSessionToken(ctx, token)
		if err != nil {
			t.Fatalf("ValidateSessionToken() error = %v", err)
		}

		t.Logf("Step 4: Token validation successful")

		// Step 5: Session management
		if session.DID != validatedSession.DID {
			t.Errorf("Session DIDs don't match: %s vs %s", session.DID, validatedSession.DID)
		}

		t.Logf("Step 5: Session management successful")

		// Step 6: User activity (simulated)
		time.Sleep(100 * time.Millisecond) // Simulate user activity

		t.Logf("Step 6: User activity simulated")

		t.Logf("End-to-end external PDS flow completed successfully")
	})
}
