package appview

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// Database represents the AppView database connection
type Database struct {
	queries *generated.Queries
	logger  *slog.Logger
}

// NewDatabase creates a new AppView database connection
func NewDatabase(db *pgxpool.Pool, logger *slog.Logger) *Database {
	queries := generated.New(db)
	return &Database{
		queries: queries,
		logger:  logger,
	}
}

// Close closes the database connection
func (d *Database) Close() error {
	// The queries don't have a Close method, so we don't need to do anything
	// The pgxpool.Pool will be closed by the server
	return nil
}

// AppView Data Models

type AppViewUser struct {
	ID           uuid.UUID  `json:"id"`
	DID          string     `json:"did"`
	Handle       string     `json:"handle"`
	DisplayName  string     `json:"display_name"`
	Bio          string     `json:"bio"`
	AvatarURL    string     `json:"avatar_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PostCount    int        `json:"post_count"`
	CommentCount int        `json:"comment_count"`
	Reputation   int        `json:"reputation"`
	PDSSource    *string    `json:"pds_source,omitempty"`   // PDS endpoint where user data is stored
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"` // Last authentication time
}

type AppViewSubforum struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	CreatedByDID    string    `json:"created_by_did"`
	CreatedByHandle string    `json:"created_by_handle"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	SubscriberCount int       `json:"subscriber_count"`
	PostCount       int       `json:"post_count"`
	CommentCount    int       `json:"comment_count"`
}

type AppViewPost struct {
	ID           uuid.UUID `json:"id"`
	AtprotoURI   string    `json:"atproto_uri"`
	AuthorDID    string    `json:"author_did"`
	AuthorHandle string    `json:"author_handle"`
	SubforumSlug string    `json:"subforum_slug"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Upvotes      int       `json:"upvotes"`
	Downvotes    int       `json:"downvotes"`
	CommentCount int       `json:"comment_count"`
	Score        int       `json:"score"`
}

type AppViewComment struct {
	ID           uuid.UUID  `json:"id"`
	AtprotoURI   string     `json:"atproto_uri"`
	AuthorDID    string     `json:"author_did"`
	AuthorHandle string     `json:"author_handle"`
	PostID       uuid.UUID  `json:"post_id"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty"`
	Content      string     `json:"content"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Upvotes      int        `json:"upvotes"`
	Downvotes    int        `json:"downvotes"`
	Score        int        `json:"score"`
}

// Database Operations

// CreateUser creates a new user in the AppView database
func (d *Database) CreateUser(user *AppViewUser) error {
	_, err := d.queries.CreateOrUpdateUserFromDID(context.Background(), &generated.CreateOrUpdateUserFromDIDParams{
		Did:         user.DID,
		Handle:      user.Handle,
		DisplayName: &user.DisplayName,
		AvatarUrl:   &user.AvatarURL,
	})

	if err != nil {
		d.logger.Error("Failed to create user", "error", err, "did", user.DID)
		return fmt.Errorf("failed to create user: %w", err)
	}

	d.logger.Info("User created/updated in AppView", "did", user.DID, "handle", user.Handle)
	return nil
}

// CreateSubforum creates a new subforum in the AppView database
func (d *Database) CreateSubforum(subforum *AppViewSubforum) error {
	_, err := d.queries.CreateAppViewSubforum(context.Background(), &generated.CreateAppViewSubforumParams{
		Name:            subforum.Name,
		Slug:            subforum.Slug,
		Description:     &subforum.Description,
		CreatedByDid:    subforum.CreatedByDID,
		CreatedByHandle: subforum.CreatedByHandle,
	})

	if err != nil {
		d.logger.Error("Failed to create subforum", "error", err, "slug", subforum.Slug)
		return fmt.Errorf("failed to create subforum: %w", err)
	}

	d.logger.Info("Subforum created/updated in AppView", "slug", subforum.Slug, "name", subforum.Name)
	return nil
}

// CreatePost creates a new post in the AppView database
func (d *Database) CreatePost(post *AppViewPost) error {
	_, err := d.queries.CreateAppViewPost(context.Background(), &generated.CreateAppViewPostParams{
		AtprotoUri:   post.AtprotoURI,
		AuthorDid:    post.AuthorDID,
		AuthorHandle: post.AuthorHandle,
		SubforumSlug: post.SubforumSlug,
		Title:        post.Title,
		Content:      post.Content,
	})

	if err != nil {
		d.logger.Error("Failed to create post", "error", err, "uri", post.AtprotoURI)
		return fmt.Errorf("failed to create post: %w", err)
	}

	d.logger.Info("Post created/updated in AppView", "uri", post.AtprotoURI, "title", post.Title)
	return nil
}

// UpdatePost updates an existing post in the AppView database
func (d *Database) UpdatePost(atprotoURI string, post *AppViewPost) error {
	_, err := d.queries.UpdatePostByAtprotoURI(context.Background(), &generated.UpdatePostByAtprotoURIParams{
		AtprotoUri: atprotoURI,
		Title:      post.Title,
		Content:    post.Content,
	})

	if err != nil {
		d.logger.Error("Failed to update post", "error", err, "uri", atprotoURI)
		return fmt.Errorf("failed to update post: %w", err)
	}

	d.logger.Info("Post updated in AppView", "uri", atprotoURI)
	return nil
}

// DeletePost removes a post from the AppView database
func (d *Database) DeletePost(atprotoURI string) error {
	err := d.queries.DeletePostByAtprotoURI(context.Background(), atprotoURI)
	if err != nil {
		d.logger.Error("Failed to delete post", "error", err, "uri", atprotoURI)
		return fmt.Errorf("failed to delete post: %w", err)
	}

	d.logger.Info("Post deleted from AppView", "uri", atprotoURI)
	return nil
}

// GetPostByURI retrieves a post by its atproto URI
func (d *Database) GetPostByURI(uri string) (*generated.AppviewPost, error) {
	post, err := d.queries.GetPostByAtprotoURI(context.Background(), uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by URI: %w", err)
	}
	return post, nil
}

// GetCommentByURI retrieves a comment by its atproto URI
func (d *Database) GetCommentByURI(uri string) (*generated.AppviewComment, error) {
	comment, err := d.queries.GetCommentByURI(context.Background(), uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment by URI: %w", err)
	}
	return comment, nil
}

// CreateComment creates a new comment in the AppView database
func (d *Database) CreateComment(comment *AppViewComment) error {
	// Convert UUIDs to pgtype.UUID
	postID := pgtype.UUID{Bytes: comment.PostID, Valid: true}
	var parentID pgtype.UUID
	if comment.ParentID != nil {
		parentID = pgtype.UUID{Bytes: *comment.ParentID, Valid: true}
	}

	_, err := d.queries.CreateComment(context.Background(), &generated.CreateCommentParams{
		AtprotoUri:   comment.AtprotoURI,
		AuthorDid:    comment.AuthorDID,
		AuthorHandle: comment.AuthorHandle,
		PostID:       postID,
		ParentID:     parentID,
		Content:      comment.Content,
	})

	if err != nil {
		d.logger.Error("Failed to create comment", "error", err, "uri", comment.AtprotoURI)
		return fmt.Errorf("failed to create comment: %w", err)
	}

	d.logger.Info("Comment created in AppView", "uri", comment.AtprotoURI)
	return nil
}

// UpdateComment updates an existing comment in the AppView database
func (d *Database) UpdateComment(atprotoURI string, comment *AppViewComment) error {
	_, err := d.queries.UpdateCommentByURI(context.Background(), &generated.UpdateCommentByURIParams{
		AtprotoUri: atprotoURI,
		Content:    comment.Content,
	})

	if err != nil {
		d.logger.Error("Failed to update comment", "error", err, "uri", atprotoURI)
		return fmt.Errorf("failed to update comment: %w", err)
	}

	d.logger.Info("Comment updated in AppView", "uri", atprotoURI)
	return nil
}

// DeleteCommentByURI removes a comment from the AppView database
func (d *Database) DeleteCommentByURI(atprotoURI string) error {
	err := d.queries.DeleteCommentByURI(context.Background(), atprotoURI)
	if err != nil {
		d.logger.Error("Failed to delete comment", "error", err, "uri", atprotoURI)
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	d.logger.Info("Comment deleted from AppView", "uri", atprotoURI)
	return nil
}
