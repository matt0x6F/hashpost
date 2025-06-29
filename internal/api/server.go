package api

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/routes"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// Server represents the API server
type Server struct {
	API    huma.API
	Mux    *http.ServeMux
	Config huma.Config
}

// NewServer creates a new API server with middleware and routes
func NewServer() *Server {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Create API key DAO for authentication
	apiKeyDAO := dao.NewAPIKeyDAO(db)

	// Create DAOs
	pseudonymDAO := dao.NewPseudonymDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	postDAO := dao.NewPostDAO(db)
	commentDAO := dao.NewCommentDAO(db)
	userDAO := dao.NewUserDAO(db)
	userPreferencesDAO := dao.NewUserPreferencesDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)

	// Create IBE system for correlation operations
	ibeSystem := ibe.NewIBESystem()

	// Create auth middleware with configuration
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret, apiKeyDAO, &cfg.JWT, &cfg.Security)

	// Set the global auth middleware for Huma functions
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	// Create a new HTTP mux
	mux := http.NewServeMux()

	// Create Huma configuration
	config := huma.DefaultConfig("HashPost API", "1.0.0")

	// Create a new Huma API with humago adapter
	api := humago.New(mux, config)

	// Add router-agnostic middleware
	api.UseMiddleware(middleware.LoggingMiddleware)
	api.UseMiddleware(middleware.CORSMiddleware)

	// Add authentication middleware to extract user context
	api.UseMiddleware(middleware.AuthenticateUserHuma)

	// Note: Authentication middleware is applied per-route as needed
	// Public routes (like register, login) don't require authentication
	log.Info().Str("jwt_secret_length", fmt.Sprintf("%d", len(cfg.JWT.Secret))).Msg("JWT configuration loaded")

	// Register routes
	routes.RegisterHealthRoutes(api)
	routes.RegisterHelloRoutes(api)
	routes.RegisterAuthRoutes(api, cfg, db)
	routes.RegisterUserRoutes(api, userDAO, pseudonymDAO, userPreferencesDAO, userBlocksDAO, postDAO, commentDAO)
	routes.RegisterSubforumRoutes(api, db)
	routes.RegisterMessagesRoutes(api)
	routes.RegisterSearchRoutes(api)
	routes.RegisterModerationRoutes(api)
	routes.RegisterContentRoutes(api, db)
	routes.RegisterCorrelationRoutes(api, db, ibeSystem, pseudonymDAO, identityMappingDAO, postDAO, commentDAO)

	return &Server{
		API:    api,
		Mux:    mux,
		Config: config,
	}
}

// GetMux returns the HTTP mux for server setup
func (s *Server) GetMux() *http.ServeMux {
	return s.Mux
}

// GetHandler returns the HTTP handler with router-specific middleware applied
func (s *Server) GetHandler() http.Handler {
	// Apply router-specific middleware to the entire mux
	return middleware.NewRouterMiddleware(s.Mux)
}
