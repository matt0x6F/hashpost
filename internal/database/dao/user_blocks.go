package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// UserBlocksDAO provides data access operations for user blocks
type UserBlocksDAO struct {
	db bob.Executor
}

// NewUserBlocksDAO creates a new UserBlocksDAO
func NewUserBlocksDAO(db bob.Executor) *UserBlocksDAO {
	return &UserBlocksDAO{
		db: db,
	}
}

// CreateUserBlock creates a new user block
func (dao *UserBlocksDAO) CreateUserBlock(ctx context.Context, blockerPseudonymID string, blockedPseudonymID string, blockedUserID int64) (*models.UserBlock, error) {
	log.Debug().
		Str("blocker_pseudonym_id", blockerPseudonymID).
		Str("blocked_pseudonym_id", blockedPseudonymID).
		Int64("blocked_user_id", blockedUserID).
		Msg("Creating user block")

	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	blockSetter := &models.UserBlockSetter{
		BlockerPseudonymID: &blockerPseudonymID,
		BlockedPseudonymID: func() *sql.Null[string] {
			v := sql.Null[string]{V: blockedPseudonymID, Valid: true}
			return &v
		}(),
		BlockedUserID: func() *sql.Null[int64] {
			v := sql.Null[int64]{V: blockedUserID, Valid: true}
			return &v
		}(),
		CreatedAt: &now,
	}

	// Use the generated UserBlocks table helper
	userBlock, err := models.UserBlocks.Insert(blockSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create user block: %w", err)
	}

	return userBlock, nil
}

// GetUserBlock retrieves a user block by blocker and blocked pseudonym IDs
func (dao *UserBlocksDAO) GetUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (*models.UserBlock, error) {
	log.Debug().
		Str("blocker_pseudonym_id", blockerPseudonymID).
		Str("blocked_pseudonym_id", blockedPseudonymID).
		Msg("Getting user block")

	// Use the generated UserBlocks table helper with where clause
	blocks, err := models.UserBlocks.Query(
		models.SelectWhere.UserBlocks.BlockerPseudonymID.EQ(blockerPseudonymID),
		models.SelectWhere.UserBlocks.BlockedPseudonymID.EQ(blockedPseudonymID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user block: %w", err)
	}

	if len(blocks) == 0 {
		return nil, nil
	}

	return blocks[0], nil
}

// GetUserBlocksByBlocker retrieves all blocks created by a specific pseudonym
func (dao *UserBlocksDAO) GetUserBlocksByBlocker(ctx context.Context, blockerPseudonymID string) ([]*models.UserBlock, error) {
	log.Debug().
		Str("blocker_pseudonym_id", blockerPseudonymID).
		Msg("Getting user blocks by blocker")

	// Use the generated UserBlocks table helper with where clause
	blocks, err := models.UserBlocks.Query(
		models.SelectWhere.UserBlocks.BlockerPseudonymID.EQ(blockerPseudonymID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user blocks by blocker: %w", err)
	}

	return blocks, nil
}

// GetUserBlocksByBlockedUser retrieves all blocks for a specific user
func (dao *UserBlocksDAO) GetUserBlocksByBlockedUser(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error) {
	log.Debug().
		Int64("blocked_user_id", blockedUserID).
		Msg("Getting user blocks by blocked user")

	// Use the generated UserBlocks table helper with where clause
	blocks, err := models.UserBlocks.Query(
		models.SelectWhere.UserBlocks.BlockedUserID.EQ(blockedUserID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user blocks by blocked user: %w", err)
	}

	return blocks, nil
}

// DeleteUserBlock deletes a user block
func (dao *UserBlocksDAO) DeleteUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) error {
	log.Debug().
		Str("blocker_pseudonym_id", blockerPseudonymID).
		Str("blocked_pseudonym_id", blockedPseudonymID).
		Msg("Deleting user block")

	// First get the block
	block, err := dao.GetUserBlock(ctx, blockerPseudonymID, blockedPseudonymID)
	if err != nil {
		return fmt.Errorf("failed to get user block for deletion: %w", err)
	}
	if block == nil {
		return fmt.Errorf("user block not found")
	}

	// Use the generated Delete method
	err = block.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete user block: %w", err)
	}

	return nil
}

// IsUserBlocked checks if a user is blocked by another user
func (dao *UserBlocksDAO) IsUserBlocked(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (bool, error) {
	log.Debug().
		Str("blocker_pseudonym_id", blockerPseudonymID).
		Str("blocked_pseudonym_id", blockedPseudonymID).
		Msg("Checking if user is blocked")

	block, err := dao.GetUserBlock(ctx, blockerPseudonymID, blockedPseudonymID)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is blocked: %w", err)
	}

	return block != nil, nil
}
