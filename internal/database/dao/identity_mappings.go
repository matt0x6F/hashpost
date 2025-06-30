package dao

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// IdentityMappingDAO provides database operations for identity mappings
type IdentityMappingDAO struct {
	db bob.Executor
}

// NewIdentityMappingDAO creates a new identity mapping DAO
func NewIdentityMappingDAO(db bob.Executor) *IdentityMappingDAO {
	return &IdentityMappingDAO{
		db: db,
	}
}

// GetIdentityMappingByPseudonymID retrieves an identity mapping by pseudonym ID
func (dao *IdentityMappingDAO) GetIdentityMappingByPseudonymID(ctx context.Context, pseudonymID string) (*models.IdentityMapping, error) {
	return models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).One(ctx, dao.db)
}

// GetIdentityMappingsByPseudonymID retrieves all identity mappings for a pseudonym ID
func (dao *IdentityMappingDAO) GetIdentityMappingsByPseudonymID(ctx context.Context, pseudonymID string) (models.IdentityMappingSlice, error) {
	return models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
}

// GetIdentityMappingsByFingerprint retrieves all identity mappings for a given fingerprint
func (dao *IdentityMappingDAO) GetIdentityMappingsByFingerprint(ctx context.Context, fingerprint string) (models.IdentityMappingSlice, error) {
	return models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.Fingerprint.EQ(fingerprint),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
}

// GetIdentityMappingsByUserID retrieves all identity mappings for a given user ID
func (dao *IdentityMappingDAO) GetIdentityMappingsByUserID(ctx context.Context, userID int64) (models.IdentityMappingSlice, error) {
	return models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.UserID.EQ(userID),
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
}

// GetAllActiveIdentityMappings retrieves all active identity mappings
func (dao *IdentityMappingDAO) GetAllActiveIdentityMappings(ctx context.Context) (models.IdentityMappingSlice, error) {
	return models.IdentityMappings.Query(
		models.SelectWhere.IdentityMappings.IsActive.EQ(true),
	).All(ctx, dao.db)
}

// CreateIdentityMapping creates a new identity mapping
func (dao *IdentityMappingDAO) CreateIdentityMapping(ctx context.Context, mapping *models.IdentityMappingSetter) (*models.IdentityMapping, error) {
	return models.IdentityMappings.Insert(mapping).One(ctx, dao.db)
}

// UpdateIdentityMapping updates an existing identity mapping
func (dao *IdentityMappingDAO) UpdateIdentityMapping(ctx context.Context, mappingID string, updates *models.IdentityMappingSetter) error {
	// Use a direct update query approach
	_, err := models.IdentityMappings.Update(
		updates.UpdateMod(),
		models.UpdateWhere.IdentityMappings.PseudonymID.EQ(mappingID),
	).One(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to update identity mapping: %w", err)
	}

	return nil
}

// DeactivateIdentityMapping marks an identity mapping as inactive
func (dao *IdentityMappingDAO) DeactivateIdentityMapping(ctx context.Context, mappingID string) error {
	isActive := sql.Null[bool]{}
	isActive.Scan(false)

	setter := &models.IdentityMappingSetter{
		IsActive: &isActive,
	}
	return dao.UpdateIdentityMapping(ctx, mappingID, setter)
}

// GetCorrelationData retrieves correlation data for a given pseudonym
// This includes all pseudonyms that share the same fingerprint
func (dao *IdentityMappingDAO) GetCorrelationData(ctx context.Context, pseudonymID string) (*models.IdentityMapping, models.IdentityMappingSlice, error) {
	// Get the requested identity mapping
	requestedMapping, err := dao.GetIdentityMappingByPseudonymID(ctx, pseudonymID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get requested identity mapping: %w", err)
	}
	if requestedMapping == nil {
		return nil, nil, fmt.Errorf("identity mapping not found for pseudonym: %s", pseudonymID)
	}

	// Get all identity mappings with the same fingerprint
	relatedMappings, err := dao.GetIdentityMappingsByFingerprint(ctx, requestedMapping.Fingerprint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get related identity mappings: %w", err)
	}

	return requestedMapping, relatedMappings, nil
}
