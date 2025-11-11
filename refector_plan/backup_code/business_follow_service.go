// application/serviceimpl/business_follow_service.go
package serviceimpl

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

type businessFollowService struct {
	businessFollowRepo repository.BusinessFollowRepository
	businessRepo       repository.BusinessAccountRepository
	userRepo           repository.UserRepository
	analyticsRepo      repository.AnalyticsDailyRepository // จะต้องสร้างเพิ่มในภายหลัง
}

// NewBusinessFollowService สร้าง instance ใหม่ของ BusinessFollowService
func NewBusinessFollowService(
	businessFollowRepo repository.BusinessFollowRepository,
	businessRepo repository.BusinessAccountRepository,
	userRepo repository.UserRepository,
	analyticsRepo repository.AnalyticsDailyRepository,
) service.BusinessFollowService {
	return &businessFollowService{
		businessFollowRepo: businessFollowRepo,
		businessRepo:       businessRepo,
		userRepo:           userRepo,
		analyticsRepo:      analyticsRepo,
	}
}

// FollowBusiness - ผู้ใช้ติดตามธุรกิจ
// application/serviceimpl/business_follow_service.go

// ปรับปรุงเมธอด FollowBusiness และ UnfollowBusiness ที่มีการใช้งาน Analytics

// FollowBusiness - ผู้ใช้ติดตามธุรกิจ
func (s *businessFollowService) FollowBusiness(userID, businessID uuid.UUID, source string) error {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessRepo.ExistsById(businessID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("business not found")
	}

	// ตรวจสอบว่าผู้ใช้มีอยู่จริง
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// ตรวจสอบว่าไม่ใช่การติดตามตัวเอง
	business, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return err
	}

	// ถ้าธุรกิจมีเจ้าของ และเจ้าของคือผู้ใช้ที่กำลังจะติดตาม
	if business.OwnerID != nil && *business.OwnerID == userID {
		return errors.New("cannot follow your own business")
	}

	// สร้างข้อมูลการติดตาม
	follow := &models.UserBusinessFollow{
		ID:         uuid.New(),
		UserID:     userID,
		BusinessID: businessID,
		FollowedAt: time.Now(),
		Source:     source,
		User:       user,
		Business:   business,
	}

	// บันทึกการติดตาม
	isFollowing, err := s.businessFollowRepo.IsFollowing(userID, businessID)
	if err != nil {
		return err
	}

	// ถ้าไม่ได้ติดตามอยู่แล้ว ให้เพิ่มสถิติด้วย
	if !isFollowing {
		// เพิ่มการติดตามและบันทึกสถิติ
		err = s.businessFollowRepo.Follow(follow)
		if err != nil {
			return err
		}

		// บันทึกสถิติการติดตามใหม่
		today := time.Now().Truncate(24 * time.Hour)
		s.analyticsRepo.IncrementNewFollowers(businessID, today)

		return nil
	}

	// ถ้าติดตามอยู่แล้ว ไม่ต้องทำอะไร (ถือว่าสำเร็จ)
	return nil
}

// UnfollowBusiness - ผู้ใช้เลิกติดตามธุรกิจ
func (s *businessFollowService) UnfollowBusiness(userID, businessID uuid.UUID) error {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessRepo.ExistsById(businessID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("business not found")
	}

	// ตรวจสอบว่ามีการติดตามอยู่
	isFollowing, err := s.businessFollowRepo.IsFollowing(userID, businessID)
	if err != nil {
		return err
	}

	// ถ้ามีการติดตามอยู่ ให้ลบการติดตามและบันทึกสถิติ
	if isFollowing {
		// ลบการติดตาม
		if err := s.businessFollowRepo.Unfollow(userID, businessID); err != nil {
			return err
		}

		// บันทึกสถิติการเลิกติดตาม
		today := time.Now().Truncate(24 * time.Hour)
		s.analyticsRepo.IncrementUnfollows(businessID, today)
	}

	return nil
}

// IsFollowing - ตรวจสอบว่าผู้ใช้ติดตามธุรกิจอยู่หรือไม่
func (s *businessFollowService) IsFollowing(userID, businessID uuid.UUID) (bool, error) {
	return s.businessFollowRepo.IsFollowing(userID, businessID)
}

// GetBusinessFollowers - ดึงรายชื่อผู้ติดตามของธุรกิจ
func (s *businessFollowService) GetBusinessFollowers(businessID uuid.UUID, limit, offset int) ([]*models.User, int64, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessRepo.ExistsById(businessID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, errors.New("business not found")
	}

	// ดึงข้อมูลผู้ติดตาม
	followers, total, err := s.businessFollowRepo.GetFollowers(businessID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็นรายชื่อผู้ใช้
	users := make([]*models.User, 0, len(followers))
	for _, follower := range followers {
		if follower.User != nil {
			users = append(users, follower.User)
		}
	}

	return users, total, nil
}

// GetUserFollowedBusinesses - ดึงรายชื่อธุรกิจที่ผู้ใช้ติดตาม
func (s *businessFollowService) GetUserFollowedBusinesses(userID uuid.UUID, limit, offset int) ([]*models.BusinessAccount, int64, error) {
	// ตรวจสอบว่าผู้ใช้มีอยู่จริง
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, 0, errors.New("user not found")
	}

	// ดึงข้อมูลธุรกิจที่ติดตาม
	followed, total, err := s.businessFollowRepo.GetFollowedBusinesses(userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็นรายชื่อธุรกิจ
	businesses := make([]*models.BusinessAccount, 0, len(followed))
	for _, follow := range followed {
		if follow.Business != nil {
			businesses = append(businesses, follow.Business)
		}
	}

	return businesses, total, nil
}
