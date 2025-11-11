// domain/repository/message_repository.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// MessageRepository เป็น interface สำหรับจัดการข้อมูลข้อความ
type MessageRepository interface {
	// การดึงข้อมูลข้อความ
	GetByID(id uuid.UUID) (*models.Message, error)
	GetMessagesByConversationID(conversationID uuid.UUID, limit, offset int) ([]*models.Message, int64, error)

	// การสร้างและแก้ไขข้อความ
	Create(message *models.Message) error
	Update(message *models.Message) error
	Delete(id uuid.UUID) error

	// การจัดการประวัติการแก้ไขและลบ
	CreateEditHistory(history *models.MessageEditHistory) error
	GetEditHistory(messageID uuid.UUID) ([]*models.MessageEditHistory, error)
	CreateDeleteHistory(history *models.MessageDeleteHistory) error
	GetDeleteHistory(messageID uuid.UUID) ([]*models.MessageDeleteHistory, error)

	// การจัดการการอ่านข้อความ
	MarkAsRead(messageID, userID uuid.UUID, readAt time.Time) error
	GetReads(messageID uuid.UUID) ([]*models.MessageRead, error)
	IsMessageRead(messageID, userID uuid.UUID) (bool, error)
	MarkAllAsRead(conversationID, userID uuid.UUID, readAt time.Time) error

	// ตรวจสอบความเป็นเจ้าของและสิทธิ์
	IsSender(messageID, userID uuid.UUID) (bool, error)
	IsConversationAdmin(conversationID, userID uuid.UUID) (bool, error)

	// อัพเดตข้อความล่าสุดในการสนทนา
	UpdateConversationLastMessage(conversationID uuid.UUID, lastMessageText string, lastMessageAt time.Time) error

	GetLastMessageByConversation(conversationID uuid.UUID) (*models.Message, error)
	GetLastNonDeletedMessageByConversation(conversationID uuid.UUID) (*models.Message, error)

	// GetMessagesBefore ดึงข้อความที่เก่ากว่า ID ที่ระบุ
	GetMessagesBefore(conversationID, messageID uuid.UUID, limit int) ([]*models.Message, error)

	// GetMessagesAfter ดึงข้อความที่ใหม่กว่า ID ที่ระบุ
	GetMessagesAfter(conversationID, messageID uuid.UUID, limit int) ([]*models.Message, error)

	// CountAllMessages นับจำนวนข้อความทั้งหมดในการสนทนา
	CountAllMessages(conversationID uuid.UUID) (int64, error)

	// เมธอดสำหรับดึงข้อความหลังเวลาที่กำหนดและไม่ใช่ของผู้ใช้
	GetMessagesAfterTime(conversationID uuid.UUID, afterTime time.Time, excludeUserID uuid.UUID) ([]*models.Message, error)

	// เมธอดสำหรับดึงข้อความทั้งหมดที่ไม่ใช่ของผู้ใช้ (สำหรับกรณีไม่มี LastReadAt)
	GetAllUnreadMessages(conversationID uuid.UUID, excludeUserID uuid.UUID) ([]*models.Message, error)

	GetCustomerMessagesAfterTime(conversationID uuid.UUID, afterTime time.Time, businessID uuid.UUID) ([]*models.Message, error)
	GetAllCustomerMessages(conversationID uuid.UUID, businessID uuid.UUID) ([]*models.Message, error)
}
