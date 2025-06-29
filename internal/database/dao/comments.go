package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// CommentDAO provides data access operations for comments
type CommentDAO struct {
	db bob.Executor
}

// NewCommentDAO creates a new CommentDAO
func NewCommentDAO(db bob.Executor) *CommentDAO {
	return &CommentDAO{
		db: db,
	}
}

// CreateComment creates a new comment
func (dao *CommentDAO) CreateComment(ctx context.Context, postID int64, pseudonymID, content string, parentCommentID *int64) (*models.Comment, error) {
	log.Debug().
		Int64("post_id", postID).
		Str("pseudonym_id", pseudonymID).
		Str("content", content).
		Msg("Creating comment")

	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	parentCommentIDNull := sql.Null[int64]{Valid: false}
	if parentCommentID != nil {
		parentCommentIDNull.Scan(*parentCommentID)
	}

	commentSetter := &models.CommentSetter{
		PostID:          &postID,
		PseudonymID:     &pseudonymID,
		Content:         &content,
		ParentCommentID: &parentCommentIDNull,
		CreatedAt:       &now,
		UpdatedAt:       &now,
	}

	// Use the generated Comments table helper
	comment, err := models.Comments.Insert(commentSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// GetCommentByID retrieves a comment by ID with related data
func (dao *CommentDAO) GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error) {
	comment, err := models.FindComment(ctx, dao.db, commentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get comment by ID: %w", err)
	}

	// Load related data
	if err := comment.LoadPseudonym(ctx, dao.db); err != nil {
		log.Warn().Err(err).Int64("comment_id", commentID).Msg("Failed to load comment pseudonym")
	}

	return comment, nil
}

// GetCommentsByPost retrieves comments for a post
func (dao *CommentDAO) GetCommentsByPost(ctx context.Context, postID int64) ([]*models.Comment, error) {
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.EQ(postID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.OrderBy("score DESC NULLS LAST, created_at ASC"),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post: %w", err)
	}

	// Load related data for all comments
	for _, comment := range comments {
		if err := comment.LoadPseudonym(ctx, dao.db); err != nil {
			log.Warn().Err(err).Int64("comment_id", comment.CommentID).Msg("Failed to load comment pseudonym")
		}
	}

	return comments, nil
}

// GetCommentsByPostWithNestedReplies retrieves comments for a post and builds nested reply structure
func (dao *CommentDAO) GetCommentsByPostWithNestedReplies(ctx context.Context, postID int64) ([]*models.Comment, error) {
	// Get all comments for the post, ordered by score (descending) then creation time (ascending)
	allComments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.EQ(postID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.OrderBy("score DESC NULLS LAST, created_at ASC"),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post: %w", err)
	}

	// Load related data for all comments
	for _, comment := range allComments {
		if err := comment.LoadPseudonym(ctx, dao.db); err != nil {
			log.Warn().Err(err).Int64("comment_id", comment.CommentID).Msg("Failed to load comment pseudonym")
		}
	}

	// Build nested structure
	return dao.buildNestedCommentTree(allComments), nil
}

// buildNestedCommentTree builds a nested tree structure from flat comments
func (dao *CommentDAO) buildNestedCommentTree(allComments []*models.Comment) []*models.Comment {
	// Create maps for quick lookup
	commentMap := make(map[int64]*models.Comment)
	rootComments := make([]*models.Comment, 0)

	// First pass: create map and identify root comments
	for _, comment := range allComments {
		commentMap[comment.CommentID] = comment
		comment.R.ReverseComments = make([]*models.Comment, 0) // Initialize replies slice

		// If no parent, it's a root comment
		if !comment.ParentCommentID.Valid {
			rootComments = append(rootComments, comment)
		}
	}

	// Second pass: build the tree structure
	for _, comment := range allComments {
		if comment.ParentCommentID.Valid {
			parent, exists := commentMap[comment.ParentCommentID.V]
			if exists {
				parent.R.ReverseComments = append(parent.R.ReverseComments, comment)
			}
		}
	}

	return rootComments
}

// CountCommentsByPost counts total comments for a post
func (dao *CommentDAO) CountCommentsByPost(ctx context.Context, postID int64) (int64, error) {
	count, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.EQ(postID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by post: %w", err)
	}

	return count, nil
}

// UpdateCommentScore updates the comment score and vote counts
func (dao *CommentDAO) UpdateCommentScore(ctx context.Context, commentID int64, score, upvotes, downvotes int32) error {
	updates := &models.CommentSetter{
		Score:     &sql.Null[int32]{Valid: true, V: score},
		Upvotes:   &sql.Null[int32]{Valid: true, V: upvotes},
		Downvotes: &sql.Null[int32]{Valid: true, V: downvotes},
		UpdatedAt: &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}

	comment, err := models.FindComment(ctx, dao.db, commentID)
	if err != nil {
		return fmt.Errorf("failed to find comment for score update: %w", err)
	}

	err = comment.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update comment score: %w", err)
	}

	return nil
}

// CountCommentsByPseudonym counts total comments by a pseudonym
func (dao *CommentDAO) CountCommentsByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	count, err := models.Comments.Query(
		models.SelectWhere.Comments.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by pseudonym: %w", err)
	}

	return count, nil
}

// CountCommentsByPseudonymInSubforum counts comments by a pseudonym in a specific subforum
func (dao *CommentDAO) CountCommentsByPseudonymInSubforum(ctx context.Context, pseudonymID string, subforumID int32) (int64, error) {
	// Get all comments by the pseudonym
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments by pseudonym: %w", err)
	}

	// Count comments in the specific subforum
	count := int64(0)
	for _, comment := range comments {
		post, err := models.FindPost(ctx, dao.db, comment.PostID)
		if err == nil && post != nil && post.SubforumID == subforumID {
			count++
		}
	}

	return count, nil
}

// GetSubforumsByPseudonymComments gets all subforums where a pseudonym has commented
func (dao *CommentDAO) GetSubforumsByPseudonymComments(ctx context.Context, pseudonymID string) ([]int32, error) {
	// Get all comments by the pseudonym
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by pseudonym: %w", err)
	}

	// Extract unique subforum IDs
	subforumMap := make(map[int32]bool)
	for _, comment := range comments {
		post, err := models.FindPost(ctx, dao.db, comment.PostID)
		if err == nil && post != nil {
			subforumMap[post.SubforumID] = true
		}
	}

	subforums := make([]int32, 0, len(subforumMap))
	for subforumID := range subforumMap {
		subforums = append(subforums, subforumID)
	}

	return subforums, nil
}
