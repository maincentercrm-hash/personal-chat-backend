// domain/service/business_admin_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type BusinessAdminService interface {
	// GetAdmins ดึงรายชื่อแอดมินของธุรกิจ
	GetAdmins(businessID uuid.UUID, userID uuid.UUID) ([]*models.BusinessAdmin, error)

	// AddAdmin เพิ่มแอดมินให้ธุรกิจ
	AddAdmin(businessID uuid.UUID, requestedBy uuid.UUID, newAdminUserID uuid.UUID, role string) (*models.BusinessAdmin, error)

	// RemoveAdmin ลบแอดมินออกจากธุรกิจ
	RemoveAdmin(businessID uuid.UUID, requestedBy uuid.UUID, targetUserID uuid.UUID) error

	// ChangeAdminRole เปลี่ยนบทบาทของแอดมิน
	ChangeAdminRole(businessID uuid.UUID, requestedBy uuid.UUID, targetUserID uuid.UUID, newRole string) (*models.BusinessAdmin, error)

	// CheckAdminPermission ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่และมีบทบาทตามที่กำหนดหรือไม่
	CheckAdminPermission(userID uuid.UUID, businessID uuid.UUID, allowedRoles []string) (bool, error)

	GetAdminByUserAndBusinessID(userID, businessID uuid.UUID) (*models.BusinessAdmin, error) // 🆕 เพิ่มบรรทัดนี้

}
