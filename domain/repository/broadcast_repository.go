// domain/repository/broadcast_repository.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// BroadcastRepository interface สำหรับจัดการข้อมูล Broadcast
type BroadcastRepository interface {
	// Create สร้าง broadcast ใหม่
	Create(broadcast *models.Broadcast) error

	// GetByID ดึงข้อมูล broadcast ตาม ID
	GetByID(id uuid.UUID) (*models.Broadcast, error)

	// GetByBusinessID ดึงข้อมูล broadcast ทั้งหมดของธุรกิจ
	GetByBusinessID(businessID uuid.UUID, status string, limit, offset int) ([]*models.Broadcast, int64, error)

	// Update อัพเดทข้อมูล broadcast
	Update(broadcast *models.Broadcast) error

	// Delete ลบ broadcast
	Delete(id uuid.UUID) error

	// UpdateStatus อัพเดทสถานะของ broadcast
	UpdateStatus(id uuid.UUID, status string) error

	// ScheduleBroadcast กำหนดเวลาส่ง broadcast
	ScheduleBroadcast(id uuid.UUID, scheduledAt time.Time) error

	// MarkAsSent บันทึกว่า broadcast ถูกส่งแล้ว
	MarkAsSent(id uuid.UUID) error

	// UpdateMetrics อัพเดทค่าสถิติใน metrics field
	UpdateMetrics(id uuid.UUID, metrics map[string]interface{}) error

	// SearchBroadcasts ค้นหา broadcasts ตามเงื่อนไข
	SearchBroadcasts(businessID uuid.UUID, query string, messageType, status string, startDate, endDate time.Time, limit, offset int) ([]*models.Broadcast, int64, error)

	// CountByStatus นับจำนวน broadcasts ตามสถานะ
	CountByStatus(businessID uuid.UUID, status string) (int64, error)

	// GetScheduledBroadcasts ดึง broadcasts ที่กำหนดเวลาส่งและถึงเวลาส่งแล้ว
	GetScheduledBroadcasts(currentTime time.Time) ([]*models.Broadcast, error)

	// ExistsByID ตรวจสอบว่ามี broadcast ตาม ID หรือไม่
	ExistsByID(id uuid.UUID) (bool, error)
}
