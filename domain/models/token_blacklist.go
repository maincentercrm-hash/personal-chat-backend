// domain/models/token_blacklist.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// TokenBlacklist - รายการโทเคนที่ถูกปฏิเสธ
type TokenBlacklist struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Token     string    `json:"token" gorm:"type:text;not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	ExpiredAt time.Time `json:"expired_at" gorm:"type:timestamp with time zone;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`

	// Associations
	User *User `json:"user,omitempty" gorm:"foreignkey:UserID"`
}

// TableName - ระบุชื่อตารางใน database
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}
