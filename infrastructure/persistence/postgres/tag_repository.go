// infrastructure/persistence/postgres/tag_repository.go
package postgres

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository สร้าง instance ใหม่ของ TagRepository
func NewTagRepository(db *gorm.DB) repository.TagRepository {
	return &tagRepository{db: db}
}

// Create สร้าง Tag ใหม่
func (r *tagRepository) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

// GetByBusinessID ดึง Tags ทั้งหมดของธุรกิจ
func (r *tagRepository) GetByBusinessID(businessID uuid.UUID) ([]*models.Tag, error) {
	var tags []*models.Tag

	err := r.db.
		Preload("Business").
		Preload("Creator").
		Where("business_id = ?", businessID).
		Order("name ASC").
		Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

// GetByID ดึง Tag ตาม ID
func (r *tagRepository) GetByID(id uuid.UUID) (*models.Tag, error) {
	var tag models.Tag

	err := r.db.
		Preload("Business").
		Preload("Creator").
		Where("id = ?", id).
		First(&tag).Error

	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// Update อัปเดต Tag
func (r *tagRepository) Update(tag *models.Tag) error {
	return r.db.Save(tag).Error
}

// Delete ลบ Tag
func (r *tagRepository) Delete(id uuid.UUID) error {
	// ลบ Tag (จะลบ UserTags ที่เกี่ยวข้องด้วย via CASCADE foreign key)
	return r.db.Where("id = ?", id).Delete(&models.Tag{}).Error
}

// GetByName ดึง Tag ตามชื่อในธุรกิจ
func (r *tagRepository) GetByName(businessID uuid.UUID, name string) (*models.Tag, error) {
	var tag models.Tag

	err := r.db.
		Preload("Business").
		Preload("Creator").
		Where("business_id = ? AND name = ?", businessID, name).
		First(&tag).Error

	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// ExistsByName ตรวจสอบว่ามีชื่อแท็กนี้อยู่แล้วหรือไม่
func (r *tagRepository) ExistsByName(businessID uuid.UUID, name string) (bool, error) {
	var count int64

	err := r.db.Model(&models.Tag{}).
		Where("business_id = ? AND name = ?", businessID, name).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetTagsWithUserCount ดึง Tags พร้อมจำนวนผู้ใช้
func (r *tagRepository) GetTagsWithUserCount(businessID uuid.UUID) ([]TagWithUserCount, error) {
	var results []TagWithUserCount

	err := r.db.Table("tags").
		Select("tags.*, COUNT(user_tags.id) as user_count").
		Joins("LEFT JOIN user_tags ON tags.id = user_tags.tag_id").
		Where("tags.business_id = ?", businessID).
		Group("tags.id").
		Order("tags.name ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetPopularTags ดึงแท็กยอดนิยม (เรียงตามจำนวนผู้ใช้มากไปน้อย)
func (r *tagRepository) GetPopularTags(businessID uuid.UUID, limit int) ([]TagWithUserCount, error) {
	var results []TagWithUserCount

	if limit <= 0 {
		limit = 10 // ค่าเริ่มต้น
	}

	err := r.db.Table("tags").
		Select("tags.*, COUNT(user_tags.id) as user_count").
		Joins("LEFT JOIN user_tags ON tags.id = user_tags.tag_id").
		Where("tags.business_id = ?", businessID).
		Group("tags.id").
		Having("COUNT(user_tags.id) > 0").
		Order("user_count DESC, tags.name ASC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUnusedTags ดึงแท็กที่ไม่มีผู้ใช้
func (r *tagRepository) GetUnusedTags(businessID uuid.UUID) ([]*models.Tag, error) {
	var tags []*models.Tag

	err := r.db.
		Preload("Business").
		Preload("Creator").
		Where("business_id = ?", businessID).
		Where("id NOT IN (SELECT DISTINCT tag_id FROM user_tags WHERE business_id = ?)", businessID).
		Order("name ASC").
		Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

// SearchTags ค้นหาแท็ก
func (r *tagRepository) SearchTags(businessID uuid.UUID, query string, limit, offset int) ([]*models.Tag, int64, error) {
	var tags []*models.Tag
	var total int64

	// Base query
	baseQuery := r.db.Model(&models.Tag{}).
		Where("business_id = ?", businessID)

	// เพิ่มเงื่อนไขการค้นหา
	if query != "" {
		baseQuery = baseQuery.Where("name ILIKE ?", "%"+query+"%")
	}

	// นับจำนวนทั้งหมด
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	searchQuery := r.db.
		Preload("Business").
		Preload("Creator").
		Where("business_id = ?", businessID)

	if query != "" {
		searchQuery = searchQuery.Where("name ILIKE ?", "%"+query+"%")
	}

	err = searchQuery.
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&tags).Error

	if err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// GetTagsByColor ดึงแท็กตามสี
func (r *tagRepository) GetTagsByColor(businessID uuid.UUID, color string) ([]*models.Tag, error) {
	var tags []*models.Tag

	err := r.db.
		Preload("Business").
		Preload("Creator").
		Where("business_id = ? AND color = ?", businessID, color).
		Order("name ASC").
		Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

// BulkCreate สร้างแท็กหลายตัวพร้อมกัน
func (r *tagRepository) BulkCreate(tags []*models.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return r.db.Create(&tags).Error
}

// BulkDelete ลบแท็กหลายตัวพร้อมกัน
func (r *tagRepository) BulkDelete(tagIDs []uuid.UUID) error {
	if len(tagIDs) == 0 {
		return nil
	}

	return r.db.Where("id IN ?", tagIDs).Delete(&models.Tag{}).Error
}

// GetTagStatistics ดึงสถิติแท็ก
func (r *tagRepository) GetTagStatistics(businessID uuid.UUID) (*TagStatistics, error) {
	var stats TagStatistics

	// นับจำนวนแท็กทั้งหมด
	err := r.db.Model(&models.Tag{}).
		Where("business_id = ?", businessID).
		Count(&stats.TotalTags).Error
	if err != nil {
		return nil, err
	}

	// นับจำนวนแท็กที่มีผู้ใช้
	err = r.db.Model(&models.Tag{}).
		Where("business_id = ?", businessID).
		Where("id IN (SELECT DISTINCT tag_id FROM user_tags WHERE business_id = ?)", businessID).
		Count(&stats.UsedTags).Error
	if err != nil {
		return nil, err
	}

	// นับจำนวนแท็กที่ไม่มีผู้ใช้
	stats.UnusedTags = stats.TotalTags - stats.UsedTags

	// นับจำนวน user_tags ทั้งหมด
	err = r.db.Table("user_tags").
		Where("business_id = ?", businessID).
		Count(&stats.TotalUserTags).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *tagRepository) GetBusinessTagsWithUserCount(businessID uuid.UUID) ([]dto.TagInfo, error) {
	var tagInfos []dto.TagInfo

	// ใช้ Raw SQL กับ GORM
	// GORM ใช้ Raw สำหรับการเขียน Raw SQL
	// โดยข้อมูลจะถูก scan เข้าไปใน struct ที่กำหนดได้โดยตรง
	err := r.db.Raw(`
        SELECT 
            t.id, t.business_id, t.name, t.color, t.created_at, t.created_by_id,
            COUNT(ut.user_id) as user_count
        FROM 
            tags t
        LEFT JOIN 
            user_tags ut ON t.id = ut.tag_id AND ut.business_id = t.business_id
        WHERE 
            t.business_id = ?
        GROUP BY 
            t.id
    `, businessID).Scan(&tagInfos).Error

	if err != nil {
		return nil, err
	}

	return tagInfos, nil
}

// TagWithUserCount โครงสร้างสำหรับเก็บแท็กพร้อมจำนวนผู้ใช้
type TagWithUserCount struct {
	models.Tag
	UserCount int64 `json:"user_count"`
}

// TagStatistics สถิติของแท็ก
type TagStatistics struct {
	TotalTags     int64 `json:"total_tags"`
	UsedTags      int64 `json:"used_tags"`
	UnusedTags    int64 `json:"unused_tags"`
	TotalUserTags int64 `json:"total_user_tags"`
}
