// infrastructure/persistence/postgres/message_read_repository.go
package postgres

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

// messageReadRepository เป็น implementation ของ MessageReadRepository
type messageReadRepository struct {
	db *gorm.DB
}

// NewMessageReadRepository สร้าง repository ใหม่
func NewMessageReadRepository(db *gorm.DB) repository.MessageReadRepository {
	return &messageReadRepository{
		db: db,
	}
}

// CreateRead สร้างบันทึกการอ่านข้อความ (ป้องกันการซ้ำซ้อน)
func (r *messageReadRepository) CreateRead(read *models.MessageRead) error {
	// ใช้ Raw SQL เพื่อทำ "INSERT ... ON CONFLICT DO NOTHING"
	return r.db.Exec(`
    INSERT INTO message_reads (id, message_id, user_id, read_at)
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (message_id, user_id) DO NOTHING
`, read.ID, read.MessageID, read.UserID, read.ReadAt).Error
}

// GetByMessageID ดึงรายการการอ่านของข้อความ
func (r *messageReadRepository) GetByMessageID(messageID uuid.UUID) ([]*models.MessageRead, error) {
	var reads []*models.MessageRead
	err := r.db.Where("message_id = ?", messageID).
		Order("read_at ASC").
		Find(&reads).Error

	if err != nil {
		return nil, err
	}

	return reads, nil
}

// GetUnreadMessageIDs ดึงรายการ ID ของข้อความที่ยังไม่ได้อ่าน
func (r *messageReadRepository) GetUnreadMessageIDs(conversationID, userID uuid.UUID) ([]uuid.UUID, error) {
	// ดึงข้อความทั้งหมดในการสนทนาที่ไม่ได้ส่งโดยผู้ใช้นี้
	subQuery := r.db.Model(&models.Message{}).
		Select("id").
		Where("conversation_id = ? AND sender_id != ? AND is_deleted = ?",
			conversationID, userID, false)

	// ดึงข้อความที่ผู้ใช้ยังไม่ได้อ่าน
	var unreadMessageIDs []uuid.UUID

	err := r.db.Model(&models.Message{}).
		Select("messages.id").
		Where("messages.id IN (?)", subQuery).
		Joins("LEFT JOIN message_reads ON messages.id = message_reads.message_id AND message_reads.user_id = ?", userID).
		Where("message_reads.id IS NULL").
		Pluck("messages.id", &unreadMessageIDs).Error

	if err != nil {
		return nil, err
	}

	return unreadMessageIDs, nil
}

// DeleteRead ลบบันทึกการอ่านข้อความ
func (r *messageReadRepository) DeleteRead(messageID, userID uuid.UUID) error {
	return r.db.Where("message_id = ? AND user_id = ?", messageID, userID).
		Delete(&models.MessageRead{}).Error
}

// CountReads นับจำนวนการอ่านของข้อความ
func (r *messageReadRepository) CountReads(messageID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&models.MessageRead{}).
		Where("message_id = ?", messageID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}
