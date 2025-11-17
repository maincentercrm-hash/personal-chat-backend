// domain/models/rich_menu_area.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// RichMenuArea - พื้นที่คลิกในเมนูหลายฟังก์ชัน
type RichMenuArea struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	RichMenuID  uuid.UUID `json:"rich_menu_id" gorm:"type:uuid;not null"`
	X           int       `json:"x" gorm:"type:integer;not null"`
	Y           int       `json:"y" gorm:"type:integer;not null"`
	Width       int       `json:"width" gorm:"type:integer;not null"`
	Height      int       `json:"height" gorm:"type:integer;not null"`
	ActionType  string    `json:"action_type" gorm:"type:varchar(50);not null"`
	ActionLabel string    `json:"action_label,omitempty" gorm:"type:varchar(100)"`
	ActionData  string    `json:"action_data" gorm:"type:text;not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`

	// Associations
	RichMenu *RichMenu `json:"rich_menu,omitempty" gorm:"foreignkey:RichMenuID"`
}

// TableName - ระบุชื่อตารางใน database
func (RichMenuArea) TableName() string {
	return "rich_menu_areas"
}
