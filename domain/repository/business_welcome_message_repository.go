// domain/repository/business_welcome_message_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// BusinessWelcomeMessageRepository คือ interface สำหรับจัดการกับข้อมูล welcome message ในฐานข้อมูล
type BusinessWelcomeMessageRepository interface {
	// Create สร้าง welcome message ใหม่
	Create(message *models.BusinessWelcomeMessage) error

	// GetByID ดึงข้อมูล welcome message ตาม ID
	GetByID(id uuid.UUID) (*models.BusinessWelcomeMessage, error)

	// GetActiveByBusinessID ดึงข้อมูล welcome message ที่เปิดใช้งานอยู่ของธุรกิจ เรียงตาม SortOrder
	GetActiveByBusinessID(businessID uuid.UUID) ([]*models.BusinessWelcomeMessage, error)

	// GetAllByBusinessID ดึงข้อมูล welcome message ทั้งหมดของธุรกิจ เรียงตาม SortOrder
	GetAllByBusinessID(businessID uuid.UUID) ([]*models.BusinessWelcomeMessage, error)

	// Update อัพเดทข้อมูล welcome message
	Update(message *models.BusinessWelcomeMessage) error

	// Delete ลบ welcome message ตาม ID
	Delete(id uuid.UUID) error

	// SetActive กำหนดสถานะการใช้งานของ welcome message
	SetActive(id uuid.UUID, isActive bool) error

	// UpdateSortOrder อัพเดทลำดับการแสดงผลของ welcome message
	UpdateSortOrder(id uuid.UUID, sortOrder int) error

	// UpdateMetrics อัพเดทค่าสถิติต่างๆ (SentCount, ClickCount, ReplyCount)
	UpdateMetrics(id uuid.UUID, sentDelta, clickDelta, replyDelta int) error

	// GetByTriggerType ดึงข้อมูล welcome message ที่เปิดใช้งานตามประเภททริกเกอร์
	GetByTriggerType(businessID uuid.UUID, triggerType string) ([]*models.BusinessWelcomeMessage, error)

	// ExistsByID ตรวจสอบว่ามี welcome message ตาม ID หรือไม่
	ExistsByID(id uuid.UUID) (bool, error)

	// CountByBusinessID นับจำนวน welcome message ทั้งหมดของธุรกิจ
	CountByBusinessID(businessID uuid.UUID) (int64, error)
}
