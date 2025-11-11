// domain/service/message_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// MessageService เป็น interface ที่กำหนดฟังก์ชันของ Message Service
type MessageService interface {
	// ส่งข้อความต่างๆ
	SendTextMessage(conversationID uuid.UUID, userID uuid.UUID, content string, metadata map[string]interface{}) (*models.Message, error)
	SendStickerMessage(conversationID uuid.UUID, userID uuid.UUID, stickerID uuid.UUID, stickerSetID uuid.UUID, mediaURL string, thumbnailURL string, metadata map[string]interface{}) (*models.Message, error)
	SendImageMessage(conversationID uuid.UUID, userID uuid.UUID, mediaURL string, thumbnailURL string, caption string, metadata map[string]interface{}) (*models.Message, error)
	SendFileMessage(conversationID uuid.UUID, userID uuid.UUID, mediaURL string, fileName string, fileSize int64, fileType string, metadata map[string]interface{}) (*models.Message, error)

	// ส่งข้อความในนามธุรกิจ
	SendBusinessTextMessage(businessID uuid.UUID, conversationID uuid.UUID, adminID uuid.UUID, content string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error)
	SendBusinessStickerMessage(businessID uuid.UUID, conversationID uuid.UUID, adminID uuid.UUID, stickerID uuid.UUID, stickerSetID uuid.UUID, mediaURL string, thumbnailURL string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error)
	SendBusinessImageMessage(businessID uuid.UUID, conversationID uuid.UUID, adminID uuid.UUID, mediaURL string, thumbnailURL string, caption string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error)
	SendBusinessFileMessage(businessID uuid.UUID, conversationID uuid.UUID, adminID uuid.UUID, mediaURL string, fileName string, fileSize int64, fileType string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error)

	// เพิ่มเมธอดใหม่สำหรับ Welcome Message โดยเฉพาะ
	SendWelcomeTextMessage(conversationID, businessID uuid.UUID, content string) error
	SendWelcomeImageMessage(conversationID, businessID uuid.UUID, imageURL, thumbnailURL string) error
	SendWelcomeCustomMessage(conversationID, businessID uuid.UUID, messageType, contentStr string) error

	// เพิ่มเมธอดสำหรับ Broadcast Message
	SendBroadcastTextMessage(conversationID, businessID, userID uuid.UUID, content string) error
	SendBroadcastImageMessage(conversationID, businessID, userID uuid.UUID, mediaURL, thumbnailURL string) error
	SendBroadcastCustomMessage(conversationID, businessID, userID uuid.UUID, messageType, contentStr string) error
	SendBroadcastStickerMessage(conversationID, businessID, userID uuid.UUID, stickerID, stickerSetID uuid.UUID, mediaURL, thumbnailURL string) error
	SendBroadcastFileMessage(conversationID, businessID, userID uuid.UUID, mediaURL, fileName string, fileSize int64, fileType string) error

	// จัดการข้อความ
	EditMessage(messageID uuid.UUID, userID uuid.UUID, newContent string) (*models.Message, error)
	DeleteMessage(messageID uuid.UUID, userID uuid.UUID) error
	ReplyToMessage(replyToID uuid.UUID, userID uuid.UUID, messageType string, content string, mediaURL string, thumbnailURL string, metadata map[string]interface{}) (*models.Message, error)

	// ดูประวัติข้อความ
	GetMessageEditHistory(messageID uuid.UUID, userID uuid.UUID) ([]*models.MessageEditHistory, error)
	GetMessageDeleteHistory(messageID uuid.UUID, userID uuid.UUID) ([]*models.MessageDeleteHistory, error)

	// ตรวจสอบสิทธิ์
	CheckBusinessAdmin(userID uuid.UUID, businessID uuid.UUID) (bool, bool, error) // คืนค่า (isAdmin, isHighLevelAdmin, error)
	CheckBusinessFollower(userID uuid.UUID, businessID uuid.UUID) (bool, error)
}
