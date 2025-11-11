// infrastructure/persistence/postgres/sticker_repository.go
package postgres

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type stickerRepository struct {
	db *gorm.DB
}

// NewStickerRepository สร้าง repository สำหรับจัดการข้อมูลสติกเกอร์
func NewStickerRepository(db *gorm.DB) repository.StickerRepository {
	return &stickerRepository{db: db}
}

// CreateStickerSet สร้างชุดสติกเกอร์ใหม่
func (r *stickerRepository) CreateStickerSet(stickerSet *models.StickerSet) error {
	return r.db.Create(stickerSet).Error
}

// GetStickerSetByID ดึงข้อมูลชุดสติกเกอร์ตาม ID
func (r *stickerRepository) GetStickerSetByID(id uuid.UUID) (*models.StickerSet, error) {
	var stickerSet models.StickerSet
	err := r.db.First(&stickerSet, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &stickerSet, nil
}

// GetAllStickerSets ดึงข้อมูลชุดสติกเกอร์ทั้งหมด
func (r *stickerRepository) GetAllStickerSets(limit, offset int) ([]*models.StickerSet, int64, error) {
	var stickerSets []*models.StickerSet
	var total int64

	// นับจำนวนทั้งหมด
	if err := r.db.Model(&models.StickerSet{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม limit และ offset
	err := r.db.Order("sort_order ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&stickerSets).Error

	if err != nil {
		return nil, 0, err
	}

	return stickerSets, total, nil
}

// GetDefaultStickerSets ดึงข้อมูลชุดสติกเกอร์เริ่มต้น
func (r *stickerRepository) GetDefaultStickerSets() ([]*models.StickerSet, error) {
	var stickerSets []*models.StickerSet
	err := r.db.Where("is_default = ?", true).
		Order("sort_order ASC, created_at DESC").
		Find(&stickerSets).Error

	if err != nil {
		return nil, err
	}

	return stickerSets, nil
}

// UpdateStickerSet อัปเดตข้อมูลชุดสติกเกอร์
func (r *stickerRepository) UpdateStickerSet(stickerSet *models.StickerSet) error {
	return r.db.Save(stickerSet).Error
}

// DeleteStickerSet ลบชุดสติกเกอร์
func (r *stickerRepository) DeleteStickerSet(id uuid.UUID) error {
	// ลบสติกเกอร์ทั้งหมดในชุดก่อน
	if err := r.db.Where("sticker_set_id = ?", id).Delete(&models.Sticker{}).Error; err != nil {
		return err
	}

	// ลบความสัมพันธ์กับผู้ใช้
	if err := r.db.Where("sticker_set_id = ?", id).Delete(&models.UserStickerSet{}).Error; err != nil {
		return err
	}

	// ลบชุดสติกเกอร์
	return r.db.Delete(&models.StickerSet{}, "id = ?", id).Error
}

// CreateSticker สร้างสติกเกอร์ใหม่
func (r *stickerRepository) CreateSticker(sticker *models.Sticker) error {
	return r.db.Create(sticker).Error
}

// GetStickerByID ดึงข้อมูลสติกเกอร์ตาม ID
func (r *stickerRepository) GetStickerByID(id uuid.UUID) (*models.Sticker, error) {
	var sticker models.Sticker
	err := r.db.First(&sticker, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &sticker, nil
}

// GetStickersBySetID ดึงข้อมูลสติกเกอร์ทั้งหมดในชุด
func (r *stickerRepository) GetStickersBySetID(setID uuid.UUID) ([]*models.Sticker, error) {
	var stickers []*models.Sticker
	err := r.db.Where("sticker_set_id = ?", setID).
		Order("sort_order ASC").
		Find(&stickers).Error

	if err != nil {
		return nil, err
	}

	return stickers, nil
}

// UpdateSticker อัปเดตข้อมูลสติกเกอร์
func (r *stickerRepository) UpdateSticker(sticker *models.Sticker) error {
	return r.db.Save(sticker).Error
}

// DeleteSticker ลบสติกเกอร์
func (r *stickerRepository) DeleteSticker(id uuid.UUID) error {
	// ลบประวัติการใช้สติกเกอร์
	if err := r.db.Where("sticker_id = ?", id).Delete(&models.UserRecentSticker{}).Error; err != nil {
		return err
	}

	// ลบสติกเกอร์โปรด
	if err := r.db.Where("sticker_id = ?", id).Delete(&models.UserFavoriteSticker{}).Error; err != nil {
		return err
	}

	// ลบสติกเกอร์
	return r.db.Delete(&models.Sticker{}, "id = ?", id).Error
}

// AddStickerSetToUser เพิ่มชุดสติกเกอร์ให้ผู้ใช้
func (r *stickerRepository) AddStickerSetToUser(userStickerSet *models.UserStickerSet) error {
	// ตรวจสอบว่ามีอยู่แล้วหรือไม่
	var count int64
	r.db.Model(&models.UserStickerSet{}).
		Where("user_id = ? AND sticker_set_id = ?", userStickerSet.UserID, userStickerSet.StickerSetID).
		Count(&count)

	if count > 0 {
		// มีอยู่แล้ว ไม่ต้องเพิ่มใหม่
		return nil
	}

	return r.db.Create(userStickerSet).Error
}

// GetUserStickerSets ดึงชุดสติกเกอร์ของผู้ใช้
func (r *stickerRepository) GetUserStickerSets(userID uuid.UUID) ([]*models.StickerSet, error) {
	var stickerSets []*models.StickerSet

	err := r.db.Model(&models.StickerSet{}).
		Joins("JOIN user_sticker_sets ON user_sticker_sets.sticker_set_id = sticker_sets.id").
		Where("user_sticker_sets.user_id = ?", userID).
		Order("user_sticker_sets.is_favorite DESC, sticker_sets.sort_order ASC").
		Find(&stickerSets).Error

	if err != nil {
		return nil, err
	}

	return stickerSets, nil
}

// SetStickerSetAsFavorite ตั้งค่าชุดสติกเกอร์เป็นรายการโปรด
func (r *stickerRepository) SetStickerSetAsFavorite(userID, stickerSetID uuid.UUID, isFavorite bool) error {
	return r.db.Model(&models.UserStickerSet{}).
		Where("user_id = ? AND sticker_set_id = ?", userID, stickerSetID).
		Update("is_favorite", isFavorite).Error
}

// RemoveStickerSetFromUser ลบชุดสติกเกอร์ออกจากผู้ใช้
func (r *stickerRepository) RemoveStickerSetFromUser(userID, stickerSetID uuid.UUID) error {
	return r.db.Where("user_id = ? AND sticker_set_id = ?", userID, stickerSetID).
		Delete(&models.UserStickerSet{}).Error
}

// AddRecentSticker บันทึกการใช้งานสติกเกอร์ล่าสุด
func (r *stickerRepository) AddRecentSticker(userRecentSticker *models.UserRecentSticker) error {
	// ลบรายการเก่าของสติกเกอร์เดียวกัน (ถ้ามี)
	r.db.Where("user_id = ? AND sticker_id = ?", userRecentSticker.UserID, userRecentSticker.StickerID).
		Delete(&models.UserRecentSticker{})

	// สร้างรายการใหม่
	return r.db.Create(userRecentSticker).Error
}

// GetUserRecentStickers ดึงสติกเกอร์ที่ใช้ล่าสุดของผู้ใช้
func (r *stickerRepository) GetUserRecentStickers(userID uuid.UUID, limit int) ([]*models.Sticker, error) {
	var stickers []*models.Sticker

	err := r.db.Model(&models.Sticker{}).
		Joins("JOIN user_recent_stickers ON user_recent_stickers.sticker_id = stickers.id").
		Where("user_recent_stickers.user_id = ?", userID).
		Order("user_recent_stickers.used_at DESC").
		Limit(limit).
		Find(&stickers).Error

	if err != nil {
		return nil, err
	}

	return stickers, nil
}

// GetUserFavoriteStickers ดึงสติกเกอร์โปรดของผู้ใช้
func (r *stickerRepository) GetUserFavoriteStickers(userID uuid.UUID) ([]*models.Sticker, error) {
	var stickers []*models.Sticker

	err := r.db.Model(&models.Sticker{}).
		Joins("JOIN user_favorite_stickers ON user_favorite_stickers.sticker_id = stickers.id").
		Where("user_favorite_stickers.user_id = ?", userID).
		Order("user_favorite_stickers.created_at DESC").
		Find(&stickers).Error

	if err != nil {
		return nil, err
	}

	return stickers, nil
}
