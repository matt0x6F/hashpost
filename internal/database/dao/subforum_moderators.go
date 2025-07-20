package dao

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// SubforumModeratorDAO provides data access operations for subforum moderators
type SubforumModeratorDAO struct {
	db bob.Executor
}

// NewSubforumModeratorDAO creates a new SubforumModeratorDAO
func NewSubforumModeratorDAO(db bob.Executor) *SubforumModeratorDAO {
	return &SubforumModeratorDAO{
		db: db,
	}
}

// GetModeratorsBySubforum retrieves all moderators for a subforum
func (dao *SubforumModeratorDAO) GetModeratorsBySubforum(ctx context.Context, subforumID int32) ([]*models.SubforumModerator, error) {
	moderators, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subforum moderators: %w", err)
	}

	// Load pseudonym relationships for all moderators
	if len(moderators) > 0 {
		err = models.SubforumModeratorSlice(moderators).LoadPseudonym(ctx, dao.db)
		if err != nil {
			return nil, fmt.Errorf("failed to load moderator pseudonyms: %w", err)
		}
	}

	return moderators, nil
}

// GetModeratorByPseudonym retrieves a moderator by pseudonym ID and subforum ID
func (dao *SubforumModeratorDAO) GetModeratorByPseudonym(ctx context.Context, pseudonymID string, subforumID int32) (*models.SubforumModerator, error) {
	moderator, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
	).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get moderator by pseudonym: %w", err)
	}

	return moderator, nil
}

// CreateModerator creates a new moderator
func (dao *SubforumModeratorDAO) CreateModerator(ctx context.Context, subforumID int32, pseudonymID, role string, addedByPseudonymID string) (*models.SubforumModerator, error) {
	addedByPseudonymIDNull := sql.Null[string]{V: addedByPseudonymID, Valid: true}

	moderatorSetter := &models.SubforumModeratorSetter{
		SubforumID:         &subforumID,
		PseudonymID:        &pseudonymID,
		Role:               &role,
		AddedByPseudonymID: &addedByPseudonymIDNull,
	}

	moderator, err := models.SubforumModerators.Insert(moderatorSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create moderator: %w", err)
	}

	return moderator, nil
}

// DeleteModerator removes a moderator
func (dao *SubforumModeratorDAO) DeleteModerator(ctx context.Context, pseudonymID string, subforumID int32) error {
	_, err := models.SubforumModerators.Delete(
		models.DeleteWhere.SubforumModerators.PseudonymID.EQ(pseudonymID),
		models.DeleteWhere.SubforumModerators.SubforumID.EQ(subforumID),
	).Exec(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete moderator: %w", err)
	}

	return nil
}

// UpdateModeratorRole updates the role of a moderator
func (dao *SubforumModeratorDAO) UpdateModeratorRole(ctx context.Context, pseudonymID string, subforumID int32, newRole string) error {
	updates := &models.SubforumModeratorSetter{
		Role: &newRole,
	}

	moderator, err := dao.GetModeratorByPseudonym(ctx, pseudonymID, subforumID)
	if err != nil {
		return fmt.Errorf("failed to find moderator for update: %w", err)
	}
	if moderator == nil {
		return fmt.Errorf("moderator not found")
	}

	err = moderator.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update moderator role: %w", err)
	}

	return nil
}
