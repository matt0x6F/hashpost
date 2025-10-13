package pds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtprotoValidator_ValidateAtIdentifier(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		identifier  string
		expectError bool
	}{
		{
			name:        "valid_did",
			identifier:  "did:plc:abc123def456",
			expectError: false,
		},
		{
			name:        "valid_handle",
			identifier:  "alice.hashpost.local",
			expectError: false,
		},
		{
			name:        "empty_identifier",
			identifier:  "",
			expectError: true,
		},
		{
			name:        "invalid_did_format",
			identifier:  "did:",
			expectError: true,
		},
		{
			name:        "invalid_handle_format",
			identifier:  "alice",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAtIdentifier(tt.identifier)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateDID(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		did         string
		expectError bool
	}{
		{
			name:        "valid_plc_did",
			did:         "did:plc:abc123def456",
			expectError: false,
		},
		{
			name:        "valid_key_did",
			did:         "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			expectError: false,
		},
		{
			name:        "empty_did",
			did:         "",
			expectError: true,
		},
		{
			name:        "missing_did_prefix",
			did:         "plc:abc123",
			expectError: true,
		},
		{
			name:        "invalid_method",
			did:         "did:123:abc",
			expectError: true,
		},
		{
			name:        "empty_method",
			did:         "did::abc",
			expectError: true,
		},
		{
			name:        "invalid_identifier",
			did:         "did:plc:abc@123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDID(tt.did)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateHandle(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		handle      string
		expectError bool
	}{
		{
			name:        "valid_handle",
			handle:      "alice.hashpost.local",
			expectError: false,
		},
		{
			name:        "valid_handle_with_hyphens",
			handle:      "alice-user.hashpost.local",
			expectError: false,
		},
		{
			name:        "valid_handle_with_underscores",
			handle:      "alice_user.hashpost.local",
			expectError: false,
		},
		{
			name:        "valid_handle_with_numbers",
			handle:      "alice123.hashpost.local",
			expectError: false,
		},
		{
			name:        "empty_handle",
			handle:      "",
			expectError: true,
		},
		{
			name:        "handle_without_domain",
			handle:      "alice",
			expectError: true,
		},
		{
			name:        "handle_with_empty_local_part",
			handle:      ".hashpost.local",
			expectError: true,
		},
		{
			name:        "handle_with_empty_domain",
			handle:      "alice.",
			expectError: true,
		},
		{
			name:        "handle_with_invalid_characters",
			handle:      "alice@user.hashpost.local",
			expectError: true,
		},
		{
			name:        "handle_with_consecutive_dots",
			handle:      "alice..user.hashpost.local",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateHandle(tt.handle)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateAtURI(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		uri         string
		expectError bool
	}{
		{
			name:        "valid_at_uri_with_did",
			uri:         "at://did:plc:abc123def456/app.bsky.feed.post/123",
			expectError: false,
		},
		{
			name:        "valid_at_uri_with_handle",
			uri:         "at://alice.hashpost.local/app.bsky.feed.post/123",
			expectError: false,
		},
		{
			name:        "empty_uri",
			uri:         "",
			expectError: true,
		},
		{
			name:        "missing_at_protocol",
			uri:         "https://alice.hashpost.local/app.bsky.feed.post/123",
			expectError: true,
		},
		{
			name:        "missing_authority",
			uri:         "at:///app.bsky.feed.post/123",
			expectError: true,
		},
		{
			name:        "missing_path",
			uri:         "at://alice.hashpost.local",
			expectError: true,
		},
		{
			name:        "missing_collection",
			uri:         "at://alice.hashpost.local/123",
			expectError: true,
		},
		{
			name:        "missing_rkey",
			uri:         "at://alice.hashpost.local/app.bsky.feed.post",
			expectError: true,
		},
		{
			name:        "invalid_collection",
			uri:         "at://alice.hashpost.local/invalid.collection/123",
			expectError: true,
		},
		{
			name:        "invalid_rkey",
			uri:         "at://alice.hashpost.local/app.bsky.feed.post/invalid@key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAtURI(tt.uri)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateCollection(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		collection  string
		expectError bool
	}{
		{
			name:        "valid_collection",
			collection:  "app.bsky.feed.post",
			expectError: false,
		},
		{
			name:        "valid_simple_collection",
			collection:  "app.bsky",
			expectError: false,
		},
		{
			name:        "empty_collection",
			collection:  "",
			expectError: true,
		},
		{
			name:        "collection_starting_with_number",
			collection:  "123app.bsky",
			expectError: true,
		},
		{
			name:        "collection_with_invalid_characters",
			collection:  "app@bsky.feed",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCollection(tt.collection)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateRkey(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		rkey        string
		expectError bool
	}{
		{
			name:        "valid_rkey",
			rkey:        "123",
			expectError: false,
		},
		{
			name:        "valid_rkey_with_hyphens",
			rkey:        "abc-123-def",
			expectError: false,
		},
		{
			name:        "valid_rkey_with_underscores",
			rkey:        "abc_123_def",
			expectError: false,
		},
		{
			name:        "valid_rkey_with_dots",
			rkey:        "abc.123.def",
			expectError: false,
		},
		{
			name:        "empty_rkey",
			rkey:        "",
			expectError: true,
		},
		{
			name:        "rkey_with_invalid_characters",
			rkey:        "abc@123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRkey(tt.rkey)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateCID(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		cid         string
		expectError bool
	}{
		{
			name:        "valid_base32_cid",
			cid:         "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			expectError: false,
		},
		{
			name:        "valid_base58_cid",
			cid:         "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
			expectError: false,
		},
		{
			name:        "empty_cid",
			cid:         "",
			expectError: true,
		},
		{
			name:        "invalid_cid_prefix",
			cid:         "invalid123",
			expectError: true,
		},
		{
			name:        "cid_with_invalid_characters",
			cid:         "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi@",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCID(tt.cid)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateDateTime(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		dateTime    string
		expectError bool
	}{
		{
			name:        "valid_rfc3339_datetime",
			dateTime:    "2023-12-01T10:30:00Z",
			expectError: false,
		},
		{
			name:        "valid_rfc3339_datetime_with_timezone",
			dateTime:    "2023-12-01T10:30:00+01:00",
			expectError: false,
		},
		{
			name:        "empty_datetime",
			dateTime:    "",
			expectError: true,
		},
		{
			name:        "invalid_datetime_format",
			dateTime:    "2023-12-01 10:30:00",
			expectError: true,
		},
		{
			name:        "invalid_datetime",
			dateTime:    "not-a-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDateTime(tt.dateTime)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateSessionToken(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid_jwt_token",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectError: false,
		},
		{
			name:        "empty_token",
			token:       "",
			expectError: true,
		},
		{
			name:        "invalid_jwt_format",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "jwt_with_missing_parts",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			expectError: true,
		},
		{
			name:        "jwt_with_empty_parts",
			token:       "..",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSessionToken(tt.token)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidatePassword(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid_password",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "valid_long_password",
			password:    "this-is-a-very-long-password-with-many-characters",
			expectError: false,
		},
		{
			name:        "empty_password",
			password:    "",
			expectError: true,
		},
		{
			name:        "short_password",
			password:    "1234567",
			expectError: true,
		},
		{
			name:        "too_long_password",
			password:    "this-password-is-way-too-long-and-exceeds-the-maximum-length-of-128-characters-which-should-cause-a-validation-error-and-is-actually-longer-than-128-characters-now",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePassword(tt.password)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateEmail(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "valid_email",
			email:       "alice@hashpost.local",
			expectError: false,
		},
		{
			name:        "valid_email_with_subdomain",
			email:       "alice@mail.hashpost.local",
			expectError: false,
		},
		{
			name:        "valid_email_with_plus",
			email:       "alice+test@hashpost.local",
			expectError: false,
		},
		{
			name:        "empty_email",
			email:       "",
			expectError: true,
		},
		{
			name:        "invalid_email_format",
			email:       "alice",
			expectError: true,
		},
		{
			name:        "email_without_domain",
			email:       "alice@",
			expectError: true,
		},
		{
			name:        "email_without_local_part",
			email:       "@hashpost.local",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAtprotoValidator_ValidateInviteCode(t *testing.T) {
	validator := NewAtprotoValidator()

	tests := []struct {
		name        string
		code        string
		expectError bool
	}{
		{
			name:        "valid_invite_code",
			code:        "ABC123DEF456",
			expectError: false,
		},
		{
			name:        "valid_invite_code_lowercase",
			code:        "abc123def456",
			expectError: false,
		},
		{
			name:        "valid_invite_code_mixed_case",
			code:        "Abc123Def456",
			expectError: false,
		},
		{
			name:        "empty_invite_code",
			code:        "",
			expectError: true,
		},
		{
			name:        "short_invite_code",
			code:        "ABC123",
			expectError: true,
		},
		{
			name:        "invite_code_with_invalid_characters",
			code:        "ABC123-DEF456",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateInviteCode(tt.code)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
