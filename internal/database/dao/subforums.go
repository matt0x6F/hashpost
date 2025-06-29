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
