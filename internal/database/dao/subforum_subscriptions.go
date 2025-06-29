package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// SubforumSubscriptionDAO provides data access operations for subforum subscriptions
type SubforumSubscriptionDAO struct {
	db bob.Executor
}

// NewSubforumSubscriptionDAO creates a new SubforumSubscriptionDAO
func NewSubforumSubscriptionDAO(db bob.Executor) *SubforumSubscriptionDAO {
	return &SubforumSubscriptionDAO{
		db: db,
	}
}

// GetSubscription retrieves a subscription by pseudonym ID and subforum ID
func (dao *SubforumSubscriptionDAO) GetSubscription(ctx context.Context, pseudonymID string, subforumID int32) (*models.SubforumSubscription, error) {
	subscription, err := models.SubforumSubscriptions.Query(
		models.SelectWhere.SubforumSubscriptions.PseudonymID.EQ(pseudonymID),
		models.SelectWhere.SubforumSubscriptions.SubforumID.EQ(subforumID),
	).One(ctx, dao.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}

// CreateSubscription creates a new subscription
func (dao *SubforumSubscriptionDAO) CreateSubscription(ctx context.Context, pseudonymID string, subforumID int32, isFavorite bool) (*models.SubforumSubscription, error) {
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	favorite := sql.Null[bool]{}
	favorite.Scan(isFavorite)

	subscriptionSetter := &models.SubforumSubscriptionSetter{
		PseudonymID:  &pseudonymID,
		SubforumID:   &subforumID,
		SubscribedAt: &now,
		IsFavorite:   &favorite,
	}

	subscription, err := models.SubforumSubscriptions.Insert(subscriptionSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return subscription, nil
}

// DeleteSubscription removes a subscription
func (dao *SubforumSubscriptionDAO) DeleteSubscription(ctx context.Context, pseudonymID string, subforumID int32) error {
	_, err := models.SubforumSubscriptions.Delete(
		models.DeleteWhere.SubforumSubscriptions.PseudonymID.EQ(pseudonymID),
		models.DeleteWhere.SubforumSubscriptions.SubforumID.EQ(subforumID),
	).Exec(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

// UpdateSubscriptionFavorite updates the favorite status of a subscription
func (dao *SubforumSubscriptionDAO) UpdateSubscriptionFavorite(ctx context.Context, pseudonymID string, subforumID int32, isFavorite bool) error {
	favorite := sql.Null[bool]{}
	favorite.Scan(isFavorite)

	updates := &models.SubforumSubscriptionSetter{
		IsFavorite: &favorite,
	}

	subscription, err := dao.GetSubscription(ctx, pseudonymID, subforumID)
	if err != nil {
		return fmt.Errorf("failed to find subscription for update: %w", err)
	}
	if subscription == nil {
		return fmt.Errorf("subscription not found")
	}

	err = subscription.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update subscription favorite: %w", err)
	}

	return nil
}

// GetSubscriptionsByPseudonym retrieves all subscriptions for a pseudonym
func (dao *SubforumSubscriptionDAO) GetSubscriptionsByPseudonym(ctx context.Context, pseudonymID string) ([]*models.SubforumSubscription, error) {
	subscriptions, err := models.SubforumSubscriptions.Query(
		models.SelectWhere.SubforumSubscriptions.PseudonymID.EQ(pseudonymID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by pseudonym: %w", err)
	}

	return subscriptions, nil
}

// GetSubscriptionsBySubforum retrieves all subscriptions for a subforum
func (dao *SubforumSubscriptionDAO) GetSubscriptionsBySubforum(ctx context.Context, subforumID int32) ([]*models.SubforumSubscription, error) {
	subscriptions, err := models.SubforumSubscriptions.Query(
		models.SelectWhere.SubforumSubscriptions.SubforumID.EQ(subforumID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by subforum: %w", err)
	}

	return subscriptions, nil
}

// CountSubscriptionsBySubforum counts the number of subscriptions for a subforum
func (dao *SubforumSubscriptionDAO) CountSubscriptionsBySubforum(ctx context.Context, subforumID int32) (int64, error) {
	count, err := models.SubforumSubscriptions.Query(
		models.SelectWhere.SubforumSubscriptions.SubforumID.EQ(subforumID),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count subscriptions by subforum: %w", err)
	}

	return count, nil
}

// IsSubscribed checks if a pseudonym is subscribed to a subforum
func (dao *SubforumSubscriptionDAO) IsSubscribed(ctx context.Context, pseudonymID string, subforumID int32) (bool, error) {
	subscription, err := dao.GetSubscription(ctx, pseudonymID, subforumID)
	if err != nil {
		return false, err
	}
	return subscription != nil, nil
}

// IsFavorite checks if a subscription is marked as favorite
func (dao *SubforumSubscriptionDAO) IsFavorite(ctx context.Context, pseudonymID string, subforumID int32) (bool, error) {
	subscription, err := dao.GetSubscription(ctx, pseudonymID, subforumID)
	if err != nil {
		return false, err
	}
	if subscription == nil {
		return false, nil
	}
	return subscription.IsFavorite.Valid && subscription.IsFavorite.V, nil
}
