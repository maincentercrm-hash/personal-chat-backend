// infrastructure/persistence/postgres/business_follow_repository.go
package postgres

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type businessFollowRepository struct {
	db *gorm.DB
}

// NewBusinessFollowRepository สร้าง instance ใหม่ของ BusinessFollowRepository
func NewBusinessFollowRepository(db *gorm.DB) repository.BusinessFollowRepository {
	return &businessFollowRepository{
		db: db,
	}
}

// Follow - ผู้ใช้ติดตามธุรกิจ
func (r *businessFollowRepository) Follow(follow *models.UserBusinessFollow) error {
	// ตรวจสอบว่ามีการติดตามอยู่แล้วหรือไม่
	var count int64
	if err := r.db.Model(&models.UserBusinessFollow{}).
		Where("user_id = ? AND business_id = ?", follow.UserID, follow.BusinessID).
		Count(&count).Error; err != nil {
		return err
	}

	// ถ้ามีการติดตามอยู่แล้ว ให้ถือว่าสำเร็จโดยไม่ต้องทำอะไรเพิ่ม
	if count > 0 {
		return nil
	}

	// บันทึกการติดตาม
	return r.db.Create(follow).Error
}

// Unfollow - ผู้ใช้เลิกติดตามธุรกิจ
func (r *businessFollowRepository) Unfollow(userID, businessID uuid.UUID) error {
	result := r.db.Where("user_id = ? AND business_id = ?", userID, businessID).
		Delete(&models.UserBusinessFollow{})

	if result.Error != nil {
		return result.Error
	}

	// ถือว่าสำเร็จแม้ไม่มีการลบเกิดขึ้น (อาจไม่ได้ติดตามอยู่แล้ว)
	return nil
}

// IsFollowing - ตรวจสอบว่าผู้ใช้ติดตามธุรกิจอยู่หรือไม่
func (r *businessFollowRepository) IsFollowing(userID, businessID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserBusinessFollow{}).
		Where("user_id = ? AND business_id = ?", userID, businessID).
		Count(&count).Error

	return count > 0, err
}

// GetFollowers - ดึงรายชื่อผู้ติดตามของธุรกิจ
func (r *businessFollowRepository) GetFollowers(businessID uuid.UUID, limit, offset int) ([]*models.UserBusinessFollow, int64, error) {
	var follows []*models.UserBusinessFollow
	var total int64

	// นับจำนวนทั้งหมด
	if err := r.db.Model(&models.UserBusinessFollow{}).
		Where("business_id = ?", businessID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม limit และ offset
	err := r.db.Where("business_id = ?", businessID).
		Preload("User"). // โหลดข้อมูลผู้ใช้ด้วย
		Order("followed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&follows).Error

	return follows, total, err
}

// GetFollowedBusinesses - ดึงรายชื่อธุรกิจที่ผู้ใช้ติดตาม
func (r *businessFollowRepository) GetFollowedBusinesses(userID uuid.UUID, limit, offset int) ([]*models.UserBusinessFollow, int64, error) {
	var follows []*models.UserBusinessFollow
	var total int64

	// นับจำนวนทั้งหมด
	if err := r.db.Model(&models.UserBusinessFollow{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม limit และ offset
	err := r.db.Where("user_id = ?", userID).
		Preload("Business"). // โหลดข้อมูลธุรกิจด้วย
		Order("followed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&follows).Error

	return follows, total, err
}

// GetFollowerCount - นับจำนวนผู้ติดตามของธุรกิจ
func (r *businessFollowRepository) GetFollowerCount(businessID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserBusinessFollow{}).
		Where("business_id = ?", businessID).
		Count(&count).Error

	return count, err
}

// CountFollowers นับจำนวนผู้ติดตามทั้งหมดของธุรกิจ
func (r *businessFollowRepository) CountFollowers(businessID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserBusinessFollow{}).
		Where("business_id = ?", businessID).
		Count(&count).Error
	return count, err
}

// GetAllFollowerIDs ดึงรายการ ID ของผู้ติดตามทั้งหมดของธุรกิจ
func (r *businessFollowRepository) GetAllFollowerIDs(businessID uuid.UUID) ([]uuid.UUID, error) {
	var follows []models.UserBusinessFollow
	err := r.db.Select("user_id").
		Where("business_id = ?", businessID).
		Find(&follows).Error
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, len(follows))
	for i, follow := range follows {
		userIDs[i] = follow.UserID
	}
	return userIDs, nil
}
