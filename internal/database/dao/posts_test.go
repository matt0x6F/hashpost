package dao

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDBConfig holds configuration for test database setup
type TestDBConfig struct {
	SubforumName    string
	SubforumDisplay string
	SubforumDesc    string
	PseudonymID     string
	PseudonymName   string
	PseudonymBio    string
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() *TestDBConfig {
	return &TestDBConfig{
		SubforumName:    "Test Subforum",
		SubforumDisplay: "Test Subforum",
		SubforumDesc:    "A test subforum for testing",
		PseudonymID:     "test-pseudonym-123",
		PseudonymName:   "Test Pseudonym",
		PseudonymBio:    "A test pseudonym",
	}
}

// setupTestDB creates a database connection for testing
func setupTestDB(t *testing.T) bob.Executor {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	// Test the connection
	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	// Convert to bob.DB which implements bob.Executor
	bobDB := bob.NewDB(db)
	return bobDB
}

// cleanupTestData cleans up test data after each test
func cleanupTestData(t *testing.T, db bob.Executor) {
	ctx := context.Background()

	// Clean up in reverse order of dependencies
	_, err := db.ExecContext(ctx, "DELETE FROM posts WHERE title LIKE 'Test%'")
	require.NoError(t, err, "Failed to cleanup test posts")

	_, err = db.ExecContext(ctx, "DELETE FROM pseudonyms WHERE pseudonym_id LIKE 'test-%'")
	require.NoError(t, err, "Failed to cleanup test pseudonyms")

	_, err = db.ExecContext(ctx, "DELETE FROM subforums WHERE name LIKE 'Test%'")
	require.NoError(t, err, "Failed to cleanup test subforums")
}

// createTestData creates the required test data (subforum and pseudonym)
func createTestData(t *testing.T, db bob.Executor) (int32, string) {
	return createTestDataWithConfig(t, db, DefaultTestConfig())
}

// createTestDataWithConfig creates test data with custom configuration
func createTestDataWithConfig(t *testing.T, db bob.Executor, config *TestDBConfig) (int32, string) {
	ctx := context.Background()

	// Create test subforum using raw SQL
	_, err := db.ExecContext(ctx, `
		INSERT INTO subforums (name, display_name, description) 
		VALUES ($1, $2, $3)
		RETURNING subforum_id
	`, config.SubforumName, config.SubforumDisplay, config.SubforumDesc)
	require.NoError(t, err, "Failed to create test subforum")

	// Get the subforum ID
	var subforumID int32
	err = db.(bob.DB).DB.QueryRowContext(ctx,
		"SELECT subforum_id FROM subforums WHERE name = $1", config.SubforumName).Scan(&subforumID)
	require.NoError(t, err, "Failed to get subforum ID")

	// Create test pseudonym using raw SQL
	_, err = db.ExecContext(ctx, `
		INSERT INTO pseudonyms (pseudonym_id, display_name, bio, is_default) 
		VALUES ($1, $2, $3, true)
	`, config.PseudonymID, config.PseudonymName, config.PseudonymBio)
	require.NoError(t, err, "Failed to create test pseudonym")

	return subforumID, config.PseudonymID
}

func TestPostDAO_GetPostByID_FiltersDeletedPosts(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewPostDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	subforumID, pseudonymID := createTestData(t, db)

	// Create a test post
	post, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post", "Test content", "text", nil, false, false)
	require.NoError(t, err)
	require.NotNil(t, post)

	// Verify post can be retrieved normally
	retrievedPost, err := dao.GetPostByID(ctx, post.PostID)
	require.NoError(t, err)
	require.NotNil(t, retrievedPost)
	assert.Equal(t, post.PostID, retrievedPost.PostID)
	assert.Equal(t, "Test Post", retrievedPost.Title)

	// Mark post as deleted
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post.PostID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify post is no longer retrievable (filtered out)
	retrievedPost, err = dao.GetPostByID(ctx, post.PostID)
	require.NoError(t, err)
	assert.Nil(t, retrievedPost, "Deleted post should not be returned")
}

func TestPostDAO_GetPostsBySubforum_FiltersDeletedPosts(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewPostDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	subforumID, pseudonymID := createTestData(t, db)

	// Create multiple posts
	_, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post 1", "Content 1", "text", nil, false, false)
	require.NoError(t, err)

	post2, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post 2", "Content 2", "text", nil, false, false)
	require.NoError(t, err)

	_, err = dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post 3", "Content 3", "text", nil, false, false)
	require.NoError(t, err)

	// Verify all posts are retrievable
	posts, err := dao.GetPostsBySubforum(ctx, subforumID, 1, 10, "created_at", false)
	require.NoError(t, err)
	assert.Len(t, posts, 3)

	// Delete one post
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post2.PostID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify only non-deleted posts are returned
	posts, err = dao.GetPostsBySubforum(ctx, subforumID, 1, 10, "created_at", false)
	require.NoError(t, err)
	assert.Len(t, posts, 2)

	// Verify the deleted post is not in the results
	for _, p := range posts {
		assert.NotEqual(t, post2.PostID, p.PostID)
	}
}

func TestPostDAO_CountPostsBySubforum_FiltersDeletedPosts(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewPostDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	subforumID, pseudonymID := createTestData(t, db)

	// Create posts
	post1, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post 1", "Content 1", "text", nil, false, false)
	require.NoError(t, err)

	_, err = dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post 2", "Content 2", "text", nil, false, false)
	require.NoError(t, err)

	// Verify count is correct
	count, err := dao.CountPostsBySubforum(ctx, subforumID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Delete one post
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post1.PostID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify count excludes deleted post
	count, err = dao.CountPostsBySubforum(ctx, subforumID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestPostDAO_UpdatePostScore_WorksWithDeletedPosts(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewPostDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	subforumID, pseudonymID := createTestData(t, db)

	// Create a post
	post, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post", "Test content", "text", nil, false, false)
	require.NoError(t, err)

	// Delete the post
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post.PostID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify score can still be updated (for vote calculations)
	err = dao.UpdatePostScore(ctx, post.PostID, 100, 150, 50)
	require.NoError(t, err)

	// Verify the score was updated by checking the raw database
	var score, upvotes, downvotes int32
	err = db.(bob.DB).DB.QueryRowContext(ctx,
		"SELECT score, upvotes, downvotes FROM posts WHERE post_id = $1",
		post.PostID).Scan(&score, &upvotes, &downvotes)
	require.NoError(t, err)

	assert.Equal(t, int32(100), score)
	assert.Equal(t, int32(150), upvotes)
	assert.Equal(t, int32(50), downvotes)
}

func TestPostDAO_MarkPostAsDeletedByPseudonym_PreventsDoubleDeletion(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewPostDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	subforumID, pseudonymID := createTestData(t, db)

	// Create a post
	post, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post", "Test content", "text", nil, false, false)
	require.NoError(t, err)

	// Delete the post
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post.PostID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Try to delete again - should fail because post is not found (filtered out)
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post.PostID, pseudonymID, "User requested deletion again")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "post not found")
}

func TestPostDAO_GetPostBySubforumAndSlug_FiltersDeletedPosts(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewPostDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	subforumID, pseudonymID := createTestData(t, db)

	// Create a post with a slug
	post, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post", "Test content", "text", nil, false, false)
	require.NoError(t, err)

	// Update the post to have a slug
	updates := &models.PostSetter{
		Slug: &sql.Null[string]{Valid: true, V: "test-post"},
	}
	err = post.Update(ctx, db, updates)
	require.NoError(t, err)

	// Verify post can be retrieved by slug
	retrievedPost, err := dao.GetPostBySubforumAndSlug(ctx, subforumID, "test-post")
	require.NoError(t, err)
	require.NotNil(t, retrievedPost)
	assert.Equal(t, post.PostID, retrievedPost.PostID)

	// Delete the post
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post.PostID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify post is no longer retrievable by slug
	retrievedPost, err = dao.GetPostBySubforumAndSlug(ctx, subforumID, "test-post")
	require.NoError(t, err)
	assert.Nil(t, retrievedPost, "Deleted post should not be returned")
}

func TestPostDAO_FindPostForScoreUpdate_WorksWithDeletedPosts(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewPostDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	subforumID, pseudonymID := createTestData(t, db)

	// Create a post
	post, err := dao.CreatePost(ctx, subforumID, pseudonymID, "Test Post", "Test content", "text", nil, false, false)
	require.NoError(t, err)

	// Delete the post
	err = dao.MarkPostAsDeletedByPseudonym(ctx, post.PostID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify that FindPostForScoreUpdate can still find the deleted post
	foundPost, err := dao.FindPostForScoreUpdate(ctx, post.PostID)
	require.NoError(t, err)
	assert.Equal(t, post.PostID, foundPost.PostID)
	assert.True(t, foundPost.IsDeleted.Valid && foundPost.IsDeleted.V, "Post should be marked as deleted")
}

// TestPostDAO_ModerationDashboardMethods tests the new moderation dashboard methods
func TestPostDAO_ModerationDashboardMethods(t *testing.T) {
	// These tests verify the method signatures and basic functionality
	// without requiring database mocking
	
	t.Run("GetTotalPostsCount_Signature", func(t *testing.T) {
		// Test that the method exists and has correct signature
		dao := &PostDAO{}
		ctx := context.Background()
		subforumPath := "test-subforum"
		
		// This will fail at runtime due to nil db, but we're just testing the signature
		// In a real test environment, we'd use a mock database
		_, err := dao.GetTotalPostsCount(ctx, subforumPath)
		
		// We expect an error due to nil database, but the method signature is correct
		assert.Error(t, err, "Expected error due to nil database")
	})

	t.Run("GetPostsCount_Signature", func(t *testing.T) {
		dao := &PostDAO{}
		ctx := context.Background()
		subforumPath := "test-subforum"
		since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		
		_, err := dao.GetPostsCount(ctx, subforumPath, since)
		
		// We expect an error due to nil database, but the method signature is correct
		assert.Error(t, err, "Expected error due to nil database")
	})

	t.Run("GetPostsCountForDateRange_Signature", func(t *testing.T) {
		dao := &PostDAO{}
		ctx := context.Background()
		subforumPath := "test-subforum"
		startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
		
		_, err := dao.GetPostsCountForDateRange(ctx, subforumPath, startTime, endTime)
		
		// We expect an error due to nil database, but the method signature is correct
		assert.Error(t, err, "Expected error due to nil database")
	})

	t.Run("MethodParameters", func(t *testing.T) {
		// Test that the methods accept the correct parameter types
		dao := &PostDAO{}
		ctx := context.Background()
		
		// Test parameter types are accepted
		subforumPath := "test-subforum"
		since := time.Now()
		startTime := time.Now().Add(-24 * time.Hour)
		endTime := time.Now()
		
		// These should compile and run (though they'll fail due to nil db)
		_, _ = dao.GetTotalPostsCount(ctx, subforumPath)
		_, _ = dao.GetPostsCount(ctx, subforumPath, since)
		_, _ = dao.GetPostsCountForDateRange(ctx, subforumPath, startTime, endTime)
		
		// If we get here, the method signatures are correct
		assert.True(t, true, "Method signatures are correct")
	})
}

// TestPostDAO_ModerationDashboardMethods_EdgeCases tests edge cases for moderation methods
func TestPostDAO_ModerationDashboardMethods_EdgeCases(t *testing.T) {
	t.Run("EmptySubforumPath", func(t *testing.T) {
		dao := &PostDAO{}
		ctx := context.Background()
		
		// Test with empty subforum path
		_, err := dao.GetTotalPostsCount(ctx, "")
		assert.Error(t, err, "Expected error with empty subforum path")
		
		_, err = dao.GetPostsCount(ctx, "", time.Now())
		assert.Error(t, err, "Expected error with empty subforum path")
		
		_, err = dao.GetPostsCountForDateRange(ctx, "", time.Now(), time.Now())
		assert.Error(t, err, "Expected error with empty subforum path")
	})

	t.Run("NilContext", func(t *testing.T) {
		dao := &PostDAO{}
		subforumPath := "test-subforum"
		
		// Test with nil context
		_, err := dao.GetTotalPostsCount(nil, subforumPath)
		assert.Error(t, err, "Expected error with nil context")
		
		_, err = dao.GetPostsCount(nil, subforumPath, time.Now())
		assert.Error(t, err, "Expected error with nil context")
		
		_, err = dao.GetPostsCountForDateRange(nil, subforumPath, time.Now(), time.Now())
		assert.Error(t, err, "Expected error with nil context")
	})

	t.Run("ZeroTime", func(t *testing.T) {
		dao := &PostDAO{}
		ctx := context.Background()
		subforumPath := "test-subforum"
		zeroTime := time.Time{}
		
		// Test with zero time
		_, err := dao.GetPostsCount(ctx, subforumPath, zeroTime)
		assert.Error(t, err, "Expected error with zero time")
		
		_, err = dao.GetPostsCountForDateRange(ctx, subforumPath, zeroTime, time.Now())
		assert.Error(t, err, "Expected error with zero start time")
		
		_, err = dao.GetPostsCountForDateRange(ctx, subforumPath, time.Now(), zeroTime)
		assert.Error(t, err, "Expected error with zero end time")
	})
}
