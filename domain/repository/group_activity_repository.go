// domain/repository/group_activity_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// GroupActivityRepository interface สำหรับจัดการ group activities
type GroupActivityRepository interface {
	// Create สร้าง activity ใหม่
	Create(activity *models.GroupActivity) error

	// GetByConversationID ดึง activities ของ conversation (ถ้า activityType ไม่ว่าง จะ filter ตาม type)
	GetByConversationID(conversationID uuid.UUID, limit, offset int, activityType string) ([]*models.GroupActivity, int64, error)

	// GetByID ดึง activity จาก ID
	GetByID(id uuid.UUID) (*models.GroupActivity, error)

	// Delete ลบ activity
	Delete(id uuid.UUID) error

	// DeleteByConversationID ลบ activities ทั้งหมดของ conversation
	DeleteByConversationID(conversationID uuid.UUID) error
}
