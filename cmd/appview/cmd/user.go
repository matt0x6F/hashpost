/*
Copyright © 2025 HashPost Team
*/
package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matt0x6f/hashpost/internal/config"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Manage users including password reset and account operations.`,
}

// resetPasswordCmd represents the reset-password command
var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset a user's password",
	Long:  `Reset a user's password by email or handle. This will generate a new password and update the user's account.`,
	Run: func(cmd *cobra.Command, args []string) {
		resetUserPassword()
	},
}

var (
	// User management flags
	userEmail      string
	userHandle     string
	newPassword    string
	userServerURL  string
	userConfigFile string
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(resetPasswordCmd)

	// Reset password command flags
	resetPasswordCmd.Flags().StringVar(&userEmail, "email", "", "User's email address")
	resetPasswordCmd.Flags().StringVar(&userHandle, "handle", "", "User's handle")
	resetPasswordCmd.Flags().StringVar(&newPassword, "password", "", "New password (if not provided, a random password will be generated)")
	resetPasswordCmd.Flags().StringVar(&userServerURL, "server", "http://localhost:8080", "PDS server URL")
	resetPasswordCmd.Flags().StringVar(&userConfigFile, "config", "config/dev.yaml", "Config file path")

	// Require either email or handle
	resetPasswordCmd.MarkFlagsOneRequired("email", "handle")
}

func resetUserPassword() {
	// Load configuration
	cfg, err := config.LoadConfig(userConfigFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to PDS database
	dbURL := cfg.GetDatabaseURL()
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Create database queries instance
	queries := generated.New(dbPool)

	// Find user by email or handle
	var foundUser *generated.User
	if userEmail != "" {
		// Find user by email
		user, err := queries.GetUserByEmail(context.Background(), &userEmail)
		if err != nil {
			log.Fatalf("Failed to find user by email %s: %v", userEmail, err)
		}
		foundUser = user
	} else if userHandle != "" {
		// Find user by handle
		user, err := queries.GetUserByHandle(context.Background(), userHandle)
		if err != nil {
			log.Fatalf("Failed to find user by handle %s: %v", userHandle, err)
		}
		foundUser = user
	}

	// Generate new password if not provided
	if newPassword == "" {
		newPassword = generateRandomPassword()
		fmt.Printf("Generated new password: %s\n", newPassword)
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Update user's password in database
	hashedPasswordStr := string(hashedPassword)
	err = queries.UpdateUserPasswordHash(context.Background(), &generated.UpdateUserPasswordHashParams{
		PasswordHash: &hashedPasswordStr,
		ID:           foundUser.ID,
	})
	if err != nil {
		log.Fatalf("Failed to update user password: %v", err)
	}

	fmt.Printf("Password successfully reset for user %s (%s)\n", *foundUser.Email, foundUser.Did)
	fmt.Printf("New password: %s\n", newPassword)
	fmt.Println("Please share this password securely with the user.")
}

// generateRandomPassword generates a secure random password
func generateRandomPassword() string {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatalf("Failed to generate random password: %v", err)
	}

	// Convert to hex string and take first 12 characters
	password := hex.EncodeToString(bytes)[:12]

	// Add some complexity
	return password + "!@#"
}
