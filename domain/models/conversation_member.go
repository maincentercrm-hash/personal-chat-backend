// domain/models/conversation_member.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ConversationMember - สมาชิกในการสนทนา
type ConversationMember struct {
	ID                   uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ConversationID       uuid.UUID   `json:"conversation_id" gorm:"type:uuid;not null"`
	UserID               uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	IsAdmin              bool        `json:"is_admin" gorm:"default:false"`
	JoinedAt             time.Time   `json:"joined_at" gorm:"type:timestamp with time zone;default:now()"`
	LastReadAt           *time.Time  `json:"last_read_at,omitempty" gorm:"type:timestamp with time zone"`
	IsMuted              bool        `json:"is_muted" gorm:"default:false"`
	IsPinned             bool        `json:"is_pinned" gorm:"default:false"`
	Nickname             string      `json:"nickname,omitempty" gorm:"type:varchar(100)"`
	NotificationSettings types.JSONB `json:"notification_settings,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Associations
	Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignkey:ConversationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignkey:UserID"`
}

// TableName - ระบุชื่อตารางใน database
func (ConversationMember) TableName() string {
	return "conversation_members"
}
