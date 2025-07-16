package dao

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// generateSlug creates a URL-friendly slug from a title and post ID
func generateSlug(title string, postID int64) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	slug := re.ReplaceAllString(strings.ToLower(title), "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 50 {
		slug = slug[:50]
	}
	if slug == "" {
		slug = "post"
	}
	return fmt.Sprintf("%s-%d", slug, postID)
}

// PostDAO provides data access operations for posts
type PostDAO struct {
	db bob.Executor
}

// NewPostDAO creates a new PostDAO
func NewPostDAO(db bob.Executor) *PostDAO {
	return &PostDAO{
		db: db,
	}
}

// CreatePost creates a new post
func (dao *PostDAO) CreatePost(ctx context.Context, subforumID int32, pseudonymID, title, content, postType string, url *string, isNSFW, isSpoiler bool) (*models.Post, error) {
	log.Debug().
		Int32("subforum_id", subforumID).
		Str("pseudonym_id", pseudonymID).
		Str("title", title).
		Str("post_type", postType).
		Msg("Creating post")

	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	isSelfPost := sql.Null[bool]{}
	isSelfPost.Scan(url == nil || *url == "")

	nsfw := sql.Null[bool]{}
	nsfw.Scan(isNSFW)

	spoiler := sql.Null[bool]{}
	spoiler.Scan(isSpoiler)

	contentNull := sql.Null[string]{}
	if content != "" {
		contentNull.Scan(content)
	}

	urlNull := sql.Null[string]{}
	if url != nil && *url != "" {
		urlNull.Scan(*url)
	}

	postSetter := &models.PostSetter{
		SubforumID:  &subforumID,
		PseudonymID: &pseudonymID,
		Title:       &title,
		Content:     &contentNull,
		PostType:    &postType,
		URL:         &urlNull,
		IsSelfPost:  &isSelfPost,
		IsNSFW:      &nsfw,
		IsSpoiler:   &spoiler,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// Use the generated Posts table helper
	post, err := models.Posts.Insert(postSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Generate slug after post is created (we need the post ID)
	slug := generateSlug(title, post.PostID)

	// Update the post with the slug
	slugNull := sql.Null[string]{}
	slugNull.Scan(slug)

	updateSetter := &models.PostSetter{
		Slug: &slugNull,
	}

	err = post.Update(ctx, dao.db, updateSetter)
	if err != nil {
		log.Warn().Err(err).Int64("post_id", post.PostID).Msg("Failed to update post with slug")
		// Don't fail the request, just log the warning
	}

	return post, nil
}

// GetPostByID retrieves a post by ID with related data
func (dao *PostDAO) GetPostByID(ctx context.Context, postID int64) (*models.Post, error) {
	post, err := models.FindPost(ctx, dao.db, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post by ID: %w", err)
	}

	// Load related data
	if err := post.LoadPseudonym(ctx, dao.db); err != nil {
		log.Warn().Err(err).Int64("post_id", postID).Msg("Failed to load post pseudonym")
	}

	if err := post.LoadSubforum(ctx, dao.db); err != nil {
		log.Warn().Err(err).Int64("post_id", postID).Msg("Failed to load post subforum")
	}

	return post, nil
}

// GetPostsBySubforum retrieves posts from a subforum with pagination and sorting
func (dao *PostDAO) GetPostsBySubforum(ctx context.Context, subforumID int32, page, limit int, sortField string, sortDesc bool) ([]*models.Post, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}

	// Determine the column to sort by
	var orderExpr psql.Expression
	switch sortField {
	case "score":
		orderExpr = models.PostColumns.Score
	case "created_at":
		orderExpr = models.PostColumns.CreatedAt
	case "updated_at":
		orderExpr = models.PostColumns.UpdatedAt
	case "comment_count":
		orderExpr = models.PostColumns.CommentCount
	case "view_count":
		orderExpr = models.PostColumns.ViewCount
	default:
		orderExpr = models.PostColumns.CreatedAt // default
	}

	direction := "ASC"
	if sortDesc {
		direction = "DESC"
	}

	// Always order by is_stickied DESC first, then by the selected field
	orderByClause := fmt.Sprintf("is_stickied DESC, %s %s", orderExpr, direction)

	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforumID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("posts", "is_removed").IsNull(),
			psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.OrderBy(orderByClause),
		sm.Limit(limit),
		sm.Offset((page-1)*limit),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by subforum: %w", err)
	}

	// Load related data for paginated posts
	for _, post := range posts {
		if err := post.LoadPseudonym(ctx, dao.db); err != nil {
			log.Warn().Err(err).Int64("post_id", post.PostID).Msg("Failed to load post pseudonym")
		}
		if err := post.LoadSubforum(ctx, dao.db); err != nil {
			log.Warn().Err(err).Int64("post_id", post.PostID).Msg("Failed to load post subforum")
		}
	}

	return posts, nil
}

// CountPostsBySubforum counts total posts in a subforum
func (dao *PostDAO) CountPostsBySubforum(ctx context.Context, subforumID int32) (int64, error) {
	count, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforumID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("posts", "is_removed").IsNull(),
			psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
		))),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts by subforum: %w", err)
	}

	return count, nil
}

// UpdatePostScore updates the post score and vote counts
func (dao *PostDAO) UpdatePostScore(ctx context.Context, postID int64, score, upvotes, downvotes int32) error {
	updates := &models.PostSetter{
		Score:     &sql.Null[int32]{Valid: true, V: score},
		Upvotes:   &sql.Null[int32]{Valid: true, V: upvotes},
		Downvotes: &sql.Null[int32]{Valid: true, V: downvotes},
		UpdatedAt: &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}

	post, err := models.FindPost(ctx, dao.db, postID)
	if err != nil {
		return fmt.Errorf("failed to find post for score update: %w", err)
	}

	err = post.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update post score: %w", err)
	}

	return nil
}

// IncrementViewCount increments the view count for a post
func (dao *PostDAO) IncrementViewCount(ctx context.Context, postID int64) error {
	post, err := models.FindPost(ctx, dao.db, postID)
	if err != nil {
		return fmt.Errorf("failed to find post for view count increment: %w", err)
	}

	currentViewCount := int32(0)
	if post.ViewCount.Valid {
		currentViewCount = post.ViewCount.V
	}

	newViewCount := currentViewCount + 1
	updates := &models.PostSetter{
		ViewCount: &sql.Null[int32]{Valid: true, V: newViewCount},
		UpdatedAt: &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}

	err = post.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

// UpdateCommentCount updates the comment count for a post
func (dao *PostDAO) UpdateCommentCount(ctx context.Context, postID int64, commentCount int32) error {
	updates := &models.PostSetter{
		CommentCount: &sql.Null[int32]{Valid: true, V: commentCount},
		UpdatedAt:    &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}

	post, err := models.FindPost(ctx, dao.db, postID)
	if err != nil {
		return fmt.Errorf("failed to find post for comment count update: %w", err)
	}

	err = post.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update comment count: %w", err)
	}

	return nil
}

// CountPostsByPseudonym counts total posts by a pseudonym
func (dao *PostDAO) CountPostsByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	count, err := models.Posts.Query(
		models.SelectWhere.Posts.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("posts", "is_removed").IsNull(),
			psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
		))),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts by pseudonym: %w", err)
	}

	return count, nil
}

// CountPostsByPseudonymInSubforum counts posts by a pseudonym in a specific subforum
func (dao *PostDAO) CountPostsByPseudonymInSubforum(ctx context.Context, pseudonymID string, subforumID int32) (int64, error) {
	count, err := models.Posts.Query(
		models.SelectWhere.Posts.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.Posts.SubforumID.EQ(subforumID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("posts", "is_removed").IsNull(),
			psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
		))),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts by pseudonym in subforum: %w", err)
	}

	return count, nil
}

// GetSubforumsByPseudonym gets all subforums where a pseudonym has posted
func (dao *PostDAO) GetSubforumsByPseudonym(ctx context.Context, pseudonymID string) ([]int32, error) {
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("posts", "is_removed").IsNull(),
			psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.Columns(models.PostColumns.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subforums by pseudonym: %w", err)
	}

	// Extract unique subforum IDs
	subforumMap := make(map[int32]bool)
	for _, post := range posts {
		subforumMap[post.SubforumID] = true
	}

	subforums := make([]int32, 0, len(subforumMap))
	for subforumID := range subforumMap {
		subforums = append(subforums, subforumID)
	}

	return subforums, nil
}

// SetLocked sets the is_locked field for a post
func (dao *PostDAO) SetLocked(ctx context.Context, postID int64, locked bool) error {
	post, err := models.FindPost(ctx, dao.db, postID)
	if err != nil {
		return fmt.Errorf("failed to find post for lock update: %w", err)
	}
	updates := &models.PostSetter{
		IsLocked:  &sql.Null[bool]{Valid: true, V: locked},
		UpdatedAt: &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}
	return post.Update(ctx, dao.db, updates)
}

// SetSticky sets the is_stickied field for a post
func (dao *PostDAO) SetSticky(ctx context.Context, postID int64, sticky bool) error {
	post, err := models.FindPost(ctx, dao.db, postID)
	if err != nil {
		return fmt.Errorf("failed to find post for sticky update: %w", err)
	}
	updates := &models.PostSetter{
		IsStickied: &sql.Null[bool]{Valid: true, V: sticky},
		UpdatedAt:  &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}
	return post.Update(ctx, dao.db, updates)
}

// SetRemoved sets the is_removed field for a post
func (dao *PostDAO) SetRemoved(ctx context.Context, postID int64, removed bool) error {
	post, err := models.FindPost(ctx, dao.db, postID)
	if err != nil {
		return fmt.Errorf("failed to find post for remove update: %w", err)
	}
	updates := &models.PostSetter{
		IsRemoved: &sql.Null[bool]{Valid: true, V: removed},
		UpdatedAt: &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}
	return post.Update(ctx, dao.db, updates)
}

// GetPostBySubforumAndSlug retrieves a post by subforum ID and slug with related data
func (dao *PostDAO) GetPostBySubforumAndSlug(ctx context.Context, subforumID int32, slug string) (*models.Post, error) {
	post, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforumID),
		models.SelectWhere.Posts.Slug.EQ(slug),
	).One(ctx, dao.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post by subforum and slug: %w", err)
	}

	// Load related data
	if err := post.LoadPseudonym(ctx, dao.db); err != nil {
		log.Warn().Err(err).Int64("post_id", post.PostID).Msg("Failed to load post pseudonym")
	}

	if err := post.LoadSubforum(ctx, dao.db); err != nil {
		log.Warn().Err(err).Int64("post_id", post.PostID).Msg("Failed to load post subforum")
	}

	return post, nil
}

// MarkPostAsDeletedByPseudonym marks a post as deleted by the pseudonym
func (dao *PostDAO) MarkPostAsDeletedByPseudonym(ctx context.Context, postID int64, pseudonymID string, reason string) error {
	post, err := models.Posts.Query(
		models.SelectWhere.Posts.PostID.EQ(postID),
	).One(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to find post: %w", err)
	}

	// Check if the post belongs to the pseudonym
	if post.PseudonymID != pseudonymID {
		return fmt.Errorf("post does not belong to pseudonym")
	}

	// Check if post is already deleted
	if post.IsDeleted.Valid && post.IsDeleted.V {
		return fmt.Errorf("post is already deleted")
	}

	now := time.Now()
	updates := &models.PostSetter{
		IsDeleted:                &sql.Null[bool]{V: true, Valid: true},
		DeletedByPseudonymID:     &sql.Null[string]{V: pseudonymID, Valid: true},
		DeletedByPseudonymAt:     &sql.Null[time.Time]{V: now, Valid: true},
		DeletedByPseudonymReason: &sql.Null[string]{V: reason, Valid: true},
	}

	err = post.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to mark post as deleted by pseudonym: %w", err)
	}

	return nil
}
