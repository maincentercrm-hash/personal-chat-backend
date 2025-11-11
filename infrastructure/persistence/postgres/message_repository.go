// infrastructure/persistence/postgres/message_repository.go
package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository สร้าง repository ใหม่
func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{
		db: db,
	}
}

// GetByID ดึงข้อมูลข้อความตาม ID
func (r *messageRepository) GetByID(id uuid.UUID) (*models.Message, error) {
	var message models.Message
	if err := r.db.First(&message, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// GetMessagesByConversationID ดึงข้อความทั้งหมดในการสนทนา
func (r *messageRepository) GetMessagesByConversationID(conversationID uuid.UUID, limit, offset int) ([]*models.Message, int64, error) {
	var count int64
	if err := r.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	var messages []*models.Message
	if err := r.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, count, nil
}

// Create สร้างข้อความใหม่
func (r *messageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

// Update อัพเดตข้อความ
func (r *messageRepository) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

// Delete ลบข้อความ (soft delete)
func (r *messageRepository) Delete(id uuid.UUID) error {
	result := r.db.Model(&models.Message{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":          true,
			"content":             nil,
			"media_url":           nil,
			"media_thumbnail_url": nil,
			"metadata":            "{}",
			"updated_at":          time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("message not found")
	}
	return nil
}

// CreateEditHistory บันทึกประวัติการแก้ไขข้อความ
func (r *messageRepository) CreateEditHistory(history *models.MessageEditHistory) error {
	return r.db.Create(history).Error
}

// GetEditHistory ดึงประวัติการแก้ไขข้อความ
func (r *messageRepository) GetEditHistory(messageID uuid.UUID) ([]*models.MessageEditHistory, error) {
	var history []*models.MessageEditHistory
	if err := r.db.Where("message_id = ?", messageID).
		Order("edited_at DESC").
		Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

// CreateDeleteHistory บันทึกประวัติการลบข้อความ
func (r *messageRepository) CreateDeleteHistory(history *models.MessageDeleteHistory) error {
	return r.db.Create(history).Error
}

// GetDeleteHistory ดึงประวัติการลบข้อความ
func (r *messageRepository) GetDeleteHistory(messageID uuid.UUID) ([]*models.MessageDeleteHistory, error) {
	var history []*models.MessageDeleteHistory
	if err := r.db.Where("message_id = ?", messageID).
		Order("deleted_at DESC").
		Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

// MarkAsRead ทำเครื่องหมายว่าข้อความถูกอ่านแล้ว
func (r *messageRepository) MarkAsRead(messageID, userID uuid.UUID, readAt time.Time) error {
	// ตรวจสอบว่ามีการอ่านแล้วหรือไม่
	var count int64
	if err := r.db.Model(&models.MessageRead{}).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // มีการอ่านแล้ว ไม่ต้องทำอะไร
	}

	// สร้างบันทึกการอ่าน
	read := models.MessageRead{
		ID:        uuid.New(),
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    readAt,
	}

	return r.db.Create(&read).Error
}

// GetReads ดึงรายการการอ่านข้อความ
func (r *messageRepository) GetReads(messageID uuid.UUID) ([]*models.MessageRead, error) {
	var reads []*models.MessageRead
	if err := r.db.Where("message_id = ?", messageID).
		Order("read_at ASC").
		Find(&reads).Error; err != nil {
		return nil, err
	}
	return reads, nil
}

// IsMessageRead ตรวจสอบว่าข้อความถูกอ่านโดยผู้ใช้แล้วหรือไม่
func (r *messageRepository) IsMessageRead(messageID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.MessageRead{}).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkAllAsRead ทำเครื่องหมายว่าข้อความทั้งหมดในการสนทนาถูกอ่านแล้ว
func (r *messageRepository) MarkAllAsRead(conversationID, userID uuid.UUID, readAt time.Time) error {
	// ดึงรายการข้อความที่ยังไม่ได้อ่าน
	rows, err := r.db.Raw(`
		SELECT m.id 
		FROM messages m 
		LEFT JOIN message_reads mr ON m.id = mr.message_id AND mr.user_id = ? 
		WHERE m.conversation_id = ? AND m.sender_id != ? AND mr.id IS NULL AND m.is_deleted = false
	`, userID, conversationID, userID).Rows()

	if err != nil {
		return err
	}
	defer rows.Close()

	var messageIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		messageIDs = append(messageIDs, id)
	}

	if len(messageIDs) == 0 {
		return nil // ไม่มีข้อความที่ต้องมาร์ค
	}

	// สร้างบันทึกการอ่านเป็นชุด
	reads := make([]models.MessageRead, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		reads = append(reads, models.MessageRead{
			ID:        uuid.New(),
			MessageID: messageID,
			UserID:    userID,
			ReadAt:    readAt,
		})
	}

	return r.db.CreateInBatches(reads, 100).Error
}

// IsSender ตรวจสอบว่าผู้ใช้เป็นผู้ส่งข้อความหรือไม่
func (r *messageRepository) IsSender(messageID, userID uuid.UUID) (bool, error) {
	var message models.Message
	err := r.db.Select("sender_id").First(&message, "id = ?", messageID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("message not found")
		}
		return false, err
	}
	return message.SenderID != nil && *message.SenderID == userID, nil
}

// IsConversationAdmin ตรวจสอบว่าผู้ใช้เป็นแอดมินของการสนทนาหรือไม่
func (r *messageRepository) IsConversationAdmin(conversationID, userID uuid.UUID) (bool, error) {
	var member models.ConversationMember
	err := r.db.Select("is_admin").First(&member, "conversation_id = ? AND user_id = ?", conversationID, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return member.IsAdmin, nil
}

// UpdateConversationLastMessage อัพเดตข้อความล่าสุดในการสนทนา
func (r *messageRepository) UpdateConversationLastMessage(conversationID uuid.UUID, lastMessageText string, lastMessageAt time.Time) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ?", conversationID).
		Updates(map[string]interface{}{
			"last_message_text": lastMessageText,
			"last_message_at":   lastMessageAt,
			"updated_at":        time.Now(),
		}).Error
}

// GetLastMessageByConversation ดึงข้อความล่าสุดของการสนทนา
func (r *messageRepository) GetLastMessageByConversation(conversationID uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&message).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// GetLastNonDeletedMessageByConversation ดึงข้อความล่าสุดที่ไม่ถูกลบของการสนทนา
func (r *messageRepository) GetLastNonDeletedMessageByConversation(conversationID uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.Where("conversation_id = ? AND is_deleted = ?", conversationID, false).
		Order("created_at DESC").
		First(&message).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// GetMessagesBefore ดึงข้อความที่เก่ากว่า ID ที่ระบุ
func (r *messageRepository) GetMessagesBefore(conversationID, messageID uuid.UUID, limit int) ([]*models.Message, error) {
	var targetMessage models.Message
	if err := r.db.First(&targetMessage, "id = ?", messageID).Error; err != nil {
		return nil, err
	}

	var messages []*models.Message

	// ดึงข้อความที่เก่ากว่าข้อความเป้าหมาย โดยเรียงจากใหม่ไปเก่า
	if err := r.db.Where("conversation_id = ? AND created_at < ?", conversationID, targetMessage.CreatedAt).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	// กลับลำดับให้เป็นจากเก่าไปใหม่
	for i := 0; i < len(messages)/2; i++ {
		messages[i], messages[len(messages)-1-i] = messages[len(messages)-1-i], messages[i]
	}

	return messages, nil
}

// GetMessagesAfter ดึงข้อความที่ใหม่กว่า ID ที่ระบุ
func (r *messageRepository) GetMessagesAfter(conversationID, messageID uuid.UUID, limit int) ([]*models.Message, error) {
	var targetMessage models.Message
	if err := r.db.First(&targetMessage, "id = ?", messageID).Error; err != nil {
		return nil, err
	}

	var messages []*models.Message

	// ดึงข้อความที่ใหม่กว่าข้อความเป้าหมาย โดยเรียงจากเก่าไปใหม่
	if err := r.db.Where("conversation_id = ? AND created_at > ?", conversationID, targetMessage.CreatedAt).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// CountAllMessages นับจำนวนข้อความทั้งหมดในการสนทนา
func (r *messageRepository) CountAllMessages(conversationID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID).Count(&count).Error
	return count, err
}

// infrastructure/persistence/postgres/message_repository.go
func (r *messageRepository) GetMessagesAfterTime(conversationID uuid.UUID, afterTime time.Time, excludeUserID uuid.UUID) ([]*models.Message, error) {
	var messages []*models.Message

	// ดึงข้อความที่สร้างหลังเวลาที่กำหนด ไม่ใช่ของผู้ใช้ที่กำหนด และไม่ถูกลบ
	err := r.db.Where("conversation_id = ? AND created_at > ? AND sender_id != ? AND is_deleted = ?",
		conversationID, afterTime, excludeUserID, false).
		Find(&messages).Error

	return messages, err
}

func (r *messageRepository) GetAllUnreadMessages(conversationID uuid.UUID, excludeUserID uuid.UUID) ([]*models.Message, error) {
	var messages []*models.Message

	// ดึงข้อความทั้งหมดในการสนทนาที่ไม่ใช่ของผู้ใช้ที่กำหนด และไม่ถูกลบ
	err := r.db.Where("conversation_id = ? AND sender_id != ? AND is_deleted = ?",
		conversationID, excludeUserID, false).
		Find(&messages).Error

	return messages, err
}

// GetCustomerMessagesAfterTime ดึงข้อความจากลูกค้าหลังจากเวลาที่กำหนด
func (r *messageRepository) GetCustomerMessagesAfterTime(conversationID uuid.UUID, afterTime time.Time, businessID uuid.UUID) ([]*models.Message, error) {
	var messages []*models.Message

	// ดึงรายการ business admin IDs
	var adminUserIDs []uuid.UUID
	err := r.db.Model(&models.BusinessAdmin{}).
		Select("user_id").
		Where("business_id = ?", businessID).
		Find(&adminUserIDs).Error
	if err != nil {
		return nil, err
	}

	// ดึงข้อความที่ไม่ใช่จากแอดมิน และส่งหลังจากเวลาที่กำหนด
	query := r.db.Where("conversation_id = ? AND created_at > ? AND is_deleted = ?",
		conversationID, afterTime, false)

	if len(adminUserIDs) > 0 {
		// ไม่รวมข้อความจากแอดมิน
		query = query.Where("sender_id NOT IN ?", adminUserIDs)
	}

	// ไม่รวมข้อความระบบ
	query = query.Where("message_type != ?", "system")

	err = query.Find(&messages).Error
	return messages, err
}

// GetAllCustomerMessages ดึงข้อความจากลูกค้าทั้งหมด
func (r *messageRepository) GetAllCustomerMessages(conversationID uuid.UUID, businessID uuid.UUID) ([]*models.Message, error) {
	var messages []*models.Message

	// ดึงรายการ business admin IDs
	var adminUserIDs []uuid.UUID
	err := r.db.Model(&models.BusinessAdmin{}).
		Select("user_id").
		Where("business_id = ?", businessID).
		Find(&adminUserIDs).Error
	if err != nil {
		return nil, err
	}

	// ดึงข้อความที่ไม่ใช่จากแอดมิน
	query := r.db.Where("conversation_id = ? AND is_deleted = ?",
		conversationID, false)

	if len(adminUserIDs) > 0 {
		query = query.Where("sender_id NOT IN ?", adminUserIDs)
	}

	// ไม่รวมข้อความระบบ
	query = query.Where("message_type != ?", "system")

	err = query.Find(&messages).Error
	return messages, err
}
