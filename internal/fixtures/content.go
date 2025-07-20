package fixtures

import (
	"database/sql"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
)

// CreateTestPost creates a test post
func CreateTestPost() *models.Post {
	return &models.Post{
		PostID:      123,
		SubforumID:  1,
		PseudonymID: "test-pseudonym-id",
		Title:       "Test Post",
		Content:     sql.Null[string]{V: "Test content", Valid: true},
		PostType:    "text",
		Score:       sql.Null[int32]{V: 10, Valid: true},
		Upvotes:     sql.Null[int32]{V: 15, Valid: true},
		Downvotes:   sql.Null[int32]{V: 5, Valid: true},
		IsDeleted:   sql.Null[bool]{V: false, Valid: true},
		CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestComment creates a test comment
func CreateTestComment() *models.Comment {
	return &models.Comment{
		CommentID:   456,
		PostID:      123,
		PseudonymID: "test-pseudonym-id",
		Content:     "Test comment",
		Score:       sql.Null[int32]{V: 5, Valid: true},
		Upvotes:     sql.Null[int32]{V: 8, Valid: true},
		Downvotes:   sql.Null[int32]{V: 3, Valid: true},
		IsDeleted:   sql.Null[bool]{V: false, Valid: true},
		CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestSubforum creates a test subforum
func CreateTestSubforum() *models.Subforum {
	return &models.Subforum{
		SubforumID:   1,
		Name:         "test-subforum",
		DisplayName:  "Test Subforum",
		Description:  sql.Null[string]{V: "A test subforum", Valid: true},
		IsPrivate:    sql.Null[bool]{V: false, Valid: true},
		IsNSFW:       sql.Null[bool]{V: false, Valid: true},
		IsRestricted: sql.Null[bool]{V: false, Valid: true},
	}
}
