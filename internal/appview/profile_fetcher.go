package appview

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// ProfileData represents user profile information fetched from a PDS
type ProfileData struct {
	DID         string
	Handle      string
	DisplayName *string
	Bio         *string
	AvatarURL   *string
	PDSSource   string
}

// ProfileFetcher handles fetching user profiles from external PDS servers
type ProfileFetcher struct {
	queries generated.Querier
	logger  *slog.Logger
	client  *http.Client
}

// NewProfileFetcher creates a new profile fetcher
func NewProfileFetcher(queries generated.Querier, logger *slog.Logger) *ProfileFetcher {
	return &ProfileFetcher{
		queries: queries,
		logger:  logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchProfileFromPDS fetches user profile data from any PDS server (including HashPost's own)
func (pf *ProfileFetcher) FetchProfileFromPDS(ctx context.Context, did string) (*ProfileData, error) {
	pf.logger.Debug("Fetching profile from PDS", "did", did)

	// Step 1: Resolve DID to get PDS endpoint
	pdsEndpoint, err := pf.resolvePDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS endpoint for DID %s: %w", did, err)
	}

	// Step 2: Fetch profile record from PDS
	profileData, err := pf.fetchProfileRecord(ctx, pdsEndpoint, did)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile record: %w", err)
	}

	// Step 3: Cache the profile data in our database
	err = pf.cacheProfileData(ctx, profileData)
	if err != nil {
		pf.logger.Warn("Failed to cache profile data", "error", err, "did", did)
		// Don't return error - we still have the profile data
	}

	return profileData, nil
}

// resolvePDSEndpoint resolves a DID to its PDS endpoint URL
func (pf *ProfileFetcher) resolvePDSEndpoint(ctx context.Context, did string) (string, error) {
	// For now, implement a simple resolution
	// In production, this would:
	// 1. Fetch DID document from PLC directory
	// 2. Extract PDS endpoint from DID document
	// 3. Return the endpoint URL

	// Check if this is a local HashPost user
	if strings.Contains(did, "hashpost") {
		return "http://hashpost-pds:8080", nil
	}

	// For external DIDs, we'd need to implement proper DID resolution
	// For now, return a mock external PDS
	return "https://bsky.network", nil
}

// fetchProfileRecord fetches the app.bsky.actor.profile record from a PDS
func (pf *ProfileFetcher) fetchProfileRecord(ctx context.Context, pdsEndpoint, did string) (*ProfileData, error) {
	// Construct the AT Protocol repo.getRecord request
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord", pdsEndpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("repo", did)
	q.Add("collection", "app.bsky.actor.profile")
	q.Add("rkey", "self")
	req.URL.RawQuery = q.Encode()

	// Make the request
	resp, err := pf.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PDS returned status %d", resp.StatusCode)
	}

	// Parse the response
	var recordResponse struct {
		URI   string `json:"uri"`
		Value struct {
			DisplayName *string `json:"displayName"`
			Description *string `json:"description"`
			Avatar      *string `json:"avatar"`
		} `json:"value"`
	}

	err = json.NewDecoder(resp.Body).Decode(&recordResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract handle from DID (simplified - in production, resolve handle from DID)
	handle := pf.extractHandleFromDID(did)

	return &ProfileData{
		DID:         did,
		Handle:      handle,
		DisplayName: recordResponse.Value.DisplayName,
		Bio:         recordResponse.Value.Description,
		AvatarURL:   recordResponse.Value.Avatar,
		PDSSource:   pdsEndpoint,
	}, nil
}

// extractHandleFromDID extracts a handle from a DID (simplified implementation)
func (pf *ProfileFetcher) extractHandleFromDID(did string) string {
	// In production, this would:
	// 1. Check if we have the handle cached in our database
	// 2. If not, resolve the handle from the DID document
	// 3. Return the handle

	// For now, return a mock handle
	if did == "did:plc:hashpost-local" {
		return "user.hashpost.local"
	}
	return "user.bsky.social"
}

// cacheProfileData stores the profile data in our database
func (pf *ProfileFetcher) cacheProfileData(ctx context.Context, profile *ProfileData) error {
	_, err := pf.queries.UpdateUserProfile(ctx, &generated.UpdateUserProfileParams{
		Did:         profile.DID,
		DisplayName: profile.DisplayName,
		Bio:         profile.Bio,
		AvatarUrl:   profile.AvatarURL,
	})

	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	pf.logger.Info("Profile data cached", "did", profile.DID, "display_name", profile.DisplayName)
	return nil
}

// GetCachedProfile retrieves cached profile data for a user
func (pf *ProfileFetcher) GetCachedProfile(ctx context.Context, did string) (*ProfileData, error) {
	user, err := pf.queries.GetUserByDID(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &ProfileData{
		DID:         user.Did,
		Handle:      user.Handle,
		DisplayName: user.DisplayName,
		Bio:         user.Bio,
		AvatarURL:   user.AvatarUrl,
		PDSSource:   pf.getStringValue(user.PdsSource),
	}, nil
}

// Helper function to safely get string value from pointer
func (pf *ProfileFetcher) getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
