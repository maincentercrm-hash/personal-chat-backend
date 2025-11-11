// domain/models/message_edit_history.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// MessageEditHistory - ประวัติการแก้ไขข้อความ
type MessageEditHistory struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MessageID       uuid.UUID   `json:"message_id" gorm:"type:uuid;not null"`
	PreviousContent string      `json:"previous_content" gorm:"type:text;not null"`
	EditedAt        time.Time   `json:"edited_at" gorm:"type:timestamp with time zone;default:now()"`
	EditedBy        uuid.UUID   `json:"edited_by" gorm:"type:uuid;not null"`
	Metadata        types.JSONB `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Associations
	Message *Message `json:"message,omitempty" gorm:"foreignkey:MessageID"`
	Editor  *User    `json:"editor,omitempty" gorm:"foreignkey:EditedBy"`
}

// TableName - ระบุชื่อตารางใน database
func (MessageEditHistory) TableName() string {
	return "message_edit_history"
}
