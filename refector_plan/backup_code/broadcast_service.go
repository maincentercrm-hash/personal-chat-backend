// application/serviceimpl/broadcast_service.go
package serviceimpl

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type broadcastService struct {
	broadcastRepo         repository.BroadcastRepository
	broadcastDeliveryRepo repository.BroadcastDeliveryRepository
	businessAccountRepo   repository.BusinessAccountRepository
	businessAdminRepo     repository.BusinessAdminRepository
	userRepo              repository.UserRepository
	businessFollowRepo    repository.BusinessFollowRepository
	userTagRepo           repository.UserTagRepository
	tagRepo               repository.TagRepository
	customerProfileRepo   repository.CustomerProfileRepository
	messageService        service.MessageService
	userTagService        service.UserTagService
	conversationRepo      repository.ConversationRepository
}

// NewBroadcastService สร้าง instance ใหม่ของ BroadcastService
func NewBroadcastService(
	broadcastRepo repository.BroadcastRepository,
	broadcastDeliveryRepo repository.BroadcastDeliveryRepository,
	businessAccountRepo repository.BusinessAccountRepository,
	businessAdminRepo repository.BusinessAdminRepository,
	userRepo repository.UserRepository,
	businessFollowRepo repository.BusinessFollowRepository,
	userTagRepo repository.UserTagRepository,
	tagRepo repository.TagRepository,
	customerProfileRepo repository.CustomerProfileRepository,
	messageService service.MessageService,
	userTagService service.UserTagService,
	conversationRepo repository.ConversationRepository,
) service.BroadcastService {
	return &broadcastService{
		broadcastRepo:         broadcastRepo,
		broadcastDeliveryRepo: broadcastDeliveryRepo,
		businessAccountRepo:   businessAccountRepo,
		businessAdminRepo:     businessAdminRepo,
		userRepo:              userRepo,
		businessFollowRepo:    businessFollowRepo,
		userTagRepo:           userTagRepo,
		tagRepo:               tagRepo,
		customerProfileRepo:   customerProfileRepo,
		messageService:        messageService,
		userTagService:        userTagService,
		conversationRepo:      conversationRepo,
	}
}

// CreateBroadcast สร้าง broadcast ใหม่
func (s *broadcastService) CreateBroadcast(
	businessID, createdByID uuid.UUID,
	title, messageType, content string,
	mediaURL string,
	bubbleType string,
	bubbleData types.JSONB,
) (*models.Broadcast, error) {
	// ตรวจสอบสิทธิ์ของผู้สร้าง
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(createdByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to create broadcasts for this business")
	}

	// ตรวจสอบความถูกต้องของข้อมูล
	if err := s.ValidateBroadcastContent(messageType, content, mediaURL, bubbleType, bubbleData); err != nil {
		return nil, err
	}

	// ตรวจสอบชื่อเรื่อง
	if title == "" {
		return nil, errors.New("title is required")
	}
	if len(title) > 100 {
		return nil, errors.New("title is too long (max 100 characters)")
	}

	// สร้าง Broadcast ใหม่
	broadcast := &models.Broadcast{
		ID:          uuid.New(),
		BusinessID:  businessID,
		Title:       title,
		MessageType: messageType,
		Content:     content,
		MediaURL:    mediaURL,
		CreatedAt:   time.Now(),
		CreatedBy:   &createdByID,
		Status:      "draft",
		TargetType:  "all", // ค่าเริ่มต้น: ส่งถึงทั้งหมด
		BubbleType:  bubbleType,
		BubbleData:  bubbleData,
		Metrics:     types.JSONB{"total_targeted": 0, "delivered": 0, "opened": 0, "clicked": 0},
	}

	// บันทึกลงฐานข้อมูล
	if err := s.broadcastRepo.Create(broadcast); err != nil {
		return nil, err
	}

	return broadcast, nil
}

// GetBroadcastByID ดึงข้อมูล broadcast ตาม ID
func (s *broadcastService) GetBroadcastByID(id, businessID, requestedByID uuid.UUID) (*models.Broadcast, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to access this broadcast")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return nil, errors.New("broadcast does not belong to the specified business")
	}

	return broadcast, nil
}

// GetBusinessBroadcasts ดึงข้อมูล broadcasts ทั้งหมดของธุรกิจ
func (s *broadcastService) GetBusinessBroadcasts(businessID, requestedByID uuid.UUID, status string, limit, offset int) ([]*models.Broadcast, int64, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{})
	if err != nil {
		return nil, 0, err
	}
	if !hasPermission {
		return nil, 0, errors.New("you don't have permission to access broadcasts of this business")
	}

	// ตั้งค่า pagination
	if limit <= 0 || limit > 100 {
		limit = 20 // ค่าเริ่มต้น
	}
	if offset < 0 {
		offset = 0
	}

	// ดึงข้อมูล broadcasts
	return s.broadcastRepo.GetByBusinessID(businessID, status, limit, offset)
}

// UpdateBroadcast อัพเดทข้อมูล broadcast
func (s *broadcastService) UpdateBroadcast(id, businessID, requestedByID uuid.UUID, updateData types.JSONB) (*models.Broadcast, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to update broadcasts of this business")
	}

	// ดึงข้อมูล broadcast ปัจจุบัน
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return nil, errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถแก้ไขได้หรือไม่
	if broadcast.Status != "draft" {
		return nil, errors.New("only broadcasts in draft status can be updated")
	}

	// อัพเดทข้อมูลตามที่ระบุ
	updated := false

	if title, ok := updateData["title"].(string); ok && title != "" {
		if len(title) > 100 {
			return nil, errors.New("title is too long (max 100 characters)")
		}
		broadcast.Title = title
		updated = true
	}

	if messageType, ok := updateData["message_type"].(string); ok && messageType != "" {
		broadcast.MessageType = messageType
		updated = true
	}

	if content, ok := updateData["content"].(string); ok {
		broadcast.Content = content
		updated = true
	}

	if mediaURL, ok := updateData["media_url"].(string); ok {
		broadcast.MediaURL = mediaURL
		updated = true
	}

	if bubbleType, ok := updateData["bubble_type"].(string); ok {
		broadcast.BubbleType = bubbleType
		updated = true
	}

	if bubbleData, ok := updateData["bubble_data"].(map[string]interface{}); ok {
		broadcast.BubbleData = types.JSONB(bubbleData)
		updated = true
	}

	// ตรวจสอบความถูกต้องของข้อมูลหลังอัพเดท
	if err := s.ValidateBroadcastContent(
		broadcast.MessageType,
		broadcast.Content,
		broadcast.MediaURL,
		broadcast.BubbleType,
		broadcast.BubbleData,
	); err != nil {
		return nil, err
	}

	// ถ้าไม่มีข้อมูลที่ต้องอัพเดท
	if !updated {
		return nil, errors.New("no valid data to update")
	}

	// บันทึกการอัพเดท
	if err := s.broadcastRepo.Update(broadcast); err != nil {
		return nil, err
	}

	return broadcast, nil
}

// DeleteBroadcast ลบ broadcast
func (s *broadcastService) DeleteBroadcast(id, businessID, requestedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to delete broadcasts of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถลบได้หรือไม่
	if broadcast.Status != "draft" && broadcast.Status != "failed" && broadcast.Status != "completed" {
		return errors.New("broadcasts that are being sent or scheduled cannot be deleted")
	}

	// ลบข้อมูลที่เกี่ยวข้องในตาราง broadcast_deliveries ก่อน
	if err := s.broadcastDeliveryRepo.DeleteByBroadcastID(id); err != nil {
		return err
	}

	// ลบ broadcast
	return s.broadcastRepo.Delete(id)
}

// ScheduleBroadcast กำหนดเวลาส่ง broadcast
func (s *broadcastService) ScheduleBroadcast(id, businessID, requestedByID uuid.UUID, scheduledAt time.Time) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to schedule broadcasts of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถกำหนดเวลาได้หรือไม่
	if broadcast.Status != "draft" {
		return errors.New("only broadcasts in draft status can be scheduled")
	}

	// ตรวจสอบเวลาที่กำหนด
	now := time.Now()
	if scheduledAt.Before(now) {
		return errors.New("scheduled time must be in the future")
	}

	// ตรวจสอบเนื้อหาว่าสามารถส่งได้หรือไม่
	if err := s.ValidateBroadcastContent(
		broadcast.MessageType,
		broadcast.Content,
		broadcast.MediaURL,
		broadcast.BubbleType,
		broadcast.BubbleData,
	); err != nil {
		return err
	}

	// ตรวจสอบว่ามีกลุ่มเป้าหมายหรือไม่
	if _, err := s.getTargetUsers(broadcast); err != nil {
		return fmt.Errorf("failed to get target users: %v", err)
	}

	// กำหนดเวลาส่ง
	broadcast.ScheduledAt = &scheduledAt
	broadcast.Status = "scheduled"

	// บันทึกการอัพเดท
	return s.broadcastRepo.Update(broadcast)
}

// CancelScheduledBroadcast ยกเลิกการกำหนดเวลาส่ง
func (s *broadcastService) CancelScheduledBroadcast(id, businessID, requestedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to cancel scheduled broadcasts of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถยกเลิกการกำหนดเวลาได้หรือไม่
	if broadcast.Status != "scheduled" {
		return errors.New("only scheduled broadcasts can be canceled")
	}

	// ยกเลิกการกำหนดเวลาส่ง
	broadcast.ScheduledAt = nil
	broadcast.Status = "draft"

	// บันทึกการอัพเดท
	return s.broadcastRepo.Update(broadcast)
}

// SendBroadcast ส่ง broadcast ทันที
func (s *broadcastService) SendBroadcast(id, businessID, requestedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to send broadcasts of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// แก้ไขตรงนี้ให้รองรับทั้ง draft และ scheduled
	if broadcast.Status != "draft" && broadcast.Status != "scheduled" {
		return errors.New("only broadcasts in draft or scheduled status can be sent")
	}

	// ตรวจสอบเนื้อหาว่าสามารถส่งได้หรือไม่
	if err := s.ValidateBroadcastContent(
		broadcast.MessageType,
		broadcast.Content,
		broadcast.MediaURL,
		broadcast.BubbleType,
		broadcast.BubbleData,
	); err != nil {
		return err
	}

	// ดึงรายชื่อผู้รับตามเงื่อนไข target
	targetUsers, err := s.getTargetUsers(broadcast)
	if err != nil {
		return fmt.Errorf("failed to get target users: %v", err)
	}

	if len(targetUsers) == 0 {
		return errors.New("no target users found for this broadcast")
	}

	// อัพเดทสถานะ broadcast เป็น sending
	broadcast.Status = "sending"
	now := time.Now()
	broadcast.SentAt = &now
	if err := s.broadcastRepo.Update(broadcast); err != nil {
		return err
	}

	// สร้าง delivery records
	if err := s.createDeliveries(broadcast, targetUsers); err != nil {
		// ถ้าเกิดข้อผิดพลาด ให้อัพเดทสถานะเป็น failed
		broadcast.Status = "failed"
		broadcast.ErrorMessage = fmt.Sprintf("Failed to create deliveries: %v", err)
		s.broadcastRepo.Update(broadcast)
		return err
	}

	// อัพเดท metrics
	metrics := make(map[string]interface{})
	metrics["total_targeted"] = len(targetUsers)
	metrics["pending"] = len(targetUsers)
	metrics["delivered"] = 0
	metrics["opened"] = 0
	metrics["clicked"] = 0
	if err := s.broadcastRepo.UpdateMetrics(broadcast.ID, metrics); err != nil {
		return err
	}

	// ส่งงานเข้า queue หรือใช้ goroutines สำหรับการส่งข้อความจริง
	go s.processBroadcastDeliveries(broadcast)

	return nil
}

// GetBroadcastStats ดึงสถิติของ broadcast
func (s *broadcastService) GetBroadcastStats(id, businessID, requestedByID uuid.UUID) (*dto.BroadcastStats, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to view broadcast stats of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return nil, errors.New("broadcast does not belong to the specified business")
	}

	// ดึงสถิติการส่ง
	stats, err := s.broadcastDeliveryRepo.GetDeliveryStats(id)
	if err != nil {
		return nil, err
	}

	// คำนวณอัตราส่วนต่างๆ
	var deliveryRate, openRate, clickRate float64
	totalTargeted := stats["total"] // จำนวนทั้งหมด
	if totalTargeted > 0 {
		deliveryRate = float64(stats["delivered"]) / float64(totalTargeted) * 100
		openRate = float64(stats["opened"]) / float64(totalTargeted) * 100
		clickRate = float64(stats["clicked"]) / float64(totalTargeted) * 100
	}

	// สร้าง BroadcastStats
	broadcastStats := &dto.BroadcastStats{
		TotalTargeted: totalTargeted,
		Pending:       stats["pending"],
		Delivered:     stats["delivered"],
		Failed:        stats["failed"],
		Opened:        stats["opened"],
		Clicked:       stats["clicked"],
		DeliveryRate:  deliveryRate,
		OpenRate:      openRate,
		ClickRate:     clickRate,
	}

	return broadcastStats, nil
}

// GetBroadcastDeliveries ดึงรายการส่ง broadcast
func (s *broadcastService) GetBroadcastDeliveries(id, businessID, requestedByID uuid.UUID, status string, limit, offset int) ([]*models.BroadcastDelivery, int64, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{})
	if err != nil {
		return nil, 0, err
	}
	if !hasPermission {
		return nil, 0, errors.New("you don't have permission to view broadcast deliveries of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return nil, 0, err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return nil, 0, errors.New("broadcast does not belong to the specified business")
	}

	// ตั้งค่า pagination
	if limit <= 0 || limit > 100 {
		limit = 20 // ค่าเริ่มต้น
	}
	if offset < 0 {
		offset = 0
	}

	// ดึงรายการส่ง broadcast
	return s.broadcastDeliveryRepo.GetByBroadcastID(id, status, limit, offset)
}

// SearchBroadcasts ค้นหา broadcasts ตามเงื่อนไข
func (s *broadcastService) SearchBroadcasts(businessID, requestedByID uuid.UUID, query, messageType, status string, startDate, endDate time.Time, limit, offset int) ([]*models.Broadcast, int64, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{})
	if err != nil {
		return nil, 0, err
	}
	if !hasPermission {
		return nil, 0, errors.New("you don't have permission to search broadcasts of this business")
	}

	// ตั้งค่า pagination
	if limit <= 0 || limit > 100 {
		limit = 20 // ค่าเริ่มต้น
	}
	if offset < 0 {
		offset = 0
	}

	// ค้นหา broadcasts
	return s.broadcastRepo.SearchBroadcasts(businessID, query, messageType, status, startDate, endDate, limit, offset)
}

// SetTargetAll กำหนดให้ส่งถึงผู้ติดตามทุกคน
func (s *broadcastService) SetTargetAll(id, businessID, requestedByID uuid.UUID) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to update broadcast target of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถอัพเดทได้หรือไม่
	if broadcast.Status != "draft" {
		return errors.New("only broadcasts in draft status can be updated")
	}

	// กำหนดเป้าหมายเป็น all
	broadcast.TargetType = "all"
	broadcast.TargetData = types.JSONB{}

	// ประมาณจำนวนผู้รับ
	followerCount, err := s.businessFollowRepo.CountFollowers(businessID)
	if err == nil && followerCount > 0 {
		// อัพเดท metrics
		metrics := make(map[string]interface{})
		metrics["total_targeted"] = followerCount
		if err := s.broadcastRepo.UpdateMetrics(broadcast.ID, metrics); err != nil {
			return err
		}
	}

	// บันทึกการอัพเดท
	return s.broadcastRepo.Update(broadcast)
}

// SetTargetTags กำหนดให้ส่งถึงผู้ใช้ตาม tags
func (s *broadcastService) SetTargetTags(id, businessID, requestedByID uuid.UUID, includeTags, excludeTags []uuid.UUID, matchType string) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to update broadcast target of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถอัพเดทได้หรือไม่
	if broadcast.Status != "draft" {
		return errors.New("only broadcasts in draft status can be updated")
	}

	// ตรวจสอบ matchType
	if matchType != "all" && matchType != "any" {
		matchType = "all" // ค่าเริ่มต้น
	}

	// ตรวจสอบว่า includeTags มีค่าหรือไม่
	if len(includeTags) == 0 {
		return errors.New("include_tags must not be empty")
	}

	// ตรวจสอบว่า tags ทั้งหมดเป็นของธุรกิจนี้
	for _, tagID := range includeTags {
		tag, err := s.tagRepo.GetByID(tagID)
		if err != nil {
			return fmt.Errorf("tag not found: %v", tagID)
		}
		if tag.BusinessID != businessID {
			return fmt.Errorf("tag %v does not belong to this business", tagID)
		}
	}

	for _, tagID := range excludeTags {
		tag, err := s.tagRepo.GetByID(tagID)
		if err != nil {
			return fmt.Errorf("tag not found: %v", tagID)
		}
		if tag.BusinessID != businessID {
			return fmt.Errorf("tag %v does not belong to this business", tagID)
		}
	}

	// กำหนดเป้าหมายเป็น tags
	broadcast.TargetType = "tags"
	broadcast.TargetData = types.JSONB{
		"include_tags": includeTags,
		"exclude_tags": excludeTags,
		"match_type":   matchType,
	}

	// ประมาณจำนวนผู้รับ
	targetCount, err := s.GetEstimatedTargetCount(businessID, requestedByID, "tags", &dto.BroadcastTargetCriteria{
		IncludeTags:  includeTags,
		ExcludeTags:  excludeTags,
		TagMatchType: matchType,
	})
	if err == nil && targetCount > 0 {
		// อัพเดท metrics
		metrics := make(map[string]interface{})
		metrics["total_targeted"] = targetCount
		if err := s.broadcastRepo.UpdateMetrics(broadcast.ID, metrics); err != nil {
			return err
		}
	}

	// บันทึกการอัพเดท
	return s.broadcastRepo.Update(broadcast)
}

// SetTargetUsers กำหนดให้ส่งถึงผู้ใช้เฉพาะราย
func (s *broadcastService) SetTargetUsers(id, businessID, requestedByID uuid.UUID, userIDs []uuid.UUID) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to update broadcast target of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถอัพเดทได้หรือไม่
	if broadcast.Status != "draft" {
		return errors.New("only broadcasts in draft status can be updated")
	}

	// ตรวจสอบว่า userIDs มีค่าหรือไม่
	if len(userIDs) == 0 {
		return errors.New("user_ids must not be empty")
	}

	// ตรวจสอบว่าผู้ใช้ทั้งหมดติดตามธุรกิจนี้
	validUserIDs := []uuid.UUID{}
	for _, userID := range userIDs {
		isFollowing, err := s.businessFollowRepo.IsFollowing(userID, businessID)
		if err != nil {
			continue
		}
		if isFollowing {
			validUserIDs = append(validUserIDs, userID)
		}
	}

	if len(validUserIDs) == 0 {
		return errors.New("no valid users found (all specified users are not following this business)")
	}

	// กำหนดเป้าหมายเป็น specific_users
	broadcast.TargetType = "specific_users"
	broadcast.TargetData = types.JSONB{
		"user_ids": validUserIDs,
	}

	// อัพเดท metrics
	metrics := make(map[string]interface{})
	metrics["total_targeted"] = len(validUserIDs)
	if err := s.broadcastRepo.UpdateMetrics(broadcast.ID, metrics); err != nil {
		return err
	}

	// บันทึกการอัพเดท
	return s.broadcastRepo.Update(broadcast)
}

// SetTargetCustomerProfile กำหนดให้ส่งถึงผู้ใช้ตามข้อมูล customer profile
func (s *broadcastService) SetTargetCustomerProfile(id, businessID, requestedByID uuid.UUID, criteria *dto.BroadcastTargetCriteria) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to update broadcast target of this business")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return errors.New("broadcast does not belong to the specified business")
	}

	// ตรวจสอบว่า broadcast สามารถอัพเดทได้หรือไม่
	if broadcast.Status != "draft" {
		return errors.New("only broadcasts in draft status can be updated")
	}

	// ตรวจสอบว่า criteria มีค่าหรือไม่
	if criteria == nil {
		return errors.New("criteria must not be nil")
	}

	// ตรวจสอบว่ามีเงื่อนไขอย่างน้อยหนึ่งข้อ
	hasCondition := len(criteria.CustomerTypes) > 0 ||
		criteria.LastContactFrom != nil ||
		criteria.LastContactTo != nil ||
		len(criteria.Statuses) > 0 ||
		len(criteria.CustomQuery) > 0

	if !hasCondition {
		return errors.New("at least one condition must be specified")
	}

	// กำหนดเป้าหมายเป็น customer_profile
	broadcast.TargetType = "customer_profile"

	// แปลง criteria เป็น TargetData
	targetData := make(map[string]interface{})
	if len(criteria.CustomerTypes) > 0 {
		targetData["customer_types"] = criteria.CustomerTypes
	}
	if criteria.LastContactFrom != nil {
		targetData["last_contact_from"] = criteria.LastContactFrom.Format(time.RFC3339)
	}
	if criteria.LastContactTo != nil {
		targetData["last_contact_to"] = criteria.LastContactTo.Format(time.RFC3339)
	}
	if len(criteria.Statuses) > 0 {
		targetData["statuses"] = criteria.Statuses
	}
	if len(criteria.CustomQuery) > 0 {
		targetData["custom_query"] = criteria.CustomQuery
	}

	broadcast.TargetData = types.JSONB(targetData)

	// ประมาณจำนวนผู้รับ
	targetCount, err := s.GetEstimatedTargetCount(businessID, requestedByID, "customer_profile", criteria)
	if err == nil && targetCount > 0 {
		// อัพเดท metrics
		metrics := make(map[string]interface{})
		metrics["total_targeted"] = targetCount
		if err := s.broadcastRepo.UpdateMetrics(broadcast.ID, metrics); err != nil {
			return err
		}
	}

	// บันทึกการอัพเดท
	return s.broadcastRepo.Update(broadcast)
}

// GetEstimatedTargetCount ประมาณจำนวนผู้รับตามเงื่อนไขที่กำหนด
func (s *broadcastService) GetEstimatedTargetCount(businessID, requestedByID uuid.UUID, targetType string, targetCriteria *dto.BroadcastTargetCriteria) (int64, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{})
	if err != nil {
		return 0, err
	}
	if !hasPermission {
		return 0, errors.New("you don't have permission to access this business")
	}

	// สร้าง broadcast ชั่วคราวเพื่อใช้ในการประมาณจำนวนผู้รับ
	tempBroadcast := &models.Broadcast{
		ID:         uuid.New(), // ไม่มีผลเพราะเป็นการใช้ชั่วคราว
		BusinessID: businessID,
		TargetType: targetType,
	}

	// กำหนด TargetData ตาม targetCriteria
	switch targetType {
	case "all":
		// ในกรณี all จะใช้จำนวนผู้ติดตามทั้งหมด
		return s.businessFollowRepo.CountFollowers(businessID)

	case "tags":
		if targetCriteria == nil {
			return 0, errors.New("targetCriteria must not be nil for tags target type")
		}

		includeTagsInterface := make([]interface{}, len(targetCriteria.IncludeTags))
		for i, tag := range targetCriteria.IncludeTags {
			includeTagsInterface[i] = tag.String()
		}

		excludeTagsInterface := make([]interface{}, len(targetCriteria.ExcludeTags))
		for i, tag := range targetCriteria.ExcludeTags {
			excludeTagsInterface[i] = tag.String()
		}

		tempBroadcast.TargetData = types.JSONB{
			"include_tags": includeTagsInterface,
			"exclude_tags": excludeTagsInterface,
			"match_type":   targetCriteria.TagMatchType,
		}

	case "specific_users":
		if targetCriteria == nil || len(targetCriteria.UserIDs) == 0 {
			return 0, errors.New("userIDs must not be empty for specific_users target type")
		}

		tempBroadcast.TargetData = types.JSONB{
			"user_ids": targetCriteria.UserIDs,
		}

		// สำหรับ specific_users เราสามารถนับจำนวนได้ทันที
		return int64(len(targetCriteria.UserIDs)), nil

	case "customer_profile":
		if targetCriteria == nil {
			return 0, errors.New("targetCriteria must not be nil for customer_profile target type")
		}

		// แปลง criteria เป็น TargetData
		targetData := make(map[string]interface{})
		if len(targetCriteria.CustomerTypes) > 0 {
			targetData["customer_types"] = targetCriteria.CustomerTypes
		}
		if targetCriteria.LastContactFrom != nil {
			targetData["last_contact_from"] = targetCriteria.LastContactFrom.Format(time.RFC3339)
		}
		if targetCriteria.LastContactTo != nil {
			targetData["last_contact_to"] = targetCriteria.LastContactTo.Format(time.RFC3339)
		}
		if len(targetCriteria.Statuses) > 0 {
			targetData["statuses"] = targetCriteria.Statuses
		}
		if len(targetCriteria.CustomQuery) > 0 {
			targetData["custom_query"] = targetCriteria.CustomQuery
		}

		tempBroadcast.TargetData = types.JSONB(targetData)

	default:
		return 0, fmt.Errorf("invalid target type: %s", targetType)
	}

	// ดึงรายชื่อผู้รับ
	targetUsers, err := s.getTargetUsers(tempBroadcast)
	if err != nil {
		return 0, err
	}

	return int64(len(targetUsers)), nil
}

// PreviewBroadcast ดูตัวอย่างข้อความก่อนส่ง
func (s *broadcastService) PreviewBroadcast(id, businessID, requestedByID uuid.UUID) (types.JSONB, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to access this broadcast")
	}

	// ดึงข้อมูล broadcast
	broadcast, err := s.broadcastRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if broadcast.BusinessID != businessID {
		return nil, errors.New("broadcast does not belong to the specified business")
	}

	// สร้างตัวอย่างข้อความ
	preview := make(types.JSONB)
	preview["message_type"] = broadcast.MessageType
	preview["title"] = broadcast.Title
	preview["content"] = broadcast.Content
	preview["media_url"] = broadcast.MediaURL

	// กรณีเป็น bubble
	if broadcast.BubbleType != "" {
		preview["bubble_type"] = broadcast.BubbleType
		preview["bubble_data"] = broadcast.BubbleData
	}

	// ข้อมูลเพิ่มเติม
	preview["estimated_recipients"] = broadcast.Metrics["total_targeted"]

	return preview, nil
}

// TrackBroadcastOpen บันทึกการเปิดอ่าน broadcast
func (s *broadcastService) TrackBroadcastOpen(broadcastID, userID uuid.UUID) error {
	// ตรวจสอบว่ามี delivery record หรือไม่
	deliveries, _, err := s.broadcastDeliveryRepo.GetByBroadcastID(broadcastID, "", 1000, 0)
	if err != nil {
		return err
	}

	// ค้นหา delivery record ของผู้ใช้นี้
	for _, delivery := range deliveries {
		if delivery.UserID == userID {
			// บันทึกเวลาเปิดอ่าน
			now := time.Now()
			delivery.OpenedAt = &now
			if err := s.broadcastDeliveryRepo.MarkAsOpened(delivery.ID, now); err != nil {
				return err
			}

			// อัพเดท metrics ของ broadcast
			broadcast, err := s.broadcastRepo.GetByID(broadcastID)
			if err != nil {
				return err
			}

			metrics := broadcast.Metrics
			openCount, _ := metrics["opened"].(float64)
			metrics["opened"] = openCount + 1

			if err := s.broadcastRepo.UpdateMetrics(broadcastID, metrics); err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("delivery record not found")
}

// TrackBroadcastClick บันทึกการคลิก broadcast
func (s *broadcastService) TrackBroadcastClick(broadcastID, userID uuid.UUID) error {
	// ตรวจสอบว่ามี delivery record หรือไม่
	deliveries, _, err := s.broadcastDeliveryRepo.GetByBroadcastID(broadcastID, "", 1000, 0)
	if err != nil {
		return err
	}

	// ค้นหา delivery record ของผู้ใช้นี้
	for _, delivery := range deliveries {
		if delivery.UserID == userID {
			// บันทึกเวลาคลิก
			now := time.Now()
			delivery.ClickedAt = &now
			if err := s.broadcastDeliveryRepo.MarkAsClicked(delivery.ID, now); err != nil {
				return err
			}

			// อัพเดท metrics ของ broadcast
			broadcast, err := s.broadcastRepo.GetByID(broadcastID)
			if err != nil {
				return err
			}

			metrics := broadcast.Metrics
			clickCount, _ := metrics["clicked"].(float64)
			metrics["clicked"] = clickCount + 1

			if err := s.broadcastRepo.UpdateMetrics(broadcastID, metrics); err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("delivery record not found")
}

// ValidateBroadcastContent ตรวจสอบความถูกต้องของเนื้อหา broadcast
func (s *broadcastService) ValidateBroadcastContent(messageType string, content string, mediaURL string, bubbleType string, bubbleData types.JSONB) error {
	// ตรวจสอบประเภทข้อความ
	validMessageTypes := map[string]bool{
		"text":     true,
		"image":    true,
		"carousel": true,
		"flex":     true,
	}

	if !validMessageTypes[messageType] {
		return fmt.Errorf("invalid message type: %s", messageType)
	}

	// ตรวจสอบตามประเภทข้อความ
	switch messageType {
	case "text":
		if content == "" {
			return errors.New("content is required for text message")
		}
	case "image":
		if mediaURL == "" {
			return errors.New("media_url is required for image message")
		}
		// ตรวจสอบว่า URL เป็นรูปภาพหรือไม่
		if !(strings.HasSuffix(mediaURL, ".jpg") ||
			strings.HasSuffix(mediaURL, ".jpeg") ||
			strings.HasSuffix(mediaURL, ".png") ||
			strings.HasSuffix(mediaURL, ".gif")) && !strings.Contains(mediaURL, "cloudinary") {
			return errors.New("media_url must be an image file (jpg, jpeg, png, gif) or from cloudinary")
		}
	case "carousel":
		if bubbleData == nil || len(bubbleData) == 0 {
			return errors.New("bubble_data is required for carousel message")
		}
		// ตรวจสอบว่ามี items หรือไม่
		items, ok := bubbleData["items"]
		if !ok || items == nil {
			return errors.New("items are required in bubble_data for carousel message")
		}
	case "flex":
		if bubbleType == "" {
			return errors.New("bubble_type is required for flex message")
		}
		if bubbleData == nil || len(bubbleData) == 0 {
			return errors.New("bubble_data is required for flex message")
		}
	}

	return nil
}

// DuplicateBroadcast สร้าง broadcast ใหม่โดยคัดลอกจาก broadcast เดิม
func (s *broadcastService) DuplicateBroadcast(sourceID, businessID, requestedByID uuid.UUID) (*models.Broadcast, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedByID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to duplicate broadcasts of this business")
	}

	// ดึงข้อมูล broadcast ต้นฉบับ
	sourceBroadcast, err := s.broadcastRepo.GetByID(sourceID)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า broadcast นี้เป็นของธุรกิจที่ระบุหรือไม่
	if sourceBroadcast.BusinessID != businessID {
		return nil, errors.New("source broadcast does not belong to the specified business")
	}

	// สร้าง broadcast ใหม่
	newBroadcast := &models.Broadcast{
		ID:          uuid.New(),
		BusinessID:  businessID,
		Title:       sourceBroadcast.Title + " (copy)",
		MessageType: sourceBroadcast.MessageType,
		Content:     sourceBroadcast.Content,
		MediaURL:    sourceBroadcast.MediaURL,
		CreatedAt:   time.Now(),
		CreatedBy:   &requestedByID,
		Status:      "draft",
		TargetType:  sourceBroadcast.TargetType,
		TargetData:  sourceBroadcast.TargetData,
		BubbleType:  sourceBroadcast.BubbleType,
		BubbleData:  sourceBroadcast.BubbleData,
		Metrics:     types.JSONB{"total_targeted": 0, "delivered": 0, "opened": 0, "clicked": 0},
	}

	// บันทึกลงฐานข้อมูล
	if err := s.broadcastRepo.Create(newBroadcast); err != nil {
		return nil, err
	}

	return newBroadcast, nil
}

// ฟังก์ชันภายใน (private methods)

// getTargetUsers ดึงรายชื่อผู้รับตามเงื่อนไขที่กำหนด
func (s *broadcastService) getTargetUsers(broadcast *models.Broadcast) ([]uuid.UUID, error) {
	switch broadcast.TargetType {
	case "all":
		// ดึงรายชื่อผู้ติดตามทั้งหมด
		followers, err := s.businessFollowRepo.GetAllFollowerIDs(broadcast.BusinessID)
		if err != nil {
			return nil, err
		}
		return followers, nil

	case "tags":
		// ตรวจสอบ TargetData
		includeTags, _ := broadcast.TargetData["include_tags"].([]interface{})
		excludeTags, _ := broadcast.TargetData["exclude_tags"].([]interface{})
		matchType, _ := broadcast.TargetData["match_type"].(string)

		if len(includeTags) == 0 {
			return nil, errors.New("include_tags must not be empty")
		}

		// แปลงเป็น UUID
		includeTagIDs := make([]uuid.UUID, 0, len(includeTags))
		for _, tag := range includeTags {
			switch t := tag.(type) {
			case string:
				tagID, err := uuid.Parse(t)
				if err != nil {
					continue
				}
				includeTagIDs = append(includeTagIDs, tagID)
			case map[string]interface{}:
				if tagIDStr, ok := t["id"].(string); ok {
					tagID, err := uuid.Parse(tagIDStr)
					if err != nil {
						continue
					}
					includeTagIDs = append(includeTagIDs, tagID)
				}
			}
		}

		excludeTagIDs := make([]uuid.UUID, 0, len(excludeTags))
		for _, tag := range excludeTags {
			switch t := tag.(type) {
			case string:
				tagID, err := uuid.Parse(t)
				if err != nil {
					continue
				}
				excludeTagIDs = append(excludeTagIDs, tagID)
			case map[string]interface{}:
				if tagIDStr, ok := t["id"].(string); ok {
					tagID, err := uuid.Parse(tagIDStr)
					if err != nil {
						continue
					}
					excludeTagIDs = append(excludeTagIDs, tagID)
				}
			}
		}

		// ค้นหาผู้ใช้ตามแท็ก
		if matchType == "" {
			matchType = "all" // ค่าเริ่มต้น
		}

		// ใช้ UserTagService ในการค้นหาผู้ใช้ตาม tags
		criteria := service.TagSearchCriteria{
			IncludeTags: includeTagIDs,
			ExcludeTags: excludeTagIDs,
			MatchType:   service.TagMatchType(matchType),
		}

		// ต้องใช้ Admin ID ในการเรียก SearchUsersByTags
		// ในที่นี้ใช้ CreatedBy ของ broadcast
		createdByID := broadcast.CreatedBy
		if createdByID == nil {
			// ถ้าไม่มี CreatedBy ให้ดึง admin คนแรกของธุรกิจนี้
			admins, err := s.businessAdminRepo.GetAdminsByBusiness(broadcast.BusinessID)
			if err != nil || len(admins) == 0 {
				return nil, errors.New("cannot find admin for this business")
			}
			createdByID = &admins[0].UserID
		}

		userTags, _, err := s.userTagService.SearchUsersByTags(broadcast.BusinessID, criteria, *createdByID, 100000, 0)
		if err != nil {
			return nil, err
		}

		// แปลงเป็นรายการ userIDs
		uniqueUsers := make(map[uuid.UUID]bool)
		var userIDs []uuid.UUID

		for _, userTag := range userTags {
			if !uniqueUsers[userTag.UserID] {
				userIDs = append(userIDs, userTag.UserID)
				uniqueUsers[userTag.UserID] = true
			}
		}

		return userIDs, nil

	case "specific_users":
		// ตรวจสอบ TargetData
		userIDsInterface, ok := broadcast.TargetData["user_ids"]
		if !ok {
			return nil, errors.New("user_ids not found in target_data")
		}

		userIDsList, ok := userIDsInterface.([]interface{})
		if !ok {
			return nil, errors.New("invalid user_ids format")
		}

		// แปลงเป็น UUID
		userIDs := make([]uuid.UUID, 0, len(userIDsList))
		for _, id := range userIDsList {
			switch userID := id.(type) {
			case string:
				uid, err := uuid.Parse(userID)
				if err != nil {
					continue
				}
				userIDs = append(userIDs, uid)
			case map[string]interface{}:
				if userIDStr, ok := userID["id"].(string); ok {
					uid, err := uuid.Parse(userIDStr)
					if err != nil {
						continue
					}
					userIDs = append(userIDs, uid)
				}
			}
		}

		if len(userIDs) == 0 {
			return nil, errors.New("no valid user_ids found")
		}

		return userIDs, nil

	case "customer_profile":
		// ดึงข้อมูล customer profiles ตามเงื่อนไข
		customerTypes, _ := broadcast.TargetData["customer_types"].([]interface{})
		lastContactFromStr, _ := broadcast.TargetData["last_contact_from"].(string)
		lastContactToStr, _ := broadcast.TargetData["last_contact_to"].(string)
		statuses, _ := broadcast.TargetData["statuses"].([]interface{})
		customQuery, _ := broadcast.TargetData["custom_query"].(map[string]interface{})

		// แปลงเป็นรูปแบบที่ใช้ได้
		customerTypesList := make([]string, 0, len(customerTypes))
		for _, ct := range customerTypes {
			if ctStr, ok := ct.(string); ok {
				customerTypesList = append(customerTypesList, ctStr)
			}
		}

		statusesList := make([]string, 0, len(statuses))
		for _, s := range statuses {
			if sStr, ok := s.(string); ok {
				statusesList = append(statusesList, sStr)
			}
		}

		var lastContactFrom, lastContactTo *time.Time
		if lastContactFromStr != "" {
			t, err := time.Parse(time.RFC3339, lastContactFromStr)
			if err == nil {
				lastContactFrom = &t
			}
		}
		if lastContactToStr != "" {
			t, err := time.Parse(time.RFC3339, lastContactToStr)
			if err == nil {
				lastContactTo = &t
			}
		}

		// ดึงรายชื่อผู้ใช้ตามเงื่อนไข
		profiles, err := s.customerProfileRepo.FindByConditions(
			broadcast.BusinessID,
			customerTypesList,
			lastContactFrom,
			lastContactTo,
			statusesList,
			customQuery,
		)
		if err != nil {
			return nil, err
		}

		// แปลงเป็นรายการ userIDs
		userIDs := make([]uuid.UUID, 0, len(profiles))
		for _, profile := range profiles {
			userIDs = append(userIDs, profile.UserID)
		}

		return userIDs, nil

	default:
		return nil, fmt.Errorf("invalid target type: %s", broadcast.TargetType)
	}
}

// createDeliveries สร้าง delivery records สำหรับผู้รับทั้งหมด
func (s *broadcastService) createDeliveries(broadcast *models.Broadcast, userIDs []uuid.UUID) error {
	batchSize := 1000 // จำนวนที่จะสร้างในแต่ละครั้ง
	for i := 0; i < len(userIDs); i += batchSize {
		end := i + batchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}

		batch := userIDs[i:end]
		var deliveries []*models.BroadcastDelivery

		for _, userID := range batch {
			delivery := &models.BroadcastDelivery{
				ID:          uuid.New(),
				BroadcastID: broadcast.ID,
				UserID:      userID,
				Status:      "pending",
			}
			deliveries = append(deliveries, delivery)
		}

		// สร้าง delivery records แบบ batch
		if err := s.broadcastDeliveryRepo.CreateBatch(deliveries); err != nil {
			return err
		}
	}

	return nil
}

// processBroadcastDeliveries ประมวลผลและส่ง broadcast deliveries
func (s *broadcastService) processBroadcastDeliveries(broadcast *models.Broadcast) {
	// อัพเดทสถานะเป็น sending
	broadcast.Status = "sending"
	s.broadcastRepo.Update(broadcast)

	// กำหนด batch size และ rate limit
	batchSize := 100                   // จำนวนที่จะประมวลผลในแต่ละครั้ง
	maxRetries := 3                    // จำนวนครั้งที่จะลองใหม่หากเกิดข้อผิดพลาด
	sendDelay := 50 * time.Millisecond // ระยะเวลาระหว่างการส่งแต่ละข้อความ

	var successCount, failedCount int64

	// ประมวลผลไปจนกว่าจะไม่มี pending deliveries
	for {
		// ดึง pending deliveries
		pendingDeliveries, err := s.broadcastDeliveryRepo.GetPendingDeliveries(broadcast.ID, batchSize)
		if err != nil || len(pendingDeliveries) == 0 {
			break // หยุดเมื่อไม่มี pending deliveries หรือเกิดข้อผิดพลาด
		}

		// ประมวลผลแต่ละ delivery
		for _, delivery := range pendingDeliveries {
			var sendError error
			var retryCount int

			// ลองส่งหลายครั้งหากเกิดข้อผิดพลาด
			for retryCount < maxRetries {
				sendError = s.sendBroadcastToUser(broadcast, delivery)
				if sendError == nil {
					break // ส่งสำเร็จ
				}
				retryCount++
				time.Sleep(100 * time.Millisecond) // รอสักครู่ก่อนลองใหม่
			}

			now := time.Now()
			if sendError != nil {
				// ส่งไม่สำเร็จ
				delivery.Status = "failed"
				delivery.ErrorMessage = sendError.Error()
				failedCount++
			} else {
				// ส่งสำเร็จ
				delivery.Status = "delivered"
				delivery.DeliveredAt = &now
				successCount++
			}

			// อัพเดทสถานะ delivery
			if sendError == nil {
				s.broadcastDeliveryRepo.MarkAsDelivered(delivery.ID, now)
			} else {
				s.broadcastDeliveryRepo.UpdateStatus(delivery.ID, delivery.Status, delivery.ErrorMessage)
			}

			// รอสักครู่เพื่อป้องกัน rate limit
			time.Sleep(sendDelay)
		}
	}

	// อัพเดทสถานะและสถิติของ broadcast
	broadcast.Status = "completed"
	now := time.Now()
	broadcast.SentAt = &now

	// อัพเดท metrics
	metrics := make(map[string]interface{})
	metrics["delivered"] = successCount
	metrics["failed"] = failedCount
	s.broadcastRepo.UpdateMetrics(broadcast.ID, metrics)

	// บันทึกการอัพเดท
	s.broadcastRepo.Update(broadcast)
}

// sendBroadcastToUser ส่ง broadcast ไปยังผู้ใช้คนหนึ่ง
func (s *broadcastService) sendBroadcastToUser(broadcast *models.Broadcast, delivery *models.BroadcastDelivery) error {
	// ตรวจสอบว่าผู้ใช้ยังติดตามธุรกิจอยู่หรือไม่
	isFollowing, err := s.businessFollowRepo.IsFollowing(delivery.UserID, broadcast.BusinessID)
	if err != nil {
		return err
	}
	if !isFollowing {
		return errors.New("user is not following this business anymore")
	}

	// ดึงข้อมูลการสนทนา (หรือสร้างใหม่ถ้ายังไม่มี)
	conversationID, err := s.getOrCreateConversation(broadcast.BusinessID, delivery.UserID)
	if err != nil {
		return err
	}

	// ส่งข้อความตามประเภท
	switch broadcast.MessageType {
	case "text":
		// ส่งข้อความธรรมดา
		if broadcast.Content == "" {
			return errors.New("content is empty")
		}

		return s.messageService.SendBroadcastTextMessage(
			conversationID,
			broadcast.BusinessID,
			delivery.UserID,
			broadcast.Content,
		)

	case "image":
		// ส่งข้อความรูปภาพ
		if broadcast.MediaURL == "" {
			return errors.New("media_url is empty")
		}

		return s.messageService.SendBroadcastImageMessage(
			conversationID,
			broadcast.BusinessID,
			delivery.UserID,
			broadcast.MediaURL,
			"", // ไม่มี ThumbnailURL ใน model จึงใช้ค่าว่าง
		)

	case "carousel", "flex":
		// สำหรับข้อความประเภทซับซ้อน
		content := ""
		bubbleData := broadcast.BubbleData
		if bubbleData != nil && len(bubbleData) > 0 {
			// แปลง bubbleData เป็น JSON string
			bubbleDataBytes, err := json.Marshal(bubbleData)
			if err != nil {
				return fmt.Errorf("error marshalling bubble data: %v", err)
			}
			content = string(bubbleDataBytes)
		} else if broadcast.Content != "" {
			content = broadcast.Content
		} else {
			return errors.New("no content or bubble data")
		}

		return s.messageService.SendBroadcastCustomMessage(
			conversationID,
			broadcast.BusinessID,
			delivery.UserID,
			broadcast.MessageType,
			content,
		)

	default:
		return fmt.Errorf("unsupported message type: %s", broadcast.MessageType)
	}
}

// getOrCreateConversation ดึงหรือสร้างการสนทนาระหว่างธุรกิจและผู้ใช้
func (s *broadcastService) getOrCreateConversation(businessID, userID uuid.UUID) (uuid.UUID, error) {
	// ตรวจสอบว่ามีการสนทนาอยู่แล้วหรือไม่
	conversationID, err := s.conversationRepo.FindBusinessUserConversation(businessID, userID)
	if err == nil && conversationID != uuid.Nil {
		// มีการสนทนาอยู่แล้ว
		return conversationID, nil
	}

	// สร้างการสนทนาใหม่
	conversation := &models.Conversation{
		ID:         uuid.New(),
		Type:       "business",
		BusinessID: &businessID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsActive:   true,
	}

	// บันทึกการสนทนา
	if err := s.conversationRepo.Create(conversation); err != nil {
		return uuid.Nil, err
	}

	// เพิ่มผู้ใช้เป็นสมาชิกในการสนทนา
	member := &models.ConversationMember{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		UserID:         userID,
		JoinedAt:       time.Now(),
	}

	if err := s.conversationRepo.AddMember(member); err != nil {
		return uuid.Nil, err
	}

	return conversation.ID, nil
}
