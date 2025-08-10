package commands

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// InitializeIBEForCommand initializes the IBE system for commands that need it
// This function is called before running commands that require IBE functionality
func InitializeIBEForCommand() error {
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

	log.Info().
		Str("ibe_master_key", hex.EncodeToString(ibeSystem.GetMasterSecret())).
		Str("ibe_salt", ibeSystem.GetSalt()).
		Int("ibe_key_version", ibeSystem.GetKeyVersion()).
		Msg("IBE system initialized for command")
	return nil
}
