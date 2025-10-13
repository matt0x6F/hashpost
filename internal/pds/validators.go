package pds

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// AtprotoValidator provides validation for atproto string formats and protocols
type AtprotoValidator struct{}

// NewAtprotoValidator creates a new atproto validator
func NewAtprotoValidator() *AtprotoValidator {
	return &AtprotoValidator{}
}

// ValidateAtIdentifier validates an at-identifier (handle or DID)
func (v *AtprotoValidator) ValidateAtIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	// Check if it's a DID
	if strings.HasPrefix(identifier, "did:") {
		return v.ValidateDID(identifier)
	}

	// Check if it's a handle
	return v.ValidateHandle(identifier)
}

// ValidateDID validates a DID (Decentralized Identifier)
func (v *AtprotoValidator) ValidateDID(did string) error {
	if !strings.HasPrefix(did, "did:") {
		return fmt.Errorf("DID must start with 'did:'")
	}

	// Basic DID format validation
	// Format: did:method:identifier
	parts := strings.Split(did, ":")
	if len(parts) < 3 {
		return fmt.Errorf("invalid DID format: %s", did)
	}

	method := parts[1]
	if method == "" {
		return fmt.Errorf("DID method cannot be empty")
	}

	// Validate method characters (must start with letter, then alphanumeric and hyphens)
	methodRegex := regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	if !methodRegex.MatchString(method) {
		return fmt.Errorf("invalid DID method: %s", method)
	}

	// Validate identifier characters
	identifier := strings.Join(parts[2:], ":")
	identifierRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !identifierRegex.MatchString(identifier) {
		return fmt.Errorf("invalid DID identifier: %s", identifier)
	}

	return nil
}

// ValidateHandle validates a handle
func (v *AtprotoValidator) ValidateHandle(handle string) error {
	if handle == "" {
		return fmt.Errorf("handle cannot be empty")
	}

	// Handle format: local-part.domain
	parts := strings.Split(handle, ".")
	if len(parts) < 2 {
		return fmt.Errorf("handle must contain at least one dot: %s", handle)
	}

	// Validate local part (first part before the first dot)
	localPart := parts[0]
	if localPart == "" {
		return fmt.Errorf("handle local part cannot be empty")
	}

	// Check for consecutive dots in the full handle
	if strings.Contains(handle, "..") {
		return fmt.Errorf("handle cannot contain consecutive dots")
	}

	// Local part can contain letters, numbers, hyphens, and underscores
	localPartRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !localPartRegex.MatchString(localPart) {
		return fmt.Errorf("invalid handle local part: %s", localPart)
	}

	// Validate domain part (after the last dot)
	domain := parts[len(parts)-1]
	if domain == "" {
		return fmt.Errorf("handle domain cannot be empty")
	}

	// Domain can contain letters, numbers, hyphens, and dots
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid handle domain: %s", domain)
	}

	// Check for consecutive dots
	if strings.Contains(domain, "..") {
		return fmt.Errorf("handle domain cannot contain consecutive dots")
	}

	return nil
}

// ValidateAtURI validates an at-uri (record URI)
func (v *AtprotoValidator) ValidateAtURI(uri string) error {
	if uri == "" {
		return fmt.Errorf("URI cannot be empty")
	}

	if !strings.HasPrefix(uri, "at://") {
		return fmt.Errorf("at-uri must start with 'at://'")
	}

	// Parse the at-uri manually since url.Parse doesn't handle at:// properly
	// Format: at://authority/path
	uriWithoutProtocol := uri[5:] // Remove "at://"

	// Find the first slash to separate authority from path
	slashIndex := strings.Index(uriWithoutProtocol, "/")
	if slashIndex == -1 {
		return fmt.Errorf("URI must contain a path")
	}

	authority := uriWithoutProtocol[:slashIndex]
	path := uriWithoutProtocol[slashIndex:]

	// Validate the authority (DID or handle)
	if authority == "" {
		return fmt.Errorf("URI authority cannot be empty")
	}

	if err := v.ValidateAtIdentifier(authority); err != nil {
		return fmt.Errorf("invalid URI authority: %w", err)
	}

	// Validate the path (collection/rkey)
	if path == "" {
		return fmt.Errorf("URI path cannot be empty")
	}

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Split into collection and rkey
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		return fmt.Errorf("URI path must contain collection and rkey")
	}

	collection := pathParts[0]
	rkey := strings.Join(pathParts[1:], "/")

	// Validate collection format
	if err := v.ValidateCollection(collection); err != nil {
		return fmt.Errorf("invalid collection: %w", err)
	}

	// Validate rkey format
	if err := v.ValidateRkey(rkey); err != nil {
		return fmt.Errorf("invalid rkey: %w", err)
	}

	return nil
}

// ValidateCollection validates a collection name
func (v *AtprotoValidator) ValidateCollection(collection string) error {
	if collection == "" {
		return fmt.Errorf("collection cannot be empty")
	}

	// Collection format: domain.name (e.g., app.bsky.feed.post)
	// Must start with a letter, then alphanumeric, with dots separating parts
	collectionRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*(\.[a-zA-Z][a-zA-Z0-9]*)*$`)
	if !collectionRegex.MatchString(collection) {
		return fmt.Errorf("invalid collection format: %s", collection)
	}

	// Additional validation: only allow known valid collection names
	validCollections := []string{
		"app.bsky",
		"app.bsky.feed.post",
		"app.bsky.feed.like",
		"app.bsky.feed.repost",
		"app.bsky.graph.follow",
		"app.bsky.actor.profile",
	}

	isValid := false
	for _, valid := range validCollections {
		if collection == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid collection name: %s", collection)
	}

	return nil
}

// ValidateRkey validates a record key
func (v *AtprotoValidator) ValidateRkey(rkey string) error {
	if rkey == "" {
		return fmt.Errorf("rkey cannot be empty")
	}

	// Rkey can contain letters, numbers, hyphens, underscores, and dots
	rkeyRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !rkeyRegex.MatchString(rkey) {
		return fmt.Errorf("invalid rkey format: %s", rkey)
	}

	return nil
}

// ValidateCID validates a CID (Content Identifier)
func (v *AtprotoValidator) ValidateCID(cid string) error {
	if cid == "" {
		return fmt.Errorf("CID cannot be empty")
	}

	// Basic CID format validation
	// CIDs typically start with 'bafy' for base32 or 'Qm' for base58
	if !strings.HasPrefix(cid, "bafy") && !strings.HasPrefix(cid, "Qm") {
		return fmt.Errorf("invalid CID format: %s", cid)
	}

	// CID should be alphanumeric
	cidRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !cidRegex.MatchString(cid) {
		return fmt.Errorf("CID contains invalid characters: %s", cid)
	}

	return nil
}

// ValidateDateTime validates an ISO 8601 datetime string
func (v *AtprotoValidator) ValidateDateTime(dateTime string) error {
	if dateTime == "" {
		return fmt.Errorf("datetime cannot be empty")
	}

	// Try to parse as RFC3339 (ISO 8601)
	_, err := time.Parse(time.RFC3339, dateTime)
	if err != nil {
		return fmt.Errorf("invalid datetime format: %w", err)
	}

	return nil
}

// ValidateSessionToken validates a session token format
func (v *AtprotoValidator) ValidateSessionToken(token string) error {
	if token == "" {
		return fmt.Errorf("session token cannot be empty")
	}

	// Basic JWT format validation (3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Each part should be base64 encoded
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("JWT part %d cannot be empty", i+1)
		}
	}

	return nil
}

// ValidatePassword validates a password
func (v *AtprotoValidator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be no more than 128 characters long")
	}

	return nil
}

// ValidateEmail validates an email address
func (v *AtprotoValidator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Basic email format validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	return nil
}

// ValidateInviteCode validates an invite code
func (v *AtprotoValidator) ValidateInviteCode(code string) error {
	if code == "" {
		return fmt.Errorf("invite code cannot be empty")
	}

	// Invite code should be alphanumeric
	codeRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !codeRegex.MatchString(code) {
		return fmt.Errorf("invalid invite code format: %s", code)
	}

	if len(code) < 8 {
		return fmt.Errorf("invite code must be at least 8 characters long")
	}

	return nil
}
