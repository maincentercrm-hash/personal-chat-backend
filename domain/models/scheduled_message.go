// domain/models/scheduled_message.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ScheduledMessage - ข้อความที่กำหนดเวลาส่ง
type ScheduledMessage struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ConversationID uuid.UUID   `json:"conversation_id" gorm:"type:uuid;not null"`
	SenderID       uuid.UUID   `json:"sender_id" gorm:"type:uuid;not null"`
	MessageType    string      `json:"message_type" gorm:"type:varchar(20);not null"` // text, image, file, sticker
	Content        string      `json:"content,omitempty" gorm:"type:text"`
	MediaURL       string      `json:"media_url,omitempty" gorm:"type:text"`
	Metadata       types.JSONB `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	ScheduledAt time.Time  `json:"scheduled_at" gorm:"type:timestamp with time zone;not null"`
	Status      string     `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, sent, cancelled, failed
	SentAt      *time.Time `json:"sent_at,omitempty" gorm:"type:timestamp with time zone"`
	MessageID   *uuid.UUID `json:"message_id,omitempty" gorm:"type:uuid"` // ID ของข้อความที่ส่งแล้ว
	ErrorReason string     `json:"error_reason,omitempty" gorm:"type:text"` // เก็บข้อผิดพลาดถ้าส่งไม่สำเร็จ

	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp with time zone;default:now()"`

	// Associations
	Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignkey:ConversationID"`
	Sender       *User         `json:"sender,omitempty" gorm:"foreignkey:SenderID"`
	Message      *Message      `json:"message,omitempty" gorm:"foreignkey:MessageID"`
}

// TableName - ระบุชื่อตารางใน database
func (ScheduledMessage) TableName() string {
	return "scheduled_messages"
}
