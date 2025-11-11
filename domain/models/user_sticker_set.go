// domain/models/user_sticker_set.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// UserStickerSet - ชุดสติกเกอร์ของผู้ใช้
type UserStickerSet struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	StickerSetID uuid.UUID `json:"sticker_set_id" gorm:"type:uuid;not null"`
	PurchasedAt  time.Time `json:"purchased_at" gorm:"type:timestamp with time zone;default:now()"`
	IsFavorite   bool      `json:"is_favorite" gorm:"default:false"`

	// Associations
	User       *User       `json:"user,omitempty" gorm:"foreignkey:UserID"`
	StickerSet *StickerSet `json:"sticker_set,omitempty" gorm:"foreignkey:StickerSetID"`
}

// TableName - ระบุชื่อตารางใน database
func (UserStickerSet) TableName() string {
	return "user_sticker_sets"
}
