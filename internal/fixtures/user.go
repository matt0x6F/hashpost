package fixtures

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/database/models"
)

// CreateTestUserContext creates a test user context
func CreateTestUserContext() *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "test-pseudonym-id",
		DisplayName:       "TestUser",
		Roles:             []string{"user"},
		Capabilities:      []string{"user"},
		TokenType:         "jwt",
	}
}

// CreateTestUserContextForBlocking creates a test user context for blocking tests
func CreateTestUserContextForBlocking(userID int64, activePseudonymID string) *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		Roles:             []string{"user"},
		Capabilities:      []string{"create_content", "vote", "message", "report"},
		ActivePseudonymID: activePseudonymID,
		DisplayName:       "TestUser",
	}
}

// CreateTestPseudonym creates a test pseudonym
func CreateTestPseudonym() *models.Pseudonym {
	return &models.Pseudonym{
		PseudonymID:         "test-pseudonym-id",
		DisplayName:         "TestUser",
		Bio:                 sql.Null[string]{V: "Test bio", Valid: true},
		WebsiteURL:          sql.Null[string]{V: "https://example.com", Valid: true},
		KarmaScore:          sql.Null[int32]{V: 100, Valid: true},
		ShowKarma:           sql.Null[bool]{V: true, Valid: true},
		AllowDirectMessages: sql.Null[bool]{V: true, Valid: true},
		IsActive:            sql.Null[bool]{V: true, Valid: true},
		CreatedAt:           sql.Null[time.Time]{V: time.Now(), Valid: true},
		LastActiveAt:        sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestPseudonymForBlocking creates a test pseudonym for blocking tests
func CreateTestPseudonymForBlocking(pseudonymID, displayName string) *models.Pseudonym {
	return &models.Pseudonym{
		PseudonymID: pseudonymID,
		DisplayName: displayName,
		IsActive:    sql.Null[bool]{V: true, Valid: true},
		CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestUserBlock creates a test user block
func CreateTestUserBlock(blockID int64, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) *models.UserBlock {
	block := &models.UserBlock{
		BlockID:            blockID,
		BlockerPseudonymID: blockerPseudonymID,
		CreatedAt:          sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	if blockedPseudonymID != "" {
		block.BlockedPseudonymID = sql.Null[string]{V: blockedPseudonymID, Valid: true}
		block.BlockedUserID = sql.Null[int64]{Valid: false}
	} else {
		block.BlockedPseudonymID = sql.Null[string]{Valid: false}
		block.BlockedUserID = sql.Null[int64]{V: blockedUserID, Valid: true}
	}

	return block
}

// GenerateTestJWTToken generates a valid JWT token for testing
func GenerateTestJWTToken(userID int64, activePseudonymID string) (string, error) {
	// Create a JWT token using the actual JWT generation logic
	// This ensures the token is valid for the test environment
	claims := &middleware.JWTClaims{
		UserID:            userID,
		Email:             "test@example.com",
		Roles:             []string{"user"},
		Capabilities:      []string{"create_content", "vote", "message", "report", "create_subforum"},
		MFAEnabled:        false,
		ActivePseudonymID: activePseudonymID,
		DisplayName:       "TestUser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		return "", fmt.Errorf("failed to generate test JWT token: %w", err)
	}

	return tokenString, nil
}

// MustGenerateTestJWTToken generates a test JWT token and panics if it fails
// This is a convenience function for tests where token generation failure should cause the test to fail
func MustGenerateTestJWTToken(userID int64, activePseudonymID string) string {
	token, err := GenerateTestJWTToken(userID, activePseudonymID)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test JWT token: %v", err))
	}
	return token
}

// CreateTestRoleKey creates a test role key
func CreateTestRoleKey(keyID string, roleName, scope string, capabilities []string) *models.RoleKey {
	// Create a UUID from the string keyID for testing
	uuid, _ := uuid.FromString(keyID)
	return &models.RoleKey{
		KeyID:     uuid,
		RoleName:  roleName,
		Scope:     scope,
		KeyData:   []byte("test_key_data"),
		ExpiresAt: time.Now().AddDate(1, 0, 0),
		CreatedBy: 1,
	}
}
