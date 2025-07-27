package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// DirectMessageDAOInterface defines the interface for direct message data access operations
type DirectMessageDAOInterface interface {
	CreateDirectMessage(ctx context.Context, senderPseudonymID, recipientPseudonymID, content string) (*models.DirectMessage, error)
	GetDirectMessageByID(ctx context.Context, messageID int64) (*models.DirectMessage, error)
	GetDirectMessagesByPseudonym(ctx context.Context, pseudonymID string, page, limit int) ([]*models.DirectMessage, error)
	GetDirectMessagesBetweenPseudonyms(ctx context.Context, pseudonymID1, pseudonymID2 string, page, limit int) ([]*models.DirectMessage, error)
	CountDirectMessagesByPseudonym(ctx context.Context, pseudonymID string) (int64, error)
	MarkMessageAsRead(ctx context.Context, messageID int64) error
	DeleteDirectMessage(ctx context.Context, messageID int64) error
	IsUserBlocked(ctx context.Context, senderPseudonymID, recipientPseudonymID string) (bool, error)
}

// DirectMessageDAO implements DirectMessageDAOInterface
type DirectMessageDAO struct {
	db            bob.Executor
	userBlocksDAO *UserBlocksDAO
}

// NewDirectMessageDAO creates a new direct message DAO
func NewDirectMessageDAO(db bob.Executor) *DirectMessageDAO {
	return &DirectMessageDAO{
		db:            db,
		userBlocksDAO: NewUserBlocksDAO(db),
	}
}

// CreateDirectMessage creates a new direct message
func (dao *DirectMessageDAO) CreateDirectMessage(ctx context.Context, senderPseudonymID, recipientPseudonymID, content string) (*models.DirectMessage, error) {
	// Check if sender is blocked by recipient
	isBlocked, err := dao.IsUserBlocked(ctx, senderPseudonymID, recipientPseudonymID)
	if err != nil {
		return nil, err
	}
	if isBlocked {
		return nil, fmt.Errorf("cannot send message: user is blocked")
	}

	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	messageSetter := &models.DirectMessageSetter{
		SenderPseudonymID:    &senderPseudonymID,
		RecipientPseudonymID: &recipientPseudonymID,
		Content:              &content,
		IsRead:               &sql.Null[bool]{V: false, Valid: true},
		CreatedAt:            &now,
	}

	message, err := models.DirectMessages.Insert(messageSetter).One(ctx, dao.db)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetDirectMessageByID retrieves a direct message by ID
func (dao *DirectMessageDAO) GetDirectMessageByID(ctx context.Context, messageID int64) (*models.DirectMessage, error) {
	message, err := models.FindDirectMessage(ctx, dao.db, messageID)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetDirectMessagesByPseudonym retrieves direct messages for a pseudonym
func (dao *DirectMessageDAO) GetDirectMessagesByPseudonym(ctx context.Context, pseudonymID string, page, limit int) ([]*models.DirectMessage, error) {
	offset := (page - 1) * limit

	// Get sent messages
	sentMessages, err := models.DirectMessages.Query(
		models.SelectWhere.DirectMessages.SenderPseudonymID.EQ(pseudonymID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, err
	}

	// Get received messages
	receivedMessages, err := models.DirectMessages.Query(
		models.SelectWhere.DirectMessages.RecipientPseudonymID.EQ(pseudonymID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, err
	}

	// Combine messages
	allMessages := append(sentMessages, receivedMessages...)

	// Apply pagination manually
	if offset >= len(allMessages) {
		return []*models.DirectMessage{}, nil
	}

	end := offset + limit
	if end > len(allMessages) {
		end = len(allMessages)
	}

	return allMessages[offset:end], nil
}

// GetDirectMessagesBetweenPseudonyms retrieves direct messages between two specific pseudonyms
func (dao *DirectMessageDAO) GetDirectMessagesBetweenPseudonyms(ctx context.Context, pseudonymID1, pseudonymID2 string, page, limit int) ([]*models.DirectMessage, error) {
	offset := (page - 1) * limit

	// Get messages from pseudonym1 to pseudonym2
	messages1to2, err := models.DirectMessages.Query(
		models.SelectWhere.DirectMessages.SenderPseudonymID.EQ(pseudonymID1),
		models.SelectWhere.DirectMessages.RecipientPseudonymID.EQ(pseudonymID2),
	).All(ctx, dao.db)
	if err != nil {
		return nil, err
	}

	// Get messages from pseudonym2 to pseudonym1
	messages2to1, err := models.DirectMessages.Query(
		models.SelectWhere.DirectMessages.SenderPseudonymID.EQ(pseudonymID2),
		models.SelectWhere.DirectMessages.RecipientPseudonymID.EQ(pseudonymID1),
	).All(ctx, dao.db)
	if err != nil {
		return nil, err
	}

	// Combine messages
	allMessages := append(messages1to2, messages2to1...)

	// Apply pagination manually
	if offset >= len(allMessages) {
		return []*models.DirectMessage{}, nil
	}

	end := offset + limit
	if end > len(allMessages) {
		end = len(allMessages)
	}

	return allMessages[offset:end], nil
}

// CountDirectMessagesByPseudonym counts direct messages for a pseudonym
func (dao *DirectMessageDAO) CountDirectMessagesByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	// Count sent messages
	sentCount, err := models.DirectMessages.Query(
		models.SelectWhere.DirectMessages.SenderPseudonymID.EQ(pseudonymID),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, err
	}

	// Count received messages
	receivedCount, err := models.DirectMessages.Query(
		models.SelectWhere.DirectMessages.RecipientPseudonymID.EQ(pseudonymID),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, err
	}

	return sentCount + receivedCount, nil
}

// MarkMessageAsRead marks a message as read
func (dao *DirectMessageDAO) MarkMessageAsRead(ctx context.Context, messageID int64) error {
	message, err := models.FindDirectMessage(ctx, dao.db, messageID)
	if err != nil {
		return err
	}

	updates := &models.DirectMessageSetter{
		IsRead: &sql.Null[bool]{V: true, Valid: true},
	}

	return message.Update(ctx, dao.db, updates)
}

// DeleteDirectMessage deletes a direct message
func (dao *DirectMessageDAO) DeleteDirectMessage(ctx context.Context, messageID int64) error {
	message, err := models.FindDirectMessage(ctx, dao.db, messageID)
	if err != nil {
		return err
	}

	return message.Delete(ctx, dao.db)
}

// IsUserBlocked checks if a user is blocked by another user
func (dao *DirectMessageDAO) IsUserBlocked(ctx context.Context, senderPseudonymID, recipientPseudonymID string) (bool, error) {
	// Use the existing UserBlocksDAO to check if the sender is blocked by the recipient
	// This checks if the recipient has blocked the sender (either at pseudonym level or fingerprint level)
	return dao.userBlocksDAO.IsUserBlocked(ctx, recipientPseudonymID, senderPseudonymID)
}
