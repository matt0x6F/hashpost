package mocks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockVoteDAO is a mock implementation of VoteDAOInterface with data injection support
type MockVoteDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	votes          map[string]*models.Vote                            // key: "pseudonymID-contentType-contentID"
	votesByContent map[string][]*models.Vote                          // key: "contentType-contentID"
	counts         map[string]int                                     // key: "contentType-contentID"
	summaries      map[string]struct{ upvotes, downvotes, total int } // key: "contentType-contentID"
}

// NewMockVoteDAO creates a new mock VoteDAO with optional initial data
func NewMockVoteDAO() *MockVoteDAO {
	return &MockVoteDAO{
		votes:          make(map[string]*models.Vote),
		votesByContent: make(map[string][]*models.Vote),
		counts:         make(map[string]int),
		summaries:      make(map[string]struct{ upvotes, downvotes, total int }),
	}
}

// InjectVote injects a vote into the mock for testing
func (m *MockVoteDAO) InjectVote(pseudonymID, contentType string, contentID int64, vote *models.Vote) {
	key := fmt.Sprintf("%s-%s-%d", pseudonymID, contentType, contentID)
	m.votes[key] = vote
}

// InjectVotesByContent injects votes that should be returned when querying by content
func (m *MockVoteDAO) InjectVotesByContent(contentType string, contentID int64, votes []*models.Vote) {
	key := fmt.Sprintf("%s-%d", contentType, contentID)
	m.votesByContent[key] = votes
}

// InjectCount injects a count that should be returned for count operations
func (m *MockVoteDAO) InjectCount(contentType string, contentID int64, count int) {
	key := fmt.Sprintf("%s-%d", contentType, contentID)
	m.counts[key] = count
}

// InjectSummary injects a vote summary that should be returned for summary operations
func (m *MockVoteDAO) InjectSummary(contentType string, contentID int64, upvotes, downvotes, total int) {
	key := fmt.Sprintf("%s-%d", contentType, contentID)
	m.summaries[key] = struct{ upvotes, downvotes, total int }{upvotes, downvotes, total}
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockVoteDAO) SetDefaultBehavior() {
	// Default behavior for GetVoteByPseudonymAndContent
	m.On("GetVoteByPseudonymAndContent", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, pseudonymID, contentType string, contentID int64) *models.Vote {
			key := fmt.Sprintf("%s-%s-%d", pseudonymID, contentType, contentID)
			return m.votes[key]
		},
		func(ctx context.Context, pseudonymID, contentType string, contentID int64) error {
			key := fmt.Sprintf("%s-%s-%d", pseudonymID, contentType, contentID)
			if m.votes[key] == nil {
				return sql.ErrNoRows
			}
			return nil
		},
	)

	// Default behavior for GetVotesByContent
	m.On("GetVotesByContent", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, contentType string, contentID int64) []*models.Vote {
			key := fmt.Sprintf("%s-%d", contentType, contentID)
			return m.votesByContent[key]
		},
		func(ctx context.Context, contentType string, contentID int64) error {
			return nil
		},
	)

	// Default behavior for CountVotesByContent
	m.On("CountVotesByContent", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, contentType string, contentID int64) int {
			key := fmt.Sprintf("%s-%d", contentType, contentID)
			return m.counts[key]
		},
		func(ctx context.Context, contentType string, contentID int64) error {
			return nil
		},
	)

	// Default behavior for GetVoteSummaryByContent
	m.On("GetVoteSummaryByContent", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, contentType string, contentID int64) (int, int, int, error) {
			key := fmt.Sprintf("%s-%d", contentType, contentID)
			if summary, exists := m.summaries[key]; exists {
				return summary.upvotes, summary.downvotes, summary.total, nil
			}
			return 0, 0, 0, nil
		},
	)

	// Default behavior for UpsertVote
	m.On("UpsertVote", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) *models.Vote {
			key := fmt.Sprintf("%s-%s-%d", pseudonymID, contentType, contentID)
			if vote, exists := m.votes[key]; exists {
				vote.VoteValue = voteValue
				vote.UpdatedAt = sql.Null[time.Time]{V: time.Now(), Valid: true}
				return vote
			}
			// Create new vote
			newVote := &models.Vote{
				PseudonymID: pseudonymID,
				ContentType: contentType,
				ContentID:   contentID,
				VoteValue:   voteValue,
				CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
				UpdatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
			}
			m.votes[key] = newVote
			return newVote
		},
		func(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) error {
			return nil
		},
	)

	// Default behavior for moderation dashboard methods
	m.On("GetVotesCount", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(0, nil)
	m.On("GetPostVotesCount", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(0, nil)
	m.On("GetCommentVotesCount", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(0, nil)
	m.On("GetPostUpvotesCount", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(0, nil)
	m.On("GetPostDownvotesCount", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(0, nil)
	m.On("GetCommentUpvotesCount", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(0, nil)
	m.On("GetCommentDownvotesCount", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(0, nil)
}

// CreateVote creates a new vote
func (m *MockVoteDAO) CreateVote(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) (*models.Vote, error) {
	args := m.Called(ctx, pseudonymID, contentType, contentID, voteValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vote), args.Error(1)
}

// GetVoteByPseudonymAndContent retrieves a vote by pseudonym and content
func (m *MockVoteDAO) GetVoteByPseudonymAndContent(ctx context.Context, pseudonymID, contentType string, contentID int64) (*models.Vote, error) {
	args := m.Called(ctx, pseudonymID, contentType, contentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vote), args.Error(1)
}

// UpdateVote updates a vote
func (m *MockVoteDAO) UpdateVote(ctx context.Context, voteID int64, voteValue int32) (*models.Vote, error) {
	args := m.Called(ctx, voteID, voteValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vote), args.Error(1)
}

// DeleteVote deletes a vote
func (m *MockVoteDAO) DeleteVote(ctx context.Context, voteID int64) error {
	args := m.Called(ctx, voteID)
	return args.Error(0)
}

// UpsertVote upserts a vote
func (m *MockVoteDAO) UpsertVote(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) (*models.Vote, error) {
	args := m.Called(ctx, pseudonymID, contentType, contentID, voteValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vote), args.Error(1)
}

// GetVotesByContent retrieves votes by content
func (m *MockVoteDAO) GetVotesByContent(ctx context.Context, contentType string, contentID int64) ([]*models.Vote, error) {
	args := m.Called(ctx, contentType, contentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Vote), args.Error(1)
}

// CountVotesByContent counts votes by content
func (m *MockVoteDAO) CountVotesByContent(ctx context.Context, contentType string, contentID int64) (int, error) {
	args := m.Called(ctx, contentType, contentID)
	return args.Get(0).(int), args.Error(1)
}

// GetVoteSummaryByContent returns the vote summary for a specific content item
func (m *MockVoteDAO) GetVoteSummaryByContent(ctx context.Context, contentType string, contentID int64) (upvotes, downvotes, total int, err error) {
	args := m.Called(ctx, contentType, contentID)
	return args.Get(0).(int), args.Get(1).(int), args.Get(2).(int), args.Error(3)
}

// GetVotesCount returns the total count of votes in a subforum since a given time
func (m *MockVoteDAO) GetVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}

// GetPostVotesCount returns the count of votes on posts in a subforum since a given time
func (m *MockVoteDAO) GetPostVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}

// GetCommentVotesCount returns the count of votes on comments in a subforum since a given time
func (m *MockVoteDAO) GetCommentVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}

// GetPostUpvotesCount returns the count of upvotes on posts in a subforum since a given time
func (m *MockVoteDAO) GetPostUpvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}

// GetPostDownvotesCount returns the count of downvotes on posts in a subforum since a given time
func (m *MockVoteDAO) GetPostDownvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}

// GetCommentUpvotesCount returns the count of upvotes on comments in a subforum since a given time
func (m *MockVoteDAO) GetCommentUpvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}

// GetCommentDownvotesCount returns the count of downvotes on comments in a subforum since a given time
func (m *MockVoteDAO) GetCommentDownvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}
