package appview

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/matt0x6f/hashpost/internal/config"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
	"github.com/matt0x6f/hashpost/internal/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Server represents the HashPost AppView server
type Server struct {
	config               *config.Config
	httpServer           *http.Server
	database             *pgxpool.Pool
	handlers             *Handlers
	eventConsumer        *EventConsumer
	rbacService          *RBACService
	authMiddleware       *AuthMiddleware
	rbacHandlers         *RBACHandlers
	voteHandlers         *VoteHandlers
	subscriptionHandlers *SubscriptionHandlers
	pdsHandlers          *PDSHandlers
	logger               *slog.Logger
}

// NewServer creates a new AppView server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Set up logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Connect to database using pgxpool
	dbURL := cfg.GetAppViewDatabaseURL()
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create RBAC service
	rbacService := NewRBACService(dbPool, logger)

	// Create authentication middleware
	authMiddleware := NewAuthMiddleware(rbacService, logger)

	// Create database queries
	queries := generated.New(dbPool)

	// Create handlers
	handlers := NewHandlers(queries, logger, rbacService)

	// Create RBAC handlers
	rbacHandlers := NewRBACHandlers(rbacService, logger)

	// Create vote handlers
	voteHandlers := NewVoteHandlers(queries, logger)

	// Create subscription handlers
	subscriptionHandlers := NewSubscriptionHandlers(queries, logger)

	// Create PDS handlers
	pdsHandlers := NewPDSHandlers(queries, logger)

	// Get NATS URL from environment
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	// Create AppView database wrapper
	appViewDB := NewDatabase(dbPool, logger)

	// Create event consumer
	eventConsumer, err := NewEventConsumer(natsURL, appViewDB, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create event consumer: %w", err)
	}

	server := &Server{
		config:         cfg,
		database:       dbPool,
		handlers:       handlers,
		eventConsumer:  eventConsumer,
		rbacService:    rbacService,
		authMiddleware: authMiddleware,
		logger:         logger,
	}

	// Store handlers for use in route registration
	server.rbacHandlers = rbacHandlers
	server.voteHandlers = voteHandlers
	server.subscriptionHandlers = subscriptionHandlers
	server.pdsHandlers = pdsHandlers

	return server, nil
}

// Start starts the AppView server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register API endpoints
	s.registerAPIEndpoints(mux)

	// Add CORS middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Add logging middleware
	handler := middleware.LoggingMiddleware(s.logger)(corsHandler(mux))

	s.httpServer = &http.Server{
		Addr:    s.config.GetAppViewServerAddress(),
		Handler: handler,
	}

	// Start event consumer in background
	go func() {
		ctx := context.Background()
		if err := s.eventConsumer.StartConsuming(ctx); err != nil {
			s.logger.Error("Event consumer stopped", "error", err)
		}
	}()

	s.logger.Info("Starting HashPost AppView server", "address", s.config.GetAppViewServerAddress())
	return s.httpServer.ListenAndServe()
}

// Stop stops the AppView server
func (s *Server) Stop(ctx context.Context) error {
	// Close event consumer
	if s.eventConsumer != nil {
		if err := s.eventConsumer.Close(); err != nil {
			s.logger.Error("Failed to close event consumer", "error", err)
		}
	}

	// Shutdown HTTP server
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// registerAPIEndpoints registers all API endpoints using the generated server wrapper
func (s *Server) registerAPIEndpoints(mux *http.ServeMux) {
	// Swagger UI and OpenAPI spec endpoints
	mux.HandleFunc("/docs", SwaggerUIHandler)
	mux.HandleFunc("/docs/", SwaggerUIHandler)
	mux.HandleFunc("/openapi.yaml", OpenAPISpecHandler)
	mux.HandleFunc("/api-docs", OpenAPISpecHandler)

	// Health endpoint (no authentication required)
	mux.HandleFunc("/health", s.handlers.Health)

	// Password update endpoint (requires authentication)
	mux.HandleFunc("/api/v1/auth/updatePassword", s.handlers.UpdatePassword)

	// Use the generated server wrapper for all OpenAPI endpoints
	// This ensures spec-handler parity
	handler := HandlerWithOptions(s, StdHTTPServerOptions{
		Middlewares: []MiddlewareFunc{
			// Apply authentication middleware to all routes
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Skip authentication for certain public endpoints
					// For /api/v1/subforums, only GET requests are public (listing), POST requires auth (creating)
					if s.isPublicEndpoint(r.URL.Path, r.Method) {
						next.ServeHTTP(w, r)
						return
					}

					// Apply authentication middleware
					s.authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
						next.ServeHTTP(w, r)
					})(w, r)
				})
			},
		},
	})
	mux.Handle("/api/v1/", handler)
}

// isPublicEndpoint checks if an endpoint should skip authentication
func (s *Server) isPublicEndpoint(path string, method string) bool {
	// Always public endpoints (any method)
	alwaysPublicEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
	}

	for _, endpoint := range alwaysPublicEndpoints {
		if path == endpoint {
			return true
		}
	}

	// Method-specific public endpoints
	if path == "/api/v1/subforums" && method == "GET" {
		return true // GET /api/v1/subforums is public (list subforums)
	}

	// Check for subforum detail endpoint (GET /api/v1/subforums/{slug})
	if method == "GET" && len(path) > 18 && path[:18] == "/api/v1/subforums/" {
		return true // GET /api/v1/subforums/{slug} is public (view subforum)
	}

	// Check for posts endpoint (GET /api/v1/posts)
	if path == "/api/v1/posts" && method == "GET" {
		return true // GET /api/v1/posts is public (list posts)
	}

	// Check for individual post endpoints (GET /api/v1/posts/{id})
	// But exclude user-vote endpoints which require authentication
	if method == "GET" && len(path) > 15 && strings.HasPrefix(path, "/api/v1/posts/") && !strings.Contains(path, "/user-vote") {
		return true // GET /api/v1/posts/{id} is public (view post)
	}

	// Check for comments endpoint (GET /api/v1/comments)
	if path == "/api/v1/comments" && method == "GET" {
		return true // GET /api/v1/comments is public (list comments)
	}

	// Check for individual comment endpoints (GET /api/v1/comments/{id})
	if method == "GET" && len(path) > 18 && strings.HasPrefix(path, "/api/v1/comments/") {
		return true // GET /api/v1/comments/{id} is public (view comment)
	}

	return false
}

// Implement the generated ServerInterface

// Authentication Endpoints
func (s *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
	s.handlers.AuthHandler(w, r)
}

func (s *Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	s.handlers.AuthHandler(w, r)
}

func (s *Server) GetCurrentUserSession(w http.ResponseWriter, r *http.Request) {
	s.handlers.AuthHandler(w, r)
}

func (s *Server) LogoutUser(w http.ResponseWriter, r *http.Request) {
	s.handlers.AuthHandler(w, r)
}

// RBAC Management Endpoints
func (s *Server) AssignRole(w http.ResponseWriter, r *http.Request) {
	s.rbacHandlers.AssignRole(w, r)
}

func (s *Server) ListPermissions(w http.ResponseWriter, r *http.Request) {
	s.rbacHandlers.ListPermissions(w, r)
}

func (s *Server) RevokeRole(w http.ResponseWriter, r *http.Request) {
	s.rbacHandlers.RevokeRole(w, r)
}

func (s *Server) ListRoles(w http.ResponseWriter, r *http.Request) {
	s.rbacHandlers.ListRoles(w, r)
}

func (s *Server) GetUserRoles(w http.ResponseWriter, r *http.Request, params GetUserRolesParams) {
	// Convert params to query parameters for existing handler
	r.URL.RawQuery = fmt.Sprintf("user_did=%s", params.UserDid)
	if params.SubforumId != nil {
		r.URL.RawQuery += fmt.Sprintf("&subforum_id=%s", *params.SubforumId)
	}
	s.rbacHandlers.GetUserRoles(w, r)
}

func (s *Server) ListAllUsers(w http.ResponseWriter, r *http.Request, params ListAllUsersParams) {
	// Convert params to query parameters for existing handler
	limit := 20
	offset := 0

	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	query := fmt.Sprintf("limit=%d&offset=%d", limit, offset)
	r.URL.RawQuery = query
	s.rbacHandlers.ListUsers(w, r)
}

// CRUD Endpoints
func (s *Server) ListComments(w http.ResponseWriter, r *http.Request, params ListCommentsParams) {
	// Convert params to query parameters for existing handler
	r.URL.RawQuery = fmt.Sprintf("post_id=%s&limit=%d&offset=%d", params.PostId, params.Limit, params.Offset)
	s.handlers.CommentsHandler(w, r)
}

func (s *Server) ListPosts(w http.ResponseWriter, r *http.Request, params ListPostsParams) {
	// Use the generated parameters directly instead of re-parsing URL
	limit := 20
	offset := 0
	subforum := ""

	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}
	if params.Subforum != nil {
		subforum = *params.Subforum
	}

	// Call the handler with the parsed parameters
	s.handlers.ListPostsWithParams(w, r, limit, offset, subforum)
}

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	s.handlers.PostsHandler(w, r)
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the existing handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s", id.String())
	s.handlers.PostByIDHandler(w, r)
}

func (s *Server) GetPostByID(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the existing handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s", id.String())
	s.handlers.PostByIDHandler(w, r)
}

func (s *Server) UpdatePost(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the existing handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s", id.String())
	s.handlers.PostByIDHandler(w, r)
}

// Comment endpoints
func (s *Server) CreateComment(w http.ResponseWriter, r *http.Request) {
	s.handlers.CreateComment(w, r)
}

func (s *Server) GetCommentByID(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the existing handler
	r.URL.Path = fmt.Sprintf("/api/v1/comments/%s", id.String())
	s.handlers.CommentByIDHandler(w, r)
}

func (s *Server) UpdateComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the existing handler
	r.URL.Path = fmt.Sprintf("/api/v1/comments/%s", id.String())
	s.handlers.CommentByIDHandler(w, r)
}

func (s *Server) DeleteComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the existing handler
	r.URL.Path = fmt.Sprintf("/api/v1/comments/%s", id.String())
	s.handlers.CommentByIDHandler(w, r)
}

func (s *Server) ListSubforums(w http.ResponseWriter, r *http.Request, params ListSubforumsParams) {
	// Use the generated parameters directly instead of re-parsing URL
	limit := 20
	offset := 0

	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	// Call the handler with the parsed parameters
	s.handlers.ListSubforumsWithParams(w, r, limit, offset)
}

func (s *Server) GetSubforumBySlug(w http.ResponseWriter, r *http.Request, slug string) {
	// Set the slug in the URL path for the existing handler
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s", slug)
	s.handlers.SubforumBySlugHandler(w, r)
}

func (s *Server) CreateSubforum(w http.ResponseWriter, r *http.Request) {
	s.handlers.CreateSubforum(w, r)
}

// Vote Endpoints
func (s *Server) VoteOnPost(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the vote handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s/vote", id.String())
	s.voteHandlers.VoteOnPost(w, r)
}

func (s *Server) RemoveVoteFromPost(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the vote handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s/vote", id.String())
	s.voteHandlers.VoteOnPost(w, r)
}

func (s *Server) GetUserVoteOnPost(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the vote handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s/vote", id.String())
	s.voteHandlers.GetUserVoteOnPost(w, r)
}

func (s *Server) VoteOnComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the vote handler
	r.URL.Path = fmt.Sprintf("/api/v1/comments/%s/vote", id.String())
	s.voteHandlers.VoteOnComment(w, r)
}

func (s *Server) RemoveVoteFromComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the vote handler
	r.URL.Path = fmt.Sprintf("/api/v1/comments/%s/vote", id.String())
	s.voteHandlers.VoteOnComment(w, r)
}

func (s *Server) GetUserVoteOnComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the vote handler
	r.URL.Path = fmt.Sprintf("/api/v1/comments/%s/vote", id.String())
	s.voteHandlers.GetUserVoteOnComment(w, r)
}

// Subscription Endpoints
func (s *Server) SubscribeToSubforum(w http.ResponseWriter, r *http.Request, slug string) {
	// Set the slug in the URL path for the subscription handler
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s/subscribe", slug)
	s.subscriptionHandlers.SubscribeToSubforum(w, r)
}

func (s *Server) UnsubscribeFromSubforum(w http.ResponseWriter, r *http.Request, slug string) {
	// Set the slug in the URL path for the subscription handler
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s/subscribe", slug)
	s.subscriptionHandlers.UnsubscribeFromSubforum(w, r)
}

func (s *Server) GetUserSubscription(w http.ResponseWriter, r *http.Request, slug string) {
	// Set the slug in the URL path for the subscription handler
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s/subscribe", slug)
	s.subscriptionHandlers.GetUserSubscription(w, r)
}

func (s *Server) ListMySubscriptions(w http.ResponseWriter, r *http.Request, params ListMySubscriptionsParams) {
	// Use the generated parameters directly instead of re-parsing URL
	limit := 20
	offset := 0

	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	// Call the handler with the parsed parameters
	s.subscriptionHandlers.ListMySubscriptionsWithParams(w, r, limit, offset)
}

func (s *Server) ListSubforumSubscribers(w http.ResponseWriter, r *http.Request, slug string, params ListSubforumSubscribersParams) {
	// Convert params to query parameters for the subscription handler
	query := fmt.Sprintf("limit=%d&offset=%d", params.Limit, params.Offset)
	r.URL.RawQuery = query
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s/subscribers", slug)
	s.subscriptionHandlers.ListSubforumSubscribers(w, r)
}

func (s *Server) ListMyModeratedSubforums(w http.ResponseWriter, r *http.Request) {
	s.subscriptionHandlers.ListMyModeratedSubforums(w, r)
}

// New hierarchical RBAC methods

// ListSubforumMembers lists members of a specific subforum
func (s *Server) ListSubforumMembers(w http.ResponseWriter, r *http.Request, slug string, params ListSubforumMembersParams) {
	// Convert params to query parameters for existing handler
	query := fmt.Sprintf("limit=%d&offset=%d", params.Limit, params.Offset)
	r.URL.RawQuery = query
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s/members", slug)

	// Create a new handler for subforum members
	s.handleSubforumMembers(w, r)
}

// AssignSubforumRole assigns a role to a user in a subforum
func (s *Server) AssignSubforumRole(w http.ResponseWriter, r *http.Request, slug string, userDid string) {
	// Set the path and user DID for the handler
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s/members/%s/roles", slug, userDid)

	// Create a new handler for subforum role assignment
	s.handleSubforumRoleAssignment(w, r)
}

// RevokeSubforumRole revokes a role from a user in a subforum
func (s *Server) RevokeSubforumRole(w http.ResponseWriter, r *http.Request, slug string, userDid string, params RevokeSubforumRoleParams) {
	// Set the path and user DID for the handler
	r.URL.Path = fmt.Sprintf("/api/v1/subforums/%s/members/%s/roles", slug, userDid)
	r.URL.RawQuery = fmt.Sprintf("role_name=%s", params.RoleName)

	// Create a new handler for subforum role revocation
	s.handleSubforumRoleRevocation(w, r)
}

// GetMyRoles returns the authenticated user's roles
func (s *Server) GetMyRoles(w http.ResponseWriter, r *http.Request, params GetMyRolesParams) {
	// Convert params to query parameters for existing handler
	if params.SubforumId != nil {
		r.URL.RawQuery = fmt.Sprintf("subforum_id=%s", params.SubforumId.String())
	}

	// Use existing GetUserRoles handler but with authenticated user's DID
	s.handleMyRoles(w, r)
}

// GetMyPermissions returns the authenticated user's permissions
func (s *Server) GetMyPermissions(w http.ResponseWriter, r *http.Request, params GetMyPermissionsParams) {
	// Convert params to query parameters for existing handler
	if params.SubforumId != nil {
		r.URL.RawQuery = fmt.Sprintf("subforum_id=%s", params.SubforumId.String())
	}

	// Use existing CheckPermission handler but for authenticated user
	s.handleMyPermissions(w, r)
}

// Handler implementations for new RBAC endpoints

// handleSubforumMembers handles listing subforum members
func (s *Server) handleSubforumMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	slug := parts[3]

	// Get limit and offset from query
	limit := 20
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get subforum members with roles
	members, err := s.rbacService.GetSubforumMembers(r.Context(), slug, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get subforum members", "error", err, "slug", slug)
		http.Error(w, "Failed to get subforum members", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"members": members,
		"limit":   limit,
		"offset":  offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSubforumRoleAssignment handles assigning roles in subforums
func (s *Server) handleSubforumRoleAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		RoleName  string  `json:"role_name"`
		ExpiresAt *string `json:"expires_at,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract subforum slug and user DID from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	slug := parts[3]
	userDid := parts[5]

	// Get authenticated user
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Assign the role
	err := s.rbacService.AssignSubforumRole(r.Context(), slug, userDid, req.RoleName, userCtx.Did, req.ExpiresAt)
	if err != nil {
		s.logger.Error("Failed to assign subforum role", "error", err, "slug", slug, "user_did", userDid)
		http.Error(w, "Failed to assign role", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Role assigned successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSubforumRoleRevocation handles revoking roles in subforums
func (s *Server) handleSubforumRoleRevocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug and user DID from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	slug := parts[3]
	userDid := parts[5]

	// Get role name from query
	roleName := r.URL.Query().Get("role_name")
	if roleName == "" {
		http.Error(w, "role_name parameter required", http.StatusBadRequest)
		return
	}

	// Get authenticated user
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Revoke the role
	err := s.rbacService.RevokeSubforumRole(r.Context(), slug, userDid, roleName, userCtx.Did)
	if err != nil {
		s.logger.Error("Failed to revoke subforum role", "error", err, "slug", slug, "user_did", userDid)
		http.Error(w, "Failed to revoke role", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Role revoked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMyRoles handles getting the authenticated user's roles
func (s *Server) handleMyRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get subforum ID from query if provided
	var subforumID *string
	if subforumIDStr := r.URL.Query().Get("subforum_id"); subforumIDStr != "" {
		subforumID = &subforumIDStr
	}

	// Get user roles
	roles, err := s.rbacService.GetUserRoles(r.Context(), userCtx.Did, subforumID)
	if err != nil {
		s.logger.Error("Failed to get user roles", "error", err, "user_did", userCtx.Did)
		http.Error(w, "Failed to get roles", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"roles": roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMyPermissions handles getting the authenticated user's permissions
func (s *Server) handleMyPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get subforum ID from query if provided
	var subforumID *string
	if subforumIDStr := r.URL.Query().Get("subforum_id"); subforumIDStr != "" {
		subforumID = &subforumIDStr
	}

	// Get user permissions
	permissions, err := s.rbacService.GetUserPermissions(r.Context(), userCtx.Did, subforumID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", "error", err, "user_did", userCtx.Did)
		http.Error(w, "Failed to get permissions", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"permissions": permissions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPostMetrics handles GET /api/v1/posts/{id}/metrics
func (s *Server) GetPostMetrics(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the post handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s/metrics", id.String())
	s.handlers.GetPostMetrics(w, r)
}

// GetPostModerationState handles GET /api/v1/posts/{id}/moderation
func (s *Server) GetPostModerationState(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the post handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s/moderation", id.String())
	s.handlers.GetPostModerationState(w, r)
}

// GetPostUserVote handles GET /api/v1/posts/{id}/user-vote
func (s *Server) GetPostUserVote(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Set the ID in the URL path for the vote handler
	r.URL.Path = fmt.Sprintf("/api/v1/posts/%s/user-vote", id.String())
	s.voteHandlers.GetUserVoteOnly(w, r)
}

// PDS Management Endpoints

// ListPDSServers handles GET /api/v1/admin/pds/servers
func (s *Server) ListPDSServers(w http.ResponseWriter, r *http.Request) {
	s.pdsHandlers.ListPDSServers(w, r)
}

// GetPDSServerDetails handles GET /api/v1/admin/pds/{endpoint}
func (s *Server) GetPDSServerDetails(w http.ResponseWriter, r *http.Request, endpoint string) {
	s.pdsHandlers.GetPDSServerDetails(w, r, endpoint)
}

// ListPDSServerUsers handles GET /api/v1/admin/pds/{endpoint}/users
func (s *Server) ListPDSServerUsers(w http.ResponseWriter, r *http.Request, endpoint string, params ListPDSServerUsersParams) {
	s.pdsHandlers.ListPDSServerUsers(w, r, endpoint, params)
}
