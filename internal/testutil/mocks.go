package testutil

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	appview "github.com/matt0x6f/hashpost/internal/database/generated/appview"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
)

// MockQueries is a mock implementation of the generated.Queries interface
type MockQueries struct {
	// User-related mocks
	CreateUserWithPasswordFunc func(ctx context.Context, arg *generated.CreateUserWithPasswordParams) (*generated.User, error)
	GetUserByDIDFunc           func(ctx context.Context, did string) (*generated.User, error)
	GetUserByHandleFunc        func(ctx context.Context, handle string) (*generated.User, error)
	GetUserByIDFunc            func(ctx context.Context, id uuid.UUID) (*generated.User, error)
	UpdateUserPasswordHashFunc func(ctx context.Context, arg *generated.UpdateUserPasswordHashParams) error
	DeleteUserFunc             func(ctx context.Context, id uuid.UUID) error

	// Session-related mocks
	CreateUserSessionFunc             func(ctx context.Context, arg *generated.CreateUserSessionParams) (*generated.UserSession, error)
	GetUserSessionFunc                func(ctx context.Context, sessionID string) (*generated.UserSession, error)
	UpdateUserSessionLastAccessedFunc func(ctx context.Context, sessionID string) error
	DeleteUserSessionFunc             func(ctx context.Context, sessionID string) error

	// OAuth-related mocks
	GetOAuthClientFunc            func(ctx context.Context, clientID string) (*generated.OauthClient, error)
	CreateAuthorizationCodeFunc   func(ctx context.Context, arg *generated.CreateAuthorizationCodeParams) (*generated.OauthAuthorizationCode, error)
	GetAuthorizationCodeFunc      func(ctx context.Context, code string) (*generated.OauthAuthorizationCode, error)
	MarkAuthorizationCodeUsedFunc func(ctx context.Context, code string) error

	// DPoP-related mocks
	CreateDPoPNonceFunc          func(ctx context.Context, arg *generated.CreateDPoPNonceParams) (*generated.DpopNonce, error)
	GetDPoPNonceFunc             func(ctx context.Context, nonce string) (*generated.DpopNonce, error)
	MarkDPoPNonceUsedFunc        func(ctx context.Context, nonce string) error
	CleanupExpiredDPoPNoncesFunc func(ctx context.Context) error

	// Post-related mocks
	CreatePostFunc             func(ctx context.Context, arg *generated.CreatePostParams) (*generated.Post, error)
	GetPostByIDFunc            func(ctx context.Context, id uuid.UUID) (*generated.Post, error)
	GetPostByAtprotoURIFunc    func(ctx context.Context, atprotoUri *string) (*generated.Post, error)
	UpdatePostByAtprotoURIFunc func(ctx context.Context, arg *generated.UpdatePostByAtprotoURIParams) (*generated.Post, error)
	DeletePostByAtprotoURIFunc func(ctx context.Context, atprotoUri *string) error
	ListPostsFunc              func(ctx context.Context, arg *generated.ListPostsParams) ([]generated.ListPostsRow, error)
	ListPostsWithCursorFunc    func(ctx context.Context, arg *generated.ListPostsWithCursorParams) ([]*generated.Post, error)

	// Subforum-related mocks
	CreateSubforumFunc          func(ctx context.Context, arg *generated.CreateSubforumParams) (*generated.Subforum, error)
	GetSubforumByIDFunc         func(ctx context.Context, id uuid.UUID) (*generated.Subforum, error)
	GetSubforumBySlugFunc       func(ctx context.Context, slug string) (*generated.Subforum, error)
	UpdateSubforumByIDFunc      func(ctx context.Context, arg *generated.UpdateSubforumByIDParams) (*generated.Subforum, error)
	DeleteSubforumByIDFunc      func(ctx context.Context, id uuid.UUID) error
	ListSubforumsFunc           func(ctx context.Context, arg *generated.ListSubforumsParams) ([]generated.ListSubforumsRow, error)
	ListSubforumsWithCursorFunc func(ctx context.Context, arg *generated.ListSubforumsWithCursorParams) ([]*generated.Subforum, error)

	// Processed events mocks
	IsEventProcessedFunc     func(ctx context.Context, eventID string) (bool, error)
	CreateProcessedEventFunc func(ctx context.Context, arg *appview.CreateProcessedEventParams) (*appview.ProcessedEvent, error)
}

// Implement the generated.Queries interface methods
func (m *MockQueries) CreateUserWithPassword(ctx context.Context, arg *generated.CreateUserWithPasswordParams) (*generated.User, error) {
	if m.CreateUserWithPasswordFunc != nil {
		return m.CreateUserWithPasswordFunc(ctx, arg)
	}
	return &generated.User{
		ID:           uuid.New(),
		Handle:       arg.Handle,
		Did:          arg.Did,
		Email:        arg.Email,
		PasswordHash: arg.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetUserByDID(ctx context.Context, did string) (*generated.User, error) {
	if m.GetUserByDIDFunc != nil {
		return m.GetUserByDIDFunc(ctx, did)
	}
	return &generated.User{
		ID:        uuid.New(),
		Handle:    "testuser.hashpost.local",
		Did:       did,
		Email:     stringPtr("test@example.com"),
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetUserByHandle(ctx context.Context, handle string) (*generated.User, error) {
	if m.GetUserByHandleFunc != nil {
		return m.GetUserByHandleFunc(ctx, handle)
	}
	return &generated.User{
		ID:        uuid.New(),
		Handle:    handle,
		Did:       "did:plc:test-user",
		Email:     stringPtr("test@example.com"),
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetUserByID(ctx context.Context, id uuid.UUID) (*generated.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return &generated.User{
		ID:        id,
		Handle:    "testuser.hashpost.local",
		Did:       "did:plc:test-user",
		Email:     stringPtr("test@example.com"),
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) UpdateUserPasswordHash(ctx context.Context, arg *generated.UpdateUserPasswordHashParams) error {
	if m.UpdateUserPasswordHashFunc != nil {
		return m.UpdateUserPasswordHashFunc(ctx, arg)
	}
	return nil
}

func (m *MockQueries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func (m *MockQueries) CreateUserSession(ctx context.Context, arg *generated.CreateUserSessionParams) (*generated.UserSession, error) {
	if m.CreateUserSessionFunc != nil {
		return m.CreateUserSessionFunc(ctx, arg)
	}
	return &generated.UserSession{
		SessionID: arg.SessionID,
		UserDid:   arg.UserDid,
		Handle:    arg.Handle,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetUserSession(ctx context.Context, sessionID string) (*generated.UserSession, error) {
	if m.GetUserSessionFunc != nil {
		return m.GetUserSessionFunc(ctx, sessionID)
	}
	return &generated.UserSession{
		SessionID: sessionID,
		UserDid:   "did:plc:test-user",
		Handle:    "testuser.hashpost.local",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) UpdateUserSessionLastAccessed(ctx context.Context, sessionID string) error {
	if m.UpdateUserSessionLastAccessedFunc != nil {
		return m.UpdateUserSessionLastAccessedFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockQueries) DeleteUserSession(ctx context.Context, sessionID string) error {
	if m.DeleteUserSessionFunc != nil {
		return m.DeleteUserSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockQueries) GetOAuthClient(ctx context.Context, clientID string) (*generated.OauthClient, error) {
	if m.GetOAuthClientFunc != nil {
		return m.GetOAuthClientFunc(ctx, clientID)
	}
	return &generated.OauthClient{
		ClientID:      clientID,
		ClientName:    "Test Client",
		RedirectUris:  []string{"http://localhost:3000/callback"},
		Scopes:        []string{"read", "write"},
		GrantTypes:    []string{"authorization_code"},
		ResponseTypes: []string{"code"},
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) CreateAuthorizationCode(ctx context.Context, arg *generated.CreateAuthorizationCodeParams) (*generated.OauthAuthorizationCode, error) {
	if m.CreateAuthorizationCodeFunc != nil {
		return m.CreateAuthorizationCodeFunc(ctx, arg)
	}
	return &generated.OauthAuthorizationCode{
		Code:        arg.Code,
		ClientID:    arg.ClientID,
		UserDid:     arg.UserDid,
		RedirectUri: arg.RedirectUri,
		Scope:       arg.Scope,
		Nonce:       arg.Nonce,
		ExpiresAt:   arg.ExpiresAt,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetAuthorizationCode(ctx context.Context, code string) (*generated.OauthAuthorizationCode, error) {
	if m.GetAuthorizationCodeFunc != nil {
		return m.GetAuthorizationCodeFunc(ctx, code)
	}
	return &generated.OauthAuthorizationCode{
		Code:        code,
		ClientID:    "test-client",
		UserDid:     "did:plc:test-user",
		RedirectUri: "http://localhost:3000/callback",
		Scope:       "read write",
		Nonce:       stringPtr("test-nonce"),
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) MarkAuthorizationCodeUsed(ctx context.Context, code string) error {
	if m.MarkAuthorizationCodeUsedFunc != nil {
		return m.MarkAuthorizationCodeUsedFunc(ctx, code)
	}
	return nil
}

func (m *MockQueries) CreateDPoPNonce(ctx context.Context, arg *generated.CreateDPoPNonceParams) (*generated.DpopNonce, error) {
	if m.CreateDPoPNonceFunc != nil {
		return m.CreateDPoPNonceFunc(ctx, arg)
	}
	return &generated.DpopNonce{
		Nonce:     arg.Nonce,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetDPoPNonce(ctx context.Context, nonce string) (*generated.DpopNonce, error) {
	if m.GetDPoPNonceFunc != nil {
		return m.GetDPoPNonceFunc(ctx, nonce)
	}
	return &generated.DpopNonce{
		Nonce:     nonce,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) MarkDPoPNonceUsed(ctx context.Context, nonce string) error {
	if m.MarkDPoPNonceUsedFunc != nil {
		return m.MarkDPoPNonceUsedFunc(ctx, nonce)
	}
	return nil
}

func (m *MockQueries) CleanupExpiredDPoPNonces(ctx context.Context) error {
	if m.CleanupExpiredDPoPNoncesFunc != nil {
		return m.CleanupExpiredDPoPNoncesFunc(ctx)
	}
	return nil
}

func (m *MockQueries) CreatePost(ctx context.Context, arg *generated.CreatePostParams) (*generated.Post, error) {
	if m.CreatePostFunc != nil {
		return m.CreatePostFunc(ctx, arg)
	}
	return &generated.Post{
		ID:         uuid.New(),
		UserID:     arg.UserID,
		SubforumID: arg.SubforumID,
		Title:      arg.Title,
		Content:    arg.Content,
		AtprotoUri: arg.AtprotoUri,
		CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetPostByID(ctx context.Context, id uuid.UUID) (*generated.Post, error) {
	if m.GetPostByIDFunc != nil {
		return m.GetPostByIDFunc(ctx, id)
	}
	return &generated.Post{
		ID:         id,
		UserID:     pgtype.UUID{Bytes: uuid.New(), Valid: true},
		SubforumID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Title:      "Test Post",
		Content:    "Test content",
		AtprotoUri: stringPtr("at://did:plc:test-user/com.hashpost.feed.post/123"),
		CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetPostByAtprotoURI(ctx context.Context, atprotoUri *string) (*generated.Post, error) {
	if m.GetPostByAtprotoURIFunc != nil {
		return m.GetPostByAtprotoURIFunc(ctx, atprotoUri)
	}
	return &generated.Post{
		ID:         uuid.New(),
		UserID:     pgtype.UUID{Bytes: uuid.New(), Valid: true},
		SubforumID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Title:      "Test Post",
		Content:    "Test content",
		AtprotoUri: atprotoUri,
		CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) UpdatePostByAtprotoURI(ctx context.Context, arg *generated.UpdatePostByAtprotoURIParams) (*generated.Post, error) {
	if m.UpdatePostByAtprotoURIFunc != nil {
		return m.UpdatePostByAtprotoURIFunc(ctx, arg)
	}
	return &generated.Post{
		ID:         uuid.New(),
		UserID:     pgtype.UUID{Bytes: uuid.New(), Valid: true},
		SubforumID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Title:      arg.Title,
		Content:    arg.Content,
		AtprotoUri: arg.AtprotoUri,
		CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) DeletePostByAtprotoURI(ctx context.Context, atprotoUri *string) error {
	if m.DeletePostByAtprotoURIFunc != nil {
		return m.DeletePostByAtprotoURIFunc(ctx, atprotoUri)
	}
	return nil
}

func (m *MockQueries) ListPosts(ctx context.Context, arg *generated.ListPostsParams) ([]generated.ListPostsRow, error) {
	if m.ListPostsFunc != nil {
		return m.ListPostsFunc(ctx, arg)
	}
	return []generated.ListPostsRow{}, nil
}

func (m *MockQueries) ListPostsWithCursor(ctx context.Context, arg *generated.ListPostsWithCursorParams) ([]*generated.Post, error) {
	if m.ListPostsWithCursorFunc != nil {
		return m.ListPostsWithCursorFunc(ctx, arg)
	}
	return []*generated.Post{}, nil
}

func (m *MockQueries) CreateSubforum(ctx context.Context, arg *generated.CreateSubforumParams) (*generated.Subforum, error) {
	if m.CreateSubforumFunc != nil {
		return m.CreateSubforumFunc(ctx, arg)
	}
	return &generated.Subforum{
		ID:          uuid.New(),
		Name:        arg.Name,
		Slug:        arg.Slug,
		Description: arg.Description,
		CreatedBy:   arg.CreatedBy,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetSubforumByID(ctx context.Context, id uuid.UUID) (*generated.Subforum, error) {
	if m.GetSubforumByIDFunc != nil {
		return m.GetSubforumByIDFunc(ctx, id)
	}
	return &generated.Subforum{
		ID:          id,
		Name:        "Test Subforum",
		Slug:        "test",
		Description: stringPtr("Test description"),
		CreatedBy:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) GetSubforumBySlug(ctx context.Context, slug string) (*generated.Subforum, error) {
	if m.GetSubforumBySlugFunc != nil {
		return m.GetSubforumBySlugFunc(ctx, slug)
	}
	return &generated.Subforum{
		ID:          uuid.New(),
		Name:        "Test Subforum",
		Slug:        slug,
		Description: stringPtr("Test description"),
		CreatedBy:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) UpdateSubforumByID(ctx context.Context, arg *generated.UpdateSubforumByIDParams) (*generated.Subforum, error) {
	if m.UpdateSubforumByIDFunc != nil {
		return m.UpdateSubforumByIDFunc(ctx, arg)
	}
	return &generated.Subforum{
		ID:          arg.ID,
		Name:        arg.Name,
		Slug:        arg.Slug,
		Description: arg.Description,
		CreatedBy:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *MockQueries) DeleteSubforumByID(ctx context.Context, id uuid.UUID) error {
	if m.DeleteSubforumByIDFunc != nil {
		return m.DeleteSubforumByIDFunc(ctx, id)
	}
	return nil
}

func (m *MockQueries) ListSubforums(ctx context.Context, arg *generated.ListSubforumsParams) ([]generated.ListSubforumsRow, error) {
	if m.ListSubforumsFunc != nil {
		return m.ListSubforumsFunc(ctx, arg)
	}
	return []generated.ListSubforumsRow{}, nil
}

func (m *MockQueries) ListSubforumsWithCursor(ctx context.Context, arg *generated.ListSubforumsWithCursorParams) ([]*generated.Subforum, error) {
	if m.ListSubforumsWithCursorFunc != nil {
		return m.ListSubforumsWithCursorFunc(ctx, arg)
	}
	return []*generated.Subforum{}, nil
}

// Processed events methods
func (m *MockQueries) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	if m.IsEventProcessedFunc != nil {
		return m.IsEventProcessedFunc(ctx, eventID)
	}
	return false, nil
}

func (m *MockQueries) CreateProcessedEvent(ctx context.Context, arg *appview.CreateProcessedEventParams) (*appview.ProcessedEvent, error) {
	if m.CreateProcessedEventFunc != nil {
		return m.CreateProcessedEventFunc(ctx, arg)
	}
	now := time.Now()
	return &appview.ProcessedEvent{
		EventID:     arg.EventID,
		Subject:     arg.Subject,
		Sequence:    arg.Sequence,
		ProcessedAt: pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}, nil
}

// MockAppViewQueries is a mock implementation of the AppView generated.Queries interface
type MockAppViewQueries struct {
	// Processed events mocks
	IsEventProcessedFunc     func(ctx context.Context, eventID string) (bool, error)
	CreateProcessedEventFunc func(ctx context.Context, arg *appview.CreateProcessedEventParams) (*appview.ProcessedEvent, error)

	// Internal state to track processed events
	processedEvents map[string]bool
}

// NewMockAppViewQueries creates a new mock AppView queries instance
func NewMockAppViewQueries() *MockAppViewQueries {
	return &MockAppViewQueries{
		processedEvents: make(map[string]bool),
	}
}

// Implement the AppView generated.Querier interface methods
func (m *MockAppViewQueries) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	if m.IsEventProcessedFunc != nil {
		return m.IsEventProcessedFunc(ctx, eventID)
	}
	// Check internal state
	return m.processedEvents[eventID], nil
}

func (m *MockAppViewQueries) CreateProcessedEvent(ctx context.Context, arg *appview.CreateProcessedEventParams) (*appview.ProcessedEvent, error) {
	if m.CreateProcessedEventFunc != nil {
		return m.CreateProcessedEventFunc(ctx, arg)
	}
	// Update internal state
	m.processedEvents[arg.EventID] = true
	now := time.Now()
	return &appview.ProcessedEvent{
		EventID:     arg.EventID,
		Subject:     arg.Subject,
		Sequence:    arg.Sequence,
		ProcessedAt: pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}, nil
}

// Stub implementations for all other Querier interface methods
// These are not used in the tests but required by the interface
func (m *MockAppViewQueries) AssignSubforumRole(ctx context.Context, arg *appview.AssignSubforumRoleParams) (*appview.AssignSubforumRoleRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) AssignUserRole(ctx context.Context, arg *appview.AssignUserRoleParams) (*appview.AssignUserRoleRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CheckUserPermission(ctx context.Context, arg *appview.CheckUserPermissionParams) (bool, error) {
	return false, nil
}
func (m *MockAppViewQueries) CleanupOldProcessedEvents(ctx context.Context) error {
	return nil
}
func (m *MockAppViewQueries) CountCommentsByAuthor(ctx context.Context, authorDid string) (int64, error) {
	return 0, nil
}
func (m *MockAppViewQueries) CountCommentsByPost(ctx context.Context, postID pgtype.UUID) (int64, error) {
	return 0, nil
}
func (m *MockAppViewQueries) CountSubforumSubscribers(ctx context.Context, subforumSlug string) (int64, error) {
	return 0, nil
}
func (m *MockAppViewQueries) CountUserSubscriptions(ctx context.Context, userDid string) (int64, error) {
	return 0, nil
}
func (m *MockAppViewQueries) CreateAppViewComment(ctx context.Context, arg *appview.CreateAppViewCommentParams) (*appview.AppviewComment, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateAppViewPost(ctx context.Context, arg *appview.CreateAppViewPostParams) (*appview.AppviewPost, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateAppViewSubforum(ctx context.Context, arg *appview.CreateAppViewSubforumParams) (*appview.CreateAppViewSubforumRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateComment(ctx context.Context, arg *appview.CreateCommentParams) (*appview.AppviewComment, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateOrUpdateSession(ctx context.Context, arg *appview.CreateOrUpdateSessionParams) (*appview.AppviewSession, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateOrUpdateUserFromDID(ctx context.Context, arg *appview.CreateOrUpdateUserFromDIDParams) (*appview.CreateOrUpdateUserFromDIDRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateSubscription(ctx context.Context, arg *appview.CreateSubscriptionParams) (*appview.AppviewSubscription, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateVote(ctx context.Context, arg *appview.CreateVoteParams) (*appview.AppviewVote, error) {
	return nil, nil
}
func (m *MockAppViewQueries) CreateVoteOnComment(ctx context.Context, arg *appview.CreateVoteOnCommentParams) (*appview.AppviewVote, error) {
	return nil, nil
}
func (m *MockAppViewQueries) DeleteAppViewComment(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockAppViewQueries) DeleteAppViewPost(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockAppViewQueries) DeleteAppViewSubforum(ctx context.Context, slug string) error {
	return nil
}
func (m *MockAppViewQueries) DeleteComment(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockAppViewQueries) DeleteCommentByURI(ctx context.Context, atprotoUri string) error {
	return nil
}
func (m *MockAppViewQueries) DeletePostByAtprotoURI(ctx context.Context, atprotoUri string) error {
	return nil
}
func (m *MockAppViewQueries) DeleteSubscription(ctx context.Context, arg *appview.DeleteSubscriptionParams) error {
	return nil
}
func (m *MockAppViewQueries) DeleteVote(ctx context.Context, arg *appview.DeleteVoteParams) error {
	return nil
}
func (m *MockAppViewQueries) DeleteVoteOnComment(ctx context.Context, arg *appview.DeleteVoteOnCommentParams) error {
	return nil
}
func (m *MockAppViewQueries) GetAllPermissions(ctx context.Context) ([]*appview.GetAllPermissionsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetAllRoles(ctx context.Context) ([]*appview.GetAllRolesRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetAppViewPostByID(ctx context.Context, id uuid.UUID) (*appview.GetAppViewPostByIDRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetAppViewPostByURI(ctx context.Context, atprotoUri string) (*appview.GetAppViewPostByURIRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetAppViewSubforumBySlug(ctx context.Context, slug string) (*appview.GetAppViewSubforumBySlugRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetCommentByID(ctx context.Context, id uuid.UUID) (*appview.AppviewComment, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetCommentByURI(ctx context.Context, atprotoUri string) (*appview.AppviewComment, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetCommentsByAuthorDID(ctx context.Context, arg *appview.GetCommentsByAuthorDIDParams) ([]*appview.GetCommentsByAuthorDIDRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetPostsByAuthorDID(ctx context.Context, arg *appview.GetPostsByAuthorDIDParams) ([]*appview.GetPostsByAuthorDIDRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUserByHandle(ctx context.Context, handle string) (*appview.GetUserByHandleRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) IncrementUserCommentCount(ctx context.Context, userDid string) error {
	return nil
}
func (m *MockAppViewQueries) IncrementUserPostCount(ctx context.Context, userDid string) error {
	return nil
}
func (m *MockAppViewQueries) UpdateUserProfileVisibility(ctx context.Context, arg *appview.UpdateUserProfileVisibilityParams) (*appview.UpdateUserProfileVisibilityRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetCommentVoteCounts(ctx context.Context, id uuid.UUID) (*appview.GetCommentVoteCountsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetCommentWithPost(ctx context.Context, id uuid.UUID) (*appview.GetCommentWithPostRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetModeratedSubforums(ctx context.Context, userDid string) ([]*appview.GetModeratedSubforumsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetPDSServerDetails(ctx context.Context, pdsSource *string) (*appview.GetPDSServerDetailsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetPDSServerStats(ctx context.Context) ([]*appview.GetPDSServerStatsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetPostByAtprotoURI(ctx context.Context, atprotoUri string) (*appview.AppviewPost, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetPostCommentCount(ctx context.Context, id uuid.UUID) (*int32, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetPostVoteCounts(ctx context.Context, id uuid.UUID) (*appview.GetPostVoteCountsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetProcessedEvent(ctx context.Context, eventID string) (*appview.ProcessedEvent, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetProcessedEventsCount(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockAppViewQueries) GetSubforumByID(ctx context.Context, id uuid.UUID) (*appview.AppviewSubforum, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetSubforumBySlug(ctx context.Context, slug string) (*appview.AppviewSubforum, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetSubforumByURI(ctx context.Context, atprotoUri *string) (*appview.AppviewSubforum, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetSubforumMembers(ctx context.Context, arg *appview.GetSubforumMembersParams) ([]*appview.GetSubforumMembersRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetSubforumStats(ctx context.Context, slug string) (*appview.GetSubforumStatsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetSubforumSubscriberCount(ctx context.Context, slug string) (*int32, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUserByDID(ctx context.Context, did string) (*appview.GetUserByDIDRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUserPermissions(ctx context.Context, arg *appview.GetUserPermissionsParams) ([]string, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUserRoles(ctx context.Context, userDid string) ([]*appview.GetUserRolesRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUserSubscription(ctx context.Context, arg *appview.GetUserSubscriptionParams) (*appview.AppviewSubscription, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUserVoteOnComment(ctx context.Context, arg *appview.GetUserVoteOnCommentParams) (*appview.AppviewVote, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUserVoteOnPost(ctx context.Context, arg *appview.GetUserVoteOnPostParams) (*appview.AppviewVote, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUsersByDIDs(ctx context.Context, dollar_1 []string) ([]*appview.GetUsersByDIDsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUsersByPDSSource(ctx context.Context, arg *appview.GetUsersByPDSSourceParams) ([]*appview.GetUsersByPDSSourceRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) GetUsersWithRoles(ctx context.Context, arg *appview.GetUsersWithRolesParams) ([]*appview.GetUsersWithRolesRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) HasUserRole(ctx context.Context, arg *appview.HasUserRoleParams) (bool, error) {
	return false, nil
}
func (m *MockAppViewQueries) ListAppViewPosts(ctx context.Context, arg *appview.ListAppViewPostsParams) ([]*appview.ListAppViewPostsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListAppViewPostsBySubforum(ctx context.Context, arg *appview.ListAppViewPostsBySubforumParams) ([]*appview.ListAppViewPostsBySubforumRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListAppViewSubforums(ctx context.Context, arg *appview.ListAppViewSubforumsParams) ([]*appview.ListAppViewSubforumsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListCommentsByAuthor(ctx context.Context, arg *appview.ListCommentsByAuthorParams) ([]*appview.ListCommentsByAuthorRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListCommentsByPost(ctx context.Context, arg *appview.ListCommentsByPostParams) ([]*appview.ListCommentsByPostRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListCommentsByPostWithReplies(ctx context.Context, arg *appview.ListCommentsByPostWithRepliesParams) ([]*appview.ListCommentsByPostWithRepliesRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListCommentsByPostWithUserVotes(ctx context.Context, arg *appview.ListCommentsByPostWithUserVotesParams) ([]*appview.ListCommentsByPostWithUserVotesRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListSubforumSubscribers(ctx context.Context, arg *appview.ListSubforumSubscribersParams) ([]*appview.ListSubforumSubscribersRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) ListUserSubscriptions(ctx context.Context, arg *appview.ListUserSubscriptionsParams) ([]*appview.ListUserSubscriptionsRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) RevokeSubforumRole(ctx context.Context, arg *appview.RevokeSubforumRoleParams) error {
	return nil
}
func (m *MockAppViewQueries) RevokeUserRole(ctx context.Context, arg *appview.RevokeUserRoleParams) error {
	return nil
}
func (m *MockAppViewQueries) UpdateAppViewComment(ctx context.Context, arg *appview.UpdateAppViewCommentParams) (*appview.AppviewComment, error) {
	return nil, nil
}
func (m *MockAppViewQueries) UpdateAppViewPost(ctx context.Context, arg *appview.UpdateAppViewPostParams) (*appview.AppviewPost, error) {
	return nil, nil
}
func (m *MockAppViewQueries) UpdateAppViewSubforum(ctx context.Context, arg *appview.UpdateAppViewSubforumParams) (*appview.UpdateAppViewSubforumRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) UpdateComment(ctx context.Context, arg *appview.UpdateCommentParams) (*appview.UpdateCommentRow, error) {
	return nil, nil
}
func (m *MockAppViewQueries) UpdateCommentByURI(ctx context.Context, arg *appview.UpdateCommentByURIParams) (*appview.AppviewComment, error) {
	return nil, nil
}
func (m *MockAppViewQueries) UpdateCommentVoteCounts(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockAppViewQueries) UpdatePostByAtprotoURI(ctx context.Context, arg *appview.UpdatePostByAtprotoURIParams) (*appview.AppviewPost, error) {
	return nil, nil
}
func (m *MockAppViewQueries) UpdatePostCommentCount(ctx context.Context, postID pgtype.UUID) error {
	return nil
}
func (m *MockAppViewQueries) UpdatePostVoteCounts(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockAppViewQueries) UpdateSubforumCommentCount(ctx context.Context, arg *appview.UpdateSubforumCommentCountParams) error {
	return nil
}
func (m *MockAppViewQueries) UpdateSubforumPostCount(ctx context.Context, arg *appview.UpdateSubforumPostCountParams) error {
	return nil
}
func (m *MockAppViewQueries) UpdateSubforumSubscriberCount(ctx context.Context, subforumSlug string) error {
	return nil
}
func (m *MockAppViewQueries) UpdateUserLastSeen(ctx context.Context, did string) error {
	return nil
}
func (m *MockAppViewQueries) UpdateUserProfile(ctx context.Context, arg *appview.UpdateUserProfileParams) (*appview.UpdateUserProfileRow, error) {
	return nil, nil
}

// MockIdentityDirectory is a mock implementation of identity.Directory
type MockIdentityDirectory struct {
	identities map[string]identity.Identity
}

// NewMockIdentityDirectory creates a new mock identity directory
func NewMockIdentityDirectory() *MockIdentityDirectory {
	return &MockIdentityDirectory{
		identities: make(map[string]identity.Identity),
	}
}

// Insert adds an identity to the mock directory
func (m *MockIdentityDirectory) Insert(identity identity.Identity) {
	m.identities[identity.DID.String()] = identity
	m.identities[identity.Handle.String()] = identity
}

// LookupHandle looks up a handle in the mock directory
func (m *MockIdentityDirectory) LookupHandle(ctx context.Context, handle syntax.Handle) (identity.Identity, error) {
	if identity, exists := m.identities[handle.String()]; exists {
		return identity, nil
	}
	return identity.Identity{}, identity.ErrHandleNotFound
}

// LookupDID looks up a DID in the mock directory
func (m *MockIdentityDirectory) LookupDID(ctx context.Context, did syntax.DID) (identity.Identity, error) {
	if identity, exists := m.identities[did.String()]; exists {
		return identity, nil
	}
	return identity.Identity{}, identity.ErrDIDNotFound
}

// Lookup looks up an identity by handle or DID
func (m *MockIdentityDirectory) Lookup(ctx context.Context, identifier string) (identity.Identity, error) {
	// Try as handle first
	if identity, exists := m.identities[identifier]; exists {
		return identity, nil
	}

	// Try as DID
	if identity, exists := m.identities[identifier]; exists {
		return identity, nil
	}

	return identity.Identity{}, identity.ErrHandleNotFound
}

// CreateTestIdentities creates common test identities
func CreateTestIdentities() *MockIdentityDirectory {
	dir := NewMockIdentityDirectory()

	// Add test identities
	dir.Insert(identity.Identity{
		DID:    syntax.DID("did:plc:test-user-1"),
		Handle: syntax.Handle("testuser1.hashpost.local"),
	})

	dir.Insert(identity.Identity{
		DID:    syntax.DID("did:plc:test-user-2"),
		Handle: syntax.Handle("testuser2.hashpost.local"),
	})

	dir.Insert(identity.Identity{
		DID:    syntax.DID("did:plc:test-admin"),
		Handle: syntax.Handle("admin.hashpost.local"),
	})

	return dir
}

// CreateMockLogger creates a mock logger for testing
func CreateMockLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))
}
