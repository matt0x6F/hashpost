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
func (s *RoleKeyService) GetKeyForOperation(ctx context.Context, roleName, scope, operation string) ([]byte, error) {
	// Get the role key from the database
	roleKey, err := s.roleKeyDAO.GetRoleKey(ctx, roleName, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get role key: %w", err)
	}

	// Validate that the key has the required capability
	hasCapability, err := s.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to validate key capability: %w", err)
	}

	if !hasCapability {
		return nil, fmt.Errorf("role key for role=%s scope=%s does not have capability=%s", roleName, scope, operation)
	}

	return roleKey.KeyData, nil
}

// GenerateAndStoreKey generates a new IBE key and stores it in the database
func (s *RoleKeyService) GenerateAndStoreKey(ctx context.Context, roleName, scope string, capabilities []string, expiresAt time.Time, createdBy int64) error {
	// Generate IBE key for the role and scope
	ibeKey := s.ibeSystem.GenerateRoleKey(roleName, scope, expiresAt)

	// Store the key in the database
	_, err := s.roleKeyDAO.CreateRoleKey(ctx, roleName, scope, ibeKey, capabilities, expiresAt, createdBy)
	if err != nil {
		return fmt.Errorf("failed to store role key: %w", err)
	}

	return nil
}

// Helper to extract roles from user.Roles
func extractUserRoles(user *models.User) []string {
	var roles []string
	if user.Roles.Valid {
		var raw json.RawMessage
		if err := user.Roles.Scan(&raw); err == nil {
			_ = json.Unmarshal(raw, &roles)
		}
	}
	return roles
}

// ValidateUserAccess validates if a user can access a specific operation
func (s *RoleKeyService) ValidateUserAccess(ctx context.Context, userID int64, roleName, scope, operation string) (bool, error) {
	// Fetch user from DB
	user, err := s.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	// Check if user has the required role
	userRoles := extractUserRoles(user)
	hasRole := false
	for _, r := range userRoles {
		if r == roleName {
			hasRole = true
			break
		}
	}
	if !hasRole {
		return false, nil
	}

	// Check if the key exists and has the required capability
	hasCapability, err := s.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, operation)
	if err != nil {
		return false, fmt.Errorf("failed to validate key capability: %w", err)
	}
	if !hasCapability {
		return false, nil
	}

	return true, nil
}

// EnsureDefaultKeys ensures that default role keys exist in the database
func (s *RoleKeyService) EnsureDefaultKeys(ctx context.Context, createdBy int64) error {
	return s.roleKeyDAO.EnsureDefaultKeys(ctx, s.ibeSystem, createdBy)
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
	userRoles := extractUserRoles(user)

	// Get all keys and filter by user roles
	allKeys, err := s.roleKeyDAO.ListRoleKeys(ctx)
	if err != nil {
		return nil, err
	}
	var filtered []*models.RoleKey
	for _, key := range allKeys {
		for _, r := range userRoles {
			if key.RoleName == r {
				filtered = append(filtered, key)
				break
			}
		}
	}
	return filtered, nil
}

// DeactivateKey deactivates a role key
func (s *RoleKeyService) DeactivateKey(ctx context.Context, keyID string) error {
	return s.roleKeyDAO.DeactivateRoleKey(ctx, keyID)
}

// GetKeyCapabilities returns the capabilities of a specific role key
func (s *RoleKeyService) GetKeyCapabilities(ctx context.Context, roleName, scope string) ([]string, error) {
	roleKey, err := s.roleKeyDAO.GetRoleKey(ctx, roleName, scope)
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
