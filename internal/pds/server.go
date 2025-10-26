package pds

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/matt0x6f/hashpost/internal/config"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	jwtservice "github.com/matt0x6f/hashpost/internal/jwt"
	"github.com/matt0x6f/hashpost/internal/lexicons"
	"github.com/matt0x6f/hashpost/internal/middleware"
)

// Server represents the HashPost PDS server
type Server struct {
	config       *config.Config
	db           *generated.Queries
	eventStream  *EventStreamer
	httpServer   *http.Server
	logger       *slog.Logger
	authService  *AuthService
	oauthService *OAuthService
	dpopService  *DPoPService
	cidService   *CIDService
}

// NewServer creates a new PDS server instance
func NewServer(cfg *config.Config, db *generated.Queries) (*Server, error) {
	// Validate required configuration
	if cfg.PDS.Atproto.HandleBase == "" {
		return nil, fmt.Errorf("pds.atproto.handle_base is required in configuration")
	}

	// Create structured logger with text handler for better Docker visibility
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Get NATS URL from config
	natsURL := cfg.GetNATSURL()

	// Create event streamer
	logger.Info("Creating event streamer", "natsURL", natsURL)
	eventStream, err := NewEventStreamer(natsURL, logger)
	if err != nil {
		logger.Error("Failed to create event streamer", "error", err)
		return nil, fmt.Errorf("failed to create event streamer: %w", err)
	}
	logger.Info("Event streamer created successfully")

	// Create JWT service for production with shared signing key
	// Use a fixed signing key so AppView can validate PDS-issued tokens
	jwtService := jwtservice.NewProductionJWTService(jwtservice.JWTServiceConfig{
		Algorithm:          "ES256K",
		Expiration:         time.Hour,
		ValidateSignatures: true,
	})

	// Create authentication service
	authService := NewAuthService(db, logger, jwtService)

	// Create OAuth service
	oauthService := NewOAuthService(authService, db, logger)

	// Create DPoP service
	dpopService := NewDPoPService(authService, db, logger, jwtService)

	// Create CID service
	cidService := NewCIDService(logger)

	server := &Server{
		config:       cfg,
		db:           db,
		eventStream:  eventStream,
		logger:       logger,
		authService:  authService,
		oauthService: oauthService,
		dpopService:  dpopService,
		cidService:   cidService,
	}

	return server, nil
}

// Start starts the PDS server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register atproto endpoints
	s.registerAtprotoEndpoints(mux)

	// Wrap the mux with CORS and logging middleware
	handler := s.corsMiddleware(middleware.LoggingMiddleware(s.logger)(mux))

	s.httpServer = &http.Server{
		Addr:    s.config.GetPDSServerAddress(),
		Handler: handler,
	}

	s.logger.Info("Starting HashPost PDS server", "addr", s.config.GetPDSServerAddress())
	return s.httpServer.ListenAndServe()
}

// Stop stops the PDS server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// registerAtprotoEndpoints registers standard atproto endpoints
func (s *Server) registerAtprotoEndpoints(mux *http.ServeMux) {
	// Identity endpoints
	mux.HandleFunc("/xrpc/com.atproto.identity.resolveHandle", s.handleResolveHandle)

	// Server endpoints
	mux.HandleFunc("/xrpc/com.atproto.server.createSession", s.handleCreateSession)
	mux.HandleFunc("/xrpc/com.atproto.server.createAccount", s.handleCreateAccount)
	mux.HandleFunc("/xrpc/com.atproto.server.getSession", s.handleGetSession)
	mux.HandleFunc("/xrpc/com.atproto.server.refreshSession", s.handleRefreshSession)
	mux.HandleFunc("/xrpc/com.atproto.server.deleteSession", s.handleDeleteSession)
	mux.HandleFunc("/xrpc/com.atproto.server.updatePassword", s.handleUpdatePassword)
	mux.HandleFunc("/xrpc/com.atproto.server.describeServer", s.handleDescribeServer)

	// Repository endpoints
	mux.HandleFunc("/xrpc/com.atproto.repo.createRecord", s.handleCreateRecord)
	mux.HandleFunc("/xrpc/com.atproto.repo.getRecord", s.handleGetRecord)
	mux.HandleFunc("/xrpc/com.atproto.repo.listRecords", s.handleListRecords)
	mux.HandleFunc("/xrpc/com.atproto.repo.putRecord", s.handlePutRecord)
	mux.HandleFunc("/xrpc/com.atproto.repo.deleteRecord", s.handleDeleteRecord)

	// OAuth endpoints
	mux.HandleFunc("/oauth/client-metadata", s.oauthService.GetClientMetadata)
	mux.HandleFunc("/oauth/authorize", s.oauthService.HandleAuthorization)
	mux.HandleFunc("/oauth/token", s.oauthService.HandleToken)

	// DPoP endpoints
	mux.HandleFunc("/oauth/dpop-nonce", s.dpopService.GenerateNonce)
}

// Handler functions moved to domain-specific files:
// - Identity handlers: identity.go
// - Repository handlers: repo.go
// - Session handlers: auth.go
// - Account handlers: accounts.go

// Helper methods for record creation

func (s *Server) createHashPostRecord(ctx context.Context, repo, collection, recordID, uri, cid string, record map[string]interface{}) error {
	// For HashPost collections, we need to create the appropriate database record
	switch collection {
	case lexicons.CollectionFeedPost:
		return s.createPostRecord(ctx, repo, recordID, uri, cid, record)
	case lexicons.CollectionFeedSubforum:
		return s.createSubforumRecord(ctx, repo, recordID, uri, cid, record)
	case lexicons.CollectionFeedComment:
		return s.createCommentRecord(ctx, repo, recordID, uri, cid, record)
	case "com.hashpost.feed.vote":
		return s.createVoteRecord(ctx, repo, recordID, uri, cid, record)
	case "com.hashpost.graph.subscription":
		return s.createSubscriptionRecord(ctx, repo, recordID, uri, cid, record)
	default:
		return fmt.Errorf("unsupported HashPost collection: %s", collection)
	}
}

func (s *Server) createPostRecord(ctx context.Context, repo, recordID, uri, cid string, record map[string]interface{}) error {
	// Extract data from the record
	title, _ := record["title"].(string)
	content, _ := record["content"].(string)
	subforumSlug, _ := record["subforumSlug"].(string)

	// Look up the user by DID (repo)
	user, err := s.db.GetUserByDID(ctx, repo)
	if err != nil {
		s.logger.Error("Failed to find user by DID", "error", err, "did", repo)
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Get the subforum by slug
	subforum, err := s.db.GetSubforumBySlug(ctx, subforumSlug)
	if err != nil {
		s.logger.Error("Failed to find subforum", "error", err, "subforum_slug", subforumSlug)
		return fmt.Errorf("failed to find subforum: %w", err)
	}

	// Convert UUIDs to pgtype.UUID
	userIDPg := pgtype.UUID{Bytes: user.ID, Valid: true}
	subforumIDPg := pgtype.UUID{Bytes: subforum.ID, Valid: true}

	_, err = s.db.CreatePost(ctx, &generated.CreatePostParams{
		UserID:     userIDPg,
		SubforumID: subforumIDPg,
		Title:      title,
		Content:    content,
		AtprotoUri: &uri,
	})

	if err != nil {
		s.logger.Error("Failed to create post in database", "error", err, "uri", uri)
		return fmt.Errorf("failed to create post: %w", err)
	}

	s.logger.Info("Created post record", "uri", uri, "title", title, "content", content)
	return nil
}

func (s *Server) createCommentRecord(ctx context.Context, repo, recordID, uri, cid string, record map[string]interface{}) error {
	// Extract data from record
	text, _ := record[lexicons.FieldText].(string)
	postURI, _ := record[lexicons.FieldPost].(string)
	parentURI, _ := record[lexicons.FieldParent].(string) // optional

	// Look up user by DID
	user, err := s.db.GetUserByDID(ctx, repo)
	if err != nil {
		s.logger.Error("Failed to find user by DID", "error", err, "did", repo)
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Look up post by atproto URI
	post, err := s.db.GetPostByAtprotoURI(ctx, &postURI)
	if err != nil {
		s.logger.Error("Failed to find post by URI", "error", err, "post_uri", postURI)
		return fmt.Errorf("failed to find post: %w", err)
	}

	// Look up parent comment if specified
	var parentID pgtype.UUID
	if parentURI != "" {
		parent, err := s.db.GetCommentByURI(ctx, &parentURI)
		if err != nil {
			s.logger.Error("Failed to find parent comment by URI", "error", err, "parent_uri", parentURI)
			return fmt.Errorf("failed to find parent comment: %w", err)
		}
		parentID = pgtype.UUID{Bytes: parent.ID, Valid: true}
	}

	// Create comment
	_, err = s.db.CreateComment(ctx, &generated.CreateCommentParams{
		UserID:     pgtype.UUID{Bytes: user.ID, Valid: true},
		PostID:     pgtype.UUID{Bytes: post.ID, Valid: true},
		ParentID:   parentID,
		Content:    text,
		AtprotoUri: &uri,
	})

	if err != nil {
		s.logger.Error("Failed to create comment in database", "error", err, "uri", uri)
		return fmt.Errorf("failed to create comment: %w", err)
	}

	s.logger.Info("Created comment record", "uri", uri, "text", text, "post_uri", postURI)
	return nil
}

// validateAppViewToken validates tokens from the AppView service
func (s *Server) validateAppViewToken(token string) (*Session, error) {
	// Parse token without signature validation (AppView already validated it)
	parsedToken, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract user information
	did, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing subject (DID) in token")
	}

	handle, ok := claims["handle"].(string)
	if !ok {
		return nil, fmt.Errorf("missing handle in token")
	}

	// Check expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing expiration in token")
	}

	expirationTime := time.Unix(int64(exp), 0)
	if time.Now().After(expirationTime) {
		return nil, fmt.Errorf("token has expired")
	}

	// Create session from token claims
	session := &Session{
		ID:        "appview-service",
		DID:       did,
		Handle:    handle,
		CreatedAt: time.Now().Add(-1 * time.Hour), // Approximate
		ExpiresAt: expirationTime,
	}

	s.logger.Debug("Validated AppView token", "did", session.DID, "handle", session.Handle)
	return session, nil
}

func (s *Server) createSubforumRecord(ctx context.Context, repo, recordID, uri, cid string, record map[string]interface{}) error {
	// Extract data from the record
	name, _ := record[lexicons.FieldName].(string)
	slug, _ := record[lexicons.FieldSlug].(string)
	description, _ := record[lexicons.FieldDescription].(string)
	prefixType, _ := record[lexicons.FieldPrefixType].(string)

	if name == "" || slug == "" {
		return fmt.Errorf("name and slug are required for subforum")
	}

	// Validate prefix_type if provided
	if prefixType != "" && prefixType != "h" && prefixType != "r" && prefixType != "t" {
		return fmt.Errorf("invalid prefix_type: must be 'h', 'r', or 't'")
	}

	// Default to 't' if not provided
	if prefixType == "" {
		prefixType = "t"
	}

	// Look up the user by DID (repo)
	user, err := s.db.GetUserByDID(ctx, repo)
	if err != nil {
		s.logger.Error("Failed to find user by DID", "error", err, "did", repo)
		return fmt.Errorf("failed to find user: %w", err)
	}

	createdByPg := pgtype.UUID{Bytes: user.ID, Valid: true}

	_, err = s.db.CreateSubforum(ctx, &generated.CreateSubforumParams{
		Name:        name,
		Slug:        slug,
		Description: &description,
		CreatedBy:   createdByPg,
		PrefixType:  prefixType,
	})

	if err != nil {
		s.logger.Error("Failed to create subforum in database", "error", err, "uri", uri)
		return fmt.Errorf("failed to create subforum: %w", err)
	}

	s.logger.Info("Created subforum record", "uri", uri, "name", name, "slug", slug)
	return nil
}

func (s *Server) createVoteRecord(ctx context.Context, repo, recordID, uri, cid string, record map[string]interface{}) error {
	// Extract data from the record
	subject, _ := record["subject"].(string)
	direction, _ := record["direction"].(string)

	if subject == "" || direction == "" {
		return fmt.Errorf("subject and direction are required for vote")
	}

	if direction != "up" && direction != "down" {
		return fmt.Errorf("invalid direction: must be 'up' or 'down'")
	}

	// Look up the user by DID (repo)
	user, err := s.db.GetUserByDID(ctx, repo)
	if err != nil {
		s.logger.Error("Failed to find user by DID", "error", err, "did", repo)
		return fmt.Errorf("failed to find user: %w", err)
	}

	userIDPg := pgtype.UUID{Bytes: user.ID, Valid: true}

	// Try to find the subject as a post first
	post, err := s.db.GetPostByAtprotoURI(ctx, &subject)
	if err == nil {
		// Vote on post
		postIDPg := pgtype.UUID{Bytes: post.ID, Valid: true}

		// Check if user already voted on this post
		existingVote, err := s.db.GetVoteByUserAndPost(ctx, &generated.GetVoteByUserAndPostParams{
			UserID: userIDPg,
			PostID: postIDPg,
		})
		if err == nil && existingVote != nil {
			// User already voted, delete the existing vote first
			err = s.db.DeleteVote(ctx, existingVote.ID)
			if err != nil {
				s.logger.Error("Failed to delete existing vote", "error", err, "vote_id", existingVote.ID)
				return fmt.Errorf("failed to delete existing vote: %w", err)
			}
			s.logger.Info("Deleted existing vote", "vote_id", existingVote.ID, "post_id", post.ID)
		}

		_, err = s.db.CreateVote(ctx, &generated.CreateVoteParams{
			UserID:    userIDPg,
			PostID:    postIDPg,
			CommentID: pgtype.UUID{Valid: false}, // NULL for post votes
			VoteType:  direction,
		})
		if err != nil {
			s.logger.Error("Failed to create vote on post in database", "error", err, "uri", uri, "post_id", post.ID)
			return fmt.Errorf("failed to create vote on post: %w", err)
		}
		s.logger.Info("Created vote on post record", "uri", uri, "post_id", post.ID, "direction", direction)
		return nil
	}

	// Try to find as a comment
	comment, err := s.db.GetCommentByAtprotoURI(ctx, &subject)
	if err != nil {
		s.logger.Warn("Could not find post or comment for vote subject", "subject", subject)
		return fmt.Errorf("subject not found: %s", subject)
	}

	// Vote on comment
	commentIDPg := pgtype.UUID{Bytes: comment.ID, Valid: true}
	_, err = s.db.CreateVote(ctx, &generated.CreateVoteParams{
		UserID:    userIDPg,
		PostID:    pgtype.UUID{Valid: false}, // NULL for comment votes
		CommentID: commentIDPg,
		VoteType:  direction,
	})
	if err != nil {
		s.logger.Error("Failed to create vote on comment in database", "error", err, "uri", uri, "comment_id", comment.ID)
		return fmt.Errorf("failed to create vote on comment: %w", err)
	}

	s.logger.Info("Created vote on comment record", "uri", uri, "comment_id", comment.ID, "direction", direction)
	return nil
}

func (s *Server) createSubscriptionRecord(ctx context.Context, repo, recordID, uri, cid string, record map[string]interface{}) error {
	// For now, just log that we received the subscription record
	// In a real implementation, we'd store subscription data
	s.logger.Info("Received subscription record", "uri", uri)
	return nil
}

func (s *Server) createGenericRecord(ctx context.Context, repo, collection, recordID, uri, cid string, record map[string]interface{}) error {
	// For generic collections, we could store in a generic records table
	// For now, just log that we received the record
	s.logger.Info("Received generic record", "collection", collection, "uri", uri)
	return nil
}

// Helper methods for record retrieval

func (s *Server) getPostRecord(ctx context.Context, uri string) (map[string]interface{}, error) {
	post, err := s.db.GetPostByAtprotoURI(ctx, &uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// For now, use a simple CID (in production, this would be computed from the record content)
	cid := fmt.Sprintf("bafybeigdyrzt5sfpudm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi-%s", post.ID)

	// Format the timestamp properly
	var createdAtStr string
	if post.CreatedAt.Valid {
		createdAtStr = post.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z")
	} else {
		createdAtStr = "2025-01-01T00:00:00.000Z" // fallback
	}

	response := map[string]interface{}{
		"uri": uri,
		"cid": cid,
		"record": map[string]interface{}{
			"$type":     lexicons.RecordTypePost,
			"text":      post.Content,
			"createdAt": createdAtStr,
		},
	}

	return response, nil
}

func (s *Server) getSubforumRecord(ctx context.Context, uri string) (map[string]interface{}, error) {
	// Extract slug from URI (this is a simplified approach)
	// In production, we'd need a more robust way to map URIs to database records
	// For now, we'll need to implement a different approach or store the URI mapping

	// This is a placeholder - we'd need to implement proper URI to record mapping
	return nil, fmt.Errorf("subforum record retrieval not yet implemented")
}

// Helper methods for record listing

func (s *Server) listPostRecords(ctx context.Context, repo string, limit int, cursor string) (map[string]interface{}, error) {
	var posts []*generated.Post

	if cursor != "" {
		// Parse cursor as timestamp
		cursorTime, err := time.Parse(time.RFC3339, cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor format: %w", err)
		}

		// Use cursor-based pagination for the specific user
		posts, err = s.db.ListPostsWithCursorByUser(ctx, &generated.ListPostsWithCursorByUserParams{
			Did:       repo,
			CreatedAt: pgtype.Timestamptz{Time: cursorTime, Valid: true},
			Limit:     int32(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list posts with cursor: %w", err)
		}
	} else {
		// First page - get most recent posts for the specific user
		postRows, err := s.db.ListPostsByUser(ctx, &generated.ListPostsByUserParams{
			Did:    repo,
			Limit:  int32(limit),
			Offset: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list posts: %w", err)
		}

		// Convert rows to posts
		posts = make([]*generated.Post, len(postRows))
		for i, row := range postRows {
			posts[i] = &generated.Post{
				ID:         row.ID,
				UserID:     row.UserID,
				SubforumID: row.SubforumID,
				Title:      row.Title,
				Content:    row.Content,
				AtprotoUri: row.AtprotoUri,
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			}
		}
	}

	records := make([]map[string]interface{}, 0, len(posts))
	for _, post := range posts {
		// Format the timestamp properly
		var createdAtStr string
		if post.CreatedAt.Valid {
			createdAtStr = post.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z")
		} else {
			createdAtStr = "2025-01-01T00:00:00.000Z" // fallback
		}

		// Generate URI from post data
		uri := fmt.Sprintf("at://%s/%s/%s", repo, lexicons.PathFeedPost, post.ID)
		cid := fmt.Sprintf("bafybeigdyrzt5sfpudm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi-%s", post.ID)

		record := map[string]interface{}{
			"uri": uri,
			"cid": cid,
			"record": map[string]interface{}{
				"$type":     lexicons.RecordTypePost,
				"text":      post.Content,
				"createdAt": createdAtStr,
			},
		}
		records = append(records, record)
	}

	response := map[string]interface{}{
		"records": records,
		"cursor":  s.generateNextCursor(records, limit),
	}

	return response, nil
}

func (s *Server) listSubforumRecords(ctx context.Context, repo string, limit int, cursor string) (map[string]interface{}, error) {
	var subforums []*generated.Subforum

	if cursor != "" {
		// Parse cursor as timestamp
		cursorTime, err := time.Parse(time.RFC3339, cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor format: %w", err)
		}

		// Use cursor-based pagination
		subforums, err = s.db.ListSubforumsWithCursor(ctx, &generated.ListSubforumsWithCursorParams{
			CreatedAt: pgtype.Timestamptz{Time: cursorTime, Valid: true},
			Limit:     int32(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list subforums with cursor: %w", err)
		}
	} else {
		// First page - get most recent subforums
		subforumRows, err := s.db.ListSubforums(ctx, &generated.ListSubforumsParams{
			Limit:  int32(limit),
			Offset: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list subforums: %w", err)
		}

		// Convert rows to subforums
		subforums = make([]*generated.Subforum, len(subforumRows))
		for i, row := range subforumRows {
			subforums[i] = &generated.Subforum{
				ID:          row.ID,
				Name:        row.Name,
				Slug:        row.Slug,
				Description: row.Description,
				CreatedBy:   row.CreatedBy,
				CreatedAt:   row.CreatedAt,
				UpdatedAt:   row.UpdatedAt,
			}
		}
	}

	records := make([]map[string]interface{}, 0, len(subforums))
	for _, subforum := range subforums {
		// Format the timestamp properly
		var createdAtStr string
		if subforum.CreatedAt.Valid {
			createdAtStr = subforum.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z")
		} else {
			createdAtStr = "2025-01-01T00:00:00.000Z" // fallback
		}

		// Generate URI from subforum data
		uri := fmt.Sprintf("at://%s/%s/%s", repo, lexicons.PathFeedSubforum, subforum.ID)
		cid := fmt.Sprintf("bafybeigdyrzt5sfpudm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi-%s", subforum.ID)

		record := map[string]interface{}{
			"uri": uri,
			"cid": cid,
			"record": map[string]interface{}{
				"$type":       lexicons.RecordTypeSubforum,
				"name":        subforum.Name,
				"slug":        subforum.Slug,
				"description": subforum.Description,
				"createdAt":   createdAtStr,
			},
		}
		records = append(records, record)
	}

	response := map[string]interface{}{
		"records": records,
		"cursor":  s.generateNextCursor(records, limit),
	}

	return response, nil
}

// Helper methods for record updates

func (s *Server) updatePostRecord(ctx context.Context, uri string, record map[string]interface{}) (map[string]interface{}, error) {
	// Extract data from the record
	title, _ := record["title"].(string)
	text, _ := record[lexicons.FieldText].(string)

	if text == "" {
		return nil, fmt.Errorf("text is required for post update")
	}

	// Update the post in database
	updatedPost, err := s.db.UpdatePostByAtprotoURI(ctx, &generated.UpdatePostByAtprotoURIParams{
		AtprotoUri: &uri,
		Title:      title, // Use the title field for title
		Content:    text,  // Use the text field for content
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Generate new CID
	cid := fmt.Sprintf("bafybeigdyrzt5sfpudm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi-%s", updatedPost.ID)

	// Format the timestamp properly
	var updatedAtStr string
	if updatedPost.UpdatedAt.Valid {
		updatedAtStr = updatedPost.UpdatedAt.Time.Format("2006-01-02T15:04:05.000Z")
	} else {
		updatedAtStr = "2025-01-01T00:00:00.000Z" // fallback
	}

	response := map[string]interface{}{
		"uri": uri,
		"cid": cid,
		"record": map[string]interface{}{
			"$type":     lexicons.RecordTypePost,
			"text":      updatedPost.Content,
			"createdAt": updatedAtStr,
		},
	}

	s.logger.Info("Updated post record", "uri", uri, "text", text)
	return response, nil
}

func (s *Server) updateSubforumRecord(ctx context.Context, rkey string, record map[string]interface{}) (map[string]interface{}, error) {
	// Extract data from the record
	name, _ := record[lexicons.FieldName].(string)
	slug, _ := record[lexicons.FieldSlug].(string)
	description, _ := record[lexicons.FieldDescription].(string)

	if name == "" || slug == "" {
		return nil, fmt.Errorf("name and slug are required for subforum update")
	}

	// Convert rkey to UUID (assuming rkey is the subforum ID)
	subforumID, err := uuid.Parse(rkey)
	if err != nil {
		return nil, fmt.Errorf("invalid subforum ID: %w", err)
	}

	// Update the subforum in database
	updatedSubforum, err := s.db.UpdateSubforumByID(ctx, &generated.UpdateSubforumByIDParams{
		ID:          subforumID,
		Name:        name,
		Slug:        slug,
		Description: &description,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update subforum: %w", err)
	}

	// Generate new CID
	cid := fmt.Sprintf("bafybeigdyrzt5sfpudm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi-%s", updatedSubforum.ID)

	// Format the timestamp properly
	var updatedAtStr string
	if updatedSubforum.UpdatedAt.Valid {
		updatedAtStr = updatedSubforum.UpdatedAt.Time.Format("2006-01-02T15:04:05.000Z")
	} else {
		updatedAtStr = "2025-01-01T00:00:00.000Z" // fallback
	}

	// Generate URI
	uri := fmt.Sprintf("at://%s/%s/%s", "did:plc:hashpost-binding-test", lexicons.PathFeedSubforum, rkey)

	response := map[string]interface{}{
		"uri": uri,
		"cid": cid,
		"record": map[string]interface{}{
			"$type":       lexicons.RecordTypeSubforum,
			"name":        updatedSubforum.Name,
			"slug":        updatedSubforum.Slug,
			"description": updatedSubforum.Description,
			"createdAt":   updatedAtStr,
		},
	}

	s.logger.Info("Updated subforum record", "uri", uri, "name", name, "slug", slug)
	return response, nil
}

// Helper methods for record deletion

func (s *Server) deletePostRecord(ctx context.Context, uri string) error {
	// Delete the post from database
	err := s.db.DeletePostByAtprotoURI(ctx, &uri)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	s.logger.Info("Deleted post record", "uri", uri)
	return nil
}

func (s *Server) deleteSubforumRecord(ctx context.Context, rkey string) error {
	// Convert rkey to UUID (assuming rkey is the subforum ID)
	subforumID, err := uuid.Parse(rkey)
	if err != nil {
		return fmt.Errorf("invalid subforum ID: %w", err)
	}

	// Delete the subforum from database
	err = s.db.DeleteSubforumByID(ctx, subforumID)
	if err != nil {
		return fmt.Errorf("failed to delete subforum: %w", err)
	}

	s.logger.Info("Deleted subforum record", "rkey", rkey)
	return nil
}

func (s *Server) deleteCommentRecord(ctx context.Context, uri string) error {
	// Delete the comment from database
	err := s.db.DeleteCommentByAtprotoURI(ctx, &uri)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	s.logger.Info("Deleted comment record", "uri", uri)
	return nil
}

// getUserEmailFromDID retrieves the user's email from the database by DID
func (s *Server) getUserEmailFromDID(ctx context.Context, did string) string {
	user, err := s.db.GetUserByDID(ctx, did)
	if err != nil {
		s.logger.Error("Failed to get user by DID", "error", err, "did", did)
		return "" // Return empty string if user not found
	}

	if user.Email != nil {
		return *user.Email
	}

	return "" // Return empty string if no email
}

// validateInviteCode validates an invite code
func (s *Server) validateInviteCode(ctx context.Context, inviteCode string) error {
	// For now, implement a simple validation
	// In production, this would check against a database table of valid invite codes
	if len(inviteCode) < 8 {
		return fmt.Errorf("invite code too short")
	}

	// Simple validation - in production, check against database
	validCodes := []string{"hashpost2024", "beta-invite", "early-access"}
	for _, validCode := range validCodes {
		if inviteCode == validCode {
			return nil
		}
	}

	return fmt.Errorf("invalid invite code")
}

// generateNextCursor generates the next cursor for pagination
func (s *Server) generateNextCursor(records []map[string]interface{}, limit int) string {
	if len(records) > 0 && len(records) == limit {
		// If we got a full page, there might be more records
		lastRecord := records[len(records)-1]
		if createdAt, ok := lastRecord["record"].(map[string]interface{})["createdAt"].(string); ok {
			return createdAt
		}
	}
	return ""
}

// corsMiddleware handles CORS headers for cross-origin requests
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
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
