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
	"github.com/matt0x6f/hashpost/internal/api"
	"github.com/matt0x6f/hashpost/internal/api/logger"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
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

	// Check if user already exists
	ctx := context.Background()
	existingUser, err := userDAO.GetUserByEmail(ctx, input.Email)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check existing user")
	}
	if existingUser != nil {
		log.Fatal().Msg("User with this email already exists")
	}

	// Hash the password
	passwordHash := hashPassword(input.Password)

	// Generate admin username if not provided
	adminUsername := input.DisplayName
	if adminUsername == "" {
		adminUsername = generateAdminUsername(input.Email)
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

	// Create the user
	user, err := userDAO.CreateUser(ctx, input.Email, passwordHash)
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

	// Create a pseudonym for the admin user
	pseudonymDAO := dao.NewPseudonymDAO(db)

	// Use display name for pseudonym (it's required)
	displayName := input.DisplayName
	if displayName == "" {
		log.Fatal().Msg("Display name is required for admin user creation")
	}

	pseudonym, err := pseudonymDAO.CreatePseudonym(ctx, user.UserID, displayName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create pseudonym for admin user")
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

	fmt.Printf("✅ Admin user created successfully!\n")
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

	fmt.Print("Display Name (optional): ")
	fmt.Scanln(&input.DisplayName)

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
	if input.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if input.DisplayName == "" {
		return fmt.Errorf("display name is required")
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

// generateAdminUsername generates an admin username from email
func generateAdminUsername(email string) string {
	// Extract username part from email
	for i, char := range email {
		if char == '@' {
			return email[:i] + "_admin"
		}
	}
	return email + "_admin"
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
