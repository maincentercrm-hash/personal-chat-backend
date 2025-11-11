// domain/repository/message_read_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// MessageReadRepository เป็น interface สำหรับจัดการข้อมูลการอ่านข้อความ
type MessageReadRepository interface {
	CreateRead(read *models.MessageRead) error
	GetByMessageID(messageID uuid.UUID) ([]*models.MessageRead, error)
	GetUnreadMessageIDs(conversationID, userID uuid.UUID) ([]uuid.UUID, error)

	// อาจเพิ่มเมธอดอื่นๆ ถ้าจำเป็น
	DeleteRead(messageID, userID uuid.UUID) error
	CountReads(messageID uuid.UUID) (int, error)
}
