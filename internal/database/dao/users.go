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

// UserDAO provides data access operations for users
type UserDAO struct {
	db bob.Executor
}

// NewUserDAO creates a new UserDAO
func NewUserDAO(db bob.Executor) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

// CreateUser creates a new user
func (dao *UserDAO) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	log.Debug().
		Str("email", email).
		Msg("Creating user")

	// Create a null time for now
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	// Create a null bool for is_active
	isActive := sql.Null[bool]{}
	isActive.Scan(true)

	userSetter := &models.UserSetter{
		Email:        &email,
		PasswordHash: &passwordHash,
		CreatedAt:    &now,
		IsActive:     &isActive,
	}

	// Use the generated Users table helper
	user, err := models.Users.Insert(userSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (dao *UserDAO) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	// Use the generated FindUser function
	user, err := models.FindUser(ctx, dao.db, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (dao *UserDAO) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Use the generated Users table helper with where clause
	users, err := models.Users.Query(
		models.SelectWhere.Users.Email.EQ(email),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

// UpdateUser updates a user
func (dao *UserDAO) UpdateUser(ctx context.Context, userID int64, updates *models.UserSetter) error {
	// First get the user
	user, err := dao.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user for update: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Use the generated Update method
	err = user.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (dao *UserDAO) DeleteUser(ctx context.Context, userID int64) error {
	// First get the user
	user, err := dao.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user for deletion: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Use the generated Delete method
	err = user.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a list of users with pagination
func (dao *UserDAO) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	// Use the generated Users table helper
	users, err := models.Users.Query(
		models.SelectWhere.Users.UserID.GT(0), // Simple condition to get all users
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Apply pagination manually
	if offset >= len(users) {
		return []*models.User{}, nil
	}

	end := offset + limit
	if end > len(users) {
		end = len(users)
	}

	return users[offset:end], nil
}

// CountUsers counts the total number of users
func (dao *UserDAO) CountUsers(ctx context.Context) (int64, error) {
	// Use the generated Users table helper with count
	count, err := models.Users.Query().Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// UpdateLastActive updates the user's last active timestamp
func (dao *UserDAO) UpdateLastActive(ctx context.Context, userID int64) error {
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	updates := &models.UserSetter{
		LastActiveAt: &now,
	}

	return dao.UpdateUser(ctx, userID, updates)
}

// SuspendUser suspends a user
func (dao *UserDAO) SuspendUser(ctx context.Context, userID int64, reason string, expiresAt *time.Time) error {
	isSuspended := sql.Null[bool]{}
	isSuspended.Scan(true)

	suspensionReason := sql.Null[string]{}
	suspensionReason.Scan(reason)

	updates := &models.UserSetter{
		IsSuspended:      &isSuspended,
		SuspensionReason: &suspensionReason,
	}

	if expiresAt != nil {
		suspensionExpiresAt := sql.Null[time.Time]{}
		suspensionExpiresAt.Scan(*expiresAt)
		updates.SuspensionExpiresAt = &suspensionExpiresAt
	}

	return dao.UpdateUser(ctx, userID, updates)
}

// UnsuspendUser removes suspension from a user
func (dao *UserDAO) UnsuspendUser(ctx context.Context, userID int64) error {
	isSuspended := sql.Null[bool]{}
	isSuspended.Scan(false)

	updates := &models.UserSetter{
		IsSuspended:         &isSuspended,
		SuspensionReason:    &sql.Null[string]{Valid: false},
		SuspensionExpiresAt: &sql.Null[time.Time]{Valid: false},
	}

	return dao.UpdateUser(ctx, userID, updates)
}
