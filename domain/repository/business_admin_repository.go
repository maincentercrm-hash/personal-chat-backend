// domain/repository/business_admin_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type BusinessAdminRepository interface {
	// Create เพิ่มแอดมินให้ธุรกิจ
	Create(admin *models.BusinessAdmin) error

	// GetByID ดึงข้อมูลแอดมินตาม ID
	GetByID(id uuid.UUID) (*models.BusinessAdmin, error)

	// GetByUserAndBusinessID ดึงข้อมูลแอดมิน
	GetByUserAndBusinessID(userID, businessID uuid.UUID) (*models.BusinessAdmin, error)

	// GetAdminsByBusinessID ดึงรายชื่อแอดมินของธุรกิจ
	GetAdminsByBusinessID(businessID uuid.UUID) ([]*models.BusinessAdmin, error)

	// Update อัพเดทข้อมูลแอดมิน
	Update(admin *models.BusinessAdmin) error

	// Delete ลบแอดมินออกจากธุรกิจ
	Delete(id uuid.UUID) error

	// DeleteByUserAndBusinessID ลบแอดมินโดยใช้ userID และ businessID
	DeleteByUserAndBusinessID(userID, businessID uuid.UUID) error

	// CheckAdminPermission ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่และมีบทบาทตามที่กำหนดหรือไม่
	CheckAdminPermission(userID, businessID uuid.UUID, allowedRoles []string) (bool, error)

	// GetAdminsByBusiness ดึงรายการแอดมินทั้งหมดของธุรกิจ
	GetAdminsByBusiness(businessID uuid.UUID) ([]*models.BusinessAdmin, error)
}
