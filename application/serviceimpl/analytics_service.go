// application/serviceimpl/analytics_service.go
package serviceimpl

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

type analyticsService struct {
	analyticsRepo     repository.AnalyticsDailyRepository
	businessRepo      repository.BusinessAccountRepository
	businessAdminRepo repository.BusinessAdminRepository // เพิ่มตรงนี้
}

// NewAnalyticsService สร้าง instance ใหม่ของ AnalyticsService
func NewAnalyticsService(
	analyticsRepo repository.AnalyticsDailyRepository,
	businessRepo repository.BusinessAccountRepository,
	businessAdminRepo repository.BusinessAdminRepository,
) service.AnalyticsService {
	return &analyticsService{
		analyticsRepo:     analyticsRepo,
		businessRepo:      businessRepo,
		businessAdminRepo: businessAdminRepo,
	}
}

// ตรวจสอบสิทธิ์การเข้าถึงข้อมูลวิเคราะห์
func (s *analyticsService) checkAnalyticsPermission(businessID, userID uuid.UUID) error {
	// ดึงธุรกิจตาม ID
	business, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของหรือ admin
	if business.OwnerID == nil || *business.OwnerID != userID {
		// ตรวจสอบสิทธิ์ admin
		isAdmin, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{})
		if err != nil || !isAdmin {
			return errors.New("you don't have permission to access analytics data")
		}
	}

	return nil
}

// GetDailyAnalytics ดึงข้อมูลวิเคราะห์รายวัน
func (s *analyticsService) GetDailyAnalytics(businessID, userID uuid.UUID, startDate, endDate time.Time) ([]*models.AnalyticsDaily, error) {
	// ตรวจสอบสิทธิ์
	if err := s.checkAnalyticsPermission(businessID, userID); err != nil {
		return nil, err
	}

	// ดึงข้อมูลวิเคราะห์รายวัน
	return s.analyticsRepo.GetDailyAnalytics(businessID, startDate, endDate)
}

// GetSummaryAnalytics สรุปข้อมูลวิเคราะห์ตามช่วงเวลา
func (s *analyticsService) GetSummaryAnalytics(businessID, userID uuid.UUID, days int) (map[string]interface{}, error) {
	// ตรวจสอบสิทธิ์
	if err := s.checkAnalyticsPermission(businessID, userID); err != nil {
		return nil, err
	}

	// ตรวจสอบค่า days
	if days <= 0 {
		days = 30 // ค่าเริ่มต้น 30 วัน
	} else if days > 365 {
		days = 365 // จำกัดไม่ให้เกิน 1 ปี
	}

	// ดึงข้อมูลสรุป
	return s.analyticsRepo.GetSummaryAnalytics(businessID, days)
}

// TrackNewFollower บันทึกการติดตามใหม่
func (s *analyticsService) TrackNewFollower(businessID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementNewFollowers(businessID, today)
}

// TrackUnfollow บันทึกการเลิกติดตาม
func (s *analyticsService) TrackUnfollow(businessID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementUnfollows(businessID, today)
}

// TrackMessageReceived บันทึกการรับข้อความ
func (s *analyticsService) TrackMessageReceived(businessID uuid.UUID, count int) error {
	if count <= 0 {
		count = 1 // ถ้าไม่ระบุจำนวน ให้ใช้ 1 เป็นค่าเริ่มต้น
	}
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementMessagesReceived(businessID, today, count)
}

// TrackMessageSent บันทึกการส่งข้อความ
func (s *analyticsService) TrackMessageSent(businessID uuid.UUID, count int) error {
	if count <= 0 {
		count = 1 // ถ้าไม่ระบุจำนวน ให้ใช้ 1 เป็นค่าเริ่มต้น
	}
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementMessagesSent(businessID, today, count)
}

// TrackActiveUser บันทึกผู้ใช้ที่มีปฏิสัมพันธ์
func (s *analyticsService) TrackActiveUser(businessID, userID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementActiveUsers(businessID, today, userID)
}

// TrackBroadcastOpen บันทึกการเปิดการแจ้งเตือน
func (s *analyticsService) TrackBroadcastOpen(businessID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementBroadcastOpens(businessID, today)
}

// TrackBroadcastClick บันทึกการคลิกการแจ้งเตือน
func (s *analyticsService) TrackBroadcastClick(businessID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementBroadcastClicks(businessID, today)
}

// TrackRichMenuClick บันทึกการคลิกเมนูลัด
func (s *analyticsService) TrackRichMenuClick(businessID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)
	return s.analyticsRepo.IncrementRichMenuClicks(businessID, today)
}
