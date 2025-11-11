// domain/models/user_event.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// UserEvent - เหตุการณ์ที่เกี่ยวข้องกับผู้ใช้
type UserEvent struct {
	ID         uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID     uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	BusinessID *uuid.UUID  `json:"business_id,omitempty" gorm:"type:uuid"`
	EventType  string      `json:"event_type" gorm:"type:varchar(50);not null"`
	EventData  types.JSONB `json:"event_data" gorm:"type:jsonb;not null"`
	OccurredAt time.Time   `json:"occurred_at" gorm:"type:timestamp with time zone;default:now()"`
	SessionID  *uuid.UUID  `json:"session_id,omitempty" gorm:"type:uuid"`

	// Associations
	User     *User            `json:"user,omitempty" gorm:"foreignkey:UserID"`
	Business *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
}

// TableName - ระบุชื่อตารางใน database
func (UserEvent) TableName() string {
	return "user_events"
}
