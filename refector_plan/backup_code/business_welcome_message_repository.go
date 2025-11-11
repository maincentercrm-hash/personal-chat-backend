// infrastructure/persistence/postgres/business_welcome_message_repository.go

package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type businessWelcomeMessageRepository struct {
	db *gorm.DB
}

// NewBusinessWelcomeMessageRepository สร้าง repository ใหม่
func NewBusinessWelcomeMessageRepository(db *gorm.DB) repository.BusinessWelcomeMessageRepository {
	return &businessWelcomeMessageRepository{
		db: db,
	}
}

// Create สร้าง welcome message ใหม่
func (r *businessWelcomeMessageRepository) Create(message *models.BusinessWelcomeMessage) error {
	return r.db.Create(message).Error
}

// GetByID ดึงข้อมูล welcome message ตาม ID
func (r *businessWelcomeMessageRepository) GetByID(id uuid.UUID) (*models.BusinessWelcomeMessage, error) {
	var message models.BusinessWelcomeMessage
	err := r.db.First(&message, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// GetActiveByBusinessID ดึงข้อมูล welcome message ที่เปิดใช้งานอยู่ของธุรกิจ เรียงตาม SortOrder
func (r *businessWelcomeMessageRepository) GetActiveByBusinessID(businessID uuid.UUID) ([]*models.BusinessWelcomeMessage, error) {
	var messages []*models.BusinessWelcomeMessage
	err := r.db.Where("business_id = ? AND is_active = ?", businessID, true).
		Order("sort_order ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// GetAllByBusinessID ดึงข้อมูล welcome message ทั้งหมดของธุรกิจ เรียงตาม SortOrder
func (r *businessWelcomeMessageRepository) GetAllByBusinessID(businessID uuid.UUID) ([]*models.BusinessWelcomeMessage, error) {
	var messages []*models.BusinessWelcomeMessage
	err := r.db.Where("business_id = ?", businessID).
		Order("sort_order ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// Update อัพเดทข้อมูล welcome message
func (r *businessWelcomeMessageRepository) Update(message *models.BusinessWelcomeMessage) error {
	// ตรวจสอบว่า message มีอยู่จริง
	exists, err := r.ExistsByID(message.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("welcome message not found")
	}

	// อัพเดทข้อมูล
	return r.db.Save(message).Error
}

// Delete ลบ welcome message ตาม ID
func (r *businessWelcomeMessageRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&models.BusinessWelcomeMessage{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("welcome message not found")
	}
	return nil
}

// SetActive กำหนดสถานะการใช้งานของ welcome message
func (r *businessWelcomeMessageRepository) SetActive(id uuid.UUID, isActive bool) error {
	result := r.db.Model(&models.BusinessWelcomeMessage{}).
		Where("id = ?", id).
		Update("is_active", isActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("welcome message not found")
	}
	return nil
}

// UpdateSortOrder อัพเดทลำดับการแสดงผลของ welcome message
func (r *businessWelcomeMessageRepository) UpdateSortOrder(id uuid.UUID, sortOrder int) error {
	result := r.db.Model(&models.BusinessWelcomeMessage{}).
		Where("id = ?", id).
		Update("sort_order", sortOrder)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("welcome message not found")
	}
	return nil
}

// UpdateMetrics อัพเดทค่าสถิติต่างๆ (SentCount, ClickCount, ReplyCount)
func (r *businessWelcomeMessageRepository) UpdateMetrics(id uuid.UUID, sentDelta, clickDelta, replyDelta int) error {
	// ใช้ raw SQL เพื่ออัพเดทค่า counter โดยไม่เกิด race condition
	// สำหรับ PostgreSQL ใช้ "+" แทน increment
	query := `
		UPDATE business_welcome_messages
		SET 
			sent_count = sent_count + ?,
			click_count = click_count + ?,
			reply_count = reply_count + ?,
			updated_at = ?
		WHERE id = ?
	`
	result := r.db.Exec(query, sentDelta, clickDelta, replyDelta, time.Now(), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("welcome message not found")
	}
	return nil
}

// GetByTriggerType ดึงข้อมูล welcome message ที่เปิดใช้งานตามประเภททริกเกอร์
func (r *businessWelcomeMessageRepository) GetByTriggerType(businessID uuid.UUID, triggerType string) ([]*models.BusinessWelcomeMessage, error) {
	var messages []*models.BusinessWelcomeMessage
	err := r.db.Where("business_id = ? AND trigger_type = ? AND is_active = ?", businessID, triggerType, true).
		Order("sort_order ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// ExistsByID ตรวจสอบว่ามี welcome message ตาม ID หรือไม่
func (r *businessWelcomeMessageRepository) ExistsByID(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.BusinessWelcomeMessage{}).
		Where("id = ?", id).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountByBusinessID นับจำนวน welcome message ทั้งหมดของธุรกิจ
func (r *businessWelcomeMessageRepository) CountByBusinessID(businessID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.BusinessWelcomeMessage{}).
		Where("business_id = ?", businessID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
