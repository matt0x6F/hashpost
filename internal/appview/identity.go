package appview

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/patrickmn/go-cache"
)

// IdentityResolver handles DID resolution and caching for AppView
type IdentityResolver struct {
	directory identity.Directory
	logger    *slog.Logger
	cache     *cache.Cache
	mu        sync.RWMutex
}

// UserInfo represents resolved user information
type UserInfo struct {
	DID         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// NewIdentityResolver creates a new identity resolver with caching
func NewIdentityResolver(logger *slog.Logger) *IdentityResolver {
	// Use the same identity directory as PDS for consistency
	directory := identity.NewMockDirectory()

	// Add the same test identities as PDS
	testUser := identity.Identity{
		DID:    syntax.DID("did:plc:hashpost-binding-test"),
		Handle: syntax.Handle("testuser.hashpost.local"),
	}
	adminUser := identity.Identity{
		DID:    syntax.DID("did:plc:hashpost-admin-test"),
		Handle: syntax.Handle("admin.hashpost.local"),
	}

	directory.Insert(testUser)
	directory.Insert(adminUser)

	// Create cache with TTL: 1 hour for successful resolutions, 5 minutes for failures
	c := cache.New(1*time.Hour, 5*time.Minute)

	return &IdentityResolver{
		directory: &directory,
		logger:    logger,
		cache:     c,
	}
}

// ResolveHandleFromDID resolves a DID to a handle
func (ir *IdentityResolver) ResolveHandleFromDID(ctx context.Context, did string) (string, error) {
	// Check cache first
	if cached, found := ir.cache.Get("handle:" + did); found {
		ir.logger.Debug("DID resolution cache hit", "did", did)
		return cached.(string), nil
	}

	ir.logger.Debug("Resolving DID to handle", "did", did)

	// Resolve DID using identity directory
	identity, err := ir.directory.LookupDID(ctx, syntax.DID(did))
	if err != nil {
		ir.logger.Error("Failed to resolve DID", "error", err, "did", did)
		// Cache negative result for shorter time
		ir.cache.Set("handle:"+did, "", 5*time.Minute)
		return "", fmt.Errorf("DID resolution failed: %w", err)
	}

	handle := identity.Handle.String()

	// Cache successful resolution
	ir.cache.Set("handle:"+did, handle, 1*time.Hour)

	ir.logger.Info("DID resolved to handle", "did", did, "handle", handle)
	return handle, nil
}

// GetUserInfoFromDID resolves a DID to complete user information
func (ir *IdentityResolver) GetUserInfoFromDID(ctx context.Context, did string) (*UserInfo, error) {
	// Check cache first
	if cached, found := ir.cache.Get("userinfo:" + did); found {
		ir.logger.Debug("User info cache hit", "did", did)
		return cached.(*UserInfo), nil
	}

	ir.logger.Debug("Resolving user info from DID", "did", did)

	// Resolve DID using identity directory
	identity, err := ir.directory.LookupDID(ctx, syntax.DID(did))
	if err != nil {
		ir.logger.Error("Failed to resolve user info", "error", err, "did", did)
		// Cache negative result for shorter time
		ir.cache.Set("userinfo:"+did, nil, 5*time.Minute)
		return nil, fmt.Errorf("DID resolution failed: %w", err)
	}

	userInfo := &UserInfo{
		DID:    did,
		Handle: identity.Handle.String(),
		// In a real implementation, we'd extract display name and avatar from the DID document
		DisplayName: identity.Handle.String(), // Use handle as display name for now
		AvatarURL:   "",                       // No avatar URL in mock directory
	}

	// Cache successful resolution
	ir.cache.Set("userinfo:"+did, userInfo, 1*time.Hour)

	ir.logger.Info("User info resolved", "did", did, "handle", userInfo.Handle)
	return userInfo, nil
}

// AddUserToMockDirectory adds a user to the mock identity directory (development only)
func (ir *IdentityResolver) AddUserToMockDirectory(ctx context.Context, did, handle string) error {
	// Only works in development mode with mock directory
	if mockDir, ok := ir.directory.(*identity.MockDirectory); ok {
		userIdentity := identity.Identity{
			DID:    syntax.DID(did),
			Handle: syntax.Handle(handle),
		}
		mockDir.Insert(userIdentity)
		ir.logger.Debug("Added user to mock identity directory", "did", did, "handle", handle)
		return nil
	}
	return fmt.Errorf("not a mock directory")
}

// GetCacheStats returns cache statistics for monitoring
func (ir *IdentityResolver) GetCacheStats() map[string]interface{} {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	stats := ir.cache.ItemCount()
	return map[string]interface{}{
		"cache_items": stats,
		"cache_size":  stats,
	}
}

// ClearCache clears the identity resolution cache
func (ir *IdentityResolver) ClearCache() {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.cache.Flush()
	ir.logger.Info("Identity resolution cache cleared")
}
