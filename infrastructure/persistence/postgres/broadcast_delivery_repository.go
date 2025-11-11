// infrastructure/persistence/postgres/broadcast_delivery_repository.go
package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type broadcastDeliveryRepository struct {
	db *gorm.DB
}

// NewBroadcastDeliveryRepository สร้าง instance ใหม่ของ BroadcastDeliveryRepository
func NewBroadcastDeliveryRepository(db *gorm.DB) repository.BroadcastDeliveryRepository {
	return &broadcastDeliveryRepository{
		db: db,
	}
}

// Create สร้าง broadcast delivery ใหม่
func (r *broadcastDeliveryRepository) Create(delivery *models.BroadcastDelivery) error {
	return r.db.Create(delivery).Error
}

// CreateBatch สร้าง broadcast deliveries หลายรายการพร้อมกัน
func (r *broadcastDeliveryRepository) CreateBatch(deliveries []*models.BroadcastDelivery) error {
	return r.db.CreateInBatches(deliveries, 1000).Error
}

// GetByID ดึงข้อมูล broadcast delivery ตาม ID
func (r *broadcastDeliveryRepository) GetByID(id uuid.UUID) (*models.BroadcastDelivery, error) {
	var delivery models.BroadcastDelivery
	if err := r.db.Where("id = ?", id).First(&delivery).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("broadcast delivery not found")
		}
		return nil, err
	}
	return &delivery, nil
}

// GetByBroadcastID ดึงข้อมูล broadcast deliveries ตาม broadcastID
func (r *broadcastDeliveryRepository) GetByBroadcastID(broadcastID uuid.UUID, status string, limit, offset int) ([]*models.BroadcastDelivery, int64, error) {
	var deliveries []*models.BroadcastDelivery
	var count int64

	query := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ?", broadcastID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// นับจำนวนทั้งหมด
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม pagination
	if err := query.Limit(limit).Offset(offset).Find(&deliveries).Error; err != nil {
		return nil, 0, err
	}

	return deliveries, count, nil
}

// GetByUserID ดึงข้อมูล broadcast deliveries ของผู้ใช้
func (r *broadcastDeliveryRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]*models.BroadcastDelivery, int64, error) {
	var deliveries []*models.BroadcastDelivery
	var count int64

	// นับจำนวนทั้งหมด
	if err := r.db.Model(&models.BroadcastDelivery{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// เรียงตาม delivered_at ถ้ามีค่า หากไม่มีค่าก็จะเรียงตาม ID
	if err := r.db.Where("user_id = ?", userID).
		Order("CASE WHEN delivered_at IS NULL THEN 0 ELSE 1 END DESC, delivered_at DESC, id DESC").
		Limit(limit).Offset(offset).Find(&deliveries).Error; err != nil {
		return nil, 0, err
	}

	return deliveries, count, nil
}

// Update อัพเดทข้อมูล broadcast delivery
func (r *broadcastDeliveryRepository) Update(delivery *models.BroadcastDelivery) error {
	return r.db.Save(delivery).Error
}

// UpdateStatus อัพเดทสถานะและข้อความผิดพลาด
func (r *broadcastDeliveryRepository) UpdateStatus(id uuid.UUID, status string, errorMessage string) error {
	return r.db.Model(&models.BroadcastDelivery{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        status,
			"error_message": errorMessage,
		}).Error
}

// MarkAsDelivered บันทึกว่าส่งแล้ว
func (r *broadcastDeliveryRepository) MarkAsDelivered(id uuid.UUID, deliveredAt time.Time) error {
	return r.db.Model(&models.BroadcastDelivery{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"delivered_at": deliveredAt,
			"status":       "delivered",
		}).Error
}

// MarkAsOpened บันทึกว่าเปิดอ่านแล้ว
func (r *broadcastDeliveryRepository) MarkAsOpened(id uuid.UUID, openedAt time.Time) error {
	return r.db.Model(&models.BroadcastDelivery{}).Where("id = ?", id).
		Update("opened_at", openedAt).Error
}

// MarkAsClicked บันทึกว่าคลิกแล้ว
func (r *broadcastDeliveryRepository) MarkAsClicked(id uuid.UUID, clickedAt time.Time) error {
	return r.db.Model(&models.BroadcastDelivery{}).Where("id = ?", id).
		Update("clicked_at", clickedAt).Error
}

// GetDeliveryStats ดึงสถิติการส่ง
func (r *broadcastDeliveryRepository) GetDeliveryStats(broadcastID uuid.UUID) (map[string]int64, error) {
	var total, pending, delivered, failed, opened, clicked int64

	// นับจำนวนทั้งหมด
	if err := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ?", broadcastID).Count(&total).Error; err != nil {
		return nil, err
	}

	// นับจำนวนตามสถานะ
	if err := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND status = ?", broadcastID, "pending").Count(&pending).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND status = ?", broadcastID, "delivered").Count(&delivered).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND status = ?", broadcastID, "failed").Count(&failed).Error; err != nil {
		return nil, err
	}

	// นับจำนวนที่เปิดอ่านแล้ว
	if err := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND opened_at IS NOT NULL", broadcastID).Count(&opened).Error; err != nil {
		return nil, err
	}

	// นับจำนวนที่คลิกแล้ว
	if err := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND clicked_at IS NOT NULL", broadcastID).Count(&clicked).Error; err != nil {
		return nil, err
	}

	stats := map[string]int64{
		"total":     total,
		"pending":   pending,
		"delivered": delivered,
		"failed":    failed,
		"opened":    opened,
		"clicked":   clicked,
	}

	return stats, nil
}

// GetUserDeliveryStatus ตรวจสอบสถานะการส่งให้กับผู้ใช้
func (r *broadcastDeliveryRepository) GetUserDeliveryStatus(broadcastID, userID uuid.UUID) (string, error) {
	var delivery models.BroadcastDelivery
	if err := r.db.Where("broadcast_id = ? AND user_id = ?", broadcastID, userID).First(&delivery).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("delivery record not found")
		}
		return "", err
	}
	return delivery.Status, nil
}

// GetPendingDeliveries ดึงรายการที่รอการส่ง
func (r *broadcastDeliveryRepository) GetPendingDeliveries(broadcastID uuid.UUID, limit int) ([]*models.BroadcastDelivery, error) {
	var deliveries []*models.BroadcastDelivery
	if err := r.db.Where("broadcast_id = ? AND status = ?", broadcastID, "pending").
		Limit(limit).Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}

// CountByStatus นับจำนวน deliveries ตามสถานะ
func (r *broadcastDeliveryRepository) CountByStatus(broadcastID uuid.UUID, status string) (int64, error) {
	var count int64
	query := r.db.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ?", broadcastID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteByBroadcastID ลบข้อมูล delivery ทั้งหมดที่เกี่ยวข้องกับ broadcast ID
func (r *broadcastDeliveryRepository) DeleteByBroadcastID(broadcastID uuid.UUID) error {
	return r.db.Where("broadcast_id = ?", broadcastID).Delete(&models.BroadcastDelivery{}).Error
}
