// domain/models/broadcast_delivery.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// BroadcastDelivery - การส่ง broadcast ไปยังผู้ใช้
type BroadcastDelivery struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BroadcastID  uuid.UUID  `json:"broadcast_id" gorm:"type:uuid;not null"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty" gorm:"type:timestamp with time zone"`
	OpenedAt     *time.Time `json:"opened_at,omitempty" gorm:"type:timestamp with time zone"`
	ClickedAt    *time.Time `json:"clicked_at,omitempty" gorm:"type:timestamp with time zone"`
	Status       string     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	ErrorMessage string     `json:"error_message,omitempty" gorm:"type:text"`

	// Associations
	Broadcast *Broadcast `json:"broadcast,omitempty" gorm:"foreignkey:BroadcastID"`
	User      *User      `json:"user,omitempty" gorm:"foreignkey:UserID"`
}

// TableName - ระบุชื่อตารางใน database
func (BroadcastDelivery) TableName() string {
	return "broadcast_deliveries"
}
