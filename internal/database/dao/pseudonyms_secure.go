package dao

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// SecurePseudonymDAO handles pseudonym operations with role-based access control
type SecurePseudonymDAO struct {
	db                 bob.Executor
	ibeSystem          *ibe.IBESystem
	identityMappingDAO *IdentityMappingDAO
	userDAO            *UserDAO
	roleKeyDAO         *RoleKeyDAO
}

// NewSecurePseudonymDAO creates a new SecurePseudonymDAO
func NewSecurePseudonymDAO(db bob.Executor, ibeSystem *ibe.IBESystem, identityMappingDAO *IdentityMappingDAO, userDAO *UserDAO, roleKeyDAO *RoleKeyDAO) *SecurePseudonymDAO {
	return &SecurePseudonymDAO{
		db:                 db,
		ibeSystem:          ibeSystem,
		identityMappingDAO: identityMappingDAO,
		userDAO:            userDAO,
		roleKeyDAO:         roleKeyDAO,
	}
}

// GetPseudonymsByUserID retrieves all pseudonyms for a user using role-based access control
func (dao *SecurePseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64, roleName, scope string) ([]*models.Pseudonym, error) {
	// Validate that the key has the required capability
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, "access_own_pseudonyms")
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key does not have permission to access own pseudonyms")
	}

	// Get the role key for this operation
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, roleName, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key: %w", err)
	}

	// Use the key to access pseudonyms
	return dao.getPseudonymsByUserIDWithKey(ctx, userID, keyData)
}

// GetPseudonymsByRealIdentity retrieves all pseudonyms for a real identity using role-based access control
func (dao *SecurePseudonymDAO) GetPseudonymsByRealIdentity(ctx context.Context, realIdentity string, roleName, scope string) ([]*models.Pseudonym, error) {
	// Validate that the operation is allowed for this role/scope
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, "access_all_pseudonyms")
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key does not have permission to access all pseudonyms")
	}

	// Get the role key for this operation
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, roleName, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key: %w", err)
	}

	// Use the key to access pseudonyms
	return dao.getPseudonymsByRealIdentityWithKey(ctx, realIdentity, keyData)
}

// VerifyPseudonymOwnership verifies if a user owns a pseudonym using role-based access control
func (dao *SecurePseudonymDAO) VerifyPseudonymOwnership(ctx context.Context, pseudonymID string, userID int64, roleName, scope string) (bool, error) {
	// Validate that the key has the required capability
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, "verify_own_pseudonym_ownership")
	if err != nil {
		return false, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return false, fmt.Errorf("role key does not have permission to verify pseudonym ownership")
	}

	// Get the role key for this operation
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, roleName, scope)
	if err != nil {
		return false, fmt.Errorf("failed to get role key: %w", err)
	}

	// Use the key to verify ownership
	return dao.verifyPseudonymOwnershipWithKey(ctx, pseudonymID, userID, keyData)
}

// GetRealIdentityByPseudonym retrieves the real identity fingerprint for a pseudonym using role-based access control
func (dao *SecurePseudonymDAO) GetRealIdentityByPseudonym(ctx context.Context, pseudonymID string, roleName, scope string) (string, error) {
	// Validate that the operation is allowed for this role/scope
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, "cross_user_correlation")
	if err != nil {
		return "", fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return "", fmt.Errorf("role key does not have permission for cross-user correlation")
	}

	// Get the role key for this operation
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, roleName, scope)
	if err != nil {
		return "", fmt.Errorf("failed to get role key: %w", err)
	}

	// Use the key to get real identity
	return dao.getRealIdentityByPseudonymWithKey(ctx, pseudonymID, keyData)
}

// Internal methods that use the actual IBE keys

func (dao *SecurePseudonymDAO) getPseudonymsByUserIDWithKey(ctx context.Context, userID int64, keyData []byte) ([]*models.Pseudonym, error) {
	// 1. Get user's real identity (email)
	user, err := dao.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 2. Use IBE correlation to find all pseudonyms for this real identity
	return dao.getPseudonymsByRealIdentityWithKey(ctx, user.Email, keyData)
}

func (dao *SecurePseudonymDAO) getPseudonymsByRealIdentityWithKey(ctx context.Context, realIdentity string, keyData []byte) ([]*models.Pseudonym, error) {
	// 1. Generate fingerprint from real identity
	fingerprint := dao.ibeSystem.GenerateFingerprint(realIdentity)
	log.Info().
		Str("real_identity", realIdentity).
		Str("fingerprint", fingerprint).
		Msg("Generated fingerprint for real identity")

	// 2. Get all identity mappings for this fingerprint
	mappings, err := dao.identityMappingDAO.GetIdentityMappingsByFingerprint(ctx, fingerprint)
	if err != nil {
		log.Error().
			Err(err).
			Str("fingerprint", fingerprint).
			Msg("Failed to get identity mappings")
		return nil, fmt.Errorf("failed to get identity mappings: %w", err)
	}

	log.Info().
		Str("fingerprint", fingerprint).
		Int("mapping_count", len(mappings)).
		Msg("Found identity mappings for fingerprint")

	// 3. Extract pseudonym IDs and fetch pseudonyms (deduplicate by pseudonym ID)
	pseudonymMap := make(map[string]*models.Pseudonym)
	for _, mapping := range mappings {
		log.Info().
			Str("pseudonym_id", mapping.PseudonymID).
			Msg("Processing identity mapping")

		// Skip if we've already processed this pseudonym
		if _, exists := pseudonymMap[mapping.PseudonymID]; exists {
			continue
		}

		pseudonym, err := dao.GetPseudonymByID(ctx, mapping.PseudonymID)
		if err != nil {
			return nil, fmt.Errorf("failed to get pseudonym %s: %w", mapping.PseudonymID, err)
		}
		if pseudonym != nil {
			pseudonymMap[mapping.PseudonymID] = pseudonym
			log.Info().
				Str("pseudonym_id", mapping.PseudonymID).
				Str("display_name", pseudonym.DisplayName).
				Msg("Added pseudonym to results")
		}
	}

	// 4. Convert map to slice
	var pseudonyms []*models.Pseudonym
	for _, pseudonym := range pseudonymMap {
		pseudonyms = append(pseudonyms, pseudonym)
	}

	log.Info().
		Str("fingerprint", fingerprint).
		Int("pseudonym_count", len(pseudonyms)).
		Msg("Retrieved pseudonyms for real identity")

	return pseudonyms, nil
}

func (dao *SecurePseudonymDAO) verifyPseudonymOwnershipWithKey(ctx context.Context, pseudonymID string, userID int64, keyData []byte) (bool, error) {
	// 1. Get user's real identity
	user, err := dao.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	// 2. Get pseudonym's real identity fingerprint via IBE
	pseudonymFingerprint, err := dao.getRealIdentityByPseudonymWithKey(ctx, pseudonymID, keyData)
	if err != nil {
		return false, fmt.Errorf("failed to get pseudonym fingerprint: %w", err)
	}

	// 3. Compare fingerprints
	userFingerprint := dao.ibeSystem.GenerateFingerprint(user.Email)
	return pseudonymFingerprint == userFingerprint, nil
}

func (dao *SecurePseudonymDAO) getRealIdentityByPseudonymWithKey(ctx context.Context, pseudonymID string, keyData []byte) (string, error) {
	// 1. Get identity mapping for pseudonym with the correct key scope
	// For admin correlation, we need to get the correlation mapping
	// For self-correlation, we need to get the self_correlation mapping
	// We can determine the scope by trying to decrypt with the provided key
	// and checking which mapping works

	// Get all identity mappings for this pseudonym
	mappings, err := dao.identityMappingDAO.GetIdentityMappingsByPseudonymID(ctx, pseudonymID)
	if err != nil {
		return "", fmt.Errorf("failed to get identity mappings: %w", err)
	}
	if len(mappings) == 0 {
		return "", fmt.Errorf("no identity mappings found for pseudonym")
	}

	// Try to decrypt each mapping until we find one that works
	var decryptedMapping string
	for _, mapping := range mappings {
		decrypted, _, err := dao.ibeSystem.DecryptIdentity(mapping.EncryptedRealIdentity, keyData)
		if err == nil {
			decryptedMapping = decrypted
			break
		}
	}

	if decryptedMapping == "" {
		return "", fmt.Errorf("failed to decrypt any identity mapping with provided key")
	}

	// 3. Parse fingerprint from mapping
	mappingParts := strings.Split(decryptedMapping, ":")
	if len(mappingParts) != 2 {
		return "", fmt.Errorf("invalid decrypted mapping format")
	}

	// Return the fingerprint (not the real identity for privacy)
	return mappingParts[0], nil
}

// Helper method to get pseudonym by ID (reused from original DAO)
func (dao *SecurePseudonymDAO) GetPseudonymByID(ctx context.Context, pseudonymID string) (*models.Pseudonym, error) {
	// Use the generated FindPseudonym function
	pseudonym, err := models.FindPseudonym(ctx, dao.db, pseudonymID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pseudonym by ID: %w", err)
	}

	return pseudonym, nil
}

// ValidateKeyCapability is a helper method to validate key capabilities
func (dao *SecurePseudonymDAO) ValidateKeyCapability(ctx context.Context, roleName, scope, capability string) (bool, error) {
	return dao.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, capability)
}

// GetPseudonymByDisplayName retrieves a pseudonym by display name
func (dao *SecurePseudonymDAO) GetPseudonymByDisplayName(ctx context.Context, displayName string) (*models.Pseudonym, error) {
	pseudonyms, err := models.Pseudonyms.Query(
		models.SelectWhere.Pseudonyms.DisplayName.EQ(displayName),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get pseudonym by display name: %w", err)
	}
	if len(pseudonyms) == 0 {
		return nil, nil
	}
	return pseudonyms[0], nil
}

// UpdatePseudonym updates a pseudonym
func (dao *SecurePseudonymDAO) UpdatePseudonym(ctx context.Context, pseudonymID string, updates *models.PseudonymSetter) error {
	// First get the pseudonym
	pseudonym, err := dao.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		return fmt.Errorf("failed to get pseudonym for update: %w", err)
	}
	if pseudonym == nil {
		return fmt.Errorf("pseudonym not found")
	}

	// Use the generated Update method
	err = pseudonym.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update pseudonym: %w", err)
	}

	return nil
}

// DeletePseudonym deletes a pseudonym
func (dao *SecurePseudonymDAO) DeletePseudonym(ctx context.Context, pseudonymID string) error {
	// First get the pseudonym
	pseudonym, err := dao.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		return fmt.Errorf("failed to get pseudonym for deletion: %w", err)
	}
	if pseudonym == nil {
		return fmt.Errorf("pseudonym not found")
	}

	// Use the generated Delete method
	err = pseudonym.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete pseudonym: %w", err)
	}

	return nil
}

// UpdateLastActive updates the pseudonym's last active timestamp
func (dao *SecurePseudonymDAO) UpdateLastActive(ctx context.Context, pseudonymID string) error {
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	updates := &models.PseudonymSetter{
		LastActiveAt: &now,
	}

	return dao.UpdatePseudonym(ctx, pseudonymID, updates)
}

// GetDefaultPseudonymByUserID retrieves the default pseudonym for a user using role-based access control
func (dao *SecurePseudonymDAO) GetDefaultPseudonymByUserID(ctx context.Context, userID int64, roleName, scope string) (*models.Pseudonym, error) {
	// Validate that the key has the required capability
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, "access_own_pseudonyms")
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key does not have permission to access own pseudonyms")
	}

	// Get the role key for this operation
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, roleName, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key: %w", err)
	}

	// Use the key to access pseudonyms
	return dao.getDefaultPseudonymByUserIDWithKey(ctx, userID, keyData)
}

// getDefaultPseudonymByUserIDWithKey retrieves the default pseudonym for a user using the provided key
func (dao *SecurePseudonymDAO) getDefaultPseudonymByUserIDWithKey(ctx context.Context, userID int64, keyData []byte) (*models.Pseudonym, error) {
	// 1. Get user's real identity (email)
	user, err := dao.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 2. Use IBE correlation to find all pseudonyms for this real identity
	pseudonyms, err := dao.getPseudonymsByRealIdentityWithKey(ctx, user.Email, keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to get pseudonyms: %w", err)
	}

	// 3. Find the default pseudonym
	for _, pseudonym := range pseudonyms {
		if pseudonym.IsDefault {
			return pseudonym, nil
		}
	}

	// 4. If no default pseudonym found, return the first one (fallback)
	if len(pseudonyms) > 0 {
		log.Warn().
			Int64("user_id", userID).
			Msg("No default pseudonym found, using first pseudonym as fallback")
		return pseudonyms[0], nil
	}

	return nil, fmt.Errorf("no pseudonyms found for user")
}

// CreatePseudonymWithIdentityMapping creates a pseudonym and its identity mapping using role-based access control
func (dao *SecurePseudonymDAO) CreatePseudonymWithIdentityMapping(ctx context.Context, userID int64, displayName string) (*models.Pseudonym, error) {
	// 1. Create the pseudonym (set is_default if needed)
	pseudonym, err := dao.createPseudonym(ctx, displayName, &userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pseudonym: %w", err)
	}

	// 2. Get user's real identity (email) and role
	user, err := dao.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 3. Determine the user's role for encryption
	userRoles := []string{"user"} // Default role
	if user.Roles.Valid {
		var roles []string
		rolesBytes, err := user.Roles.V.Value()
		if err == nil {
			if err := json.Unmarshal(rolesBytes.([]byte), &roles); err == nil && len(roles) > 0 {
				userRoles = roles
			}
		}
	}

	// 4. Generate fingerprint for the real identity
	fingerprint := dao.ibeSystem.GenerateFingerprint(user.Email)
	log.Info().
		Str("real_identity", user.Email).
		Str("fingerprint", fingerprint).
		Str("user_role", userRoles[0]).
		Msg("Generated fingerprint during pseudonym creation")

	// 5. Create identity mappings using IBE
	// Create two identity mappings: one for self-correlation and one for admin correlation
	userRole := userRoles[0] // Use the first role for consistency

	// Get actual role keys from the database
	selfCorrelationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, userRole, "self_correlation")
	if err != nil {
		return nil, fmt.Errorf("failed to get self-correlation role key: %w", err)
	}

	// Create self-correlation mapping (for user self-verification)
	selfCorrelationFingerprint, err := dao.ibeSystem.EncryptIdentity(user.Email, pseudonym.PseudonymID, selfCorrelationKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt self-correlation identity mapping: %w", err)
	}

	// Create self-correlation identity mapping using Bob ORM
	selfCorrelationMapping := &models.IdentityMappingSetter{
		Fingerprint:               &fingerprint,
		PseudonymID:               &pseudonym.PseudonymID,
		EncryptedRealIdentity:     &selfCorrelationFingerprint,
		EncryptedPseudonymMapping: &selfCorrelationFingerprint,
		KeyVersion:                &[]int32{int32(dao.ibeSystem.GetKeyVersion())}[0],
		UserID:                    &userID,
		KeyScope:                  &[]string{"self_correlation"}[0],
	}

	_, err = models.IdentityMappings.Insert(selfCorrelationMapping).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create self-correlation identity mapping: %w", err)
	}

	// Only create correlation mapping for admin roles (they have correlation keys)
	adminRoles := []string{"platform_admin", "trust_safety", "legal_team"}
	isAdminRole := false
	for _, adminRole := range adminRoles {
		if userRole == adminRole {
			isAdminRole = true
			break
		}
	}

	if isAdminRole {
		// Get correlation key for admin role
		correlationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, userRole, "correlation")
		if err != nil {
			return nil, fmt.Errorf("failed to get correlation role key: %w", err)
		}

		// Create correlation mapping (for admin correlation)
		correlationFingerprint, err := dao.ibeSystem.EncryptIdentity(user.Email, pseudonym.PseudonymID, correlationKeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt correlation identity mapping: %w", err)
		}

		// Create correlation identity mapping using Bob ORM
		correlationMapping := &models.IdentityMappingSetter{
			Fingerprint:               &fingerprint,
			PseudonymID:               &pseudonym.PseudonymID,
			EncryptedRealIdentity:     &correlationFingerprint,
			EncryptedPseudonymMapping: &correlationFingerprint,
			KeyVersion:                &[]int32{int32(dao.ibeSystem.GetKeyVersion())}[0],
			UserID:                    &userID,
			KeyScope:                  &[]string{"correlation"}[0],
		}

		_, err = models.IdentityMappings.Insert(correlationMapping).One(ctx, dao.db)
		if err != nil {
			return nil, fmt.Errorf("failed to create correlation identity mapping: %w", err)
		}
	}

	return pseudonym, nil
}

// createPseudonym creates a new pseudonym (internal method)
func (dao *SecurePseudonymDAO) createPseudonym(ctx context.Context, displayName string, userID *int64) (*models.Pseudonym, error) {
	log.Debug().
		Str("display_name", displayName).
		Msg("Creating pseudonym")

	// Generate a unique pseudonym ID
	pseudonymID := generatePseudonymID()

	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	isActive := sql.Null[bool]{}
	isActive.Scan(true)

	isDefaultVal := false
	if userID != nil {
		// Check if user already has a default pseudonym using bob
		existingPseudonyms, err := models.Pseudonyms.Query(
			models.SelectWhere.Pseudonyms.IsDefault.EQ(true),
			models.SelectWhere.Pseudonyms.IsActive.EQ(true),
		).All(ctx, dao.db)
		if err == nil && len(existingPseudonyms) == 0 {
			isDefaultVal = true
		}
	}

	pseudonymSetter := &models.PseudonymSetter{
		PseudonymID: &pseudonymID,
		DisplayName: &displayName,
		CreatedAt:   &now,
		IsActive:    &isActive,
		IsDefault:   &isDefaultVal,
	}

	pseudonym, err := models.Pseudonyms.Insert(pseudonymSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create pseudonym: %w", err)
	}

	return pseudonym, nil
}

// generatePseudonymID generates a unique pseudonym ID
func generatePseudonymID() string {
	// Generate 32 random bytes and encode as hex
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
