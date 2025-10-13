/*
Copyright © 2025 HashPost Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matt0x6f/hashpost/internal/config"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	"github.com/matt0x6f/hashpost/internal/pds"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pds",
	Short: "HashPost PDS (Personal Data Server)",
	Long: `HashPost PDS is the Personal Data Server component that handles
atproto protocol compliance, identity management, and data storage.

The PDS provides:
- Atproto protocol endpoints (/xrpc/com.atproto.*)
- DID and handle resolution
- Session management
- Repository operations
- Custom HashPost features`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is config/dev.yaml)")
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the PDS server",
	Long:  `Start the HashPost PDS server with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runServer() {
	// Load configuration
	if configFile == "" {
		configFile = "config/dev.yaml"
	}

	// Debug: Check environment variables
	dbURL := os.Getenv("DATABASE_URL")
	log.Printf("DEBUG: DATABASE_URL environment variable: %s", dbURL)
	if dbURL == "" {
		log.Printf("DEBUG: DATABASE_URL is empty!")
	} else {
		log.Printf("DEBUG: DATABASE_URL found: %s", dbURL)
	}
	log.Printf("DEBUG: All environment variables:")
	for _, env := range os.Environ() {
		if strings.Contains(env, "DATABASE") {
			log.Printf("DEBUG: %s", env)
		}
	}
	
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dbURL = cfg.GetDatabaseURL()
	log.Printf("Connecting to database with URL: %s", dbURL)
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Create database queries instance
	queries := generated.New(dbPool)

	// Create PDS server
	server, err := pds.NewServer(cfg, queries)
	if err != nil {
		log.Fatalf("Failed to create PDS server: %v", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down PDS server...")
		os.Exit(0)
	}()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start PDS server: %v", err)
	}
}
