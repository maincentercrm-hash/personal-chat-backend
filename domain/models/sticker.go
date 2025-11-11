// domain/models/sticker.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// Sticker - สติกเกอร์เดี่ยวในชุดสติกเกอร์
type Sticker struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	StickerSetID uuid.UUID `json:"sticker_set_id" gorm:"type:uuid;not null"`
	Name         string    `json:"name,omitempty" gorm:"type:varchar(100)"`
	StickerURL   string    `json:"sticker_url" gorm:"type:text;not null"`
	ThumbnailURL string    `json:"thumbnail_url" gorm:"type:text;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	IsAnimated   bool      `json:"is_animated" gorm:"default:false"`
	SortOrder    int       `json:"sort_order" gorm:"default:0"`

	// Associations
	StickerSet    *StickerSet            `json:"sticker_set,omitempty" gorm:"foreignkey:StickerSetID"`
	FavoriteUsers []*UserFavoriteSticker `json:"favorite_users,omitempty" gorm:"foreignkey:StickerID"`
	RecentUsers   []*UserRecentSticker   `json:"recent_users,omitempty" gorm:"foreignkey:StickerID"`
}

// TableName - ระบุชื่อตารางใน database
func (Sticker) TableName() string {
	return "stickers"
}
