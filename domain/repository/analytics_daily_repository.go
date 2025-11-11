// domain/repository/analytics_daily_repository.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type AnalyticsDailyRepository interface {
	// GetDailyAnalytics ดึงข้อมูลวิเคราะห์รายวัน
	GetDailyAnalytics(businessID uuid.UUID, startDate, endDate time.Time) ([]*models.AnalyticsDaily, error)

	// GetSummaryAnalytics สรุปข้อมูลวิเคราะห์
	GetSummaryAnalytics(businessID uuid.UUID, days int) (map[string]interface{}, error)

	// IncrementNewFollowers เพิ่มจำนวนผู้ติดตามใหม่
	IncrementNewFollowers(businessID uuid.UUID, date time.Time) error

	// IncrementUnfollows เพิ่มจำนวนการเลิกติดตาม
	IncrementUnfollows(businessID uuid.UUID, date time.Time) error

	// IncrementMessagesReceived เพิ่มจำนวนข้อความที่ได้รับ
	IncrementMessagesReceived(businessID uuid.UUID, date time.Time, count int) error

	// IncrementMessagesSent เพิ่มจำนวนข้อความที่ส่ง
	IncrementMessagesSent(businessID uuid.UUID, date time.Time, count int) error

	// IncrementActiveUsers เพิ่มจำนวนผู้ใช้ที่มีปฏิสัมพันธ์
	IncrementActiveUsers(businessID uuid.UUID, date time.Time, userID uuid.UUID) error

	// IncrementBroadcastOpens เพิ่มจำนวนการเปิดการแจ้งเตือน
	IncrementBroadcastOpens(businessID uuid.UUID, date time.Time) error

	// IncrementBroadcastClicks เพิ่มจำนวนการคลิกการแจ้งเตือน
	IncrementBroadcastClicks(businessID uuid.UUID, date time.Time) error

	// IncrementRichMenuClicks เพิ่มจำนวนการคลิกเมนูลัด
	IncrementRichMenuClicks(businessID uuid.UUID, date time.Time) error
}
