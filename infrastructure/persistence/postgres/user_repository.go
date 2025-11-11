// /infrastructure/persistence/postgres/user_repository.go
package postgres

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.Status == "" {
		user.Status = "active"
	}

	return r.db.Create(user).Error
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// SearchUsers ค้นหาผู้ใช้ตามคำค้นหา
func (r *userRepository) SearchUsers(query string, limit, offset int) ([]*models.User, int, error) {
	var users []*models.User
	var total int64

	// สร้าง query สำหรับค้นหา
	searchQuery := "%" + strings.ToLower(query) + "%"

	// นับจำนวนผลลัพธ์ทั้งหมด
	err := r.db.Model(&models.User{}).
		Where("LOWER(username) LIKE ? OR LOWER(display_name) LIKE ?", searchQuery, searchQuery).
		Where("status = ?", "active").
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม limit และ offset
	err = r.db.Where("LOWER(username) LIKE ? OR LOWER(display_name) LIKE ?", searchQuery, searchQuery).
		Where("status = ?", "active").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

// SearchUsersExact ค้นหาผู้ใช้แบบตรงกับทั้งหมด
func (r *userRepository) SearchUsersExact(query string, limit, offset int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// นับจำนวนผลลัพธ์ทั้งหมด
	err := r.db.Model(&models.User{}).
		Where("username = ? OR display_name = ?", query, query).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตาม limit และ offset
	err = r.db.Where("username = ? OR display_name = ?", query, query).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// เพิ่มเมธอดนี้
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
