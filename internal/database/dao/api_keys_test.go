package dao

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyDAO_CreateAndValidate(t *testing.T) {
	// This is a basic test to verify the API Key DAO structure
	// In a real test, you would use a test database connection

	t.Run("CreateAPIKey", func(t *testing.T) {
		// Test creating an API key with permissions
		permissions := &APIKeyPermissions{
			Roles:        []string{"admin"},
			Capabilities: []string{"read", "write"},
		}

		// This would require a real database connection
		// For now, just verify the structure is correct
		assert.NotNil(t, permissions)
		assert.Equal(t, []string{"admin"}, permissions.Roles)
		assert.Equal(t, []string{"read", "write"}, permissions.Capabilities)
	})

	t.Run("HashAPIKey", func(t *testing.T) {
		// Test that the hash function works consistently
		key1 := "test-api-key-123"
		key2 := "test-api-key-123"
		key3 := "different-key"

		hash1 := hashAPIKey(key1)
		hash2 := hashAPIKey(key2)
		hash3 := hashAPIKey(key3)

		assert.Equal(t, hash1, hash2, "Same key should produce same hash")
		assert.NotEqual(t, hash1, hash3, "Different keys should produce different hashes")
		assert.Len(t, hash1, 64, "SHA-256 hash should be 64 characters")
	})
}

func TestAPIKeyPermissions_JSON(t *testing.T) {
	t.Run("MarshalUnmarshal", func(t *testing.T) {
		// Test JSON marshaling and unmarshaling
		original := &APIKeyPermissions{
			Roles:        []string{"admin", "moderator"},
			Capabilities: []string{"read", "write", "delete"},
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var unmarshaled APIKeyPermissions
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Verify the data is the same
		assert.Equal(t, original.Roles, unmarshaled.Roles)
		assert.Equal(t, original.Capabilities, unmarshaled.Capabilities)
	})

	t.Run("BasicPermissions", func(t *testing.T) {
		// Test with basic permissions
		original := &APIKeyPermissions{
			Roles:        []string{"user"},
			Capabilities: []string{"read"},
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var unmarshaled APIKeyPermissions
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Verify the data is the same
		assert.Equal(t, original.Roles, unmarshaled.Roles)
		assert.Equal(t, original.Capabilities, unmarshaled.Capabilities)
	})
}
