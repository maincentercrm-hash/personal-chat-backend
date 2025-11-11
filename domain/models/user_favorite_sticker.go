// domain/models/user_favorite_sticker.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// UserFavoriteSticker - สติกเกอร์โปรดของผู้ใช้
type UserFavoriteSticker struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	StickerID uuid.UUID `json:"sticker_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`

	// Associations
	User    *User    `json:"user,omitempty" gorm:"foreignkey:UserID"`
	Sticker *Sticker `json:"sticker,omitempty" gorm:"foreignkey:StickerID"`
}

// TableName - ระบุชื่อตารางใน database
func (UserFavoriteSticker) TableName() string {
	return "user_favorite_stickers"
}
