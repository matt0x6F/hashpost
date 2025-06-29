package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentDAO_GetCommentsByPostWithNestedReplies(t *testing.T) {
	// This test requires a database connection
	// For now, we'll test the tree building logic directly

	dao := &CommentDAO{}

	// Create test comments with nested structure
	now := time.Now()

	// Root comment 1 (score: 10)
	root1 := &models.Comment{
		CommentID: 1,
		PostID:    1,
		Content:   "Root comment 1",
		Score:     sql.Null[int32]{Valid: true, V: 10},
		CreatedAt: sql.Null[time.Time]{Valid: true, V: now},
	}

	// Root comment 2 (score: 5)
	root2 := &models.Comment{
		CommentID: 2,
		PostID:    1,
		Content:   "Root comment 2",
		Score:     sql.Null[int32]{Valid: true, V: 5},
		CreatedAt: sql.Null[time.Time]{Valid: true, V: now.Add(time.Minute)},
	}

	// Reply to root1 (score: 8)
	reply1 := &models.Comment{
		CommentID:       3,
		PostID:          1,
		ParentCommentID: sql.Null[int64]{Valid: true, V: 1},
		Content:         "Reply to root 1",
		Score:           sql.Null[int32]{Valid: true, V: 8},
		CreatedAt:       sql.Null[time.Time]{Valid: true, V: now.Add(2 * time.Minute)},
	}

	// Reply to reply1 (score: 3)
	reply2 := &models.Comment{
		CommentID:       4,
		PostID:          1,
		ParentCommentID: sql.Null[int64]{Valid: true, V: 3},
		Content:         "Reply to reply 1",
		Score:           sql.Null[int32]{Valid: true, V: 3},
		CreatedAt:       sql.Null[time.Time]{Valid: true, V: now.Add(3 * time.Minute)},
	}

	// Reply to root2 (score: 2)
	reply3 := &models.Comment{
		CommentID:       5,
		PostID:          1,
		ParentCommentID: sql.Null[int64]{Valid: true, V: 2},
		Content:         "Reply to root 2",
		Score:           sql.Null[int32]{Valid: true, V: 2},
		CreatedAt:       sql.Null[time.Time]{Valid: true, V: now.Add(4 * time.Minute)},
	}

	allComments := []*models.Comment{root1, root2, reply1, reply2, reply3}

	// Test the tree building logic
	result := dao.buildNestedCommentTree(allComments)

	// Verify the structure
	require.Len(t, result, 2, "Should have 2 root comments")

	// Root comments should be ordered by score (descending)
	assert.Equal(t, int64(1), result[0].CommentID, "First root comment should be root1 (higher score)")
	assert.Equal(t, int64(2), result[1].CommentID, "Second root comment should be root2 (lower score)")

	// Check that root1 has reply1 as a child
	require.Len(t, result[0].R.ReverseComments, 1, "Root1 should have 1 reply")
	assert.Equal(t, int64(3), result[0].R.ReverseComments[0].CommentID, "Root1's reply should be reply1")

	// Check that reply1 has reply2 as a child
	require.Len(t, result[0].R.ReverseComments[0].R.ReverseComments, 1, "Reply1 should have 1 reply")
	assert.Equal(t, int64(4), result[0].R.ReverseComments[0].R.ReverseComments[0].CommentID, "Reply1's reply should be reply2")

	// Check that root2 has reply3 as a child
	require.Len(t, result[1].R.ReverseComments, 1, "Root2 should have 1 reply")
	assert.Equal(t, int64(5), result[1].R.ReverseComments[0].CommentID, "Root2's reply should be reply3")
}

func TestCommentDAO_GetCommentsByPost(t *testing.T) {
	// Test that the original method still works and includes ordering
	dao := &CommentDAO{}

	// This test would require a real database connection
	// For now, we'll just verify the method signature and basic logic

	// Test with a mock executor (this won't actually work, but tests the interface)
	var mockDB bob.Executor
	dao.db = mockDB

	// The method should not panic even with a nil executor
	// (it will return an error, but that's expected)
	// We'll skip this test since it requires a real database connection
	t.Skip("Skipping test that requires real database connection")
}
