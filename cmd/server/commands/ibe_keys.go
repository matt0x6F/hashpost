package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// IBEKeyOptions defines options for IBE key generation
type IBEKeyOptions struct {
	OutputDir      string
	KeyVersion     int32
	Salt           string
	NonInteractive bool
	KeySize        int
}

// GenerateIBEKeys generates IBE keys for enhanced architecture
func GenerateIBEKeys(options *IBEKeyOptions, cfg *config.Config) error {
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

	log.Info().Msg("IBE key generation completed successfully")
	return nil
}

// initializeIBESystem initializes the IBE system with automatically generated domain keys
func initializeIBESystem(opts *IBEKeyOptions) (*ibe.IBESystem, error) {
	// Always generate new domain keys - no more master key concept
	domainMasters := make(map[string][]byte)

	// Define all domains that need keys
	domains := []string{
		ibe.DOMAIN_USER_PSEUDONYMS,
		ibe.DOMAIN_USER_CORRELATION,
		ibe.DOMAIN_MOD_CORRELATION,
		ibe.DOMAIN_ADMIN_CORRELATION,
		ibe.DOMAIN_LEGAL_CORRELATION,
	}

	// Generate a unique master key for each domain
	for _, domain := range domains {
		master := make([]byte, opts.KeySize)
		if _, err := rand.Read(master); err != nil {
			return nil, fmt.Errorf("failed to generate domain key for %s: %w", domain, err)
		}
		domainMasters[domain] = master
		log.Info().Str("domain", domain).Msg("Generated domain master key")
	}

	// Create IBE system with generated domain keys
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
		DomainMasters: domainMasters,
		KeyVersion:    opts.KeyVersion,
		Salt:          opts.Salt,
	})

	return ibeSystem, nil
}

// generateDomainKeys generates and saves domain-specific master keys
func generateDomainKeys(ibeSystem *ibe.IBESystem, outputDir string) error {
	domainKeysDir := filepath.Join(outputDir, "domains")
	if err := os.MkdirAll(domainKeysDir, 0755); err != nil {
		return fmt.Errorf("failed to create domain keys directory: %w", err)
	}

	// Save all domain masters
	if err := ibeSystem.SaveDomainMastersToDir(domainKeysDir); err != nil {
		return fmt.Errorf("failed to save domain keys: %w", err)
	}

	log.Info().
		Str("domain_keys_dir", domainKeysDir).
		Msg("Domain keys generated and saved")

	return nil
}

// generateRoleKeys generates role-specific keys with automatic domain and scope mapping
func generateRoleKeys(ibeSystem *ibe.IBESystem, outputDir string) error {
	roleKeysDir := filepath.Join(outputDir, "roles")
	if err := os.MkdirAll(roleKeysDir, 0755); err != nil {
		return fmt.Errorf("failed to create role keys directory: %w", err)
	}

	// Get all roles from constants
	roles := constants.GetAllRoles()

	// Define standard time windows for key expiration
	timeWindows := []time.Duration{
		time.Hour,           // 1 hour
		24 * time.Hour,      // 1 day
		7 * 24 * time.Hour,  // 1 week
		30 * 24 * time.Hour, // 1 month
	}

	for _, role := range roles {
		roleDir := filepath.Join(roleKeysDir, role)
		if err := os.MkdirAll(roleDir, 0755); err != nil {
			return fmt.Errorf("failed to create role directory: %w", err)
		}

		// Get scopes for this role from constants
		roleDefinition := constants.GetRoleDefinition(role)
		if roleDefinition == nil {
			log.Warn().Str("role", role).Msg("No role definition found, skipping")
			continue
		}

		for _, scope := range roleDefinition.Scopes {
			scopeDir := filepath.Join(roleDir, scope)
			if err := os.MkdirAll(scopeDir, 0755); err != nil {
				return fmt.Errorf("failed to create scope directory: %w", err)
			}

			for _, timeWindow := range timeWindows {
				// Generate time-bounded key
				roleKey := ibeSystem.GenerateTimeBoundedKey(role, scope, timeWindow)

				// Create filename with time window
				timeWindowStr := formatTimeWindow(timeWindow)
				keyPath := filepath.Join(scopeDir, fmt.Sprintf("%s.key", timeWindowStr))

				if err := saveKeyToFile(roleKey, keyPath); err != nil {
					return fmt.Errorf("failed to save role key: %w", err)
				}

				log.Info().
					Str("role", role).
					Str("scope", scope).
					Str("time_window", timeWindowStr).
					Str("key_path", keyPath).
					Str("key_hash", hex.EncodeToString(roleKey)).
					Msg("Role key generated and saved")
			}
		}
	}

	return nil
}

// generateTestKeys generates test keys for development and testing
func generateTestKeys(ibeSystem *ibe.IBESystem, outputDir string) error {
	testKeysDir := filepath.Join(outputDir, "test")
	if err := os.MkdirAll(testKeysDir, 0755); err != nil {
		return fmt.Errorf("failed to create test keys directory: %w", err)
	}

	// Generate test pseudonyms
	testPseudonyms := []struct {
		userID  int64
		context string
	}{
		{1, "test_user_1"},
		{2, "test_user_2"},
		{3, "test_moderator"},
		{4, "test_admin"},
	}

	for _, test := range testPseudonyms {
		// Generate legacy pseudonym (version 1)
		pseudonymV1 := ibeSystem.CreateEnhancedPseudonym(test.userID, test.context)

		// Save test pseudonym
		testFile := filepath.Join(testKeysDir, fmt.Sprintf("pseudonym_%d_v1.txt", test.userID))
		if err := saveStringToFile(pseudonymV1, testFile); err != nil {
			return fmt.Errorf("failed to save test pseudonym: %w", err)
		}

		log.Info().
			Int64("user_id", test.userID).
			Str("context", test.context).
			Str("pseudonym_v1", pseudonymV1).
			Str("test_file", testFile).
			Msg("Test pseudonym generated")
	}

	// Generate test role keys for common roles and scopes
	testRoles := []string{"user", "moderator", "platform_admin"}
	testScopes := []string{"authentication", "correlation"}

	for _, role := range testRoles {
		for _, scope := range testScopes {
			testKey := ibeSystem.GenerateTestRoleKey(role, scope)
			testKeyPath := filepath.Join(testKeysDir, fmt.Sprintf("test_%s_%s.key", role, scope))

			if err := saveKeyToFile(testKey, testKeyPath); err != nil {
				return fmt.Errorf("failed to save test role key: %w", err)
			}

			log.Info().
				Str("role", role).
				Str("scope", scope).
				Str("test_key_path", testKeyPath).
				Str("test_key_hash", hex.EncodeToString(testKey)).
				Msg("Test role key generated")
		}
	}

	return nil
}

// saveIBEConfiguration saves the IBE system configuration
func saveIBEConfiguration(ibeSystem *ibe.IBESystem, outputDir string) error {
	configPath := filepath.Join(outputDir, "ibe_config.json")

	// Get role definitions for configuration
	roleDefinitions := constants.GetRoleDefinitions()
	roleConfig := make(map[string]interface{})

	for _, roleDef := range roleDefinitions {
		roleConfig[roleDef.RoleName] = map[string]interface{}{
			"scopes":       roleDef.Scopes,
			"capabilities": roleDef.Capabilities,
		}
	}

	config := map[string]interface{}{
		"key_version": ibeSystem.GetKeyVersion(),
		"salt":        ibeSystem.GetSalt(),
		"domains": map[string]string{
			"user_pseudonyms":   ibe.DOMAIN_USER_PSEUDONYMS,
			"user_correlation":  ibe.DOMAIN_USER_CORRELATION,
			"mod_correlation":   ibe.DOMAIN_MOD_CORRELATION,
			"admin_correlation": ibe.DOMAIN_ADMIN_CORRELATION,
			"legal_correlation": ibe.DOMAIN_LEGAL_CORRELATION,
		},
		"roles":        roleConfig,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Info().
		Str("config_path", configPath).
		Msg("IBE configuration saved")

	return nil
}

// Helper functions

func formatTimeWindow(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/24/7))
	default:
		return fmt.Sprintf("%dm", int(d.Hours()/24/30))
	}
}

func saveKeyToFile(key []byte, path string) error {
	hexKey := hex.EncodeToString(key)
	return os.WriteFile(path, []byte(hexKey), 0600)
}

func saveStringToFile(content string, path string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// NewGenerateIBEKeysCommand creates and returns the generate-ibe-keys command
func NewGenerateIBEKeysCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-ibe-keys",
		Short: "Generate IBE keys for enhanced architecture",
		Long:  "Generate Identity-Based Encryption keys with automatic domain separation and role-scope mapping. This command automatically generates all necessary domain keys and role keys based on the system's predefined mappings.",
		Run: func(cmd *cobra.Command, args []string) {
			// This command generates IBE keys from scratch - no need to initialize existing system

			// Parse command line flags
			outputDir, _ := cmd.Flags().GetString("output-dir")
			keyVersion, _ := cmd.Flags().GetInt32("key-version")
			salt, _ := cmd.Flags().GetString("salt")
			nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
			keySize, _ := cmd.Flags().GetInt("key-size")

			// Use configuration salt if not specified via command line
			if salt == "" {
				salt = cfg.IBE.Salt
			}

			// Create IBE key options
			ibeOptions := &IBEKeyOptions{
				OutputDir:      outputDir,
				KeyVersion:     keyVersion,
				Salt:           salt,
				NonInteractive: nonInteractive,
				KeySize:        keySize,
			}

			// Generate IBE keys
			if err := GenerateIBEKeys(ibeOptions, cfg); err != nil {
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
	cmd.Flags().String("output-dir", "./keys", "Output directory for generated keys")
	cmd.Flags().Int32("key-version", 1, "Key version to generate")
	cmd.Flags().String("salt", "", "Salt for fingerprint generation (defaults to config value)")
	cmd.Flags().Bool("non-interactive", false, "Non-interactive mode")
	cmd.Flags().Int("key-size", 32, "Key size in bytes (default 32, i.e., 256 bits)")

	return cmd
}
