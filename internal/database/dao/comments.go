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
	comment, err := models.Comments.Query(
		models.SelectWhere.Comments.CommentID.EQ(commentID),
		sm.Where(psql.Group(psql.And(
			psql.Group(psql.Or(
				psql.Quote("comments", "is_removed").IsNull(),
				psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
			)),
			psql.Group(psql.Or(
				psql.Quote("comments", "is_deleted").IsNull(),
				psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
			)),
		))),
	).One(ctx, dao.db)
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

// MarkCommentAsDeletedByPseudonym marks a comment as deleted by the pseudonym
func (dao *CommentDAO) MarkCommentAsDeletedByPseudonym(ctx context.Context, commentID int64, pseudonymID string, reason string) error {
	comment, err := models.Comments.Query(
		models.SelectWhere.Comments.CommentID.EQ(commentID),
		sm.Where(psql.Group(psql.And(
			psql.Group(psql.Or(
				psql.Quote("comments", "is_removed").IsNull(),
				psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
			)),
			psql.Group(psql.Or(
				psql.Quote("comments", "is_deleted").IsNull(),
				psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
			)),
		))),
	).One(ctx, dao.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("comment not found")
		}
		return fmt.Errorf("failed to find comment: %w", err)
	}

	// Check if the comment belongs to the pseudonym
	if comment.PseudonymID != pseudonymID {
		return fmt.Errorf("comment does not belong to pseudonym")
	}

	now := time.Now()
	updates := &models.CommentSetter{
		IsDeleted:                &sql.Null[bool]{V: true, Valid: true},
		DeletedByPseudonymID:     &sql.Null[string]{V: pseudonymID, Valid: true},
		DeletedByPseudonymAt:     &sql.Null[time.Time]{V: now, Valid: true},
		DeletedByPseudonymReason: &sql.Null[string]{V: reason, Valid: true},
	}

	err = comment.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to mark comment as deleted by pseudonym: %w", err)
	}

	return nil
}

// GetCommentsByPost retrieves comments for a post (excluding user-deleted ones)
func (dao *CommentDAO) GetCommentsByPost(ctx context.Context, postID int64) ([]*models.Comment, error) {
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.EQ(postID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_deleted").IsNull(),
			psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
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
// Includes deleted comments but clears their content and user info, and freezes voting
func (dao *CommentDAO) GetCommentsByPostWithNestedReplies(ctx context.Context, postID int64) ([]*models.Comment, error) {
	// Get all comments for the post, including deleted ones (but excluding moderator-removed ones)
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

		// Load deleted by pseudonym if comment is deleted
		if comment.IsDeleted.Valid && comment.IsDeleted.V && comment.DeletedByPseudonymID.Valid {
			if err := comment.LoadDeletedByPseudonymPseudonym(ctx, dao.db); err != nil {
				log.Warn().Err(err).Int64("comment_id", comment.CommentID).Msg("Failed to load deleted by pseudonym")
			}
		}

		// Clear content and user info for deleted comments
		if comment.IsDeleted.Valid && comment.IsDeleted.V {
			// Clear content
			comment.Content = "[deleted]"

			// Clear author info but keep the pseudonym ID for reference
			if comment.R.Pseudonym != nil {
				comment.R.Pseudonym.DisplayName = "[deleted]"
				comment.R.Pseudonym.Bio = sql.Null[string]{Valid: false}
			}

			// Note: UserVote is computed at the API level, not stored in the database
			// The API layer will handle freezing voting for deleted comments
		}
	}

	// Build nested structure
	commentMap := make(map[int64]*models.Comment)
	var rootComments []*models.Comment

	for _, comment := range allComments {
		commentMap[comment.CommentID] = comment
	}

	for _, comment := range allComments {
		if comment.ParentCommentID.Valid {
			parent, exists := commentMap[comment.ParentCommentID.V]
			if exists {
				parent.R.ReverseComments = append(parent.R.ReverseComments, comment)
			}
		} else {
			rootComments = append(rootComments, comment)
		}
	}

	return rootComments, nil
}

// CountCommentsByPost counts total comments for a post (excluding user-deleted ones)
func (dao *CommentDAO) CountCommentsByPost(ctx context.Context, postID int64) (int64, error) {
	count, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.EQ(postID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_deleted").IsNull(),
			psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
		))),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by post: %w", err)
	}

	return count, nil
}

// FindCommentForScoreUpdate retrieves a comment by ID for score updates, including deleted comments
// This method is specifically for score updates where we need to update scores even for deleted content
func (dao *CommentDAO) FindCommentForScoreUpdate(ctx context.Context, commentID int64) (*models.Comment, error) {
	comment, err := models.FindComment(ctx, dao.db, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find comment for score update: %w", err)
	}
	return comment, nil
}

// UpdateCommentScore updates the comment score and vote counts
func (dao *CommentDAO) UpdateCommentScore(ctx context.Context, commentID int64, score, upvotes, downvotes int32) error {
	updates := &models.CommentSetter{
		Score:     &sql.Null[int32]{Valid: true, V: score},
		Upvotes:   &sql.Null[int32]{Valid: true, V: upvotes},
		Downvotes: &sql.Null[int32]{Valid: true, V: downvotes},
		UpdatedAt: &sql.Null[time.Time]{Valid: true, V: time.Now()},
	}

	// For score updates, we need to find the comment even if it's deleted
	comment, err := dao.FindCommentForScoreUpdate(ctx, commentID)
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
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_deleted").IsNull(),
			psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
		))),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by pseudonym: %w", err)
	}

	return count, nil
}

// CountCommentsByPseudonymInSubforum counts total comments by a pseudonym in a specific subforum
func (dao *CommentDAO) CountCommentsByPseudonymInSubforum(ctx context.Context, pseudonymID string, subforumID int32) (int64, error) {
	// Get all comments by the pseudonym
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PseudonymID.EQ(pseudonymID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_deleted").IsNull(),
			psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
		))),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments by pseudonym: %w", err)
	}

	// Count comments in the specific subforum
	count := int64(0)
	for _, comment := range comments {
		post, err := models.Posts.Query(
			models.SelectWhere.Posts.PostID.EQ(comment.PostID),
			sm.Where(psql.Group(psql.And(
				psql.Group(psql.Or(
					psql.Quote("posts", "is_removed").IsNull(),
					psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
				)),
				psql.Group(psql.Or(
					psql.Quote("posts", "is_deleted").IsNull(),
					psql.Quote("posts", "is_deleted").EQ(psql.Arg(false)),
				)),
			))),
		).One(ctx, dao.db)
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
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_deleted").IsNull(),
			psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
		))),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by pseudonym: %w", err)
	}

	// Extract unique subforum IDs
	subforumMap := make(map[int32]bool)
	for _, comment := range comments {
		post, err := models.Posts.Query(
			models.SelectWhere.Posts.PostID.EQ(comment.PostID),
			sm.Where(psql.Group(psql.And(
				psql.Group(psql.Or(
					psql.Quote("posts", "is_removed").IsNull(),
					psql.Quote("posts", "is_removed").EQ(psql.Arg(false)),
				)),
				psql.Group(psql.Or(
					psql.Quote("posts", "is_deleted").IsNull(),
					psql.Quote("posts", "is_deleted").EQ(psql.Arg(false)),
				)),
			))),
		).One(ctx, dao.db)
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

// DeleteCommentByUser marks a comment as deleted by the user
func (dao *CommentDAO) DeleteCommentByUser(ctx context.Context, commentID int64, reason string) error {
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	reasonNull := sql.Null[string]{Valid: false}
	if reason != "" {
		reasonNull.Scan(reason)
	}

	updates := &models.CommentSetter{
		IsDeleted:                &sql.Null[bool]{Valid: true, V: true},
		DeletedByPseudonymID:     &sql.Null[string]{Valid: true, V: ""}, // This will be set by the caller
		DeletedByPseudonymAt:     &now,
		DeletedByPseudonymReason: &reasonNull,
		UpdatedAt:                &now,
	}

	comment, err := models.Comments.Query(
		models.SelectWhere.Comments.CommentID.EQ(commentID),
		sm.Where(psql.Group(psql.And(
			psql.Group(psql.Or(
				psql.Quote("comments", "is_removed").IsNull(),
				psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
			)),
			psql.Group(psql.Or(
				psql.Quote("comments", "is_deleted").IsNull(),
				psql.Quote("comments", "is_deleted").EQ(psql.Arg(false)),
			)),
		))),
	).One(ctx, dao.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("comment not found")
		}
		return fmt.Errorf("failed to find comment for user deletion: %w", err)
	}

	err = comment.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to mark comment as deleted by user: %w", err)
	}

	return nil
}

// GetCommentsByPostWithDeleted retrieves comments for a post including user-deleted ones (for display)
func (dao *CommentDAO) GetCommentsByPostWithDeleted(ctx context.Context, postID int64) ([]*models.Comment, error) {
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.EQ(postID),
		sm.Where(psql.Group(psql.Or(
			psql.Quote("comments", "is_removed").IsNull(),
			psql.Quote("comments", "is_removed").EQ(psql.Arg(false)),
		))),
		sm.OrderBy("score DESC NULLS LAST, created_at ASC"),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post with deleted: %w", err)
	}

	// Load related data for all comments
	for _, comment := range comments {
		if err := comment.LoadPseudonym(ctx, dao.db); err != nil {
			log.Warn().Err(err).Int64("comment_id", comment.CommentID).Msg("Failed to load comment pseudonym")
		}

		// Load deleted by pseudonym if comment is deleted
		if comment.IsDeleted.Valid && comment.IsDeleted.V && comment.DeletedByPseudonymID.Valid {
			if err := comment.LoadDeletedByPseudonymPseudonym(ctx, dao.db); err != nil {
				log.Warn().Err(err).Int64("comment_id", comment.CommentID).Msg("Failed to load deleted by pseudonym")
			}
		}
	}

	return comments, nil
}
