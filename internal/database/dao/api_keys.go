package dao

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
)

// APIKeyDAO provides data access operations for API keys
type APIKeyDAO struct {
	db bob.Executor
}

// NewAPIKeyDAO creates a new APIKeyDAO
func NewAPIKeyDAO(db bob.Executor) *APIKeyDAO {
	return &APIKeyDAO{
		db: db,
	}
}

// APIKeyPermissions represents the permissions structure for API keys
type APIKeyPermissions struct {
	Roles        []string `json:"roles"`
	Capabilities []string `json:"capabilities"`
}

// CreateAPIKey creates a new API key
func (dao *APIKeyDAO) CreateAPIKey(ctx context.Context, keyName string, rawKey string, pseudonymID string, permissions *APIKeyPermissions, expiresAt *time.Time) (*models.APIKey, error) {
	log.Debug().
		Str("key_name", keyName).
		Str("pseudonym_id", pseudonymID).
		Msg("Creating API key")

	// Hash the raw key for storage
	keyHash := hashAPIKey(rawKey)

	// Convert permissions to JSON
	var permissionsJSON sql.Null[types.JSON[json.RawMessage]]
	if permissions != nil {
		permissionsBytes, err := json.Marshal(permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal permissions: %w", err)
		}
		permissionsJSON.Scan(permissionsBytes)
	}

	// Set creation time
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	// Set active status
	isActive := sql.Null[bool]{}
	isActive.Scan(true)

	// Set expiration if provided
	var expiresAtNull sql.Null[time.Time]
	if expiresAt != nil {
		expiresAtNull.Scan(*expiresAt)
	}

	// Set pseudonym ID
	pseudonymIDNull := sql.Null[string]{}
	pseudonymIDNull.Scan(pseudonymID)

	apiKeySetter := &models.APIKeySetter{
		KeyName:     &keyName,
		KeyHash:     &keyHash,
		PseudonymID: &pseudonymIDNull,
		Permissions: &permissionsJSON,
		CreatedAt:   &now,
		ExpiresAt:   &expiresAtNull,
		IsActive:    &isActive,
	}

	// Use the generated APIKeys table helper
	apiKey, err := models.APIKeys.Insert(apiKeySetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	log.Info().
		Int64("key_id", apiKey.KeyID).
		Str("key_name", keyName).
		Str("pseudonym_id", pseudonymID).
		Msg("API key created successfully")

	return apiKey, nil
}

// GetAPIKeyByHash retrieves an API key by its hash
func (dao *APIKeyDAO) GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	// Use the generated APIKeys table helper with where clause
	apiKeys, err := models.APIKeys.Query(
		models.SelectWhere.APIKeys.KeyHash.EQ(keyHash),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key by hash: %w", err)
	}

	if len(apiKeys) == 0 {
		return nil, nil
	}

	return apiKeys[0], nil
}

// GetAPIKeyByID retrieves an API key by ID
func (dao *APIKeyDAO) GetAPIKeyByID(ctx context.Context, keyID int64) (*models.APIKey, error) {
	// Use the generated FindAPIKey function
	apiKey, err := models.FindAPIKey(ctx, dao.db, keyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by ID: %w", err)
	}

	return apiKey, nil
}

// ValidateAPIKey validates an API key and returns user context if valid
func (dao *APIKeyDAO) ValidateAPIKey(ctx context.Context, rawKey string) (*APIKeyPermissions, string, error) {
	// Hash the raw key
	keyHash := hashAPIKey(rawKey)

	// Get the API key from database
	apiKey, err := dao.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, "", fmt.Errorf("failed to validate API key: %w", err)
	}

	if apiKey == nil {
		return nil, "", fmt.Errorf("API key not found")
	}

	// Check if key is active
	if !apiKey.IsActive.Valid || !apiKey.IsActive.V {
		return nil, "", fmt.Errorf("API key is inactive")
	}

	// Check if key has expired
	if apiKey.ExpiresAt.Valid && apiKey.ExpiresAt.V.Before(time.Now()) {
		return nil, "", fmt.Errorf("API key has expired")
	}

	// Check if pseudonym_id is set
	if !apiKey.PseudonymID.Valid || apiKey.PseudonymID.V == "" {
		return nil, "", fmt.Errorf("API key is not associated with a pseudonym")
	}

	// Update last used timestamp
	err = dao.UpdateLastUsed(ctx, apiKey.KeyID)
	if err != nil {
		log.Warn().Err(err).Int64("key_id", apiKey.KeyID).Msg("Failed to update API key last used timestamp")
	}

	// Parse permissions
	var permissions APIKeyPermissions
	if apiKey.Permissions.Valid {
		rawValue, err := apiKey.Permissions.V.Value()
		if err != nil {
			return nil, "", fmt.Errorf("failed to get API key permissions value: %w", err)
		}
		err = json.Unmarshal(rawValue.([]byte), &permissions)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse API key permissions: %w", err)
		}
	}

	log.Debug().
		Int64("key_id", apiKey.KeyID).
		Str("key_name", apiKey.KeyName).
		Str("pseudonym_id", apiKey.PseudonymID.V).
		Msg("API key validated successfully")

	return &permissions, apiKey.PseudonymID.V, nil
}

// UpdateAPIKey updates an API key
func (dao *APIKeyDAO) UpdateAPIKey(ctx context.Context, keyID int64, updates *models.APIKeySetter) error {
	// First get the API key
	apiKey, err := dao.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get API key for update: %w", err)
	}
	if apiKey == nil {
		return fmt.Errorf("API key not found")
	}

	// Use the generated Update method
	err = apiKey.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	log.Info().
		Int64("key_id", keyID).
		Msg("API key updated successfully")

	return nil
}

// DeleteAPIKey deletes an API key
func (dao *APIKeyDAO) DeleteAPIKey(ctx context.Context, keyID int64) error {
	// First get the API key
	apiKey, err := dao.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get API key for deletion: %w", err)
	}
	if apiKey == nil {
		return fmt.Errorf("API key not found")
	}

	// Use the generated Delete method
	err = apiKey.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	log.Info().
		Int64("key_id", keyID).
		Msg("API key deleted successfully")

	return nil
}

// DeactivateAPIKey deactivates an API key
func (dao *APIKeyDAO) DeactivateAPIKey(ctx context.Context, keyID int64) error {
	isActive := sql.Null[bool]{}
	isActive.Scan(false)

	updates := &models.APIKeySetter{
		IsActive: &isActive,
	}

	return dao.UpdateAPIKey(ctx, keyID, updates)
}

// ActivateAPIKey activates an API key
func (dao *APIKeyDAO) ActivateAPIKey(ctx context.Context, keyID int64) error {
	isActive := sql.Null[bool]{}
	isActive.Scan(true)

	updates := &models.APIKeySetter{
		IsActive: &isActive,
	}

	return dao.UpdateAPIKey(ctx, keyID, updates)
}

// UpdateLastUsed updates the last used timestamp for an API key
func (dao *APIKeyDAO) UpdateLastUsed(ctx context.Context, keyID int64) error {
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	updates := &models.APIKeySetter{
		LastUsedAt: &now,
	}

	return dao.UpdateAPIKey(ctx, keyID, updates)
}

// ListAPIKeys retrieves a list of API keys with pagination
func (dao *APIKeyDAO) ListAPIKeys(ctx context.Context, limit, offset int) ([]*models.APIKey, error) {
	// Use the generated APIKeys table helper
	apiKeys, err := models.APIKeys.Query(
		models.SelectWhere.APIKeys.KeyID.GT(0), // Simple condition to get all keys
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Apply pagination manually
	if offset >= len(apiKeys) {
		return []*models.APIKey{}, nil
	}

	end := offset + limit
	if end > len(apiKeys) {
		end = len(apiKeys)
	}

	return apiKeys[offset:end], nil
}

// CountAPIKeys counts the total number of API keys
func (dao *APIKeyDAO) CountAPIKeys(ctx context.Context) (int64, error) {
	// Use the generated APIKeys table helper with count
	count, err := models.APIKeys.Query().Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count API keys: %w", err)
	}

	return count, nil
}

// GetExpiredAPIKeys retrieves API keys that have expired
func (dao *APIKeyDAO) GetExpiredAPIKeys(ctx context.Context) ([]*models.APIKey, error) {
	now := time.Now()

	// Use the generated APIKeys table helper with where clause
	apiKeys, err := models.APIKeys.Query(
		models.SelectWhere.APIKeys.ExpiresAt.LT(now),
		models.SelectWhere.APIKeys.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired API keys: %w", err)
	}

	return apiKeys, nil
}

// CleanupExpiredAPIKeys deactivates expired API keys
func (dao *APIKeyDAO) CleanupExpiredAPIKeys(ctx context.Context) (int, error) {
	expiredKeys, err := dao.GetExpiredAPIKeys(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired API keys for cleanup: %w", err)
	}

	cleanedCount := 0
	for _, key := range expiredKeys {
		err = dao.DeactivateAPIKey(ctx, key.KeyID)
		if err != nil {
			log.Warn().Err(err).Int64("key_id", key.KeyID).Msg("Failed to deactivate expired API key")
			continue
		}
		cleanedCount++
	}

	if cleanedCount > 0 {
		log.Info().Int("cleaned_count", cleanedCount).Msg("Cleaned up expired API keys")
	}

	return cleanedCount, nil
}

// hashAPIKey creates a SHA-256 hash of the API key
func hashAPIKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// GetAPIKeysByPseudonymID retrieves all API keys for a specific pseudonym
func (dao *APIKeyDAO) GetAPIKeysByPseudonymID(ctx context.Context, pseudonymID string) ([]*models.APIKey, error) {
	// Use the generated APIKeys table helper with where clause
	apiKeys, err := models.APIKeys.Query(
		models.SelectWhere.APIKeys.PseudonymID.EQ(pseudonymID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys by pseudonym ID: %w", err)
	}

	return apiKeys, nil
}

// GetAPIKeyWithPseudonym retrieves an API key with its associated pseudonym information
func (dao *APIKeyDAO) GetAPIKeyWithPseudonym(ctx context.Context, keyID int64) (*models.APIKey, error) {
	// Use the generated FindAPIKey function
	apiKey, err := models.FindAPIKey(ctx, dao.db, keyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by ID: %w", err)
	}

	if apiKey == nil {
		return nil, nil
	}

	// Load the associated pseudonym
	err = apiKey.LoadPseudonym(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to load pseudonym for API key: %w", err)
	}

	return apiKey, nil
}
