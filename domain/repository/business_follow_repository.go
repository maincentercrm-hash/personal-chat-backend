// domain/repository/business_follow_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type BusinessFollowRepository interface {
	// Follow - ผู้ใช้ติดตามธุรกิจ
	Follow(follow *models.UserBusinessFollow) error

	// Unfollow - ผู้ใช้เลิกติดตามธุรกิจ
	Unfollow(userID, businessID uuid.UUID) error

	// IsFollowing - ตรวจสอบว่าผู้ใช้ติดตามธุรกิจอยู่หรือไม่
	IsFollowing(userID, businessID uuid.UUID) (bool, error)

	// GetFollowers - ดึงรายชื่อผู้ติดตามของธุรกิจ
	GetFollowers(businessID uuid.UUID, limit, offset int) ([]*models.UserBusinessFollow, int64, error)

	// GetFollowedBusinesses - ดึงรายชื่อธุรกิจที่ผู้ใช้ติดตาม
	GetFollowedBusinesses(userID uuid.UUID, limit, offset int) ([]*models.UserBusinessFollow, int64, error)

	// GetFollowerCount - นับจำนวนผู้ติดตามของธุรกิจ
	GetFollowerCount(businessID uuid.UUID) (int64, error)

	// CountFollowers นับจำนวนผู้ติดตามทั้งหมดของธุรกิจ
	CountFollowers(businessID uuid.UUID) (int64, error)

	// GetAllFollowerIDs ดึงรายการ ID ของผู้ติดตามทั้งหมดของธุรกิจ
	GetAllFollowerIDs(businessID uuid.UUID) ([]uuid.UUID, error)
}
