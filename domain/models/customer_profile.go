// domain/models/customer_profile.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// CustomerProfile - โปรไฟล์ลูกค้าในมุมมองของแต่ละธุรกิจ
type CustomerProfile struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BusinessID uuid.UUID `json:"business_id" gorm:"type:uuid;not null;index:idx_customer_business_user,unique,priority:1"` // เพิ่ม unique index
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index:idx_customer_business_user,unique,priority:2"`     // เพิ่ม unique index

	// ชื่อเล่น/ชื่องานที่ธุรกิจตั้งให้
	Nickname string `json:"nickname,omitempty" gorm:"type:varchar(100)"`

	// ข้อมูลเพิ่มเติมที่ธุรกิจจดไว้
	Notes        string `json:"notes,omitempty" gorm:"type:text"`
	CustomerType string `json:"customer_type,omitempty" gorm:"type:varchar(50)"` // VIP, Regular, New, etc.

	// 🆕 เพิ่มฟิลด์สำหรับ Analytics และ CRM
	LastContactAt *time.Time `json:"last_contact_at,omitempty" gorm:"type:timestamp with time zone"`
	Status        string     `json:"status" gorm:"type:varchar(20);default:'active'"` // active, inactive, blocked, archived

	// 🆕 สำหรับ Analytics (ไม่บังคับ - ใช้ computed values ก็ได้)
	// InteractionCount int `json:"interaction_count" gorm:"type:integer;default:0"` // จำนวนครั้งที่ติดต่อ
	// LastPurchaseAt  *time.Time `json:"last_purchase_at,omitempty" gorm:"type:timestamp with time zone"` // ถ้ามี e-commerce

	// Metadata
	Metadata types.JSONB `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"type:timestamp with time zone;default:now()"`
	CreatedByID *uuid.UUID `json:"created_by_id,omitempty" gorm:"type:uuid"`
	UpdatedByID *uuid.UUID `json:"updated_by_id,omitempty" gorm:"type:uuid"`

	// Associations
	Business  *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
	User      *User            `json:"user,omitempty" gorm:"foreignkey:UserID"`
	CreatedBy *User            `json:"created_by,omitempty" gorm:"foreignkey:CreatedByID"`
	UpdatedBy *User            `json:"updated_by,omitempty" gorm:"foreignkey:UpdatedByID"`
	Tags      []*UserTag       `json:"tags,omitempty" gorm:"-"` // ปิดการสร้าง foreign key

}

// TableName - ระบุชื่อตารางใน database
func (CustomerProfile) TableName() string {
	return "customer_profiles"
}
