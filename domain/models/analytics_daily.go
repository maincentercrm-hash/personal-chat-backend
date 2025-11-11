// domain/models/analytics_daily.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsDaily - ข้อมูลวิเคราะห์รายวันของธุรกิจ
type AnalyticsDaily struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BusinessID       uuid.UUID `json:"business_id" gorm:"type:uuid;not null"`
	Date             time.Time `json:"date" gorm:"type:date;not null"`
	NewFollowers     int       `json:"new_followers" gorm:"default:0"`
	Unfollows        int       `json:"unfollows" gorm:"default:0"`
	MessagesReceived int       `json:"messages_received" gorm:"default:0"`
	MessagesSent     int       `json:"messages_sent" gorm:"default:0"`
	ActiveUsers      int       `json:"active_users" gorm:"default:0"`
	BroadcastOpens   int       `json:"broadcast_opens" gorm:"default:0"`
	BroadcastClicks  int       `json:"broadcast_clicks" gorm:"default:0"`
	RichMenuClicks   int       `json:"rich_menu_clicks" gorm:"default:0"`

	// Associations
	Business *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
}

// TableName - ระบุชื่อตารางใน database
func (AnalyticsDaily) TableName() string {
	return "analytics_daily"
}
