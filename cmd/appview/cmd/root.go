/*
Copyright © 2025 HashPost Team
*/
package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/matt0x6f/hashpost/internal/appview"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "appview",
	Short: "HashPost AppView (Stateless Aggregator)",
	Long: `HashPost AppView is a stateless aggregator that provides a unified API
for accessing HashPost data aggregated from the PDS.

The AppView provides:
- Unified API for HashPost features
- Data aggregation from PDS
- OpenAPI specification
- Type-safe client generation`,
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
	Short: "Start the AppView server",
	Long:  `Start the HashPost AppView server with the specified configuration.`,
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

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create AppView server
	server, err := appview.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create AppView server: %v", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down AppView server...")
		os.Exit(0)
	}()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start AppView server: %v", err)
	}
}
