package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
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

	// Define all roles and their capabilities
	allRoles := []struct {
		roleName     string
		scopes       []string
		capabilities map[string][]string
	}{
		{
			roleName: "user",
			scopes:   []string{"authentication", "self_correlation"},
			capabilities: map[string][]string{
				"authentication":   {"access_own_pseudonyms", "login", "session_management"},
				"self_correlation": {"verify_own_pseudonym_ownership", "manage_own_profile"},
			},
		},
		{
			roleName: "moderator",
			scopes:   []string{"authentication", "self_correlation", "correlation"},
			capabilities: map[string][]string{
				"authentication":   {"access_own_pseudonyms", "login", "session_management"},
				"self_correlation": {"verify_own_pseudonym_ownership", "manage_own_profile"},
				"correlation":      {"access_subforum_pseudonyms", "correlate_fingerprints", "moderate_content"},
			},
		},
		{
			roleName: "subforum_owner",
			scopes:   []string{"authentication", "self_correlation", "correlation"},
			capabilities: map[string][]string{
				"authentication":   {"access_own_pseudonyms", "login", "session_management"},
				"self_correlation": {"verify_own_pseudonym_ownership", "manage_own_profile"},
				"correlation":      {"access_subforum_pseudonyms", "correlate_fingerprints", "moderate_content", "manage_moderators"},
			},
		},
		{
			roleName: "platform_admin",
			scopes:   []string{"authentication", "self_correlation", "correlation"},
			capabilities: map[string][]string{
				"authentication":   {"access_own_pseudonyms", "login", "session_management"},
				"self_correlation": {"verify_own_pseudonym_ownership", "manage_own_profile"},
				"correlation":      {"access_all_pseudonyms", "cross_user_correlation", "moderation", "compliance", "legal_requests"},
			},
		},
		{
			roleName: "trust_safety",
			scopes:   []string{"authentication", "self_correlation", "correlation"},
			capabilities: map[string][]string{
				"authentication":   {"access_own_pseudonyms", "login", "session_management"},
				"self_correlation": {"verify_own_pseudonym_ownership", "manage_own_profile"},
				"correlation":      {"access_all_pseudonyms", "cross_user_correlation", "moderation", "compliance"},
			},
		},
		{
			roleName: "legal_team",
			scopes:   []string{"authentication", "self_correlation", "correlation"},
			capabilities: map[string][]string{
				"authentication":   {"access_own_pseudonyms", "login", "session_management"},
				"self_correlation": {"verify_own_pseudonym_ownership", "manage_own_profile"},
				"correlation":      {"access_all_pseudonyms", "cross_user_correlation", "compliance", "legal_requests"},
			},
		},
	}

	// Create role keys for each admin role
	for _, adminRole := range allRoles {
		log.Info().Str("role", adminRole.roleName).Msg("Creating role keys")

		for _, scope := range adminRole.scopes {
			capabilities := adminRole.capabilities[scope]

			// Check if role key already exists
			existingKey, err := roleKeyDAO.GetRoleKey(ctx, adminRole.roleName, scope)
			if err == nil && existingKey != nil {
				log.Info().Str("role", adminRole.roleName).Str("scope", scope).Msg("Role key already exists, skipping")
				continue
			}

			// Create the role key
			expiresAt := time.Now().AddDate(1, 0, 0) // Expire in 1 year
			keyData := ibeSystem.GenerateTestRoleKey(adminRole.roleName, scope)

			_, err = roleKeyDAO.CreateRoleKey(ctx, adminRole.roleName, scope, keyData, capabilities, expiresAt, creatorUserID)
			if err != nil {
				log.Error().Str("role", adminRole.roleName).Str("scope", scope).Err(err).Msg("Failed to create role key")
				continue
			}

			log.Info().Str("role", adminRole.roleName).Str("scope", scope).Strs("capabilities", capabilities).Msg("Role key created successfully")
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
