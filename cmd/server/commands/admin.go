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

// AdminCreateInput defines the input for creating an admin user
type AdminCreateInput struct {
	Email          string `doc:"Email address for the admin user" json:"email"`
	Password       string `doc:"Password for the admin user" json:"password"`
	AdminRole      string `doc:"Admin role (platform_admin, trust_safety, legal_team)" json:"admin_role" default:"platform_admin"`
	DisplayName    string `doc:"Display name for the admin user" json:"display_name"`
	AdminScope     string `doc:"Admin scope (optional)" json:"admin_scope"`
	MFAEnabled     bool   `doc:"Enable MFA for the admin user" json:"mfa_enabled" default:"true"`
	NonInteractive bool   `doc:"Non-interactive mode (requires all flags)" json:"non_interactive"`
}

// SetModeratorInput defines the input for setting a pseudonym as a forum moderator
type SetModeratorInput struct {
	SubforumName   string `doc:"Name of the subforum" json:"subforum_name"`
	PseudonymID    string `doc:"Pseudonym ID to set as moderator" json:"pseudonym_id"`
	NonInteractive bool   `doc:"Non-interactive mode (requires all flags)" json:"non_interactive"`
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

		// Prepare roles and capabilities
		roles := []string{"user", input.AdminRole} // Include both user and admin roles
		capabilities := getCapabilitiesForRole(input.AdminRole)

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
		rolesNull.Scan(rolesJSON)

		capabilitiesNull := sql.Null[types.JSON[json.RawMessage]]{}
		capabilitiesNull.Scan(capabilitiesJSON)

		adminUsernameNull := sql.Null[string]{}
		adminUsernameNull.Scan(adminUsername)

		adminPasswordHashNull := sql.Null[string]{}
		adminPasswordHashNull.Scan(adminPasswordHash)

		mfaEnabledNull := sql.Null[bool]{}
		mfaEnabledNull.Scan(input.MFAEnabled)

		adminScopeNull := sql.Null[string]{}
		if input.AdminScope != "" {
			adminScopeNull.Scan(input.AdminScope)
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

		// Prepare roles and capabilities
		roles := []string{"user", input.AdminRole} // Include both user and admin roles
		capabilities := getCapabilitiesForRole(input.AdminRole)

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
		rolesNull.Scan(rolesJSON)

		capabilitiesNull := sql.Null[types.JSON[json.RawMessage]]{}
		capabilitiesNull.Scan(capabilitiesJSON)

		adminUsernameNull := sql.Null[string]{}
		adminUsernameNull.Scan(adminUsername)

		adminPasswordHashNull := sql.Null[string]{}
		adminPasswordHashNull.Scan(adminPasswordHash)

		mfaEnabledNull := sql.Null[bool]{}
		mfaEnabledNull.Scan(input.MFAEnabled)

		adminScopeNull := sql.Null[string]{}
		if input.AdminScope != "" {
			adminScopeNull.Scan(input.AdminScope)
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

	// Ensure default role keys for the admin user
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	if err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, user.UserID); err != nil {
		return fmt.Errorf("failed to create default role keys for admin user: %w", err)
	}

	// Create a pseudonym for the admin user with identity mapping
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	securePseudonymDAO := dao.NewSecurePseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)

	// Use display name for pseudonym (it's required)
	displayName := input.DisplayName
	if displayName == "" {
		return fmt.Errorf("display name is required for admin user creation")
	}

	// Check if user already has a pseudonym
	existingPseudonyms, err := securePseudonymDAO.GetPseudonymsByUserID(ctx, user.UserID, input.AdminRole, "authentication")
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
		pseudonym, err = securePseudonymDAO.CreatePseudonymWithIdentityMapping(ctx, user.UserID, displayName)
		if err != nil {
			return fmt.Errorf("failed to create pseudonym for admin user: %w", err)
		}

		// Set admin capabilities on the pseudonym
		capabilities := getCapabilitiesForRole(input.AdminRole)
		capabilitiesJSON, err := json.Marshal(capabilities)
		if err != nil {
			return fmt.Errorf("failed to marshal capabilities: %w", err)
		}

		capabilitiesNull := sql.Null[types.JSON[json.RawMessage]]{}
		capabilitiesNull.Scan(capabilitiesJSON)

		// Update the pseudonym with admin capabilities
		pseudonymUpdates := &models.PseudonymSetter{
			Capabilities: &capabilitiesNull,
		}

		if err := securePseudonymDAO.UpdatePseudonym(ctx, pseudonym.PseudonymID, pseudonymUpdates); err != nil {
			return fmt.Errorf("failed to update pseudonym with admin capabilities: %w", err)
		}

		log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Created new pseudonym with admin capabilities")
	}

	log.Info().
		Int64("user_id", user.UserID).
		Str("email", input.Email).
		Str("admin_username", adminUsername).
		Str("role", input.AdminRole).
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
	fmt.Printf("   Role: %s\n", input.AdminRole)
	fmt.Printf("   MFA Enabled: %t\n", input.MFAEnabled)
	fmt.Printf("   Pseudonym ID: %s\n", pseudonym.PseudonymID)
	fmt.Printf("   Display Name: %s\n", pseudonym.DisplayName)
	if input.AdminScope != "" {
		fmt.Printf("   Admin Scope: %s\n", input.AdminScope)
	}

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
	pseudonymDAO := dao.NewSecurePseudonymDAO(db, ibe.NewIBESystemFromEnv(), dao.NewIdentityMappingDAO(db), dao.NewUserDAO(db), dao.NewRoleKeyDAO(db), dao.NewUserBlocksDAO(db))

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

	// Check if pseudonym is already a moderator
	existingModerators, err := subforumDAO.GetSubforumModerators(ctx, subforum.SubforumID)
	if err != nil {
		return fmt.Errorf("failed to get existing moderators: %w", err)
	}

	for _, mod := range existingModerators {
		if mod.PseudonymID == input.PseudonymID {
			fmt.Printf("✅ Pseudonym '%s' is already a moderator of subforum '%s'\n", input.PseudonymID, input.SubforumName)
			return nil
		}
	}

	// Add the pseudonym as a moderator
	err = subforumDAO.AddSubforumModerator(ctx, subforum.SubforumID, input.PseudonymID)
	if err != nil {
		return fmt.Errorf("failed to add moderator: %w", err)
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

// getAdminCreateInput prompts for admin user creation input
func getAdminCreateInput() *AdminCreateInput {
	input := &AdminCreateInput{}

	// Check if we're in non-interactive mode
	cmd := cobra.Command{}
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("password", "", "")
	cmd.Flags().String("role", "platform_admin", "")
	cmd.Flags().String("display-name", "", "")
	cmd.Flags().String("scope", "", "")
	cmd.Flags().Bool("mfa-enabled", true, "")
	cmd.Flags().Bool("non-interactive", false, "")

	// Parse flags from os.Args
	cmd.ParseFlags(os.Args[1:])

	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive {
		// Non-interactive mode - get values from flags
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		role, _ := cmd.Flags().GetString("role")
		displayName, _ := cmd.Flags().GetString("display-name")
		scope, _ := cmd.Flags().GetString("scope")
		mfaEnabled, _ := cmd.Flags().GetBool("mfa-enabled")

		if email == "" || password == "" {
			log.Fatal().Msg("email and password are required in non-interactive mode")
		}

		if displayName == "" {
			log.Fatal().Msg("display name is required in non-interactive mode")
		}

		input.Email = email
		input.Password = password
		input.AdminRole = role
		input.DisplayName = displayName
		input.AdminScope = scope
		input.MFAEnabled = mfaEnabled
		input.NonInteractive = true

		return input
	}

	// Interactive mode
	fmt.Println("Create Admin User")
	fmt.Println("=================")

	fmt.Print("Email: ")
	fmt.Scanln(&input.Email)

	// Get password with hidden input
	input.Password = getPasswordInput("Password: ")

	// Confirm password
	confirmPassword := getPasswordInput("Confirm Password: ")
	if input.Password != confirmPassword {
		log.Fatal().Msg("passwords do not match")
	}

	fmt.Print("Display Name (required, cannot be email): ")
	fmt.Scanln(&input.DisplayName)
	if strings.TrimSpace(input.DisplayName) == "" {
		log.Fatal().Msg("display name is required")
	}
	if strings.ToLower(strings.TrimSpace(input.DisplayName)) == strings.ToLower(strings.TrimSpace(input.Email)) {
		log.Fatal().Msg("display name cannot be the same as email address")
	}

	fmt.Print("Admin Role (platform_admin, trust_safety, legal_team) [platform_admin]: ")
	fmt.Scanln(&input.AdminRole)
	if input.AdminRole == "" {
		input.AdminRole = "platform_admin"
	}

	fmt.Print("Admin Scope (optional): ")
	fmt.Scanln(&input.AdminScope)

	fmt.Print("Enable MFA (y/n) [y]: ")
	var mfaInput string
	fmt.Scanln(&mfaInput)
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
	cmd.ParseFlags(os.Args[1:])

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
	fmt.Scanln(&input.SubforumName)

	fmt.Print("Pseudonym ID: ")
	fmt.Scanln(&input.PseudonymID)

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
	if strings.ToLower(strings.TrimSpace(input.DisplayName)) == strings.ToLower(strings.TrimSpace(input.Email)) {
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

	validRoles := []string{constants.RolePlatformAdmin, constants.RoleTrustSafety, constants.RoleLegalTeam}
	roleValid := false
	for _, role := range validRoles {
		if input.AdminRole == role {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return fmt.Errorf("invalid admin role: %s. Valid roles are: %v", input.AdminRole, validRoles)
	}

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
