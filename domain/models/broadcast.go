// domain/models/broadcast.go
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// Broadcast - ข้อความ broadcast ที่ส่งไปยังผู้ติดตาม
type Broadcast struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BusinessID   uuid.UUID   `json:"business_id" gorm:"type:uuid;not null;index"`
	Title        string      `json:"title" gorm:"type:varchar(100);not null"`
	MessageType  string      `json:"message_type" gorm:"type:varchar(20);not null"` // text, image, carousel
	Content      string      `json:"content,omitempty" gorm:"type:text"`
	MediaURL     string      `json:"media_url,omitempty" gorm:"type:text"`
	ScheduledAt  *time.Time  `json:"scheduled_at,omitempty" gorm:"type:timestamp with time zone"`
	SentAt       *time.Time  `json:"sent_at,omitempty" gorm:"type:timestamp with time zone"`
	CreatedAt    time.Time   `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	CreatedBy    *uuid.UUID  `json:"created_by,omitempty" gorm:"type:uuid"`
	Status       string      `json:"status" gorm:"type:varchar(20);default:'draft'"`    // draft, scheduled, sending, completed, failed
	TargetType   string      `json:"target_type" gorm:"type:varchar(20);default:'all'"` // all, segment
	TargetData   types.JSONB `json:"target_data,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`
	Metrics      types.JSONB `json:"metrics,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`
	ErrorMessage string      `json:"error_message,omitempty" gorm:"type:text"`

	// NEW: เพิ่มฟิลด์สำหรับข้อมูล bubble
	BubbleType string      `json:"bubble_type,omitempty" gorm:"type:varchar(30)"`
	BubbleData types.JSONB `json:"bubble_data,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Associations
	Business   *BusinessAccount     `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
	Creator    *User                `json:"creator,omitempty" gorm:"foreignkey:CreatedBy"`
	Deliveries []*BroadcastDelivery `json:"deliveries,omitempty" gorm:"foreignkey:BroadcastID"`
}

// TableName - ระบุชื่อตารางใน database
func (Broadcast) TableName() string {
	return "broadcasts"
}
