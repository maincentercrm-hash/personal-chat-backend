// domain/service/business_follow_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type BusinessFollowService interface {
	// FollowBusiness - ผู้ใช้ติดตามธุรกิจ
	FollowBusiness(userID, businessID uuid.UUID, source string) error

	// UnfollowBusiness - ผู้ใช้เลิกติดตามธุรกิจ
	UnfollowBusiness(userID, businessID uuid.UUID) error

	// IsFollowing - ตรวจสอบว่าผู้ใช้ติดตามธุรกิจอยู่หรือไม่
	IsFollowing(userID, businessID uuid.UUID) (bool, error)

	// GetBusinessFollowers - ดึงรายชื่อผู้ติดตามของธุรกิจ
	GetBusinessFollowers(businessID uuid.UUID, limit, offset int) ([]*models.User, int64, error)

	// GetUserFollowedBusinesses - ดึงรายชื่อธุรกิจที่ผู้ใช้ติดตาม
	GetUserFollowedBusinesses(userID uuid.UUID, limit, offset int) ([]*models.BusinessAccount, int64, error)
}
