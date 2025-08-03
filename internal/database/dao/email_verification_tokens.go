package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// EmailVerificationTokenDAOInterface defines the interface for email verification token operations
type EmailVerificationTokenDAOInterface interface {
	CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetToken(ctx context.Context, token string) (*models.EmailVerificationToken, error)
	MarkTokenAsUsed(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteTokensByUserID(ctx context.Context, userID int64) error
}

// EmailVerificationTokenDAO implements EmailVerificationTokenDAOInterface
type EmailVerificationTokenDAO struct {
	db bob.Executor
}

// NewEmailVerificationTokenDAO creates a new EmailVerificationTokenDAO
func NewEmailVerificationTokenDAO(db bob.Executor) *EmailVerificationTokenDAO {
	return &EmailVerificationTokenDAO{db: db}
}

// CreateToken creates a new email verification token
func (d *EmailVerificationTokenDAO) CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	tokenSetter := &models.EmailVerificationTokenSetter{
		UserID:    &userID,
		Token:     &token,
		ExpiresAt: &expiresAt,
	}

	_, err := models.EmailVerificationTokens.Insert(tokenSetter).One(ctx, d.db)
	return err
}

// GetToken retrieves a token by its value
func (d *EmailVerificationTokenDAO) GetToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	tokens, err := models.EmailVerificationTokens.Query(
		models.SelectWhere.EmailVerificationTokens.Token.EQ(token),
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
func (d *EmailVerificationTokenDAO) MarkTokenAsUsed(ctx context.Context, token string) error {
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

	updates := &models.EmailVerificationTokenSetter{
		UsedAt: &now,
	}

	err = tokenModel.Update(ctx, d.db, updates)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// DeleteExpiredTokens deletes all expired tokens
func (d *EmailVerificationTokenDAO) DeleteExpiredTokens(ctx context.Context) error {
	// Get all expired tokens
	tokens, err := models.EmailVerificationTokens.Query(
		models.SelectWhere.EmailVerificationTokens.ExpiresAt.LT(time.Now()),
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
func (d *EmailVerificationTokenDAO) DeleteTokensByUserID(ctx context.Context, userID int64) error {
	// Get all tokens for the user
	tokens, err := models.EmailVerificationTokens.Query(
		models.SelectWhere.EmailVerificationTokens.UserID.EQ(userID),
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
