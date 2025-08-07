package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
)

// RoleKeyService handles role key operations with IBE integration
type RoleKeyService struct {
	roleKeyDAO *dao.RoleKeyDAO
	userDAO    *dao.UserDAO
	ibeSystem  *ibe.IBESystem
}

// NewRoleKeyService creates a new RoleKeyService
func NewRoleKeyService(roleKeyDAO *dao.RoleKeyDAO, userDAO *dao.UserDAO, ibeSystem *ibe.IBESystem) *RoleKeyService {
	return &RoleKeyService{
		roleKeyDAO: roleKeyDAO,
		userDAO:    userDAO,
		ibeSystem:  ibeSystem,
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
func (s *RoleKeyService) ListUserKeys(ctx context.Context, userID int64) ([]*models.RoleKey, error) {
	// Fetch user from DB
	user, err := s.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get all keys - since users don't have roles anymore, we'll return all keys
	// In a more sophisticated implementation, we would filter based on pseudonym access
	allKeys, err := s.roleKeyDAO.ListRoleKeys(ctx)
	if err != nil {
		return nil, err
	}

	// For now, return all keys since users don't have direct roles
	// In the future, this could be filtered based on pseudonym access patterns
	return allKeys, nil
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
