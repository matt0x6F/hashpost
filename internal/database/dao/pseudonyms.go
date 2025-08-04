package dao

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// PseudonymDAO handles pseudonym operations with role-based access control
type PseudonymDAO struct {
	db                 bob.Executor
	ibeSystem          *ibe.IBESystem
	identityMappingDAO *IdentityMappingDAO
	userDAO            *UserDAO
	roleKeyDAO         *RoleKeyDAO
	userBlocksDAO      *UserBlocksDAO
}

// NewPseudonymDAO creates a new PseudonymDAO
func NewPseudonymDAO(db bob.Executor, ibeSystem *ibe.IBESystem, identityMappingDAO *IdentityMappingDAO, userDAO *UserDAO, roleKeyDAO *RoleKeyDAO, userBlocksDAO *UserBlocksDAO) *PseudonymDAO {
	return &PseudonymDAO{
		db:                 db,
		ibeSystem:          ibeSystem,
		identityMappingDAO: identityMappingDAO,
		userDAO:            userDAO,
		roleKeyDAO:         roleKeyDAO,
		userBlocksDAO:      userBlocksDAO,
	}
}

// GetPseudonymsByUserID retrieves all pseudonyms for a user using role-based access control
func (dao *PseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64, roleName, scope string) ([]*models.Pseudonym, error) {
	// Get the user's default pseudonym for validation
	defaultPseudonym, err := dao.getDefaultPseudonymByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default pseudonym for validation: %w", err)
	}

	// Validate that the key has the required capability using the default pseudonym
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, defaultPseudonym.PseudonymID, scope, constants.CapabilityAccessOwnPseudonyms, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key does not have permission to access own pseudonyms")
	}

	// Use the key to access pseudonyms
	return dao.getPseudonymsByUserID(ctx, userID)
}

// GetPseudonymsByRealIdentity retrieves all pseudonyms for a real identity using role-based access control
func (dao *PseudonymDAO) GetPseudonymsByRealIdentity(ctx context.Context, realIdentity string, roleName, scope string) ([]*models.Pseudonym, error) {
	// This is a privileged operation that requires admin-level access
	// We'll validate using a system-level pseudonym check
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, constants.SystemPseudonymID, scope, constants.CapabilityAccessAllPseudonyms, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key does not have permission to access all pseudonyms")
	}

	// Use the key to access pseudonyms
	return dao.getPseudonymsByRealIdentity(ctx, realIdentity)
}

// VerifyPseudonymOwnership verifies if a user owns a pseudonym using role-based access control
func (dao *PseudonymDAO) VerifyPseudonymOwnership(ctx context.Context, pseudonymID string, userID int64, roleName, scope string) (bool, error) {
	// Get the user's default pseudonym for validation
	defaultPseudonym, err := dao.getDefaultPseudonymByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get default pseudonym for validation: %w", err)
	}

	// Validate that the key has the required capability using the default pseudonym
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, defaultPseudonym.PseudonymID, scope, constants.CapabilityVerifyOwnPseudonymOwnership, nil)
	if err != nil {
		return false, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return false, fmt.Errorf("role key does not have permission to verify pseudonym ownership")
	}

	// Get the role key for this operation using the default pseudonym
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, defaultPseudonym.PseudonymID, scope, nil)
	if err != nil {
		return false, fmt.Errorf("failed to get role key: %w", err)
	}

	// Use the key to verify ownership
	return dao.verifyPseudonymOwnershipWithKey(ctx, pseudonymID, userID, keyData)
}

// GetRealIdentityByPseudonym retrieves the real identity fingerprint for a pseudonym using role-based access control
func (dao *PseudonymDAO) GetRealIdentityByPseudonym(ctx context.Context, pseudonymID string, roleName, scope string) (string, error) {
	// This is a privileged operation that requires admin-level access
	// We'll validate using a system-level pseudonym check
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, constants.SystemPseudonymID, scope, constants.CapabilityCrossUserCorrelation, nil)
	if err != nil {
		return "", fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return "", fmt.Errorf("role key does not have permission for cross-user correlation")
	}

	// Get the role key for this operation using system admin
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, constants.SystemPseudonymID, scope, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get role key: %w", err)
	}

	// Use the key to get real identity
	return dao.getRealIdentityByPseudonymWithKey(ctx, pseudonymID, keyData)
}

// Internal methods that use the actual IBE keys

func (dao *PseudonymDAO) getPseudonymsByUserID(ctx context.Context, userID int64) ([]*models.Pseudonym, error) {
	// 1. Get user's real identity (email)
	user, err := dao.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 2. Use IBE correlation to find all pseudonyms for this real identity
	return dao.getPseudonymsByRealIdentity(ctx, user.Email)
}

func (dao *PseudonymDAO) getPseudonymsByRealIdentity(ctx context.Context, realIdentity string) ([]*models.Pseudonym, error) {
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

func (dao *PseudonymDAO) verifyPseudonymOwnershipWithKey(ctx context.Context, pseudonymID string, userID int64, keyData []byte) (bool, error) {
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

func (dao *PseudonymDAO) getRealIdentityByPseudonymWithKey(ctx context.Context, pseudonymID string, keyData []byte) (string, error) {
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
func (dao *PseudonymDAO) GetPseudonymByID(ctx context.Context, pseudonymID string) (*models.Pseudonym, error) {
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

// GetPseudonymByDisplayName retrieves a pseudonym by display name
func (dao *PseudonymDAO) GetPseudonymByDisplayName(ctx context.Context, displayName string) (*models.Pseudonym, error) {
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

// GetPseudonymBySlug retrieves a pseudonym by slug
func (dao *PseudonymDAO) GetPseudonymBySlug(ctx context.Context, slug string) (*models.Pseudonym, error) {
	pseudonyms, err := models.Pseudonyms.Query(
		models.SelectWhere.Pseudonyms.Slug.EQ(slug),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get pseudonym by slug: %w", err)
	}
	if len(pseudonyms) == 0 {
		return nil, nil
	}
	return pseudonyms[0], nil
}

// GenerateSlugFromDisplayName generates a URL-friendly slug from a display name
func (dao *PseudonymDAO) GenerateSlugFromDisplayName(ctx context.Context, displayName string) (string, error) {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(strings.ReplaceAll(displayName, " ", "-"))
	
	// Remove any non-alphanumeric characters except hyphens
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	
	// Remove multiple consecutive hyphens
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	
	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Ensure slug is not empty
	if slug == "" {
		slug = "user"
	}
	
	// Check if slug already exists and append number if needed
	originalSlug := slug
	counter := 1
	for {
		existing, err := dao.GetPseudonymBySlug(ctx, slug)
		if err != nil {
			return "", fmt.Errorf("failed to check slug availability: %w", err)
		}
		if existing == nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}
	
	return slug, nil
}

// CalculateKarmaForPseudonym calculates the total karma for a pseudonym based on their posts and comments
func (dao *PseudonymDAO) CalculateKarmaForPseudonym(ctx context.Context, pseudonymID string) (int32, error) {
	// Get all posts by the pseudonym
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.And(
			psql.Group(psql.Or(
				psql.Quote("posts", "is_removed").IsNull(),
				psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
			)),
			psql.Group(psql.Or(
				psql.Quote("posts", "is_deleted").IsNull(),
				psql.Quote("posts", "is_deleted").EQ(psql.Arg(false)),
			)),
		))),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts for karma calculation: %w", err)
	}

	// Get all comments by the pseudonym
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_deleted").IsNull(),
			psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
		))),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments for karma calculation: %w", err)
	}

	// Calculate total karma from posts
	postKarma := int32(0)
	for _, post := range posts {
		if post.Score.Valid {
			postKarma += post.Score.V
		}
	}

	// Calculate total karma from comments
	commentKarma := int32(0)
	for _, comment := range comments {
		if comment.Score.Valid {
			commentKarma += comment.Score.V
		}
	}

	totalKarma := postKarma + commentKarma
	return totalKarma, nil
}

// UpdateKarmaForPseudonym calculates and updates the karma for a pseudonym
func (dao *PseudonymDAO) UpdateKarmaForPseudonym(ctx context.Context, pseudonymID string) error {
	karma, err := dao.CalculateKarmaForPseudonym(ctx, pseudonymID)
	if err != nil {
		return fmt.Errorf("failed to calculate karma: %w", err)
	}

	updates := &models.PseudonymSetter{
		KarmaScore: &sql.Null[int32]{V: karma, Valid: true},
	}

	err = dao.UpdatePseudonym(ctx, pseudonymID, updates)
	if err != nil {
		return fmt.Errorf("failed to update karma: %w", err)
	}

	return nil
}

// UpdatePseudonym updates a pseudonym
func (dao *PseudonymDAO) UpdatePseudonym(ctx context.Context, pseudonymID string, updates *models.PseudonymSetter) error {
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
func (dao *PseudonymDAO) DeletePseudonym(ctx context.Context, pseudonymID string) error {
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
func (dao *PseudonymDAO) UpdateLastActive(ctx context.Context, pseudonymID string) error {
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	updates := &models.PseudonymSetter{
		LastActiveAt: &now,
	}

	return dao.UpdatePseudonym(ctx, pseudonymID, updates)
}

// GetDefaultPseudonymByUserID retrieves the default pseudonym for a user using role-based access control
func (dao *PseudonymDAO) GetDefaultPseudonymByUserID(ctx context.Context, userID int64, roleName, scope string) (*models.Pseudonym, error) {
	// Get the user's default pseudonym for validation
	defaultPseudonym, err := dao.getDefaultPseudonymByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default pseudonym for validation: %w", err)
	}

	// Validate that the key has the required capability using the default pseudonym
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, defaultPseudonym.PseudonymID, scope, constants.CapabilityAccessOwnPseudonyms, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key does not have permission to access own pseudonyms")
	}

	// Return the default pseudonym
	return defaultPseudonym, nil
}

// getDefaultPseudonymByUserID retrieves the default pseudonym for a user using the provided key
func (dao *PseudonymDAO) getDefaultPseudonymByUserID(ctx context.Context, userID int64) (*models.Pseudonym, error) {
	// 1. Get user's real identity (email)
	user, err := dao.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 2. Use IBE correlation to find all pseudonyms for this real identity
	pseudonyms, err := dao.getPseudonymsByRealIdentity(ctx, user.Email)
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
func (dao *PseudonymDAO) CreatePseudonymWithIdentityMapping(ctx context.Context, userID int64, displayName string) (*models.Pseudonym, error) {
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
	// Create identity mappings for both authentication and self-correlation scopes
	userRole := userRoles[0] // Use the first role for consistency

	// Get authentication role key from the database
	// Since we're creating a new pseudonym, we need to use a system-level key for initial setup
	authenticationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, constants.SystemPseudonymID, "authentication", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication role key: %w", err)
	}

	// Create authentication mapping (for login and session management)
	authenticationFingerprint, err := dao.ibeSystem.EncryptIdentity(user.Email, pseudonym.PseudonymID, authenticationKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt authentication identity mapping: %w", err)
	}

	// Create authentication identity mapping using Bob ORM
	authenticationMapping := &models.IdentityMappingSetter{
		Fingerprint:               &fingerprint,
		PseudonymID:               &pseudonym.PseudonymID,
		EncryptedRealIdentity:     &authenticationFingerprint,
		EncryptedPseudonymMapping: &authenticationFingerprint,
		KeyVersion:                &[]int32{int32(dao.ibeSystem.GetKeyVersion())}[0],
		UserID:                    &userID,
		KeyScope:                  &[]string{constants.ScopeAuthentication}[0],
	}

	_, err = models.IdentityMappings.Insert(authenticationMapping).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication identity mapping: %w", err)
	}

	// Get self-correlation role key from the database
	// Since we're creating a new pseudonym, we need to use a system-level key for initial setup
	selfCorrelationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, constants.SystemPseudonymID, constants.ScopeSelfCorrelation, nil)
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
		KeyScope:                  &[]string{constants.ScopeSelfCorrelation}[0],
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
		// Since we're creating a new pseudonym, we need to use a system-level key for initial setup
		correlationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, constants.SystemPseudonymID, constants.ScopeCorrelation, nil)
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
			// Ensure KeyVersion is within int32 bounds
			UserID:                    &userID,
			KeyScope:                  &[]string{constants.ScopeCorrelation}[0],
		}

		_, err = models.IdentityMappings.Insert(correlationMapping).One(ctx, dao.db)
		if err != nil {
			return nil, fmt.Errorf("failed to create correlation identity mapping: %w", err)
		}
	}

	return pseudonym, nil
}

// createPseudonym creates a new pseudonym (internal method)
func (dao *PseudonymDAO) createPseudonym(ctx context.Context, displayName string, userID *int64) (*models.Pseudonym, error) {
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
		// Check if user already has a default pseudonym by getting all pseudonyms for this user
		// and checking if any are marked as default
		user, err := dao.userDAO.GetUserByID(ctx, *userID)
		if err == nil && user != nil {
			// Get all pseudonyms for this user's real identity
			pseudonyms, err := dao.getPseudonymsByRealIdentity(ctx, user.Email)
			if err == nil {
				// Check if any existing pseudonym is default
				hasDefault := false
				for _, pseudonym := range pseudonyms {
					if pseudonym.IsDefault {
						hasDefault = true
						break
					}
				}
				// If no default pseudonym exists, make this one default
				if !hasDefault {
					isDefaultVal = true
				}
			}
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

// GetUserIDByPseudonym gets the user ID for a pseudonym using IBE correlation
func (dao *PseudonymDAO) GetUserIDByPseudonym(ctx context.Context, pseudonymID, roleName, scope string) (int64, error) {
	log.Debug().
		Str("pseudonym_id", pseudonymID).
		Str("role_name", roleName).
		Str("scope", scope).
		Msg("Getting user ID by pseudonym using IBE correlation")

	// This is a privileged operation that requires admin-level access
	// We'll validate using a system-level pseudonym check
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, constants.SystemPseudonymID, scope, constants.CapabilityCrossUserCorrelation, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return 0, fmt.Errorf("role key does not have permission for cross-user correlation")
	}

	// Get the identity mapping for this pseudonym
	mapping, err := dao.identityMappingDAO.GetIdentityMappingByPseudonymID(ctx, pseudonymID)
	if err != nil {
		return 0, fmt.Errorf("failed to get identity mapping: %w", err)
	}
	if mapping == nil {
		return 0, fmt.Errorf("identity mapping not found for pseudonym")
	}

	// Get the user ID from the mapping
	return mapping.UserID, nil
}

// DeactivatePseudonym deactivates a pseudonym, preventing it from being used for new content
// Users cannot reactivate deactivated pseudonyms
func (dao *PseudonymDAO) DeactivatePseudonym(ctx context.Context, pseudonymID string, userID int64, roleName, scope string) error {
	// Get the user's default pseudonym for validation
	defaultPseudonym, err := dao.getDefaultPseudonymByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get default pseudonym for validation: %w", err)
	}

	// Validate that the key has the required capability using the default pseudonym
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, defaultPseudonym.PseudonymID, scope, constants.CapabilityManageOwnPseudonyms, nil)
	if err != nil {
		return fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return fmt.Errorf("role key does not have permission to manage own pseudonyms")
	}

	// Verify that the user owns this pseudonym
	ownsPseudonym, err := dao.VerifyPseudonymOwnership(ctx, pseudonymID, userID, roleName, scope)
	if err != nil {
		return fmt.Errorf("failed to verify pseudonym ownership: %w", err)
	}

	if !ownsPseudonym {
		return fmt.Errorf("user does not own this pseudonym")
	}

	// Get the pseudonym to check if it's already deactivated
	pseudonym, err := dao.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		return fmt.Errorf("failed to get pseudonym: %w", err)
	}

	if pseudonym == nil {
		return fmt.Errorf("pseudonym not found")
	}

	// Check if pseudonym is already deactivated
	if !pseudonym.IsActive.Valid || !pseudonym.IsActive.V {
		return fmt.Errorf("pseudonym is already deactivated")
	}

	// Check if this is the user's default pseudonym
	if pseudonym.IsDefault {
		return fmt.Errorf("cannot deactivate default pseudonym")
	}

	// Deactivate the pseudonym
	return dao.deactivatePseudonym(ctx, pseudonymID)
}

// deactivatePseudonym deactivates a pseudonym using the provided key
func (dao *PseudonymDAO) deactivatePseudonym(ctx context.Context, pseudonymID string) error {
	// Update the pseudonym to set is_active to false
	updates := &models.PseudonymSetter{
		IsActive: &sql.Null[bool]{V: false, Valid: true},
	}

	// Use the existing UpdatePseudonym method which doesn't require role-based access
	// since we've already validated ownership and permissions
	pseudonym, err := models.FindPseudonym(ctx, dao.db, pseudonymID)
	if err != nil {
		return fmt.Errorf("failed to find pseudonym: %w", err)
	}

	if pseudonym == nil {
		return fmt.Errorf("pseudonym not found")
	}

	err = pseudonym.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to deactivate pseudonym: %w", err)
	}

	log.Info().
		Str("pseudonym_id", pseudonymID).
		Str("display_name", pseudonym.DisplayName).
		Msg("Pseudonym deactivated successfully")

	return nil
}

// DeleteByUserID deletes all pseudonyms for a specific user
func (dao *PseudonymDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	// 1. Get user's real identity (email) first
	user, err := dao.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", userID, err)
	}
	if user == nil {
		return fmt.Errorf("user %d not found", userID)
	}

	// 2. Generate fingerprint from real identity
	fingerprint := dao.ibeSystem.GenerateFingerprint(user.Email)
	log.Info().
		Str("real_identity", user.Email).
		Str("fingerprint", fingerprint).
		Msg("Generated fingerprint for real identity")

	// 3. Get all identity mappings for this fingerprint
	mappings, err := dao.identityMappingDAO.GetIdentityMappingsByFingerprint(ctx, fingerprint)
	if err != nil {
		return fmt.Errorf("failed to get identity mappings for fingerprint %s: %w", fingerprint, err)
	}

	log.Info().
		Str("fingerprint", fingerprint).
		Int("mapping_count", len(mappings)).
		Msg("Found identity mappings for fingerprint")

	// 4. Extract pseudonym IDs and delete each pseudonym (deduplicate by pseudonym ID)
	pseudonymIDs := make(map[string]bool)
	for _, mapping := range mappings {
		pseudonymIDs[mapping.PseudonymID] = true
	}

	// 5. Delete each unique pseudonym
	for pseudonymID := range pseudonymIDs {
		pseudonym, err := models.FindPseudonym(ctx, dao.db, pseudonymID)
		if err != nil {
			return fmt.Errorf("failed to find pseudonym %s: %w", pseudonymID, err)
		}
		if pseudonym != nil {
			err = pseudonym.Delete(ctx, dao.db)
			if err != nil {
				return fmt.Errorf("failed to delete pseudonym %s: %w", pseudonymID, err)
			}
			log.Info().
				Str("pseudonym_id", pseudonymID).
				Str("display_name", pseudonym.DisplayName).
				Msg("Deleted pseudonym")
		}
	}

	return nil
}
