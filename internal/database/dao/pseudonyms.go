package dao

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

// NewPseudonymDAOForKarma creates a new PseudonymDAO specifically for karma operations
// This constructor only requires the database connection since karma operations don't need
// the other dependencies like IBE system, role keys, etc.
func NewPseudonymDAOForKarma(db bob.Executor) *PseudonymDAO {
	return &PseudonymDAO{
		db: db,
		// Other fields remain nil as they're not needed for karma operations
	}
}

// GetPseudonymsByUserID retrieves all pseudonyms for a user using role-based access control
func (dao *PseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64, activePseudonymID, roleName, scope string) ([]*models.Pseudonym, error) {
	// Get the role key for the active pseudonym to verify ownership
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, activePseudonymID, scope, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key for pseudonym verification: %w", err)
	}

	// Validate that the active pseudonym belongs to the user
	ownsPseudonym, err := dao.verifyPseudonymOwnershipWithKey(ctx, activePseudonymID, userID, keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify active pseudonym ownership: %w", err)
	}

	if !ownsPseudonym {
		return nil, fmt.Errorf("active pseudonym does not belong to user")
	}

	// Validate that the key has the required capability using the active pseudonym
	hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, activePseudonymID, scope, constants.CapabilityAccessOwnPseudonyms, nil)
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
	// We'll validate using platform admin role
	hasCapability, err := dao.roleKeyDAO.ValidatePlatformKeyCapability(ctx, constants.RolePlatformAdmin, scope, constants.CapabilityAccessAllPseudonyms)
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
func (dao *PseudonymDAO) VerifyPseudonymOwnership(ctx context.Context, pseudonymID string, userID int64, activePseudonymID, roleName, scope string) (bool, error) {
	// Get the role key for the active pseudonym to verify ownership
	activePseudonymKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, activePseudonymID, scope, nil)
	if err != nil {
		return false, fmt.Errorf("failed to get role key for active pseudonym verification: %w", err)
	}

	// Validate that the active pseudonym belongs to the user first
	ownsActivePseudonym, err := dao.verifyPseudonymOwnershipWithKey(ctx, activePseudonymID, userID, activePseudonymKeyData)
	if err != nil {
		return false, fmt.Errorf("failed to verify active pseudonym ownership: %w", err)
	}

	if !ownsActivePseudonym {
		return false, fmt.Errorf("active pseudonym does not belong to user")
	}

	log.Info().
		Str("active_pseudonym_id", activePseudonymID).
		Str("target_pseudonym_id", pseudonymID).
		Str("scope", scope).
		Msg("Verifying pseudonym ownership with active pseudonym")

	// Get the role key for this operation using the active pseudonym
	keyData, err := dao.roleKeyDAO.GetKeyData(ctx, activePseudonymID, scope, nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("active_pseudonym_id", activePseudonymID).
			Str("scope", scope).
			Msg("Failed to get role key")
		return false, fmt.Errorf("failed to get role key: %w", err)
	}

	log.Info().
		Str("active_pseudonym_id", activePseudonymID).
		Str("target_pseudonym_id", pseudonymID).
		Str("scope", scope).
		Msg("Got role key, verifying ownership")

	// Use the key to verify ownership
	return dao.verifyPseudonymOwnershipWithKey(ctx, pseudonymID, userID, keyData)
}

// GetRealIdentityByPseudonym retrieves the real identity fingerprint for a pseudonym using role-based access control
func (dao *PseudonymDAO) GetRealIdentityByPseudonym(ctx context.Context, pseudonymID string, roleName, scope string) (string, error) {
	// This is a privileged operation that requires admin-level access
	// We'll validate using platform admin role
	hasCapability, err := dao.roleKeyDAO.ValidatePlatformKeyCapability(ctx, constants.RolePlatformAdmin, scope, constants.CapabilityCrossUserCorrelation)
	if err != nil {
		return "", fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return "", fmt.Errorf("role key does not have permission for cross-user correlation")
	}

	// Get the role key for this operation using platform admin role
	keyData, err := dao.roleKeyDAO.GetPlatformKeyData(ctx, constants.RolePlatformAdmin, scope)
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

// GetPseudonymsByRealIdentityDirect retrieves all pseudonyms for a real identity without role-based access control
// This method is intended for administrative operations like admin user creation where role verification
// is not possible or necessary (e.g., when creating the first pseudonym for a user)
func (dao *PseudonymDAO) GetPseudonymsByRealIdentityDirect(ctx context.Context, realIdentity string) ([]*models.Pseudonym, error) {
	return dao.getPseudonymsByRealIdentity(ctx, realIdentity)
}

// GetIBESystemSalt returns the current salt used by the IBE system for debugging purposes
func (dao *PseudonymDAO) GetIBESystemSalt() string {
	return dao.ibeSystem.GetSalt()
}

// GenerateFingerprintForEmail generates a fingerprint for the given email using the current IBE system salt
// This method is for debugging purposes to help troubleshoot salt mismatches
func (dao *PseudonymDAO) GenerateFingerprintForEmail(email string) string {
	return dao.ibeSystem.GenerateFingerprint(email)
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

	log.Info().
		Str("pseudonym_id", pseudonymID).
		Int64("user_id", userID).
		Str("user_email", user.Email).
		Msg("Getting pseudonym fingerprint")

	// 2. Get pseudonym's real identity fingerprint via IBE
	pseudonymFingerprint, err := dao.getRealIdentityByPseudonymWithKey(ctx, pseudonymID, keyData)
	if err != nil {
		log.Error().
			Err(err).
			Str("pseudonym_id", pseudonymID).
			Msg("Failed to get pseudonym fingerprint")
		return false, fmt.Errorf("failed to get pseudonym fingerprint: %w", err)
	}

	log.Info().
		Str("pseudonym_id", pseudonymID).
		Str("pseudonym_fingerprint", pseudonymFingerprint).
		Msg("Got pseudonym fingerprint")

	// 3. Compare fingerprints
	userFingerprint := dao.ibeSystem.GenerateFingerprint(user.Email)

	log.Info().
		Str("pseudonym_id", pseudonymID).
		Str("pseudonym_fingerprint", pseudonymFingerprint).
		Str("user_fingerprint", userFingerprint).
		Bool("fingerprints_match", pseudonymFingerprint == userFingerprint).
		Msg("Comparing fingerprints")

	return pseudonymFingerprint == userFingerprint, nil
}

func (dao *PseudonymDAO) getRealIdentityByPseudonymWithKey(ctx context.Context, pseudonymID string, keyData []byte) (string, error) {
	// 1. Get identity mapping for pseudonym with the correct key scope
	// For admin correlation, we need to get the correlation mapping
	// For self-correlation, we need to get the self_correlation mapping
	// We can determine the scope by trying to decrypt with the provided key
	// and checking which mapping works

	log.Info().
		Str("pseudonym_id", pseudonymID).
		Msg("Getting identity mappings for pseudonym")

	// Get all identity mappings for this pseudonym
	mappings, err := dao.identityMappingDAO.GetIdentityMappingsByPseudonymID(ctx, pseudonymID)
	if err != nil {
		return "", fmt.Errorf("failed to get identity mappings: %w", err)
	}
	if len(mappings) == 0 {
		return "", fmt.Errorf("no identity mappings found for pseudonym")
	}

	log.Info().
		Str("pseudonym_id", pseudonymID).
		Int("mapping_count", len(mappings)).
		Msg("Found identity mappings")

	// Try to decrypt each mapping until we find one that works
	var decryptedMapping string
	for i, mapping := range mappings {
		log.Info().
			Str("pseudonym_id", pseudonymID).
			Int("mapping_index", i).
			Str("key_scope", mapping.KeyScope).
			Msg("Trying to decrypt mapping")

		decrypted, _, err := dao.ibeSystem.DecryptIdentity(mapping.EncryptedRealIdentity, keyData)
		if err == nil {
			decryptedMapping = decrypted
			log.Info().
				Str("pseudonym_id", pseudonymID).
				Int("mapping_index", i).
				Str("key_scope", mapping.KeyScope).
				Msg("Successfully decrypted mapping")
			break
		} else {
			log.Info().
				Str("pseudonym_id", pseudonymID).
				Int("mapping_index", i).
				Str("key_scope", mapping.KeyScope).
				Err(err).
				Msg("Failed to decrypt mapping")
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

	log.Info().
		Str("pseudonym_id", pseudonymID).
		Str("fingerprint", mappingParts[0]).
		Msg("Extracted fingerprint from mapping")

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
	// Get all posts by the pseudonym (excluding only moderator-removed content)
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("posts", "is_removed").IsNull(),
			psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
		))),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts for karma calculation: %w", err)
	}

	// Get all comments by the pseudonym (excluding only moderator-removed content)
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
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

	// Get the default pseudonym for the user to determine roles
	defaultPseudonym, err := dao.getDefaultPseudonymByUserID(ctx, userID)
	if err == nil && defaultPseudonym != nil {
		// Get role keys for the default pseudonym to determine user roles
		roleKeys, err := dao.roleKeyDAO.ListRoleKeysByPseudonym(ctx, defaultPseudonym.PseudonymID)
		if err == nil && len(roleKeys) > 0 {
			roleSet := make(map[string]bool)
			for _, roleKey := range roleKeys {
				// Skip subforum-specific keys for role determination
				if roleKey.SubforumID.Valid {
					continue
				}
				roleSet[roleKey.RoleName] = true
			}

			// Convert set to slice
			if len(roleSet) > 0 {
				userRoles = make([]string, 0, len(roleSet))
				for role := range roleSet {
					userRoles = append(userRoles, role)
				}
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

	// Ensure the user has role keys for this pseudonym
	err = dao.roleKeyDAO.EnsureDefaultKeys(ctx, dao.ibeSystem, pseudonym.PseudonymID, userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure default keys: %w", err)
	}

	// Get authentication role key for this pseudonym using the user's role
	authenticationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, pseudonym.PseudonymID, constants.ScopeAuthentication, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication role key: %w", err)
	}

	// Create authentication mapping (for login and session management)
	// Use the correct domain based on the user's role
	authenticationDomain := dao.ibeSystem.GetDomainForRole(userRole)
	authenticationFingerprint, err := dao.ibeSystem.EncryptIdentityWithDomain(user.Email, pseudonym.PseudonymID, authenticationDomain, authenticationKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt authentication identity mapping: %w", err)
	}

	// Create authentication identity mapping using Bob ORM
	keyVersion := dao.ibeSystem.GetKeyVersion()
	authenticationMapping := &models.IdentityMappingSetter{
		Fingerprint:               &fingerprint,
		PseudonymID:               &pseudonym.PseudonymID,
		EncryptedRealIdentity:     &authenticationFingerprint,
		EncryptedPseudonymMapping: &authenticationFingerprint,
		KeyVersion:                &keyVersion,
		UserID:                    &userID,
		KeyScope:                  &[]string{constants.ScopeAuthentication}[0],
	}

	_, err = models.IdentityMappings.Insert(authenticationMapping).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication identity mapping: %w", err)
	}

	// Get self-correlation role key for this pseudonym using the user's role
	selfCorrelationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, pseudonym.PseudonymID, constants.ScopeSelfCorrelation, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get self-correlation role key: %w", err)
	}

	// Create self-correlation mapping (for user self-verification)
	// Use the correct domain based on the user's role
	selfCorrelationDomain := dao.ibeSystem.GetDomainForRole(userRole)
	selfCorrelationFingerprint, err := dao.ibeSystem.EncryptIdentityWithDomain(user.Email, pseudonym.PseudonymID, selfCorrelationDomain, selfCorrelationKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt self-correlation identity mapping: %w", err)
	}

	// Create self-correlation identity mapping using Bob ORM
	selfCorrelationMapping := &models.IdentityMappingSetter{
		Fingerprint:               &fingerprint,
		PseudonymID:               &pseudonym.PseudonymID,
		EncryptedRealIdentity:     &selfCorrelationFingerprint,
		EncryptedPseudonymMapping: &selfCorrelationFingerprint,
		KeyVersion:                &keyVersion,
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
		// Get correlation key for admin role using the user's role
		correlationKeyData, err := dao.roleKeyDAO.GetKeyData(ctx, pseudonym.PseudonymID, constants.ScopeCorrelation, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get correlation role key: %w", err)
		}

		// Create correlation mapping (for admin correlation)
		// Use the correct domain based on the user's role
		correlationDomain := dao.ibeSystem.GetDomainForRole(userRole)
		correlationFingerprint, err := dao.ibeSystem.EncryptIdentityWithDomain(user.Email, pseudonym.PseudonymID, correlationDomain, correlationKeyData)
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
			UserID:   &userID,
			KeyScope: &[]string{constants.ScopeCorrelation}[0],
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
	if userID == nil {
		return nil, fmt.Errorf("userID cannot be nil")
	}

	log.Debug().
		Str("display_name", displayName).
		Msg("Creating pseudonym")

	// Generate a unique pseudonym ID using IBE
	pseudonymID := dao.generatePseudonymID(*userID)

	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	isActive := sql.Null[bool]{}
	isActive.Scan(true)

	isDefaultVal := false
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

// generatePseudonymID generates a unique pseudonym ID using IBE
func (dao *PseudonymDAO) generatePseudonymID(userID int64) string {
	// Use IBE system to generate pseudonym with user_pseudonyms domain
	// Use display name as context for uniqueness
	context := fmt.Sprintf("pseudonym_%d_%d", userID, time.Now().Unix())
	return dao.ibeSystem.GeneratePseudonym(userID, context, dao.ibeSystem.GetKeyVersion())
}

// GetUserIDByPseudonym gets the user ID for a pseudonym using IBE correlation
func (dao *PseudonymDAO) GetUserIDByPseudonym(ctx context.Context, pseudonymID, roleName, scope string) (int64, error) {
	log.Debug().
		Str("pseudonym_id", pseudonymID).
		Str("role_name", roleName).
		Str("scope", scope).
		Msg("Getting user ID by pseudonym using IBE correlation")

	// This is a privileged operation that requires admin-level access
	// We'll validate using platform admin role
	hasCapability, err := dao.roleKeyDAO.ValidatePlatformKeyCapability(ctx, constants.RolePlatformAdmin, scope, constants.CapabilityCrossUserCorrelation)
	if err != nil {
		return 0, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return 0, fmt.Errorf("role key does not have permission for cross-user correlation")
	}

	// Get the role key for this operation using platform admin role
	keyData, err := dao.roleKeyDAO.GetPlatformKeyData(ctx, constants.RolePlatformAdmin, scope)
	if err != nil {
		return 0, fmt.Errorf("failed to get role key: %w", err)
	}

	// Use IBE decryption to get the real identity (email)
	realIdentity, err := dao.getRealIdentityByPseudonymWithKey(ctx, pseudonymID, keyData)
	if err != nil {
		return 0, fmt.Errorf("failed to get real identity by pseudonym: %w", err)
	}

	// Get the user by email (real identity)
	user, err := dao.userDAO.GetUserByEmail(ctx, realIdentity)
	if err != nil {
		return 0, fmt.Errorf("failed to get user by email: %w", err)
	}
	if user == nil {
		return 0, fmt.Errorf("user not found for real identity")
	}

	return user.UserID, nil
}

// ArePseudonymsOwnedBySameUser checks if two pseudonyms belong to the same user using IBE correlation
func (dao *PseudonymDAO) ArePseudonymsOwnedBySameUser(ctx context.Context, pseudonymID1, pseudonymID2 string) (bool, error) {
	log.Debug().
		Str("pseudonym_id_1", pseudonymID1).
		Str("pseudonym_id_2", pseudonymID2).
		Msg("Checking if pseudonyms are owned by same user")

	// Get platform admin key for cross-user correlation
	keyData, err := dao.roleKeyDAO.GetPlatformKeyData(ctx, constants.RolePlatformAdmin, constants.ScopeCorrelation)
	if err != nil {
		return false, fmt.Errorf("failed to get platform admin key: %w", err)
	}

	// Get real identity fingerprints for both pseudonyms
	fingerprint1, err := dao.getRealIdentityFingerprintByPseudonymWithKey(ctx, pseudonymID1, keyData)
	if err != nil {
		return false, fmt.Errorf("failed to get fingerprint for pseudonym1: %w", err)
	}

	fingerprint2, err := dao.getRealIdentityFingerprintByPseudonymWithKey(ctx, pseudonymID2, keyData)
	if err != nil {
		return false, fmt.Errorf("failed to get fingerprint for pseudonym2: %w", err)
	}

	// Compare fingerprints
	sameUser := fingerprint1 == fingerprint2

	// Hash fingerprints for logging to avoid leaking sensitive correlation data
	hashFingerprint := func(fp string) string {
		sum := sha256.Sum256([]byte(fp))
		return hex.EncodeToString(sum[:])[:8] // log only first 8 hex chars
	}

	log.Debug().
		Str("pseudonym_id_1", pseudonymID1).
		Str("pseudonym_id_2", pseudonymID2).
		Str("fingerprint_1_hash", hashFingerprint(fingerprint1)).
		Str("fingerprint_2_hash", hashFingerprint(fingerprint2)).
		Bool("same_user", sameUser).
		Msg("Pseudonym ownership comparison result")

	return sameUser, nil
}

// getRealIdentityFingerprintByPseudonymWithKey gets the fingerprint for a pseudonym using the provided key
func (dao *PseudonymDAO) getRealIdentityFingerprintByPseudonymWithKey(ctx context.Context, pseudonymID string, keyData []byte) (string, error) {
	// Get identity mappings for this pseudonym
	mappings, err := dao.identityMappingDAO.GetIdentityMappingsByPseudonymID(ctx, pseudonymID)
	if err != nil {
		return "", fmt.Errorf("failed to get identity mappings: %w", err)
	}
	if len(mappings) == 0 {
		return "", fmt.Errorf("no identity mappings found for pseudonym")
	}

	// Try to decrypt each mapping until we find one that works
	for _, mapping := range mappings {
		decrypted, _, err := dao.ibeSystem.DecryptIdentity(mapping.EncryptedRealIdentity, keyData)
		if err == nil {
			// Parse fingerprint from mapping (format: "fingerprint:pseudonymID")
			mappingParts := strings.Split(decrypted, ":")
			if len(mappingParts) == 2 {
				return mappingParts[0], nil
			}
		}
	}

	return "", fmt.Errorf("failed to decrypt any identity mapping with provided key")
}

// DeactivatePseudonym deactivates a pseudonym using role-based access control
func (dao *PseudonymDAO) DeactivatePseudonym(ctx context.Context, pseudonymID string, userID int64, activePseudonymID, roleName, scope string) error {
	// Verify that the user owns this pseudonym using the active pseudonym
	ownsPseudonym, err := dao.VerifyPseudonymOwnership(ctx, pseudonymID, userID, activePseudonymID, roleName, scope)
	if err != nil {
		return fmt.Errorf("failed to verify pseudonym ownership: %w", err)
	}

	if !ownsPseudonym {
		return fmt.Errorf("user does not own this pseudonym")
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

// SearchPseudonyms searches for pseudonyms by display name, slug, or ID
func (dao *PseudonymDAO) SearchPseudonyms(ctx context.Context, query string, page, limit int) ([]*models.Pseudonym, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Validate and set defaults for pagination
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 25
	}

	// Build database query with WHERE clauses for better performance
	queryBuilder := models.Pseudonyms.Query(
		sm.Where(psql.Group(psql.Or(
			models.PseudonymColumns.DisplayName.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.Slug.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.PseudonymID.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
		))),
		sm.Limit(limit),
		sm.Offset((page-1)*limit),
	)

	// Execute query with pagination
	pseudonyms, err := queryBuilder.All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to search pseudonyms: %w", err)
	}

	log.Info().Int("result_count", len(pseudonyms)).Str("query", query).Int("page", page).Int("limit", limit).Msg("Search: Returning paginated results")

	return pseudonyms, nil
}

// SearchPseudonymsPublic searches for active pseudonyms by display name, slug, or ID (public endpoint)
func (dao *PseudonymDAO) SearchPseudonymsPublic(ctx context.Context, query string, page, limit int) ([]*models.Pseudonym, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Validate and set defaults for pagination
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 25
	}

	// Build database query with WHERE clauses for better performance
	queryBuilder := models.Pseudonyms.Query(
		sm.Where(models.PseudonymColumns.IsActive.EQ(psql.Arg(true))),
		sm.Where(psql.Group(psql.Or(
			models.PseudonymColumns.DisplayName.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.Slug.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.PseudonymID.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
		))),
		sm.Limit(limit),
		sm.Offset((page-1)*limit),
	)

	// Execute query with pagination
	pseudonyms, err := queryBuilder.All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to search pseudonyms: %w", err)
	}

	log.Info().Int("result_count", len(pseudonyms)).Str("query", query).Int("page", page).Int("limit", limit).Msg("Public Search: Returning paginated results")

	return pseudonyms, nil
}

// CountSearchPseudonyms counts the total number of pseudonyms matching the search criteria
func (dao *PseudonymDAO) CountSearchPseudonyms(ctx context.Context, query string) (int64, error) {
	if query == "" {
		return 0, fmt.Errorf("search query is required")
	}

	// Build database query with WHERE clauses for better performance
	queryBuilder := models.Pseudonyms.Query(
		sm.Where(psql.Group(psql.Or(
			models.PseudonymColumns.DisplayName.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.Slug.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.PseudonymID.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
		))),
	)

	// Execute count query
	pseudonyms, err := queryBuilder.All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count pseudonyms: %w", err)
	}

	total := int64(len(pseudonyms))
	log.Info().Int64("total_matching_pseudonyms", total).Str("query", query).Msg("Search: Count completed")

	return total, nil
}

// CountSearchPseudonymsPublic counts the total number of active pseudonyms matching the search criteria for public search
func (dao *PseudonymDAO) CountSearchPseudonymsPublic(ctx context.Context, query string) (int64, error) {
	if query == "" {
		return 0, fmt.Errorf("search query is required")
	}

	// Build database query with WHERE clauses for better performance
	queryBuilder := models.Pseudonyms.Query(
		sm.Where(models.PseudonymColumns.IsActive.EQ(psql.Arg(true))),
		sm.Where(psql.Group(psql.Or(
			models.PseudonymColumns.DisplayName.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.Slug.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
			models.PseudonymColumns.PseudonymID.ILike(psql.Arg("%"+strings.ToLower(query)+"%")),
		))),
	)

	// Execute count query
	pseudonyms, err := queryBuilder.All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count pseudonyms: %w", err)
	}

	total := int64(len(pseudonyms))
	log.Info().Int64("total_matching_pseudonyms", total).Str("query", query).Msg("Public Search: Count completed")

	return total, nil
}
