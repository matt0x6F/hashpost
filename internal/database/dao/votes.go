package dao

import (
	"context"
	"database/sql"
	"fmt"
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
		if vote.VoteValue == 1 {
			upvotes++
		} else if vote.VoteValue == -1 {
			downvotes++
		}
	}

	return upvotes, downvotes, total, nil
}
