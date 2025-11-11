// domain/service/analytics_service.go
package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type AnalyticsService interface {
	// GetDailyAnalytics ดึงข้อมูลวิเคราะห์รายวัน
	GetDailyAnalytics(businessID, userID uuid.UUID, startDate, endDate time.Time) ([]*models.AnalyticsDaily, error)

	// GetSummaryAnalytics สรุปข้อมูลวิเคราะห์ตามช่วงเวลา
	GetSummaryAnalytics(businessID, userID uuid.UUID, days int) (map[string]interface{}, error)

	// TrackNewFollower บันทึกการติดตามใหม่
	TrackNewFollower(businessID uuid.UUID) error

	// TrackUnfollow บันทึกการเลิกติดตาม
	TrackUnfollow(businessID uuid.UUID) error

	// TrackMessageReceived บันทึกการรับข้อความ
	TrackMessageReceived(businessID uuid.UUID, count int) error

	// TrackMessageSent บันทึกการส่งข้อความ
	TrackMessageSent(businessID uuid.UUID, count int) error

	// TrackActiveUser บันทึกผู้ใช้ที่มีปฏิสัมพันธ์
	TrackActiveUser(businessID, userID uuid.UUID) error

	// TrackBroadcastOpen บันทึกการเปิดการแจ้งเตือน
	TrackBroadcastOpen(businessID uuid.UUID) error

	// TrackBroadcastClick บันทึกการคลิกการแจ้งเตือน
	TrackBroadcastClick(businessID uuid.UUID) error

	// TrackRichMenuClick บันทึกการคลิกเมนูลัด
	TrackRichMenuClick(businessID uuid.UUID) error
}
