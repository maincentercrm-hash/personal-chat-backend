// domain/models/business_account.go

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// BusinessAccount - บัญชีธุรกิจสำหรับ chat
type BusinessAccount struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name            string      `json:"name" gorm:"type:varchar(100);not null"`
	Description     string      `json:"description,omitempty" gorm:"type:text"`
	ProfileImageURL string      `json:"profile_image_url,omitempty" gorm:"type:text"`
	CoverImageURL   string      `json:"cover_image_url,omitempty" gorm:"type:text"`
	CreatedAt       time.Time   `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	OwnerID         *uuid.UUID  `json:"owner_id,omitempty" gorm:"type:uuid"`
	Status          string      `json:"status" gorm:"type:varchar(20);default:'active'"`
	Settings        types.JSONB `json:"settings,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`
	Username        string      `json:"username" gorm:"type:varchar(50);not null;unique"`

	// ฟิลด์เดิม WelcomeMessage ถูกลบออก และจะย้ายไปอยู่ในตาราง business_welcome_messages

	// Associations
	Owner            *User                     `json:"owner,omitempty" gorm:"foreignkey:OwnerID"`
	Admins           []*BusinessAdmin          `json:"admins,omitempty" gorm:"foreignkey:BusinessID"`
	Followers        []*UserBusinessFollow     `json:"followers,omitempty" gorm:"foreignkey:BusinessID"`
	Broadcasts       []*Broadcast              `json:"broadcasts,omitempty" gorm:"foreignkey:BusinessID"`
	Conversations    []*Conversation           `json:"conversations,omitempty" gorm:"foreignkey:BusinessID"`
	CustomerProfiles []*CustomerProfile        `json:"customer_profiles,omitempty" gorm:"foreignkey:BusinessID"`
	RichMenus        []*RichMenu               `json:"rich_menus,omitempty" gorm:"foreignkey:BusinessID"`
	Tags             []*Tag                    `json:"tags,omitempty" gorm:"foreignkey:BusinessID"`
	Analytics        []*AnalyticsDaily         `json:"analytics,omitempty" gorm:"foreignkey:BusinessID"`
	WelcomeMessages  []*BusinessWelcomeMessage `json:"welcome_messages,omitempty" gorm:"foreignkey:BusinessID"` // เพิ่มความสัมพันธ์กับ BusinessWelcomeMessage
}

// TableName - ระบุชื่อตารางใน database
func (BusinessAccount) TableName() string {
	return "business_accounts"
}
