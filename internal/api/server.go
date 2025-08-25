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
	API       huma.API
	Mux       *http.ServeMux
	Config    huma.Config
	AppConfig *config.Config
}

// NewServer creates a new API server with middleware and routes
func NewServer(cfg *config.Config) *Server {

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Get the raw *sql.DB from bob.DB
	rawDB := db.DB

	// Create IBE system from configuration
	ibeSystem, err := ibe.NewIBESystemFromConfig(cfg.IBE.DomainKeysDir, cfg.IBE.KeyVersion, cfg.IBE.Salt)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create IBE system from configuration")
	}

	// Create DAOs
	userDAO := dao.NewUserDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	pseudonymDAO := dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)
	postDAO := dao.NewPostDAO(db)
	commentDAO := dao.NewCommentDAO(db)
	voteDAO := dao.NewVoteDAO(db)
	userPreferencesDAO := dao.NewUserPreferencesDAO(db)
	apiKeyDAO := dao.NewAPIKeyDAO(db)
	subforumDAO := dao.NewSubforumDAO(db)

	// Create moderation DAOs
	reportDAO := dao.NewReportDAO(db)
	moderationActionDAO := dao.NewModerationActionDAO(db)
	userBanDAO := dao.NewUserBanDAO(db)
	permissionDAO := dao.NewPermissionDAO(db)

	// Create auth middleware with configuration
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret, apiKeyDAO, &cfg.JWT, &cfg.Security)

	// Set the global auth middleware for Huma functions
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	// Create a new HTTP mux
	mux := http.NewServeMux()

	// Create Huma configuration
	config := huma.DefaultConfig("HashPost API", "1.0.0")

	contact := &huma.Contact{
		Name:  "Matt Ouille",
		URL:   "https://github.com/matt0x6f",
		Email: "matt@hashpost.dev",
	}

	config.Info.Contact = contact

	config.Info.Description = "HashPost is a modern forum platform with enhanced security and privacy features."

	// Create a new Huma API with humago adapter
	api := humago.New(mux, config)

	// Add router-agnostic middleware
	api.UseMiddleware(middleware.LoggingMiddleware)
	api.UseMiddleware(middleware.CORSMiddleware(&cfg.CORS))

	// Note: Authentication middleware is applied per-route as needed
	// Public routes (like register, login) don't require authentication
	// Protected routes use AuthInput structs to handle authentication
	log.Info().Str("jwt_secret_length", fmt.Sprintf("%d", len(cfg.JWT.Secret))).Msg("JWT configuration loaded")

	// Register routes
	routes.RegisterHealthRoutes(api)
	routes.RegisterAuthRoutes(api, cfg, db, rawDB, ibeSystem)
	routes.RegisterUserRoutes(api, userDAO, pseudonymDAO, userPreferencesDAO, userBlocksDAO, postDAO, commentDAO, ibeSystem)
	routes.RegisterSubforumRoutes(api, db, pseudonymDAO)
	routes.RegisterMessagesRoutes(api, db, pseudonymDAO)
	routes.RegisterSearchRoutes(api, db, ibeSystem)
	routes.RegisterModerationRoutes(api, reportDAO, moderationActionDAO, userBanDAO, pseudonymDAO, subforumDAO, postDAO, commentDAO, voteDAO, permissionDAO)
	routes.RegisterRulesRoutes(api, db, pseudonymDAO)
	routes.RegisterSystemSettingsRoutes(api, db, pseudonymDAO)
	routes.RegisterContentRoutes(api, db, rawDB, ibeSystem, identityMappingDAO, userDAO)
	routes.RegisterCorrelationRoutes(api, db, ibeSystem, pseudonymDAO, identityMappingDAO, postDAO, commentDAO, subforumDAO)

	return &Server{
		API:       api,
		Mux:       mux,
		Config:    config,
		AppConfig: cfg,
	}
}

// NewServerForOpenAPI creates a minimal server just for generating OpenAPI specs
// This function doesn't require IBE or database connections
func NewServerForOpenAPI() *Server {
	// Create a new HTTP mux
	mux := http.NewServeMux()

	// Create Huma configuration
	config := huma.DefaultConfig("HashPost API", "1.0.0")

	// Create a new Huma API with humago adapter
	api := humago.New(mux, config)

	// Register only the routes that don't require IBE or database
	// These are typically just the basic structure routes
	routes.RegisterHealthRoutes(api)

	return &Server{
		API:       api,
		Mux:       mux,
		Config:    config,
		AppConfig: nil, // No config needed for OpenAPI generation
	}
}

// GetMux returns the HTTP mux for server setup
func (s *Server) GetMux() *http.ServeMux {
	return s.Mux
}

// GetHandler returns the HTTP handler with router-specific middleware applied
func (s *Server) GetHandler() http.Handler {
	// Apply CORS middleware first, then router middleware
	return middleware.CORSMiddlewareHTTP(&s.AppConfig.CORS)(middleware.NewRouterMiddleware(s.Mux))
}
