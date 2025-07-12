package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
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
			if err := commands.CreateAdminUser(); err != nil {
				log.Fatal().Err(err).Msg("Failed to create admin user")
			}
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
			if err := commands.SetupRoles(); err != nil {
				log.Fatal().Err(err).Msg("Failed to setup roles")
			}
		}),
	}

	cli.Root().AddCommand(setupRolesCmd)

	// Add set-moderator subcommand
	setModeratorCmd := &cobra.Command{
		Use:   "set-moderator",
		Short: "Set a pseudonym as a forum moderator",
		Long:  "Set a pseudonym as a moderator of a specific subforum",
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, options *Options) {
			if err := commands.SetModerator(); err != nil {
				log.Fatal().Err(err).Msg("Failed to set moderator")
			}
		}),
	}

	// Add flags for set-moderator command
	setModeratorCmd.Flags().String("subforum", "", "Name of the subforum")
	setModeratorCmd.Flags().String("pseudonym", "", "Pseudonym ID to set as moderator")
	setModeratorCmd.Flags().Bool("non-interactive", false, "Non-interactive mode (requires all flags)")

	cli.Root().AddCommand(setModeratorCmd)
	fmt.Println("DEBUG: Added set-moderator command")

	// Add generate-ibe-keys subcommand
	generateIBEKeysCmd := &cobra.Command{
		Use:   "generate-ibe-keys",
		Short: "Generate IBE keys for enhanced architecture",
		Long:  "Generate Identity-Based Encryption keys with domain separation and time-bounded access",
		Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, options *Options) {
			// Parse command line flags
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
