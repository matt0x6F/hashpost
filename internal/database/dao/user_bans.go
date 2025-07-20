package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// UserBanDAO implements UserBanDAOInterface
type UserBanDAO struct {
	db bob.Executor
}

// NewUserBanDAO creates a new UserBanDAO
func NewUserBanDAO(db bob.Executor) *UserBanDAO {
	return &UserBanDAO{db: db}
}

// CreateUserBan creates a new user ban
func (dao *UserBanDAO) CreateUserBan(ctx context.Context, ban *models.UserBanSetter) (*models.UserBan, error) {
	log.Debug().
		Int32("subforum_id", *ban.SubforumID).
		Int64("banned_user_id", *ban.BannedUserID).
		Int64("banned_by_user_id", *ban.BannedByUserID).
		Str("banned_by_pseudonym_id", *ban.BannedByPseudonymID).
		Msg("Creating user ban")

	// Set default values if not provided
	if ban.CreatedAt == nil {
		now := sql.Null[time.Time]{}
		now.Scan(time.Now())
		ban.CreatedAt = &now
	}

	if ban.IsActive == nil {
		isActive := sql.Null[bool]{}
		isActive.Scan(true)
		ban.IsActive = &isActive
	}

	// Use the generated UserBans table helper
	userBan, err := models.UserBans.Insert(ban).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create user ban: %w", err)
	}

	return userBan, nil
}

// GetUserBanByID retrieves a user ban by ID
func (dao *UserBanDAO) GetUserBanByID(ctx context.Context, banID int64) (*models.UserBan, error) {
	// Use the generated FindUserBan function
	userBan, err := models.FindUserBan(ctx, dao.db, banID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user ban by ID: %w", err)
	}

	return userBan, nil
}

// GetUserBansBySubforum retrieves user bans for a specific subforum
func (dao *UserBanDAO) GetUserBansBySubforum(ctx context.Context, subforumID int32, page, limit int) ([]*models.UserBan, error) {
	offset := page * limit

	// Use the generated UserBans table helper with where clause and pagination
	userBans, err := models.UserBans.Query(
		models.SelectWhere.UserBans.SubforumID.EQ(subforumID),
		models.SelectWhere.UserBans.IsActive.EQ(true),
		sm.Limit(limit),
		sm.Offset(offset),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bans by subforum: %w", err)
	}

	return userBans, nil
}

// GetUserBansByUser retrieves user bans for a specific user
func (dao *UserBanDAO) GetUserBansByUser(ctx context.Context, bannedUserID int64, page, limit int) ([]*models.UserBan, error) {
	offset := page * limit

	// Use the generated UserBans table helper with where clause and pagination
	userBans, err := models.UserBans.Query(
		models.SelectWhere.UserBans.BannedUserID.EQ(bannedUserID),
		models.SelectWhere.UserBans.IsActive.EQ(true),
		sm.Limit(limit),
		sm.Offset(offset),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bans by user: %w", err)
	}

	return userBans, nil
}

// CountUserBansBySubforum counts user bans for a specific subforum
func (dao *UserBanDAO) CountUserBansBySubforum(ctx context.Context, subforumID int32) (int64, error) {
	// Use the generated UserBans table helper with count
	count, err := models.UserBans.Query(
		models.SelectWhere.UserBans.SubforumID.EQ(subforumID),
		models.SelectWhere.UserBans.IsActive.EQ(true),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count user bans by subforum: %w", err)
	}

	return count, nil
}

// CountUserBansByUser counts user bans for a specific user
func (dao *UserBanDAO) CountUserBansByUser(ctx context.Context, bannedUserID int64) (int64, error) {
	// Use the generated UserBans table helper with count
	count, err := models.UserBans.Query(
		models.SelectWhere.UserBans.BannedUserID.EQ(bannedUserID),
		models.SelectWhere.UserBans.IsActive.EQ(true),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count user bans by user: %w", err)
	}

	return count, nil
}

// IsUserBannedFromSubforum checks if a user is banned from a specific subforum
func (dao *UserBanDAO) IsUserBannedFromSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	// Check for active bans for this user in this subforum
	userBans, err := models.UserBans.Query(
		models.SelectWhere.UserBans.BannedUserID.EQ(userID),
		models.SelectWhere.UserBans.SubforumID.EQ(subforumID),
		models.SelectWhere.UserBans.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is banned from subforum: %w", err)
	}

	// Check if any of the bans are still active (not expired or permanent)
	for _, ban := range userBans {
		// If ban is permanent, user is banned
		if ban.IsPermanent.Valid && ban.IsPermanent.V {
			return true, nil
		}

		// If ban has expiration and it's not expired, user is banned
		if ban.ExpiresAt.Valid && ban.ExpiresAt.V.After(time.Now()) {
			return true, nil
		}
	}

	return false, nil
}

// UpdateUserBan updates a user ban
func (dao *UserBanDAO) UpdateUserBan(ctx context.Context, banID int64, updates *models.UserBanSetter) error {
	// First get the user ban
	userBan, err := dao.GetUserBanByID(ctx, banID)
	if err != nil {
		return fmt.Errorf("failed to get user ban for update: %w", err)
	}
	if userBan == nil {
		return fmt.Errorf("user ban not found")
	}

	// Use the generated Update method
	err = userBan.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update user ban: %w", err)
	}

	return nil
}

// DeactivateUserBan deactivates a user ban
func (dao *UserBanDAO) DeactivateUserBan(ctx context.Context, banID int64) error {
	isActive := sql.Null[bool]{}
	isActive.Scan(false)

	updates := &models.UserBanSetter{
		IsActive: &isActive,
	}

	return dao.UpdateUserBan(ctx, banID, updates)
}
