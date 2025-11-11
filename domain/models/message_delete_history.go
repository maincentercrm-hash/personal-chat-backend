// domain/models/message_delete_history.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// MessageDeleteHistory - ประวัติการลบข้อความ
type MessageDeleteHistory struct {
	ID                uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MessageID         uuid.UUID   `json:"message_id" gorm:"type:uuid;not null"`
	Content           string      `json:"content,omitempty" gorm:"type:text"`
	MediaURL          string      `json:"media_url,omitempty" gorm:"type:text"`
	MediaThumbnailURL string      `json:"media_thumbnail_url,omitempty" gorm:"type:text"`
	Metadata          types.JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`
	DeletedAt         time.Time   `json:"deleted_at" gorm:"type:timestamp with time zone;default:now()"`
	DeletedBy         uuid.UUID   `json:"deleted_by" gorm:"type:uuid;not null"`

	// Associations
	Message *Message `json:"message,omitempty" gorm:"foreignkey:MessageID"`
	Deleter *User    `json:"deleter,omitempty" gorm:"foreignkey:DeletedBy"`
}

// TableName - ระบุชื่อตารางใน database
func (MessageDeleteHistory) TableName() string {
	return "message_delete_history"
}
