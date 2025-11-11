// infrastructure/persistence/postgres/business_admin_repository.go
package postgres

import (
	"errors"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type businessAdminRepository struct {
	db *gorm.DB
}

// NewBusinessAdminRepository สร้าง instance ใหม่ของ BusinessAdminRepository
func NewBusinessAdminRepository(db *gorm.DB) repository.BusinessAdminRepository {
	return &businessAdminRepository{db: db}
}

// Create เพิ่มแอดมินให้ธุรกิจ
func (r *businessAdminRepository) Create(admin *models.BusinessAdmin) error {
	return r.db.Create(admin).Error
}

// GetByID ดึงข้อมูลแอดมินตาม ID
func (r *businessAdminRepository) GetByID(id uuid.UUID) (*models.BusinessAdmin, error) {
	var admin models.BusinessAdmin
	if err := r.db.First(&admin, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}
	return &admin, nil
}

// GetByUserAndBusinessID ดึงข้อมูลแอดมิน
func (r *businessAdminRepository) GetByUserAndBusinessID(userID, businessID uuid.UUID) (*models.BusinessAdmin, error) {
	var admin models.BusinessAdmin
	if err := r.db.Where("business_id = ? AND user_id = ?", businessID, userID).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}
	return &admin, nil
}

// GetAdminsByBusinessID ดึงรายชื่อแอดมินของธุรกิจ
func (r *businessAdminRepository) GetAdminsByBusinessID(businessID uuid.UUID) ([]*models.BusinessAdmin, error) {
	var admins []*models.BusinessAdmin
	if err := r.db.Where("business_id = ?", businessID).Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

// Update อัพเดทข้อมูลแอดมิน
func (r *businessAdminRepository) Update(admin *models.BusinessAdmin) error {
	return r.db.Save(admin).Error
}

// Delete ลบแอดมินออกจากธุรกิจ
func (r *businessAdminRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&models.BusinessAdmin{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("admin not found")
	}
	return nil
}

// DeleteByUserAndBusinessID ลบแอดมินโดยใช้ userID และ businessID
func (r *businessAdminRepository) DeleteByUserAndBusinessID(userID, businessID uuid.UUID) error {
	// ตรวจสอบว่าไม่ได้ลบ owner
	var admin models.BusinessAdmin
	if err := r.db.Where("business_id = ? AND user_id = ?", businessID, userID).First(&admin).Error; err != nil {
		return err
	}

	if admin.Role == "owner" {
		return errors.New("cannot remove the owner")
	}

	result := r.db.Where("business_id = ? AND user_id = ?", businessID, userID).Delete(&models.BusinessAdmin{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("admin not found")
	}
	return nil
}

// CheckAdminPermission ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่และมีบทบาทตามที่กำหนดหรือไม่
func (r *businessAdminRepository) CheckAdminPermission(userID, businessID uuid.UUID, allowedRoles []string) (bool, error) {
	var admin models.BusinessAdmin
	err := r.db.Where("business_id = ? AND user_id = ?", businessID, userID).First(&admin).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// ถ้าไม่ได้ระบุบทบาทที่อนุญาต หมายถึงตรวจสอบเพียงว่าเป็นแอดมินหรือไม่
	if len(allowedRoles) == 0 {
		return true, nil
	}

	// ตรวจสอบว่าผู้ใช้มีบทบาทที่อนุญาตหรือไม่
	for _, role := range allowedRoles {
		if admin.Role == role {
			return true, nil
		}
	}

	return false, nil
}

func (r *businessAdminRepository) GetAdminsByBusiness(businessID uuid.UUID) ([]*models.BusinessAdmin, error) {
	var admins []*models.BusinessAdmin
	err := r.db.Where("business_id = ?", businessID).Find(&admins).Error
	return admins, err
}
