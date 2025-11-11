// domain/models/business_welcome_message.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// BusinessWelcomeMessage - ข้อความต้อนรับของบัญชีธุรกิจ
type BusinessWelcomeMessage struct {
	ID            uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BusinessID    uuid.UUID   `json:"business_id" gorm:"type:uuid;not null;index"`
	IsActive      bool        `json:"is_active" gorm:"default:true"`
	MessageType   string      `json:"message_type" gorm:"type:varchar(50);default:'text'"` // text, image, card, carousel, flex
	Title         string      `json:"title,omitempty" gorm:"type:varchar(100)"`
	Content       string      `json:"content,omitempty" gorm:"type:text"`
	ImageURL      string      `json:"image_url,omitempty" gorm:"type:text"`
	ThumbnailURL  string      `json:"thumbnail_url,omitempty" gorm:"type:text"`
	ActionButtons types.JSONB `json:"action_buttons,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// สำหรับข้อความประเภท carousel หรือ flex
	Components types.JSONB `json:"components,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// การตั้งค่าเงื่อนไขการส่ง
	TriggerType   string      `json:"trigger_type" gorm:"type:varchar(50);default:'follow'"` // follow, inactive, schedule
	TriggerParams types.JSONB `json:"trigger_params,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// ลำดับการแสดงผล (ถ้ามีหลายข้อความ)
	SortOrder int `json:"sort_order" gorm:"default:0"`

	// ข้อมูลเพิ่มเติมและสถิติ
	SentCount  int `json:"sent_count" gorm:"default:0"`
	ClickCount int `json:"click_count" gorm:"default:0"`
	ReplyCount int `json:"reply_count" gorm:"default:0"`

	// ข้อมูลเวลา
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamp with time zone;default:now()"`
	CreatedByID uuid.UUID `json:"created_by_id" gorm:"type:uuid;column:created_by"`
	UpdatedByID uuid.UUID `json:"updated_by_id" gorm:"type:uuid;column:updated_by"`

	// ความสัมพันธ์
	Business      *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
	CreatedByUser *User            `json:"created_by_user,omitempty" gorm:"foreignkey:CreatedByID"`
	UpdatedByUser *User            `json:"updated_by_user,omitempty" gorm:"foreignkey:UpdatedByID"`
}

// TableName - ระบุชื่อตารางใน database
func (BusinessWelcomeMessage) TableName() string {
	return "business_welcome_messages"
}
