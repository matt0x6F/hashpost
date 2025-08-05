package commands

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/stephenafamo/bob/types"
	"golang.org/x/term"
)

// AdminCreateInput defines the input for creating an admin user.
type AdminCreateInput struct {
	Email          string `doc:"Email address for the admin user" json:"email"`
	Password       string `doc:"Password for the admin user" json:"password"`
	DisplayName    string `doc:"Display name for the admin user" json:"display_name"`
	MFAEnabled     bool   `doc:"Enable MFA for the admin user" json:"mfa_enabled" default:"true"`
	NonInteractive bool   `doc:"Non-interactive mode (requires all flags)" json:"non_interactive"`
}

// SetModeratorInput defines the input for setting a pseudonym as a forum moderator
type SetModeratorInput struct {
	SubforumName   string `doc:"Name of the subforum" json:"subforum_name"`
	PseudonymID    string `doc:"Pseudonym ID to set as moderator" json:"pseudonym_id"`
	NonInteractive bool   `doc:"Non-interactive mode (requires all flags)" json:"non_interactive"`
}

// DeleteUserInput defines the input for deleting a user
type DeleteUserInput struct {
	Email          string `doc:"Email address of the user to delete" json:"email"`
	NonInteractive bool   `doc:"Non-interactive mode (requires all flags)" json:"non_interactive"`
	Force          bool   `doc:"Force deletion without confirmation" json:"force"`
}

// CreateAdminUser creates a new admin user
func CreateAdminUser() error {
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

	// Get admin creation input
	input := getAdminCreateInput()

	// Validate input
	if err := validateAdminInput(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	// Create user DAO
	userDAO := dao.NewUserDAO(db)

	// Initialize IBE system and identity mapping DAO
	ibeSystem := ibe.NewIBESystemFromEnv()
	identityMappingDAO := dao.NewIdentityMappingDAO(db)

	// Check if user already exists
	ctx := context.Background()
	existingUser, err := userDAO.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	var user *models.User
	var adminUsername string

	// Use display name as admin username (display name is required)
	adminUsername = input.DisplayName

	if existingUser != nil {
		// User exists - update them with admin role and capabilities
		log.Info().Str("email", input.Email).Msg("User already exists, updating with admin role")

		// Hash the password if provided
		var passwordHash string
		if input.Password != "" {
			passwordHash = hashPassword(input.Password)
		} else {
			// Keep existing password if not provided
			passwordHash = existingUser.PasswordHash
		}

		// Generate admin password hash
		adminPasswordHash := hashPassword(input.Password)

		// Prepare roles and capabilities for platform admin
		adminRole := constants.RolePlatformAdmin
		roles := []string{"user", adminRole} // Include both user and platform admin roles

		// Get capabilities for platform admin role
		capabilities := getCapabilitiesForRole(adminRole)

		// Convert to JSON
		rolesJSON, err := json.Marshal(roles)
		if err != nil {
			return fmt.Errorf("failed to marshal roles: %w", err)
		}

		capabilitiesJSON, err := json.Marshal(capabilities)
		if err != nil {
			return fmt.Errorf("failed to marshal capabilities: %w", err)
		}

		// Update user with admin-specific fields
		rolesNull := sql.Null[types.JSON[json.RawMessage]]{}
		if err := rolesNull.Scan(rolesJSON); err != nil {
			return fmt.Errorf("failed to scan roles: %w", err)
		}

		capabilitiesNull := sql.Null[types.JSON[json.RawMessage]]{}
		if err := capabilitiesNull.Scan(capabilitiesJSON); err != nil {
			return fmt.Errorf("failed to scan capabilities: %w", err)
		}

		adminUsernameNull := sql.Null[string]{}
		if err := adminUsernameNull.Scan(adminUsername); err != nil {
			return fmt.Errorf("failed to scan admin username: %w", err)
		}

		adminPasswordHashNull := sql.Null[string]{}
		if err := adminPasswordHashNull.Scan(adminPasswordHash); err != nil {
			return fmt.Errorf("failed to scan admin password hash: %w", err)
		}

		mfaEnabledNull := sql.Null[bool]{}
		if err := mfaEnabledNull.Scan(input.MFAEnabled); err != nil {
			return fmt.Errorf("failed to scan MFA enabled: %w", err)
		}

		// Derive scopes from platform admin role
		scopes := getScopesForRole(adminRole)
		scopesJSON, err := json.Marshal(scopes)
		if err != nil {
			return fmt.Errorf("failed to marshal scopes: %w", err)
		}

		adminScopeNull := sql.Null[string]{}
		if err := adminScopeNull.Scan(string(scopesJSON)); err != nil {
			return fmt.Errorf("failed to scan admin scope: %w", err)
		}

		updates := &models.UserSetter{
			PasswordHash:      &passwordHash,
			Roles:             &rolesNull,
			AdminUsername:     &adminUsernameNull,
			AdminPasswordHash: &adminPasswordHashNull,
			MfaEnabled:        &mfaEnabledNull,
			AdminScope:        &adminScopeNull,
		}

		if err := userDAO.UpdateUser(ctx, existingUser.UserID, updates); err != nil {
			return fmt.Errorf("failed to update user with admin fields: %w", err)
		}

		user = existingUser
		log.Info().Int64("user_id", user.UserID).Msg("User updated with admin role")
	} else {
		// User doesn't exist - create new user
		log.Info().Str("email", input.Email).Msg("Creating new admin user")

		// Hash the password
		passwordHash := hashPassword(input.Password)

		// Use display name as admin username (display name is required)
		adminUsername = input.DisplayName

		// Generate admin password hash
		adminPasswordHash := hashPassword(input.Password)

		// Prepare roles and capabilities for platform admin
		adminRole := constants.RolePlatformAdmin
		roles := []string{"user", adminRole} // Include both user and platform admin roles

		// Get capabilities for platform admin role
		capabilities := getCapabilitiesForRole(adminRole)

		// Convert to JSON
		rolesJSON, err := json.Marshal(roles)
		if err != nil {
			return fmt.Errorf("failed to marshal roles: %w", err)
		}

		capabilitiesJSON, err := json.Marshal(capabilities)
		if err != nil {
			return fmt.Errorf("failed to marshal capabilities: %w", err)
		}

		// Create the user
		user, err = userDAO.CreateUser(ctx, input.Email, passwordHash)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Update user with admin-specific fields
		rolesNull := sql.Null[types.JSON[json.RawMessage]]{}
		if err := rolesNull.Scan(rolesJSON); err != nil {
			return fmt.Errorf("failed to scan roles: %w", err)
		}

		capabilitiesNull := sql.Null[types.JSON[json.RawMessage]]{}
		if err := capabilitiesNull.Scan(capabilitiesJSON); err != nil {
			return fmt.Errorf("failed to scan capabilities: %w", err)
		}

		adminUsernameNull := sql.Null[string]{}
		if err := adminUsernameNull.Scan(adminUsername); err != nil {
			return fmt.Errorf("failed to scan admin username: %w", err)
		}

		adminPasswordHashNull := sql.Null[string]{}
		if err := adminPasswordHashNull.Scan(adminPasswordHash); err != nil {
			return fmt.Errorf("failed to scan admin password hash: %w", err)
		}

		mfaEnabledNull := sql.Null[bool]{}
		if err := mfaEnabledNull.Scan(input.MFAEnabled); err != nil {
			return fmt.Errorf("failed to scan MFA enabled: %w", err)
		}

		// Derive scopes from platform admin role
		scopes := getScopesForRole(adminRole)
		scopesJSON, err := json.Marshal(scopes)
		if err != nil {
			return fmt.Errorf("failed to marshal scopes: %w", err)
		}

		adminScopeNull := sql.Null[string]{}
		if err := adminScopeNull.Scan(string(scopesJSON)); err != nil {
			return fmt.Errorf("failed to scan admin scope: %w", err)
		}

		updates := &models.UserSetter{
			Roles:             &rolesNull,
			AdminUsername:     &adminUsernameNull,
			AdminPasswordHash: &adminPasswordHashNull,
			MfaEnabled:        &mfaEnabledNull,
			AdminScope:        &adminScopeNull,
		}

		if err := userDAO.UpdateUser(ctx, user.UserID, updates); err != nil {
			return fmt.Errorf("failed to update user with admin fields: %w", err)
		}

		log.Info().Int64("user_id", user.UserID).Msg("New admin user created")
	}

	// Create a pseudonym for the admin user with identity mapping
	pseudonymDAO := dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAO, dao.NewUserDAO(db), dao.NewRoleKeyDAO(db), dao.NewUserBlocksDAO(db))

	// Use display name for pseudonym (it's required)
	displayName := input.DisplayName
	if displayName == "" {
		return fmt.Errorf("display name is required for admin user creation")
	}

	// Define admin role for platform admin
	adminRole := constants.RolePlatformAdmin

	// Check if user already has a pseudonym
	existingPseudonyms, err := pseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, adminRole, "authentication")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check existing pseudonyms, will create new one")
	}

	var pseudonym *models.Pseudonym
	if len(existingPseudonyms) > 0 {
		// User already has pseudonyms, use the first one
		pseudonym = existingPseudonyms[0]
		log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Using existing pseudonym")
	} else {
		// Create new pseudonym and identity mapping
		pseudonym, err = pseudonymDAO.CreatePseudonymWithIdentityMapping(ctx, user.UserID, displayName)
		if err != nil {
			return fmt.Errorf("failed to create pseudonym for admin user: %w", err)
		}

		log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Created new pseudonym")
	}

	// Ensure default role keys for the admin user's pseudonym
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	if err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, pseudonym.PseudonymID, []string{adminRole}); err != nil {
		return fmt.Errorf("failed to create default role keys for admin user: %w", err)
	}

	log.Info().
		Int64("user_id", user.UserID).
		Str("email", input.Email).
		Str("admin_username", adminUsername).
		Str("role", adminRole).
		Bool("mfa_enabled", input.MFAEnabled).
		Str("pseudonym_id", pseudonym.PseudonymID).
		Str("display_name", pseudonym.DisplayName).
		Msg("Admin user created successfully")

	action := "created"
	if existingUser != nil {
		action = "updated"
	}

	fmt.Printf("✅ Admin user %s successfully!\n", action)
	fmt.Printf("   User ID: %d\n", user.UserID)
	fmt.Printf("   Email: %s\n", input.Email)
	fmt.Printf("   Admin Username: %s\n", adminUsername)
	fmt.Printf("   Role: %s\n", adminRole)
	fmt.Printf("   MFA Enabled: %t\n", input.MFAEnabled)
	fmt.Printf("   Pseudonym ID: %s\n", pseudonym.PseudonymID)
	fmt.Printf("   Display Name: %s\n", pseudonym.DisplayName)
	scopes := getScopesForRole(adminRole)
	fmt.Printf("   Admin Scopes: %s\n", strings.Join(scopes, ", "))

	return nil
}

// SetModerator sets a pseudonym as a forum moderator
func SetModerator() error {
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

	// Get moderator input
	input := getSetModeratorInput()

	// Validate input
	if err := validateSetModeratorInput(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	// Create DAOs
	subforumDAO := dao.NewSubforumDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	ibeSystem := ibe.NewIBESystemFromEnv()
	pseudonymDAO := dao.NewPseudonymDAO(db, ibeSystem, dao.NewIdentityMappingDAO(db), dao.NewUserDAO(db), roleKeyDAO, dao.NewUserBlocksDAO(db))

	ctx := context.Background()

	// Find the subforum
	subforum, err := subforumDAO.GetSubforumByName(ctx, input.SubforumName)
	if err != nil {
		return fmt.Errorf("failed to get subforum: %w", err)
	}
	if subforum == nil {
		return fmt.Errorf("subforum '%s' not found", input.SubforumName)
	}

	// Find the pseudonym
	pseudonym, err := pseudonymDAO.GetPseudonymByID(ctx, input.PseudonymID)
	if err != nil {
		return fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		return fmt.Errorf("pseudonym '%s' not found", input.PseudonymID)
	}

	// Check if pseudonym already has moderator role keys for this subforum
	existingModeratorKey, err := roleKeyDAO.GetRoleKey(ctx, input.PseudonymID, "moderation", &subforum.SubforumID)
	if err != nil {
		return fmt.Errorf("failed to check existing moderator role keys: %w", err)
	}

	if existingModeratorKey != nil {
		fmt.Printf("✅ Pseudonym '%s' is already a moderator of subforum '%s'\n", input.PseudonymID, input.SubforumName)
		return nil
	}

	// Create moderator role keys for the pseudonym in this subforum
	// Get the moderator role definition
	moderatorRole := constants.GetRoleDefinition(constants.RoleModerator)
	if moderatorRole == nil {
		return fmt.Errorf("failed to get moderator role definition")
	}

	// Create role keys for each scope that the moderator role has
	for _, scope := range moderatorRole.Scopes {
		capabilities := moderatorRole.Capabilities[scope]
		if len(capabilities) == 0 {
			continue
		}

		// Generate role key for this scope
		keyData := ibeSystem.GenerateRoleKey(constants.RoleModerator, scope, time.Now().AddDate(1, 0, 0))

		// Store the role key
		_, err = roleKeyDAO.CreateRoleKey(ctx, constants.RoleModerator, scope, keyData, capabilities, time.Now().AddDate(1, 0, 0), constants.SystemPseudonymID, input.PseudonymID, &subforum.SubforumID)
		if err != nil {
			return fmt.Errorf("failed to create role key for scope %s: %w", scope, err)
		}
	}

	log.Info().
		Str("subforum_name", input.SubforumName).
		Int32("subforum_id", subforum.SubforumID).
		Str("pseudonym_id", input.PseudonymID).
		Str("pseudonym_display_name", pseudonym.DisplayName).
		Msg("Moderator added successfully")

	fmt.Printf("✅ Successfully set pseudonym '%s' (%s) as moderator of subforum '%s'\n",
		input.PseudonymID, pseudonym.DisplayName, input.SubforumName)

	return nil
}

// DeleteUser deletes a user and all associated data
func DeleteUser() error {
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

	// Get delete user input
	input := getDeleteUserInput()

	// Validate input
	if err := validateDeleteUserInput(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	// Create DAOs
	userDAO := dao.NewUserDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	pseudonymDAO := dao.NewPseudonymDAO(db, ibe.NewIBESystemFromEnv(), identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)
	emailVerificationTokenDAO := dao.NewEmailVerificationTokenDAO(db)
	passwordResetTokenDAO := dao.NewPasswordResetTokenDAO(db)

	// Find the user
	ctx := context.Background()
	user, err := userDAO.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user with email %s not found", input.Email)
	}

	// Show user info and ask for confirmation
	log.Info().
		Int64("user_id", user.UserID).
		Str("email", user.Email).
		Msg("Found user to delete")

	if !input.Force && !input.NonInteractive {
		fmt.Printf("Are you sure you want to delete user %s (ID: %d)? This action cannot be undone. (y/N): ", input.Email, user.UserID)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Start transaction for atomic deletion
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Delete in the correct order to respect foreign key constraints:
	// 1. Email verification tokens
	// 2. Password reset tokens
	// 3. Pseudonyms (before identity mappings since pseudonym deletion depends on them)
	// 4. Identity mappings
	// 5. Role keys
	// 6. User

	log.Info().Msg("Starting user deletion process...")

	// 1. Delete email verification tokens
	if err := emailVerificationTokenDAO.DeleteTokensByUserID(ctx, user.UserID); err != nil {
		return fmt.Errorf("failed to delete email verification tokens: %w", err)
	}
	log.Info().Msg("Deleted email verification tokens")

	// 2. Delete password reset tokens
	if err := passwordResetTokenDAO.DeleteTokensByUserID(ctx, user.UserID); err != nil {
		return fmt.Errorf("failed to delete password reset tokens: %w", err)
	}
	log.Info().Msg("Deleted password reset tokens")

	// 3. Delete pseudonyms (before identity mappings since pseudonym deletion depends on them)
	if err := pseudonymDAO.DeleteByUserID(ctx, user.UserID); err != nil {
		return fmt.Errorf("failed to delete pseudonyms: %w", err)
	}
	log.Info().Msg("Deleted pseudonyms")

	// 4. Delete identity mappings
	if err := identityMappingDAO.DeleteByUserID(ctx, user.UserID); err != nil {
		return fmt.Errorf("failed to delete identity mappings: %w", err)
	}
	log.Info().Msg("Deleted identity mappings")

	// 5. Delete role keys for all pseudonyms
	pseudonyms, err := pseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, "", "authentication")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get pseudonyms for role key deletion, continuing")
	} else {
		for _, pseudonym := range pseudonyms {
			if err := roleKeyDAO.DeleteByPseudonymID(ctx, pseudonym.PseudonymID); err != nil {
				log.Warn().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to delete role keys for pseudonym")
			}
		}
	}
	log.Info().Msg("Deleted role keys")

	// 6. Delete user
	if err := userDAO.DeleteUser(ctx, user.UserID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	log.Info().Msg("Deleted user")

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Int64("user_id", user.UserID).
		Str("email", input.Email).
		Msg("User deleted successfully")

	fmt.Printf("User %s (ID: %d) has been deleted successfully.\n", input.Email, user.UserID)
	return nil
}

// getAdminCreateInput prompts for admin user creation input
func getAdminCreateInput() *AdminCreateInput {
	input := &AdminCreateInput{}

	// Check if we're in non-interactive mode
	cmd := cobra.Command{}
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("display-name", "", "")
	cmd.Flags().Bool("mfa-enabled", true, "")
	cmd.Flags().Bool("non-interactive", false, "")

	// Parse flags from os.Args
	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		log.Fatal().Err(err).Msg("failed to parse flags")
	}

	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive {
		// Non-interactive mode - get values from flags
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		displayName, _ := cmd.Flags().GetString("display-name")
		mfaEnabled, _ := cmd.Flags().GetBool("mfa-enabled")

		if email == "" || password == "" {
			log.Fatal().Msg("email and password are required in non-interactive mode")
		}

		if displayName == "" {
			log.Fatal().Msg("display name is required in non-interactive mode")
		}

		input.Email = email
		input.Password = password
		input.DisplayName = displayName
		input.MFAEnabled = mfaEnabled
		input.NonInteractive = true

		return input
	}

	// Interactive mode
	fmt.Println("Create Admin User")
	fmt.Println("=================")

	fmt.Print("Email: ")
	if _, err := fmt.Scanln(&input.Email); err != nil {
		log.Fatal().Err(err).Msg("failed to read email")
	}

	// Get password with hidden input
	input.Password = getPasswordInput("Password: ")

	// Confirm password
	confirmPassword := getPasswordInput("Confirm Password: ")
	if input.Password != confirmPassword {
		log.Fatal().Msg("passwords do not match")
	}

	fmt.Print("Display Name (required, cannot be email): ")
	if _, err := fmt.Scanln(&input.DisplayName); err != nil {
		log.Fatal().Err(err).Msg("failed to read display name")
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		log.Fatal().Msg("display name is required")
	}
	if strings.EqualFold(strings.TrimSpace(input.DisplayName), strings.TrimSpace(input.Email)) {
		log.Fatal().Msg("display name cannot be the same as email address")
	}

		// Admin role is automatically set to platform_admin for initial admin creation

	// Scopes are automatically derived from roles, no need to prompt for them

	fmt.Print("Enable MFA (y/n) [y]: ")
	var mfaInput string
	if _, err := fmt.Scanln(&mfaInput); err != nil {
		log.Fatal().Err(err).Msg("failed to read MFA input")
	}
	input.MFAEnabled = mfaInput != "n" && mfaInput != "N"

	return input
}

// getSetModeratorInput prompts for moderator input
func getSetModeratorInput() *SetModeratorInput {
	input := &SetModeratorInput{}

	// Check if we're in non-interactive mode
	cmd := cobra.Command{}
	cmd.Flags().String("subforum", "", "")
	cmd.Flags().String("pseudonym", "", "")
	cmd.Flags().Bool("non-interactive", false, "")

	// Parse flags from os.Args
	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		log.Fatal().Err(err).Msg("failed to parse flags")
	}

	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive {
		// Non-interactive mode - get values from flags
		subforum, _ := cmd.Flags().GetString("subforum")
		pseudonym, _ := cmd.Flags().GetString("pseudonym")

		if subforum == "" || pseudonym == "" {
			log.Fatal().Msg("subforum and pseudonym are required in non-interactive mode")
		}

		input.SubforumName = subforum
		input.PseudonymID = pseudonym
		input.NonInteractive = true

		return input
	}

	// Interactive mode
	fmt.Println("Set Forum Moderator")
	fmt.Println("===================")

	fmt.Print("Subforum Name: ")
	if _, err := fmt.Scanln(&input.SubforumName); err != nil {
		log.Fatal().Err(err).Msg("failed to read subforum name")
	}

	fmt.Print("Pseudonym ID: ")
	if _, err := fmt.Scanln(&input.PseudonymID); err != nil {
		log.Fatal().Err(err).Msg("failed to read pseudonym ID")
	}

	return input
}

// getDeleteUserInput prompts for user deletion input
func getDeleteUserInput() *DeleteUserInput {
	input := &DeleteUserInput{}

	// Check if we're in non-interactive mode
	cmd := cobra.Command{}
	cmd.Flags().String("email", "", "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("non-interactive", false, "")

	// Parse flags from os.Args
	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		log.Fatal().Err(err).Msg("failed to parse flags")
	}

	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive {
		// Non-interactive mode - get values from flags
		email, _ := cmd.Flags().GetString("email")
		force, _ := cmd.Flags().GetBool("force")

		if email == "" {
			log.Fatal().Msg("email is required in non-interactive mode")
		}

		input.Email = email
		input.Force = force
		input.NonInteractive = true

		return input
	}

	// Interactive mode
	fmt.Println("Delete User")
	fmt.Println("===========")

	fmt.Print("Email: ")
	if _, err := fmt.Scanln(&input.Email); err != nil {
		log.Fatal().Err(err).Msg("failed to read email")
	}

	fmt.Print("Force deletion without confirmation (y/N) [N]: ")
	var forceInput string
	if _, err := fmt.Scanln(&forceInput); err != nil {
		log.Fatal().Err(err).Msg("failed to read force input")
	}
	input.Force = forceInput != "n" && forceInput != "N"

	return input
}

// getPasswordInput prompts for a password with hidden input
func getPasswordInput(prompt string) string {
	fmt.Print(prompt)

	// Get the file descriptor for stdin
	fd := int(syscall.Stdin)

	// Read password with hidden input
	bytePassword, err := term.ReadPassword(fd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read password")
	}

	// Print newline after password input
	fmt.Println()

	// Convert to string and trim whitespace
	password := strings.TrimSpace(string(bytePassword))

	return password
}

// validateAdminInput validates the admin creation input
func validateAdminInput(input *AdminCreateInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Password is only required for new users or when explicitly provided
	if input.Password == "" && !input.NonInteractive {
		return fmt.Errorf("password is required")
	}

	if input.Password != "" && len(input.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if input.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}

	// Prevent using email as display name
	if strings.EqualFold(strings.TrimSpace(input.DisplayName), strings.TrimSpace(input.Email)) {
		return fmt.Errorf("display name cannot be the same as email address")
	}

	// Additional display name validation
	displayName := strings.TrimSpace(input.DisplayName)
	if len(displayName) < 2 {
		return fmt.Errorf("display name must be at least 2 characters long")
	}
	if len(displayName) > 50 {
		return fmt.Errorf("display name must be 50 characters or less")
	}

		// Admin role is automatically set to platform_admin, no validation needed

	return nil
}

// validateSetModeratorInput validates the set moderator input
func validateSetModeratorInput(input *SetModeratorInput) error {
	if input.SubforumName == "" {
		return fmt.Errorf("subforum name is required")
	}

	if input.PseudonymID == "" {
		return fmt.Errorf("pseudonym ID is required")
	}

	// Additional validation
	subforumName := strings.TrimSpace(input.SubforumName)
	if len(subforumName) < 2 {
		return fmt.Errorf("subforum name must be at least 2 characters long")
	}
	if len(subforumName) > 50 {
		return fmt.Errorf("subforum name must be 50 characters or less")
	}

	return nil
}

// validateDeleteUserInput validates the delete user input
func validateDeleteUserInput(input *DeleteUserInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}

	return nil
}

// hashPassword hashes a password using SHA-256
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// getCapabilitiesForRole returns the capabilities for a given admin role
func getCapabilitiesForRole(role string) []string {
	// Basic user capabilities that all users should have
	basicUserCapabilities := []string{
		constants.CapabilityCreateContent,
		constants.CapabilityVote,
		constants.CapabilityMessage,
		constants.CapabilityReport,
		constants.CapabilityCreateSubforum,
	}

	// Use the constants package to get role capabilities
	roleCapabilities := constants.GetRoleCapabilities(role)

	// Start with basic user capabilities
	capabilities := append([]string{}, basicUserCapabilities...)

	// Add role-specific capabilities
	capabilities = append(capabilities, roleCapabilities...)

	return capabilities
}

// getCapabilitiesForRoles returns the capabilities for multiple roles
func getCapabilitiesForRoles(roles []string) []string {
	// Basic user capabilities that all users should have
	basicUserCapabilities := []string{
		constants.CapabilityCreateContent,
		constants.CapabilityVote,
		constants.CapabilityMessage,
		constants.CapabilityReport,
		constants.CapabilityCreateSubforum,
	}

	// Start with basic user capabilities
	capabilities := append([]string{}, basicUserCapabilities...)

	// Add capabilities from all roles
	for _, role := range roles {
		roleCapabilities := constants.GetRoleCapabilities(role)
		capabilities = append(capabilities, roleCapabilities...)
	}

	return capabilities
}

// getScopesForRole returns the scopes for a single role
func getScopesForRole(role string) []string {
	// Always include basic scopes that every user should have
	basicScopes := []string{
		constants.ScopeAuthentication,
		constants.ScopeSelfCorrelation,
	}
	
	// Get scopes from the role
	roleScopes := constants.GetRoleScopes(role)
	
	// Combine basic scopes with role scopes, avoiding duplicates
	scopeSet := make(map[string]bool)
	var allScopes []string
	
	// Add basic scopes first
	for _, scope := range basicScopes {
		scopeSet[scope] = true
		allScopes = append(allScopes, scope)
	}
	
	// Add role scopes
	for _, scope := range roleScopes {
		if !scopeSet[scope] {
			scopeSet[scope] = true
			allScopes = append(allScopes, scope)
		}
	}
	
	return allScopes
}

// getScopesForRoles returns the scopes for multiple roles
func getScopesForRoles(roles []string) []string {
	var allScopes []string
	scopeSet := make(map[string]bool)

	// Always include basic scopes that every user should have
	basicScopes := []string{
		constants.ScopeAuthentication,
		constants.ScopeSelfCorrelation,
	}
	
	for _, scope := range basicScopes {
		scopeSet[scope] = true
		allScopes = append(allScopes, scope)
	}

	// Collect additional scopes from all roles
	for _, role := range roles {
		roleScopes := constants.GetRoleScopes(role)
		for _, scope := range roleScopes {
			if !scopeSet[scope] {
				scopeSet[scope] = true
				allScopes = append(allScopes, scope)
			}
		}
	}

	return allScopes
}
