// domain/service/broadcast_delivery_service.go
package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// BroadcastDeliveryService interface สำหรับจัดการบริการ broadcast delivery
type BroadcastDeliveryService interface {
	// GetByID ดึงข้อมูล broadcast delivery ตาม ID
	GetByID(id uuid.UUID) (*models.BroadcastDelivery, error)

	// GetByBroadcastID ดึงข้อมูล broadcast deliveries ตาม broadcastID
	GetByBroadcastID(broadcastID uuid.UUID, status string, limit, offset int) ([]*models.BroadcastDelivery, int64, error)

	// GetByUserID ดึงข้อมูล broadcast deliveries ของผู้ใช้
	GetByUserID(userID uuid.UUID, limit, offset int) ([]*models.BroadcastDelivery, int64, error)

	// MarkAsOpened บันทึกว่าเปิดอ่านแล้ว
	MarkAsOpened(id uuid.UUID, openedAt time.Time) error

	// MarkAsClicked บันทึกว่าคลิกแล้ว
	MarkAsClicked(id uuid.UUID, clickedAt time.Time) error

	// GetDeliveryStats ดึงสถิติการส่ง
	GetDeliveryStats(broadcastID uuid.UUID) (map[string]int64, error)
}
