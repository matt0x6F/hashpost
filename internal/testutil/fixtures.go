package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// TestUser represents a test user fixture
type TestUser struct {
	ID           uuid.UUID `json:"id"`
	DID          string    `json:"did"`
	Handle       string    `json:"handle"`
	Email        *string   `json:"email,omitempty"`
	PasswordHash *string   `json:"password_hash,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TestSubforum represents a test subforum fixture
type TestSubforum struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TestPost represents a test post fixture
type TestPost struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	SubforumID uuid.UUID `json:"subforum_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	AtprotoURI *string   `json:"atproto_uri,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TestComment represents a test comment fixture
type TestComment struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PostID    uuid.UUID `json:"post_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TestRole represents a test role fixture
type TestRole struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TestUserRole represents a test user role assignment fixture
type TestUserRole struct {
	ID         uuid.UUID  `json:"id"`
	UserDID    string     `json:"user_did"`
	RoleName   string     `json:"role_name"`
	SubforumID *uuid.UUID `json:"subforum_id,omitempty"`
	GrantedBy  string     `json:"granted_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TestSession represents a test session fixture
type TestSession struct {
	ID        string    `json:"id"`
	UserDID   string    `json:"user_did"`
	Handle    string    `json:"handle"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TestFixtures holds all test fixture data
type TestFixtures struct {
	Users     []TestUser     `json:"users"`
	Subforums []TestSubforum `json:"subforums"`
	Posts     []TestPost     `json:"posts"`
	Comments  []TestComment  `json:"comments"`
	Roles     []TestRole     `json:"roles"`
	UserRoles []TestUserRole `json:"user_roles"`
	Sessions  []TestSession  `json:"sessions"`
}

// CreateDefaultFixtures creates default test fixtures
func CreateDefaultFixtures() *TestFixtures {
	now := time.Now()

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	adminID := uuid.New()

	users := []TestUser{
		{
			ID:           user1ID,
			DID:          "did:plc:test-user-1",
			Handle:       "testuser1.hashpost.local",
			Email:        stringPtr("testuser1@example.com"),
			PasswordHash: stringPtr("$2a$10$dummy.hash.for.testing"),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           user2ID,
			DID:          "did:plc:test-user-2",
			Handle:       "testuser2.hashpost.local",
			Email:        stringPtr("testuser2@example.com"),
			PasswordHash: stringPtr("$2a$10$dummy.hash.for.testing"),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           adminID,
			DID:          "did:plc:test-admin",
			Handle:       "admin.hashpost.local",
			Email:        stringPtr("admin@example.com"),
			PasswordHash: stringPtr("$2a$10$dummy.hash.for.testing"),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	// Create test subforums
	subforum1ID := uuid.New()
	subforum2ID := uuid.New()

	subforums := []TestSubforum{
		{
			ID:          subforum1ID,
			Name:        "General Discussion",
			Slug:        "general",
			Description: stringPtr("General discussion about HashPost"),
			CreatedBy:   user1ID,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          subforum2ID,
			Name:        "Technical Support",
			Slug:        "tech-support",
			Description: stringPtr("Technical support and help"),
			CreatedBy:   adminID,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Create test posts
	posts := []TestPost{
		{
			ID:         uuid.New(),
			UserID:     user1ID,
			SubforumID: subforum1ID,
			Title:      "Welcome to HashPost!",
			Content:    "This is a test post to welcome users to HashPost.",
			AtprotoURI: stringPtr("at://did:plc:test-user-1/com.hashpost.feed.post/123"),
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			UserID:     user2ID,
			SubforumID: subforum1ID,
			Title:      "Test Post 2",
			Content:    "This is another test post.",
			AtprotoURI: stringPtr("at://did:plc:test-user-2/com.hashpost.feed.post/456"),
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	// Create test comments
	comments := []TestComment{
		{
			ID:        uuid.New(),
			UserID:    user2ID,
			PostID:    posts[0].ID,
			Content:   "Great post! Thanks for sharing.",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Create test roles
	roles := []TestRole{
		{
			ID:          uuid.New(),
			Name:        "user",
			Description: stringPtr("Regular user role"),
			Permissions: []string{"post.create", "post.read", "comment.create", "comment.read"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Name:        "moderator",
			Description: stringPtr("Moderator role"),
			Permissions: []string{"post.create", "post.read", "post.moderate", "comment.create", "comment.read", "comment.moderate"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			Name:        "platform_admin",
			Description: stringPtr("Platform administrator role"),
			Permissions: []string{"platform.manage_users", "platform.manage_roles", "platform.manage_settings"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Create test user roles
	userRoles := []TestUserRole{
		{
			ID:        uuid.New(),
			UserDID:   "did:plc:test-user-1",
			RoleName:  "user",
			GrantedBy: "did:plc:test-admin",
			CreatedAt: now,
		},
		{
			ID:        uuid.New(),
			UserDID:   "did:plc:test-user-2",
			RoleName:  "user",
			GrantedBy: "did:plc:test-admin",
			CreatedAt: now,
		},
		{
			ID:        uuid.New(),
			UserDID:   "did:plc:test-admin",
			RoleName:  "platform_admin",
			GrantedBy: "did:plc:test-admin",
			CreatedAt: now,
		},
	}

	// Create test sessions
	sessions := []TestSession{
		{
			ID:        "test-session-1",
			UserDID:   "did:plc:test-user-1",
			Handle:    "testuser1.hashpost.local",
			ExpiresAt: now.Add(24 * time.Hour),
			CreatedAt: now,
		},
		{
			ID:        "test-session-2",
			UserDID:   "did:plc:test-user-2",
			Handle:    "testuser2.hashpost.local",
			ExpiresAt: now.Add(24 * time.Hour),
			CreatedAt: now,
		},
	}

	return &TestFixtures{
		Users:     users,
		Subforums: subforums,
		Posts:     posts,
		Comments:  comments,
		Roles:     roles,
		UserRoles: userRoles,
		Sessions:  sessions,
	}
}

// SaveFixturesToFile saves fixtures to a YAML file
func SaveFixturesToFile(fixtures *TestFixtures, filename string) error {
	// Ensure testdata directory exists
	if err := os.MkdirAll("testdata/fixtures", 0755); err != nil {
		return fmt.Errorf("failed to create testdata directory: %w", err)
	}

	filepath := filepath.Join("testdata/fixtures", filename)
	data, err := json.MarshalIndent(fixtures, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal fixtures: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write fixtures file: %w", err)
	}

	return nil
}

// LoadFixturesFromFile loads fixtures from a JSON file
func LoadFixturesFromFile(filename string) (*TestFixtures, error) {
	filepath := filepath.Join("testdata/fixtures", filename)
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixtures file: %w", err)
	}

	var fixtures TestFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fixtures: %w", err)
	}

	return &fixtures, nil
}

// CreateTestUser creates a test user with default values
func CreateTestUser(did, handle string) TestUser {
	now := time.Now()
	return TestUser{
		ID:           uuid.New(),
		DID:          did,
		Handle:       handle,
		Email:        stringPtr(fmt.Sprintf("%s@example.com", handle)),
		PasswordHash: stringPtr("$2a$10$dummy.hash.for.testing"),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// CreateTestSubforum creates a test subforum with default values
func CreateTestSubforum(name, slug string, createdBy uuid.UUID) TestSubforum {
	now := time.Now()
	return TestSubforum{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: stringPtr(fmt.Sprintf("Description for %s", name)),
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateTestPost creates a test post with default values
func CreateTestPost(userID, subforumID uuid.UUID, title, content string) TestPost {
	now := time.Now()
	return TestPost{
		ID:         uuid.New(),
		UserID:     userID,
		SubforumID: subforumID,
		Title:      title,
		Content:    content,
		AtprotoURI: stringPtr(fmt.Sprintf("at://did:plc:test-user/com.hashpost.feed.post/%s", uuid.New().String())),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CreateTestSession creates a test session with default values
func CreateTestSession(userDID, handle string) TestSession {
	now := time.Now()
	return TestSession{
		ID:        uuid.New().String(),
		UserDID:   userDID,
		Handle:    handle,
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
}
