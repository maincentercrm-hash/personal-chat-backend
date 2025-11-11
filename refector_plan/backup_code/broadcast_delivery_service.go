// application/serviceimpl/broadcast_delivery_service.go
package serviceimpl

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

type broadcastDeliveryService struct {
	broadcastDeliveryRepo repository.BroadcastDeliveryRepository
}

// NewBroadcastDeliveryService สร้าง instance ใหม่ของ BroadcastDeliveryService
func NewBroadcastDeliveryService(
	broadcastDeliveryRepo repository.BroadcastDeliveryRepository,
) service.BroadcastDeliveryService {
	return &broadcastDeliveryService{
		broadcastDeliveryRepo: broadcastDeliveryRepo,
	}
}

// GetByID ดึงข้อมูล broadcast delivery ตาม ID
func (s *broadcastDeliveryService) GetByID(id uuid.UUID) (*models.BroadcastDelivery, error) {
	return s.broadcastDeliveryRepo.GetByID(id)
}

// GetByBroadcastID ดึงข้อมูล broadcast deliveries ตาม broadcastID
func (s *broadcastDeliveryService) GetByBroadcastID(broadcastID uuid.UUID, status string, limit, offset int) ([]*models.BroadcastDelivery, int64, error) {
	return s.broadcastDeliveryRepo.GetByBroadcastID(broadcastID, status, limit, offset)
}

// GetByUserID ดึงข้อมูล broadcast deliveries ของผู้ใช้
func (s *broadcastDeliveryService) GetByUserID(userID uuid.UUID, limit, offset int) ([]*models.BroadcastDelivery, int64, error) {
	return s.broadcastDeliveryRepo.GetByUserID(userID, limit, offset)
}

// MarkAsOpened บันทึกว่าเปิดอ่านแล้ว
func (s *broadcastDeliveryService) MarkAsOpened(id uuid.UUID, openedAt time.Time) error {
	return s.broadcastDeliveryRepo.MarkAsOpened(id, openedAt)
}

// MarkAsClicked บันทึกว่าคลิกแล้ว
func (s *broadcastDeliveryService) MarkAsClicked(id uuid.UUID, clickedAt time.Time) error {
	return s.broadcastDeliveryRepo.MarkAsClicked(id, clickedAt)
}

// GetDeliveryStats ดึงสถิติการส่ง
func (s *broadcastDeliveryService) GetDeliveryStats(broadcastID uuid.UUID) (map[string]int64, error) {
	return s.broadcastDeliveryRepo.GetDeliveryStats(broadcastID)
}
