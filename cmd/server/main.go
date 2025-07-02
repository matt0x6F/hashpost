package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/matt0x6f/hashpost/cmd/server/commands"
	"github.com/matt0x6f/hashpost/internal/api"
	"github.com/matt0x6f/hashpost/internal/api/logger"
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

// Options defines the CLI options
type Options struct {
	Debug bool   `doc:"Enable debug logging"`
	Host  string `doc:"Hostname to listen on."`
	Port  int    `doc:"Port to listen on." short:"p" default:"8888"`
}

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

func main() {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		// If config loading fails, initialize logger with default level
		logger.Init()
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize the logger with the configured level
	logger.InitWithLevel(cfg.Logging.Level)

	// Initialize IBE system and identity mapping DAO
	ibeSystem := ibe.NewIBESystemFromEnv()
	log.Info().Str("ibe_master_key", hex.EncodeToString(ibeSystem.GetMasterSecret())).Str("ibe_salt", ibeSystem.GetSalt()).Int("ibe_key_version", ibeSystem.GetKeyVersion()).Msg("IBE system configuration (CLI/server startup)")

	// Create the CLI
	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		// Create the API server with all middleware and routes
		server := api.NewServer()

		// Create the HTTP server with graceful shutdown
		httpServer := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler: server.GetHandler(),
		}

		hooks.OnStart(func() {
			log.Info().Str("addr", httpServer.Addr).Msg("Server listening")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Error starting server")
			}
		})

		hooks.OnStop(func() {
			log.Info().Msg("Start shutdown")
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(ctx); err != nil {
				log.Error().Err(err).Msg("Could not stop server gracefully")
				if err := httpServer.Close(); err != nil {
					log.Fatal().Err(err).Msg("Could not force close server")
				}
			}
			log.Info().Msg("Server stopped")
		})
	})

	// Set app name and version
	cmd := cli.Root()
	cmd.Use = "hashpost"
	cmd.Version = "1.0.0"

	// Add create-admin subcommand
	createAdminCmd := &cobra.Command{
		Use:   "create-admin",
		Short: "Create a new admin user",
		Long:  "Create a new admin user with specified role and capabilities",
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, options *Options) {
			createAdminUser(options)
		}),
	}

	// Add flags for create-admin command
	createAdminCmd.Flags().String("email", "", "Email address for the admin user")
	createAdminCmd.Flags().String("password", "", "Password for the admin user")
	createAdminCmd.Flags().String("role", "platform_admin", "Admin role (platform_admin, trust_safety, legal_team)")
	createAdminCmd.Flags().String("display-name", "", "Display name for the admin user")
	createAdminCmd.Flags().String("scope", "", "Admin scope (optional)")
	createAdminCmd.Flags().Bool("mfa-enabled", true, "Enable MFA for the admin user")
	createAdminCmd.Flags().Bool("non-interactive", false, "Non-interactive mode (requires all flags)")

	cli.Root().AddCommand(createAdminCmd)

	// Add setup-roles subcommand
	setupRolesCmd := &cobra.Command{
		Use:   "setup-roles",
		Short: "Setup role keys for all roles",
		Long:  "Create the necessary role keys for all roles: user, moderator, subforum_owner, platform_admin, trust_safety, and legal_team",
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, options *Options) {
			setupRoles(options)
		}),
	}

	cli.Root().AddCommand(setupRolesCmd)

	// Add generate-ibe-keys subcommand
	generateIBEKeysCmd := &cobra.Command{
		Use:   "generate-ibe-keys",
		Short: "Generate IBE keys for enhanced architecture",
		Long:  "Generate Identity-Based Encryption keys with domain separation and time-bounded access",
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, options *Options) {
			generateIBEKeys(options)
		}),
	}

	// Add flags for generate-ibe-keys command
	generateIBEKeysCmd.Flags().String("output-dir", "./keys", "Output directory for generated keys")
	generateIBEKeysCmd.Flags().Int("key-version", 1, "Key version to generate")
	generateIBEKeysCmd.Flags().String("salt", "fingerprint_salt_v1", "Salt for fingerprint generation")
	generateIBEKeysCmd.Flags().String("master-key-path", "", "Path to existing master key file (optional)")
	generateIBEKeysCmd.Flags().Bool("generate-new", false, "Generate new master key")
	generateIBEKeysCmd.Flags().String("domains", "", "Comma-separated list of domains to generate keys for")
	generateIBEKeysCmd.Flags().String("time-windows", "", "Comma-separated list of time windows (e.g., 1h,24h,7d,30d)")
	generateIBEKeysCmd.Flags().String("roles", "", "Comma-separated list of roles to generate keys for")
	generateIBEKeysCmd.Flags().String("scopes", "", "Comma-separated list of scopes to generate keys for")
	generateIBEKeysCmd.Flags().Bool("non-interactive", false, "Non-interactive mode")
	generateIBEKeysCmd.Flags().Int("key-size", 32, "Key size in bytes (default 32, i.e., 256 bits)")

	cli.Root().AddCommand(generateIBEKeysCmd)

	// Add openapi subcommand
	cli.Root().AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec",
		Run: func(cmd *cobra.Command, args []string) {
			// Create a temporary server to get the API
			server := api.NewServer()
			b, err := server.API.OpenAPI().YAML()
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to generate OpenAPI spec")
			}
			fmt.Println(string(b))
		},
	})

	// Run the CLI
	cli.Run()
}

// createAdminUser creates a new admin user
func createAdminUser(opts *Options) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Get admin creation input
	input := getAdminCreateInput()

	// Validate input
	if err := validateAdminInput(input); err != nil {
		log.Fatal().Err(err).Msg("Invalid input")
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
		log.Fatal().Err(err).Msg("Failed to check existing user")
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
		roles := []string{input.AdminRole}
		capabilities := getCapabilitiesForRole(input.AdminRole)

		// Convert to JSON
		rolesJSON, err := json.Marshal(roles)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal roles")
		}

		capabilitiesJSON, err := json.Marshal(capabilities)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal capabilities")
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
			Capabilities:      &capabilitiesNull,
			AdminUsername:     &adminUsernameNull,
			AdminPasswordHash: &adminPasswordHashNull,
			MfaEnabled:        &mfaEnabledNull,
			AdminScope:        &adminScopeNull,
		}

		if err := userDAO.UpdateUser(ctx, existingUser.UserID, updates); err != nil {
			log.Fatal().Err(err).Msg("Failed to update user with admin fields")
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
		roles := []string{input.AdminRole}
		capabilities := getCapabilitiesForRole(input.AdminRole)

		// Convert to JSON
		rolesJSON, err := json.Marshal(roles)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal roles")
		}

		capabilitiesJSON, err := json.Marshal(capabilities)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal capabilities")
		}

		// Create the user
		user, err = userDAO.CreateUser(ctx, input.Email, passwordHash)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create user")
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
			Capabilities:      &capabilitiesNull,
			AdminUsername:     &adminUsernameNull,
			AdminPasswordHash: &adminPasswordHashNull,
			MfaEnabled:        &mfaEnabledNull,
			AdminScope:        &adminScopeNull,
		}

		if err := userDAO.UpdateUser(ctx, user.UserID, updates); err != nil {
			log.Fatal().Err(err).Msg("Failed to update user with admin fields")
		}

		log.Info().Int64("user_id", user.UserID).Msg("New admin user created")
	}

	// Ensure default role keys for the admin user
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	if err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, user.UserID); err != nil {
		log.Fatal().Err(err).Msg("Failed to create default role keys for admin user")
	}

	// Create a pseudonym for the admin user with identity mapping
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	securePseudonymDAO := dao.NewSecurePseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)

	// Use display name for pseudonym (it's required)
	displayName := input.DisplayName
	if displayName == "" {
		log.Fatal().Msg("Display name is required for admin user creation")
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
			log.Fatal().Err(err).Msg("Failed to create pseudonym for admin user")
		}
		log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Created new pseudonym")
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
			log.Fatal().Msg("Email and password are required in non-interactive mode")
		}

		if displayName == "" {
			log.Fatal().Msg("Display name is required in non-interactive mode")
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
		log.Fatal().Msg("Passwords do not match")
	}

	fmt.Print("Display Name (required, cannot be email): ")
	fmt.Scanln(&input.DisplayName)
	if strings.TrimSpace(input.DisplayName) == "" {
		log.Fatal().Msg("Display name is required")
	}
	if strings.ToLower(strings.TrimSpace(input.DisplayName)) == strings.ToLower(strings.TrimSpace(input.Email)) {
		log.Fatal().Msg("Display name cannot be the same as email address")
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

	validRoles := []string{"platform_admin", "trust_safety", "legal_team"}
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

// hashPassword hashes a password using SHA-256
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// getCapabilitiesForRole returns the capabilities for a given admin role
func getCapabilitiesForRole(role string) []string {
	roleCapabilities := map[string][]string{
		"platform_admin": {
			"system_admin",
			"user_management",
			"correlate_identities",
			"access_private_subforums",
			"cross_platform_access",
			"system_moderation",
			"create_subforum",
		},
		"trust_safety": {
			"correlate_identities",
			"cross_platform_access",
			"system_moderation",
			"harassment_investigation",
		},
		"legal_team": {
			"correlate_identities",
			"legal_compliance",
			"court_orders",
			"cross_platform_access",
		},
	}

	if caps, exists := roleCapabilities[role]; exists {
		return caps
	}
	return []string{}
}

// setupRoles creates the necessary role keys for all admin roles
func setupRoles(opts *Options) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
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
			log.Fatal().Err(err).Msg("Failed to create system user for role key creation")
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
}

// generateIBEKeys generates IBE keys for the enhanced architecture
func generateIBEKeys(opts *Options) {
	// Parse command line flags
	cmd := cobra.Command{}
	cmd.Flags().String("output-dir", "./keys", "")
	cmd.Flags().Int("key-version", 1, "")
	cmd.Flags().String("salt", "fingerprint_salt_v1", "")
	cmd.Flags().String("domain-keys-dir", "", "")
	cmd.Flags().Bool("generate-new", false, "")
	cmd.Flags().String("domains", "", "")
	cmd.Flags().String("time-windows", "", "")
	cmd.Flags().String("roles", "", "")
	cmd.Flags().String("scopes", "", "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.Flags().Int("key-size", 32, "Key size in bytes (default 32, i.e., 256 bits)")

	// Parse flags from os.Args
	cmd.ParseFlags(os.Args[1:])

	// Get flag values
	outputDir, _ := cmd.Flags().GetString("output-dir")
	keyVersion, _ := cmd.Flags().GetInt("key-version")
	salt, _ := cmd.Flags().GetString("salt")
	domainKeysDir, _ := cmd.Flags().GetString("domain-keys-dir")
	generateNew, _ := cmd.Flags().GetBool("generate-new")
	domains, _ := cmd.Flags().GetString("domains")
	timeWindows, _ := cmd.Flags().GetString("time-windows")
	roles, _ := cmd.Flags().GetString("roles")
	scopes, _ := cmd.Flags().GetString("scopes")
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	keySize, _ := cmd.Flags().GetInt("key-size")

	// Create IBE key options
	ibeOptions := &commands.IBEKeyOptions{
		OutputDir:      outputDir,
		KeyVersion:     keyVersion,
		Salt:           salt,
		DomainKeysDir:  domainKeysDir,
		GenerateNew:    generateNew,
		Domains:        domains,
		TimeWindows:    timeWindows,
		Roles:          roles,
		Scopes:         scopes,
		NonInteractive: nonInteractive,
		KeySize:        keySize,
	}

	// Generate IBE keys
	if err := commands.GenerateIBEKeys(ibeOptions); err != nil {
		log.Fatal().Err(err).Msg("Failed to generate IBE keys")
	}

	fmt.Println("✅ IBE keys generated successfully!")
	fmt.Printf("   Output directory: %s\n", outputDir)
	fmt.Printf("   Key version: %d\n", keyVersion)
	fmt.Printf("   Salt: %s\n", salt)
	if generateNew {
		fmt.Println("   Generated new domain keys")
	}
	if domainKeysDir != "" {
		fmt.Printf("   Used existing domain keys: %s\n", domainKeysDir)
	}
}
