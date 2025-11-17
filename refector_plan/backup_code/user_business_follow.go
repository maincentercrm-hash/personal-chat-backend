// domain/models/user_business_follow.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// UserBusinessFollow - การติดตามบัญชีธุรกิจโดยผู้ใช้
type UserBusinessFollow struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	BusinessID uuid.UUID `json:"business_id" gorm:"type:uuid;not null"`
	FollowedAt time.Time `json:"followed_at" gorm:"type:timestamp with time zone;default:now()"`
	Source     string    `json:"source,omitempty" gorm:"type:varchar(50)"`

	// Associations
	User     *User            `json:"user,omitempty" gorm:"foreignkey:UserID"`
	Business *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
}

// TableName - ระบุชื่อตารางใน database
func (UserBusinessFollow) TableName() string {
	return "user_business_follows"
}
