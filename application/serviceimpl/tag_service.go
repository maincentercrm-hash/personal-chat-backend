// application/serviceimpl/tag_service.go
package serviceimpl

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

type tagService struct {
	tagRepo             repository.TagRepository
	userTagRepo         repository.UserTagRepository
	customerProfileRepo repository.CustomerProfileRepository
	businessAdminRepo   repository.BusinessAdminRepository
	businessAccountRepo repository.BusinessAccountRepository
}

// NewTagService สร้าง instance ใหม่ของ TagService
func NewTagService(
	tagRepo repository.TagRepository,
	userTagRepo repository.UserTagRepository,
	customerProfileRepo repository.CustomerProfileRepository,
	businessAdminRepo repository.BusinessAdminRepository,
	businessAccountRepo repository.BusinessAccountRepository,
) service.TagService {
	return &tagService{
		tagRepo:             tagRepo,
		userTagRepo:         userTagRepo,
		customerProfileRepo: customerProfileRepo,
		businessAdminRepo:   businessAdminRepo,
		businessAccountRepo: businessAccountRepo,
	}
}

// CreateTag สร้างแท็กใหม่
func (s *tagService) CreateTag(businessID, createdByID uuid.UUID, name, color string) (*models.Tag, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(createdByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to create tags")
	}

	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	// ตรวจสอบชื่อแท็ก
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("tag name is required")
	}
	if len(name) > 50 {
		return nil, errors.New("tag name too long (max 50 characters)")
	}

	// ตรวจสอบสี (ถ้ามี)
	if color != "" && !isValidColor(color) {
		return nil, errors.New("invalid color format")
	}

	// สร้างแท็กใหม่
	tag := &models.Tag{
		ID:          uuid.New(),
		BusinessID:  businessID,
		Name:        name,
		Color:       color,
		CreatedAt:   time.Now(),
		CreatedByID: &createdByID,
	}

	// บันทึกลงฐานข้อมูล
	err = s.tagRepo.Create(tag)
	if err != nil {
		// ตรวจสอบว่าเป็น duplicate name หรือไม่
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.New("tag name already exists in this business")
		}
		return nil, err
	}

	return tag, nil
}

// GetBusinessTags ดึงแท็กทั้งหมดของธุรกิจ
func (s *tagService) GetBusinessTags(businessID uuid.UUID) ([]*models.Tag, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	// ดึงแท็กทั้งหมด
	tags, err := s.tagRepo.GetByBusinessID(businessID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// AddTagToUser เพิ่มแท็กให้กับผู้ใช้
func (s *tagService) AddTagToUser(businessID, userID, tagID, addedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(addedByID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to manage user tags")
	}

	// ตรวจสอบว่าแท็กมีอยู่จริงและเป็นของธุรกิจนี้
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return errors.New("tag not found")
	}
	if tag.BusinessID != businessID {
		return errors.New("tag does not belong to this business")
	}

	// ตรวจสอบว่ามี CustomerProfile หรือไม่ (สร้างถ้าไม่มี)
	_, err = s.customerProfileRepo.GetByBusinessAndUser(businessID, userID)
	if err != nil {
		// ถ้าไม่มีโปรไฟล์ ให้สร้างใหม่ผ่าน CustomerProfileService
		return errors.New("customer profile not found, please create customer profile first")
	}

	// ตรวจสอบว่าผู้ใช้มีแท็กนี้อยู่แล้วหรือไม่
	userTags, err := s.userTagRepo.GetUserTags(businessID, userID)
	if err != nil {
		return err
	}

	// ตรวจสอบการซ้ำซ้อน
	for _, existingTag := range userTags {
		if existingTag.TagID == tagID {
			return errors.New("user already has this tag")
		}
	}

	// สร้าง UserTag
	userTag := &models.UserTag{
		ID:         uuid.New(),
		UserID:     userID,
		TagID:      tagID,
		BusinessID: businessID,
		AddedAt:    time.Now(),
		AddedByID:  &addedByID,
	}

	// บันทึกลงฐานข้อมูล
	err = s.userTagRepo.Create(userTag)
	if err != nil {
		// ตรวจสอบว่าเป็น duplicate หรือไม่
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return errors.New("user already has this tag")
		}
		return err
	}

	return nil
}

// RemoveTagFromUser ลบแท็กออกจากผู้ใช้
func (s *tagService) RemoveTagFromUser(businessID, userID, tagID uuid.UUID) error {
	// ลบ UserTag
	return s.userTagRepo.Delete(businessID, userID, tagID)
}

// GetUserTags ดึงแท็กทั้งหมดของผู้ใช้
func (s *tagService) GetUserTags(businessID, userID uuid.UUID) ([]*models.Tag, error) {
	// ดึง UserTags
	userTags, err := s.userTagRepo.GetUserTags(businessID, userID)
	if err != nil {
		return nil, err
	}

	// แปลงเป็น Tags
	tags := make([]*models.Tag, 0, len(userTags))
	for _, userTag := range userTags {
		if userTag.Tag != nil {
			tags = append(tags, userTag.Tag)
		}
	}

	return tags, nil
}

// GetUsersByTag ดึงรายชื่อผู้ใช้ที่มีแท็กนี้
func (s *tagService) GetUsersByTag(businessID, tagID uuid.UUID) ([]*models.CustomerProfile, error) {
	// ตรวจสอบว่าแท็กมีอยู่จริงและเป็นของธุรกิจนี้
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return nil, errors.New("tag not found")
	}
	if tag.BusinessID != businessID {
		return nil, errors.New("tag does not belong to this business")
	}

	// ดึง UserTags
	userTags, err := s.userTagRepo.GetUsersByTag(businessID, tagID)
	if err != nil {
		return nil, err
	}

	// ดึง CustomerProfiles สำหรับแต่ละ user
	profiles := make([]*models.CustomerProfile, 0, len(userTags))
	for _, userTag := range userTags {
		profile, err := s.customerProfileRepo.GetByBusinessAndUser(businessID, userTag.UserID)
		if err != nil {
			continue // ข้ามกรณีที่ไม่พบโปรไฟล์
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// UpdateTag อัปเดตแท็ก
func (s *tagService) UpdateTag(businessID, tagID, updatedByID uuid.UUID, name, color string) (*models.Tag, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(updatedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to update tags")
	}

	// ดึงแท็กปัจจุบัน
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return nil, errors.New("tag not found")
	}
	if tag.BusinessID != businessID {
		return nil, errors.New("tag does not belong to this business")
	}

	// อัปเดตข้อมูล
	if name != "" {
		name = strings.TrimSpace(name)
		if len(name) > 50 {
			return nil, errors.New("tag name too long (max 50 characters)")
		}
		tag.Name = name
	}

	if color != "" && isValidColor(color) {
		tag.Color = color
	}

	// บันทึกการเปลี่ยนแปลง
	err = s.tagRepo.Update(tag)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.New("tag name already exists in this business")
		}
		return nil, err
	}

	return tag, nil
}

// DeleteTag ลบแท็ก
func (s *tagService) DeleteTag(businessID, tagID, deletedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(deletedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to delete tags")
	}

	// ตรวจสอบว่าแท็กมีอยู่จริงและเป็นของธุรกิจนี้
	tag, err := s.tagRepo.GetByID(tagID)
	if err != nil {
		return errors.New("tag not found")
	}
	if tag.BusinessID != businessID {
		return errors.New("tag does not belong to this business")
	}

	// ลบแท็ก (จะลบ UserTags ที่เกี่ยวข้องด้วย via CASCADE)
	return s.tagRepo.Delete(tagID)
}

// isValidColor ตรวจสอบรูปแบบสี
func isValidColor(color string) bool {
	// รองรับ hex color (#RRGGBB หรือ #RGB)
	if strings.HasPrefix(color, "#") {
		color = color[1:]
		return len(color) == 3 || len(color) == 6
	}

	// รองรับสีตามชื่อพื้นฐาน
	basicColors := []string{
		"red", "blue", "green", "yellow", "orange", "purple",
		"pink", "brown", "gray", "black", "white",
	}

	for _, basicColor := range basicColors {
		if strings.ToLower(color) == basicColor {
			return true
		}
	}

	return false
}

func (s *tagService) GetBusinessTagsWithInfo(businessID uuid.UUID) ([]dto.TagInfo, error) {
	// ดึงแท็กทั้งหมด
	tags, err := s.GetBusinessTags(businessID)
	if err != nil {
		return nil, err
	}

	// แปลงเป็น TagInfo DTOs
	tagInfos := make([]dto.TagInfo, 0, len(tags))
	for _, tag := range tags {
		// นับจำนวนผู้ใช้ที่มีแท็กนี้
		userTags, err := s.userTagRepo.GetUsersByTag(businessID, tag.ID)
		userCount := 0
		if err == nil {
			userCount = len(userTags)
		}

		// แปลงเป็น TagInfo DTO
		tagInfo := dto.TagInfo{
			ID:         tag.ID.String(),
			BusinessID: tag.BusinessID.String(),
			Name:       tag.Name,
			Color:      tag.Color,
			CreatedAt:  tag.CreatedAt,
			UserCount:  userCount,
		}

		// ถ้ามีข้อมูลผู้สร้าง
		if tag.CreatedByID != nil {
			tagInfo.CreatedBy = tag.CreatedByID.String()
		}

		tagInfos = append(tagInfos, tagInfo)
	}

	return tagInfos, nil
}
