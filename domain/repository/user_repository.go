// domain/repository/user_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByUsername(username string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)    // มีอยู่แล้ว
	FindByEmail(email string) (*models.User, error) // เพิ่มเมธอดนี้
	Update(user *models.User) error
	SearchUsers(query string, limit, offset int) ([]*models.User, int, error)

	// ไม่ต้องเพิ่ม GetByID เพราะมี FindByID อยู่แล้ว แค่เปลี่ยนการเรียกใช้ในโค้ดเป็น FindByID แทน
	// เพิ่มฟังก์ชันใหม่
	SearchUsersExact(query string, limit, offset int) ([]*models.User, int64, error)
}
