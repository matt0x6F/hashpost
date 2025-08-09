package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/matt0x6f/hashpost/cmd/server/commands"
	"github.com/matt0x6f/hashpost/internal/api"
	"github.com/matt0x6f/hashpost/internal/api/logger"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Options defines the CLI options
type Options struct {
	Debug bool   `doc:"Enable debug logging"`
	Host  string `doc:"Hostname to listen on."`
	Port  int    `doc:"Port to listen on." short:"p" default:"8888"`
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

	// Create the root command
	rootCmd := &cobra.Command{
		Use:     "hashpost",
		Version: "1.0.0",
		Short:   "HashPost - A modern forum platform",
		Long:    "HashPost is a modern forum platform with enhanced security and privacy features.",
	}

	// Add global flags
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().String("host", "localhost", "Hostname to listen on")
	rootCmd.PersistentFlags().IntP("port", "p", 8888, "Port to listen on")

	// Create server command with humacli
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the HashPost server",
		Long:  "Start the HashPost server with the specified configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Create the CLI for server mode
			cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
				// This function only runs when starting the server
				// IBE system is initialized here only for server mode
				ibeSystem := ibe.NewIBESystemFromEnv()
				log.Info().Str("ibe_master_key", hex.EncodeToString(ibeSystem.GetMasterSecret())).Str("ibe_salt", ibeSystem.GetSalt()).Int("ibe_key_version", ibeSystem.GetKeyVersion()).Msg("IBE system configuration (server startup)")

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

			// Run the server
			cli.Run()
		},
	}

	// Set the default command to start the server
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		serverCmd.Run(cmd, args)
		return nil
	}

	// Add server command
	rootCmd.AddCommand(serverCmd)

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
			// This command needs IBE - initialize it
			if err := initializeIBEForCommand(); err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize IBE system")
			}
			if err := commands.SetupRoles(); err != nil {
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
			if err := commands.ListRoles(); err != nil {
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
			if err := commands.ListRoleKeys(); err != nil {
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
			if err := commands.RotateRoleKeys(roleName, force); err != nil {
				log.Fatal().Err(err).Msg("Failed to rotate role keys")
			}
		},
	}
	rotateRoleKeysCmd.Flags().String("role", "", "Specific role to rotate keys for (optional, rotates all roles if not specified)")
	rotateRoleKeysCmd.Flags().Bool("force", false, "Force rotation even if keys already exist")
	rolesCmd.AddCommand(rotateRoleKeysCmd)

	// Register the roles group
	rootCmd.AddCommand(rolesCmd)

	// Add create-admin subcommand
	createAdminCmd := &cobra.Command{
		Use:   "create-admin",
		Short: "Create a new admin user",
		Long:  "Create a new admin user with specified role and capabilities",
		Run: func(cmd *cobra.Command, args []string) {
			// This command needs IBE - initialize it
			if err := initializeIBEForCommand(); err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize IBE system")
			}
			if err := commands.CreateAdminUser(); err != nil {
				log.Fatal().Err(err).Msg("Failed to create admin user")
			}
		},
	}

	// Add flags for create-admin command
	createAdminCmd.Flags().String("email", "", "Email address for the admin user")
	createAdminCmd.Flags().String("password", "", "Password for the admin user")
	createAdminCmd.Flags().String("role", "platform_admin", "Admin role (platform_admin, trust_safety, legal_team)")
	createAdminCmd.Flags().String("display-name", "", "Display name for the admin user")
	createAdminCmd.Flags().String("scope", "", "Admin scope (optional)")
	createAdminCmd.Flags().Bool("mfa-enabled", true, "Enable MFA for the admin user")
	createAdminCmd.Flags().Bool("non-interactive", false, "Non-interactive mode (requires all flags)")

	rootCmd.AddCommand(createAdminCmd)

	// Add set-moderator subcommand
	setModeratorCmd := &cobra.Command{
		Use:   "set-moderator",
		Short: "Set a pseudonym as a forum moderator",
		Long:  "Set a pseudonym as a moderator of a specific subforum",
		Run: func(cmd *cobra.Command, args []string) {
			// This command needs IBE - initialize it
			if err := initializeIBEForCommand(); err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize IBE system")
			}
			if err := commands.SetModerator(); err != nil {
				log.Fatal().Err(err).Msg("Failed to set moderator")
			}
		},
	}

	// Add flags for set-moderator command
	setModeratorCmd.Flags().String("subforum", "", "Name of the subforum")
	setModeratorCmd.Flags().String("pseudonym", "", "Pseudonym ID to set as moderator")
	setModeratorCmd.Flags().Bool("non-interactive", false, "Non-interactive mode (requires all flags)")

	rootCmd.AddCommand(setModeratorCmd)

	// Add delete-user subcommand
	deleteUserCmd := &cobra.Command{
		Use:   "delete-user",
		Short: "Delete a user and all associated data",
		Long:  "Delete a user account and all associated data including pseudonyms, role keys, and tokens",
		Run: func(cmd *cobra.Command, args []string) {
			// This command needs IBE - initialize it
			if err := initializeIBEForCommand(); err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize IBE system")
			}
			if err := commands.DeleteUser(); err != nil {
				log.Fatal().Err(err).Msg("Failed to delete user")
			}
		},
	}

	// Add flags for delete-user command
	deleteUserCmd.Flags().String("email", "", "Email address of the user to delete")
	deleteUserCmd.Flags().Bool("force", false, "Force deletion without confirmation")
	deleteUserCmd.Flags().Bool("non-interactive", false, "Non-interactive mode (requires all flags)")

	rootCmd.AddCommand(deleteUserCmd)

	// Add update-admin subcommand
	updateAdminCmd := &cobra.Command{
		Use:   "update-admin",
		Short: "Update an admin user and fix their pseudonym mappings",
		Long:  "Update an existing admin user's role and optionally fix missing identity mappings",
		Run: func(cmd *cobra.Command, args []string) {
			// This command needs IBE - initialize it
			if err := initializeIBEForCommand(); err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize IBE system")
			}
			if err := commands.UpdateAdminUserWithCommand(cmd); err != nil {
				log.Fatal().Err(err).Msg("Failed to update admin user")
			}
		},
	}

	// Add flags for update-admin command
	updateAdminCmd.Flags().String("email", "", "Email address of the admin user to update")
	updateAdminCmd.Flags().String("role", "platform_admin", "Admin role (platform_admin, trust_safety, legal_team)")
	updateAdminCmd.Flags().Bool("fix-mappings", false, "Fix missing identity mappings for the user's pseudonyms")
	updateAdminCmd.Flags().Bool("non-interactive", false, "Non-interactive mode (requires all flags)")

	rootCmd.AddCommand(updateAdminCmd)

	// Add generate-ibe-keys subcommand
	generateIBEKeysCmd := &cobra.Command{
		Use:   "generate-ibe-keys",
		Short: "Generate IBE keys for enhanced architecture",
		Long:  "Generate Identity-Based Encryption keys with automatic domain separation and role-scope mapping. This command automatically generates all necessary domain keys and role keys based on the system's predefined mappings.",
		Run: func(cmd *cobra.Command, args []string) {
			// This command generates IBE keys from scratch - no need to initialize existing system

			// Parse command line flags
			outputDir, _ := cmd.Flags().GetString("output-dir")
			keyVersion, _ := cmd.Flags().GetInt("key-version")
			salt, _ := cmd.Flags().GetString("salt")
			nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
			keySize, _ := cmd.Flags().GetInt("key-size")

			// Create IBE key options
			ibeOptions := &commands.IBEKeyOptions{
				OutputDir:      outputDir,
				KeyVersion:     keyVersion,
				Salt:           salt,
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
			fmt.Println("   Generated domain keys for all domains")
			fmt.Println("   Generated role keys for all roles with appropriate scopes")
		},
	}

	// Add flags for generate-ibe-keys command
	generateIBEKeysCmd.Flags().String("output-dir", "./keys", "Output directory for generated keys")
	generateIBEKeysCmd.Flags().Int("key-version", 1, "Key version to generate")
	generateIBEKeysCmd.Flags().String("salt", "fingerprint_salt_v1", "Salt for fingerprint generation")
	generateIBEKeysCmd.Flags().Bool("non-interactive", false, "Non-interactive mode")
	generateIBEKeysCmd.Flags().Int("key-size", 32, "Key size in bytes (default 32, i.e., 256 bits)")

	rootCmd.AddCommand(generateIBEKeysCmd)

	// Add key-rotation subcommand group
	keyRotationCmd := &cobra.Command{
		Use:   "key-rotation",
		Short: "Manage key rotation migrations",
		Long:  "Commands for managing IBE key rotation migrations with resumable functionality",
	}

	// Add key-rotation subcommands
	startKeyRotationCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new key rotation migration",
		Long:  "Start a new key rotation migration for a specific domain and key versions",
		Run: func(cmd *cobra.Command, args []string) {
			// This command needs IBE - initialize it
			if err := initializeIBEForCommand(); err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize IBE system")
			}

			// Parse command line flags
			domain, _ := cmd.Flags().GetString("domain")
			oldKeyVersion, _ := cmd.Flags().GetInt("old-version")
			newKeyVersion, _ := cmd.Flags().GetInt("new-version")
			createdBy, _ := cmd.Flags().GetInt64("created-by")

			if err := commands.StartKeyRotationMigration(domain, oldKeyVersion, newKeyVersion, createdBy); err != nil {
				log.Fatal().Err(err).Msg("Failed to start key rotation migration")
			}
		},
	}
	startKeyRotationCmd.Flags().String("domain", "", "Domain to migrate (e.g., user_correlation, admin_correlation)")
	startKeyRotationCmd.Flags().Int("old-version", 0, "Old key version to migrate from")
	startKeyRotationCmd.Flags().Int("new-version", 0, "New key version to migrate to")
	startKeyRotationCmd.Flags().Int64("created-by", 1, "User ID who created the migration")
	startKeyRotationCmd.MarkFlagRequired("domain")
	startKeyRotationCmd.MarkFlagRequired("old-version")
	startKeyRotationCmd.MarkFlagRequired("new-version")

	statusKeyRotationCmd := &cobra.Command{
		Use:   "status",
		Short: "Check key rotation migration status",
		Long:  "Display the current status of key rotation migrations",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE
			if err := commands.GetKeyRotationStatus(); err != nil {
				log.Fatal().Err(err).Msg("Failed to get key rotation status")
			}
		},
	}

	pauseKeyRotationCmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause a running key rotation migration",
		Long:  "Pause a currently running key rotation migration",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE
			migrationID, _ := cmd.Flags().GetInt64("migration-id")
			if err := commands.PauseKeyRotationMigration(migrationID); err != nil {
				log.Fatal().Err(err).Msg("Failed to pause key rotation migration")
			}
		},
	}
	pauseKeyRotationCmd.Flags().Int64("migration-id", 0, "ID of the migration to pause")
	pauseKeyRotationCmd.MarkFlagRequired("migration-id")

	resumeKeyRotationCmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a paused key rotation migration",
		Long:  "Resume a previously paused key rotation migration",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE
			migrationID, _ := cmd.Flags().GetInt64("migration-id")
			if err := commands.ResumeKeyRotationMigration(migrationID); err != nil {
				log.Fatal().Err(err).Msg("Failed to resume key rotation migration")
			}
		},
	}
	resumeKeyRotationCmd.Flags().Int64("migration-id", 0, "ID of the migration to resume")
	resumeKeyRotationCmd.MarkFlagRequired("migration-id")

	recoverKeyRotationCmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover a failed key rotation migration",
		Long:  "Attempt to recover a failed key rotation migration",
		Run: func(cmd *cobra.Command, args []string) {
			// This command needs IBE - initialize it
			if err := initializeIBEForCommand(); err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize IBE system")
			}

			migrationID, _ := cmd.Flags().GetInt64("migration-id")
			if err := commands.RecoverKeyRotationMigration(migrationID); err != nil {
				log.Fatal().Err(err).Msg("Failed to recover key rotation migration")
			}
		},
	}
	recoverKeyRotationCmd.Flags().Int64("migration-id", 0, "ID of the migration to recover")
	recoverKeyRotationCmd.MarkFlagRequired("migration-id")

	listKeyRotationCmd := &cobra.Command{
		Use:   "list",
		Short: "List all key rotation migrations",
		Long:  "Display all key rotation migrations with their current status",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE
			if err := commands.ListKeyRotationMigrations(); err != nil {
				log.Fatal().Err(err).Msg("Failed to list key rotation migrations")
			}
		},
	}

	// Add all key-rotation subcommands
	keyRotationCmd.AddCommand(
		startKeyRotationCmd,
		statusKeyRotationCmd,
		pauseKeyRotationCmd,
		resumeKeyRotationCmd,
		recoverKeyRotationCmd,
		listKeyRotationCmd,
	)

	// Register the key-rotation group
	rootCmd.AddCommand(keyRotationCmd)

	// Add openapi subcommand
	rootCmd.AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec",
		Run: func(cmd *cobra.Command, args []string) {
			// This command doesn't need IBE - use minimal server
			server := api.NewServerForOpenAPI()
			b, err := server.API.OpenAPI().YAML()
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to generate OpenAPI spec")
			}
			fmt.Println(string(b))
		},
	})

	// Run the CLI
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

// initializeIBEForCommand initializes the IBE system for commands that need it
// This function is called before running commands that require IBE functionality
func initializeIBEForCommand() error {
	// Check if IBE environment variables are set
	domainKeysDir := os.Getenv("IBE_DOMAIN_KEYS_DIR")
	if domainKeysDir == "" {
		domainKeysDir = "./keys/domains"
	}

	// Try to initialize IBE system
	ibeSystem, err := ibe.NewIBESystemFromConfig(domainKeysDir, 1, "fingerprint_salt_v1")
	if err != nil {
		return fmt.Errorf("failed to initialize IBE system: %w", err)
	}

	log.Info().Str("ibe_master_key", hex.EncodeToString(ibeSystem.GetMasterSecret())).Str("ibe_salt", ibeSystem.GetSalt()).Int("ibe_key_version", ibeSystem.GetKeyVersion()).Msg("IBE system initialized for command")
	return nil
}
