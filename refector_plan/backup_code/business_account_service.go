// domain/service/business_account_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type BusinessAccountService interface {
	// GetBusinessByID ดึงข้อมูลธุรกิจตาม ID
	GetBusinessByID(id uuid.UUID, userID uuid.UUID) (*models.BusinessAccount, error)

	// GetBusinessByUsername ดึงข้อมูลธุรกิจตาม username
	GetBusinessByUsername(username string, userID uuid.UUID) (*models.BusinessAccount, error)

	// GetUserBusinesses ดึงรายการธุรกิจที่ผู้ใช้เป็นแอดมิน
	GetUserBusinesses(userID uuid.UUID) ([]*models.BusinessAccount, error)

	// GetUserBusinesses ดึงรายการธุรกิจที่ผู้ใช้เป็นแอดมิน และแปลงเป็น DTO
	GetUserBusinessDTOs(userID uuid.UUID) ([]dto.BusinessItem, error)

	// CreateBusiness สร้างธุรกิจใหม่
	CreateBusiness(userID uuid.UUID, name string, username string, description string, welcomeMessage string) (*models.BusinessAccount, error)

	// UpdateBusiness อัพเดทข้อมูลธุรกิจ
	UpdateBusiness(id uuid.UUID, userID uuid.UUID, updateData types.JSONB) (*models.BusinessAccount, error)

	// DeleteBusiness ลบธุรกิจ
	DeleteBusiness(id uuid.UUID, userID uuid.UUID) error

	UploadBusinessProfileImage(id uuid.UUID, userID uuid.UUID, imageURL string) error

	// UploadBusinessCoverImage อัปโหลดรูปปกธุรกิจ
	UploadBusinessCoverImage(id uuid.UUID, userID uuid.UUID, imageURL string) error

	// SearchBusinesses ค้นหาธุรกิจ
	SearchBusinesses(query string, limit, offset int, userID uuid.UUID) ([]*models.BusinessAccount, int64, error)

	// เพิ่มฟังก์ชันใหม่สำหรับการค้นหาแบบตรงกับทั้งหมด
	GetBusinessByUsernameExact(username string, userID uuid.UUID) (*models.BusinessAccount, error)
	SearchBusinessesExact(query string, limit, offset int, userID uuid.UUID) ([]*models.BusinessAccount, int64, error)
}
