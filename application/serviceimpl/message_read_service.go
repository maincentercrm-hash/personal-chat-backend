// application/serviceimpl/message_read_service.go
package serviceimpl

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

// messageReadService เป็น implementation ของ MessageReadService
type messageReadService struct {
	messageRepo      repository.MessageRepository
	messageReadRepo  repository.MessageReadRepository
	conversationRepo repository.ConversationRepository
}

// NewMessageReadService สร้าง instance ใหม่ของ MessageReadService
func NewMessageReadService(
	messageRepo repository.MessageRepository,
	messageReadRepo repository.MessageReadRepository,
	conversationRepo repository.ConversationRepository,
) service.MessageReadService {
	return &messageReadService{
		messageRepo:      messageRepo,
		messageReadRepo:  messageReadRepo,
		conversationRepo: conversationRepo,
	}
}

// MarkMessageAsRead ทำเครื่องหมายว่าข้อความถูกอ่านแล้ว
func (s *messageReadService) MarkMessageAsRead(messageID, userID uuid.UUID) (uuid.UUID, error) {

	// ดึงข้อมูลข้อความ
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return uuid.Nil, err
	}

	if message == nil {
		return uuid.Nil, errors.New("message not found")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(message.ConversationID, userID)
	if err != nil {
		return uuid.Nil, err
	}

	if !isMember {
		return uuid.Nil, errors.New("you are not a member of this conversation")
	}

	// ถ้าผู้ใช้เป็นผู้ส่งข้อความ ไม่ต้องทำอะไร
	if message.SenderID != nil && *message.SenderID == userID {
		return uuid.Nil, errors.New("you are not a member of this conversation")
	}

	// ตรวจสอบว่าอ่านแล้วหรือยัง
	isRead, err := s.messageRepo.IsMessageRead(messageID, userID)
	if err != nil {
		return uuid.Nil, err
	}

	if isRead {
		return message.ConversationID, nil
	}

	// สร้างบันทึกการอ่าน
	now := time.Now()
	read := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(read); err != nil {
		return uuid.Nil, err
	}

	// อ่านข้อความเก่ากว่าทั้งหมดในการสนทนาเดียวกัน
	unreadMessages, err := s.messageReadRepo.GetUnreadMessageIDs(message.ConversationID, userID)
	if err == nil && len(unreadMessages) > 0 {
		// กรองเอาเฉพาะข้อความที่เก่ากว่า
		olderMessages := make([]uuid.UUID, 0)
		for _, msgID := range unreadMessages {
			msg, err := s.messageRepo.GetByID(msgID)
			if err != nil {
				continue
			}

			if msg.CreatedAt.Before(message.CreatedAt) || msg.CreatedAt.Equal(message.CreatedAt) {
				olderMessages = append(olderMessages, msgID)
			}
		}

		// มาร์คข้อความที่เก่ากว่าทั้งหมดเป็นอ่านแล้ว
		for _, msgID := range olderMessages {
			// สร้างบันทึกการอ่าน
			read := &models.MessageRead{
				ID:        uuid.New(),
				MessageID: msgID,
				UserID:    userID,
				ReadAt:    now,
			}
			s.messageReadRepo.CreateRead(read)
		}
	}

	// อัปเดตเวลาอ่านล่าสุดในตาราง conversation_members
	err = s.conversationRepo.UpdateMemberLastRead(message.ConversationID, userID, message.CreatedAt)
	if err != nil {
		return uuid.Nil, err
	}

	// คืนค่า conversationID สำหรับการแจ้งเตือน
	return message.ConversationID, nil
}

// GetMessageReads ดึงข้อมูลผู้ที่อ่านข้อความแล้ว
func (s *messageReadService) GetMessageReads(messageID, userID uuid.UUID) ([]*models.MessageRead, error) {

	// ดึงข้อมูลข้อความ
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, errors.New("message not found")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(message.ConversationID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("you are not a member of this conversation")
	}

	// ดึงข้อมูลการอ่านทั้งหมด
	return s.messageReadRepo.GetByMessageID(messageID)
}

// MarkAllMessagesAsRead ทำเครื่องหมายว่าข้อความทั้งหมดในการสนทนาถูกอ่านแล้ว
func (s *messageReadService) MarkAllMessagesAsRead(conversationID, userID uuid.UUID) (int, error) {

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return 0, err
	}

	if !isMember {
		return 0, errors.New("you are not a member of this conversation")
	}

	// อัปเดต last_read_at ก่อน
	now := time.Now()
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, userID, now); err != nil {
		// บันทึกข้อผิดพลาดแต่ไม่หยุดการทำงาน
	}

	// ดึงข้อความทั้งหมดที่ยังไม่ได้อ่าน
	unreadMessageIDs, err := s.messageReadRepo.GetUnreadMessageIDs(conversationID, userID)
	if err != nil {
		return 0, err
	}

	if len(unreadMessageIDs) == 0 {
		return 0, nil
	}

	// มาร์คข้อความทั้งหมดเป็นอ่านแล้ว
	for _, messageID := range unreadMessageIDs {
		// สร้างบันทึกการอ่าน
		read := &models.MessageRead{
			ID:        uuid.New(),
			MessageID: messageID,
			UserID:    userID,
			ReadAt:    now,
		}
		if err := s.messageReadRepo.CreateRead(read); err != nil {
			// บันทึกข้อผิดพลาดแต่ไม่หยุดการทำงาน
		}
	}

	return len(unreadMessageIDs), nil
}

// GetUnreadCount ดึงจำนวนข้อความที่ยังไม่ได้อ่านในการสนทนา
func (s *messageReadService) GetUnreadCount(conversationID, userID uuid.UUID) (int, error) {

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return 0, err
	}

	if !isMember {
		return 0, errors.New("you are not a member of this conversation")
	}

	// ดึงข้อความทั้งหมดที่ยังไม่ได้อ่าน
	unreadMessageIDs, err := s.messageReadRepo.GetUnreadMessageIDs(conversationID, userID)
	if err != nil {
		return 0, err
	}

	return len(unreadMessageIDs), nil
}
