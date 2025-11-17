// application/serviceimpl/message_service.go
package serviceimpl

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// urlRegex สำหรับตรวจจับ URL ในข้อความ
var urlRegex = regexp.MustCompile(`https?://[^\s]+`)

// messageService เป็น implementation ของ MessageService interface
type messageService struct {
	messageRepo         repository.MessageRepository
	messageReadRepo     repository.MessageReadRepository
	conversationRepo    repository.ConversationRepository
	userRepo            repository.UserRepository
	notificationService service.NotificationService
}

// NewMessageService สร้าง instance ใหม่ของ MessageService
func NewMessageService(
	messageRepo repository.MessageRepository,
	messageReadRepo repository.MessageReadRepository,
	conversationRepo repository.ConversationRepository,
	userRepo repository.UserRepository,
	notificationService service.NotificationService,
) service.MessageService {
	return &messageService{
		messageRepo:         messageRepo,
		messageReadRepo:     messageReadRepo,
		conversationRepo:    conversationRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

// CheckBusinessAdmin ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่

// createMessageRead สร้างบันทึกการอ่านข้อความ
func (s *messageService) createMessageRead(messageID, userID uuid.UUID) error {
	// ตรวจสอบว่ามีบันทึกการอ่านแล้วหรือไม่

	isRead, err := s.messageRepo.IsMessageRead(messageID, userID)
	if err != nil {
		return err
	}

	if isRead {
		return nil // ถ้าอ่านแล้ว ไม่ต้องทำอะไร
	}

	// สร้างบันทึกการอ่าน
	now := time.Now()
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    now,
	}

	return s.messageReadRepo.CreateRead(messageRead)
}

// updateConversationLastRead อัปเดต last_read_at ในข้อมูลสมาชิกการสนทนา
func (s *messageService) updateConversationLastRead(conversationID, userID uuid.UUID, readTime time.Time) error {

	return s.conversationRepo.UpdateMemberLastRead(conversationID, userID, readTime)
}

func (s *messageService) convertMetadataToJSON(metadata map[string]interface{}) types.JSONB {
	if metadata == nil {
		return types.JSONB{} // คืนค่า JSONB ที่เป็น empty map
	}

	// สร้าง types.JSONB ใหม่จาก metadata
	jsonb := types.JSONB{}
	for k, v := range metadata {
		jsonb[k] = v
	}

	return jsonb
}

// extractLinks ดึง URLs จากข้อความ
func (s *messageService) extractLinks(content string) []string {
	if content == "" {
		return nil
	}

	links := urlRegex.FindAllString(content, -1)
	if len(links) == 0 {
		return nil
	}

	// Remove duplicates
	uniqueLinks := make(map[string]bool)
	result := []string{}

	for _, link := range links {
		if !uniqueLinks[link] {
			uniqueLinks[link] = true
			result = append(result, link)
		}
	}

	return result
}
