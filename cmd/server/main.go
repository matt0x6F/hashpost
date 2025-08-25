package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/matt0x6f/hashpost/cmd/server/commands"
	"github.com/matt0x6f/hashpost/internal/api"
	"github.com/matt0x6f/hashpost/internal/api/logger"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ServerOptions defines the CLI options for the server command
type ServerOptions struct {
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

	// Create server command that uses the CLI
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the HashPost server",
		Long:  "Start the HashPost server with the specified configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Create the CLI for the server - IBE system will be initialized only when server starts
			cli := humacli.New(func(hooks humacli.Hooks, opts *ServerOptions) {
				var httpServer *http.Server

				hooks.OnStart(func() {
					// Create the API server with all middleware and routes ONLY when starting
					server := api.NewServer(cfg)

					// Create the HTTP server with graceful shutdown
					httpServer = &http.Server{
						Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
						Handler: server.GetHandler(),
					}

					log.Info().Str("addr", httpServer.Addr).Msg("Server listening")
					if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						log.Fatal().Err(err).Msg("Error starting server")
					}
				})

				hooks.OnStop(func() {
					if httpServer == nil {
						return
					}
					log.Info().Msg("Start shutdown")
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()

					if err := httpServer.Shutdown(ctx); err != nil {
						log.Error().Err(err).Msg("Could not stop server gracefully")
						if err := httpServer.Close(); err != nil {
							log.Fatal().Err(err).Msg("Could not force close server")
						}
					}
					log.Info().Msg("Server stopped")
				})
			})

			// Run the CLI
			cli.Run()
		},
	}

	// Add server command
	rootCmd.AddCommand(serverCmd)

	// Add roles commands
	rootCmd.AddCommand(commands.NewRolesCommands(cfg)...)

	// Add admin commands
	rootCmd.AddCommand(commands.NewAdminCommands(cfg)...)

	// Add IBE keys command
	rootCmd.AddCommand(commands.NewGenerateIBEKeysCommand(cfg))

	// Add key-rotation commands
	rootCmd.AddCommand(commands.NewKeyRotationCommands(cfg)...)

	// Run the CLI
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}
