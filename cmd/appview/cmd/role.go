/*
Copyright © 2025 HashPost Team
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// roleCmd represents the role command
var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage user roles",
	Long:  `Manage user roles including assignment and revocation.`,
}

// assignCmd represents the assign command
var assignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a role to a user",
	Long:  `Assign a role to a user. Requires platform admin authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		assignRole()
	},
}

// revokeCmd represents the revoke command
var revokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a role from a user",
	Long:  `Revoke a role from a user. Requires platform admin authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		revokeRole()
	},
}

var (
	// Common flags
	userDID    string
	roleName   string
	subforumID string
	expiresAt  string
	serverURL  string
	authToken  string
)

func init() {
	rootCmd.AddCommand(roleCmd)
	roleCmd.AddCommand(assignCmd)
	roleCmd.AddCommand(revokeCmd)

	// Assign command flags
	assignCmd.Flags().StringVar(&userDID, "user-did", "", "User DID to assign role to (required)")
	assignCmd.Flags().StringVar(&roleName, "role", "", "Role name to assign (required)")
	assignCmd.Flags().StringVar(&subforumID, "subforum-id", "", "Subforum ID (optional, for subforum-specific roles)")
	assignCmd.Flags().StringVar(&expiresAt, "expires-at", "", "Expiration date in RFC3339 format (optional)")
	assignCmd.Flags().StringVar(&serverURL, "server", "http://localhost:8081", "AppView server URL")
	assignCmd.Flags().StringVar(&authToken, "token", "", "Authentication token (required)")
	assignCmd.MarkFlagRequired("user-did")
	assignCmd.MarkFlagRequired("role")
	assignCmd.MarkFlagRequired("token")

	// Revoke command flags
	revokeCmd.Flags().StringVar(&userDID, "user-did", "", "User DID to revoke role from (required)")
	revokeCmd.Flags().StringVar(&roleName, "role", "", "Role name to revoke (required)")
	revokeCmd.Flags().StringVar(&subforumID, "subforum-id", "", "Subforum ID (optional, for subforum-specific roles)")
	revokeCmd.Flags().StringVar(&serverURL, "server", "http://localhost:8081", "AppView server URL")
	revokeCmd.Flags().StringVar(&authToken, "token", "", "Authentication token (required)")
	revokeCmd.MarkFlagRequired("user-did")
	revokeCmd.MarkFlagRequired("role")
	revokeCmd.MarkFlagRequired("token")
}

func assignRole() {
	// Validate expires-at format if provided
	var expiresAtPtr *string
	if expiresAt != "" {
		if _, err := time.Parse(time.RFC3339, expiresAt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid expires-at format. Use RFC3339 format (e.g., 2024-12-31T23:59:59Z)\n")
			os.Exit(1)
		}
		expiresAtPtr = &expiresAt
	}

	// Prepare request body
	reqBody := map[string]interface{}{
		"user_did":  userDID,
		"role_name": roleName,
	}

	if subforumID != "" {
		reqBody["subforum_id"] = subforumID
	}

	if expiresAtPtr != nil {
		reqBody["expires_at"] = *expiresAtPtr
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to marshal request body: %v\n", err)
		os.Exit(1)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/api/v1/admin/assign-role", serverURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to make request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: Server returned status %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Role assigned successfully!\n")
	fmt.Printf("User DID: %s\n", userDID)
	fmt.Printf("Role: %s\n", roleName)
	if subforumID != "" {
		fmt.Printf("Subforum ID: %s\n", subforumID)
	}
	if expiresAtPtr != nil {
		fmt.Printf("Expires at: %s\n", *expiresAtPtr)
	}
}

func revokeRole() {
	// Prepare request body
	reqBody := map[string]interface{}{
		"user_did":  userDID,
		"role_name": roleName,
	}

	if subforumID != "" {
		reqBody["subforum_id"] = subforumID
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to marshal request body: %v\n", err)
		os.Exit(1)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/api/v1/admin/revoke-role", serverURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to make request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: Server returned status %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Role revoked successfully!\n")
	fmt.Printf("User DID: %s\n", userDID)
	fmt.Printf("Role: %s\n", roleName)
	if subforumID != "" {
		fmt.Printf("Subforum ID: %s\n", subforumID)
	}
}
