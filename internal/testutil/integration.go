package testutil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	"time"

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
	roleKeys     map[string]bool // Track role keys by key ID
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
		roleKeys:     make(map[string]bool),
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

// TrackRoleKey marks a role key as created for cleanup
func (t *TestEntityTracker) TrackRoleKey(keyID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.roleKeys[keyID] = true
}

// Cleanup removes all tracked entities in the correct order
func (t *TestEntityTracker) Cleanup(ctx context.Context, db bob.DB) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	log.Info().Msg("[TestEntityTracker] Starting cleanup of test entities...")

	// Clean up in reverse dependency order
	// 1. Votes (depend on posts/comments)
	for voteID := range t.votes {
		log.Info().Int64("vote_id", voteID).Msg("[TestEntityTracker] Deleting vote")
		if _, err := db.ExecContext(ctx, "DELETE FROM votes WHERE vote_id = $1", voteID); err != nil {
			return fmt.Errorf("failed to cleanup vote %d: %w", voteID, err)
		}
	}

	// 2. Comments (depend on posts)
	for commentID := range t.comments {
		log.Info().Int64("comment_id", commentID).Msg("[TestEntityTracker] Deleting comment")
		if _, err := db.ExecContext(ctx, "DELETE FROM comments WHERE comment_id = $1", commentID); err != nil {
			return fmt.Errorf("failed to cleanup comment %d: %w", commentID, err)
		}
	}

	// 3. Posts (depend on subforums)
	for postID := range t.posts {
		log.Info().Int64("post_id", postID).Msg("[TestEntityTracker] Deleting post")
		if _, err := db.ExecContext(ctx, "DELETE FROM posts WHERE post_id = $1", postID); err != nil {
			return fmt.Errorf("failed to cleanup post %d: %w", postID, err)
		}
	}

	// 4. User blocks
	for blockID := range t.userBlocks {
		log.Info().Int64("block_id", blockID).Msg("[TestEntityTracker] Deleting user block")
		if _, err := db.ExecContext(ctx, "DELETE FROM user_blocks WHERE block_id = $1", blockID); err != nil {
			return fmt.Errorf("failed to cleanup user block %d: %w", blockID, err)
		}
	}

	// 5. User preferences
	for prefID := range t.userPrefs {
		log.Info().Int64("pref_id", prefID).Msg("[TestEntityTracker] Deleting user preferences")
		if _, err := db.ExecContext(ctx, "DELETE FROM user_preferences WHERE preference_id = $1", prefID); err != nil {
			return fmt.Errorf("failed to cleanup user preferences %d: %w", prefID, err)
		}
	}

	// 6. Reports
	for reportID := range t.reports {
		log.Info().Int64("report_id", reportID).Msg("[TestEntityTracker] Deleting report")
		if _, err := db.ExecContext(ctx, "DELETE FROM reports WHERE report_id = $1", reportID); err != nil {
			return fmt.Errorf("failed to cleanup report %d: %w", reportID, err)
		}
	}

	// 7. Moderation actions
	for actionID := range t.modActions {
		log.Info().Int64("action_id", actionID).Msg("[TestEntityTracker] Deleting moderation action")
		if _, err := db.ExecContext(ctx, "DELETE FROM moderation_actions WHERE action_id = $1", actionID); err != nil {
			return fmt.Errorf("failed to cleanup moderation action %d: %w", actionID, err)
		}
	}

	// 8. Correlations
	for correlationID := range t.correlations {
		log.Info().Int64("correlation_id", correlationID).Msg("[TestEntityTracker] Deleting correlation")
		if _, err := db.ExecContext(ctx, "DELETE FROM compliance_correlations WHERE correlation_id = $1", correlationID); err != nil {
			return fmt.Errorf("failed to cleanup correlation %d: %w", correlationID, err)
		}
	}

	// 9. API Keys
	for apiKeyID := range t.apiKeys {
		log.Info().Str("api_key_id", apiKeyID).Msg("[TestEntityTracker] Deleting API key")
		if _, err := db.ExecContext(ctx, "DELETE FROM api_keys WHERE api_key_id = $1", apiKeyID); err != nil {
			return fmt.Errorf("failed to cleanup API key %s: %w", apiKeyID, err)
		}
	}

	// 10. Role Keys (depend on users)
	for keyID := range t.roleKeys {
		log.Info().Str("key_id", keyID).Msg("[TestEntityTracker] Deleting role key")
		if _, err := db.ExecContext(ctx, "DELETE FROM role_keys WHERE key_id = $1", keyID); err != nil {
			return fmt.Errorf("failed to cleanup role key %s: %w", keyID, err)
		}
	}

	// 11. Pseudonyms (depend on users)
	for pseudonymID := range t.pseudonyms {
		log.Info().Str("pseudonym_id", pseudonymID).Msg("[TestEntityTracker] Deleting pseudonym")
		if _, err := db.ExecContext(ctx, "DELETE FROM pseudonyms WHERE pseudonym_id = $1", pseudonymID); err != nil {
			return fmt.Errorf("failed to cleanup pseudonym %s: %w", pseudonymID, err)
		}
	}

	// 12. Subforums (depend on users)
	for subforumID := range t.subforums {
		log.Info().Int64("subforum_id", subforumID).Msg("[TestEntityTracker] Deleting subforum")
		if _, err := db.ExecContext(ctx, "DELETE FROM subforums WHERE subforum_id = $1", subforumID); err != nil {
			return fmt.Errorf("failed to cleanup subforum %d: %w", subforumID, err)
		}
	}

	// 13. Users (last, as everything depends on them)
	for userID := range t.users {
		log.Info().Int64("user_id", userID).Msg("[TestEntityTracker] Deleting user")
		if _, err := db.ExecContext(ctx, "DELETE FROM users WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("failed to cleanup user %d: %w", userID, err)
		}
	}

	log.Info().Msg("[TestEntityTracker] Cleanup completed successfully")
	return nil
}

// IntegrationTestSuite provides a complete test environment
type IntegrationTestSuite struct {
	DB                 bob.DB
	Server             *api.Server
	Config             *config.Config
	UserDAO            *dao.UserDAO
	SecurePseudonymDAO *dao.SecurePseudonymDAO
	RoleKeyDAO         *dao.RoleKeyDAO
	SubforumDAO        *dao.SubforumDAO
	PostDAO            *dao.PostDAO
	CommentDAO         *dao.CommentDAO
	VoteDAO            *dao.VoteDAO
	APIKeyDAO          *dao.APIKeyDAO
	UserBlockDAO       *dao.UserBlocksDAO
	UserPrefDAO        *dao.UserPreferencesDAO
	IdentityMappingDAO *dao.IdentityMappingDAO
	Tracker            *TestEntityTracker
	Cleanup            func()
	IBESystem          *ibe.IBESystem
}

// GetIBESystem returns the IBE system instance for this test suite
func (ts *IntegrationTestSuite) GetIBESystem() *ibe.IBESystem {
	return ts.IBESystem
}

// EnsureDefaultKeys ensures default role keys exist and tracks them for cleanup
func (ts *IntegrationTestSuite) EnsureDefaultKeys(t *testing.T, createdBy int64) {
	ctx := context.Background()

	// Get the user to determine their roles
	user, err := ts.UserDAO.GetUserByID(ctx, createdBy)
	if err != nil {
		t.Fatalf("Failed to get user for role key creation: %v", err)
	}

	// Parse user roles
	var userRoles []string
	if user.Roles.Valid {
		rolesBytes, err := user.Roles.V.Value()
		if err != nil {
			t.Fatalf("Failed to get user roles value: %v", err)
		}
		if err := json.Unmarshal(rolesBytes.([]byte), &userRoles); err != nil {
			t.Fatalf("Failed to parse user roles: %v", err)
		}
	}

	// If no roles, default to "user"
	if len(userRoles) == 0 {
		userRoles = []string{"user"}
	}

	// Ensure default keys exist for the user's actual roles
	if err := ts.RoleKeyDAO.EnsureDefaultKeys(ctx, ts.IBESystem, createdBy); err != nil {
		t.Fatalf("Failed to ensure default keys: %v", err)
	}

	// DEBUG: Print all role keys for this user (createdBy)
	roleKeys, err := dbmodels.RoleKeys.Query(dbmodels.SelectWhere.RoleKeys.CreatedBy.EQ(createdBy)).All(ctx, ts.DB)
	if err != nil {
		t.Fatalf("[DEBUG] Failed to fetch role keys for user %d: %v", createdBy, err)
	}
	fmt.Printf("[DEBUG] Role keys for user %d: %v\n", createdBy, roleKeys)

	// Track role keys for cleanup
	for _, roleKey := range roleKeys {
		ts.Tracker.TrackRoleKey(roleKey.KeyID.String())
	}

	// Role key tracking removed: Role keys are global entities that persist across tests
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

// forceDropTestDatabase forcibly drops and recreates the test database
func forceDropTestDatabase(t *testing.T, dbConfig *config.DatabaseConfig) {
	// Connect to postgres database to manage the test database
	postgresConfig := *dbConfig
	postgresConfig.Database = "postgres"

	postgresDB, err := database.NewConnection(&postgresConfig)
	if err != nil {
		t.Fatalf("Failed to connect to postgres database: %v", err)
	}
	defer postgresDB.Close()

	ctx := context.Background()

	// Terminate all connections to the test database
	_, err = postgresDB.ExecContext(ctx, `
		SELECT pg_terminate_backend(pid) 
		FROM pg_stat_activity 
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`, dbConfig.Database)
	if err != nil {
		t.Logf("Warning: failed to terminate connections to test database: %v", err)
	}

	// Drop the test database
	_, err = postgresDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+dbConfig.Database)
	if err != nil {
		t.Logf("Warning: failed to drop test database: %v", err)
	}

	// Recreate the test database
	_, err = postgresDB.ExecContext(ctx, "CREATE DATABASE "+dbConfig.Database)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Logf("Test database %s dropped and recreated", dbConfig.Database)
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

	// For test databases, check if we need to drop and recreate
	// Only do this if the database doesn't exist or has no tables
	if strings.Contains(cfg.Database.Database, "test") {
		// Check if database exists and has tables
		postgresConfig := cfg.Database
		postgresConfig.Database = "postgres"

		postgresDB, err := database.NewConnection(&postgresConfig)
		if err != nil {
			t.Fatalf("Failed to connect to postgres database: %v", err)
		}
		defer postgresDB.Close()

		ctx := context.Background()

		// Check if test database exists
		var exists bool
		err = postgresDB.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
			cfg.Database.Database).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check if database exists: %v", err)
		}

		if !exists {
			// Database doesn't exist, create it
			_, err = postgresDB.ExecContext(ctx, "CREATE DATABASE "+cfg.Database.Database)
			if err != nil {
				t.Fatalf("Failed to create test database: %v", err)
			}
			t.Logf("Created test database %s", cfg.Database.Database)
		} else {
			// Database exists, check if it has tables
			testDB, err := database.NewConnection(&cfg.Database)
			if err != nil {
				t.Fatalf("Failed to connect to test database: %v", err)
			}
			defer testDB.Close()

			var tableCount int
			err = testDB.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
			if err != nil {
				t.Fatalf("Failed to check table count: %v", err)
			}

			if tableCount == 0 {
				// Database exists but has no tables, drop and recreate
				t.Logf("Test database %s exists but has no tables, dropping and recreating", cfg.Database.Database)
				forceDropTestDatabase(t, &cfg.Database)
			} else {
				t.Logf("Test database %s exists with %d tables, using existing database", cfg.Database.Database, tableCount)
			}
		}
	}

	// Create database connection
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Get the raw *sql.DB from bob.DB
	rawDB := db.DB

	// Create IBE system for correlation operations with deterministic test key
	testMasterSecret := []byte("test_master_secret_32_bytes_long_key")
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
		MasterSecret: testMasterSecret,
		KeyVersion:   1,
		Salt:         "test_fingerprint_salt_v1",
	})

	// Create DAOs
	userDAO := dao.NewUserDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)
	securePseudonymDAO := dao.NewSecurePseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO)
	postDAO := dao.NewPostDAO(db)
	commentDAO := dao.NewCommentDAO(db)
	userPreferencesDAO := dao.NewUserPreferencesDAO(db)
	userBlocksDAO := dao.NewUserBlocksDAO(db)
	subforumDAO := dao.NewSubforumDAO(db)
	voteDAO := dao.NewVoteDAO(db)
	apiKeyDAO := dao.NewAPIKeyDAO(db)

	// Create auth middleware with test configuration
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret, apiKeyDAO, &cfg.JWT, &cfg.Security)

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
	humaAPI.UseMiddleware(middleware.CORSMiddleware(&cfg.CORS))

	// Add authentication middleware to extract user context
	humaAPI.UseMiddleware(middleware.AuthenticateUserHuma)
	log.Info().Str("jwt_secret_length", fmt.Sprintf("%d", len(cfg.JWT.Secret))).Msg("JWT configuration loaded")

	// Register routes with test configuration
	routes.RegisterHealthRoutes(humaAPI)
	routes.RegisterHelloRoutes(humaAPI)
	routes.RegisterAuthRoutes(humaAPI, cfg, db, rawDB, ibeSystem)
	routes.RegisterUserRoutes(humaAPI, userDAO, securePseudonymDAO, userPreferencesDAO, userBlocksDAO, postDAO, commentDAO, ibeSystem)
	routes.RegisterSubforumRoutes(humaAPI, db)
	routes.RegisterMessagesRoutes(humaAPI)
	routes.RegisterSearchRoutes(humaAPI)
	routes.RegisterModerationRoutes(humaAPI)
	routes.RegisterContentRoutes(humaAPI, db, rawDB, ibeSystem, identityMappingDAO, userDAO)
	routes.RegisterCorrelationRoutes(humaAPI, db, ibeSystem, securePseudonymDAO, identityMappingDAO, postDAO, commentDAO)

	server := &api.Server{
		API:       humaAPI,
		Mux:       mux,
		Config:    config,
		AppConfig: cfg,
	}

	// Create entity tracker
	tracker := NewTestEntityTracker()

	// Bootstrap role keys for all test scenarios
	// This ensures role keys exist before any tests run, preventing race conditions
	if err := bootstrapRoleKeys(context.Background(), db, roleKeyDAO, ibeSystem); err != nil {
		t.Fatalf("Failed to bootstrap role keys: %v", err)
	}

	// Create test suite with consistent IBE system
	suite := &IntegrationTestSuite{
		DB:                 db,
		Server:             server,
		Config:             cfg,
		UserDAO:            userDAO,
		SecurePseudonymDAO: securePseudonymDAO,
		RoleKeyDAO:         roleKeyDAO,
		SubforumDAO:        subforumDAO,
		PostDAO:            postDAO,
		CommentDAO:         commentDAO,
		VoteDAO:            voteDAO,
		APIKeyDAO:          apiKeyDAO,
		UserBlockDAO:       userBlocksDAO,
		UserPrefDAO:        userPreferencesDAO,
		IdentityMappingDAO: identityMappingDAO,
		Tracker:            tracker,
		IBESystem:          ibeSystem,
		Cleanup: func() {
			// Clean up test data
			ctx := context.Background()
			if err := tracker.Cleanup(ctx, db); err != nil {
				t.Logf("Warning: failed to cleanup test data: %v", err)
			}
		},
	}

	return suite
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

	// Ensure default role keys are created for the user
	if err := ts.RoleKeyDAO.EnsureDefaultKeys(ctx, ts.IBESystem, user.UserID); err != nil {
		t.Fatalf("Failed to ensure default keys for test user: %v", err)
	}

	// Track role keys for cleanup
	roleKeys, err := dbmodels.RoleKeys.Query(dbmodels.SelectWhere.RoleKeys.CreatedBy.EQ(user.UserID)).All(ctx, ts.DB)
	if err != nil {
		t.Fatalf("[DEBUG] Failed to fetch role keys for user %d: %v", user.UserID, err)
	}
	for _, roleKey := range roleKeys {
		ts.Tracker.TrackRoleKey(roleKey.KeyID.String())
	}
	fmt.Printf("[DEBUG] Role keys for user %d: %v\n", user.UserID, roleKeys)

	// Create pseudonym for the user
	displayName := fmt.Sprintf("test_user_%d", user.UserID)
	pseudonym, err := ts.SecurePseudonymDAO.CreatePseudonymWithIdentityMapping(ctx, user.UserID, displayName)
	if err != nil {
		t.Fatalf("Failed to create test user pseudonym: %v", err)
	}

	// Track pseudonym for cleanup
	ts.Tracker.TrackPseudonym(pseudonym.PseudonymID)

	testUser := &TestUser{
		UserID:       user.UserID,
		Email:        email,
		Password:     password,
		PasswordHash: passwordHash,
		Roles:        roles,
		Capabilities: getCapabilitiesForRoles(roles),
		PseudonymID:  pseudonym.PseudonymID,
		DisplayName:  displayName,
	}
	return testUser
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
	pseudonyms, err := ts.SecurePseudonymDAO.GetPseudonymsByUserID(ctx, userID, "user", "authentication")
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
	pseudonym, err := ts.SecurePseudonymDAO.CreatePseudonymWithIdentityMapping(ctx, userID, displayName)
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
	// Use the existing IBE system and DAOs from the test suite
	ibeSystem := ts.IBESystem
	userDAO := ts.UserDAO
	identityMappingDAO := ts.IdentityMappingDAO
	pseudonymDAO := ts.SecurePseudonymDAO
	postDAO := ts.PostDAO
	commentDAO := ts.CommentDAO
	userPreferencesDAO := ts.UserPrefDAO
	userBlocksDAO := ts.UserBlockDAO
	apiKeyDAO := ts.APIKeyDAO

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
	routes.RegisterAuthRoutes(humaAPI, ts.Config, ts.DB, ts.DB.DB, ibeSystem)
	routes.RegisterUserRoutes(humaAPI, userDAO, pseudonymDAO, userPreferencesDAO, userBlocksDAO, postDAO, commentDAO, ibeSystem)
	routes.RegisterSubforumRoutes(humaAPI, ts.DB)
	routes.RegisterMessagesRoutes(humaAPI)
	routes.RegisterSearchRoutes(humaAPI)
	routes.RegisterModerationRoutes(humaAPI)
	routes.RegisterContentRoutes(humaAPI, ts.DB, ts.DB.DB, ibeSystem, identityMappingDAO, userDAO)
	routes.RegisterCorrelationRoutes(humaAPI, ts.DB, ibeSystem, pseudonymDAO, identityMappingDAO, postDAO, commentDAO)

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
		case "user":
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "create_subforum")
		case "platform_admin":
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "create_subforum", "moderation", "compliance", "legal_requests")
		case "trust_safety":
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "moderation", "compliance")
		case "legal_team":
			capabilities = append(capabilities, "create_content", "vote", "message", "report", "compliance", "legal_requests")
		}
	}

	return capabilities
}

// bootstrapRoleKeys creates all necessary role keys for testing scenarios
// This ensures role keys exist before any tests run, preventing race conditions
func bootstrapRoleKeys(ctx context.Context, db bob.DB, roleKeyDAO *dao.RoleKeyDAO, ibeSystem *ibe.IBESystem) error {
	// Create a bootstrap user for creating role keys
	userDAO := dao.NewUserDAO(db)
	bootstrapEmail := "bootstrap@test.local"

	// Check if bootstrap user already exists
	bootstrapUser, err := userDAO.GetUserByEmail(ctx, bootstrapEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check if bootstrap user exists: %w", err)
	}

	if bootstrapUser == nil {
		// Create bootstrap user if it doesn't exist
		bootstrapPasswordHash := hashPassword("bootstrap_password")
		bootstrapUser, err = userDAO.CreateUser(ctx, bootstrapEmail, bootstrapPasswordHash)
		if err != nil {
			return fmt.Errorf("failed to create bootstrap user: %w", err)
		}
	}

	// Define all roles that might be used in tests
	allRoles := []string{"user", "platform_admin", "trust_safety", "legal_team", "moderator", "admin"}

	// Define default keys for each role
	defaultKeys := []struct {
		roleName     string
		scope        string
		capabilities []string
	}{}

	// Add authentication and self_correlation keys for each role
	for _, roleName := range allRoles {
		defaultKeys = append(defaultKeys, struct {
			roleName     string
			scope        string
			capabilities []string
		}{
			roleName: roleName,
			scope:    "authentication",
			capabilities: []string{
				"access_own_pseudonyms",
				"login",
				"session_management",
			},
		})
		defaultKeys = append(defaultKeys, struct {
			roleName     string
			scope        string
			capabilities []string
		}{
			roleName: roleName,
			scope:    "self_correlation",
			capabilities: []string{
				"verify_own_pseudonym_ownership",
				"manage_own_profile",
			},
		})
	}

	// Add correlation keys for admin roles
	adminRoles := []string{"platform_admin", "trust_safety", "legal_team", "moderator", "admin"}
	for _, roleName := range adminRoles {
		defaultKeys = append(defaultKeys, struct {
			roleName     string
			scope        string
			capabilities []string
		}{
			roleName: roleName,
			scope:    "correlation",
			capabilities: []string{
				"access_all_pseudonyms",
				"cross_user_correlation",
				"moderation",
				"compliance",
				"legal_requests",
			},
		})
	}

	// Create each role key if it doesn't exist
	for _, keyDef := range defaultKeys {
		existingKey, err := roleKeyDAO.GetRoleKey(ctx, keyDef.roleName, keyDef.scope)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check if role key exists for role=%s scope=%s: %w", keyDef.roleName, keyDef.scope, err)
		}

		if existingKey == nil {
			// Key doesn't exist, create it
			expiresAt := time.Now().AddDate(1, 0, 0) // Expire in 1 year
			keyData := ibeSystem.GenerateTestRoleKey(keyDef.roleName, keyDef.scope)

			_, err = roleKeyDAO.CreateRoleKey(ctx, keyDef.roleName, keyDef.scope, keyData, keyDef.capabilities, expiresAt, bootstrapUser.UserID)
			if err != nil {
				return fmt.Errorf("failed to create bootstrap role key for role=%s scope=%s: %w", keyDef.roleName, keyDef.scope, err)
			}
		}
	}

	return nil
}
