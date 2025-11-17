package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Request DTOs ============

// AnalyticsDailyRequest สำหรับการดึงข้อมูลวิเคราะห์รายวัน
type AnalyticsDailyRequest struct {
	BusinessID uuid.UUID `json:"business_id" validate:"required"`
	StartDate  string    `json:"start_date" validate:"omitempty,datetime=2006-01-02"`
	EndDate    string    `json:"end_date" validate:"omitempty,datetime=2006-01-02"`
}

// AnalyticsSummaryRequest สำหรับการดึงข้อมูลสรุปวิเคราะห์
type AnalyticsSummaryRequest struct {
	BusinessID uuid.UUID `json:"business_id" validate:"required"`
	Days       int       `json:"days" validate:"omitempty,min=1,max=365"`
}

// AnalyticsTrackEventRequest สำหรับการบันทึกเหตุการณ์
type AnalyticsTrackEventRequest struct {
	EventType string    `json:"event_type" validate:"required,oneof=new_follower unfollow message_received message_sent active_user broadcast_open broadcast_click rich_menu_click"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	Count     int       `json:"count,omitempty" validate:"omitempty,min=1"`
}

// ============ Response DTOs ============

// AnalyticsDailyItem รายการข้อมูลวิเคราะห์รายวัน
type AnalyticsDailyItem struct {
	Date             time.Time `json:"date"`
	NewFollowers     int       `json:"new_followers"`
	Unfollows        int       `json:"unfollows"`
	MessagesReceived int       `json:"messages_received"`
	MessagesSent     int       `json:"messages_sent"`
	ActiveUsers      int       `json:"active_users"`
	BroadcastOpens   int       `json:"broadcast_opens"`
	BroadcastClicks  int       `json:"broadcast_clicks"`
	RichMenuClicks   int       `json:"rich_menu_clicks"`
}

// AnalyticsDailyResponse ข้อมูลตอบกลับสำหรับการดึงข้อมูลวิเคราะห์รายวัน
type AnalyticsDailyResponse struct {
	Success bool               `json:"success"`
	Data    AnalyticsDailyData `json:"data"`
}

// AnalyticsDailyData ข้อมูลวิเคราะห์รายวัน
type AnalyticsDailyData struct {
	Analytics []AnalyticsDailyItem `json:"analytics"`
	StartDate string               `json:"start_date"`
	EndDate   string               `json:"end_date"`
}

// AnalyticsSummaryResponse ข้อมูลตอบกลับสำหรับการดึงข้อมูลสรุปวิเคราะห์
type AnalyticsSummaryResponse struct {
	Success bool                 `json:"success"`
	Data    AnalyticsSummaryData `json:"data"`
}

// AnalyticsSummaryData ข้อมูลสรุปวิเคราะห์
type AnalyticsSummaryData struct {
	TotalFollowers         int     `json:"total_followers"`
	NewFollowers           int     `json:"new_followers"`
	Unfollows              int     `json:"unfollows"`
	NetFollowerGrowth      int     `json:"net_follower_growth"`
	FollowerGrowthRate     float64 `json:"follower_growth_rate"`
	MessagesReceived       int     `json:"messages_received"`
	MessagesSent           int     `json:"messages_sent"`
	TotalMessages          int     `json:"total_messages"`
	ActiveUsers            int     `json:"active_users"`
	ActiveUserRate         float64 `json:"active_user_rate"`
	BroadcastOpens         int     `json:"broadcast_opens"`
	BroadcastClicks        int     `json:"broadcast_clicks"`
	BroadcastOpenRate      float64 `json:"broadcast_open_rate"`
	BroadcastClickRate     float64 `json:"broadcast_click_rate"`
	RichMenuClicks         int     `json:"rich_menu_clicks"`
	AverageMessagesPerUser float64 `json:"average_messages_per_user"`
	Days                   int     `json:"days"`
}

// AnalyticsTrackEventResponse ข้อมูลตอบกลับสำหรับการบันทึกเหตุการณ์
type AnalyticsTrackEventResponse struct {
	GenericResponse
}
