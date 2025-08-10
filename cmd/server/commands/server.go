package commands

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/matt0x6f/hashpost/internal/api"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// ServerOptions defines the CLI options for the server command
type ServerOptions struct {
	Debug bool   `doc:"Enable debug logging"`
	Host  string `doc:"Hostname to listen on."`
	Port  int    `doc:"Port to listen on." short:"p" default:"8888"`
}

// NewServerCommand creates and returns the server command
func NewServerCommand() humacli.CLI {
	return humacli.New(func(hooks humacli.Hooks, opts *ServerOptions) {
		// Create the API server with all middleware and routes
		server := api.NewServer()

		// Create the HTTP server with graceful shutdown
		httpServer := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler: server.GetHandler(),
		}

		hooks.OnStart(func() {
			// Initialize IBE system only when actually starting the server
			ibeSystem := ibe.NewIBESystemFromEnv()
			log.Info().
				Str("ibe_master_key", hex.EncodeToString(ibeSystem.GetMasterSecret())).
				Str("ibe_salt", ibeSystem.GetSalt()).
				Int("ibe_key_version", ibeSystem.GetKeyVersion()).
				Msg("IBE system configuration (server startup)")

			log.Info().Str("addr", httpServer.Addr).Msg("Server listening")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Error starting server")
			}
		})

		hooks.OnStop(func() {
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
}
