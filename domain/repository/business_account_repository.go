// domain/repository/business_account_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type BusinessAccountRepository interface {
	// GetByID ดึงข้อมูลธุรกิจตาม ID
	GetByID(id uuid.UUID) (*models.BusinessAccount, error)

	// GetByUsername ดึงข้อมูลธุรกิจตาม username
	GetByUsername(username string) (*models.BusinessAccount, error)

	// GetBusinessesByUserID ดึงรายการธุรกิจที่ผู้ใช้เป็นแอดมิน
	GetBusinessesByUserID(userID uuid.UUID) ([]*models.BusinessAccount, error)

	// Create สร้างธุรกิจใหม่
	Create(business *models.BusinessAccount) error

	// Update อัพเดทข้อมูลธุรกิจ
	Update(business *models.BusinessAccount) error

	// Delete ลบธุรกิจ (เปลี่ยนสถานะเป็น deleted)
	Delete(id uuid.UUID) error

	// GetFollowerCount นับจำนวนผู้ติดตามของธุรกิจ
	GetFollowerCount(businessID uuid.UUID) (int64, error)

	// IsFollowing ตรวจสอบว่าผู้ใช้ติดตามธุรกิจนี้หรือไม่
	IsFollowing(userID uuid.UUID, businessID uuid.UUID) (bool, error)

	// ExistsById ตรวจสอบว่าธุรกิจมีอยู่จริงหรือไม่
	ExistsById(id uuid.UUID) (bool, error)

	// SearchBusinesses ค้นหาธุรกิจ
	SearchBusinesses(query string, limit, offset int) ([]*models.BusinessAccount, int64, error)

	// เพิ่มฟังก์ชันใหม่
	GetByUsernameExact(username string) (*models.BusinessAccount, error)
	SearchBusinessesExact(query string, limit, offset int) ([]*models.BusinessAccount, int64, error)
}
