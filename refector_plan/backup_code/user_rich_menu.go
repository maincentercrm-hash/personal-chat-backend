// domain/models/user_rich_menu.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRichMenu - การกำหนดเมนูหลายฟังก์ชันให้กับผู้ใช้
type UserRichMenu struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	RichMenuID uuid.UUID `json:"rich_menu_id" gorm:"type:uuid;not null"`
	BusinessID uuid.UUID `json:"business_id" gorm:"type:uuid;not null"`
	AssignedAt time.Time `json:"assigned_at" gorm:"type:timestamp with time zone;default:now()"`

	// Associations
	User     *User            `json:"user,omitempty" gorm:"foreignkey:UserID"`
	RichMenu *RichMenu        `json:"rich_menu,omitempty" gorm:"foreignkey:RichMenuID"`
	Business *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
}

// TableName - ระบุชื่อตารางใน database
func (UserRichMenu) TableName() string {
	return "user_rich_menus"
}
