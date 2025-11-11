// domain/models/business_admin.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// BusinessAdmin - ผู้ดูแลระบบของบัญชีธุรกิจ
type BusinessAdmin struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BusinessID uuid.UUID  `json:"business_id" gorm:"type:uuid;not null"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Role       string     `json:"role" gorm:"type:varchar(20);not null"`
	AddedAt    time.Time  `json:"added_at" gorm:"type:timestamp with time zone;default:now()"`
	AddedBy    *uuid.UUID `json:"added_by,omitempty" gorm:"type:uuid"`

	// Associations
	Business *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
	User     *User            `json:"user,omitempty" gorm:"foreignkey:UserID"`
	Adder    *User            `json:"adder,omitempty" gorm:"foreignkey:AddedBy"`
}

// TableName - ระบุชื่อตารางใน database
func (BusinessAdmin) TableName() string {
	return "business_admins"
}
