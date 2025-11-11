// domain/models/rich_menu.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// RichMenu - เมนูหลายฟังก์ชันสำหรับบัญชีธุรกิจ
type RichMenu struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BusinessID  uuid.UUID  `json:"business_id" gorm:"type:uuid;not null"`
	Name        string     `json:"name" gorm:"type:varchar(100);not null"`
	Description string     `json:"description,omitempty" gorm:"type:text"`
	ImageURL    string     `json:"image_url" gorm:"type:text;not null"`
	SizeW       int        `json:"size_w" gorm:"type:integer;not null"`
	SizeH       int        `json:"size_h" gorm:"type:integer;not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"type:timestamp with time zone;default:now()"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`

	// Associations
	Business  *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
	Creator   *User            `json:"creator,omitempty" gorm:"foreignkey:CreatedBy"`
	Areas     []*RichMenuArea  `json:"areas,omitempty" gorm:"foreignkey:RichMenuID"`
	UserMenus []*UserRichMenu  `json:"user_menus,omitempty" gorm:"foreignkey:RichMenuID"`
}

// TableName - ระบุชื่อตารางใน database
func (RichMenu) TableName() string {
	return "rich_menus"
}
