// application/serviceimpl/user_tag_service.go
package serviceimpl

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

type userTagService struct {
	userTagRepo         repository.UserTagRepository
	tagRepo             repository.TagRepository
	customerProfileRepo repository.CustomerProfileRepository
	businessAdminRepo   repository.BusinessAdminRepository
	businessAccountRepo repository.BusinessAccountRepository
	userRepo            repository.UserRepository
}

// NewUserTagService สร้าง instance ใหม่ของ UserTagService
func NewUserTagService(
	userTagRepo repository.UserTagRepository,
	tagRepo repository.TagRepository,
	customerProfileRepo repository.CustomerProfileRepository,
	businessAdminRepo repository.BusinessAdminRepository,
	businessAccountRepo repository.BusinessAccountRepository,
	userRepo repository.UserRepository,
) service.UserTagService {
	return &userTagService{
		userTagRepo:         userTagRepo,
		tagRepo:             tagRepo,
		customerProfileRepo: customerProfileRepo,
		businessAdminRepo:   businessAdminRepo,
		businessAccountRepo: businessAccountRepo,
		userRepo:            userRepo,
	}
}

// AddTagToUser เพิ่มแท็กให้กับผู้ใช้
func (s *userTagService) AddTagToUser(businessID, userID, tagID, addedByID uuid.UUID) (*models.UserTag, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(addedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to manage user tags")
	}

	// ตรวจสอบว่าแท็กมีอยู่จริงและเป็นของธุรกิจนี้
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return nil, errors.New("tag not found")
	}
	if tag.BusinessID != businessID {
		return nil, errors.New("tag does not belong to this business")
	}

	// ตรวจสอบว่ามี CustomerProfile หรือไม่
	_, err = s.customerProfileRepo.GetByBusinessAndUser(businessID, userID)
	if err != nil {
		return nil, errors.New("customer profile not found, please create customer profile first")
	}

	// ตรวจสอบว่าผู้ใช้มีแท็กนี้อยู่แล้วหรือไม่
	hasTag, err := s.CheckUserHasTag(businessID, userID, tagID)
	if err != nil {
		return nil, err
	}
	if hasTag {
		return nil, errors.New("user already has this tag")
	}

	// สร้าง UserTag ใหม่
	userTag := &models.UserTag{
		ID:         uuid.New(),
		UserID:     userID,
		TagID:      tagID,
		BusinessID: businessID,
		AddedAt:    time.Now(),
		AddedByID:  &addedByID,
		Tag:        tag,
	}

	// บันทึกลงฐานข้อมูล
	err = s.userTagRepo.Create(userTag)
	if err != nil {
		return nil, err
	}

	// โหลดข้อมูลผู้ใช้
	user, err := s.userRepo.FindByID(userID)
	if err == nil {
		userTag.User = user
	}

	return userTag, nil
}

// RemoveTagFromUser ลบแท็กออกจากผู้ใช้
func (s *userTagService) RemoveTagFromUser(businessID, userID, tagID, requestedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to manage user tags")
	}

	// ตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่
	hasTag, err := s.CheckUserHasTag(businessID, userID, tagID)
	if err != nil {
		return err
	}
	if !hasTag {
		return errors.New("user does not have this tag")
	}

	// ลบแท็ก
	return s.userTagRepo.Delete(businessID, userID, tagID)
}

// GetUserTags ดึงแท็กทั้งหมดของผู้ใช้ในธุรกิจ
func (s *userTagService) GetUserTags(businessID, userID uuid.UUID, requestedByID uuid.UUID) ([]*models.UserTag, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to view user tags")
	}

	// ดึงแท็กของผู้ใช้
	userTags, err := s.userTagRepo.GetUserTags(businessID, userID)
	if err != nil {
		return nil, err
	}

	return userTags, nil
}

// GetUsersByTag ดึงรายชื่อผู้ใช้ที่มีแท็กนี้
func (s *userTagService) GetUsersByTag(businessID, tagID uuid.UUID, requestedByID uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return nil, 0, err
	}
	if !hasPermission {
		return nil, 0, errors.New("you don't have permission to view user tags")
	}

	// ตรวจสอบว่าแท็กมีอยู่จริงและเป็นของธุรกิจนี้
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return nil, 0, errors.New("tag not found")
	}
	if tag.BusinessID != businessID {
		return nil, 0, errors.New("tag does not belong to this business")
	}

	// ตั้งค่าเริ่มต้นสำหรับ pagination
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// ดึงผู้ใช้ที่มีแท็กนี้
	userTags, err := s.userTagRepo.GetUsersByTag(businessID, tagID)
	if err != nil {
		return nil, 0, err
	}

	// จำลอง pagination (ใน implementation จริงควรทำใน repository)
	total := int64(len(userTags))
	start := offset
	end := offset + limit
	if start > len(userTags) {
		start = len(userTags)
	}
	if end > len(userTags) {
		end = len(userTags)
	}

	paginatedUserTags := userTags[start:end]

	return paginatedUserTags, total, nil
}

// BulkAddTagToUsers เพิ่มแท็กให้กับผู้ใช้หลายคน
func (s *userTagService) BulkAddTagToUsers(businessID, tagID uuid.UUID, userIDs []uuid.UUID, addedByID uuid.UUID) ([]*models.UserTag, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(addedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to manage user tags")
	}

	// ตรวจสอบว่าแท็กมีอยู่จริงและเป็นของธุรกิจนี้
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return nil, errors.New("tag not found")
	}
	if tag.BusinessID != businessID {
		return nil, errors.New("tag does not belong to this business")
	}

	var addedUserTags []*models.UserTag
	var errors []string

	for _, userID := range userIDs {
		userTag, err := s.AddTagToUser(businessID, userID, tagID, addedByID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("User %s: %s", userID.String(), err.Error()))
			continue
		}
		addedUserTags = append(addedUserTags, userTag)
	}

	// ถ้ามี error บางส่วน ให้รายงาน
	if len(errors) > 0 {
		return addedUserTags, fmt.Errorf("some operations failed: %v", errors)
	}

	return addedUserTags, nil
}

// BulkRemoveTagFromUsers ลบแท็กออกจากผู้ใช้หลายคน
func (s *userTagService) BulkRemoveTagFromUsers(businessID, tagID uuid.UUID, userIDs []uuid.UUID, requestedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to manage user tags")
	}

	var errors []string

	for _, userID := range userIDs {
		err := s.RemoveTagFromUser(businessID, userID, tagID, requestedByID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("User %s: %s", userID.String(), err.Error()))
		}
	}

	// ถ้ามี error บางส่วน ให้รายงาน
	if len(errors) > 0 {
		return fmt.Errorf("some operations failed: %v", errors)
	}

	return nil
}

// ReplaceUserTags แทนที่แท็กทั้งหมดของผู้ใช้ด้วยแท็กใหม่
func (s *userTagService) ReplaceUserTags(businessID, userID uuid.UUID, tagIDs []uuid.UUID, updatedByID uuid.UUID) ([]*models.UserTag, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(updatedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to manage user tags")
	}

	// ดึงแท็กปัจจุบันของผู้ใช้
	currentUserTags, err := s.userTagRepo.GetUserTags(businessID, userID)
	if err != nil {
		return nil, err
	}

	// ลบแท็กเก่าทั้งหมด
	for _, userTag := range currentUserTags {
		err := s.userTagRepo.Delete(businessID, userID, userTag.TagID)
		if err != nil {
			return nil, fmt.Errorf("failed to remove old tag %s: %v", userTag.TagID.String(), err)
		}
	}

	// เพิ่มแท็กใหม่
	var newUserTags []*models.UserTag
	for _, tagID := range tagIDs {
		userTag, err := s.AddTagToUser(businessID, userID, tagID, updatedByID)
		if err != nil {
			return nil, fmt.Errorf("failed to add new tag %s: %v", tagID.String(), err)
		}
		newUserTags = append(newUserTags, userTag)
	}

	return newUserTags, nil
}

// GetTagStatistics ดึงสถิติการใช้แท็ก
func (s *userTagService) GetTagStatistics(businessID uuid.UUID, requestedByID uuid.UUID) ([]service.TagStatistic, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to view tag statistics")
	}

	// ดึงแท็กทั้งหมดของธุรกิจ
	tags, err := s.tagRepo.GetByBusinessID(businessID)
	if err != nil {
		return nil, err
	}

	var statistics []service.TagStatistic

	for _, tag := range tags {
		// นับจำนวนผู้ใช้ที่มีแท็กนี้
		userTags, err := s.userTagRepo.GetUsersByTag(businessID, tag.ID)
		if err != nil {
			continue
		}

		statistic := service.TagStatistic{
			TagID:     tag.ID,
			TagName:   tag.Name,
			TagColor:  tag.Color,
			UserCount: int64(len(userTags)),
			CreatedAt: tag.CreatedAt.Format("2006-01-02"),
		}

		statistics = append(statistics, statistic)
	}

	return statistics, nil
}

// SearchUsersByTags ค้นหาผู้ใช้ที่มีแท็กตามเงื่อนไข
func (s *userTagService) SearchUsersByTags(businessID uuid.UUID, criteria service.TagSearchCriteria, requestedByID uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return nil, 0, err
	}
	if !hasPermission {
		return nil, 0, errors.New("you don't have permission to search user tags")
	}

	// ตั้งค่าเริ่มต้นสำหรับ pagination
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// สำหรับการ implement ที่ซับซ้อน ควรทำใน repository layer
	// ตอนนี้ implement แบบง่ายๆ ก่อน
	var matchedUserTags []*models.UserTag

	if len(criteria.IncludeTags) > 0 {
		// ใช้แท็กแรกเป็นฐาน
		firstTagID := criteria.IncludeTags[0]
		userTags, err := s.userTagRepo.GetUsersByTag(businessID, firstTagID)
		if err != nil {
			return nil, 0, err
		}

		// กรองตามเงื่อนไข
		if criteria.MatchType == service.TagMatchAll {
			// ต้องมีทุกแท็ก (AND logic)
			for _, userTag := range userTags {
				hasAllTags := true
				for _, requiredTagID := range criteria.IncludeTags {
					hasTag, err := s.CheckUserHasTag(businessID, userTag.UserID, requiredTagID)
					if err != nil || !hasTag {
						hasAllTags = false
						break
					}
				}
				if hasAllTags {
					matchedUserTags = append(matchedUserTags, userTag)
				}
			}
		} else {
			// มีแท็กใดแท็กหนึ่งก็พอ (OR logic)
			userMap := make(map[uuid.UUID]*models.UserTag)
			for _, tagID := range criteria.IncludeTags {
				userTags, err := s.userTagRepo.GetUsersByTag(businessID, tagID)
				if err != nil {
					continue
				}
				for _, userTag := range userTags {
					userMap[userTag.UserID] = userTag
				}
			}
			for _, userTag := range userMap {
				matchedUserTags = append(matchedUserTags, userTag)
			}
		}
	}

	// กรอง exclude tags
	if len(criteria.ExcludeTags) > 0 {
		var filteredUserTags []*models.UserTag
		for _, userTag := range matchedUserTags {
			hasExcludedTag := false
			for _, excludeTagID := range criteria.ExcludeTags {
				hasTag, err := s.CheckUserHasTag(businessID, userTag.UserID, excludeTagID)
				if err == nil && hasTag {
					hasExcludedTag = true
					break
				}
			}
			if !hasExcludedTag {
				filteredUserTags = append(filteredUserTags, userTag)
			}
		}
		matchedUserTags = filteredUserTags
	}

	// Pagination
	total := int64(len(matchedUserTags))
	start := offset
	end := offset + limit
	if start > len(matchedUserTags) {
		start = len(matchedUserTags)
	}
	if end > len(matchedUserTags) {
		end = len(matchedUserTags)
	}

	paginatedUserTags := matchedUserTags[start:end]

	return paginatedUserTags, total, nil
}

// CheckUserHasTag ตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่
func (s *userTagService) CheckUserHasTag(businessID, userID, tagID uuid.UUID) (bool, error) {
	userTags, err := s.userTagRepo.GetUserTags(businessID, userID)
	if err != nil {
		return false, err
	}

	for _, userTag := range userTags {
		if userTag.TagID == tagID {
			return true, nil
		}
	}

	return false, nil
}

// GetUsersWithMultipleTags ดึงผู้ใช้ที่มีแท็กตามที่กำหนด (AND/OR logic)
func (s *userTagService) GetUsersWithMultipleTags(businessID uuid.UUID, tagIDs []uuid.UUID, matchType service.TagMatchType, requestedByID uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error) {
	criteria := service.TagSearchCriteria{
		IncludeTags: tagIDs,
		MatchType:   matchType,
	}

	return s.SearchUsersByTags(businessID, criteria, requestedByID, limit, offset)
}
