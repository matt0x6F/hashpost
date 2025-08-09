package commands

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// IBEKeyOptions defines the options for IBE key generation
type IBEKeyOptions struct {
	OutputDir      string `doc:"Output directory for generated keys" json:"output_dir"`
	KeyVersion     int    `doc:"Key version to generate" json:"key_version" default:"1"`
	Salt           string `doc:"Salt for fingerprint generation" json:"salt" default:"fingerprint_salt_v1"`
	KeySize        int    `doc:"Key size in bytes (default 32, i.e., 256 bits)" json:"key_size" default:"32"`
	NonInteractive bool   `doc:"Non-interactive mode" json:"non_interactive"`
}

// GenerateIBEKeys generates IBE keys for the enhanced architecture
// This command automatically generates all necessary domain keys and role keys
// based on the system's predefined role-domain-scope mappings
func GenerateIBEKeys(opts *IBEKeyOptions) error {
	// Load configuration (for future use)
	_, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Initialize IBE system with automatically generated domain keys
	ibeSystem, err := initializeIBESystem(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize IBE system: %w", err)
	}

	// Generate domain keys (all domains are automatically generated)
	if err := generateDomainKeys(ibeSystem, opts.OutputDir); err != nil {
		return fmt.Errorf("failed to generate domain keys: %w", err)
	}

	// Generate role keys with automatic domain and scope mapping
	if err := generateRoleKeys(ibeSystem, opts.OutputDir); err != nil {
		return fmt.Errorf("failed to generate role keys: %w", err)
	}

	// Generate test keys for development
	if err := generateTestKeys(ibeSystem, opts.OutputDir); err != nil {
		return fmt.Errorf("failed to generate test keys: %w", err)
	}

	// Save configuration
	if err := saveIBEConfiguration(ibeSystem, opts.OutputDir); err != nil {
		return fmt.Errorf("failed to save IBE configuration: %w", err)
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
