// domain/service/group_activity_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
)

// GroupActivityService interface สำหรับจัดการ group activities
type GroupActivityService interface {
	// GetActivities ดึง activities ของ conversation (ถ้า activityType ไม่ว่าง จะ filter ตาม type)
	GetActivities(conversationID, userID uuid.UUID, limit, offset int, activityType string) ([]*dto.ActivityDTO, int64, error)

	// Helper methods สำหรับบันทึก activity แต่ละประเภท
	LogGroupCreated(conversationID, creatorID uuid.UUID) error
	LogGroupNameChanged(conversationID, actorID uuid.UUID, oldName, newName string) error
	LogGroupIconChanged(conversationID, actorID uuid.UUID, oldIcon, newIcon string) error
	LogMemberAdded(conversationID, actorID, targetID uuid.UUID) error
	LogMemberRemoved(conversationID, actorID, targetID uuid.UUID) error
	LogMemberRoleChanged(conversationID, actorID, targetID uuid.UUID, oldRole, newRole string) error
	LogOwnershipTransferred(conversationID, oldOwnerID, newOwnerID uuid.UUID) error
	LogMemberLeft(conversationID, userID uuid.UUID) error
}
