package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// PasswordResetTokenDAOInterface defines the interface for password reset token operations
type PasswordResetTokenDAOInterface interface {
	CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	MarkTokenAsUsed(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteTokensByUserID(ctx context.Context, userID int64) error
}

// PasswordResetTokenDAO implements PasswordResetTokenDAOInterface
type PasswordResetTokenDAO struct {
	db bob.Executor
}

// NewPasswordResetTokenDAO creates a new PasswordResetTokenDAO
func NewPasswordResetTokenDAO(db bob.Executor) *PasswordResetTokenDAO {
	return &PasswordResetTokenDAO{db: db}
}

// CreateToken creates a new password reset token
func (d *PasswordResetTokenDAO) CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	tokenSetter := &models.PasswordResetTokenSetter{
		UserID:    &userID,
		Token:     &token,
		ExpiresAt: &expiresAt,
	}

	_, err := models.PasswordResetTokens.Insert(tokenSetter).One(ctx, d.db)
	return err
}

// GetToken retrieves a token by its value
func (d *PasswordResetTokenDAO) GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	tokens, err := models.PasswordResetTokens.Query(
		models.SelectWhere.PasswordResetTokens.Token.EQ(token),
	).All(ctx, d.db)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	return tokens[0], nil
}

// MarkTokenAsUsed marks a token as used
func (d *PasswordResetTokenDAO) MarkTokenAsUsed(ctx context.Context, token string) error {
	// First get the token
	tokenModel, err := d.GetToken(ctx, token)
	if err != nil {
		return err
	}
	if tokenModel == nil {
		return fmt.Errorf("token not found")
	}

	// Mark as used
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	updates := &models.PasswordResetTokenSetter{
		UsedAt: &now,
	}

	err = tokenModel.Update(ctx, d.db, updates)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// DeleteExpiredTokens deletes all expired tokens
func (d *PasswordResetTokenDAO) DeleteExpiredTokens(ctx context.Context) error {
	// Get all expired tokens
	tokens, err := models.PasswordResetTokens.Query(
		models.SelectWhere.PasswordResetTokens.ExpiresAt.LT(time.Now()),
	).All(ctx, d.db)
	if err != nil {
		return fmt.Errorf("failed to get expired tokens: %w", err)
	}

	// Delete each token
	for _, token := range tokens {
		err = token.Delete(ctx, d.db)
		if err != nil {
			return fmt.Errorf("failed to delete expired token: %w", err)
		}
	}

	return nil
}

// DeleteTokensByUserID deletes all tokens for a specific user
func (d *PasswordResetTokenDAO) DeleteTokensByUserID(ctx context.Context, userID int64) error {
	// Get all tokens for the user
	tokens, err := models.PasswordResetTokens.Query(
		models.SelectWhere.PasswordResetTokens.UserID.EQ(userID),
	).All(ctx, d.db)
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Delete each token
	for _, token := range tokens {
		err = token.Delete(ctx, d.db)
		if err != nil {
			return fmt.Errorf("failed to delete user token: %w", err)
		}
	}

	return nil
}
