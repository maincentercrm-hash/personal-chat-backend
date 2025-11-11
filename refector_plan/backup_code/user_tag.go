// domain/models/user_tag.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// UserTag - แท็กที่กำหนดให้กับผู้ใช้โดยธุรกิจ
type UserTag struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_tag_business"`     // เพิ่ม uniqueIndex
	TagID      uuid.UUID  `json:"tag_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_tag_business"`      // เพิ่ม uniqueIndex
	BusinessID uuid.UUID  `json:"business_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_tag_business"` // เพิ่ม uniqueIndex
	AddedAt    time.Time  `json:"added_at" gorm:"type:timestamp with time zone;default:now()"`
	AddedByID  *uuid.UUID `json:"added_by_id,omitempty" gorm:"type:uuid"`

	// Associations
	User     *User            `json:"user,omitempty" gorm:"foreignkey:UserID"`
	Tag      *Tag             `json:"tag,omitempty" gorm:"foreignkey:TagID"`
	Business *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
	AddedBy  *User            `json:"added_by,omitempty" gorm:"foreignkey:AddedByID"`
}

// TableName - ระบุชื่อตารางใน database
func (UserTag) TableName() string {
	return "user_tags"
}
