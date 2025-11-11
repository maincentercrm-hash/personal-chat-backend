// infrastructure/persistence/postgres/analytics_daily_repository.go
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type analyticsDailyRepository struct {
	db *gorm.DB
}

// NewAnalyticsDailyRepository สร้าง instance ใหม่ของ AnalyticsDailyRepository
func NewAnalyticsDailyRepository(db *gorm.DB) repository.AnalyticsDailyRepository {
	return &analyticsDailyRepository{db: db}
}

// GetDailyAnalytics ดึงข้อมูลวิเคราะห์รายวัน
func (r *analyticsDailyRepository) GetDailyAnalytics(businessID uuid.UUID, startDate, endDate time.Time) ([]*models.AnalyticsDaily, error) {
	var analytics []*models.AnalyticsDaily

	err := r.db.Where("business_id = ? AND date >= ? AND date <= ?",
		businessID, startDate, endDate).
		Order("date ASC").
		Find(&analytics).Error

	return analytics, err
}

// GetSummaryAnalytics สรุปข้อมูลวิเคราะห์
func (r *analyticsDailyRepository) GetSummaryAnalytics(businessID uuid.UUID, days int) (map[string]interface{}, error) {
	var analytics []*models.AnalyticsDaily

	// คำนวณวันที่เริ่มต้น
	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	err := r.db.Where("business_id = ? AND date >= ?",
		businessID, startDate).
		Order("date ASC").
		Find(&analytics).Error

	if err != nil {
		return nil, err
	}

	// รวมข้อมูลวิเคราะห์
	var totalNewFollowers, totalUnfollows, totalMessagesReceived, totalMessagesSent int
	var totalActiveUsers, totalBroadcastOpens, totalBroadcastClicks, totalRichMenuClicks int

	for _, analytic := range analytics {
		totalNewFollowers += analytic.NewFollowers
		totalUnfollows += analytic.Unfollows
		totalMessagesReceived += analytic.MessagesReceived
		totalMessagesSent += analytic.MessagesSent
		totalActiveUsers += analytic.ActiveUsers
		totalBroadcastOpens += analytic.BroadcastOpens
		totalBroadcastClicks += analytic.BroadcastClicks
		totalRichMenuClicks += analytic.RichMenuClicks
	}

	// คำนวณอัตราการเติบโตของผู้ติดตาม
	followerGrowth := totalNewFollowers - totalUnfollows

	// เตรียมข้อมูลรายวัน
	dailyData := make([]map[string]interface{}, len(analytics))
	for i, analytic := range analytics {
		dailyData[i] = map[string]interface{}{
			"date":              analytic.Date.Format("2006-01-02"),
			"new_followers":     analytic.NewFollowers,
			"unfollows":         analytic.Unfollows,
			"messages_received": analytic.MessagesReceived,
			"messages_sent":     analytic.MessagesSent,
			"active_users":      analytic.ActiveUsers,
		}
	}

	// สรุปข้อมูล
	summary := map[string]interface{}{
		"period_days":             days,
		"total_new_followers":     totalNewFollowers,
		"total_unfollows":         totalUnfollows,
		"follower_growth":         followerGrowth,
		"total_messages_received": totalMessagesReceived,
		"total_messages_sent":     totalMessagesSent,
		"total_active_users":      totalActiveUsers,
		"total_broadcast_opens":   totalBroadcastOpens,
		"total_broadcast_clicks":  totalBroadcastClicks,
		"total_rich_menu_clicks":  totalRichMenuClicks,
		"daily_data":              dailyData,
	}

	return summary, nil
}

// เพิ่มหรืออัปเดตข้อมูลวิเคราะห์รายวัน
func (r *analyticsDailyRepository) getOrCreateDailyAnalytics(businessID uuid.UUID, date time.Time) (*models.AnalyticsDaily, error) {
	var analytics models.AnalyticsDaily

	// ตรวจสอบว่ามีข้อมูลของวันนี้หรือไม่
	err := r.db.Where("business_id = ? AND date = ?", businessID, date.Format("2006-01-02")).
		First(&analytics).Error

	if err != nil {
		// ถ้าไม่พบ ให้สร้างใหม่
		if err == gorm.ErrRecordNotFound {
			analytics = models.AnalyticsDaily{
				ID:               uuid.New(),
				BusinessID:       businessID,
				Date:             date,
				NewFollowers:     0,
				Unfollows:        0,
				MessagesReceived: 0,
				MessagesSent:     0,
				ActiveUsers:      0,
				BroadcastOpens:   0,
				BroadcastClicks:  0,
				RichMenuClicks:   0,
			}
			err = r.db.Create(&analytics).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &analytics, nil
}

// IncrementNewFollowers เพิ่มจำนวนผู้ติดตามใหม่
func (r *analyticsDailyRepository) IncrementNewFollowers(businessID uuid.UUID, date time.Time) error {
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนผู้ติดตามใหม่
	analytics.NewFollowers++
	return r.db.Save(analytics).Error
}

// IncrementUnfollows เพิ่มจำนวนการเลิกติดตาม
func (r *analyticsDailyRepository) IncrementUnfollows(businessID uuid.UUID, date time.Time) error {
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนการเลิกติดตาม
	analytics.Unfollows++
	return r.db.Save(analytics).Error
}

// IncrementMessagesReceived เพิ่มจำนวนข้อความที่ได้รับ
func (r *analyticsDailyRepository) IncrementMessagesReceived(businessID uuid.UUID, date time.Time, count int) error {
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนข้อความที่ได้รับ
	analytics.MessagesReceived += count
	return r.db.Save(analytics).Error
}

// IncrementMessagesSent เพิ่มจำนวนข้อความที่ส่ง
func (r *analyticsDailyRepository) IncrementMessagesSent(businessID uuid.UUID, date time.Time, count int) error {
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนข้อความที่ส่ง
	analytics.MessagesSent += count
	return r.db.Save(analytics).Error
}

// IncrementActiveUsers เพิ่มจำนวนผู้ใช้ที่มีปฏิสัมพันธ์
func (r *analyticsDailyRepository) IncrementActiveUsers(businessID uuid.UUID, date time.Time, userID uuid.UUID) error {
	// ใช้ Redis หรือตารางเพิ่มเติมเพื่อไม่ให้นับซ้ำในวันเดียวกัน
	// สำหรับตัวอย่างนี้ เราจะเพิ่มค่าทุกครั้ง
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนผู้ใช้ที่มีปฏิสัมพันธ์
	analytics.ActiveUsers++
	return r.db.Save(analytics).Error
}

// IncrementBroadcastOpens เพิ่มจำนวนการเปิดการแจ้งเตือน
func (r *analyticsDailyRepository) IncrementBroadcastOpens(businessID uuid.UUID, date time.Time) error {
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนการเปิดการแจ้งเตือน
	analytics.BroadcastOpens++
	return r.db.Save(analytics).Error
}

// IncrementBroadcastClicks เพิ่มจำนวนการคลิกการแจ้งเตือน
func (r *analyticsDailyRepository) IncrementBroadcastClicks(businessID uuid.UUID, date time.Time) error {
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนการคลิกการแจ้งเตือน
	analytics.BroadcastClicks++
	return r.db.Save(analytics).Error
}

// IncrementRichMenuClicks เพิ่มจำนวนการคลิกเมนูลัด
func (r *analyticsDailyRepository) IncrementRichMenuClicks(businessID uuid.UUID, date time.Time) error {
	analytics, err := r.getOrCreateDailyAnalytics(businessID, date)
	if err != nil {
		return err
	}

	// เพิ่มจำนวนการคลิกเมนูลัด
	analytics.RichMenuClicks++
	return r.db.Save(analytics).Error
}
