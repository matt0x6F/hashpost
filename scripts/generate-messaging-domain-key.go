package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Define the domain name
	domain := "user_messaging_v1"

	// Define the keys directory
	keysDir := "./keys/domains"

	// Create the keys directory if it doesn't exist
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		fmt.Printf("Failed to create keys directory: %v\n", err)
		os.Exit(1)
	}

	// Generate a new 32-byte (256-bit) key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		fmt.Printf("Failed to generate random key: %v\n", err)
		os.Exit(1)
	}

	// Create the key file path
	keyPath := filepath.Join(keysDir, domain+".key")

	// Check if the key file already exists
	if _, err := os.Stat(keyPath); err == nil {
		fmt.Printf("Domain key file %s already exists, skipping generation\n", keyPath)
		os.Exit(0)
	}

	// Write the key to the file
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		fmt.Printf("Failed to write key file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated domain key for %s\n", domain)
	fmt.Printf("Key file: %s\n", keyPath)
	fmt.Printf("Key fingerprint: %s\n", hex.EncodeToString(key))
}
