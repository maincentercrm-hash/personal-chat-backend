// infrastructure/persistence/postgres/business_account_repository.go
package postgres

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type businessAccountRepository struct {
	db *gorm.DB
}

// NewBusinessAccountRepository สร้าง instance ใหม่ของ BusinessAccountRepository
func NewBusinessAccountRepository(db *gorm.DB) repository.BusinessAccountRepository {
	return &businessAccountRepository{db: db}
}

// GetByID ดึงข้อมูลธุรกิจตาม ID
func (r *businessAccountRepository) GetByID(id uuid.UUID) (*models.BusinessAccount, error) {
	var business models.BusinessAccount
	if err := r.db.First(&business, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &business, nil
}

// GetByUsername ดึงข้อมูลธุรกิจตาม username
func (r *businessAccountRepository) GetByUsername(username string) (*models.BusinessAccount, error) {
	var business models.BusinessAccount
	if err := r.db.Where("username = ?", username).First(&business).Error; err != nil {
		return nil, err
	}
	return &business, nil
}

// GetBusinessesByUserID ดึงรายการธุรกิจที่ผู้ใช้เป็นแอดมิน
func (r *businessAccountRepository) GetBusinessesByUserID(userID uuid.UUID) ([]*models.BusinessAccount, error) {
	// ดึงรายการ business_id ที่ผู้ใช้เป็นแอดมิน
	var admins []models.BusinessAdmin
	if err := r.db.Where("user_id = ?", userID).Find(&admins).Error; err != nil {
		return nil, err
	}

	if len(admins) == 0 {
		return []*models.BusinessAccount{}, nil
	}

	// สร้าง slice ของ business IDs
	var businessIDs []uuid.UUID
	businessRoles := make(map[uuid.UUID]string)
	for _, admin := range admins {
		businessIDs = append(businessIDs, admin.BusinessID)
		businessRoles[admin.BusinessID] = admin.Role
	}

	// ดึงข้อมูลธุรกิจ
	var businesses []*models.BusinessAccount
	if err := r.db.Where("id IN ?", businessIDs).Find(&businesses).Error; err != nil {
		return nil, err
	}

	// หมายเหตุ: เราไม่สามารถกำหนดค่า UserRole ได้โดยตรงเนื่องจากไม่มีในโมเดล
	// แต่สามารถทำได้ในชั้น service โดยการสร้าง DTO หรือ response struct ที่มีฟิลด์นี้

	return businesses, nil
}

// Create สร้างธุรกิจใหม่
func (r *businessAccountRepository) Create(business *models.BusinessAccount) error {
	tx := r.db.Begin()

	// ตรวจสอบว่า username ซ้ำหรือไม่
	var count int64
	if err := tx.Model(&models.BusinessAccount{}).Where("username = ?", business.Username).Count(&count).Error; err != nil {
		tx.Rollback()
		return err
	}

	if count > 0 {
		tx.Rollback()
		return errors.New("username is already taken")
	}

	// สร้างข้อมูลธุรกิจ
	if err := tx.Create(business).Error; err != nil {
		tx.Rollback()
		return err
	}

	// เพิ่มผู้ใช้เป็นแอดมินของธุรกิจ (owner)
	admin := models.BusinessAdmin{
		BusinessID: business.ID,
		UserID:     *business.OwnerID, // ปรับให้เข้ากับโมเดลของคุณที่ OwnerID เป็น pointer
		Role:       "owner",
	}

	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Update อัพเดทข้อมูลธุรกิจ
func (r *businessAccountRepository) Update(business *models.BusinessAccount) error {
	return r.db.Save(business).Error
}

// Delete ลบธุรกิจ (เปลี่ยนสถานะเป็น deleted)
func (r *businessAccountRepository) Delete(id uuid.UUID) error {
	result := r.db.Model(&models.BusinessAccount{}).Where("id = ?", id).Update("status", "deleted")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("business not found")
	}
	return nil
}

// GetFollowerCount นับจำนวนผู้ติดตามของธุรกิจ
func (r *businessAccountRepository) GetFollowerCount(businessID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserBusinessFollow{}).Where("business_id = ?", businessID).Count(&count).Error
	return count, err
}

// IsFollowing ตรวจสอบว่าผู้ใช้ติดตามธุรกิจนี้หรือไม่
func (r *businessAccountRepository) IsFollowing(userID uuid.UUID, businessID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserBusinessFollow{}).Where("user_id = ? AND business_id = ?", userID, businessID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *businessAccountRepository) ExistsById(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.BusinessAccount{}).Where("id = ? AND status = ?", id, "active").Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SearchBusinesses ค้นหาธุรกิจ
func (r *businessAccountRepository) SearchBusinesses(query string, limit, offset int) ([]*models.BusinessAccount, int64, error) {
	var businesses []*models.BusinessAccount
	var total int64

	// ตัดช่องว่างและเตรียมคำค้นหา
	searchQuery := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"

	// นับจำนวนผลลัพธ์ทั้งหมด
	err := r.db.Model(&models.BusinessAccount{}).
		Where("(LOWER(name) LIKE ? OR LOWER(username) LIKE ? OR LOWER(description) LIKE ?) AND status = ?",
			searchQuery, searchQuery, searchQuery, "active").
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม limit และ offset
	err = r.db.Where("(LOWER(name) LIKE ? OR LOWER(username) LIKE ? OR LOWER(description) LIKE ?) AND status = ?",
		searchQuery, searchQuery, searchQuery, "active").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&businesses).Error

	if err != nil {
		return nil, 0, err
	}

	return businesses, total, nil
}

// GetByUsernameExact ดึงข้อมูลธุรกิจตาม username แบบตรงกับทั้งหมด
func (r *businessAccountRepository) GetByUsernameExact(username string) (*models.BusinessAccount, error) {
	var business models.BusinessAccount
	// ใช้การเปรียบเทียบแบบเท่ากันเท่านั้น ไม่ใช้ LIKE
	if err := r.db.Where("username = ?", username).First(&business).Error; err != nil {
		return nil, err
	}
	return &business, nil
}

// SearchBusinessesExact ค้นหาธุรกิจแบบตรงกับทั้งหมด
func (r *businessAccountRepository) SearchBusinessesExact(query string, limit, offset int) ([]*models.BusinessAccount, int64, error) {
	var businesses []*models.BusinessAccount
	var total int64

	// นับจำนวนผลลัพธ์ทั้งหมด - ใช้การเปรียบเทียบแบบเท่ากันเท่านั้น ไม่ใช้ LIKE
	err := r.db.Model(&models.BusinessAccount{}).
		Where("(name = ? OR username = ? OR description = ?) AND status = ?",
			query, query, query, "active").
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม limit และ offset
	err = r.db.Where("(name = ? OR username = ? OR description = ?) AND status = ?",
		query, query, query, "active").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&businesses).Error

	if err != nil {
		return nil, 0, err
	}

	return businesses, total, nil
}
