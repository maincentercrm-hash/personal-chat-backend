// application/serviceimpl/business_welcome_message_service.go
package serviceimpl

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type businessWelcomeMessageService struct {
	welcomeMessageRepo  repository.BusinessWelcomeMessageRepository
	businessAccountRepo repository.BusinessAccountRepository
	userRepo            repository.UserRepository
	businessAdminRepo   repository.BusinessAdminRepository
	messageService      service.MessageService // สำหรับส่งข้อความ
	conversationRepo    repository.ConversationRepository
}

// NewBusinessWelcomeMessageService สร้าง instance ใหม่ของ BusinessWelcomeMessageService
func NewBusinessWelcomeMessageService(
	welcomeMessageRepo repository.BusinessWelcomeMessageRepository,
	businessAccountRepo repository.BusinessAccountRepository,
	userRepo repository.UserRepository,
	businessAdminRepo repository.BusinessAdminRepository,
	messageService service.MessageService,
	conversationRepo repository.ConversationRepository,
) service.BusinessWelcomeMessageService {
	return &businessWelcomeMessageService{
		welcomeMessageRepo:  welcomeMessageRepo,
		businessAccountRepo: businessAccountRepo,
		userRepo:            userRepo,
		businessAdminRepo:   businessAdminRepo,
		messageService:      messageService,
		conversationRepo:    conversationRepo,
	}
}

// CreateWelcomeMessage สร้าง welcome message ใหม่
func (s *businessWelcomeMessageService) CreateWelcomeMessage(
	businessID uuid.UUID,
	userID uuid.UUID,
	messageType string,
	title string,
	content string,
	imageURL string,
	thumbnailURL string,
	actionButtons types.JSONB,
	components types.JSONB,
	triggerType string,
	triggerParams types.JSONB,
	sortOrder int,
) (*models.BusinessWelcomeMessage, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to create welcome message for this business")
	}

	// ตรวจสอบความถูกต้องของข้อมูล
	if err := s.validateMessageType(messageType); err != nil {
		return nil, err
	}

	if err := s.ValidateTriggerParams(triggerType, triggerParams); err != nil {
		return nil, err
	}

	if err := s.ValidateMessageComponents(messageType, components); err != nil {
		return nil, err
	}

	if err := s.ValidateActionButtons(actionButtons); err != nil {
		return nil, err
	}

	// สร้าง welcome message ใหม่
	welcomeMessage := &models.BusinessWelcomeMessage{
		ID:            uuid.New(),
		BusinessID:    businessID,
		IsActive:      true,
		MessageType:   messageType,
		Title:         title,
		Content:       content,
		ImageURL:      imageURL,
		ThumbnailURL:  thumbnailURL,
		ActionButtons: actionButtons,
		Components:    components,
		TriggerType:   triggerType,
		TriggerParams: triggerParams,
		SortOrder:     sortOrder,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CreatedByID:   userID,
		UpdatedByID:   userID,
	}

	// บันทึกลงฐานข้อมูล
	err = s.welcomeMessageRepo.Create(welcomeMessage)
	if err != nil {
		return nil, err
	}

	return welcomeMessage, nil
}

// GetWelcomeMessageByID ดึงข้อมูล welcome message ตาม ID
func (s *businessWelcomeMessageService) GetWelcomeMessageByID(id uuid.UUID, userID uuid.UUID) (*models.BusinessWelcomeMessage, error) {
	// ดึงข้อมูล welcome message
	welcomeMessage, err := s.welcomeMessageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, welcomeMessage.BusinessID, []string{})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to access this welcome message")
	}

	// โหลดข้อมูลเพิ่มเติม (ถ้าจำเป็น)
	// ในที่นี้เรายังไม่ได้โหลดข้อมูลเพิ่มเติม เช่น Business, CreatedByUser, UpdatedByUser
	// หากต้องการโหลดข้อมูลเหล่านี้ สามารถเพิ่มโค้ดได้ที่นี่

	return welcomeMessage, nil
}

// GetBusinessWelcomeMessages ดึงข้อมูล welcome message ทั้งหมดของธุรกิจ
func (s *businessWelcomeMessageService) GetBusinessWelcomeMessages(businessID uuid.UUID, userID uuid.UUID, includeInactive bool) ([]*models.BusinessWelcomeMessage, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to access welcome messages of this business")
	}

	// ดึงข้อมูล welcome message
	var welcomeMessages []*models.BusinessWelcomeMessage
	if includeInactive {
		welcomeMessages, err = s.welcomeMessageRepo.GetAllByBusinessID(businessID)
	} else {
		welcomeMessages, err = s.welcomeMessageRepo.GetActiveByBusinessID(businessID)
	}
	if err != nil {
		return nil, err
	}

	return welcomeMessages, nil
}

// UpdateWelcomeMessage อัพเดทข้อมูล welcome message
func (s *businessWelcomeMessageService) UpdateWelcomeMessage(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID, updateData types.JSONB) (*models.BusinessWelcomeMessage, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to update welcome message of this business")
	}

	// ดึงข้อมูล welcome message ปัจจุบัน
	welcomeMessage, err := s.welcomeMessageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า welcome message นี้เป็นของธุรกิจที่ระบุหรือไม่
	if welcomeMessage.BusinessID != businessID {
		return nil, errors.New("welcome message does not belong to the specified business")
	}

	// อัพเดทข้อมูลตามที่ระบุ
	updated := false

	if messageType, ok := updateData["message_type"].(string); ok && messageType != "" {
		if err := s.validateMessageType(messageType); err != nil {
			return nil, err
		}
		welcomeMessage.MessageType = messageType
		updated = true
	}

	if title, ok := updateData["title"].(string); ok {
		welcomeMessage.Title = title
		updated = true
	}

	if content, ok := updateData["content"].(string); ok {
		welcomeMessage.Content = content
		updated = true
	}

	if imageURL, ok := updateData["image_url"].(string); ok {
		welcomeMessage.ImageURL = imageURL
		updated = true
	}

	if thumbnailURL, ok := updateData["thumbnail_url"].(string); ok {
		welcomeMessage.ThumbnailURL = thumbnailURL
		updated = true
	}

	if actionButtons, ok := updateData["action_buttons"].(map[string]interface{}); ok {
		jsonbActionButtons := types.JSONB(actionButtons)
		if err := s.ValidateActionButtons(jsonbActionButtons); err != nil {
			return nil, err
		}
		welcomeMessage.ActionButtons = jsonbActionButtons
		updated = true
	}

	if components, ok := updateData["components"].(map[string]interface{}); ok {
		jsonbComponents := types.JSONB(components)
		if err := s.ValidateMessageComponents(welcomeMessage.MessageType, jsonbComponents); err != nil {
			return nil, err
		}
		welcomeMessage.Components = jsonbComponents
		updated = true
	}

	if triggerType, ok := updateData["trigger_type"].(string); ok && triggerType != "" {
		welcomeMessage.TriggerType = triggerType
		updated = true
	}

	if triggerParams, ok := updateData["trigger_params"].(map[string]interface{}); ok {
		jsonbTriggerParams := types.JSONB(triggerParams)
		if err := s.ValidateTriggerParams(welcomeMessage.TriggerType, jsonbTriggerParams); err != nil {
			return nil, err
		}
		welcomeMessage.TriggerParams = jsonbTriggerParams
		updated = true
	}

	if isActive, ok := updateData["is_active"].(bool); ok {
		welcomeMessage.IsActive = isActive
		updated = true
	}

	if sortOrder, ok := updateData["sort_order"].(float64); ok {
		welcomeMessage.SortOrder = int(sortOrder)
		updated = true
	}

	// ถ้าไม่มีข้อมูลที่ต้องอัพเดท
	if !updated {
		return nil, errors.New("no valid data to update")
	}

	// อัพเดทเวลาและผู้อัพเดท
	welcomeMessage.UpdatedAt = time.Now()
	welcomeMessage.UpdatedByID = userID

	// บันทึกการอัพเดท
	err = s.welcomeMessageRepo.Update(welcomeMessage)
	if err != nil {
		return nil, err
	}

	return welcomeMessage, nil
}

// DeleteWelcomeMessage ลบ welcome message
func (s *businessWelcomeMessageService) DeleteWelcomeMessage(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to delete welcome message of this business")
	}

	// ดึงข้อมูล welcome message
	welcomeMessage, err := s.welcomeMessageRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า welcome message นี้เป็นของธุรกิจที่ระบุหรือไม่
	if welcomeMessage.BusinessID != businessID {
		return errors.New("welcome message does not belong to the specified business")
	}

	// ลบ welcome message
	return s.welcomeMessageRepo.Delete(id)
}

// SetWelcomeMessageActive กำหนดสถานะการใช้งานของ welcome message
func (s *businessWelcomeMessageService) SetWelcomeMessageActive(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID, isActive bool) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to update welcome message of this business")
	}

	// ดึงข้อมูล welcome message
	welcomeMessage, err := s.welcomeMessageRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า welcome message นี้เป็นของธุรกิจที่ระบุหรือไม่
	if welcomeMessage.BusinessID != businessID {
		return errors.New("welcome message does not belong to the specified business")
	}

	// ถ้าสถานะเดิมเหมือนกับสถานะใหม่ ไม่ต้องทำอะไร
	if welcomeMessage.IsActive == isActive {
		return nil
	}

	// อัพเดทสถานะ
	return s.welcomeMessageRepo.SetActive(id, isActive)
}

// UpdateWelcomeMessageSortOrder อัพเดทลำดับการแสดงผลของ welcome message
func (s *businessWelcomeMessageService) UpdateWelcomeMessageSortOrder(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID, sortOrder int) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{"owner", "admin"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("you don't have permission to update welcome message of this business")
	}

	// ดึงข้อมูล welcome message
	welcomeMessage, err := s.welcomeMessageRepo.GetByID(id)
	if err != nil {
		return err
	}

	// ตรวจสอบว่า welcome message นี้เป็นของธุรกิจที่ระบุหรือไม่
	if welcomeMessage.BusinessID != businessID {
		return errors.New("welcome message does not belong to the specified business")
	}

	// ถ้าลำดับเดิมเหมือนกับลำดับใหม่ ไม่ต้องทำอะไร
	if welcomeMessage.SortOrder == sortOrder {
		return nil
	}

	// อัพเดทลำดับ
	return s.welcomeMessageRepo.UpdateSortOrder(id, sortOrder)
}

// GetWelcomeMessagesByTriggerType ดึงข้อมูล welcome message ตามประเภททริกเกอร์
func (s *businessWelcomeMessageService) GetWelcomeMessagesByTriggerType(businessID uuid.UUID, userID uuid.UUID, triggerType string) ([]*models.BusinessWelcomeMessage, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to access welcome messages of this business")
	}

	// ดึงข้อมูล welcome message
	return s.welcomeMessageRepo.GetByTriggerType(businessID, triggerType)
}

// ProcessFollowWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้เริ่มติดตามธุรกิจ
func (s *businessWelcomeMessageService) ProcessFollowWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID) error {
	// ดึงข้อมูล welcome message ประเภท follow ที่เปิดใช้งาน
	welcomeMessages, err := s.welcomeMessageRepo.GetByTriggerType(businessID, "follow")
	if err != nil {
		return err
	}

	if len(welcomeMessages) == 0 {
		return nil // ไม่มี welcome message ที่ต้องส่ง
	}

	// ดึงข้อมูลธุรกิจ
	_, err = s.businessAccountRepo.GetByID(businessID)
	if err != nil { // ถูก: return err เมื่อเกิดข้อผิดพลาด
		return err
	}

	// ดึงข้อมูลผู้ใช้
	_, err = s.userRepo.FindByID(targetUserID)
	if err != nil { // ถูก: return err เมื่อเกิดข้อผิดพลาด
		return err
	}

	// ตรวจสอบว่ามีการสนทนาระหว่างธุรกิจและผู้ใช้หรือไม่
	// ถ้ายังไม่มี ให้สร้างใหม่
	// ส่วนนี้ขึ้นอยู่กับการออกแบบของคุณว่าจะจัดการการสนทนาอย่างไร
	// ในที่นี้เราจะสมมติว่ามีฟังก์ชัน GetOrCreateConversation
	conversationID, err := s.getOrCreateConversation(businessID, targetUserID)
	if err != nil {
		return err
	}

	// ส่ง welcome message แต่ละรายการ
	for _, message := range welcomeMessages {
		// เรนเดอร์เนื้อหา welcome message พร้อมแทนที่ตัวแปร
		renderedContent, err := s.RenderWelcomeMessageContent(message, targetUserID)
		if err != nil {
			continue // ข้ามไปข้อความถัดไปหากเกิดข้อผิดพลาด
		}

		// ส่งข้อความตามประเภท
		switch message.MessageType {
		case "text":
			content := ""
			if contentValue, ok := renderedContent["content"].(string); ok {
				content = contentValue
			} else if message.Content != "" {
				content = message.Content
			}

			if content != "" {
				// ใช้ messageService เพื่อส่งข้อความ
				err = s.sendTextMessage(conversationID, businessID, targetUserID, content)
				if err != nil {
					continue
				}
			}

		case "image":
			imageURL := ""
			if urlValue, ok := renderedContent["image_url"].(string); ok {
				imageURL = urlValue
			} else if message.ImageURL != "" {
				imageURL = message.ImageURL
			}

			if imageURL != "" {
				// ส่งรูปภาพ
				err = s.sendImageMessage(conversationID, businessID, targetUserID, imageURL, message.ThumbnailURL)
				if err != nil {
					continue
				}
			}

		case "card", "carousel", "flex":
			// สำหรับข้อความประเภทซับซ้อน ส่งในรูปแบบ JSON
			err = s.sendCustomMessage(conversationID, businessID, targetUserID, message.MessageType, renderedContent)
			if err != nil {
				continue
			}
		}

		// บันทึกสถิติการส่งข้อความ
		err = s.TrackMessageSent(message.ID)
		if err != nil {
			// แค่บันทึกข้อผิดพลาด ไม่ต้องหยุดการทำงาน
			fmt.Printf("Error tracking message sent: %v\n", err)
		}
	}

	return nil
}

// ProcessCommandWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้ส่งคำสั่งเฉพาะ
func (s *businessWelcomeMessageService) ProcessCommandWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID, command string) ([]*models.BusinessWelcomeMessage, error) {
	// ดึงข้อมูล welcome message ประเภท command ที่เปิดใช้งาน
	welcomeMessages, err := s.welcomeMessageRepo.GetByTriggerType(businessID, "command")
	if err != nil {
		return nil, err
	}

	if len(welcomeMessages) == 0 {
		return nil, nil // ไม่มี welcome message ที่ต้องส่ง
	}

	// กรองเฉพาะ welcome message ที่ตรงกับคำสั่ง
	var matchedMessages []*models.BusinessWelcomeMessage
	for _, message := range welcomeMessages {
		if s.isCommandMatch(message, command) {
			matchedMessages = append(matchedMessages, message)
		}
	}

	if len(matchedMessages) == 0 {
		return nil, nil // ไม่มี welcome message ที่ตรงกับคำสั่ง
	}

	// ดึงข้อมูลการสนทนา (หรือสร้างใหม่ถ้ายังไม่มี)
	conversationID, err := s.getOrCreateConversation(businessID, targetUserID)
	if err != nil {
		return nil, err
	}

	// ส่ง welcome message ที่ตรงกับคำสั่ง
	for _, message := range matchedMessages {
		// เรนเดอร์เนื้อหา welcome message พร้อมแทนที่ตัวแปร
		renderedContent, err := s.RenderWelcomeMessageContent(message, targetUserID)
		if err != nil {
			continue
		}

		// ส่งข้อความตามประเภท (เหมือนกับใน ProcessFollowWelcomeMessages)
		switch message.MessageType {
		case "text":
			content := ""
			if contentValue, ok := renderedContent["content"].(string); ok {
				content = contentValue
			} else if message.Content != "" {
				content = message.Content
			}

			if content != "" {
				err = s.sendTextMessage(conversationID, businessID, targetUserID, content)
				if err != nil {
					continue
				}
			}

		case "image":
			imageURL := ""
			if urlValue, ok := renderedContent["image_url"].(string); ok {
				imageURL = urlValue
			} else if message.ImageURL != "" {
				imageURL = message.ImageURL
			}

			if imageURL != "" {
				err = s.sendImageMessage(conversationID, businessID, targetUserID, imageURL, message.ThumbnailURL)
				if err != nil {
					continue
				}
			}

		case "card", "carousel", "flex":
			err = s.sendCustomMessage(conversationID, businessID, targetUserID, message.MessageType, renderedContent)
			if err != nil {
				continue
			}
		}

		// บันทึกสถิติการส่งข้อความ
		err = s.TrackMessageSent(message.ID)
		if err != nil {
			fmt.Printf("Error tracking message sent: %v\n", err)
		}
	}

	return matchedMessages, nil
}

// ProcessConversationStartWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้เริ่มต้นสนทนา
func (s *businessWelcomeMessageService) ProcessConversationStartWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID, conversationID uuid.UUID) error {
	// ดึงข้อมูล welcome message ประเภท conversation_start ที่เปิดใช้งาน
	welcomeMessages, err := s.welcomeMessageRepo.GetByTriggerType(businessID, "conversation_start")
	if err != nil {
		return err
	}

	if len(welcomeMessages) == 0 {
		return nil // ไม่มี welcome message ที่ต้องส่ง
	}

	// ส่ง welcome message แต่ละรายการ (ดำเนินการเหมือนกับ ProcessFollowWelcomeMessages)
	for _, message := range welcomeMessages {
		// เรนเดอร์เนื้อหา welcome message พร้อมแทนที่ตัวแปร
		renderedContent, err := s.RenderWelcomeMessageContent(message, targetUserID)
		if err != nil {
			continue
		}

		// ส่งข้อความตามประเภท
		switch message.MessageType {
		case "text":
			content := ""
			if contentValue, ok := renderedContent["content"].(string); ok {
				content = contentValue
			} else if message.Content != "" {
				content = message.Content
			}

			if content != "" {
				err = s.sendTextMessage(conversationID, businessID, targetUserID, content)
				if err != nil {
					continue
				}
			}

		case "image":
			imageURL := ""
			if urlValue, ok := renderedContent["image_url"].(string); ok {
				imageURL = urlValue
			} else if message.ImageURL != "" {
				imageURL = message.ImageURL
			}

			if imageURL != "" {
				err = s.sendImageMessage(conversationID, businessID, targetUserID, imageURL, message.ThumbnailURL)
				if err != nil {
					continue
				}
			}

		case "card", "carousel", "flex":
			err = s.sendCustomMessage(conversationID, businessID, targetUserID, message.MessageType, renderedContent)
			if err != nil {
				continue
			}
		}

		// บันทึกสถิติการส่งข้อความ
		err = s.TrackMessageSent(message.ID)
		if err != nil {
			fmt.Printf("Error tracking message sent: %v\n", err)
		}
	}

	return nil
}

// ProcessInactiveWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้กลับมาหลังจากไม่มีกิจกรรม
func (s *businessWelcomeMessageService) ProcessInactiveWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID, inactiveDays int) error {
	// ดึงข้อมูล welcome message ประเภท inactive ที่เปิดใช้งาน
	welcomeMessages, err := s.welcomeMessageRepo.GetByTriggerType(businessID, "inactive")
	if err != nil {
		return err
	}

	if len(welcomeMessages) == 0 {
		return nil // ไม่มี welcome message ที่ต้องส่ง
	}

	// กรองเฉพาะ welcome message ที่ตรงกับจำนวนวันที่ไม่มีกิจกรรม
	var matchedMessages []*models.BusinessWelcomeMessage
	for _, message := range welcomeMessages {
		if s.isInactiveDaysMatch(message, inactiveDays) {
			matchedMessages = append(matchedMessages, message)
		}
	}

	if len(matchedMessages) == 0 {
		return nil // ไม่มี welcome message ที่ตรงกับจำนวนวันที่ไม่มีกิจกรรม
	}

	// ดึงข้อมูลการสนทนา (หรือสร้างใหม่ถ้ายังไม่มี)
	conversationID, err := s.getOrCreateConversation(businessID, targetUserID)
	if err != nil {
		return err
	}

	// ส่ง welcome message ที่ตรงกับจำนวนวันที่ไม่มีกิจกรรม
	for _, message := range matchedMessages {
		// เรนเดอร์เนื้อหา welcome message พร้อมแทนที่ตัวแปร
		renderedContent, err := s.RenderWelcomeMessageContent(message, targetUserID)
		if err != nil {
			continue
		}

		// ส่งข้อความตามประเภท
		switch message.MessageType {
		case "text":
			content := ""
			if contentValue, ok := renderedContent["content"].(string); ok {
				content = contentValue
			} else if message.Content != "" {
				content = message.Content
			}

			if content != "" {
				err = s.sendTextMessage(conversationID, businessID, targetUserID, content)
				if err != nil {
					continue
				}
			}

		case "image":
			imageURL := ""
			if urlValue, ok := renderedContent["image_url"].(string); ok {
				imageURL = urlValue
			} else if message.ImageURL != "" {
				imageURL = message.ImageURL
			}

			if imageURL != "" {
				err = s.sendImageMessage(conversationID, businessID, targetUserID, imageURL, message.ThumbnailURL)
				if err != nil {
					continue
				}
			}

		case "card", "carousel", "flex":
			err = s.sendCustomMessage(conversationID, businessID, targetUserID, message.MessageType, renderedContent)
			if err != nil {
				continue
			}
		}

		// บันทึกสถิติการส่งข้อความ
		err = s.TrackMessageSent(message.ID)
		if err != nil {
			fmt.Printf("Error tracking message sent: %v\n", err)
		}
	}

	return nil
}

// TrackMessageSent บันทึกการส่ง welcome message
func (s *businessWelcomeMessageService) TrackMessageSent(messageID uuid.UUID) error {
	return s.welcomeMessageRepo.UpdateMetrics(messageID, 1, 0, 0)
}

// TrackMessageClick บันทึกการคลิกในแต่ละแอคชั่นของ welcome message
func (s *businessWelcomeMessageService) TrackMessageClick(messageID uuid.UUID, actionType string, actionData types.JSONB) error {
	return s.welcomeMessageRepo.UpdateMetrics(messageID, 0, 1, 0)
}

// TrackMessageReply บันทึกการตอบกลับ welcome message
func (s *businessWelcomeMessageService) TrackMessageReply(messageID uuid.UUID) error {
	return s.welcomeMessageRepo.UpdateMetrics(messageID, 0, 0, 1)
}

// RenderWelcomeMessageContent เรนเดอร์เนื้อหา welcome message พร้อมแทนที่ตัวแปร
func (s *businessWelcomeMessageService) RenderWelcomeMessageContent(message *models.BusinessWelcomeMessage, targetUserID uuid.UUID) (types.JSONB, error) {
	result := make(types.JSONB)

	// ดึงข้อมูลธุรกิจ
	business, err := s.businessAccountRepo.GetByID(message.BusinessID)
	if err != nil {
		return nil, err
	}

	// ดึงข้อมูลผู้ใช้
	user, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return nil, err
	}

	// คัดลอกข้อมูลพื้นฐาน
	result["message_type"] = message.MessageType
	result["title"] = s.replaceVariables(message.Title, user, business)
	result["content"] = s.replaceVariables(message.Content, user, business)
	result["image_url"] = message.ImageURL
	result["thumbnail_url"] = message.ThumbnailURL

	// คัดลอกและปรับแต่ง action buttons
	if message.ActionButtons != nil {
		actionButtons := make(types.JSONB)
		for key, value := range message.ActionButtons {
			if strValue, ok := value.(string); ok {
				actionButtons[key] = s.replaceVariables(strValue, user, business)
			} else {
				actionButtons[key] = value
			}
		}
		result["action_buttons"] = actionButtons
	}

	// คัดลอกและปรับแต่ง components
	if message.Components != nil {
		components := s.processComponents(message.Components, user, business)
		result["components"] = components
	}

	return result, nil
}

// ValidateTriggerParams ตรวจสอบความถูกต้องของพารามิเตอร์ทริกเกอร์
func (s *businessWelcomeMessageService) ValidateTriggerParams(triggerType string, triggerParams types.JSONB) error {
	// ตรวจสอบประเภททริกเกอร์
	validTriggerTypes := map[string]bool{
		"follow":             true,
		"inactive":           true,
		"schedule":           true,
		"command":            true,
		"conversation_start": true,
		"location":           true,
		"event":              true,
	}

	if !validTriggerTypes[triggerType] {
		return fmt.Errorf("invalid trigger type: %s", triggerType)
	}

	// ตรวจสอบพารามิเตอร์ตามประเภททริกเกอร์
	switch triggerType {
	case "inactive":
		// ต้องมี days
		if triggerParams == nil {
			return errors.New("trigger_params is required for inactive trigger type")
		}
		if _, ok := triggerParams["days"]; !ok {
			return errors.New("days parameter is required for inactive trigger type")
		}
	case "schedule":
		// ต้องมี delay หรือ specific_time
		if triggerParams == nil {
			return errors.New("trigger_params is required for schedule trigger type")
		}
		if _, ok := triggerParams["delay"]; !ok && triggerParams["specific_time"] == nil {
			return errors.New("delay or specific_time parameter is required for schedule trigger type")
		}
	case "command":
		// ต้องมี commands
		if triggerParams == nil {
			return errors.New("trigger_params is required for command trigger type")
		}
		if _, ok := triggerParams["commands"]; !ok {
			return errors.New("commands parameter is required for command trigger type")
		}
	case "location":
		// ต้องมี locations
		if triggerParams == nil {
			return errors.New("trigger_params is required for location trigger type")
		}
		if _, ok := triggerParams["locations"]; !ok {
			return errors.New("locations parameter is required for location trigger type")
		}
	case "event":
		// ต้องมี event_types
		if triggerParams == nil {
			return errors.New("trigger_params is required for event trigger type")
		}
		if _, ok := triggerParams["event_types"]; !ok {
			return errors.New("event_types parameter is required for event trigger type")
		}
	}

	return nil
}

// ValidateMessageComponents ตรวจสอบความถูกต้องของคอมโพเนนต์ข้อความ
func (s *businessWelcomeMessageService) ValidateMessageComponents(messageType string, components types.JSONB) error {
	// ตรวจสอบประเภทข้อความ
	validMessageTypes := map[string]bool{
		"text":     true,
		"image":    true,
		"card":     true,
		"carousel": true,
		"flex":     true,
	}

	if !validMessageTypes[messageType] {
		return fmt.Errorf("invalid message type: %s", messageType)
	}

	// สำหรับประเภท carousel ต้องมี items
	if messageType == "carousel" && components != nil {
		if items, ok := components["items"]; !ok || items == nil {
			return errors.New("items are required for carousel message type")
		}
	}

	// สำหรับประเภท flex ต้องมี type
	if messageType == "flex" && components != nil {
		if flexType, ok := components["type"]; !ok || flexType == nil {
			return errors.New("type is required for flex message type")
		}
	}

	return nil
}

// ValidateActionButtons ตรวจสอบความถูกต้องของปุ่มดำเนินการ
func (s *businessWelcomeMessageService) ValidateActionButtons(actionButtons types.JSONB) error {
	if actionButtons == nil {
		return nil
	}

	// ถ้าเป็น array ของปุ่ม
	if buttons, ok := actionButtons["buttons"].([]interface{}); ok {
		for i, btn := range buttons {
			button, ok := btn.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid button format at index %d", i)
			}

			// ตรวจสอบประเภทปุ่ม
			btnType, ok := button["type"].(string)
			if !ok || btnType == "" {
				return fmt.Errorf("button type is required at index %d", i)
			}

			// ตรวจสอบข้อมูลตามประเภทปุ่ม
			switch btnType {
			case "url":
				if _, ok := button["url"].(string); !ok {
					return fmt.Errorf("url is required for url button at index %d", i)
				}
			case "message":
				if _, ok := button["text"].(string); !ok {
					return fmt.Errorf("text is required for message button at index %d", i)
				}
			case "phone":
				if _, ok := button["phone"].(string); !ok {
					return fmt.Errorf("phone is required for phone button at index %d", i)
				}
			case "menu":
				if _, ok := button["menu_id"].(string); !ok {
					return fmt.Errorf("menu_id is required for menu button at index %d", i)
				}
			case "custom":
				if _, ok := button["action"].(string); !ok {
					return fmt.Errorf("action is required for custom button at index %d", i)
				}
			default:
				return fmt.Errorf("invalid button type: %s at index %d", btnType, i)
			}

			// ตรวจสอบว่ามีข้อความปุ่ม
			if _, ok := button["label"].(string); !ok {
				return fmt.Errorf("label is required for button at index %d", i)
			}
		}
	}

	return nil
}

// validateMessageType ตรวจสอบความถูกต้องของประเภทข้อความ
func (s *businessWelcomeMessageService) validateMessageType(messageType string) error {
	validMessageTypes := map[string]bool{
		"text":     true,
		"image":    true,
		"card":     true,
		"carousel": true,
		"flex":     true,
	}

	if !validMessageTypes[messageType] {
		return fmt.Errorf("invalid message type: %s", messageType)
	}

	return nil
}

// isCommandMatch ตรวจสอบว่าคำสั่งตรงกับ welcome message หรือไม่
func (s *businessWelcomeMessageService) isCommandMatch(message *models.BusinessWelcomeMessage, command string) bool {
	if message.TriggerParams == nil {
		return false
	}

	commandsInterface, ok := message.TriggerParams["commands"]
	if !ok {
		return false
	}

	commands, ok := commandsInterface.([]interface{})
	if !ok {
		return false
	}

	// ตรวจสอบว่าต้องสนใจตัวพิมพ์เล็ก-ใหญ่หรือไม่
	caseSensitive := true
	if caseSensitiveVal, ok := message.TriggerParams["case_sensitive"].(bool); ok {
		caseSensitive = caseSensitiveVal
	}

	// ตรวจสอบว่าต้องตรงกันทั้งหมดหรือไม่
	exactMatch := true
	if exactMatchVal, ok := message.TriggerParams["exact_match"].(bool); ok {
		exactMatch = exactMatchVal
	}

	// ปรับคำสั่งให้เป็นตัวพิมพ์เล็กทั้งหมด ถ้าไม่สนใจตัวพิมพ์เล็ก-ใหญ่
	userCommand := command
	if !caseSensitive {
		userCommand = strings.ToLower(userCommand)
	}

	// ตรวจสอบคำสั่ง
	for _, cmd := range commands {
		cmdStr, ok := cmd.(string)
		if !ok {
			continue
		}

		// ปรับคำสั่งให้เป็นตัวพิมพ์เล็กทั้งหมด ถ้าไม่สนใจตัวพิมพ์เล็ก-ใหญ่
		if !caseSensitive {
			cmdStr = strings.ToLower(cmdStr)
		}

		if exactMatch {
			// ต้องตรงกันทั้งหมด
			if userCommand == cmdStr {
				return true
			}
		} else {
			// ตรวจสอบว่ามีคำสั่งอยู่ในข้อความหรือไม่
			if strings.Contains(userCommand, cmdStr) {
				return true
			}
		}
	}

	return false
}

// isInactiveDaysMatch ตรวจสอบว่าจำนวนวันที่ไม่มีกิจกรรมตรงกับ welcome message หรือไม่
func (s *businessWelcomeMessageService) isInactiveDaysMatch(message *models.BusinessWelcomeMessage, inactiveDays int) bool {
	if message.TriggerParams == nil {
		return false
	}

	daysInterface, ok := message.TriggerParams["days"]
	if !ok {
		return false
	}

	// แปลงเป็นจำนวนวัน
	var days int
	switch d := daysInterface.(type) {
	case float64:
		days = int(d)
	case int:
		days = d
	default:
		return false
	}

	// ตรวจสอบว่าจำนวนวันตรงกันหรือไม่
	return inactiveDays >= days
}

// replaceVariables แทนที่ตัวแปรในข้อความ
func (s *businessWelcomeMessageService) replaceVariables(text string, user *models.User, business *models.BusinessAccount) string {
	if text == "" {
		return ""
	}

	// แทนที่ตัวแปรผู้ใช้
	text = strings.ReplaceAll(text, "{{user_name}}", user.DisplayName)
	text = strings.ReplaceAll(text, "{{user_id}}", user.ID.String())
	text = strings.ReplaceAll(text, "{{user_profile_url}}", user.ProfileImageURL)

	// แทนที่ตัวแปรธุรกิจ
	text = strings.ReplaceAll(text, "{{business_name}}", business.Name)
	text = strings.ReplaceAll(text, "{{business_username}}", business.Username)
	text = strings.ReplaceAll(text, "{{business_profile_url}}", business.ProfileImageURL)

	// แทนที่ตัวแปรเวลา
	now := time.Now()
	text = strings.ReplaceAll(text, "{{current_date}}", now.Format("2006-01-02"))
	text = strings.ReplaceAll(text, "{{current_time}}", now.Format("15:04:05"))
	text = strings.ReplaceAll(text, "{{day_of_week}}", now.Weekday().String())

	return text
}

// processComponents ประมวลผลและแทนที่ตัวแปรในคอมโพเนนต์
func (s *businessWelcomeMessageService) processComponents(components types.JSONB, user *models.User, business *models.BusinessAccount) types.JSONB {
	if components == nil {
		return nil
	}

	result := make(types.JSONB)

	// คัดลอกและปรับแต่งข้อมูล
	for key, value := range components {
		switch v := value.(type) {
		case string:
			result[key] = s.replaceVariables(v, user, business)
		case map[string]interface{}:
			// แปลงเป็น types.JSONB และประมวลผลแบบเรียกซ้ำ
			subComponents := types.JSONB(v)
			result[key] = s.processComponents(subComponents, user, business)
		case []interface{}:
			// ประมวลผลแต่ละรายการในอาร์เรย์
			var processed []interface{}
			for _, item := range v {
				switch i := item.(type) {
				case string:
					processed = append(processed, s.replaceVariables(i, user, business))
				case map[string]interface{}:
					// แปลงเป็น types.JSONB และประมวลผลแบบเรียกซ้ำ
					subComponents := types.JSONB(i)
					processed = append(processed, s.processComponents(subComponents, user, business))
				default:
					processed = append(processed, i)
				}
			}
			result[key] = processed
		default:
			result[key] = v
		}
	}

	return result
}

// getOrCreateConversation ดึงหรือสร้างการสนทนาระหว่างธุรกิจและผู้ใช้
func (s *businessWelcomeMessageService) getOrCreateConversation(businessID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	// ฟังก์ชันนี้ต้องได้รับการปรับแต่งตามโครงสร้างของระบบของคุณ
	// ในที่นี้เราจะสมมติว่ามีการเรียกฟังก์ชันจาก conversationRepo

	// หาการสนทนาที่มีอยู่แล้ว
	conversationID, err := s.conversationRepo.FindBusinessUserConversation(businessID, userID)
	if err == nil && conversationID != uuid.Nil {
		return conversationID, nil
	}

	// สร้างการสนทนาใหม่
	// (ปรับแต่งตามโครงสร้างของระบบของคุณ)
	conversation := &models.Conversation{
		ID:         uuid.New(),
		Type:       "business",
		BusinessID: &businessID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsActive:   true,
	}

	err = s.conversationRepo.Create(conversation)
	if err != nil {
		return uuid.Nil, err
	}

	// เพิ่มผู้ใช้เป็นสมาชิกในการสนทนา
	// (ปรับแต่งตามโครงสร้างของระบบของคุณ)
	member := &models.ConversationMember{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		UserID:         userID,
		JoinedAt:       time.Now(),
	}

	err = s.conversationRepo.AddMember(member)
	if err != nil {
		return uuid.Nil, err
	}

	return conversation.ID, nil
}

// sendTextMessage ส่งข้อความข้อความธรรมดา
func (s *businessWelcomeMessageService) sendTextMessage(conversationID uuid.UUID, businessID uuid.UUID, targetUserID uuid.UUID, content string) error {
	// เรียกใช้ฟังก์ชัน SendWelcomeTextMessage ที่สร้างใหม่
	return s.messageService.SendWelcomeTextMessage(conversationID, businessID, content)
}

// sendImageMessage ส่งข้อความรูปภาพ
func (s *businessWelcomeMessageService) sendImageMessage(conversationID uuid.UUID, businessID uuid.UUID, targetUserID uuid.UUID, imageURL string, thumbnailURL string) error {
	// เรียกใช้ฟังก์ชัน SendWelcomeImageMessage ที่สร้างใหม่
	return s.messageService.SendWelcomeImageMessage(conversationID, businessID, imageURL, thumbnailURL)
}

// sendCustomMessage ส่งข้อความแบบกำหนดเอง
func (s *businessWelcomeMessageService) sendCustomMessage(conversationID uuid.UUID, businessID uuid.UUID, targetUserID uuid.UUID, messageType string, content types.JSONB) error {
	// แปลง content เป็น JSON string
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("error marshalling content to JSON: %w", err)
	}

	contentStr := string(contentBytes)

	// เรียกใช้ฟังก์ชัน SendWelcomeCustomMessage ที่สร้างใหม่
	return s.messageService.SendWelcomeCustomMessage(conversationID, businessID, messageType, contentStr)
}
