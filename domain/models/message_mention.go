package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageMention represents a user mention in a message
type MessageMention struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MessageID       uuid.UUID `json:"message_id" gorm:"type:uuid;not null;index"`
	MentionedUserID uuid.UUID `json:"mentioned_user_id" gorm:"type:uuid;not null;index"`
	StartIndex      *int      `json:"start_index,omitempty" gorm:"type:integer"`
	Length          *int      `json:"length,omitempty" gorm:"type:integer"`
	CreatedAt       time.Time `json:"created_at" gorm:"default:now()"`

	// Relations
	Message       *Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	MentionedUser *User    `json:"mentioned_user,omitempty" gorm:"foreignKey:MentionedUserID"`
}

// TableName specifies the table name for GORM
func (MessageMention) TableName() string {
	return "message_mentions"
}
