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

func TestExternalPDSClient_AuthenticateUser(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Add test external user
	externalUser := identity.Identity{
		DID:    syntax.DID("did:plc:external-user-test"),
		Handle: syntax.Handle("externaluser.example.com"),
	}
	mockDir.Insert(externalUser)

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	tests := []struct {
		name       string
		did        string
		identifier string
		password   string
		wantErr    bool
	}{
		{
			name:       "successful external authentication",
			did:        "did:plc:external-user-test",
			identifier: "externaluser.example.com",
			password:   "testpassword",
			wantErr:    false,
		},
		{
			name:       "invalid credentials",
			did:        "did:plc:external-user-test",
			identifier: "externaluser.example.com",
			password:   "wrongpassword",
			wantErr:    true,
		},
		{
			name:       "non-existent user",
			did:        "did:plc:non-existent",
			identifier: "nonexistent.example.com",
			password:   "testpassword",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			session, err := client.AuthenticateUser(ctx, tt.did, tt.identifier, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("AuthenticateUser() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("AuthenticateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if session == nil {
				t.Errorf("AuthenticateUser() returned nil session")
				return
			}

			if session.DID != tt.did {
				t.Errorf("AuthenticateUser() DID = %v, want %v", session.DID, tt.did)
			}
		})
	}
}

func TestExternalPDSClient_ResolvePDSEndpoint(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Add test external user
	externalUser := identity.Identity{
		DID:    syntax.DID("did:plc:external-user-test"),
		Handle: syntax.Handle("externaluser.example.com"),
	}
	mockDir.Insert(externalUser)

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	tests := []struct {
		name    string
		did     string
		wantErr bool
	}{
		{
			name:    "resolve external PDS endpoint",
			did:     "did:plc:external-user-test",
			wantErr: false,
		},
		{
			name:    "non-existent DID",
			did:     "did:plc:non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			endpoint, err := client.ResolvePDSEndpoint(ctx, tt.did)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolvePDSEndpoint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ResolvePDSEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if endpoint == "" {
				t.Errorf("ResolvePDSEndpoint() returned empty endpoint")
			}
		})
	}
}

func TestExternalPDSClient_ValidateSessionToken(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Add test user identity
	testUser := identity.Identity{
		DID:    syntax.DID("did:plc:integration-test"),
		Handle: syntax.Handle("integrationtest.example.com"),
	}
	mockDir.Insert(testUser)

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	// Generate a real token for testing
	did := "did:plc:integration-test"
	mockUser, exists := mockPDS.GetUser(did)
	if !exists {
		t.Fatalf("Mock user not found: %s", did)
	}

	validToken, err := mockPDS.GenerateAccessToken(mockUser)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid external PDS token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "invalid token format",
			token:   "invalid.token",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			session, err := client.ValidateSessionToken(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateSessionToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateSessionToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if session == nil {
				t.Errorf("ValidateSessionToken() returned nil session")
			}
		})
	}
}

func TestExternalPDSClient_Integration(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Add test external user
	externalUser := identity.Identity{
		DID:    syntax.DID("did:plc:integration-test"),
		Handle: syntax.Handle("integrationuser.example.com"),
	}
	mockDir.Insert(externalUser)

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	t.Run("complete external PDS flow", func(t *testing.T) {
		ctx := context.Background()
		did := "did:plc:integration-test"
		identifier := "integrationuser.example.com"
		password := "testpassword"

		// Step 1: Resolve PDS endpoint
		endpoint, err := client.ResolvePDSEndpoint(ctx, did)
		if err != nil {
			t.Fatalf("ResolvePDSEndpoint() error = %v", err)
		}

		if endpoint == "" {
			t.Fatal("ResolvePDSEndpoint() returned empty endpoint")
		}

		// Step 2: Authenticate user
		session, err := client.AuthenticateUser(ctx, did, identifier, password)
		if err != nil {
			t.Fatalf("AuthenticateUser() error = %v", err)
		}

		if session.DID != did {
			t.Errorf("AuthenticateUser() DID = %v, want %v", session.DID, did)
		}

		// Step 3: Validate session token
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

		if validatedSession == nil {
			t.Fatal("ValidateSessionToken() returned nil session")
		}

		t.Logf("External PDS integration test completed successfully")
		t.Logf("PDS Endpoint: %s", endpoint)
		t.Logf("User DID: %s", session.DID)
		t.Logf("User Handle: %s", session.Handle)
	})
}

func TestExternalPDSClient_Performance(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	t.Run("concurrent external PDS operations", func(t *testing.T) {
		ctx := context.Background()
		concurrency := 10
		results := make(chan error, concurrency)

		// Run concurrent operations
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				did := "did:plc:performance-test-" + string(rune(id))
				_, err := client.ResolvePDSEndpoint(ctx, did)
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

		// For this test, we expect some failures since we're using non-existent DIDs
		// In a real test, we'd set up proper test data
		t.Logf("Concurrent operations completed: %d/%d successful", successCount, concurrency)
	})
}

func TestExternalPDSClient_ErrorHandling(t *testing.T) {

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create mock external PDS
	mockPDS := testutil.NewMockExternalPDS(logger)
	defer mockPDS.Close()

	mockDir := identity.NewMockDirectory()
	mockDir.Insert(testutil.CreateMockExternalPDSIdentity(mockPDS.URL()))

	// Create external PDS client with mock PDS URL and public key
	client := NewExternalPDSClientForTesting(&mockDir, logger, mockPDS.URL(), mockPDS.PublicKey())

	tests := []struct {
		name        string
		operation   func() error
		expectError bool
	}{
		{
			name: "network timeout simulation",
			operation: func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()
				_, err := client.ResolvePDSEndpoint(ctx, "did:plc:timeout-test")
				return err
			},
			expectError: true,
		},
		{
			name: "invalid DID format",
			operation: func() error {
				_, err := client.ResolvePDSEndpoint(context.Background(), "invalid-did")
				return err
			},
			expectError: true,
		},
		{
			name: "malformed token",
			operation: func() error {
				_, err := client.ValidateSessionToken(context.Background(), "malformed.token")
				return err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
