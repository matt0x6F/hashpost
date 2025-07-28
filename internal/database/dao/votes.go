package dao

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// VoteDAO provides data access operations for votes
type VoteDAO struct {
	db bob.Executor
}

// NewVoteDAO creates a new VoteDAO
func NewVoteDAO(db bob.Executor) *VoteDAO {
	return &VoteDAO{
		db: db,
	}
}

// GetVoteByPseudonymAndContent retrieves a vote by pseudonym and content
func (dao *VoteDAO) GetVoteByPseudonymAndContent(ctx context.Context, pseudonymID, contentType string, contentID int64) (*models.Vote, error) {
	log.Debug().
		Str("pseudonym_id", pseudonymID).
		Str("content_type", contentType).
		Int64("content_id", contentID).
		Msg("Getting vote by pseudonym and content")

	// Use the generated Votes table helper with where clause
	votes, err := models.Votes.Query(
		models.SelectWhere.Votes.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.Votes.ContentType.EQ(contentType),
		models.SelectWhere.Votes.ContentID.EQ(contentID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote by pseudonym and content: %w", err)
	}

	if len(votes) == 0 {
		return nil, nil
	}

	return votes[0], nil
}

// CreateVote creates a new vote
func (dao *VoteDAO) CreateVote(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) (*models.Vote, error) {
	log.Debug().
		Str("pseudonym_id", pseudonymID).
		Str("content_type", contentType).
		Int64("content_id", contentID).
		Int32("vote_value", voteValue).
		Msg("Creating vote")

	// Create a null time for now
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	voteSetter := &models.VoteSetter{
		PseudonymID: &pseudonymID,
		ContentType: &contentType,
		ContentID:   &contentID,
		VoteValue:   &voteValue,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// Use the generated Votes table helper
	vote, err := models.Votes.Insert(voteSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create vote: %w", err)
	}

	return vote, nil
}

// UpdateVote updates an existing vote
func (dao *VoteDAO) UpdateVote(ctx context.Context, voteID int64, voteValue int32) (*models.Vote, error) {
	log.Debug().
		Int64("vote_id", voteID).
		Int32("vote_value", voteValue).
		Msg("Updating vote")

	// Create a null time for now
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	updates := &models.VoteSetter{
		VoteValue: &voteValue,
		UpdatedAt: &now,
	}

	// Use the generated Votes table helper
	vote, err := models.Votes.Update(updates.UpdateMod(), models.UpdateWhere.Votes.VoteID.EQ(voteID)).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to update vote: %w", err)
	}

	return vote, nil
}

// DeleteVote deletes a vote
func (dao *VoteDAO) DeleteVote(ctx context.Context, voteID int64) error {
	log.Debug().
		Int64("vote_id", voteID).
		Msg("Deleting vote")

	// Use the generated Votes table helper
	_, err := models.Votes.Delete(models.DeleteWhere.Votes.VoteID.EQ(voteID)).Exec(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}

	return nil
}

// UpsertVote creates or updates a vote
func (dao *VoteDAO) UpsertVote(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) (*models.Vote, error) {
	log.Debug().
		Str("pseudonym_id", pseudonymID).
		Str("content_type", contentType).
		Int64("content_id", contentID).
		Int32("vote_value", voteValue).
		Msg("Upserting vote")

	// First try to get existing vote
	existingVote, err := dao.GetVoteByPseudonymAndContent(ctx, pseudonymID, contentType, contentID)
	if err != nil {
		return nil, err
	}

	if existingVote == nil {
		// Create new vote
		return dao.CreateVote(ctx, pseudonymID, contentType, contentID, voteValue)
	}

	// Update existing vote
	return dao.UpdateVote(ctx, existingVote.VoteID, voteValue)
}

// GetVotesByContent retrieves all votes for a specific content item
func (dao *VoteDAO) GetVotesByContent(ctx context.Context, contentType string, contentID int64) ([]*models.Vote, error) {
	log.Debug().
		Str("content_type", contentType).
		Int64("content_id", contentID).
		Msg("Getting votes by content")

	// Use the generated Votes table helper with where clause
	votes, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ(contentType),
		models.SelectWhere.Votes.ContentID.EQ(contentID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes by content: %w", err)
	}

	return votes, nil
}

// CountVotesByContent counts votes for a specific content item
func (dao *VoteDAO) CountVotesByContent(ctx context.Context, contentType string, contentID int64) (int, error) {
	log.Debug().
		Str("content_type", contentType).
		Int64("content_id", contentID).
		Msg("Counting votes by content")

	// Use the generated Votes table helper with where clause and count
	count, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ(contentType),
		models.SelectWhere.Votes.ContentID.EQ(contentID),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count votes by content: %w", err)
	}

	return int(count), nil
}

// GetVoteSummaryByContent gets vote summary (upvotes, downvotes, total) for content
func (dao *VoteDAO) GetVoteSummaryByContent(ctx context.Context, contentType string, contentID int64) (upvotes, downvotes, total int, err error) {
	log.Debug().
		Str("content_type", contentType).
		Int64("content_id", contentID).
		Msg("Getting vote summary by content")

	// Get all votes for the content
	votes, err := dao.GetVotesByContent(ctx, contentType, contentID)
	if err != nil {
		return 0, 0, 0, err
	}

	// Calculate summary
	for _, vote := range votes {
		total++
		switch vote.VoteValue {
		case 1:
			upvotes++
		case -1:
			downvotes++
		}
	}

	return upvotes, downvotes, total, nil
}

// GetVotesCount returns the total count of votes in a subforum since a given time
func (dao *VoteDAO) GetVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Time("since", since).
		Msg("Getting votes count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Count votes on posts since the given time
	postVotesCount, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("post"),
		models.SelectWhere.Votes.ContentID.In(postIDs...),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count post votes: %w", err)
	}

	// Get comment IDs in the subforum
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.In(postIDs...),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments in subforum: %w", err)
	}

	if len(comments) == 0 {
		return int(postVotesCount), nil
	}

	// Extract comment IDs
	commentIDs := make([]int64, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.CommentID
	}

	// Count votes on comments since the given time
	commentVotesCount, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("comment"),
		models.SelectWhere.Votes.ContentID.In(commentIDs...),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comment votes: %w", err)
	}

	return int(postVotesCount + commentVotesCount), nil
}

// GetPostVotesCount returns the count of votes on posts in a subforum since a given time
func (dao *VoteDAO) GetPostVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Time("since", since).
		Msg("Getting post votes count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Count votes on posts since the given time
	count, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("post"),
		models.SelectWhere.Votes.ContentID.In(postIDs...),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count post votes: %w", err)
	}

	return int(count), nil
}

// GetCommentVotesCount returns the count of votes on comments in a subforum since a given time
func (dao *VoteDAO) GetCommentVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Time("since", since).
		Msg("Getting comment votes count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Get comment IDs in the subforum
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.In(postIDs...),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments in subforum: %w", err)
	}

	if len(comments) == 0 {
		return 0, nil
	}

	// Extract comment IDs
	commentIDs := make([]int64, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.CommentID
	}

	// Count votes on comments since the given time
	count, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("comment"),
		models.SelectWhere.Votes.ContentID.In(commentIDs...),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comment votes: %w", err)
	}

	return int(count), nil
}

// GetPostUpvotesCount returns the count of upvotes on posts in a subforum since a given time
func (dao *VoteDAO) GetPostUpvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Time("since", since).
		Msg("Getting post upvotes count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Count upvotes on posts since the given time
	count, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("post"),
		models.SelectWhere.Votes.ContentID.In(postIDs...),
		models.SelectWhere.Votes.VoteValue.EQ(1),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count post upvotes: %w", err)
	}

	return int(count), nil
}

// GetPostDownvotesCount returns the count of downvotes on posts in a subforum since a given time
func (dao *VoteDAO) GetPostDownvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Time("since", since).
		Msg("Getting post downvotes count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Count downvotes on posts since the given time
	count, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("post"),
		models.SelectWhere.Votes.ContentID.In(postIDs...),
		models.SelectWhere.Votes.VoteValue.EQ(-1),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count post downvotes: %w", err)
	}

	return int(count), nil
}

// GetCommentUpvotesCount returns the count of upvotes on comments in a subforum since a given time
func (dao *VoteDAO) GetCommentUpvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Time("since", since).
		Msg("Getting comment upvotes count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Get comment IDs in the subforum
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.In(postIDs...),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments in subforum: %w", err)
	}

	if len(comments) == 0 {
		return 0, nil
	}

	// Extract comment IDs
	commentIDs := make([]int64, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.CommentID
	}

	// Count upvotes on comments since the given time
	count, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("comment"),
		models.SelectWhere.Votes.ContentID.In(commentIDs...),
		models.SelectWhere.Votes.VoteValue.EQ(1),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comment upvotes: %w", err)
	}

	return int(count), nil
}

// GetCommentDownvotesCount returns the count of downvotes on comments in a subforum since a given time
func (dao *VoteDAO) GetCommentDownvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Time("since", since).
		Msg("Getting comment downvotes count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Get comment IDs in the subforum
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.In(postIDs...),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments in subforum: %w", err)
	}

	if len(comments) == 0 {
		return 0, nil
	}

	// Extract comment IDs
	commentIDs := make([]int64, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.CommentID
	}

	// Count downvotes on comments since the given time
	count, err := models.Votes.Query(
		models.SelectWhere.Votes.ContentType.EQ("comment"),
		models.SelectWhere.Votes.ContentID.In(commentIDs...),
		models.SelectWhere.Votes.VoteValue.EQ(-1),
		models.SelectWhere.Votes.CreatedAt.GTE(since),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comment downvotes: %w", err)
	}

	return int(count), nil
}
