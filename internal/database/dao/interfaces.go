package dao

import (
	"context"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
)

// UserDAOInterface defines the interface for user data access operations
type UserDAOInterface interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, userID int64, updates *models.UserSetter) error
	DeleteUser(ctx context.Context, userID int64) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	CountUsers(ctx context.Context) (int64, error)
	UpdateLastActive(ctx context.Context, userID int64) error
	SuspendUser(ctx context.Context, userID int64, reason string, expiresAt *time.Time) error
	UnsuspendUser(ctx context.Context, userID int64) error
}

// SecurePseudonymDAOInterface defines the interface for secure pseudonym data access operations
type SecurePseudonymDAOInterface interface {
	CreatePseudonymWithIdentityMapping(ctx context.Context, userID int64, displayName string) (*models.Pseudonym, error)
	GetPseudonymByID(ctx context.Context, pseudonymID string) (*models.Pseudonym, error)
	GetPseudonymByDisplayName(ctx context.Context, displayName string) (*models.Pseudonym, error)
	GetPseudonymsByUserID(ctx context.Context, userID int64, roleName, scope string) ([]*models.Pseudonym, error)
	GetDefaultPseudonymByUserID(ctx context.Context, userID int64, roleName, scope string) (*models.Pseudonym, error)
	UpdatePseudonym(ctx context.Context, pseudonymID string, updates *models.PseudonymSetter) error
	DeletePseudonym(ctx context.Context, pseudonymID string) error
	VerifyPseudonymOwnership(ctx context.Context, pseudonymID string, userID int64, roleName, scope string) (bool, error)
	GetUserIDByPseudonym(ctx context.Context, pseudonymID, roleName, scope string) (int64, error)
}

// IdentityMappingDAOInterface defines the interface for identity mapping data access operations
type IdentityMappingDAOInterface interface {
	CreateIdentityMapping(ctx context.Context, mapping *models.IdentityMappingSetter) (*models.IdentityMapping, error)
	GetIdentityMappingByPseudonymID(ctx context.Context, pseudonymID string) (*models.IdentityMapping, error)
	GetIdentityMappingsByPseudonymID(ctx context.Context, pseudonymID string) (models.IdentityMappingSlice, error)
	GetIdentityMappingsByUserID(ctx context.Context, userID int64) (models.IdentityMappingSlice, error)
	GetIdentityMappingsByFingerprint(ctx context.Context, fingerprint string) (models.IdentityMappingSlice, error)
	DeactivateIdentityMapping(ctx context.Context, mappingID string) error
}

// RoleKeyDAOInterface defines the interface for role key data access operations
type RoleKeyDAOInterface interface {
	CreateRoleKey(ctx context.Context, roleName, scope string, keyData []byte, capabilities []string, expiresAt time.Time, createdBy int64) (*models.RoleKey, error)
	GetRoleKey(ctx context.Context, roleName, scope string) (*models.RoleKey, error)
	GetPerUserRoleKey(ctx context.Context, roleName, scope string, createdBy int64) (*models.RoleKey, error)
	GetRoleKeyByID(ctx context.Context, keyID string) (*models.RoleKey, error)
	ListRoleKeys(ctx context.Context) ([]*models.RoleKey, error)
	ListRoleKeysByRole(ctx context.Context, roleName string) ([]*models.RoleKey, error)
	DeactivateRoleKey(ctx context.Context, keyID string) error
	ValidateKeyCapability(ctx context.Context, roleName, scope, requiredCapability string) (bool, error)
	GetKeyData(ctx context.Context, roleName, scope string) ([]byte, error)
	GetPerUserKeyData(ctx context.Context, roleName, scope string, createdBy int64) ([]byte, error)
	EnsureDefaultKeys(ctx context.Context, ibeSystem interface{}, userID int64) error
}

// SubforumDAOInterface defines the interface for subforum data access operations
type SubforumDAOInterface interface {
	CreateSubforum(ctx context.Context, name, displayName, description, sidebarText, rulesText string, isNSFW, isPrivate, isRestricted bool) (*models.Subforum, error)
	GetSubforumByID(ctx context.Context, subforumID int32) (*models.Subforum, error)
	GetSubforumByName(ctx context.Context, name string) (*models.Subforum, error)
	ListSubforums(ctx context.Context) ([]*models.Subforum, error)
}

// PostDAOInterface defines the interface for post data access operations
type PostDAOInterface interface {
	CreatePost(ctx context.Context, subforumID int32, pseudonymID, title, content, postType string, url *string, isNSFW, isSpoiler bool) (*models.Post, error)
	GetPostByID(ctx context.Context, postID int64) (*models.Post, error)
	GetPostsBySubforum(ctx context.Context, subforumID int32, page, limit int, sortField string, sortDesc bool) ([]*models.Post, error)
	CountPostsBySubforum(ctx context.Context, subforumID int32) (int64, error)
	CountPostsByPseudonym(ctx context.Context, pseudonymID string) (int64, error)
	GetPostBySubforumAndSlug(ctx context.Context, subforumID int32, slug string) (*models.Post, error)
	UpdatePostScore(ctx context.Context, postID int64, score, upvotes, downvotes int32) error
	IncrementViewCount(ctx context.Context, postID int64) error
	UpdateCommentCount(ctx context.Context, postID int64, commentCount int32) error
	SetLocked(ctx context.Context, postID int64, locked bool) error
	SetSticky(ctx context.Context, postID int64, sticky bool) error
	SetRemoved(ctx context.Context, postID int64, removed bool) error
	MarkPostAsDeletedByPseudonym(ctx context.Context, postID int64, pseudonymID string, reason string) error
}

// CommentDAOInterface defines the interface for comment data access operations
type CommentDAOInterface interface {
	CreateComment(ctx context.Context, postID int64, pseudonymID, content string, parentCommentID *int64) (*models.Comment, error)
	GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error)
	GetCommentsByPost(ctx context.Context, postID int64) ([]*models.Comment, error)
	GetCommentsByPostWithNestedReplies(ctx context.Context, postID int64) ([]*models.Comment, error)
	CountCommentsByPost(ctx context.Context, postID int64) (int64, error)
	CountCommentsByPseudonym(ctx context.Context, pseudonymID string) (int64, error)
	UpdateCommentScore(ctx context.Context, commentID int64, score, upvotes, downvotes int32) error
	MarkCommentAsDeletedByPseudonym(ctx context.Context, commentID int64, pseudonymID string, reason string) error
	DeleteCommentByUser(ctx context.Context, commentID int64, reason string) error
}

// VoteDAOInterface defines the interface for vote data access operations
type VoteDAOInterface interface {
	CreateVote(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) (*models.Vote, error)
	GetVoteByPseudonymAndContent(ctx context.Context, pseudonymID, contentType string, contentID int64) (*models.Vote, error)
	UpdateVote(ctx context.Context, voteID int64, voteValue int32) (*models.Vote, error)
	DeleteVote(ctx context.Context, voteID int64) error
	UpsertVote(ctx context.Context, pseudonymID, contentType string, contentID int64, voteValue int32) (*models.Vote, error)
	GetVotesByContent(ctx context.Context, contentType string, contentID int64) ([]*models.Vote, error)
	CountVotesByContent(ctx context.Context, contentType string, contentID int64) (int, error)
	GetVoteSummaryByContent(ctx context.Context, contentType string, contentID int64) (upvotes, downvotes, total int, err error)
}

// APIKeyDAOInterface defines the interface for API key data access operations
type APIKeyDAOInterface interface {
	CreateAPIKey(ctx context.Context, userID int64, pseudonymID string, permissions map[string]interface{}) (*models.APIKey, error)
	GetAPIKeyByID(ctx context.Context, apiKeyID string) (*models.APIKey, error)
	GetAPIKeysByUserID(ctx context.Context, userID int64) ([]*models.APIKey, error)
	UpdateAPIKey(ctx context.Context, apiKeyID string, updates *models.APIKeySetter) error
	DeleteAPIKey(ctx context.Context, apiKeyID string) error
	ValidateAPIKey(ctx context.Context, apiKeyHash string) (*models.APIKey, error)
}

// UserBlocksDAOInterface defines the interface for user blocks data access operations
type UserBlocksDAOInterface interface {
	CreateUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (*models.UserBlock, error)
	GetUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (*models.UserBlock, error)
	GetUserBlocksByBlocker(ctx context.Context, blockerPseudonymID string) ([]*models.UserBlock, error)
	GetUserBlocksByBlockedUser(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error)
	DeleteUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) error
	DeleteUserBlockByID(ctx context.Context, blockID int64) error
	IsUserBlocked(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (bool, error)
	IsPseudonymBlockedByUser(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (bool, error)
	IsUserBlockedAtFingerprintLevel(ctx context.Context, blockerPseudonymID string, blockedUserID int64) (bool, error)
	IsUserBlockedByAnyPseudonym(ctx context.Context, blockerUserID int64, blockedPseudonymID string) (bool, error)
	GetFingerprintLevelBlocks(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error)
}

// UserPreferencesDAOInterface defines the interface for user preferences data access operations
type UserPreferencesDAOInterface interface {
	CreateUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error)
	GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreference, error)
	UpdateUserPreferences(ctx context.Context, userID int64, updates *models.UserPreferenceSetter) error
	UpsertUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error)
}
