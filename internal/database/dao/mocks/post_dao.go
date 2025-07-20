package mocks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockPostDAO is a mock implementation of PostDAOInterface with data injection support
type MockPostDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	posts           map[int64]*models.Post
	postsBySlug     map[string]*models.Post // key: "subforumID-slug"
	postsBySubforum map[int32][]*models.Post
	counts          map[string]int64 // key: "subforumID" or "pseudonymID"
}

// NewMockPostDAO creates a new mock PostDAO with optional initial data
func NewMockPostDAO() *MockPostDAO {
	return &MockPostDAO{
		posts:           make(map[int64]*models.Post),
		postsBySlug:     make(map[string]*models.Post),
		postsBySubforum: make(map[int32][]*models.Post),
		counts:          make(map[string]int64),
	}
}

// InjectPost injects a post into the mock for testing
func (m *MockPostDAO) InjectPost(post *models.Post) {
	m.posts[post.PostID] = post
}

// InjectPostBySlug injects a post that should be returned when querying by slug
func (m *MockPostDAO) InjectPostBySlug(subforumID int32, slug string, post *models.Post) {
	key := fmt.Sprintf("%d-%s", subforumID, slug)
	m.postsBySlug[key] = post
}

// InjectPostsBySubforum injects posts that should be returned when querying by subforum
func (m *MockPostDAO) InjectPostsBySubforum(subforumID int32, posts []*models.Post) {
	m.postsBySubforum[subforumID] = posts
}

// InjectCount injects a count that should be returned for count operations
func (m *MockPostDAO) InjectCount(key string, count int64) {
	m.counts[key] = count
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockPostDAO) SetDefaultBehavior() {
	// Default behavior for GetPostByID
	m.On("GetPostByID", mock.Anything, mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, postID int64) (*models.Post, error) {
			if post, exists := m.posts[postID]; exists {
				return post, nil
			}
			return nil, sql.ErrNoRows
		},
	).Maybe()

	// Default behavior for GetPostBySubforumAndSlug
	m.On("GetPostBySubforumAndSlug", mock.Anything, mock.AnythingOfType("int32"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, subforumID int32, slug string) (*models.Post, error) {
			key := fmt.Sprintf("%d-%s", subforumID, slug)
			if post, exists := m.postsBySlug[key]; exists {
				return post, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for ListPostsBySubforum
	m.On("ListPostsBySubforum", mock.Anything, mock.AnythingOfType("int32"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, subforumID int32, limit, offset int, sort, timeFilter string) ([]*models.Post, error) {
			if posts, exists := m.postsBySubforum[subforumID]; exists {
				return posts, nil
			}
			return []*models.Post{}, nil
		},
	)

	// Default behavior for CountPostsBySubforum
	m.On("CountPostsBySubforum", mock.Anything, mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, subforumID int32) (int64, error) {
			count := m.counts[fmt.Sprintf("subforum_%d", subforumID)]
			return count, nil
		},
	)

	// Default behavior for CountPostsByPseudonym
	m.On("CountPostsByPseudonym", mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, pseudonymID string) (int64, error) {
			count := m.counts[fmt.Sprintf("pseudonym_%s", pseudonymID)]
			return count, nil
		},
	)

	// Default behavior for SetRemoved
	m.On("SetRemoved", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("bool")).Return(
		func(ctx context.Context, postID int64, removed bool) error {
			if post, exists := m.posts[postID]; exists {
				post.IsRemoved = sql.Null[bool]{V: removed, Valid: true}
				return nil
			}
			return sql.ErrNoRows
		},
	)

	// Default behavior for MarkPostAsDeletedByPseudonym
	m.On("MarkPostAsDeletedByPseudonym", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, postID int64, pseudonymID string, reason string) error {
			if post, exists := m.posts[postID]; exists {
				post.IsDeleted = sql.Null[bool]{V: true, Valid: true}
				post.DeletedByPseudonymID = sql.Null[string]{V: pseudonymID, Valid: true}
				post.DeletedByPseudonymAt = sql.Null[time.Time]{V: time.Now(), Valid: true}
				post.DeletedByPseudonymReason = sql.Null[string]{V: reason, Valid: true}
				return nil
			}
			return sql.ErrNoRows
		},
	).Maybe()
}

// CreatePost creates a new post
func (m *MockPostDAO) CreatePost(ctx context.Context, subforumID int32, pseudonymID, title, content, postType string, url *string, isNSFW, isSpoiler bool) (*models.Post, error) {
	args := m.Called(ctx, subforumID, pseudonymID, title, content, postType, url, isNSFW, isSpoiler)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

// GetPostByID retrieves a post by ID
func (m *MockPostDAO) GetPostByID(ctx context.Context, postID int64) (*models.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, sql.ErrNoRows
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64) (*models.Post, error)); ok {
		return fn(ctx, postID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*models.Post), args.Error(0)
}

// GetPostsBySubforum retrieves posts by subforum
func (m *MockPostDAO) GetPostsBySubforum(ctx context.Context, subforumID int32, page, limit int, sortField string, sortDesc bool) ([]*models.Post, error) {
	args := m.Called(ctx, subforumID, page, limit, sortField, sortDesc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

// CountPostsBySubforum counts posts by subforum
func (m *MockPostDAO) CountPostsBySubforum(ctx context.Context, subforumID int32) (int64, error) {
	args := m.Called(ctx, subforumID)
	return args.Get(0).(int64), args.Error(1)
}

// CountPostsByPseudonym counts posts by pseudonym
func (m *MockPostDAO) CountPostsByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	args := m.Called(ctx, pseudonymID)
	return args.Get(0).(int64), args.Error(1)
}

// GetPostBySubforumAndSlug retrieves a post by subforum and slug
func (m *MockPostDAO) GetPostBySubforumAndSlug(ctx context.Context, subforumID int32, slug string) (*models.Post, error) {
	args := m.Called(ctx, subforumID, slug)
	if args.Get(0) == nil {
		return nil, sql.ErrNoRows
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int32, string) (*models.Post, error)); ok {
		return fn(ctx, subforumID, slug)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(0)
	}
	return args.Get(0).(*models.Post), args.Error(0)
}

// UpdatePostScore updates post score
func (m *MockPostDAO) UpdatePostScore(ctx context.Context, postID int64, score, upvotes, downvotes int32) error {
	args := m.Called(ctx, postID, score, upvotes, downvotes)
	return args.Error(0)
}

// IncrementViewCount increments view count
func (m *MockPostDAO) IncrementViewCount(ctx context.Context, postID int64) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

// UpdateCommentCount updates comment count
func (m *MockPostDAO) UpdateCommentCount(ctx context.Context, postID int64, commentCount int32) error {
	args := m.Called(ctx, postID, commentCount)
	return args.Error(0)
}

// SetLocked sets post locked status
func (m *MockPostDAO) SetLocked(ctx context.Context, postID int64, locked bool) error {
	args := m.Called(ctx, postID, locked)
	return args.Error(0)
}

// SetSticky sets post sticky status
func (m *MockPostDAO) SetSticky(ctx context.Context, postID int64, sticky bool) error {
	args := m.Called(ctx, postID, sticky)
	return args.Error(0)
}

// SetRemoved sets post removed status
func (m *MockPostDAO) SetRemoved(ctx context.Context, postID int64, removed bool) error {
	args := m.Called(ctx, postID, removed)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, bool) error); ok {
		return fn(ctx, postID, removed)
	}

	// Fallback to direct return values
	return args.Error(0)
}

// MarkPostAsDeletedByPseudonym marks a post as deleted by pseudonym
func (m *MockPostDAO) MarkPostAsDeletedByPseudonym(ctx context.Context, postID int64, pseudonymID string, reason string) error {
	args := m.Called(ctx, postID, pseudonymID, reason)
	if args.Get(0) == nil {
		return nil
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, string, string) error); ok {
		return fn(ctx, postID, pseudonymID, reason)
	}

	// Fallback to direct return values
	return args.Error(0)
}

// CountPostsByPseudonymInSubforum counts posts by pseudonym in a specific subforum
func (m *MockPostDAO) CountPostsByPseudonymInSubforum(ctx context.Context, pseudonymID string, subforumID int32) (int64, error) {
	args := m.Called(ctx, pseudonymID, subforumID)
	return args.Get(0).(int64), args.Error(1)
}

// GetSubforumsByPseudonym gets subforums where a pseudonym has posted
func (m *MockPostDAO) GetSubforumsByPseudonym(ctx context.Context, pseudonymID string) ([]int32, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return []int32{}, args.Error(1)
	}
	return args.Get(0).([]int32), args.Error(1)
}
