// domain/models/message_read.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageRead - บันทึกการอ่านข้อความ
type MessageRead struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MessageID uuid.UUID `json:"message_id" gorm:"type:uuid;not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	ReadAt    time.Time `json:"read_at" gorm:"type:timestamp with time zone;default:now()"`

	// Associations
	Message *Message `json:"message,omitempty" gorm:"foreignkey:MessageID"`
	User    *User    `json:"user,omitempty" gorm:"foreignkey:UserID"`
}

// TableName - ระบุชื่อตารางใน database
func (MessageRead) TableName() string {
	return "message_reads"
}
