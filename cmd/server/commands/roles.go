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
	"github.com/spf13/cobra"
)

// NewRolesCommands returns all roles-related commands
func NewRolesCommands(cfg *config.Config) []*cobra.Command {
	// Create the roles subcommand
	rolesCmd := &cobra.Command{
		Use:   "roles",
		Short: "Manage roles and role keys",
		Long:  "Commands for managing roles, role keys, and related operations.",
	}

	// Add 'setup' subcommand under 'roles'
	setupRolesCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup role keys for all roles",
		Long:  "Create the necessary role keys for all roles: user, moderator, subforum_owner, platform_admin, trust_safety, and legal_team",
		Run: func(cmd *cobra.Command, args []string) {
			if err := SetupRoles(cfg); err != nil {
				log.Fatal().Err(err).Msg("Failed to setup roles")
			}
		},
	}
	rolesCmd.AddCommand(setupRolesCmd)

	// Add 'list' subcommand under 'roles'
	listRolesCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available roles and their capabilities",
		Long:  "Display all roles defined in the system with their capabilities, correlation access, scope, and time windows",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE
			if err := ListRoles(); err != nil {
				log.Fatal().Err(err).Msg("Failed to list roles")
			}
		},
	}
	rolesCmd.AddCommand(listRolesCmd)

	// Add 'keys' subcommand under 'roles'
	listRoleKeysCmd := &cobra.Command{
		Use:   "keys",
		Short: "List all active role keys",
		Long:  "Display all active role keys with their capabilities, expiration dates, and metadata",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE
			if err := ListRoleKeys(); err != nil {
				log.Fatal().Err(err).Msg("Failed to list role keys")
			}
		},
	}
	rolesCmd.AddCommand(listRoleKeysCmd)

	// Add 'rotate' subcommand under 'roles'
	rotateRoleKeysCmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate role keys for security",
		Long:  "Rotate role keys for a specific role or all roles. This deactivates existing keys and creates new ones.",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE
			roleName, _ := cmd.Flags().GetString("role")
			force, _ := cmd.Flags().GetBool("force")
			if err := RotateRoleKeys(roleName, force, cfg); err != nil {
				log.Fatal().Err(err).Msg("Failed to rotate role keys")
			}
		},
	}
	rotateRoleKeysCmd.Flags().String("role", "", "Specific role to rotate keys for (optional, rotates all roles if not specified)")
	rotateRoleKeysCmd.Flags().Bool("force", false, "Force rotation even if keys already exist")
	rolesCmd.AddCommand(rotateRoleKeysCmd)

	// Add 'audit' subcommand under 'roles'
	auditRoleKeysCmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit and rectify role keys and identity mappings",
		Long:  "Check existing role keys and identity mappings, then create what's missing. This ensures all users have the required keys for authentication, self-correlation, and messaging.",
		Run: func(cmd *cobra.Command, args []string) {
			userID, _ := cmd.Flags().GetInt64("user-id")
			allUsers, _ := cmd.Flags().GetBool("all-users")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if err := AuditRoleKeys(userID, allUsers, dryRun, cfg); err != nil {
				log.Fatal().Err(err).Msg("Failed to audit role keys")
			}
		},
	}
	auditRoleKeysCmd.Flags().Int64("user-id", 0, "Specific user ID to audit (optional)")
	auditRoleKeysCmd.Flags().Bool("all-users", false, "Audit all users in the system")
	auditRoleKeysCmd.Flags().Bool("dry-run", false, "Show what would be created without making changes")
	rolesCmd.AddCommand(auditRoleKeysCmd)

	// Add 're-encrypt' subcommand under 'roles'
	reEncryptCmd := &cobra.Command{
		Use:   "re-encrypt",
		Short: "Re-encrypt existing identity mappings with new keys",
		Long:  "Re-encrypt existing identity mappings that were encrypted with old key generation methods",
		Run: func(cmd *cobra.Command, args []string) {
			allUsers, _ := cmd.Flags().GetBool("all-users")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if err := ReEncryptIdentityMappings(allUsers, dryRun, cfg); err != nil {
				log.Fatal().Err(err).Msg("Failed to re-encrypt identity mappings")
			}
		},
	}
	reEncryptCmd.Flags().Bool("all-users", false, "Re-encrypt mappings for all users")
	reEncryptCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	rolesCmd.AddCommand(reEncryptCmd)

	return []*cobra.Command{rolesCmd}
}

// SetupRoles creates the necessary role keys for all admin roles
func SetupRoles(cfg *config.Config) error {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize IBE system using configuration instead of hardcoded defaults
	ibeSystem, err := ibe.NewIBESystemFromConfig(cfg.IBE.DomainKeysDir, cfg.IBE.KeyVersion, cfg.IBE.Salt)
	if err != nil {
		return fmt.Errorf("failed to initialize IBE system: %w", err)
	}

	// Create role key DAO
	roleKeyDAO := dao.NewRoleKeyDAO(db, nil)

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

			// Check if role key already exists for this role
			existingKey, err := roleKeyDAO.GetRoleKey(ctx, adminRole.RoleName, scope, nil)
			if err == nil && existingKey != nil {
				log.Info().Str("role", adminRole.RoleName).Str("scope", scope).Msg("Role key already exists, skipping")
				continue
			}

			// Create the role key
			expiresAt := time.Now().AddDate(1, 0, 0) // Expire in 1 year
			keyData := ibeSystem.GenerateTestRoleKey(adminRole.RoleName, scope)

			_, err = roleKeyDAO.CreateRoleKey(ctx, adminRole.RoleName, scope, keyData, capabilities, expiresAt, adminRole.RoleName, "", nil)
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
	roleKeyDAO := dao.NewRoleKeyDAO(db, nil)
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
func RotateRoleKeys(roleName string, force bool, cfg *config.Config) error {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize IBE system using configuration instead of hardcoded defaults
	ibeSystem, err := ibe.NewIBESystemFromConfig(cfg.IBE.DomainKeysDir, cfg.IBE.KeyVersion, cfg.IBE.Salt)
	if err != nil {
		return fmt.Errorf("failed to initialize IBE system: %w", err)
	}

	// Create role key DAO
	roleKeyDAO := dao.NewRoleKeyDAO(db, nil)
	ctx := context.Background()

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

			// Check if role key exists for platform admin
			existingKey, err := roleKeyDAO.GetRoleKey(ctx, constants.RolePlatformAdmin, scope, nil)
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
			keyData := ibeSystem.GenerateRoleKey(role.RoleName, scope, expiresAt)

			_, err = roleKeyDAO.CreateRoleKey(ctx, constants.RolePlatformAdmin, scope, keyData, capabilities, expiresAt, constants.RolePlatformAdmin, "", nil)
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

// AuditRoleKeys audits and rectifies role keys and identity mappings for users
func AuditRoleKeys(userID int64, allUsers bool, dryRun bool, cfg *config.Config) error {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize IBE system
	ibeSystem, err := ibe.NewIBESystemFromConfig(cfg.IBE.DomainKeysDir, cfg.IBE.KeyVersion, cfg.IBE.Salt)
	if err != nil {
		return fmt.Errorf("failed to initialize IBE system: %w", err)
	}

	// Create DAOs
	userDAO := dao.NewUserDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db, nil)

	ctx := context.Background()

	var usersToAudit []*models.User

	if userID > 0 {
		// Audit specific user
		user, err := userDAO.GetUserByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user %d: %w", userID, err)
		}
		if user == nil {
			return fmt.Errorf("user %d not found", userID)
		}
		usersToAudit = []*models.User{user}
		log.Info().Int64("user_id", userID).Msg("Auditing specific user")
	} else if allUsers {
		// Audit all users
		users, err := userDAO.ListUsers(ctx, 1000, 0) // Get up to 1000 users
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		usersToAudit = users
		log.Info().Int("user_count", len(users)).Msg("Auditing all users")
	} else {
		return fmt.Errorf("must specify either --user-id or --all-users")
	}

	if dryRun {
		log.Info().Msg("DRY RUN MODE - No changes will be made")
	}

	// Track what we're going to create
	type auditResult struct {
		userID      int64
		pseudonymID string
		missingKeys []string
		missingMaps []string
		createdKeys []string
		createdMaps []string
	}

	var results []auditResult

	for _, user := range usersToAudit {
		log.Info().Int64("user_id", user.UserID).Str("email", user.Email).Msg("Auditing user")

		// Get user's pseudonyms through identity mappings
		identityMappings, err := models.IdentityMappings.Query(
			models.SelectWhere.IdentityMappings.UserID.EQ(user.UserID),
			models.SelectWhere.IdentityMappings.IsActive.EQ(true),
		).All(ctx, db)
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get identity mappings, skipping user")
			continue
		}

		if len(identityMappings) == 0 {
			log.Warn().Int64("user_id", user.UserID).Msg("User has no identity mappings, skipping")
			continue
		}

		// Get unique pseudonym IDs from identity mappings
		pseudonymIDs := make(map[string]bool)
		for _, mapping := range identityMappings {
			pseudonymIDs[mapping.PseudonymID] = true
		}

		// Get pseudonyms by their IDs
		var pseudonyms []*models.Pseudonym
		for pseudonymID := range pseudonymIDs {
			pseudonym, err := models.Pseudonyms.Query(
				models.SelectWhere.Pseudonyms.PseudonymID.EQ(pseudonymID),
				models.SelectWhere.Pseudonyms.IsActive.EQ(true),
			).One(ctx, db)
			if err != nil {
				log.Warn().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get pseudonym, skipping")
				continue
			}
			pseudonyms = append(pseudonyms, pseudonym)
		}

		if len(pseudonyms) == 0 {
			log.Warn().Int64("user_id", user.UserID).Msg("User has no active pseudonyms, skipping")
			continue
		}

		// Audit each pseudonym
		for _, pseudonym := range pseudonyms {
			result := auditResult{
				userID:      user.UserID,
				pseudonymID: pseudonym.PseudonymID,
			}

			log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Auditing pseudonym")

			// Determine user roles based on existing role keys
			userRoles := []string{"user"} // Default role
			existingRoleKeys, err := roleKeyDAO.ListRoleKeysByPseudonym(ctx, pseudonym.PseudonymID)
			if err == nil && len(existingRoleKeys) > 0 {
				roleSet := make(map[string]bool)
				for _, roleKey := range existingRoleKeys {
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

			log.Info().Strs("user_roles", userRoles).Msg("Determined user roles")

			// Check what keys should exist
			requiredKeys := []struct {
				scope  string
				desc   string
				reason string
			}{
				{constants.ScopeAuthentication, "authentication", "Required for all users to authenticate and access the system"},
				{constants.ScopeSelfCorrelation, "self-correlation", "Required for all users to link their own pseudonyms together"},
				{constants.ScopeMessaging, "messaging", "Required for all users to send and receive encrypted direct messages"},
			}

			// Add admin-specific keys
			for _, role := range userRoles {
				if role == constants.RolePlatformAdmin || role == constants.RoleTrustSafety || role == constants.RoleLegalTeam {
					requiredKeys = append(requiredKeys, struct {
						scope  string
						desc   string
						reason string
					}{
						constants.ScopeCorrelation, "correlation", fmt.Sprintf("Required for %s role to perform cross-user correlation and moderation", role),
					})
					break
				}
			}

			// Check each required key
			for _, requiredKey := range requiredKeys {
				existing, err := roleKeyDAO.GetRoleKey(ctx, pseudonym.PseudonymID, requiredKey.scope, nil)
				if err != nil || existing == nil {
					result.missingKeys = append(result.missingKeys, requiredKey.desc)

					if !dryRun {
						log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredKey.scope).Str("reason", requiredKey.reason).Msg("Creating missing role key")

						// Use EnsureDefaultKeys for authentication and self-correlation
						if requiredKey.scope == constants.ScopeAuthentication || requiredKey.scope == constants.ScopeSelfCorrelation {
							err = roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, pseudonym.PseudonymID, userRoles)
							if err != nil {
								log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredKey.scope).Msg("Failed to create default keys")
								continue
							}
						} else if requiredKey.scope == constants.ScopeMessaging {
							// Create messaging key using the key management service
							// We need to create a mock encryption service since we don't have one
							// For now, create the messaging key directly using the role key DAO
							expiresAt := time.Now().AddDate(1, 0, 0) // 1 year expiration

							// Generate messaging key data using IBE
							messagingKeyData := ibeSystem.GenerateMessagingKey("user", "messaging", time.Hour*24*365) // 1 year duration
							if messagingKeyData == nil {
								log.Error().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredKey.scope).Msg("Failed to generate messaging key data")
								continue
							}

							// Create the messaging role key
							_, err = roleKeyDAO.CreateRoleKey(
								ctx,
								"user",                   // roleName
								constants.ScopeMessaging, // scope
								messagingKeyData,         // keyData
								[]string{ // capabilities
									constants.CapabilitySendDirectMessages,
									constants.CapabilityReceiveDirectMessages,
									constants.CapabilityManageConversationKeys,
								},
								expiresAt,             // expiresAt
								pseudonym.PseudonymID, // createdByPseudonymID
								pseudonym.PseudonymID, // pseudonymID
								nil,                   // subforumID (global scope)
							)
							if err != nil {
								log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredKey.scope).Msg("Failed to create messaging role key")
								continue
							}
						}

						result.createdKeys = append(result.createdKeys, requiredKey.desc)
						log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredKey.scope).Msg("Created role key")
					} else {
						log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredKey.scope).Str("reason", requiredKey.reason).Msg("Would create missing role key")
					}
				} else {
					log.Debug().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredKey.scope).Msg("Role key already exists")
				}
			}

			// Check identity mappings
			requiredMappings := []struct {
				scope  string
				reason string
			}{
				{constants.ScopeAuthentication, "Required for all users to authenticate and access the system"},
				{constants.ScopeSelfCorrelation, "Required for all users to link their own pseudonyms together"},
			}

			// Add admin-specific mappings
			for _, role := range userRoles {
				if role == constants.RolePlatformAdmin || role == constants.RoleTrustSafety || role == constants.RoleLegalTeam {
					requiredMappings = append(requiredMappings, struct {
						scope  string
						reason string
					}{
						constants.ScopeCorrelation, fmt.Sprintf("Required for %s role to perform cross-user correlation and moderation", role),
					})
					break
				}
			}

			// Check each required mapping
			for _, requiredMapping := range requiredMappings {
				// Check if mapping exists by looking for any mapping with this scope
				// Get all mappings for the pseudonym and check if any have the required scope
				existingMappings, err := identityMappingDAO.GetIdentityMappingsByPseudonymID(ctx, pseudonym.PseudonymID)
				hasMapping := false
				if err != nil {
					log.Warn().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredMapping.scope).Msg("Failed to get identity mappings, assuming missing")
				} else {
					// Check if any existing mapping has the required scope
					for _, mapping := range existingMappings {
						if mapping.KeyScope == requiredMapping.scope {
							hasMapping = true
							break
						}
					}
				}

				if !hasMapping {
					result.missingMaps = append(result.missingMaps, requiredMapping.scope)
				}

				if !dryRun {
					log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredMapping.scope).Str("reason", requiredMapping.reason).Msg("Creating missing identity mapping")

					// Get the role key data for this scope
					keyData, err := roleKeyDAO.GetKeyData(ctx, pseudonym.PseudonymID, requiredMapping.scope, nil)
					if err != nil {
						log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredMapping.scope).Msg("Failed to get key data for mapping")
						continue
					}

					// Create identity mapping
					err = createIdentityMapping(ctx, identityMappingDAO, ibeSystem, user, pseudonym, requiredMapping.scope, keyData)
					if err != nil {
						log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredMapping.scope).Msg("Failed to create identity mapping")
						continue
					}

					result.createdMaps = append(result.createdMaps, requiredMapping.scope)
					log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredMapping.scope).Msg("Created identity mapping")
				} else {
					log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", requiredMapping.scope).Str("reason", requiredMapping.reason).Msg("Would create missing identity mapping")
				}
			}

			results = append(results, result)
		}
	}

	// Print summary
	log.Info().Msg("=== AUDIT SUMMARY ===")
	totalMissingKeys := 0
	totalMissingMaps := 0
	totalCreatedKeys := 0
	totalCreatedMaps := 0

	for _, result := range results {
		if len(result.missingKeys) > 0 || len(result.missingMaps) > 0 {
			log.Info().
				Int64("user_id", result.userID).
				Str("pseudonym_id", result.pseudonymID).
				Strs("missing_keys", result.missingKeys).
				Strs("missing_mappings", result.missingMaps).
				Strs("created_keys", result.createdKeys).
				Strs("created_mappings", result.createdMaps).
				Msg("User audit result")
		}

		totalMissingKeys += len(result.missingKeys)
		totalMissingMaps += len(result.missingMaps)
		totalCreatedKeys += len(result.createdKeys)
		totalCreatedMaps += len(result.createdMaps)
	}

	log.Info().
		Int("total_missing_keys", totalMissingKeys).
		Int("total_missing_mappings", totalMissingMaps).
		Int("total_created_keys", totalCreatedKeys).
		Int("total_created_mappings", totalCreatedMaps).
		Bool("dry_run", dryRun).
		Msg("Audit completed")

	if dryRun && (totalMissingKeys > 0 || totalMissingMaps > 0) {
		log.Info().Msg("Run without --dry-run to create missing keys and mappings")
	}

	return nil
}

// createIdentityMapping creates an identity mapping for a user and pseudonym
func createIdentityMapping(ctx context.Context, identityMappingDAO dao.IdentityMappingDAOInterface, ibeSystem *ibe.IBESystem, user *models.User, pseudonym *models.Pseudonym, scope string, keyData []byte) error {
	// Generate fingerprint for the real identity
	fingerprint := ibeSystem.GenerateFingerprint(user.Email)

	// Determine the domain based on the scope
	domain := ibeSystem.GetDomainForRole("user") // Default domain

	// Encrypt the identity mapping
	encryptedData, err := ibeSystem.EncryptIdentityWithDomain(user.Email, pseudonym.PseudonymID, domain, keyData)
	if err != nil {
		return fmt.Errorf("failed to encrypt identity mapping: %w", err)
	}

	// Create the identity mapping
	keyVersion := ibeSystem.GetKeyVersion()
	mapping := &models.IdentityMappingSetter{
		Fingerprint:               &fingerprint,
		PseudonymID:               &pseudonym.PseudonymID,
		EncryptedRealIdentity:     &encryptedData,
		EncryptedPseudonymMapping: &encryptedData,
		KeyVersion:                &keyVersion,
		UserID:                    &user.UserID,
		KeyScope:                  &scope,
	}

	_, err = identityMappingDAO.CreateIdentityMapping(ctx, mapping)
	if err != nil {
		return fmt.Errorf("failed to create identity mapping: %w", err)
	}

	return nil
}

// ReEncryptIdentityMappings re-encrypts existing identity mappings with new keys
func ReEncryptIdentityMappings(allUsers bool, dryRun bool, cfg *config.Config) error {
	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize IBE system
	ibeSystem, err := ibe.NewIBESystemFromConfig(cfg.IBE.DomainKeysDir, cfg.IBE.KeyVersion, cfg.IBE.Salt)
	if err != nil {
		return fmt.Errorf("failed to initialize IBE system: %w", err)
	}

	// Create DAOs
	userDAO := dao.NewUserDAO(db)
	pseudonymDAO := dao.NewPseudonymDAO(db, ibeSystem, dao.NewIdentityMappingDAO(db), userDAO, dao.NewRoleKeyDAO(db, nil), dao.NewUserBlocksDAO(db))
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db, nil)

	ctx := context.Background()

	// Get all users with pagination
	var usersToProcess []*models.User
	if allUsers {
		const batchSize = 1000
		offset := 0
		for {
			users, err := userDAO.ListUsers(ctx, batchSize, offset)
			if err != nil {
				return fmt.Errorf("failed to get users at offset %d: %w", offset, err)
			}
			if len(users) == 0 {
				break
			}
			usersToProcess = append(usersToProcess, users...)
			if len(users) < batchSize {
				break
			}
			offset += batchSize
		}
	} else {
		return fmt.Errorf("--all-users flag is required for re-encryption")
	}

	log.Info().Int("user_count", len(usersToProcess)).Msg("Re-encrypting identity mappings for users")

	totalReEncrypted := 0
	totalErrors := 0

	for _, user := range usersToProcess {
		log.Info().Str("email", user.Email).Int64("user_id", user.UserID).Msg("Processing user")

		// Get all pseudonyms for this user
		pseudonyms, err := pseudonymDAO.GetPseudonymsByUserIDDirect(ctx, user.UserID)
		if err != nil {
			log.Error().Err(err).Int64("user_id", user.UserID).Msg("Failed to get pseudonyms for user")
			totalErrors++
			continue
		}

		for _, pseudonym := range pseudonyms {
			log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Processing pseudonym")

			// Get all identity mappings for this pseudonym
			mappings, err := identityMappingDAO.GetIdentityMappingsByPseudonymID(ctx, pseudonym.PseudonymID)
			if err != nil {
				log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to get identity mappings")
				totalErrors++
				continue
			}

			for _, mapping := range mappings {
				log.Info().Str("scope", mapping.KeyScope).Msg("Re-encrypting identity mapping")

				if dryRun {
					log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", mapping.KeyScope).Msg("Would re-encrypt identity mapping (dry run)")
					totalReEncrypted++
					continue
				}

				// Get the role key data for this scope
				keyData, err := roleKeyDAO.GetKeyData(ctx, pseudonym.PseudonymID, mapping.KeyScope, nil)
				if err != nil {
					log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", mapping.KeyScope).Msg("Failed to get key data for re-encryption")
					totalErrors++
					continue
				}

				// Re-encrypt the identity mapping
				err = reEncryptIdentityMapping(ctx, identityMappingDAO, ibeSystem, user, pseudonym, mapping, keyData)
				if err != nil {
					log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", mapping.KeyScope).Msg("Failed to re-encrypt identity mapping")
					totalErrors++
					continue
				}

				log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("scope", mapping.KeyScope).Msg("Successfully re-encrypted identity mapping")
				totalReEncrypted++
			}
		}
	}

	log.Info().Int("total_re_encrypted", totalReEncrypted).Int("total_errors", totalErrors).Bool("dry_run", dryRun).Msg("Re-encryption completed")

	return nil
}

// reEncryptIdentityMapping re-encrypts a single identity mapping with new key
func reEncryptIdentityMapping(ctx context.Context, identityMappingDAO dao.IdentityMappingDAOInterface, ibeSystem *ibe.IBESystem, user *models.User, pseudonym *models.Pseudonym, mapping *models.IdentityMapping, keyData []byte) error {
	// Generate fingerprint for the real identity
	fingerprint := ibeSystem.GenerateFingerprint(user.Email)

	// Determine the domain based on the scope
	domain := ibeSystem.GetDomainForRole("user") // Default domain

	// Encrypt the identity mapping with new key
	encryptedData, err := ibeSystem.EncryptIdentityWithDomain(user.Email, pseudonym.PseudonymID, domain, keyData)
	if err != nil {
		return fmt.Errorf("failed to encrypt identity mapping: %w", err)
	}

	// Update the identity mapping with new encrypted data
	keyVersion := ibeSystem.GetKeyVersion()
	updateData := &models.IdentityMappingSetter{
		Fingerprint:               &fingerprint,
		EncryptedRealIdentity:     &encryptedData,
		EncryptedPseudonymMapping: &encryptedData,
		KeyVersion:                &keyVersion,
	}

	err = identityMappingDAO.UpdateIdentityMapping(ctx, mapping.MappingID.String(), updateData)
	if err != nil {
		return fmt.Errorf("failed to update identity mapping: %w", err)
	}

	return nil
}
