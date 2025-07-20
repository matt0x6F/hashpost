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

// ModerationActionDAO implements ModerationActionDAOInterface
type ModerationActionDAO struct {
	db bob.Executor
}

// NewModerationActionDAO creates a new ModerationActionDAO
func NewModerationActionDAO(db bob.Executor) *ModerationActionDAO {
	return &ModerationActionDAO{db: db}
}

// CreateModerationAction creates a new moderation action
func (dao *ModerationActionDAO) CreateModerationAction(ctx context.Context, action *models.ModerationActionSetter) (*models.ModerationAction, error) {
	log.Debug().
		Int64("moderator_user_id", *action.ModeratorUserID).
		Str("moderator_pseudonym_id", *action.ModeratorPseudonymID).
		Str("action_type", *action.ActionType).
		Msg("Creating moderation action")

	// Set default values if not provided
	if action.CreatedAt == nil {
		now := sql.Null[time.Time]{}
		now.Scan(time.Now())
		action.CreatedAt = &now
	}

	// Use the generated ModerationActions table helper
	createdAction, err := models.ModerationActions.Insert(action).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create moderation action: %w", err)
	}

	return createdAction, nil
}

// GetModerationActionByID retrieves a moderation action by ID
func (dao *ModerationActionDAO) GetModerationActionByID(ctx context.Context, actionID int64) (*models.ModerationAction, error) {
	// Use the generated FindModerationAction function
	action, err := models.FindModerationAction(ctx, dao.db, actionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get moderation action by ID: %w", err)
	}

	return action, nil
}

// GetModerationActions retrieves moderation actions with filtering and pagination
func (dao *ModerationActionDAO) GetModerationActions(ctx context.Context, actionType string, page, limit int) ([]*models.ModerationAction, error) {
	offset := page * limit

	// Build query with optional action type filter
	var actions []*models.ModerationAction
	var err error

	if actionType != "" {
		actions, err = models.ModerationActions.Query(
			models.SelectWhere.ModerationActions.ActionType.EQ(actionType),
			sm.Limit(limit),
			sm.Offset(offset),
		).All(ctx, dao.db)
	} else {
		actions, err = models.ModerationActions.Query(
			sm.Limit(limit),
			sm.Offset(offset),
		).All(ctx, dao.db)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get moderation actions: %w", err)
	}

	return actions, nil
}

// CountModerationActions counts moderation actions with optional action type filter
func (dao *ModerationActionDAO) CountModerationActions(ctx context.Context, actionType string) (int64, error) {
	// Build query with optional action type filter
	var count int64
	var err error

	if actionType != "" {
		count, err = models.ModerationActions.Query(
			models.SelectWhere.ModerationActions.ActionType.EQ(actionType),
		).Count(ctx, dao.db)
	} else {
		count, err = models.ModerationActions.Query().Count(ctx, dao.db)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to count moderation actions: %w", err)
	}

	return count, nil
}

// GetModerationActionsByModerator retrieves moderation actions by a specific moderator
func (dao *ModerationActionDAO) GetModerationActionsByModerator(ctx context.Context, moderatorUserID int64, page, limit int) ([]*models.ModerationAction, error) {
	offset := page * limit

	// Use the generated ModerationActions table helper with where clause and pagination
	actions, err := models.ModerationActions.Query(
		models.SelectWhere.ModerationActions.ModeratorUserID.EQ(moderatorUserID),
		sm.Limit(limit),
		sm.Offset(offset),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get moderation actions by moderator: %w", err)
	}

	return actions, nil
}
