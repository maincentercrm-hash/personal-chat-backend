// infrastructure/persistence/postgres/broadcast_repository.go
package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type broadcastRepository struct {
	db *gorm.DB
}

// NewBroadcastRepository สร้าง instance ใหม่ของ BroadcastRepository
func NewBroadcastRepository(db *gorm.DB) repository.BroadcastRepository {
	return &broadcastRepository{
		db: db,
	}
}

// Create สร้าง broadcast ใหม่
func (r *broadcastRepository) Create(broadcast *models.Broadcast) error {
	return r.db.Create(broadcast).Error
}

// GetByID ดึงข้อมูล broadcast ตาม ID
func (r *broadcastRepository) GetByID(id uuid.UUID) (*models.Broadcast, error) {
	var broadcast models.Broadcast
	if err := r.db.Where("id = ?", id).First(&broadcast).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("broadcast not found")
		}
		return nil, err
	}
	return &broadcast, nil
}

// GetByBusinessID ดึงข้อมูล broadcast ทั้งหมดของธุรกิจ
func (r *broadcastRepository) GetByBusinessID(businessID uuid.UUID, status string, limit, offset int) ([]*models.Broadcast, int64, error) {
	var broadcasts []*models.Broadcast
	var count int64

	query := r.db.Model(&models.Broadcast{}).Where("business_id = ?", businessID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// นับจำนวนทั้งหมด
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม pagination
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&broadcasts).Error; err != nil {
		return nil, 0, err
	}

	return broadcasts, count, nil
}

// Update อัพเดทข้อมูล broadcast
func (r *broadcastRepository) Update(broadcast *models.Broadcast) error {
	return r.db.Save(broadcast).Error
}

// Delete ลบ broadcast
func (r *broadcastRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Broadcast{}, id).Error
}

// UpdateStatus อัพเดทสถานะของ broadcast
func (r *broadcastRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Broadcast{}).Where("id = ?", id).Update("status", status).Error
}

// ScheduleBroadcast กำหนดเวลาส่ง broadcast
func (r *broadcastRepository) ScheduleBroadcast(id uuid.UUID, scheduledAt time.Time) error {
	return r.db.Model(&models.Broadcast{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"scheduled_at": scheduledAt,
			"status":       "scheduled",
		}).Error
}

// MarkAsSent บันทึกว่า broadcast ถูกส่งแล้ว
func (r *broadcastRepository) MarkAsSent(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Broadcast{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"sent_at": now,
			"status":  "completed",
		}).Error
}

// UpdateMetrics อัพเดทค่าสถิติใน metrics field
func (r *broadcastRepository) UpdateMetrics(id uuid.UUID, metrics map[string]interface{}) error {
	var broadcast models.Broadcast
	if err := r.db.Where("id = ?", id).First(&broadcast).Error; err != nil {
		return err
	}

	// อัพเดทค่าสถิติใน metrics field
	currentMetrics := broadcast.Metrics
	for key, value := range metrics {
		currentMetrics[key] = value
	}

	return r.db.Model(&models.Broadcast{}).Where("id = ?", id).
		Update("metrics", currentMetrics).Error
}

// SearchBroadcasts ค้นหา broadcasts ตามเงื่อนไข
func (r *broadcastRepository) SearchBroadcasts(businessID uuid.UUID, query string, messageType, status string, startDate, endDate time.Time, limit, offset int) ([]*models.Broadcast, int64, error) {
	var broadcasts []*models.Broadcast
	var count int64

	dbQuery := r.db.Model(&models.Broadcast{}).Where("business_id = ?", businessID)

	// ค้นหาตามคำ
	if query != "" {
		dbQuery = dbQuery.Where("title ILIKE ? OR content ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	// กรองตามประเภทข้อความ
	if messageType != "" {
		dbQuery = dbQuery.Where("message_type = ?", messageType)
	}

	// กรองตามสถานะ
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}

	// กรองตามช่วงเวลา
	if !startDate.IsZero() {
		dbQuery = dbQuery.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		dbQuery = dbQuery.Where("created_at <= ?", endDate)
	}

	// นับจำนวนทั้งหมด
	if err := dbQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม pagination
	if err := dbQuery.Order("created_at desc").Limit(limit).Offset(offset).Find(&broadcasts).Error; err != nil {
		return nil, 0, err
	}

	return broadcasts, count, nil
}

// CountByStatus นับจำนวน broadcasts ตามสถานะ
func (r *broadcastRepository) CountByStatus(businessID uuid.UUID, status string) (int64, error) {
	var count int64
	query := r.db.Model(&models.Broadcast{}).Where("business_id = ?", businessID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// GetScheduledBroadcasts ดึง broadcasts ที่กำหนดเวลาส่งและถึงเวลาส่งแล้ว
func (r *broadcastRepository) GetScheduledBroadcasts(currentTime time.Time) ([]*models.Broadcast, error) {
	var broadcasts []*models.Broadcast

	if err := r.db.Where("status = ? AND scheduled_at <= ?", "scheduled", currentTime).
		Find(&broadcasts).Error; err != nil {
		return nil, err
	}

	return broadcasts, nil
}

// ExistsByID ตรวจสอบว่ามี broadcast ตาม ID หรือไม่
func (r *broadcastRepository) ExistsByID(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Broadcast{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
