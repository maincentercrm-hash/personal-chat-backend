// domain/models/tag.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// Tag - แท็กสำหรับจัดกลุ่มผู้ใช้ในบัญชีธุรกิจ
type Tag struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BusinessID  uuid.UUID  `json:"business_id" gorm:"type:uuid;not null"`
	Name        string     `json:"name" gorm:"type:varchar(50);not null"`
	Color       string     `json:"color,omitempty" gorm:"type:varchar(20)"`
	CreatedAt   time.Time  `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	CreatedByID *uuid.UUID `json:"created_by_id,omitempty" gorm:"type:uuid"`

	// Associations
	Business *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
	Creator  *User            `json:"creator,omitempty" gorm:"foreignkey:CreatedByID"`
	UserTags []*UserTag       `json:"user_tags,omitempty" gorm:"foreignkey:TagID"`
}

// TableName - ระบุชื่อตารางใน database
func (Tag) TableName() string {
	return "tags"
}
