// domain/models/group_activity.go
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// GroupActivity บันทึกกิจกรรมต่างๆ ในกลุ่ม
type GroupActivity struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ConversationID uuid.UUID   `json:"conversation_id" gorm:"type:uuid;not null;index:idx_group_activities_conversation"`
	Type           string      `json:"type" gorm:"type:varchar(50);not null;index:idx_group_activities_type"`
	ActorID        uuid.UUID   `json:"actor_id" gorm:"type:uuid;not null"`
	TargetID       *uuid.UUID  `json:"target_id,omitempty" gorm:"type:uuid"`
	OldValue       types.JSONB `json:"old_value,omitempty" gorm:"type:jsonb"`
	NewValue       types.JSONB `json:"new_value,omitempty" gorm:"type:jsonb"`
	CreatedAt      time.Time   `json:"created_at" gorm:"type:timestamp with time zone;default:now();index:idx_group_activities_created_at,sort:desc"`

	// Associations
	Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignkey:ConversationID"`
	Actor        *User         `json:"actor,omitempty" gorm:"foreignkey:ActorID"`
	Target       *User         `json:"target,omitempty" gorm:"foreignkey:TargetID"`
}

// TableName กำหนดชื่อตาราง
func (GroupActivity) TableName() string {
	return "group_activities"
}

// Activity Types - กำหนดประเภทของกิจกรรม
const (
	ActivityGroupCreated         = "group.created"
	ActivityGroupNameChanged     = "group.name_changed"
	ActivityGroupIconChanged     = "group.icon_changed"
	ActivityMemberAdded          = "member.added"
	ActivityMemberRemoved        = "member.removed"
	ActivityMemberRoleChanged    = "member.role_changed"
	ActivityOwnershipTransferred = "ownership.transferred"
	ActivityMemberLeft           = "member.left"
)
