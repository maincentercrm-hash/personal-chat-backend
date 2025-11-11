// domain/models/refresh_token.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken - โทเคนสำหรับรีเฟรชการเข้าถึง
type RefreshToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Token     string    `json:"token" gorm:"type:text;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"type:timestamp with time zone;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	Revoked   bool      `json:"revoked" gorm:"default:false"`

	// Associations
	User *User `json:"user,omitempty" gorm:"foreignkey:UserID"`
}

// TableName - ระบุชื่อตารางใน database
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
