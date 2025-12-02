package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// MessageMentionRepository defines methods for managing message mentions
type MessageMentionRepository interface {
	// Create a single mention
	Create(mention *models.MessageMention) error

	// Create multiple mentions at once
	CreateBatch(mentions []*models.MessageMention) error

	// Get mentions for a specific user (cursor-based pagination)
	// Returns: mentions, nextCursor, hasMore, error
	GetByUserID(
		userID uuid.UUID,
		limit int,
		cursor *string,
		direction string,
	) ([]*models.MessageMention, *string, bool, error)

	// Delete mentions for a message (when message is deleted)
	DeleteByMessageID(messageID uuid.UUID) error

	// Get all mentions in a message
	GetByMessageID(messageID uuid.UUID) ([]*models.MessageMention, error)
}
