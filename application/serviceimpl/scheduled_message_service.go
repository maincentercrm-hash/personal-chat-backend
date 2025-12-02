// application/serviceimpl/scheduled_message_service.go
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

type scheduledMessageService struct {
	scheduledMessageRepo repository.ScheduledMessageRepository
	conversationRepo     repository.ConversationRepository
	messageService       service.MessageService
}

// NewScheduledMessageService สร้าง instance ใหม่ของ ScheduledMessageService
func NewScheduledMessageService(
	scheduledMessageRepo repository.ScheduledMessageRepository,
	conversationRepo repository.ConversationRepository,
	messageService service.MessageService,
) service.ScheduledMessageService {
	return &scheduledMessageService{
		scheduledMessageRepo: scheduledMessageRepo,
		conversationRepo:     conversationRepo,
		messageService:       messageService,
	}
}

// ScheduleMessage กำหนดเวลาส่งข้อความ
func (s *scheduledMessageService) ScheduleMessage(
	conversationID, userID uuid.UUID,
	messageType, content, mediaURL string,
	metadata map[string]interface{},
	scheduledAt time.Time,
) (*models.ScheduledMessage, error) {
	// ตรวจสอบว่า user เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not a member of this conversation")
	}

	// ตรวจสอบว่า scheduled_at ต้องอยู่ในอนาคต
	if scheduledAt.Before(time.Now()) {
		return nil, errors.New("scheduled_at must be in the future")
	}

	// สร้าง metadata JSONB
	metadataJSON := make(map[string]interface{})
	if metadata != nil {
		for k, v := range metadata {
			metadataJSON[k] = v
		}
	}

	// สร้าง scheduled message
	scheduledMsg := &models.ScheduledMessage{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       userID,
		MessageType:    messageType,
		Content:        content,
		MediaURL:       mediaURL,
		Metadata:       metadataJSON,
		ScheduledAt:    scheduledAt,
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// บันทึกลงฐานข้อมูล
	if err := s.scheduledMessageRepo.Create(scheduledMsg); err != nil {
		return nil, err
	}

	return scheduledMsg, nil
}

// GetScheduledMessage ดึงข้อมูลข้อความที่กำหนดเวลาส่ง
func (s *scheduledMessageService) GetScheduledMessage(id, userID uuid.UUID) (*models.ScheduledMessage, error) {
	scheduledMsg, err := s.scheduledMessageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if scheduledMsg == nil {
		return nil, errors.New("scheduled message not found")
	}

	// ตรวจสอบว่า user เป็นเจ้าของข้อความ
	if scheduledMsg.SenderID != userID {
		return nil, errors.New("unauthorized to access this scheduled message")
	}

	return scheduledMsg, nil
}

// GetUserScheduledMessages ดึงรายการข้อความที่กำหนดเวลาส่งของผู้ใช้
func (s *scheduledMessageService) GetUserScheduledMessages(userID uuid.UUID, limit, offset int) ([]*models.ScheduledMessage, int64, error) {
	return s.scheduledMessageRepo.FindByUserID(userID, limit, offset)
}

// GetConversationScheduledMessages ดึงรายการข้อความที่กำหนดเวลาส่งในการสนทนา
func (s *scheduledMessageService) GetConversationScheduledMessages(conversationID, userID uuid.UUID, limit, offset int) ([]*models.ScheduledMessage, int64, error) {
	// ตรวจสอบว่า user เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !isMember {
		return nil, 0, errors.New("user is not a member of this conversation")
	}

	return s.scheduledMessageRepo.FindByConversationID(conversationID, limit, offset)
}

// CancelScheduledMessage ยกเลิกข้อความที่กำหนดเวลาส่ง
func (s *scheduledMessageService) CancelScheduledMessage(id, userID uuid.UUID) error {
	// ดึงข้อมูล scheduled message
	scheduledMsg, err := s.scheduledMessageRepo.GetByID(id)
	if err != nil {
		return err
	}
	if scheduledMsg == nil {
		return errors.New("scheduled message not found")
	}

	// ตรวจสอบว่า user เป็นเจ้าของข้อความ
	if scheduledMsg.SenderID != userID {
		return errors.New("unauthorized to cancel this scheduled message")
	}

	// ตรวจสอบสถานะ
	if scheduledMsg.Status != "pending" {
		return errors.New("can only cancel pending scheduled messages")
	}

	return s.scheduledMessageRepo.CancelScheduledMessage(id)
}

// ProcessScheduledMessages ประมวลผลข้อความที่ถึงเวลาส่ง
func (s *scheduledMessageService) ProcessScheduledMessages() error {
	// ดึงข้อความที่ถึงเวลาส่งแล้ว
	now := time.Now()
	scheduledMessages, err := s.scheduledMessageRepo.FindPendingMessages(now, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch pending messages: %w", err)
	}

	fmt.Printf("Processing %d scheduled messages\n", len(scheduledMessages))

	// ส่งแต่ละข้อความ
	for _, scheduledMsg := range scheduledMessages {
		if err := s.sendScheduledMessage(scheduledMsg); err != nil {
			fmt.Printf("Failed to send scheduled message %s: %v\n", scheduledMsg.ID, err)
			// อัปเดตสถานะเป็น failed
			_ = s.scheduledMessageRepo.UpdateStatus(
				scheduledMsg.ID,
				"failed",
				nil,
				nil,
				err.Error(),
			)
		}
	}

	return nil
}

// sendScheduledMessage ส่งข้อความที่กำหนดเวลาส่ง
func (s *scheduledMessageService) sendScheduledMessage(scheduledMsg *models.ScheduledMessage) error {
	var message *models.Message
	var err error

	// แปลง metadata กลับเป็น map[string]interface{}
	metadata := make(map[string]interface{})
	for k, v := range scheduledMsg.Metadata {
		metadata[k] = v
	}

	// ส่งข้อความตามประเภท
	switch scheduledMsg.MessageType {
	case "text":
		message, err = s.messageService.SendTextMessage(
			scheduledMsg.ConversationID,
			scheduledMsg.SenderID,
			scheduledMsg.Content,
			metadata,
		)
	case "image":
		message, err = s.messageService.SendImageMessage(
			scheduledMsg.ConversationID,
			scheduledMsg.SenderID,
			scheduledMsg.MediaURL,
			"", // thumbnailURL
			scheduledMsg.Content, // caption
			metadata,
		)
	case "file":
		message, err = s.messageService.SendFileMessage(
			scheduledMsg.ConversationID,
			scheduledMsg.SenderID,
			scheduledMsg.MediaURL,
			scheduledMsg.Content, // fileName
			0, // fileSize
			"", // fileType
			metadata,
		)
	case "sticker":
		// สติกเกอร์ต้องมี stickerID ใน metadata
		var stickerID, stickerSetID uuid.UUID
		if stickerIDStr, ok := metadata["sticker_id"].(string); ok {
			stickerID, _ = uuid.Parse(stickerIDStr)
		}
		if stickerSetIDStr, ok := metadata["sticker_set_id"].(string); ok {
			stickerSetID, _ = uuid.Parse(stickerSetIDStr)
		}

		message, err = s.messageService.SendStickerMessage(
			scheduledMsg.ConversationID,
			scheduledMsg.SenderID,
			stickerID,
			stickerSetID,
			scheduledMsg.MediaURL,
			"", // thumbnailURL
			metadata,
		)
	default:
		return fmt.Errorf("unsupported message type: %s", scheduledMsg.MessageType)
	}

	if err != nil {
		return err
	}

	// อัปเดตสถานะเป็น sent
	now := time.Now()
	return s.scheduledMessageRepo.UpdateStatus(
		scheduledMsg.ID,
		"sent",
		&now,
		&message.ID,
		"",
	)
}
