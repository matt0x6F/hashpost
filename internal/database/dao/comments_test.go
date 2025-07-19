package dao

import (
	"context"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestDataForComments creates the required test data (subforum, pseudonym, and post)
func createTestDataForComments(t *testing.T, db bob.Executor) (int64, string) {
	return createTestDataForCommentsWithConfig(t, db, DefaultTestConfig())
}

// createTestDataForCommentsWithConfig creates test data for comments with custom configuration
func createTestDataForCommentsWithConfig(t *testing.T, db bob.Executor, config *TestDBConfig) (int64, string) {
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

	// Create test post using raw SQL
	_, err = db.ExecContext(ctx, `
		INSERT INTO posts (subforum_id, pseudonym_id, title, content, post_type) 
		VALUES ($1, $2, 'Test Post', 'Test content', 'text')
		RETURNING post_id
	`, subforumID, config.PseudonymID)
	require.NoError(t, err, "Failed to create test post")

	// Get the post ID
	var postID int64
	err = db.(bob.DB).DB.QueryRowContext(ctx,
		"SELECT post_id FROM posts WHERE title = 'Test Post'").Scan(&postID)
	require.NoError(t, err, "Failed to get post ID")

	return postID, config.PseudonymID
}

func TestCommentDAO_GetCommentByID_FiltersDeletedComments(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create a test comment
	comment, err := dao.CreateComment(ctx, postID, pseudonymID, "Test comment", nil)
	require.NoError(t, err)
	require.NotNil(t, comment)

	// Verify comment can be retrieved normally
	retrievedComment, err := dao.GetCommentByID(ctx, comment.CommentID)
	require.NoError(t, err)
	require.NotNil(t, retrievedComment)
	assert.Equal(t, comment.CommentID, retrievedComment.CommentID)
	assert.Equal(t, "Test comment", retrievedComment.Content)

	// Mark comment as deleted
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify comment is no longer retrievable
	retrievedComment, err = dao.GetCommentByID(ctx, comment.CommentID)
	require.NoError(t, err)
	assert.Nil(t, retrievedComment, "Deleted comment should not be returned")
}

func TestCommentDAO_GetCommentsByPost_FiltersDeletedComments(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create multiple comments
	comment1, err := dao.CreateComment(ctx, postID, pseudonymID, "Test Comment 1", nil)
	require.NoError(t, err)

	comment2, err := dao.CreateComment(ctx, postID, pseudonymID, "Test Comment 2", nil)
	require.NoError(t, err)

	_, err = dao.CreateComment(ctx, postID, pseudonymID, "Test Comment 3", nil)
	require.NoError(t, err)

	// Verify all comments are retrievable
	comments, err := dao.GetCommentsByPost(ctx, postID)
	require.NoError(t, err)
	assert.Len(t, comments, 3)

	// Delete one comment
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment2.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify only non-deleted comments are returned
	comments, err = dao.GetCommentsByPost(ctx, postID)
	require.NoError(t, err)
	assert.Len(t, comments, 2)

	// Verify the deleted comment is not in the results
	for _, c := range comments {
		assert.NotEqual(t, comment2.CommentID, c.CommentID)
	}

	// Verify comment1 is still in the results
	foundComment1 := false
	for _, c := range comments {
		if c.CommentID == comment1.CommentID {
			foundComment1 = true
			break
		}
	}
	assert.True(t, foundComment1, "Comment1 should still be in the results")
}

func TestCommentDAO_CountCommentsByPost_FiltersDeletedComments(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create comments
	comment1, err := dao.CreateComment(ctx, postID, pseudonymID, "Test Comment 1", nil)
	require.NoError(t, err)

	_, err = dao.CreateComment(ctx, postID, pseudonymID, "Test Comment 2", nil)
	require.NoError(t, err)

	// Verify count is correct
	count, err := dao.CountCommentsByPost(ctx, postID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Delete one comment
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment1.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify count excludes deleted comment
	count, err = dao.CountCommentsByPost(ctx, postID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestCommentDAO_UpdateCommentScore_WorksWithDeletedComments(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create a comment
	comment, err := dao.CreateComment(ctx, postID, pseudonymID, "Test comment", nil)
	require.NoError(t, err)

	// Delete the comment
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify score can still be updated (for vote calculations)
	err = dao.UpdateCommentScore(ctx, comment.CommentID, 100, 150, 50)
	require.NoError(t, err)

	// Verify the score was updated by checking the raw database
	var score, upvotes, downvotes int32
	err = db.(bob.DB).DB.QueryRowContext(ctx,
		"SELECT score, upvotes, downvotes FROM comments WHERE comment_id = $1",
		comment.CommentID).Scan(&score, &upvotes, &downvotes)
	require.NoError(t, err)

	assert.Equal(t, int32(100), score)
	assert.Equal(t, int32(150), upvotes)
	assert.Equal(t, int32(50), downvotes)
}

func TestCommentDAO_MarkCommentAsDeletedByPseudonym_PreventsDoubleDeletion(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create a comment
	comment, err := dao.CreateComment(ctx, postID, pseudonymID, "Test comment", nil)
	require.NoError(t, err)

	// Delete the comment
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Try to delete again - should fail because comment is not found (filtered out)
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment.CommentID, pseudonymID, "User requested deletion again")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")
}

func TestCommentDAO_GetCommentsByPostWithNestedReplies_IncludesDeletedComments(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create a comment
	comment, err := dao.CreateComment(ctx, postID, pseudonymID, "Test comment", nil)
	require.NoError(t, err)

	// Delete the comment
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify deleted comment is still returned but with cleared content
	comments, err := dao.GetCommentsByPostWithNestedReplies(ctx, postID)
	require.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "[deleted]", comments[0].Content)
	assert.True(t, comments[0].IsDeleted.Valid && comments[0].IsDeleted.V)
}

func TestCommentDAO_CountCommentsByPseudonym_FiltersDeletedComments(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create comments by the same pseudonym
	comment1, err := dao.CreateComment(ctx, postID, pseudonymID, "Test Comment 1", nil)
	require.NoError(t, err)

	_, err = dao.CreateComment(ctx, postID, pseudonymID, "Test Comment 2", nil)
	require.NoError(t, err)

	// Verify count is correct
	count, err := dao.CountCommentsByPseudonym(ctx, pseudonymID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Delete one comment
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment1.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify count excludes deleted comment
	count, err = dao.CountCommentsByPseudonym(ctx, pseudonymID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestCommentDAO_FindCommentForScoreUpdate_WorksWithDeletedComments(t *testing.T) {
	db := setupTestDB(t)
	defer db.(bob.DB).DB.Close()

	dao := NewCommentDAO(db)
	ctx := context.Background()

	// Clean up after test
	defer cleanupTestData(t, db)

	// Create test data
	postID, pseudonymID := createTestDataForComments(t, db)

	// Create a comment
	comment, err := dao.CreateComment(ctx, postID, pseudonymID, "Test comment", nil)
	require.NoError(t, err)

	// Delete the comment
	err = dao.MarkCommentAsDeletedByPseudonym(ctx, comment.CommentID, pseudonymID, "User requested deletion")
	require.NoError(t, err)

	// Verify that FindCommentForScoreUpdate can still find the deleted comment
	foundComment, err := dao.FindCommentForScoreUpdate(ctx, comment.CommentID)
	require.NoError(t, err)
	assert.Equal(t, comment.CommentID, foundComment.CommentID)
	assert.True(t, foundComment.IsDeleted.Valid && foundComment.IsDeleted.V, "Comment should be marked as deleted")
}
