package testutil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/matt0x6f/hashpost/internal/api"
	"github.com/matt0x6f/hashpost/internal/api/logger"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/routes"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
)

// TestEntityTracker tracks all entities created during a test for proper cleanup
type TestEntityTracker struct {
	mu sync.Mutex
	// Track entities by type and ID for proper cleanup order
	users        map[int64]bool
	pseudonyms   map[string]bool
	subforums    map[int64]bool
	posts        map[int64]bool
	comments     map[int64]bool
	votes        map[int64]bool
	apiKeys      map[string]bool
	userBlocks   map[int64]bool
	userPrefs    map[int64]bool
	reports      map[int64]bool
	modActions   map[int64]bool
	correlations map[int64]bool
}

// NewTestEntityTracker creates a new entity tracker
func NewTestEntityTracker() *TestEntityTracker {
	return &TestEntityTracker{
		users:        make(map[int64]bool),
		pseudonyms:   make(map[string]bool),
		subforums:    make(map[int64]bool),
		posts:        make(map[int64]bool),
		comments:     make(map[int64]bool),
		votes:        make(map[int64]bool),
		apiKeys:      make(map[string]bool),
		userBlocks:   make(map[int64]bool),
		userPrefs:    make(map[int64]bool),
		reports:      make(map[int64]bool),
		modActions:   make(map[int64]bool),
		correlations: make(map[int64]bool),
	}
}

// TrackUser marks a user as created for cleanup
func (t *TestEntityTracker) TrackUser(userID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.users[userID] = true
}

// TrackPseudonym marks a pseudonym as created for cleanup
func (t *TestEntityTracker) TrackPseudonym(pseudonymID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pseudonyms[pseudonymID] = true
}

// TrackSubforum marks a subforum as created for cleanup
func (t *TestEntityTracker) TrackSubforum(subforumID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.subforums[subforumID] = true
}

// TrackPost marks a post as created for cleanup
func (t *TestEntityTracker) TrackPost(postID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.posts[postID] = true
}

// TrackComment marks a comment as created for cleanup
func (t *TestEntityTracker) TrackComment(commentID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.comments[commentID] = true
}

// TrackVote marks a vote as created for cleanup
func (t *TestEntityTracker) TrackVote(voteID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.votes[voteID] = true
}

// TrackAPIKey marks an API key as created for cleanup
func (t *TestEntityTracker) TrackAPIKey(apiKeyID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.apiKeys[apiKeyID] = true
}

// TrackUserBlock marks a user block as created for cleanup
func (t *TestEntityTracker) TrackUserBlock(blockID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.userBlocks[blockID] = true
}

// TrackUserPref marks user preferences as created for cleanup
func (t *TestEntityTracker) TrackUserPref(prefID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.userPrefs[prefID] = true
}

// TrackReport marks a report as created for cleanup
func (t *TestEntityTracker) TrackReport(reportID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reports[reportID] = true
}

// TrackModAction marks a moderation action as created for cleanup
func (t *TestEntityTracker) TrackModAction(actionID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.modActions[actionID] = true
}

// TrackCorrelation marks a correlation as created for cleanup
func (t *TestEntityTracker) TrackCorrelation(correlationID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.correlations[correlationID] = true
}

// Cleanup removes all tracked entities in the correct order
func (t *TestEntityTracker) Cleanup(ctx context.Context, db bob.DB) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Clean up in reverse dependency order
	// 1. Votes (depend on posts/comments)
	for voteID := range t.votes {
		if _, err := db.ExecContext(ctx, "DELETE FROM votes WHERE vote_id = $1", voteID); err != nil {
			return fmt.Errorf("failed to cleanup vote %d: %w", voteID, err)
		}
	}

	// 2. Comments (depend on posts)
	for commentID := range t.comments {
		if _, err := db.ExecContext(ctx, "DELETE FROM comments WHERE comment_id = $1", commentID); err != nil {
			return fmt.Errorf("failed to cleanup comment %d: %w", commentID, err)
		}
	}

	// 3. Posts (depend on subforums)
	for postID := range t.posts {
		if _, err := db.ExecContext(ctx, "DELETE FROM posts WHERE post_id = $1", postID); err != nil {
			return fmt.Errorf("failed to cleanup post %d: %w", postID, err)
		}
	}

	// 4. User blocks
	for blockID := range t.userBlocks {
		if _, err := db.ExecContext(ctx, "DELETE FROM user_blocks WHERE block_id = $1", blockID); err != nil {
			return fmt.Errorf("failed to cleanup user block %d: %w", blockID, err)
		}
	}

	// 5. User preferences
	for prefID := range t.userPrefs {
		if _, err := db.ExecContext(ctx, "DELETE FROM user_preferences WHERE preference_id = $1", prefID); err != nil {
			return fmt.Errorf("failed to cleanup user preference %d: %w", prefID, err)
		}
	}

	// 6. Reports
	for reportID := range t.reports {
		if _, err := db.ExecContext(ctx, "DELETE FROM reports WHERE report_id = $1", reportID); err != nil {
			return fmt.Errorf("failed to cleanup report %d: %w", reportID, err)
		}
	}

	// 7. Moderation actions
	for actionID := range t.modActions {
		if _, err := db.ExecContext(ctx, "DELETE FROM moderation_actions WHERE action_id = $1", actionID); err != nil {
			return fmt.Errorf("failed to cleanup moderation action %d: %w", actionID, err)
		}
	}

	// 8. Correlations
	for correlationID := range t.correlations {
		if _, err := db.ExecContext(ctx, "DELETE FROM compliance_correlations WHERE correlation_id = $1", correlationID); err != nil {
			return fmt.Errorf("failed to cleanup correlation %d: %w", correlationID, err)
		}
	}

	// 9. API keys
	for apiKeyID := range t.apiKeys {
		if _, err := db.ExecContext(ctx, "DELETE FROM api_keys WHERE api_key_id = $1", apiKeyID); err != nil {
			return fmt.Errorf("failed to cleanup API key %s: %w", apiKeyID, err)
		}
	}

	// 10. Subforums (depend on users for moderators)
	for subforumID := range t.subforums {
		if _, err := db.ExecContext(ctx, "DELETE FROM subforums WHERE subforum_id = $1", subforumID); err != nil {
			return fmt.Errorf("failed to cleanup subforum %d: %w", subforumID, err)
		}
	}

	// 11. Pseudonyms (depend on users)
	for pseudonymID := range t.pseudonyms {
		if _, err := db.ExecContext(ctx, "DELETE FROM pseudonyms WHERE pseudonym_id = $1", pseudonymID); err != nil {
			return fmt.Errorf("failed to cleanup pseudonym %s: %w", pseudonymID, err)
		}
	}

	// 12. Users (clean up last as everything depends on them)
	for userID := range t.users {
		if _, err := db.ExecContext(ctx, "DELETE FROM users WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("failed to cleanup user %d: %w", userID, err)
		}
	}

	return nil
}

// IntegrationTestSuite provides utilities for integration testing
type IntegrationTestSuite struct {
	DB           bob.DB
	Server       *api.Server
	Config       *config.Config
	UserDAO      *dao.UserDAO
	PseudonymDAO *dao.PseudonymDAO
	SubforumDAO  *dao.SubforumDAO
	PostDAO      *dao.PostDAO
	CommentDAO   *dao.CommentDAO
	VoteDAO      *dao.VoteDAO
	APIKeyDAO    *dao.APIKeyDAO
	UserBlockDAO *dao.UserBlocksDAO
	UserPrefDAO  *dao.UserPreferencesDAO
	Tracker      *TestEntityTracker
	Cleanup      func()
}

// TestUser represents a test user for integration tests
type TestUser struct {
	UserID       int64
	Email        string
	Password     string
	PasswordHash string
	Roles        []string
	Capabilities []string
	PseudonymID  string
	DisplayName  string
}

// TestSubforum represents a test subforum for integration tests
type TestSubforum struct {
	SubforumID   int64
	Name         string
	Description  string
	CreatedBy    int64
	IsPrivate    bool
	ModeratorIDs []int64
}

// TestPost represents a test post for integration tests
type TestPost struct {
	PostID      int64
	Title       string
	Content     string
	SubforumID  int64
	AuthorID    int64
	PseudonymID string
}

// TestComment represents a test comment for integration tests
type TestComment struct {
	CommentID   int64
	Content     string
	PostID      int64
	AuthorID    int64
	PseudonymID string
	ParentID    *int64
}

// parseDSN parses a PostgreSQL DSN and returns database configuration
func parseDSN(dsn string) (*config.DatabaseConfig, error) {
	// Parse the URL
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Extract components
	host := u.Hostname()
	port := 5432 // default PostgreSQL port
	if u.Port() != "" {
		if p, err := strconv.Atoi(u.Port()); err == nil {
			port = p
		}
	}

	// Extract username and password
	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	// Extract database name
	database := strings.TrimPrefix(u.Path, "/")

	// Extract SSL mode from query parameters
	sslMode := "disable"
	if u.Query().Get("sslmode") != "" {
		sslMode = u.Query().Get("sslmode")
	}

	return &config.DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     username,
		Password: password,
		Database: database,
		SSLMode:  sslMode,
	}, nil
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	// Check if we have a database URL for testing
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL environment variable not set. Skipping integration test.")
		return nil
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger with config level
	logger.InitWithLevel(cfg.Logging.Level)

	// Parse the database URL to override database configuration
	if databaseURL != "" {
		dbConfig, err := parseDSN(databaseURL)
		if err != nil {
			t.Fatalf("Failed to parse DATABASE_URL: %v", err)
		}
		cfg.Database = *dbConfig
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create server
	server := api.NewServer()

	// Create DAOs
	userDAO := dao.NewUserDAO(db)
	pseudonymDAO := dao.NewPseudonymDAO(db)
	subforumDAO := dao.NewSubforumDAO(db)
	postDAO := dao.NewPostDAO(db)
	commentDAO := dao.NewCommentDAO(db)
	voteDAO := dao.NewVoteDAO(db)
	apiKeyDAO := dao.NewAPIKeyDAO(db)
	userBlockDAO := dao.NewUserBlocksDAO(db)
	userPrefDAO := dao.NewUserPreferencesDAO(db)

	// Create entity tracker
	tracker := NewTestEntityTracker()

	// Return test suite
	return &IntegrationTestSuite{
		DB:           db,
		Server:       server,
		Config:       cfg,
		UserDAO:      userDAO,
		PseudonymDAO: pseudonymDAO,
		SubforumDAO:  subforumDAO,
		PostDAO:      postDAO,
		CommentDAO:   commentDAO,
		VoteDAO:      voteDAO,
		APIKeyDAO:    apiKeyDAO,
		UserBlockDAO: userBlockDAO,
		UserPrefDAO:  userPrefDAO,
		Tracker:      tracker,
		Cleanup: func() {
			// Clean up test data
			ctx := context.Background()
			if err := tracker.Cleanup(ctx, db); err != nil {
				t.Logf("Warning: failed to cleanup test data: %v", err)
			}
		},
	}
}

// CreateTestUser creates a test user in the database and tracks it for cleanup
func (ts *IntegrationTestSuite) CreateTestUser(t *testing.T, email, password string, roles []string) *TestUser {
	ctx := context.Background()

	// Hash password
	passwordHash := hashPassword(password)

	// Create user
	user, err := ts.UserDAO.CreateUser(ctx, email, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Track for cleanup
	ts.Tracker.TrackUser(user.UserID)

	// Set roles and capabilities if provided
	if len(roles) > 0 {
		rolesJSON, _ := json.Marshal(roles)
		capabilities := getCapabilitiesForRoles(roles)
		capabilitiesJSON, _ := json.Marshal(capabilities)

		rolesNull := sql.Null[types.JSON[json.RawMessage]]{}
		rolesNull.Scan(rolesJSON)

		capabilitiesNull := sql.Null[types.JSON[json.RawMessage]]{}
		capabilitiesNull.Scan(capabilitiesJSON)

		updates := &dbmodels.UserSetter{
			Roles:        &rolesNull,
			Capabilities: &capabilitiesNull,
		}

		if err := ts.UserDAO.UpdateUser(ctx, user.UserID, updates); err != nil {
			t.Fatalf("Failed to update test user roles: %v", err)
		}
	}

	// Create pseudonym for the user
	displayName := fmt.Sprintf("test_user_%d", user.UserID)
	pseudonym, err := ts.PseudonymDAO.CreatePseudonym(ctx, user.UserID, displayName)
	if err != nil {
		t.Fatalf("Failed to create test user pseudonym: %v", err)
	}

	// Track pseudonym for cleanup
	ts.Tracker.TrackPseudonym(pseudonym.PseudonymID)

	return &TestUser{
		UserID:       user.UserID,
		Email:        email,
		Password:     password,
		PasswordHash: passwordHash,
		Roles:        roles,
		Capabilities: getCapabilitiesForRoles(roles),
		PseudonymID:  pseudonym.PseudonymID,
		DisplayName:  displayName,
	}
}

// CreateTestSubforum creates a test subforum and tracks it for cleanup
func (ts *IntegrationTestSuite) CreateTestSubforum(t *testing.T, name, description string, createdBy int64, isPrivate bool) *TestSubforum {
	ctx := context.Background()

	// Create subforum with required parameters
	displayName := name
	sidebarText := ""
	rulesText := ""
	isNSFW := false
	isRestricted := false

	subforum, err := ts.SubforumDAO.CreateSubforum(ctx, name, displayName, description, sidebarText, rulesText, isNSFW, isPrivate, isRestricted)
	if err != nil {
		t.Fatalf("Failed to create test subforum: %v", err)
	}

	// Track for cleanup
	ts.Tracker.TrackSubforum(int64(subforum.SubforumID))

	return &TestSubforum{
		SubforumID:  int64(subforum.SubforumID),
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		IsPrivate:   isPrivate,
	}
}

// CreateTestPost creates a test post and tracks it for cleanup
func (ts *IntegrationTestSuite) CreateTestPost(t *testing.T, title, content string, subforumID, authorID int64, pseudonymID string) *TestPost {
	ctx := context.Background()

	// Create post with required parameters
	postType := "text"
	var url *string = nil
	isNSFW := false
	isSpoiler := false

	post, err := ts.PostDAO.CreatePost(ctx, int32(subforumID), pseudonymID, title, content, postType, url, isNSFW, isSpoiler)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Track for cleanup
	ts.Tracker.TrackPost(post.PostID)

	return &TestPost{
		PostID:      post.PostID,
		Title:       title,
		Content:     content,
		SubforumID:  int64(subforumID),
		AuthorID:    authorID,
		PseudonymID: pseudonymID,
	}
}

// CreateTestComment creates a test comment and tracks it for cleanup
func (ts *IntegrationTestSuite) CreateTestComment(t *testing.T, content string, postID, authorID int64, pseudonymID string, parentID *int64) *TestComment {
	ctx := context.Background()

	// Create comment
	comment, err := ts.CommentDAO.CreateComment(ctx, postID, pseudonymID, content, parentID)
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	// Track for cleanup
	ts.Tracker.TrackComment(comment.CommentID)

	return &TestComment{
		CommentID:   comment.CommentID,
		Content:     content,
		PostID:      postID,
		AuthorID:    authorID,
		PseudonymID: pseudonymID,
		ParentID:    parentID,
	}
}

// CreateTestVote creates a test vote and tracks it for cleanup
func (ts *IntegrationTestSuite) CreateTestVote(t *testing.T, userID, postID int64, voteType string) *dbmodels.Vote {
	ctx := context.Background()

	// Convert vote type to integer value
	var voteValue int32
	switch voteType {
	case "upvote":
		voteValue = 1
	case "downvote":
		voteValue = -1
	default:
		t.Fatalf("Invalid vote type: %s", voteType)
	}

	// Get user's pseudonym for voting
	user, err := ts.UserDAO.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get user for vote: %v", err)
	}
	if user == nil {
		t.Fatalf("User not found for vote")
	}

	// Get user's pseudonym
	pseudonyms, err := ts.PseudonymDAO.GetPseudonymsByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get user pseudonyms for vote: %v", err)
	}
	if len(pseudonyms) == 0 {
		t.Fatalf("No pseudonym found for user")
	}
	pseudonymID := pseudonyms[0].PseudonymID

	// Create vote
	vote, err := ts.VoteDAO.CreateVote(ctx, pseudonymID, "post", postID, voteValue)
	if err != nil {
		t.Fatalf("Failed to create test vote: %v", err)
	}

	// Track for cleanup
	ts.Tracker.TrackVote(vote.VoteID)

	return vote
}

// CreateTestAPIKey creates a test API key and tracks it for cleanup
func (ts *IntegrationTestSuite) CreateTestAPIKey(t *testing.T, userID int64, pseudonymID string, permissions map[string]interface{}) *dbmodels.APIKey {
	ctx := context.Background()

	// Convert permissions map to APIKeyPermissions struct
	var roles []string
	var capabilities []string

	if rolesVal, ok := permissions["roles"].([]string); ok {
		roles = rolesVal
	}
	if capsVal, ok := permissions["capabilities"].([]string); ok {
		capabilities = capsVal
	}

	apiKeyPermissions := &dao.APIKeyPermissions{
		Roles:        roles,
		Capabilities: capabilities,
	}

	// Generate a random API key
	keyName := fmt.Sprintf("test_key_%d", userID)
	rawKey := fmt.Sprintf("test_api_key_%d_%s", userID, pseudonymID)

	// Create API key
	apiKey, err := ts.APIKeyDAO.CreateAPIKey(ctx, keyName, rawKey, pseudonymID, apiKeyPermissions, nil)
	if err != nil {
		t.Fatalf("Failed to create test API key: %v", err)
	}

	// Track for cleanup
	ts.Tracker.TrackAPIKey(fmt.Sprintf("%d", apiKey.KeyID))

	return apiKey
}

// CreateTestPseudonym creates a test pseudonym and tracks it for cleanup
func (ts *IntegrationTestSuite) CreateTestPseudonym(t *testing.T, userID int64, displayName string) *dbmodels.Pseudonym {
	ctx := context.Background()

	// Create pseudonym
	pseudonym, err := ts.PseudonymDAO.CreatePseudonym(ctx, userID, displayName)
	if err != nil {
		t.Fatalf("Failed to create test pseudonym: %v", err)
	}

	// Track for cleanup
	ts.Tracker.TrackPseudonym(pseudonym.PseudonymID)

	return pseudonym
}

// CreateTestServer creates an HTTP test server for integration tests
func (ts *IntegrationTestSuite) CreateTestServer() *httptest.Server {
	// Create a new server with the test database configuration
	server := ts.createTestServer()
	return httptest.NewServer(server.GetHandler())
}

// createTestServer creates a new API server with test database configuration
func (ts *IntegrationTestSuite) createTestServer() *api.Server {
	// Create database connection using test configuration
	db, err := database.NewConnection(&ts.Config.Database)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
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

	// Create auth middleware with test configuration
	authMiddleware := middleware.NewAuthMiddleware(ts.Config.JWT.Secret, apiKeyDAO, &ts.Config.JWT, &ts.Config.Security)

	// Set the global auth middleware for Huma functions
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	// Create a new HTTP mux
	mux := http.NewServeMux()

	// Create Huma configuration
	config := huma.DefaultConfig("HashPost API", "1.0.0")

	// Create a new Huma API with humago adapter
	humaAPI := humago.New(mux, config)

	// Add router-agnostic middleware
	humaAPI.UseMiddleware(middleware.LoggingMiddleware)
	humaAPI.UseMiddleware(middleware.CORSMiddleware(&ts.Config.CORS))

	// Add authentication middleware to extract user context
	humaAPI.UseMiddleware(middleware.AuthenticateUserHuma)
	log.Info().Str("jwt_secret_length", fmt.Sprintf("%d", len(ts.Config.JWT.Secret))).Msg("JWT configuration loaded")

	// Register routes with test configuration
	routes.RegisterHealthRoutes(humaAPI)
	routes.RegisterHelloRoutes(humaAPI)
	routes.RegisterAuthRoutes(humaAPI, ts.Config, db)
	routes.RegisterUserRoutes(humaAPI, userDAO, pseudonymDAO, userPreferencesDAO, userBlocksDAO, postDAO, commentDAO)
	routes.RegisterSubforumRoutes(humaAPI, db)
	routes.RegisterMessagesRoutes(humaAPI)
	routes.RegisterSearchRoutes(humaAPI)
	routes.RegisterModerationRoutes(humaAPI)
	routes.RegisterContentRoutes(humaAPI, db)
	routes.RegisterCorrelationRoutes(humaAPI, db, ibeSystem, pseudonymDAO, identityMappingDAO, postDAO, commentDAO)

	return &api.Server{
		API:       humaAPI,
		Mux:       mux,
		Config:    config,
		AppConfig: ts.Config,
	}
}

// LoginUser performs a login request and returns the response
func (ts *IntegrationTestSuite) LoginUser(t *testing.T, server *httptest.Server, email, password string) *http.Response {
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}

	resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to make login request: %v", err)
	}

	return resp
}

// ExtractTokenFromResponse extracts the access token from a login response
func (ts *IntegrationTestSuite) ExtractTokenFromResponse(t *testing.T, resp *http.Response) string {
	// Read the response body into a buffer so it can be read multiple times
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Create a new reader from the buffer for JSON decoding
	bodyReader := bytes.NewReader(bodyBytes)

	var responseBody map[string]interface{}
	if err := json.NewDecoder(bodyReader).Decode(&responseBody); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// With the new Huma response structure, access_token is directly in the response
	if accessToken, ok := responseBody["access_token"].(string); ok {
		// Restore the response body so it can be read again
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		return accessToken
	}

	// Fallback: try the old nested structure just in case
	if body, ok := responseBody["body"].(map[string]interface{}); ok {
		if accessToken, ok := body["access_token"].(string); ok {
			// Restore the response body so it can be read again
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			return accessToken
		}
	}

	// Restore the response body so it can be read again
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	t.Fatalf("Failed to extract access token from response: %+v", responseBody)
	return ""
}

// MakeAuthenticatedRequest makes an authenticated HTTP request
func (ts *IntegrationTestSuite) MakeAuthenticatedRequest(t *testing.T, server *httptest.Server, method, path, token string, body interface{}) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

// Helper functions

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func getCapabilitiesForRoles(roles []string) []string {
	capabilities := []string{}
	for _, role := range roles {
		switch role {
		case "platform_admin":
			capabilities = append(capabilities, "system_admin", "user_management", "correlate_identities", "access_private_subforums", "cross_platform_access", "system_moderation")
		case "trust_safety":
			capabilities = append(capabilities, "correlate_identities", "cross_platform_access", "system_moderation", "harassment_investigation")
		case "legal_team":
			capabilities = append(capabilities, "correlate_identities", "legal_compliance", "court_orders", "cross_platform_access")
		default:
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "create_subforum")
		}
	}
	return capabilities
}
