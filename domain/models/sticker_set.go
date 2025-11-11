// domain/models/sticker_set.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// StickerSet - ชุดสติกเกอร์
type StickerSet struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name          string    `json:"name" gorm:"type:varchar(100);not null"`
	Description   string    `json:"description,omitempty" gorm:"type:text"`
	Author        string    `json:"author,omitempty" gorm:"type:varchar(100)"`
	CoverImageURL string    `json:"cover_image_url,omitempty" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	IsOfficial    bool      `json:"is_official" gorm:"default:false"`
	IsDefault     bool      `json:"is_default" gorm:"default:true"`
	SortOrder     int       `json:"sort_order" gorm:"default:0"`

	// Associations
	Stickers []*Sticker        `json:"stickers,omitempty" gorm:"foreignkey:StickerSetID"`
	UserSets []*UserStickerSet `json:"user_sets,omitempty" gorm:"foreignkey:StickerSetID"`
}

// TableName - ระบุชื่อตารางใน database
func (StickerSet) TableName() string {
	return "sticker_sets"
}
