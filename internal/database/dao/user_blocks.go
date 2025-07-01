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
		CreatedAt:          &now,
	}

	// Set either blockedPseudonymID or blockedUserID, but not both (database constraint)
	if blockedPseudonymID != "" {
		blockSetter.BlockedPseudonymID = &sql.Null[string]{V: blockedPseudonymID, Valid: true}
		blockSetter.BlockedUserID = &sql.Null[int64]{Valid: false}
	} else if blockedUserID != 0 {
		blockSetter.BlockedPseudonymID = &sql.Null[string]{Valid: false}
		blockSetter.BlockedUserID = &sql.Null[int64]{V: blockedUserID, Valid: true}
	} else {
		return nil, fmt.Errorf("either blockedPseudonymID or blockedUserID must be provided")
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

// getUserIDByPseudonym gets the user ID for a pseudonym using identity mapping
func (dao *UserBlocksDAO) getUserIDByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	log.Debug().
		Str("pseudonym_id", pseudonymID).
		Msg("Getting user ID by pseudonym")

	// Get the identity mapping for this pseudonym
	mapping, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get identity mapping: %w", err)
	}
	if mapping == nil {
		return 0, fmt.Errorf("identity mapping not found for pseudonym")
	}

	return mapping.UserID, nil
}

// getPseudonymIDsByUserID gets all pseudonym IDs for a user using identity mappings
func (dao *UserBlocksDAO) getPseudonymIDsByUserID(ctx context.Context, userID int64) ([]string, error) {
	log.Debug().
		Int64("user_id", userID).
		Msg("Getting pseudonym IDs by user ID")

	// Get all identity mappings for this user
	mappings, err := models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity mappings: %w", err)
	}

	// Extract unique pseudonym IDs
	pseudonymMap := make(map[string]bool)
	var pseudonymIDs []string
	for _, mapping := range mappings {
		if !pseudonymMap[mapping.PseudonymID] {
			pseudonymMap[mapping.PseudonymID] = true
			pseudonymIDs = append(pseudonymIDs, mapping.PseudonymID)
		}
	}

	return pseudonymIDs, nil
}

// IsUserBlocked checks if a user is blocked by another user
func (dao *UserBlocksDAO) IsUserBlocked(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (bool, error) {
	log.Debug().
		Str("blocker_pseudonym_id", blockerPseudonymID).
		Str("blocked_pseudonym_id", blockedPseudonymID).
		Msg("Checking if user is blocked")

	// Check for direct pseudonym-to-pseudonym block
	block, err := dao.GetUserBlock(ctx, blockerPseudonymID, blockedPseudonymID)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is blocked: %w", err)
	}

	if block != nil {
		return true, nil
	}

	// If no direct block found, check for fingerprint-level blocks
	// This requires getting the user ID for the blocked pseudonym
	// Get the user ID for the blocked pseudonym using identity mapping
	blockedUserID, err := dao.getUserIDByPseudonym(ctx, blockedPseudonymID)
	if err != nil {
		log.Warn().
			Err(err).
			Str("blocked_pseudonym_id", blockedPseudonymID).
			Msg("Failed to get user ID for blocked pseudonym, skipping fingerprint-level check")
		return false, nil
	}

	// Check for fingerprint-level block (blocks all personas of the user)
	fingerprintBlocked, err := dao.IsUserBlockedAtFingerprintLevel(ctx, blockerPseudonymID, blockedUserID)
	if err != nil {
		log.Warn().
			Err(err).
			Str("blocker_pseudonym_id", blockerPseudonymID).
			Int64("blocked_user_id", blockedUserID).
			Msg("Failed to check fingerprint-level block")
		return false, nil
	}

	return fingerprintBlocked, nil
}

// IsPseudonymBlockedByUser checks if a pseudonym is blocked by checking both direct blocks and fingerprint-level blocks
func (dao *UserBlocksDAO) IsPseudonymBlockedByUser(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (bool, error) {
	log.Debug().
		Str("blocker_pseudonym_id", blockerPseudonymID).
		Str("blocked_pseudonym_id", blockedPseudonymID).
		Int64("blocked_user_id", blockedUserID).
		Msg("Checking if pseudonym is blocked by user")

	// Check for direct pseudonym-to-pseudonym block
	directBlock, err := dao.GetUserBlock(ctx, blockerPseudonymID, blockedPseudonymID)
	if err != nil {
		return false, fmt.Errorf("failed to check direct block: %w", err)
	}

	if directBlock != nil {
		return true, nil
	}

	// Check for fingerprint-level block (blocks all personas of the user)
	fingerprintBlocked, err := dao.IsUserBlockedAtFingerprintLevel(ctx, blockerPseudonymID, blockedUserID)
	if err != nil {
		return false, fmt.Errorf("failed to check fingerprint-level block: %w", err)
	}

	return fingerprintBlocked, nil
}

// IsUserBlockedAtFingerprintLevel checks if a user is blocked at the fingerprint level (all personas)
func (dao *UserBlocksDAO) IsUserBlockedAtFingerprintLevel(ctx context.Context, blockerPseudonymID string, blockedUserID int64) (bool, error) {
	// Look for a block where blocked_pseudonym_id is NULL but blocked_user_id matches
	blocks, err := models.UserBlocks.Query(
		models.SelectWhere.UserBlocks.BlockerPseudonymID.EQ(blockerPseudonymID),
		models.SelectWhere.UserBlocks.BlockedUserID.EQ(blockedUserID),
		models.SelectWhere.UserBlocks.BlockedPseudonymID.IsNull(),
	).All(ctx, dao.db)

	if err != nil {
		return false, fmt.Errorf("failed to query fingerprint-level blocks: %w", err)
	}

	return len(blocks) > 0, nil
}

// IsUserBlockedByAnyPseudonym checks if a user is blocked by any pseudonym of the blocker
func (dao *UserBlocksDAO) IsUserBlockedByAnyPseudonym(ctx context.Context, blockerUserID int64, blockedPseudonymID string) (bool, error) {
	log.Debug().
		Int64("blocker_user_id", blockerUserID).
		Str("blocked_pseudonym_id", blockedPseudonymID).
		Msg("Checking if user is blocked by any pseudonym of blocker")

	// Get all pseudonyms for the blocker user using identity mappings
	blockerPseudonymIDs, err := dao.getPseudonymIDsByUserID(ctx, blockerUserID)
	if err != nil {
		log.Warn().
			Err(err).
			Int64("blocker_user_id", blockerUserID).
			Msg("Failed to get pseudonyms for blocker user, skipping block check")
		return false, nil
	}

	// Check if any of the blocker's pseudonyms block the target pseudonym
	for _, blockerPseudonymID := range blockerPseudonymIDs {
		blocked, err := dao.IsUserBlocked(ctx, blockerPseudonymID, blockedPseudonymID)
		if err != nil {
			log.Warn().
				Err(err).
				Str("blocker_pseudonym_id", blockerPseudonymID).
				Str("blocked_pseudonym_id", blockedPseudonymID).
				Msg("Failed to check block between pseudonyms")
			continue
		}
		if blocked {
			return true, nil
		}
	}

	return false, nil
}

// GetFingerprintLevelBlocks retrieves all fingerprint-level blocks for a user
func (dao *UserBlocksDAO) GetFingerprintLevelBlocks(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error) {
	log.Debug().
		Int64("blocked_user_id", blockedUserID).
		Msg("Getting fingerprint-level blocks for user")

	blocks, err := models.UserBlocks.Query(
		models.SelectWhere.UserBlocks.BlockedUserID.EQ(blockedUserID),
		models.SelectWhere.UserBlocks.BlockedPseudonymID.IsNull(),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get fingerprint-level blocks: %w", err)
	}

	return blocks, nil
}

// DeleteUserBlockByID deletes a user block by its ID
func (dao *UserBlocksDAO) DeleteUserBlockByID(ctx context.Context, blockID int64) error {
	log.Debug().
		Int64("block_id", blockID).
		Msg("Deleting user block by ID")

	// Get the block by ID
	block, err := models.UserBlocks.Query(
		models.SelectWhere.UserBlocks.BlockID.EQ(blockID),
	).One(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to get user block by ID: %w", err)
	}
	if block == nil {
		return fmt.Errorf("user block not found")
	}

	// Delete the block
	err = block.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete user block: %w", err)
	}

	return nil
}
