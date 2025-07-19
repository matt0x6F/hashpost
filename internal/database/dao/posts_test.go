package dao

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	ctx := context.Background()

	// Create test subforum using raw SQL
	_, err := db.ExecContext(ctx, `
		INSERT INTO subforums (name, display_name, description) 
		VALUES ('Test Subforum', 'Test Subforum', 'A test subforum for testing')
		RETURNING subforum_id
	`)
	require.NoError(t, err, "Failed to create test subforum")

	// Get the subforum ID
	var subforumID int32
	err = db.(bob.DB).DB.QueryRowContext(ctx,
		"SELECT subforum_id FROM subforums WHERE name = 'Test Subforum'").Scan(&subforumID)
	require.NoError(t, err, "Failed to get subforum ID")

	// Create test pseudonym using raw SQL
	_, err = db.ExecContext(ctx, `
		INSERT INTO pseudonyms (pseudonym_id, display_name, bio, is_default) 
		VALUES ('test-pseudonym-123', 'Test Pseudonym', 'A test pseudonym', true)
	`)
	require.NoError(t, err, "Failed to create test pseudonym")

	return subforumID, "test-pseudonym-123"
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
