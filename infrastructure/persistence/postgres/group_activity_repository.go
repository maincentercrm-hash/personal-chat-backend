// infrastructure/persistence/postgres/group_activity_repository.go
package postgres

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type groupActivityRepository struct {
	db *gorm.DB
}

// NewGroupActivityRepository สร้าง repository instance ใหม่
func NewGroupActivityRepository(db *gorm.DB) repository.GroupActivityRepository {
	return &groupActivityRepository{db: db}
}

// Create สร้าง activity ใหม่
func (r *groupActivityRepository) Create(activity *models.GroupActivity) error {
	return r.db.Create(activity).Error
}

// GetByConversationID ดึง activities ของ conversation พร้อม pagination และ type filter
func (r *groupActivityRepository) GetByConversationID(conversationID uuid.UUID, limit, offset int, activityType string) ([]*models.GroupActivity, int64, error) {
	var activities []*models.GroupActivity
	var total int64

	// สร้าง base query
	query := r.db.Model(&models.GroupActivity{}).
		Where("conversation_id = ?", conversationID)

	// เพิ่ม filter ตาม type ถ้ามีการระบุ
	if activityType != "" {
		query = query.Where("type = ?", activityType)
	}

	// นับจำนวนทั้งหมด
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลพร้อม preload ข้อมูลผู้ใช้
	dataQuery := r.db.
		Preload("Actor").
		Preload("Target").
		Where("conversation_id = ?", conversationID)

	// เพิ่ม filter ตาม type ถ้ามีการระบุ
	if activityType != "" {
		dataQuery = dataQuery.Where("type = ?", activityType)
	}

	err := dataQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// GetByID ดึง activity จาก ID
func (r *groupActivityRepository) GetByID(id uuid.UUID) (*models.GroupActivity, error) {
	var activity models.GroupActivity
	err := r.db.
		Preload("Actor").
		Preload("Target").
		Where("id = ?", id).
		First(&activity).Error

	if err != nil {
		return nil, err
	}

	return &activity, nil
}

// Delete ลบ activity
func (r *groupActivityRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.GroupActivity{}).Error
}

// DeleteByConversationID ลบ activities ทั้งหมดของ conversation
func (r *groupActivityRepository) DeleteByConversationID(conversationID uuid.UUID) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&models.GroupActivity{}).Error
}
