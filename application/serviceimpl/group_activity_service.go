// application/serviceimpl/group_activity_service.go
package serviceimpl

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type groupActivityService struct {
	activityRepo        repository.GroupActivityRepository
	conversationRepo    repository.ConversationRepository
	notificationService service.NotificationService
}

// NewGroupActivityService สร้าง service instance ใหม่
func NewGroupActivityService(
	activityRepo repository.GroupActivityRepository,
	conversationRepo repository.ConversationRepository,
	notificationService service.NotificationService,
) service.GroupActivityService {
	return &groupActivityService{
		activityRepo:        activityRepo,
		conversationRepo:    conversationRepo,
		notificationService: notificationService,
	}
}

// GetActivities ดึง activities ของ conversation พร้อม type filter
func (s *groupActivityService) GetActivities(conversationID, userID uuid.UUID, limit, offset int, activityType string) ([]*dto.ActivityDTO, int64, error) {
	// ตรวจสอบว่า user เป็นสมาชิกของ conversation หรือไม่
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !isMember {
		return nil, 0, errors.New("user is not a member of this conversation")
	}

	// ดึง activities จาก repository พร้อม type filter
	activities, total, err := s.activityRepo.GetByConversationID(conversationID, limit, offset, activityType)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTO
	dtos := make([]*dto.ActivityDTO, 0, len(activities))
	for _, activity := range activities {
		activityDTO := s.convertToActivityDTO(activity)
		dtos = append(dtos, activityDTO)
	}

	return dtos, total, nil
}

// convertToActivityDTO แปลง Activity model เป็น DTO
func (s *groupActivityService) convertToActivityDTO(activity *models.GroupActivity) *dto.ActivityDTO {
	activityDTO := &dto.ActivityDTO{
		ID:             activity.ID.String(),
		ConversationID: activity.ConversationID.String(),
		Type:           activity.Type,
		OldValue:       activity.OldValue,
		NewValue:       activity.NewValue,
		CreatedAt:      activity.CreatedAt,
	}

	// เพิ่มข้อมูล Actor
	if activity.Actor != nil {
		activityDTO.Actor = &dto.UserInfoDTO{
			ID:              activity.Actor.ID.String(),
			Username:        activity.Actor.Username,
			DisplayName:     activity.Actor.DisplayName,
			ProfileImageURL: activity.Actor.ProfileImageURL,
		}
	}

	// เพิ่มข้อมูล Target (ถ้ามี)
	if activity.TargetID != nil && activity.Target != nil {
		activityDTO.Target = &dto.UserInfoDTO{
			ID:              activity.Target.ID.String(),
			Username:        activity.Target.Username,
			DisplayName:     activity.Target.DisplayName,
			ProfileImageURL: activity.Target.ProfileImageURL,
		}
	}

	return activityDTO
}

// LogGroupCreated บันทึกการสร้างกลุ่ม
func (s *groupActivityService) LogGroupCreated(conversationID, creatorID uuid.UUID) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityGroupCreated,
		ActorID:        creatorID,
		CreatedAt:      time.Now(),
	}

	return s.activityRepo.Create(activity)
}

// LogGroupNameChanged บันทึกการเปลี่ยนชื่อกลุ่ม
func (s *groupActivityService) LogGroupNameChanged(conversationID, actorID uuid.UUID, oldName, newName string) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityGroupNameChanged,
		ActorID:        actorID,
		OldValue:       types.JSONB{"name": oldName},
		NewValue:       types.JSONB{"name": newName},
		CreatedAt:      time.Now(),
	}

	return s.activityRepo.Create(activity)
}

// LogGroupIconChanged บันทึกการเปลี่ยนไอคอนกลุ่ม
func (s *groupActivityService) LogGroupIconChanged(conversationID, actorID uuid.UUID, oldIcon, newIcon string) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityGroupIconChanged,
		ActorID:        actorID,
		OldValue:       types.JSONB{"icon_url": oldIcon},
		NewValue:       types.JSONB{"icon_url": newIcon},
		CreatedAt:      time.Now(),
	}

	return s.activityRepo.Create(activity)
}

// LogMemberAdded บันทึกการเพิ่มสมาชิก
func (s *groupActivityService) LogMemberAdded(conversationID, actorID, targetID uuid.UUID) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityMemberAdded,
		ActorID:        actorID,
		TargetID:       &targetID,
		CreatedAt:      time.Now(),
	}

	if err := s.activityRepo.Create(activity); err != nil {
		println("❌ [LogMemberAdded] Error creating activity:", err.Error())
		return err
	}

	println("✅ [LogMemberAdded] Activity created:", activity.ID.String())
	return nil
}

// LogMemberRemoved บันทึกการลบสมาชิก
func (s *groupActivityService) LogMemberRemoved(conversationID, actorID, targetID uuid.UUID) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityMemberRemoved,
		ActorID:        actorID,
		TargetID:       &targetID,
		CreatedAt:      time.Now(),
	}

	return s.activityRepo.Create(activity)
}

// LogMemberRoleChanged บันทึกการเปลี่ยน role ของสมาชิก
func (s *groupActivityService) LogMemberRoleChanged(conversationID, actorID, targetID uuid.UUID, oldRole, newRole string) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityMemberRoleChanged,
		ActorID:        actorID,
		TargetID:       &targetID,
		OldValue:       types.JSONB{"role": oldRole},
		NewValue:       types.JSONB{"role": newRole},
		CreatedAt:      time.Now(),
	}

	if err := s.activityRepo.Create(activity); err != nil {
		return err
	}

	// Broadcast WebSocket event พร้อม user info
	activityWithUsers, err := s.activityRepo.GetByID(activity.ID)
	if err == nil && s.notificationService != nil {
		activityDTO := s.convertToActivityDTO(activityWithUsers)
		s.notificationService.NotifyNewActivity(conversationID, activityDTO)
	}

	return nil
}

// LogOwnershipTransferred บันทึกการโอนความเป็นเจ้าของ
func (s *groupActivityService) LogOwnershipTransferred(conversationID, oldOwnerID, newOwnerID uuid.UUID) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityOwnershipTransferred,
		ActorID:        oldOwnerID,
		TargetID:       &newOwnerID,
		CreatedAt:      time.Now(),
	}

	if err := s.activityRepo.Create(activity); err != nil {
		return err
	}

	// Broadcast WebSocket event พร้อม user info
	activityWithUsers, err := s.activityRepo.GetByID(activity.ID)
	if err == nil && s.notificationService != nil {
		activityDTO := s.convertToActivityDTO(activityWithUsers)
		s.notificationService.NotifyNewActivity(conversationID, activityDTO)
	}

	return nil
}

// LogMemberLeft บันทึกการออกจากกลุ่ม
func (s *groupActivityService) LogMemberLeft(conversationID, userID uuid.UUID) error {
	activity := &models.GroupActivity{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Type:           models.ActivityMemberLeft,
		ActorID:        userID,
		CreatedAt:      time.Now(),
	}

	return s.activityRepo.Create(activity)
}
