package commands

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// IBEKeyOptions defines the options for IBE key generation
type IBEKeyOptions struct {
	OutputDir      string `doc:"Output directory for generated keys" json:"output_dir"`
	KeyVersion     int    `doc:"Key version to generate" json:"key_version" default:"1"`
	Salt           string `doc:"Salt for fingerprint generation" json:"salt" default:"fingerprint_salt_v1"`
	DomainKeysDir  string `doc:"Path to existing domain keys directory (optional)" json:"domain_keys_dir"`
	GenerateNew    bool   `doc:"Generate new domain keys" json:"generate_new"`
	Domains        string `doc:"Comma-separated list of domains to generate keys for" json:"domains"`
	TimeWindows    string `doc:"Comma-separated list of time windows (e.g., 1h,24h,7d,30d)" json:"time_windows"`
	Roles          string `doc:"Comma-separated list of roles to generate keys for" json:"roles"`
	Scopes         string `doc:"Comma-separated list of scopes to generate keys for" json:"scopes"`
	NonInteractive bool   `doc:"Non-interactive mode" json:"non_interactive"`
	KeySize        int    `doc:"Key size in bytes (default 32, i.e., 256 bits)" json:"key_size" default:"32"`
}

// GenerateIBEKeys generates IBE keys for the enhanced architecture
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

	// Initialize IBE system
	ibeSystem, err := initializeIBESystem(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize IBE system: %w", err)
	}

	// Parse options
	timeWindows := parseTimeWindows(opts.TimeWindows)
	roles := parseRoles(opts.Roles)
	scopes := parseScopes(opts.Scopes)

	// Generate master key if requested
	if opts.GenerateNew {
		if err := generateMasterKey(ibeSystem, opts.OutputDir); err != nil {
			return fmt.Errorf("failed to generate master key: %w", err)
		}
	}

	// Generate domain keys
	if err := generateDomainKeys(ibeSystem, opts.OutputDir); err != nil {
		return fmt.Errorf("failed to generate domain keys: %w", err)
	}

	// Generate role keys
	if err := generateRoleKeys(ibeSystem, roles, scopes, timeWindows, opts.OutputDir); err != nil {
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

// initializeIBESystem initializes the IBE system with the given options
func initializeIBESystem(opts *IBEKeyOptions) (*ibe.IBESystem, error) {
	var domainMasters map[string][]byte
	var err error

	if opts.DomainKeysDir != "" {
		// Load existing domain keys
		domainMasters, err = ibe.LoadDomainMastersFromDir(opts.DomainKeysDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load domain keys from %s: %w", opts.DomainKeysDir, err)
		}
		log.Info().Str("domain_keys_dir", opts.DomainKeysDir).Msg("Loaded existing domain keys")
	} else {
		// Generate new domain keys
		domainMasters = make(map[string][]byte)
		domains := []string{
			ibe.DOMAIN_USER_PSEUDONYMS,
			ibe.DOMAIN_USER_CORRELATION,
			ibe.DOMAIN_MOD_CORRELATION,
			ibe.DOMAIN_ADMIN_CORRELATION,
			ibe.DOMAIN_LEGAL_CORRELATION,
		}

		for _, domain := range domains {
			master := make([]byte, opts.KeySize)
			if _, err := rand.Read(master); err != nil {
				return nil, fmt.Errorf("failed to generate domain key for %s: %w", domain, err)
			}
			domainMasters[domain] = master
		}
		log.Info().Msg("Generated new domain keys")
	}

	// Create IBE system with options
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
		DomainMasters: domainMasters,
		KeyVersion:    opts.KeyVersion,
		Salt:          opts.Salt,
	})

	return ibeSystem, nil
}

// generateMasterKey generates and saves a new master key
func generateMasterKey(ibeSystem *ibe.IBESystem, outputDir string) error {
	masterKeyPath := filepath.Join(outputDir, "master.key")

	if err := ibeSystem.SaveMasterSecretToFile(masterKeyPath); err != nil {
		return fmt.Errorf("failed to save master key: %w", err)
	}

	log.Info().
		Str("master_key_path", masterKeyPath).
		Str("master_key_hash", hex.EncodeToString(ibeSystem.GetMasterSecret())).
		Msg("Master key generated and saved")

	return nil
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

// generateRoleKeys generates role-specific keys with time windows
func generateRoleKeys(ibeSystem *ibe.IBESystem, roles, scopes []string, timeWindows []time.Duration, outputDir string) error {
	roleKeysDir := filepath.Join(outputDir, "roles")
	if err := os.MkdirAll(roleKeysDir, 0755); err != nil {
		return fmt.Errorf("failed to create role keys directory: %w", err)
	}

	for _, role := range roles {
		roleDir := filepath.Join(roleKeysDir, role)
		if err := os.MkdirAll(roleDir, 0755); err != nil {
			return fmt.Errorf("failed to create role directory: %w", err)
		}

		for _, scope := range scopes {
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

	// Generate test role keys
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

func parseDomains(domainsStr string) []string {
	if domainsStr == "" {
		return []string{
			ibe.DOMAIN_USER_PSEUDONYMS,
			ibe.DOMAIN_USER_CORRELATION,
			ibe.DOMAIN_MOD_CORRELATION,
			ibe.DOMAIN_ADMIN_CORRELATION,
			ibe.DOMAIN_LEGAL_CORRELATION,
		}
	}
	return strings.Split(domainsStr, ",")
}

func parseTimeWindows(timeWindowsStr string) []time.Duration {
	if timeWindowsStr == "" {
		return []time.Duration{
			time.Hour,
			24 * time.Hour,
			7 * 24 * time.Hour,
			30 * 24 * time.Hour,
		}
	}

	parts := strings.Split(timeWindowsStr, ",")
	var timeWindows []time.Duration

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if duration, err := parseDuration(part); err == nil {
			timeWindows = append(timeWindows, duration)
		}
	}

	return timeWindows
}

func parseRoles(rolesStr string) []string {
	if rolesStr == "" {
		return []string{"user", "moderator", "subforum_owner", "platform_admin", "trust_safety", "legal_team"}
	}
	return strings.Split(rolesStr, ",")
}

func parseScopes(scopesStr string) []string {
	if scopesStr == "" {
		return []string{"authentication", "self_correlation", "correlation"}
	}
	return strings.Split(scopesStr, ",")
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(s)

	switch {
	case strings.HasSuffix(s, "h"):
		return time.ParseDuration(s)
	case strings.HasSuffix(s, "d"):
		days := strings.TrimSuffix(s, "d")
		return time.ParseDuration(days + "h")
	case strings.HasSuffix(s, "w"):
		weeks := strings.TrimSuffix(s, "w")
		return time.ParseDuration(weeks + "h")
	default:
		return time.ParseDuration(s)
	}
}

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

func loadMasterSecretFromFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
