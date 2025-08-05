package mocks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockCommentDAO is a mock implementation of CommentDAOInterface with data injection support
type MockCommentDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	comments       map[int64]*models.Comment
	commentsByPost map[int64][]*models.Comment
	counts         map[string]int64 // key: "postID" or "pseudonymID"
}

// NewMockCommentDAO creates a new mock CommentDAO with optional initial data
func NewMockCommentDAO() *MockCommentDAO {
	return &MockCommentDAO{
		comments:       make(map[int64]*models.Comment),
		commentsByPost: make(map[int64][]*models.Comment),
		counts:         make(map[string]int64),
	}
}

// InjectComment injects a comment into the mock for testing
func (m *MockCommentDAO) InjectComment(comment *models.Comment) {
	m.comments[comment.CommentID] = comment
}

// InjectCommentsByPost injects comments that should be returned when querying by post
func (m *MockCommentDAO) InjectCommentsByPost(postID int64, comments []*models.Comment) {
	m.commentsByPost[postID] = comments
}

// InjectCount injects a count that should be returned for count operations
func (m *MockCommentDAO) InjectCount(key string, count int64) {
	m.counts[key] = count
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockCommentDAO) SetDefaultBehavior() {
	// Default behavior for GetCommentByID
	m.On("GetCommentByID", mock.Anything, mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, commentID int64) (*models.Comment, error) {
			if comment, exists := m.comments[commentID]; exists {
				return comment, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for GetCommentsByPost
	m.On("GetCommentsByPost", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, postID int64, limit, offset int, sort string) ([]*models.Comment, error) {
			if comments, exists := m.commentsByPost[postID]; exists {
				return comments, nil
			}
			return []*models.Comment{}, nil
		},
	)

	// Default behavior for CountCommentsByPost
	m.On("CountCommentsByPost", mock.Anything, mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, postID int64) (int64, error) {
			count := m.counts[fmt.Sprintf("post_%d", postID)]
			return count, nil
		},
	)

	// Default behavior for CountCommentsByPseudonym
	m.On("CountCommentsByPseudonym", mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, pseudonymID string) (int64, error) {
			count := m.counts[fmt.Sprintf("pseudonym_%s", pseudonymID)]
			return count, nil
		},
	)

	// Default behavior for SetRemoved
	m.On("SetRemoved", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("bool")).Return(
		func(ctx context.Context, commentID int64, removed bool) error {
			if comment, exists := m.comments[commentID]; exists {
				comment.IsRemoved = sql.Null[bool]{V: removed, Valid: true}
				return nil
			}
			return sql.ErrNoRows
		},
	)
}

// CreateComment creates a new comment
func (m *MockCommentDAO) CreateComment(ctx context.Context, postID int64, pseudonymID, content string, parentCommentID *int64) (*models.Comment, error) {
	args := m.Called(ctx, postID, pseudonymID, content, parentCommentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

// GetCommentByID retrieves a comment by ID
func (m *MockCommentDAO) GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error) {
	args := m.Called(ctx, commentID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64) (*models.Comment, error)); ok {
		return fn(ctx, commentID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

// GetCommentsByPost retrieves comments by post
func (m *MockCommentDAO) GetCommentsByPost(ctx context.Context, postID int64) ([]*models.Comment, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Comment), args.Error(1)
}

// GetCommentsByPostWithNestedReplies retrieves comments by post with nested replies
func (m *MockCommentDAO) GetCommentsByPostWithNestedReplies(ctx context.Context, postID int64) ([]*models.Comment, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Comment), args.Error(1)
}

// CountCommentsByPost counts comments by post
func (m *MockCommentDAO) CountCommentsByPost(ctx context.Context, postID int64) (int64, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(int64), args.Error(1)
}

// CountCommentsByPseudonym counts comments by pseudonym
func (m *MockCommentDAO) CountCommentsByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	args := m.Called(ctx, pseudonymID)
	return args.Get(0).(int64), args.Error(1)
}

// UpdateCommentScore updates comment score
func (m *MockCommentDAO) UpdateCommentScore(ctx context.Context, commentID int64, score, upvotes, downvotes int32) error {
	args := m.Called(ctx, commentID, score, upvotes, downvotes)
	return args.Error(0)
}

// MarkCommentAsDeletedByPseudonym marks a comment as deleted by pseudonym
func (m *MockCommentDAO) MarkCommentAsDeletedByPseudonym(ctx context.Context, commentID int64, pseudonymID string, reason string) error {
	args := m.Called(ctx, commentID, pseudonymID, reason)
	return args.Error(0)
}

// DeleteCommentByUser deletes a comment by user
func (m *MockCommentDAO) DeleteCommentByUser(ctx context.Context, commentID int64, reason string) error {
	args := m.Called(ctx, commentID, reason)
	return args.Error(0)
}

// CountCommentsByPseudonymInSubforum counts comments by pseudonym in a specific subforum
func (m *MockCommentDAO) CountCommentsByPseudonymInSubforum(ctx context.Context, pseudonymID string, subforumID int32) (int64, error) {
	args := m.Called(ctx, pseudonymID, subforumID)
	return args.Get(0).(int64), args.Error(1)
}

// GetSubforumsByPseudonymComments gets subforums where a pseudonym has commented
func (m *MockCommentDAO) GetSubforumsByPseudonymComments(ctx context.Context, pseudonymID string) ([]int32, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return []int32{}, args.Error(1)
	}
	return args.Get(0).([]int32), args.Error(1)
}

// SetRemoved sets the removed status of a comment
func (m *MockCommentDAO) SetRemoved(ctx context.Context, commentID int64, removed bool) error {
	args := m.Called(ctx, commentID, removed)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, bool) error); ok {
		return fn(ctx, commentID, removed)
	}

	// Fallback to direct return values
	return args.Error(0)
}

// UpdateComment updates a comment's content and edit metadata
func (m *MockCommentDAO) UpdateComment(ctx context.Context, commentID int64, content, editReason string) error {
	args := m.Called(ctx, commentID, content, editReason)
	return args.Error(0)
}

// SetCommentRemoved sets the removal status and metadata for a comment
func (m *MockCommentDAO) SetCommentRemoved(ctx context.Context, commentID int64, removed bool, reason, removedByPseudonymID string) error {
	args := m.Called(ctx, commentID, removed, reason, removedByPseudonymID)
	return args.Error(0)
}

// GetCommentsCount returns the count of comments in a subforum since a given time
func (m *MockCommentDAO) GetCommentsCount(ctx context.Context, subforumPath string, since time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, since)
	return args.Get(0).(int), args.Error(1)
}

// GetCommentsCountForDateRange returns the count of comments created within a specific date range
func (m *MockCommentDAO) GetCommentsCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error) {
	args := m.Called(ctx, subforumPath, startTime, endTime)
	return args.Get(0).(int), args.Error(1)
}

// GetTotalCommentsCount returns the total count of comments in a subforum
func (m *MockCommentDAO) GetTotalCommentsCount(ctx context.Context, subforumPath string) (int, error) {
	args := m.Called(ctx, subforumPath)
	return args.Get(0).(int), args.Error(1)
}

// GetCommentsByPseudonym retrieves comments by pseudonym with pagination and sorting
func (m *MockCommentDAO) GetCommentsByPseudonym(ctx context.Context, pseudonymID string, page, limit int, sortField string, sortDesc bool) ([]*models.Comment, error) {
	args := m.Called(ctx, pseudonymID, page, limit, sortField, sortDesc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Comment), args.Error(1)
}
