// domain/models/user_recent_sticker.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRecentSticker - สติกเกอร์ที่ใช้ล่าสุดของผู้ใช้
type UserRecentSticker struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	StickerID uuid.UUID `json:"sticker_id" gorm:"type:uuid;not null"`
	UsedAt    time.Time `json:"used_at" gorm:"type:timestamp with time zone;default:now()"`

	// Associations
	User    *User    `json:"user,omitempty" gorm:"foreignkey:UserID"`
	Sticker *Sticker `json:"sticker,omitempty" gorm:"foreignkey:StickerID"`
}

// TableName - ระบุชื่อตารางใน database
func (UserRecentSticker) TableName() string {
	return "user_recent_stickers"
}
