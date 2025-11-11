// domain/service/message_read_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// MessageReadService เป็น interface สำหรับจัดการการอ่านข้อความ
type MessageReadService interface {
	// MarkMessageAsRead ทำเครื่องหมายว่าข้อความถูกอ่านแล้ว
	MarkMessageAsRead(messageID, userID uuid.UUID) (uuid.UUID, error)

	// GetMessageReads ดึงข้อมูลผู้ที่อ่านข้อความแล้ว
	GetMessageReads(messageID, userID uuid.UUID) ([]*models.MessageRead, error)

	// MarkAllMessagesAsRead ทำเครื่องหมายว่าข้อความทั้งหมดในการสนทนาถูกอ่านแล้ว
	MarkAllMessagesAsRead(conversationID, userID uuid.UUID) (int, error)

	// GetUnreadCount ดึงจำนวนข้อความที่ยังไม่ได้อ่านในการสนทนา
	GetUnreadCount(conversationID, userID uuid.UUID) (int, error)
}
