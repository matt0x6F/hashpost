package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// RoleKeyService handles role key operations with IBE integration
type RoleKeyService struct {
	roleKeyDAO   *dao.RoleKeyDAO
	userDAO      *dao.UserDAO
	pseudonymDAO dao.PseudonymDAOInterface
	ibeSystem    *ibe.IBESystem
}

// NewRoleKeyService creates a new RoleKeyService
func NewRoleKeyService(roleKeyDAO *dao.RoleKeyDAO, userDAO *dao.UserDAO, pseudonymDAO dao.PseudonymDAOInterface, ibeSystem *ibe.IBESystem) *RoleKeyService {
	return &RoleKeyService{
		roleKeyDAO:   roleKeyDAO,
		userDAO:      userDAO,
		pseudonymDAO: pseudonymDAO,
		ibeSystem:    ibeSystem,
	}
}

// GetKeyForOperation retrieves and validates a role key for a specific operation
func (s *RoleKeyService) GetKeyForOperation(ctx context.Context, pseudonymID, scope, operation string) ([]byte, error) {
	// Get the role key from the database using the pseudonym ID
	roleKey, err := s.roleKeyDAO.GetRoleKey(ctx, pseudonymID, scope, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key: %w", err)
	}

	// Validate that the key has the required capability
	hasCapability, err := s.roleKeyDAO.ValidateKeyCapability(ctx, pseudonymID, scope, operation, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key for pseudonym=%s scope=%s does not have capability=%s", pseudonymID, scope, operation)
	}

	return roleKey.KeyData, nil
}

// GenerateAndStoreKey generates a new IBE key and stores it in the database
func (s *RoleKeyService) GenerateAndStoreKey(ctx context.Context, pseudonymID, scope string, capabilities []string, expiresAt time.Time, createdByPseudonymID string) error {
	// Generate IBE key for the pseudonym and scope
	ibeKey := s.ibeSystem.GenerateRoleKey(pseudonymID, scope, expiresAt)

	// Store the key in the database
	_, err := s.roleKeyDAO.CreateRoleKey(ctx, pseudonymID, scope, ibeKey, capabilities, expiresAt, createdByPseudonymID, "", nil)
	if err != nil {
		return fmt.Errorf("failed to store role key: %w", err)
	}

	return nil
}

// ValidateUserAccess validates if a user can access a specific operation
func (s *RoleKeyService) ValidateUserAccess(ctx context.Context, userID int64, pseudonymID, scope, operation string) (bool, error) {
	// Fetch user from DB
	user, err := s.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	// Check if the key exists and has the required capability using the pseudonym ID
	hasCapability, err := s.roleKeyDAO.ValidateKeyCapability(ctx, pseudonymID, scope, operation, nil)
	if err != nil {
		return false, fmt.Errorf("failed to validate key capability: %w", err)
	}
	if !hasCapability {
		return false, nil
	}

	return true, nil
}

// EnsureDefaultKeys ensures that default role keys exist in the database
func (s *RoleKeyService) EnsureDefaultKeys(ctx context.Context, pseudonymID string, userRoles []string) error {
	return s.roleKeyDAO.EnsureDefaultKeys(ctx, s.ibeSystem, pseudonymID, userRoles)
}

// ListUserKeys lists all role keys that a user can access
func (s *RoleKeyService) ListUserKeys(ctx context.Context, userID int64, activePseudonymID string) ([]*models.RoleKey, error) {
	// Fetch user from DB
	user, err := s.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Validate that the active pseudonym belongs to the user
	// This ensures users can only access role keys for pseudonyms they own
	ownsPseudonym, err := s.pseudonymDAO.VerifyPseudonymOwnership(ctx, activePseudonymID, userID, activePseudonymID, "", "")
	if err != nil {
		log.Warn().
			Err(err).
			Int64("user_id", userID).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("Failed to verify pseudonym ownership")
		return []*models.RoleKey{}, nil
	}

	if !ownsPseudonym {
		log.Warn().
			Int64("user_id", userID).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("User does not own the active pseudonym")
		return []*models.RoleKey{}, nil
	}

	// Check if user has correlation capability through their active pseudonym's role keys
	hasCorrelationCapability, err := s.roleKeyDAO.ValidateKeyCapability(ctx, activePseudonymID, "correlation", constants.CapabilityAccessOwnPseudonyms, nil)
	if err != nil {
		log.Warn().
			Err(err).
			Int64("user_id", userID).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("Failed to validate correlation capability")
		return []*models.RoleKey{}, nil
	}

	if !hasCorrelationCapability {
		log.Warn().
			Int64("user_id", userID).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("User lacks correlation capability for role key access")
		return []*models.RoleKey{}, nil
	}

	// Get user's pseudonyms using their active pseudonym's role keys for correlation
	// First, get the role keys for the active pseudonym to determine the user's role
	roleKeys, err := s.roleKeyDAO.ListRoleKeysByPseudonym(ctx, activePseudonymID)
	if err != nil {
		log.Warn().
			Err(err).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("Failed to get role keys for active pseudonym")
		return []*models.RoleKey{}, nil
	}

	// Determine the user's primary role from their role keys
	// Look for correlation scope role keys to determine the user's role
	var userRole string
	for _, roleKey := range roleKeys {
		if roleKey.Scope == "correlation" {
			// Parse capabilities to check if this is a user-level role
			capabilitiesBytes, err := roleKey.Capabilities.Value()
			if err != nil {
				continue
			}
			var capabilities []string
			if err := json.Unmarshal(capabilitiesBytes.([]byte), &capabilities); err != nil {
				continue
			}

			// Check if this role key has access to own pseudonyms
			for _, capability := range capabilities {
				if capability == constants.CapabilityAccessOwnPseudonyms {
					userRole = roleKey.RoleName
					break
				}
			}
			if userRole != "" {
				break
			}
		}
	}

	// If no role found, default to "user"
	if userRole == "" {
		userRole = "user"
	}

	// Get user's pseudonyms using the active pseudonym for authorization
	pseudonyms, err := s.pseudonymDAO.GetPseudonymsByUserID(ctx, userID, activePseudonymID, userRole, "correlation")
	if err != nil {
		log.Warn().
			Err(err).
			Int64("user_id", userID).
			Str("active_pseudonym_id", activePseudonymID).
			Str("user_role", userRole).
			Msg("Failed to get user's pseudonyms for role key filtering")
		return []*models.RoleKey{}, nil
	}

	// Collect all role keys for the user's pseudonyms
	var allRoleKeys []*models.RoleKey
	for _, pseudonym := range pseudonyms {
		roleKeys, err := s.roleKeyDAO.ListRoleKeysByPseudonym(ctx, pseudonym.PseudonymID)
		if err != nil {
			log.Warn().
				Err(err).
				Str("pseudonym_id", pseudonym.PseudonymID).
				Msg("Failed to get role keys for pseudonym")
			continue
		}
		allRoleKeys = append(allRoleKeys, roleKeys...)
	}

	log.Info().
		Int64("user_id", userID).
		Str("active_pseudonym_id", activePseudonymID).
		Str("user_role", userRole).
		Int("pseudonym_count", len(pseudonyms)).
		Int("role_key_count", len(allRoleKeys)).
		Msg("Retrieved role keys for user's pseudonyms")

	return allRoleKeys, nil
}

// DeactivateKey deactivates a role key
func (s *RoleKeyService) DeactivateKey(ctx context.Context, keyID string) error {
	return s.roleKeyDAO.DeactivateRoleKey(ctx, keyID)
}

// GetKeyCapabilities returns the capabilities of a specific role key
func (s *RoleKeyService) GetKeyCapabilities(ctx context.Context, pseudonymID, scope string) ([]string, error) {
	roleKey, err := s.roleKeyDAO.GetRoleKey(ctx, pseudonymID, scope, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key: %w", err)
	}

	// Parse capabilities from JSON
	capabilitiesBytes, err := roleKey.Capabilities.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get capabilities value: %w", err)
	}

	var capabilities []string
	if err := json.Unmarshal(capabilitiesBytes.([]byte), &capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return capabilities, nil
}

// ValidateKeyForUser validates if a specific key can be used by a user for an operation
func (s *RoleKeyService) ValidateKeyForUser(ctx context.Context, userID int64, roleName, scope, operation string) (bool, error) {
	return s.ValidateUserAccess(ctx, userID, roleName, scope, operation)
}
