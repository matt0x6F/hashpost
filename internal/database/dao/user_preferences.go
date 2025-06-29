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

// UserPreferencesDAO provides data access operations for user preferences
type UserPreferencesDAO struct {
	db bob.Executor
}

// NewUserPreferencesDAO creates a new UserPreferencesDAO
func NewUserPreferencesDAO(db bob.Executor) *UserPreferencesDAO {
	return &UserPreferencesDAO{
		db: db,
	}
}

// GetUserPreferences retrieves user preferences by user ID
func (dao *UserPreferencesDAO) GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreference, error) {
	log.Debug().
		Int64("user_id", userID).
		Msg("Getting user preferences")

	// Use the generated FindUserPreference function
	preferences, err := models.FindUserPreference(ctx, dao.db, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	return preferences, nil
}

// CreateUserPreferences creates new user preferences
func (dao *UserPreferencesDAO) CreateUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error) {
	log.Debug().
		Int64("user_id", userID).
		Msg("Creating user preferences")

	// Set the user ID
	preferences.UserID = &userID

	// Set timestamps
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())
	preferences.CreatedAt = &now
	preferences.UpdatedAt = &now

	// Use the generated UserPreferences table helper
	userPreferences, err := models.UserPreferences.Insert(preferences).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create user preferences: %w", err)
	}

	return userPreferences, nil
}

// UpdateUserPreferences updates user preferences
func (dao *UserPreferencesDAO) UpdateUserPreferences(ctx context.Context, userID int64, updates *models.UserPreferenceSetter) error {
	log.Debug().
		Int64("user_id", userID).
		Msg("Updating user preferences")

	// First get the existing preferences
	preferences, err := dao.GetUserPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user preferences for update: %w", err)
	}
	if preferences == nil {
		return fmt.Errorf("user preferences not found")
	}

	// Set updated timestamp
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())
	updates.UpdatedAt = &now

	// Use the generated Update method
	err = preferences.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

// UpsertUserPreferences creates or updates user preferences
func (dao *UserPreferencesDAO) UpsertUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error) {
	log.Debug().
		Int64("user_id", userID).
		Msg("Upserting user preferences")

	// Try to get existing preferences
	existing, err := dao.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user preferences: %w", err)
	}

	if existing == nil {
		// Create new preferences
		return dao.CreateUserPreferences(ctx, userID, preferences)
	} else {
		// Update existing preferences
		err = dao.UpdateUserPreferences(ctx, userID, preferences)
		if err != nil {
			return nil, err
		}
		// Reload the updated preferences
		return dao.GetUserPreferences(ctx, userID)
	}
}
