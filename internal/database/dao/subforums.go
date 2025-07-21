package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// SubforumDAO provides data access operations for subforums
type SubforumDAO struct {
	db bob.Executor
}

// NewSubforumDAO creates a new SubforumDAO
func NewSubforumDAO(db bob.Executor) *SubforumDAO {
	return &SubforumDAO{
		db: db,
	}
}

// GetSubforumByName retrieves a subforum by name
func (dao *SubforumDAO) GetSubforumByName(ctx context.Context, name string) (*models.Subforum, error) {
	subforums, err := models.Subforums.Query(
		models.SelectWhere.Subforums.Name.EQ(name),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subforum by name: %w", err)
	}

	if len(subforums) == 0 {
		return nil, nil
	}

	return subforums[0], nil
}

// GetSubforumByID retrieves a subforum by ID
func (dao *SubforumDAO) GetSubforumByID(ctx context.Context, subforumID int32) (*models.Subforum, error) {
	subforum, err := models.FindSubforum(ctx, dao.db, subforumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get subforum by ID: %w", err)
	}

	return subforum, nil
}

// ListSubforums retrieves a list of subforums
func (dao *SubforumDAO) ListSubforums(ctx context.Context) ([]*models.Subforum, error) {
	subforums, err := models.Subforums.Query().All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list subforums: %w", err)
	}

	return subforums, nil
}

// GetSubforumModerators retrieves moderators for a subforum
func (dao *SubforumDAO) GetSubforumModerators(ctx context.Context, subforumID int32) ([]*models.SubforumModerator, error) {
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

// AddSubforumModerator adds a pseudonym as a moderator to a subforum
func (dao *SubforumDAO) AddSubforumModerator(ctx context.Context, subforumID int32, pseudonymID string) error {
	// Check if moderator already exists
	existing, err := models.SubforumModerators.Query(
		models.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
		models.SelectWhere.SubforumModerators.PseudonymID.EQ(pseudonymID),
	).One(ctx, dao.db)
	if err == nil && existing != nil {
		return fmt.Errorf("moderator already exists")
	}

	// Create the moderator
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	role := "moderator" // Default role

	moderatorSetter := &models.SubforumModeratorSetter{
		SubforumID:  &subforumID,
		PseudonymID: &pseudonymID,
		Role:        &role,
		AddedAt:     &now,
	}

	_, err = models.SubforumModerators.Insert(moderatorSetter).One(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to add moderator: %w", err)
	}

	return nil
}

// CreateSubforum creates a new subforum
func (dao *SubforumDAO) CreateSubforum(ctx context.Context, name, displayName, description, sidebarText, rulesText string, isNSFW, isPrivate, isRestricted bool) (*models.Subforum, error) {
	// Check if subforum with this name already exists
	existing, err := dao.GetSubforumByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing subforum: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("subforum with name '%s' already exists", name)
	}

	// Create null types for optional fields
	descriptionNull := sql.Null[string]{}
	descriptionNull.Scan(description)

	sidebarTextNull := sql.Null[string]{}
	sidebarTextNull.Scan(sidebarText)

	rulesTextNull := sql.Null[string]{}
	rulesTextNull.Scan(rulesText)

	isNSFWNull := sql.Null[bool]{}
	isNSFWNull.Scan(isNSFW)

	isPrivateNull := sql.Null[bool]{}
	isPrivateNull.Scan(isPrivate)

	isRestrictedNull := sql.Null[bool]{}
	isRestrictedNull.Scan(isRestricted)

	subscriberCountNull := sql.Null[int32]{}
	subscriberCountNull.Scan(0)

	postCountNull := sql.Null[int32]{}
	postCountNull.Scan(0)

	createdAtNull := sql.Null[time.Time]{}
	createdAtNull.Scan(time.Now())

	// Create the subforum using the setter pattern
	subforumSetter := &models.SubforumSetter{
		Name:            &name,
		DisplayName:     &displayName,
		Description:     &descriptionNull,
		SidebarText:     &sidebarTextNull,
		RulesText:       &rulesTextNull,
		IsNSFW:          &isNSFWNull,
		IsPrivate:       &isPrivateNull,
		IsRestricted:    &isRestrictedNull,
		SubscriberCount: &subscriberCountNull,
		PostCount:       &postCountNull,
		CreatedAt:       &createdAtNull,
	}

	// Insert into database using the generated table helper
	subforum, err := models.Subforums.Insert(subforumSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create subforum: %w", err)
	}

	return subforum, nil
}

// UpdatePostCount updates the post count for a subforum
func (dao *SubforumDAO) UpdatePostCount(ctx context.Context, subforumID int32, postCount int32) error {
	// Get the current subforum
	subforum, err := dao.GetSubforumByID(ctx, subforumID)
	if err != nil {
		return fmt.Errorf("failed to get subforum for post count update: %w", err)
	}
	if subforum == nil {
		return fmt.Errorf("subforum not found")
	}

	// Create the update setter
	postCountNull := sql.Null[int32]{}
	postCountNull.Scan(postCount)

	updatedAtNull := sql.Null[time.Time]{}
	updatedAtNull.Scan(time.Now())

	updateSetter := &models.SubforumSetter{
		PostCount: &postCountNull,
		UpdatedAt: &updatedAtNull,
	}

	// Update the subforum
	err = subforum.Update(ctx, dao.db, updateSetter)
	if err != nil {
		return fmt.Errorf("failed to update subforum post count: %w", err)
	}

	return nil
}
