// domain/service/scheduled_message_service.go
package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// ScheduledMessageService เป็น interface ที่กำหนดฟังก์ชันของ Scheduled Message Service
type ScheduledMessageService interface {
	// Create and manage scheduled messages
	ScheduleMessage(conversationID, userID uuid.UUID, messageType, content, mediaURL string, metadata map[string]interface{}, scheduledAt time.Time) (*models.ScheduledMessage, error)
	GetScheduledMessage(id, userID uuid.UUID) (*models.ScheduledMessage, error)
	GetUserScheduledMessages(userID uuid.UUID, limit, offset int) ([]*models.ScheduledMessage, int64, error)
	GetConversationScheduledMessages(conversationID, userID uuid.UUID, limit, offset int) ([]*models.ScheduledMessage, int64, error)
	CancelScheduledMessage(id, userID uuid.UUID) error

	// Background processor
	ProcessScheduledMessages() error
}
