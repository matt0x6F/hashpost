package dao

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// PseudonymDAO provides data access operations for pseudonyms
type PseudonymDAO struct {
	db bob.Executor
}

// NewPseudonymDAO creates a new PseudonymDAO
func NewPseudonymDAO(db bob.Executor) *PseudonymDAO {
	return &PseudonymDAO{
		db: db,
	}
}

// generatePseudonymID generates a unique pseudonym ID
func generatePseudonymID() string {
	// Generate 32 random bytes and encode as hex
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreatePseudonym creates a new pseudonym
func (dao *PseudonymDAO) CreatePseudonym(ctx context.Context, userID int64, displayName string) (*models.Pseudonym, error) {
	log.Debug().
		Int64("user_id", userID).
		Str("display_name", displayName).
		Msg("Creating pseudonym")

	// Generate a unique pseudonym ID
	pseudonymID := generatePseudonymID()

	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	isActive := sql.Null[bool]{}
	isActive.Scan(true)

	pseudonymSetter := &models.PseudonymSetter{
		PseudonymID: &pseudonymID,
		UserID:      &userID,
		DisplayName: &displayName,
		CreatedAt:   &now,
		IsActive:    &isActive,
	}

	// Use the generated Pseudonyms table helper
	pseudonym, err := models.Pseudonyms.Insert(pseudonymSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create pseudonym: %w", err)
	}

	return pseudonym, nil
}

// GetPseudonymByID retrieves a pseudonym by ID
func (dao *PseudonymDAO) GetPseudonymByID(ctx context.Context, pseudonymID string) (*models.Pseudonym, error) {
	// Use the generated FindPseudonym function
	pseudonym, err := models.FindPseudonym(ctx, dao.db, pseudonymID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pseudonym by ID: %w", err)
	}

	return pseudonym, nil
}

// GetPseudonymsByUserID retrieves all pseudonyms for a user
func (dao *PseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64) ([]*models.Pseudonym, error) {
	// Use the generated Pseudonyms table helper with where clause
	pseudonyms, err := models.Pseudonyms.Query(
		models.SelectWhere.Pseudonyms.UserID.EQ(userID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get pseudonyms by user ID: %w", err)
	}

	return pseudonyms, nil
}

// UpdatePseudonym updates a pseudonym
func (dao *PseudonymDAO) UpdatePseudonym(ctx context.Context, pseudonymID string, updates *models.PseudonymSetter) error {
	// First get the pseudonym
	pseudonym, err := dao.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		return fmt.Errorf("failed to get pseudonym for update: %w", err)
	}
	if pseudonym == nil {
		return fmt.Errorf("pseudonym not found")
	}

	// Use the generated Update method
	err = pseudonym.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update pseudonym: %w", err)
	}

	return nil
}

// DeletePseudonym deletes a pseudonym
func (dao *PseudonymDAO) DeletePseudonym(ctx context.Context, pseudonymID string) error {
	// First get the pseudonym
	pseudonym, err := dao.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		return fmt.Errorf("failed to get pseudonym for deletion: %w", err)
	}
	if pseudonym == nil {
		return fmt.Errorf("pseudonym not found")
	}

	// Use the generated Delete method
	err = pseudonym.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete pseudonym: %w", err)
	}

	return nil
}

// UpdateLastActive updates the pseudonym's last active timestamp
func (dao *PseudonymDAO) UpdateLastActive(ctx context.Context, pseudonymID string) error {
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	updates := &models.PseudonymSetter{
		LastActiveAt: &now,
	}

	return dao.UpdatePseudonym(ctx, pseudonymID, updates)
}

// GetPseudonymByDisplayName retrieves a pseudonym by display name
func (dao *PseudonymDAO) GetPseudonymByDisplayName(ctx context.Context, displayName string) (*models.Pseudonym, error) {
	pseudonyms, err := models.Pseudonyms.Query(
		models.SelectWhere.Pseudonyms.DisplayName.EQ(displayName),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get pseudonym by display name: %w", err)
	}
	if len(pseudonyms) == 0 {
		return nil, nil
	}
	return pseudonyms[0], nil
}
