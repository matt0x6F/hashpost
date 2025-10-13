package testutil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertAtprotoDID validates that a string is a valid atproto DID
func AssertAtprotoDID(t *testing.T, did string, msgAndArgs ...interface{}) {
	t.Helper()

	// Check basic DID format
	if !strings.HasPrefix(did, "did:") {
		assert.Fail(t, fmt.Sprintf("Expected DID to start with 'did:', got: %s", did), msgAndArgs...)
		return
	}

	// Try to parse as DID using atproto syntax
	parsedDID, err := syntax.ParseDID(did)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected valid DID format, got error: %v", err), msgAndArgs...)
		return
	}

	// Ensure it's not empty
	if parsedDID.String() == "" {
		assert.Fail(t, "Expected non-empty DID", msgAndArgs...)
	}
}

// AssertAtprotoHandle validates that a string is a valid atproto handle
func AssertAtprotoHandle(t *testing.T, handle string, msgAndArgs ...interface{}) {
	t.Helper()

	// Check basic handle format (should contain a dot)
	if !strings.Contains(handle, ".") {
		assert.Fail(t, fmt.Sprintf("Expected handle to contain '.', got: %s", handle), msgAndArgs...)
		return
	}

	// Try to parse as handle using atproto syntax
	parsedHandle, err := syntax.ParseHandle(handle)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected valid handle format, got error: %v", err), msgAndArgs...)
		return
	}

	// Ensure it's not empty
	if parsedHandle.String() == "" {
		assert.Fail(t, "Expected non-empty handle", msgAndArgs...)
	}
}

// AssertAtprotoURI validates that a string is a valid atproto URI
func AssertAtprotoURI(t *testing.T, uri string, msgAndArgs ...interface{}) {
	t.Helper()

	// Check basic URI format
	if !strings.HasPrefix(uri, "at://") {
		assert.Fail(t, fmt.Sprintf("Expected URI to start with 'at://', got: %s", uri), msgAndArgs...)
		return
	}

	// Check that it has the expected structure: at://did/collection/rkey
	parts := strings.Split(strings.TrimPrefix(uri, "at://"), "/")
	if len(parts) < 3 {
		assert.Fail(t, fmt.Sprintf("Expected URI to have at least 3 parts (did/collection/rkey), got: %s", uri), msgAndArgs...)
		return
	}

	// Validate the DID part
	AssertAtprotoDID(t, parts[0], msgAndArgs...)
}

// AssertJWTStructure validates that a JWT token has the expected structure
func AssertJWTStructure(t *testing.T, token string, expectedClaims []string, msgAndArgs ...interface{}) {
	t.Helper()

	// Check basic JWT format (should have 3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		assert.Fail(t, fmt.Sprintf("Expected JWT to have 3 parts, got %d", len(parts)), msgAndArgs...)
		return
	}

	// Decode the claims (middle part)
	claimsBytes, err := decodeBase64URL(parts[1])
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Failed to decode JWT claims: %v", err), msgAndArgs...)
		return
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		assert.Fail(t, fmt.Sprintf("Failed to unmarshal JWT claims: %v", err), msgAndArgs...)
		return
	}

	// Check for expected claims
	for _, expectedClaim := range expectedClaims {
		if _, exists := claims[expectedClaim]; !exists {
			assert.Fail(t, fmt.Sprintf("Expected JWT to contain claim '%s'", expectedClaim), msgAndArgs...)
		}
	}
}

// AssertCIDFormat validates that a string is a valid CID format
func AssertCIDFormat(t *testing.T, cid string, msgAndArgs ...interface{}) {
	t.Helper()

	// Check basic CID format (should start with 'b' for base58 or 'z' for base32)
	if !strings.HasPrefix(cid, "b") && !strings.HasPrefix(cid, "z") {
		assert.Fail(t, fmt.Sprintf("Expected CID to start with 'b' or 'z', got: %s", cid), msgAndArgs...)
		return
	}

	// Check minimum length (CIDs are typically much longer)
	if len(cid) < 10 {
		assert.Fail(t, fmt.Sprintf("Expected CID to be longer than 10 characters, got: %s", cid), msgAndArgs...)
	}
}

// AssertTimestampFormat validates that a string is a valid timestamp format
func AssertTimestampFormat(t *testing.T, timestamp string, msgAndArgs ...interface{}) {
	t.Helper()

	// Try to parse as RFC3339 timestamp
	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected valid RFC3339 timestamp, got error: %v", err), msgAndArgs...)
	}
}

// AssertSessionValid validates that a session object has required fields
func AssertSessionValid(t *testing.T, session interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	// Convert to map for easier inspection
	sessionMap, ok := session.(map[string]interface{})
	if !ok {
		// Try to marshal/unmarshal to convert
		jsonBytes, err := json.Marshal(session)
		require.NoError(t, err, "Failed to marshal session")

		err = json.Unmarshal(jsonBytes, &sessionMap)
		require.NoError(t, err, "Failed to unmarshal session")
	}

	// Check required fields
	requiredFields := []string{"id", "did", "handle", "created_at", "expires_at"}
	for _, field := range requiredFields {
		if _, exists := sessionMap[field]; !exists {
			assert.Fail(t, fmt.Sprintf("Expected session to contain field '%s'", field), msgAndArgs...)
		}
	}

	// Validate DID format
	if did, exists := sessionMap["did"].(string); exists {
		AssertAtprotoDID(t, did, msgAndArgs...)
	}

	// Validate handle format
	if handle, exists := sessionMap["handle"].(string); exists {
		AssertAtprotoHandle(t, handle, msgAndArgs...)
	}

	// Validate timestamp formats
	if createdAt, exists := sessionMap["created_at"].(string); exists {
		AssertTimestampFormat(t, createdAt, msgAndArgs...)
	}

	if expiresAt, exists := sessionMap["expires_at"].(string); exists {
		AssertTimestampFormat(t, expiresAt, msgAndArgs...)
	}
}

// AssertEventStructure validates that an event has the expected atproto event structure
func AssertEventStructure(t *testing.T, event interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	// Convert to map for easier inspection
	eventMap, ok := event.(map[string]interface{})
	if !ok {
		// Try to marshal/unmarshal to convert
		jsonBytes, err := json.Marshal(event)
		require.NoError(t, err, "Failed to marshal event")

		err = json.Unmarshal(jsonBytes, &eventMap)
		require.NoError(t, err, "Failed to unmarshal event")
	}

	// Check required fields
	requiredFields := []string{"type", "repo", "timestamp"}
	for _, field := range requiredFields {
		if _, exists := eventMap[field]; !exists {
			assert.Fail(t, fmt.Sprintf("Expected event to contain field '%s'", field), msgAndArgs...)
		}
	}

	// Validate event type
	if eventType, exists := eventMap["type"].(string); exists {
		validTypes := []string{"record.created", "record.updated", "record.deleted", "identity.resolved", "session.created"}
		found := false
		for _, validType := range validTypes {
			if eventType == validType {
				found = true
				break
			}
		}
		if !found {
			assert.Fail(t, fmt.Sprintf("Expected event type to be one of %v, got: %s", validTypes, eventType), msgAndArgs...)
		}
	}

	// Validate repo (should be a DID)
	if repo, exists := eventMap["repo"].(string); exists {
		AssertAtprotoDID(t, repo, msgAndArgs...)
	}

	// Validate timestamp
	if timestamp, exists := eventMap["timestamp"].(string); exists {
		AssertTimestampFormat(t, timestamp, msgAndArgs...)
	}
}

// AssertRecordStructure validates that a record has the expected atproto record structure
func AssertRecordStructure(t *testing.T, record interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	// Convert to map for easier inspection
	recordMap, ok := record.(map[string]interface{})
	if !ok {
		// Try to marshal/unmarshal to convert
		jsonBytes, err := json.Marshal(record)
		require.NoError(t, err, "Failed to marshal record")

		err = json.Unmarshal(jsonBytes, &recordMap)
		require.NoError(t, err, "Failed to unmarshal record")
	}

	// Check for $type field (required for atproto records)
	if recordType, exists := recordMap["$type"].(string); exists {
		// Should be a valid atproto record type
		if !strings.Contains(recordType, ".") {
			assert.Fail(t, fmt.Sprintf("Expected record type to contain '.', got: %s", recordType), msgAndArgs...)
		}
	} else {
		assert.Fail(t, "Expected record to contain '$type' field", msgAndArgs...)
	}
}

// AssertErrorContains checks that an error contains a specific substring
func AssertErrorContains(t *testing.T, err error, substring string, msgAndArgs ...interface{}) {
	t.Helper()

	if err == nil {
		assert.Fail(t, "Expected error but got nil", msgAndArgs...)
		return
	}

	if !strings.Contains(err.Error(), substring) {
		assert.Fail(t, fmt.Sprintf("Expected error to contain '%s', got: %s", substring, err.Error()), msgAndArgs...)
	}
}

// AssertNoError checks that there is no error
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if err != nil {
		assert.Fail(t, fmt.Sprintf("Expected no error, got: %v", err), msgAndArgs...)
	}
}

// AssertDurationWithin checks that a duration is within expected bounds
func AssertDurationWithin(t *testing.T, actual, min, max time.Duration, msgAndArgs ...interface{}) {
	t.Helper()

	if actual < min || actual > max {
		assert.Fail(t, fmt.Sprintf("Expected duration to be between %v and %v, got: %v", min, max, actual), msgAndArgs...)
	}
}

// decodeBase64URL decodes a base64url encoded string
func decodeBase64URL(s string) ([]byte, error) {
	// Add padding if needed
	for len(s)%4 != 0 {
		s += "="
	}

	// Replace URL-safe characters
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	// Decode using base64
	return base64.StdEncoding.DecodeString(s)
}
