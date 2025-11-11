// domain/repository/broadcast_delivery_repository.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// BroadcastDeliveryRepository interface สำหรับจัดการข้อมูล BroadcastDelivery
type BroadcastDeliveryRepository interface {
	// Create สร้าง broadcast delivery ใหม่
	Create(delivery *models.BroadcastDelivery) error

	// CreateBatch สร้าง broadcast deliveries หลายรายการพร้อมกัน
	CreateBatch(deliveries []*models.BroadcastDelivery) error

	// GetByID ดึงข้อมูล broadcast delivery ตาม ID
	GetByID(id uuid.UUID) (*models.BroadcastDelivery, error)

	// GetByBroadcastID ดึงข้อมูล broadcast deliveries ตาม broadcastID
	GetByBroadcastID(broadcastID uuid.UUID, status string, limit, offset int) ([]*models.BroadcastDelivery, int64, error)

	// GetByUserID ดึงข้อมูล broadcast deliveries ของผู้ใช้
	GetByUserID(userID uuid.UUID, limit, offset int) ([]*models.BroadcastDelivery, int64, error)

	// Update อัพเดทข้อมูล broadcast delivery
	Update(delivery *models.BroadcastDelivery) error

	// UpdateStatus อัพเดทสถานะและข้อความผิดพลาด
	UpdateStatus(id uuid.UUID, status string, errorMessage string) error

	// MarkAsDelivered บันทึกว่าส่งแล้ว
	MarkAsDelivered(id uuid.UUID, deliveredAt time.Time) error

	// MarkAsOpened บันทึกว่าเปิดอ่านแล้ว
	MarkAsOpened(id uuid.UUID, openedAt time.Time) error

	// MarkAsClicked บันทึกว่าคลิกแล้ว
	MarkAsClicked(id uuid.UUID, clickedAt time.Time) error

	// GetDeliveryStats ดึงสถิติการส่ง
	GetDeliveryStats(broadcastID uuid.UUID) (map[string]int64, error)

	// GetUserDeliveryStatus ตรวจสอบสถานะการส่งให้กับผู้ใช้
	GetUserDeliveryStatus(broadcastID, userID uuid.UUID) (string, error)

	// GetPendingDeliveries ดึงรายการที่รอการส่ง
	GetPendingDeliveries(broadcastID uuid.UUID, limit int) ([]*models.BroadcastDelivery, error)

	// CountByStatus นับจำนวน deliveries ตามสถานะ
	CountByStatus(broadcastID uuid.UUID, status string) (int64, error)

	// DeleteByBroadcastID ลบข้อมูล delivery ทั้งหมดที่เกี่ยวข้องกับ broadcast ID
	DeleteByBroadcastID(broadcastID uuid.UUID) error
}
