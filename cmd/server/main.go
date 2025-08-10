package main

import (
	"github.com/matt0x6f/hashpost/cmd/server/commands"
	"github.com/matt0x6f/hashpost/internal/api/logger"
	"github.com/matt0x6f/hashpost/internal/config"
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

	// Create the root command
	rootCmd := &cobra.Command{
		Use:     "hashpost",
		Version: "1.0.0",
		Short:   "HashPost - A modern forum platform",
		Long:    "HashPost is a modern forum platform with enhanced security and privacy features.",
	}

	// Add global flags
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().String("host", "localhost", "Hostname to listen on")
	rootCmd.PersistentFlags().IntP("port", "p", 8888, "Port to listen on")

	// Create server command
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the HashPost server",
		Long:  "Start the HashPost server with the specified configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Create the CLI for server mode using the commands package
			cli := commands.NewServerCommand()

			// Run the server
			cli.Run()
		},
	}

	// Set the default command to start the server
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		serverCmd.Run(cmd, args)
		return nil
	}

	// Add server command
	rootCmd.AddCommand(serverCmd)

	// Add roles commands
	rootCmd.AddCommand(commands.NewRolesCommands()...)

	// Add admin commands
	rootCmd.AddCommand(commands.NewAdminCommands()...)

	// Add IBE keys command
	rootCmd.AddCommand(commands.NewGenerateIBEKeysCommand())

	// Add key-rotation commands
	rootCmd.AddCommand(commands.NewKeyRotationCommands()...)

	// Add openapi command
	rootCmd.AddCommand(commands.NewOpenAPICommand())

	// Run the CLI
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}
