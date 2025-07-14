package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// SetupRoles creates the necessary role keys for all admin roles
func SetupRoles() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize IBE system
	ibeSystem := ibe.NewIBESystemFromEnv()

	// Create role key DAO
	roleKeyDAO := dao.NewRoleKeyDAO(db)

	// Find a user to use as the creator for role keys
	// First try to find any existing user
	userDAO := dao.NewUserDAO(db)
	ctx := context.Background()

	var creatorUserID int64
	var createdSystemUser bool
	var systemUserID int64

	// Try to find any existing user
	users, err := userDAO.ListUsers(ctx, 1, 0) // Get first user
	if err != nil || len(users) == 0 {
		// No users exist, create a temporary system user for role key creation
		log.Info().Msg("No users found, creating temporary system user for role key creation")

		// Create a temporary system user
		systemPasswordHash := hashPassword("system_user_temp_password")
		systemUser, err := userDAO.CreateUser(ctx, "system@hashpost.local", systemPasswordHash)
		if err != nil {
			return fmt.Errorf("failed to create system user for role key creation: %w", err)
		}

		creatorUserID = systemUser.UserID
		systemUserID = systemUser.UserID
		createdSystemUser = true
		log.Info().Int64("system_user_id", creatorUserID).Msg("Created temporary system user for role key creation")
	} else {
		// Use the first existing user
		creatorUserID = users[0].UserID
		log.Info().Int64("creator_user_id", creatorUserID).Msg("Using existing user for role key creation")
	}

	// Define all roles and their capabilities using constants
	allRoles := constants.GetRoleDefinitions()

	// Create role keys for each admin role
	for _, adminRole := range allRoles {
		log.Info().Str("role", adminRole.RoleName).Msg("Creating role keys")

		for _, scope := range adminRole.Scopes {
			capabilities := adminRole.Capabilities[scope]

			// Check if role key already exists
			existingKey, err := roleKeyDAO.GetRoleKey(ctx, adminRole.RoleName, scope)
			if err == nil && existingKey != nil {
				log.Info().Str("role", adminRole.RoleName).Str("scope", scope).Msg("Role key already exists, skipping")
				continue
			}

			// Create the role key
			expiresAt := time.Now().AddDate(1, 0, 0) // Expire in 1 year
			keyData := ibeSystem.GenerateTestRoleKey(adminRole.RoleName, scope)

			_, err = roleKeyDAO.CreateRoleKey(ctx, adminRole.RoleName, scope, keyData, capabilities, expiresAt, creatorUserID)
			if err != nil {
				log.Error().Str("role", adminRole.RoleName).Str("scope", scope).Err(err).Msg("Failed to create role key")
				continue
			}

			log.Info().Str("role", adminRole.RoleName).Str("scope", scope).Strs("capabilities", capabilities).Msg("Role key created successfully")
		}
	}

	log.Info().Msg("Role key setup completed successfully")
	fmt.Println("✅ Role keys created successfully for all roles!")
	fmt.Println("   - user: authentication, self_correlation")
	fmt.Println("   - moderator: authentication, self_correlation, correlation")
	fmt.Println("   - subforum_owner: authentication, self_correlation, correlation")
	fmt.Println("   - platform_admin: authentication, self_correlation, correlation")
	fmt.Println("   - trust_safety: authentication, self_correlation, correlation")
	fmt.Println("   - legal_team: authentication, self_correlation, correlation")

	// Clean up temporary system user if we created one
	if createdSystemUser {
		log.Info().Int64("system_user_id", systemUserID).Msg("Cleaning up temporary system user")
		if err := userDAO.DeleteUser(ctx, systemUserID); err != nil {
			log.Error().Err(err).Int64("system_user_id", systemUserID).Msg("Failed to delete temporary system user")
		} else {
			log.Info().Int64("system_user_id", systemUserID).Msg("Temporary system user cleaned up successfully")
		}
	}

	return nil
}

// ListRoles lists all available roles and their capabilities
func ListRoles() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get all role definitions directly using the model
	ctx := context.Background()
	roleDefs, err := models.RoleDefinitions.Query().All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to list role definitions: %w", err)
	}

	fmt.Println("📋 Available Roles:")
	fmt.Println("===================")

	for _, roleDef := range roleDefs {
		fmt.Printf("\n🔸 %s (%s)\n", roleDef.DisplayName, roleDef.RoleName)
		if roleDef.Description.Valid {
			fmt.Printf("   Description: %s\n", roleDef.Description.V)
		}

		// Parse capabilities
		var capabilities []string
		capBytes, err := roleDef.Capabilities.Value()
		if err == nil {
			json.Unmarshal(capBytes.([]byte), &capabilities)
		}

		if len(capabilities) > 0 {
			fmt.Printf("   Capabilities: %s\n", strings.Join(capabilities, ", "))
		}

		if roleDef.CorrelationAccess.Valid {
			fmt.Printf("   Correlation Access: %s\n", roleDef.CorrelationAccess.V)
		}

		if roleDef.Scope.Valid {
			fmt.Printf("   Scope: %s\n", roleDef.Scope.V)
		}

		if roleDef.TimeWindow.Valid {
			fmt.Printf("   Time Window: %s\n", roleDef.TimeWindow.V)
		}
	}

	return nil
}

// ListRoleKeys lists all active role keys
func ListRoleKeys() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create role key DAO
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	ctx := context.Background()

	// Get all active role keys
	roleKeys, err := roleKeyDAO.ListRoleKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to list role keys: %w", err)
	}

	fmt.Println("🔑 Active Role Keys:")
	fmt.Println("====================")

	if len(roleKeys) == 0 {
		fmt.Println("No active role keys found.")
		return nil
	}

	for _, key := range roleKeys {
		fmt.Printf("\n🔸 Role: %s\n", key.RoleName)
		fmt.Printf("   Scope: %s\n", key.Scope)
		fmt.Printf("   Key ID: %s\n", key.KeyID)
		fmt.Printf("   Version: %d\n", key.KeyVersion)
		fmt.Printf("   Expires: %s\n", key.ExpiresAt.Format("2006-01-02 15:04:05"))

		// Parse capabilities
		var capabilities []string
		capBytes, err := key.Capabilities.Value()
		if err == nil {
			json.Unmarshal(capBytes.([]byte), &capabilities)
		}

		if len(capabilities) > 0 {
			fmt.Printf("   Capabilities: %s\n", strings.Join(capabilities, ", "))
		}

		if key.CreatedAt.Valid {
			fmt.Printf("   Created: %s\n", key.CreatedAt.V.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}

// RotateRoleKeys rotates role keys for a specific role or all roles
func RotateRoleKeys(roleName string, force bool) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize IBE system
	ibeSystem := ibe.NewIBESystemFromEnv()

	// Create role key DAO
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	userDAO := dao.NewUserDAO(db)
	ctx := context.Background()

	// Find a user to use as the creator for role keys
	users, err := userDAO.ListUsers(ctx, 1, 0)
	if err != nil || len(users) == 0 {
		return fmt.Errorf("no users found to create role keys")
	}
	creatorUserID := users[0].UserID

	// Define all roles and their capabilities (same as SetupRoles)
	allRoles := constants.GetRoleDefinitions()

	// Filter roles if specific role is provided
	var rolesToRotate []constants.RoleDefinition

	if roleName != "" {
		for _, role := range allRoles {
			if role.RoleName == roleName {
				rolesToRotate = append(rolesToRotate, role)
				break
			}
		}
		if len(rolesToRotate) == 0 {
			return fmt.Errorf("role '%s' not found", roleName)
		}
	} else {
		rolesToRotate = allRoles
	}

	// Rotate keys for selected roles
	for _, role := range rolesToRotate {
		log.Info().Str("role", role.RoleName).Msg("Rotating role keys")

		for _, scope := range role.Scopes {
			capabilities := role.Capabilities[scope]

			// Check if role key exists
			existingKey, err := roleKeyDAO.GetRoleKey(ctx, role.RoleName, scope)
			if err != nil || existingKey == nil {
				log.Info().Str("role", role.RoleName).Str("scope", scope).Msg("Role key does not exist, creating new one")
			} else if !force {
				log.Info().Str("role", role.RoleName).Str("scope", scope).Msg("Role key exists and force=false, skipping")
				continue
			} else {
				// Deactivate existing key
				err = roleKeyDAO.DeactivateRoleKey(ctx, existingKey.KeyID.String())
				if err != nil {
					log.Error().Str("role", role.RoleName).Str("scope", scope).Err(err).Msg("Failed to deactivate existing role key")
					continue
				}
				log.Info().Str("role", role.RoleName).Str("scope", scope).Msg("Deactivated existing role key")
			}

			// Create new role key
			expiresAt := time.Now().AddDate(1, 0, 0) // Expire in 1 year
			keyData := ibeSystem.GenerateTestRoleKey(role.RoleName, scope)

			_, err = roleKeyDAO.CreateRoleKey(ctx, role.RoleName, scope, keyData, capabilities, expiresAt, creatorUserID)
			if err != nil {
				log.Error().Str("role", role.RoleName).Str("scope", scope).Err(err).Msg("Failed to create new role key")
				continue
			}

			log.Info().Str("role", role.RoleName).Str("scope", scope).Strs("capabilities", capabilities).Msg("New role key created successfully")
		}
	}

	if roleName != "" {
		fmt.Printf("✅ Role keys rotated successfully for role: %s\n", roleName)
	} else {
		fmt.Println("✅ Role keys rotated successfully for all roles!")
	}

	return nil
}
