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

	// เพิ่มเมธอดใหม่สำหรับ Welcome Message โดยเฉพาะ

	// เพิ่มเมธอดสำหรับ Broadcast Message

	// จัดการข้อความ
	EditMessage(messageID uuid.UUID, userID uuid.UUID, newContent string) (*models.Message, error)
	DeleteMessage(messageID uuid.UUID, userID uuid.UUID) error
	ReplyToMessage(replyToID uuid.UUID, userID uuid.UUID, messageType string, content string, mediaURL string, thumbnailURL string, metadata map[string]interface{}) (*models.Message, error)

	// ดูประวัติข้อความ
	GetMessageEditHistory(messageID uuid.UUID, userID uuid.UUID) ([]*models.MessageEditHistory, error)
	GetMessageDeleteHistory(messageID uuid.UUID, userID uuid.UUID) ([]*models.MessageDeleteHistory, error)

	// ตรวจสอบสิทธิ์
}
