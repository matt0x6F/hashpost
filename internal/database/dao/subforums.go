package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
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

// GetSubforumByCommunityTypeAndName retrieves a subforum by community type and name
func (dao *SubforumDAO) GetSubforumByCommunityTypeAndName(ctx context.Context, communityType, name string) (*models.Subforum, error) {
	subforums, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(name),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subforum by community type and name: %w", err)
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

// ListSubforumsByCommunityType retrieves subforums by community type
func (dao *SubforumDAO) ListSubforumsByCommunityType(ctx context.Context, communityType string) ([]*models.Subforum, error) {
	subforums, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list subforums by community type: %w", err)
	}

	return subforums, nil
}

// CreateSubforum creates a new subforum with community type support
func (dao *SubforumDAO) CreateSubforum(ctx context.Context, name, displayName, description, sidebarText, rulesText, communityType, governanceStyle string, isNSFW, isPrivate, isRestricted bool, ownerPseudonymID string) (*models.Subforum, error) {
	// Check if subforum with this name and community type already exists
	existing, err := dao.GetSubforumByCommunityTypeAndName(ctx, communityType, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing subforum: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("subforum with name '%s' already exists in community type '%s'", name, communityType)
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

	ownerPseudonymIDNull := sql.Null[string]{}
	if ownerPseudonymID != "" {
		ownerPseudonymIDNull.Scan(ownerPseudonymID)
	}

	// Create the subforum using the setter pattern
	subforumSetter := &models.SubforumSetter{
		Name:             &name,
		DisplayName:      &displayName,
		Description:      &descriptionNull,
		SidebarText:      &sidebarTextNull,
		RulesText:        &rulesTextNull,
		CommunityType:    &communityType,
		GovernanceStyle:  &governanceStyle,
		IsNSFW:           &isNSFWNull,
		IsPrivate:        &isPrivateNull,
		IsRestricted:     &isRestrictedNull,
		SubscriberCount:  &subscriberCountNull,
		PostCount:        &postCountNull,
		CreatedAt:        &createdAtNull,
		OwnerPseudonymID: &ownerPseudonymIDNull,
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

// UpdateSubscriberCount updates the subscriber count for a subforum
func (dao *SubforumDAO) UpdateSubscriberCount(ctx context.Context, subforumID int32, subscriberCount int32) error {
	// Get the current subforum
	subforum, err := dao.GetSubforumByID(ctx, subforumID)
	if err != nil {
		return fmt.Errorf("failed to get subforum for subscriber count update: %w", err)
	}
	if subforum == nil {
		return fmt.Errorf("subforum not found")
	}

	// Create the update setter
	subscriberCountNull := sql.Null[int32]{}
	subscriberCountNull.Scan(subscriberCount)

	updatedAtNull := sql.Null[time.Time]{}
	updatedAtNull.Scan(time.Now())

	updateSetter := &models.SubforumSetter{
		SubscriberCount: &subscriberCountNull,
		UpdatedAt:       &updatedAtNull,
	}

	// Update the subforum
	err = subforum.Update(ctx, dao.db, updateSetter)
	if err != nil {
		return fmt.Errorf("failed to update subforum subscriber count: %w", err)
	}

	return nil
}

// UpdateSettings updates subforum settings
func (dao *SubforumDAO) UpdateSettings(ctx context.Context, subforumID int32, allowImages, allowVideos, allowPolls, requireFlair, isPrivate, isRestricted, isNSFW bool, minimumAccountAgeHours, minimumKarmaRequired int, description, sidebarText string) error {
	// Get the subforum first
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.SubforumID.EQ(subforumID),
	).One(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to get subforum: %w", err)
	}

	// Create update setter
	updateSetter := &models.SubforumSetter{
		AllowImages:            &sql.Null[bool]{V: allowImages, Valid: true},
		AllowVideos:            &sql.Null[bool]{V: allowVideos, Valid: true},
		AllowPolls:             &sql.Null[bool]{V: allowPolls, Valid: true},
		RequireFlair:           &sql.Null[bool]{V: requireFlair, Valid: true},
		MinimumAccountAgeHours: &sql.Null[int32]{V: int32(minimumAccountAgeHours), Valid: true},
		MinimumKarmaRequired:   &sql.Null[int32]{V: int32(minimumKarmaRequired), Valid: true},
		IsPrivate:              &sql.Null[bool]{V: isPrivate, Valid: true},
		IsRestricted:           &sql.Null[bool]{V: isRestricted, Valid: true},
		IsNSFW:                 &sql.Null[bool]{V: isNSFW, Valid: true},
		Description:            &sql.Null[string]{V: description, Valid: true},
		SidebarText:            &sql.Null[string]{V: sidebarText, Valid: true},
		UpdatedAt:              &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Update the subforum
	err = subforum.Update(ctx, dao.db, updateSetter)
	if err != nil {
		return fmt.Errorf("failed to update subforum settings: %w", err)
	}

	return nil
}

// UpdateRules updates subforum rules
func (dao *SubforumDAO) UpdateRules(ctx context.Context, subforumID int32, rules []byte) error {
	// Get the subforum first
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.SubforumID.EQ(subforumID),
	).One(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to get subforum: %w", err)
	}

	// Create update setter
	updateSetter := &models.SubforumSetter{
		SubforumRules: &sql.Null[types.JSON[json.RawMessage]]{V: types.NewJSON[json.RawMessage](rules), Valid: true},
		UpdatedAt:     &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Update the subforum
	err = subforum.Update(ctx, dao.db, updateSetter)
	if err != nil {
		return fmt.Errorf("failed to update subforum rules: %w", err)
	}

	return nil
}
