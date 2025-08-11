package dao

import (
	"context"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
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

// PseudonymDAOInterface defines the interface for secure pseudonym data access operations
type PseudonymDAOInterface interface {
	CreatePseudonymWithIdentityMapping(ctx context.Context, userID int64, displayName string) (*models.Pseudonym, error)
	GetPseudonymByID(ctx context.Context, pseudonymID string) (*models.Pseudonym, error)
	GetPseudonymByDisplayName(ctx context.Context, displayName string) (*models.Pseudonym, error)
	GetPseudonymBySlug(ctx context.Context, slug string) (*models.Pseudonym, error)
	GetPseudonymsByUserID(ctx context.Context, userID int64, activePseudonymID, roleName, scope string) ([]*models.Pseudonym, error)
	GetDefaultPseudonymByUserID(ctx context.Context, userID int64, roleName, scope string) (*models.Pseudonym, error)
	UpdatePseudonym(ctx context.Context, pseudonymID string, updates *models.PseudonymSetter) error
	DeletePseudonym(ctx context.Context, pseudonymID string) error
	DeactivatePseudonym(ctx context.Context, pseudonymID string, userID int64, activePseudonymID, roleName, scope string) error
	VerifyPseudonymOwnership(ctx context.Context, pseudonymID string, userID int64, activePseudonymID, roleName, scope string) (bool, error)
	GetPseudonymsByRealIdentity(ctx context.Context, realIdentity string, roleName, scope string) ([]*models.Pseudonym, error)
	GetPseudonymsByRealIdentityDirect(ctx context.Context, realIdentity string) ([]*models.Pseudonym, error)
	GetRealIdentityByPseudonym(ctx context.Context, pseudonymID, roleName, scope string) (string, error)
	GetUserIDByPseudonym(ctx context.Context, pseudonymID, roleName, scope string) (int64, error)
	ArePseudonymsOwnedBySameUser(ctx context.Context, pseudonymID1, pseudonymID2 string) (bool, error)
	UpdateLastActive(ctx context.Context, pseudonymID string) error
	GenerateSlugFromDisplayName(ctx context.Context, displayName string) (string, error)
	CalculateKarmaForPseudonym(ctx context.Context, pseudonymID string) (int32, error)
	UpdateKarmaForPseudonym(ctx context.Context, pseudonymID string) error
	DeleteByUserID(ctx context.Context, userID int64) error
}

// IdentityMappingDAOInterface defines the interface for identity mapping data access operations
type IdentityMappingDAOInterface interface {
	GetIdentityMappingByPseudonymID(ctx context.Context, pseudonymID string) (*models.IdentityMapping, error)
	GetIdentityMappingsByPseudonymID(ctx context.Context, pseudonymID string) (models.IdentityMappingSlice, error)
	GetIdentityMappingsByFingerprint(ctx context.Context, fingerprint string) (models.IdentityMappingSlice, error)
	GetAllActiveIdentityMappings(ctx context.Context) (models.IdentityMappingSlice, error)
	CreateIdentityMapping(ctx context.Context, mapping *models.IdentityMappingSetter) (*models.IdentityMapping, error)
	UpdateIdentityMapping(ctx context.Context, mappingID string, updates *models.IdentityMappingSetter) error
	DeleteByUserID(ctx context.Context, userID int64) error
	DeactivateIdentityMapping(ctx context.Context, mappingID string) error
	GetCorrelationData(ctx context.Context, pseudonymID string) (*models.IdentityMapping, models.IdentityMappingSlice, error)
}

// RoleKeyDAOInterface defines the interface for role key data access operations
type RoleKeyDAOInterface interface {
	CreateRoleKey(ctx context.Context, roleName, scope string, keyData []byte, capabilities []string, expiresAt time.Time, createdByPseudonymID string, pseudonymID string, subforumID *int32) (*models.RoleKey, error)
	GetRoleKey(ctx context.Context, pseudonymID string, scope string, subforumID *int32) (*models.RoleKey, error)
	GetRoleKeyByID(ctx context.Context, keyID string) (*models.RoleKey, error)
	ListRoleKeys(ctx context.Context) ([]*models.RoleKey, error)
	ListRoleKeysByPseudonym(ctx context.Context, pseudonymID string) ([]*models.RoleKey, error)
	GetModeratorsForSubforum(ctx context.Context, subforumID int32) ([]*models.RoleKey, error)
	DeactivateRoleKey(ctx context.Context, keyID string) error
	ValidateKeyCapability(ctx context.Context, pseudonymID string, scope, requiredCapability string, subforumID *int32) (bool, error)
	GetKeyData(ctx context.Context, pseudonymID string, scope string, subforumID *int32) ([]byte, error)
	// Platform-level operations that work with roles instead of pseudonyms
	GetPlatformKeyData(ctx context.Context, roleName, scope string) ([]byte, error)
	ValidatePlatformKeyCapability(ctx context.Context, roleName, scope, requiredCapability string) (bool, error)
	CreateRoleKeyWithIBE(ctx context.Context, roleName, scope string, capabilities []string, expiresAt time.Time, createdByPseudonymID string, pseudonymID string, subforumID *int32) (*models.RoleKey, error)
	EnsureDefaultKeys(ctx context.Context, ibeSystem interface{}, pseudonymID string, userRoles []string) error
	DeleteByPseudonymID(ctx context.Context, pseudonymID string) error
}

// SubforumDAOInterface defines the interface for subforum data access operations
type SubforumDAOInterface interface {
	CreateSubforum(ctx context.Context, name, displayName, description, sidebarText, communityType, governanceStyle string, isNSFW, isPrivate, isRestricted bool, ownerPseudonymID string) (*models.Subforum, error)
	GetSubforumByID(ctx context.Context, subforumID int32) (*models.Subforum, error)
	GetSubforumByName(ctx context.Context, name string) (*models.Subforum, error)
	GetSubforumByCommunityTypeAndName(ctx context.Context, communityType, name string) (*models.Subforum, error)
	ListSubforums(ctx context.Context) ([]*models.Subforum, error)
	ListSubforumsByCommunityType(ctx context.Context, communityType string) ([]*models.Subforum, error)
	UpdatePostCount(ctx context.Context, subforumID int32, postCount int32) error
	UpdateSubscriberCount(ctx context.Context, subforumID int32, subscriberCount int32) error
	UpdateSettings(ctx context.Context, subforumID int32, allowImages, allowVideos, allowPolls, requireFlair, isPrivate, isRestricted, isNSFW bool, minimumAccountAgeHours, minimumKarmaRequired int, description, sidebarText string) error
	UpdateRules(ctx context.Context, subforumID int32, rules []byte) error
}

// PostDAOInterface defines the interface for post data access operations
type PostDAOInterface interface {
	CreatePost(ctx context.Context, subforumID int32, pseudonymID, title, content, postType string, url *string, isNSFW, isSpoiler bool) (*models.Post, error)
	GetPostByID(ctx context.Context, postID int64) (*models.Post, error)
	GetPostsBySubforum(ctx context.Context, subforumID int32, page, limit int, sortField string, sortDesc bool) ([]*models.Post, error)
	CountPostsBySubforum(ctx context.Context, subforumID int32) (int64, error)
	CountPostsByPseudonym(ctx context.Context, pseudonymID string) (int64, error)
	GetPostsByPseudonym(ctx context.Context, pseudonymID string, page, limit int, sortField string, sortDesc bool) ([]*models.Post, error)
	CountPostsByPseudonymInSubforum(ctx context.Context, pseudonymID string, subforumID int32) (int64, error)
	GetSubforumsByPseudonym(ctx context.Context, pseudonymID string) ([]int32, error)
	GetPostBySubforumAndSlug(ctx context.Context, subforumID int32, slug string) (*models.Post, error)
	UpdatePostScore(ctx context.Context, postID int64, score, upvotes, downvotes int32) error
	IncrementViewCount(ctx context.Context, postID int64) error
	UpdateCommentCount(ctx context.Context, postID int64, commentCount int32) error
	SetLocked(ctx context.Context, postID int64, locked bool) error
	SetSticky(ctx context.Context, postID int64, sticky bool) error
	SetRemoved(ctx context.Context, postID int64, removed bool) error
	MarkPostAsDeletedByPseudonym(ctx context.Context, postID int64, pseudonymID string, reason string) error
	UpdatePost(ctx context.Context, postID int64, title, content string) error
	// Moderation dashboard methods
	GetPostsCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetTotalPostsCount(ctx context.Context, subforumPath string) (int, error)
	GetPostsCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
}

// CommentDAOInterface defines the interface for comment data access operations
type CommentDAOInterface interface {
	CreateComment(ctx context.Context, postID int64, pseudonymID, content string, parentCommentID *int64) (*models.Comment, error)
	GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error)
	GetCommentsByPost(ctx context.Context, postID int64) ([]*models.Comment, error)
	GetCommentsByPostWithNestedReplies(ctx context.Context, postID int64) ([]*models.Comment, error)
	CountCommentsByPost(ctx context.Context, postID int64) (int64, error)
	CountCommentsByPseudonym(ctx context.Context, pseudonymID string) (int64, error)
	GetCommentsByPseudonym(ctx context.Context, pseudonymID string, page, limit int, sortField string, sortDesc bool) ([]*models.Comment, error)
	CountCommentsByPseudonymInSubforum(ctx context.Context, pseudonymID string, subforumID int32) (int64, error)
	GetSubforumsByPseudonymComments(ctx context.Context, pseudonymID string) ([]int32, error)
	UpdateCommentScore(ctx context.Context, commentID int64, score, upvotes, downvotes int32) error
	MarkCommentAsDeletedByPseudonym(ctx context.Context, commentID int64, pseudonymID string, reason string) error
	DeleteCommentByUser(ctx context.Context, commentID int64, reason string) error
	SetRemoved(ctx context.Context, commentID int64, removed bool) error
	SetCommentRemoved(ctx context.Context, commentID int64, removed bool, reason, removedByPseudonymID string) error
	UpdateComment(ctx context.Context, commentID int64, content, editReason string) error
	// Moderation dashboard methods
	GetCommentsCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetTotalCommentsCount(ctx context.Context, subforumPath string) (int, error)
	GetCommentsCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
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
	// Moderation dashboard methods
	GetVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetPostVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetCommentVotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetPostUpvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetPostDownvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetCommentUpvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	GetCommentDownvotesCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
	// Daily activity methods
	GetPostVotesCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
	GetCommentVotesCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
	GetPostUpvotesCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
	GetPostDownvotesCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
	GetCommentUpvotesCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
	GetCommentDownvotesCountForDateRange(ctx context.Context, subforumPath string, startTime, endTime time.Time) (int, error)
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

// UserBlocksDAOInterface defines the interface for user blocking data access operations
type UserBlocksDAOInterface interface {
	CreateUserBlock(ctx context.Context, blockerPseudonymID string, blockedPseudonymID string, blockedUserID int64) (*models.UserBlock, error)
	GetUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (*models.UserBlock, error)
	GetUserBlocksByBlocker(ctx context.Context, blockerPseudonymID string) ([]*models.UserBlock, error)
	GetUserBlocksByBlockedUser(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error)
	DeleteUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) error
	IsUserBlocked(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (bool, error)
	IsPseudonymBlockedByUser(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (bool, error)
	IsUserBlockedAtFingerprintLevel(ctx context.Context, blockerPseudonymID string, blockedUserID int64) (bool, error)
	GetFingerprintLevelBlocks(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error)
	DeleteUserBlockByID(ctx context.Context, blockID int64) error
}

// UserPreferencesDAOInterface defines the interface for user preferences data access operations
type UserPreferencesDAOInterface interface {
	CreateUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error)
	GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreference, error)
	UpdateUserPreferences(ctx context.Context, userID int64, updates *models.UserPreferenceSetter) error
	UpsertUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error)
}

// CorrelationAuditDAOInterface defines the interface for correlation audit operations
type CorrelationAuditDAOInterface interface {
	CreateCorrelationAudit(ctx context.Context, auditRecord *models.CorrelationAuditSetter) error
	GetCorrelationHistory(ctx context.Context, correlationType string, page, limit int) (models.CorrelationAuditSlice, error)
}

// ReportDAOInterface defines the interface for report data access operations
type ReportDAOInterface interface {
	CreateReport(ctx context.Context, report *models.ReportSetter) (*models.Report, error)
	GetReportByID(ctx context.Context, reportID int64) (*models.Report, error)
	GetReports(ctx context.Context, status string, page, limit int) ([]*models.Report, error)
	CountReports(ctx context.Context, status string) (int64, error)
	UpdateReport(ctx context.Context, reportID int64, updates *models.ReportSetter) error
	ResolveReport(ctx context.Context, reportID int64, resolverUserID int64, resolverPseudonymID string, resolutionNotes string, action string) error
	// Moderation dashboard methods
	GetPendingReportsCount(ctx context.Context, subforumPath string) (int, error)
}

// SystemSettingsDAOInterface defines the interface for system settings data access operations
type SystemSettingsDAOInterface interface {
	GetSetting(ctx context.Context, settingKey string) (*models.SystemSetting, error)
	SetSetting(ctx context.Context, settingKey, settingValue, settingType string, updatedBy int64) error
	GetAllSettings(ctx context.Context) ([]*models.SystemSetting, error)
}

// ModerationActionDAOInterface defines the interface for moderation action data access operations
type ModerationActionDAOInterface interface {
	CreateModerationAction(ctx context.Context, action *models.ModerationActionSetter) (*models.ModerationAction, error)
	GetModerationActionByID(ctx context.Context, actionID int64) (*models.ModerationAction, error)
	GetModerationActions(ctx context.Context, actionType string, page, limit int) ([]*models.ModerationAction, error)
	CountModerationActions(ctx context.Context, actionType string) (int64, error)
	GetModerationActionsByModerator(ctx context.Context, moderatorUserID int64, page, limit int) ([]*models.ModerationAction, error)
	// Moderation dashboard methods
	GetModActionsCount(ctx context.Context, subforumPath string, since time.Time) (int, error)
}

// UserBanDAOInterface defines the interface for user ban data access operations
type UserBanDAOInterface interface {
	CreateUserBan(ctx context.Context, ban *models.UserBanSetter) (*models.UserBan, error)
	GetUserBanByID(ctx context.Context, banID int64) (*models.UserBan, error)
	GetUserBansBySubforum(ctx context.Context, subforumID int32, page, limit int) ([]*models.UserBan, error)
	GetUserBansByUser(ctx context.Context, bannedUserID int64, page, limit int) ([]*models.UserBan, error)
	CountUserBansBySubforum(ctx context.Context, subforumID int32) (int64, error)
	CountUserBansByUser(ctx context.Context, bannedUserID int64) (int64, error)
	IsUserBannedFromSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error)
	UpdateUserBan(ctx context.Context, banID int64, updates *models.UserBanSetter) error
	DeactivateUserBan(ctx context.Context, banID int64) error
	// Moderation dashboard methods
	GetBannedUsersCount(ctx context.Context, subforumPath string) (int, error)
}

// SubforumSubscriptionDAOInterface defines the interface for subforum subscription data access operations
type SubforumSubscriptionDAOInterface interface {
	CreateSubscription(ctx context.Context, pseudonymID string, subforumID int32, isFavorite bool) (*models.SubforumSubscription, error)
	GetSubscription(ctx context.Context, pseudonymID string, subforumID int32) (*models.SubforumSubscription, error)
	GetSubscriptionsByPseudonym(ctx context.Context, pseudonymID string) ([]*models.SubforumSubscription, error)
	IsSubscribed(ctx context.Context, pseudonymID string, subforumID int32) (bool, error)
	IsFavorite(ctx context.Context, pseudonymID string, subforumID int32) (bool, error)
	DeleteSubscription(ctx context.Context, pseudonymID string, subforumID int32) error
	CountSubscriptionsBySubforum(ctx context.Context, subforumID int32) (int64, error)
}

// PermissionDAOInterface defines the interface for permission data access operations
type PermissionDAOInterface interface {
	CanAccessPrivateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error)
	CanAccessPrivateSubforumWithActivePseudonym(ctx context.Context, userID int64, subforumID int32, activePseudonymID string) (bool, error)
	HasSubforumCapability(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error)
	HasSubforumCapabilityWithActivePseudonym(ctx context.Context, userID int64, subforumID int32, capability string, activePseudonymID string) (bool, error)
	CanModerateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error)
	GetUserSubforumRoles(ctx context.Context, userID int64, subforumID int32) ([]string, error)
	GetUserSubforumCapabilities(ctx context.Context, userID int64, subforumID int32) ([]string, error)
	GetActivePseudonymRolesAndCapabilities(ctx context.Context, userID int64, activePseudonymID string) ([]string, []string, error)
	// Unified permission methods that combine global and subforum-specific capabilities
	GetUnifiedActivePseudonymRolesAndCapabilities(ctx context.Context, userID int64, activePseudonymID string, subforumID *int32) ([]string, []string, error)
	HasUnifiedCapability(ctx context.Context, userID int64, activePseudonymID string, capability string, subforumID *int32) (bool, error)
}

// KeyRotationMigrationDAOInterface defines the interface for key rotation migration operations
type KeyRotationMigrationDAOInterface interface {
	GetDB() bob.Executor
	CreateMigration(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int32, createdBy int64) (*MigrationState, error)
	GetMigrationByDomain(ctx context.Context, domain string, oldKeyVersion, newKeyVersion int32) (*MigrationState, error)
	GetMigrationByID(ctx context.Context, migrationID string) (*MigrationState, error)
	UpdateMigrationStatus(ctx context.Context, migrationID, status string) error
	UpdateMigrationProgress(ctx context.Context, migrationID string, processedRecords, failedRecords int64, lastProcessedID *string) error
	GetUnmigratedBatch(ctx context.Context, migrationID, domain string, offset int, batchSize int, lastProcessedID *string) ([]*models.IdentityMapping, error)
	MarkRecordProcessing(ctx context.Context, migrationID, mappingID string) error
	MarkRecordCompleted(ctx context.Context, migrationID, mappingID string) error
	MarkRecordFailed(ctx context.Context, migrationID, mappingID, errorMessage string) error
	IsRecordAlreadyMigrated(ctx context.Context, migrationID, mappingID string) (bool, error)
	GetStuckRecords(ctx context.Context, migrationID string, timeoutMinutes int) ([]*models.MigrationProgress, error)
	ResetRecordStatus(ctx context.Context, migrationID, mappingID, status string) error
	GetMigrationProgress(ctx context.Context, migrationID string) (*MigrationProgress, error)
}
