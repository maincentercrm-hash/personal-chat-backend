// application/serviceimpl/message_welcome_service.go

package serviceimpl

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ขยาย messageService เพื่อรองรับการส่งข้อความต้อนรับ
// ฟังก์ชันเหล่านี้จะทำงานเป็นส่วนหนึ่งของ messageService

// SendWelcomeTextMessage ส่งข้อความข้อความสำหรับ welcome message
func (s *messageService) SendWelcomeTextMessage(conversationID, businessID uuid.UUID, content string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ดึง system admin ID หรือใช้ business owner
	systemAdminID, err := s.getWelcomeMessageSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความต้อนรับ
	metadata := map[string]interface{}{
		"is_welcome_message": true,
		"automated":          true,
		"welcome_type":       "text",
		"sent_at":            time.Now().Format(time.RFC3339),
	}

	// ส่งข้อความโดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessTextMessage(
		businessID,
		conversationID,
		systemAdminID,
		content,
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// SendWelcomeImageMessage ส่งข้อความรูปภาพสำหรับ welcome message
func (s *messageService) SendWelcomeImageMessage(conversationID, businessID uuid.UUID, imageURL, thumbnailURL string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบ URL รูปภาพ
	if imageURL == "" {
		return errors.New("image URL is required for welcome image message")
	}

	// ดึง system admin ID หรือใช้ business owner
	systemAdminID, err := s.getWelcomeMessageSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความต้อนรับ
	metadata := map[string]interface{}{
		"is_welcome_message": true,
		"automated":          true,
		"welcome_type":       "image",
		"sent_at":            time.Now().Format(time.RFC3339),
	}

	// ส่งรูปภาพโดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessImageMessage(
		businessID,
		conversationID,
		systemAdminID,
		imageURL,
		thumbnailURL,
		"", // ไม่มีคำอธิบายภาพ
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// SendWelcomeCustomMessage ส่งข้อความแบบกำหนดเองสำหรับ welcome message
func (s *messageService) SendWelcomeCustomMessage(conversationID, businessID uuid.UUID, messageType, contentStr string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบประเภทข้อความ
	validMessageTypes := map[string]bool{
		"card":     true,
		"carousel": true,
		"flex":     true,
	}

	if !validMessageTypes[messageType] {
		return fmt.Errorf("invalid custom message type: %s", messageType)
	}

	// ตรวจสอบว่า contentStr สามารถแปลงเป็น JSON ได้
	var content map[string]interface{}
	err = json.Unmarshal([]byte(contentStr), &content)
	if err != nil {
		return fmt.Errorf("invalid custom message content: %w", err)
	}

	// ดึง system admin ID หรือใช้ business owner
	systemAdminID, err := s.getWelcomeMessageSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความต้อนรับ
	metadata := map[string]interface{}{
		"is_welcome_message": true,
		"automated":          true,
		"welcome_type":       messageType,
		"sent_at":            time.Now().Format(time.RFC3339),
		"content":            content,
	}

	// สร้างข้อความ
	now := time.Now()
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       &systemAdminID,
		SenderType:     "business",
		MessageType:    messageType,
		Content:        contentStr, // เก็บเนื้อหาเป็น JSON string
		BusinessID:     &businessID,
		Metadata:       s.convertMetadataToJSON(metadata),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsDeleted:      false,
	}

	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับ system admin
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    systemAdminID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		// เพียงบันทึกข้อผิดพลาด ไม่ต้องหยุดการทำงาน
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), systemAdminID)
	}

	// อัปเดต last_read_at สำหรับ system admin
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, systemAdminID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, systemAdminID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	lastMsgText := fmt.Sprintf("[%s]", messageType)
	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, lastMsgText, now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return nil
}

// SendWelcomeStickerMessage ส่งข้อความสติกเกอร์สำหรับ welcome message
func (s *messageService) SendWelcomeStickerMessage(conversationID, businessID uuid.UUID, stickerID, stickerSetID uuid.UUID, mediaURL, thumbnailURL string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบ URL สติกเกอร์
	if mediaURL == "" {
		return errors.New("sticker URL is required for welcome sticker message")
	}

	// ดึง system admin ID หรือใช้ business owner
	systemAdminID, err := s.getWelcomeMessageSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความต้อนรับ
	metadata := map[string]interface{}{
		"is_welcome_message": true,
		"automated":          true,
		"welcome_type":       "sticker",
		"sent_at":            time.Now().Format(time.RFC3339),
	}

	if stickerID != uuid.Nil {
		metadata["sticker_id"] = stickerID
	}

	if stickerSetID != uuid.Nil {
		metadata["sticker_set_id"] = stickerSetID
	}

	// ส่งสติกเกอร์โดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessStickerMessage(
		businessID,
		conversationID,
		systemAdminID,
		stickerID,
		stickerSetID,
		mediaURL,
		thumbnailURL,
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// SendWelcomeFileMessage ส่งข้อความไฟล์สำหรับ welcome message
func (s *messageService) SendWelcomeFileMessage(conversationID, businessID uuid.UUID, mediaURL, fileName string, fileSize int64, fileType string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบ URL ไฟล์
	if mediaURL == "" {
		return errors.New("file URL is required for welcome file message")
	}

	// ดึง system admin ID หรือใช้ business owner
	systemAdminID, err := s.getWelcomeMessageSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความต้อนรับ
	metadata := map[string]interface{}{
		"is_welcome_message": true,
		"automated":          true,
		"welcome_type":       "file",
		"sent_at":            time.Now().Format(time.RFC3339),
	}

	if fileName != "" {
		metadata["file_name"] = fileName
	}

	if fileSize > 0 {
		metadata["file_size"] = fileSize
	}

	if fileType != "" {
		metadata["file_type"] = fileType
	}

	// ส่งไฟล์โดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessFileMessage(
		businessID,
		conversationID,
		systemAdminID,
		mediaURL,
		fileName,
		fileSize,
		fileType,
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// getWelcomeMessageSenderID ดึง ID ของผู้ส่งข้อความต้อนรับ (system admin หรือ business owner)
func (s *messageService) getWelcomeMessageSenderID(businessID uuid.UUID) (uuid.UUID, error) {
	// ตรวจสอบว่ามี system admin ID หรือไม่
	systemAdminID := uuid.Nil

	// ถ้ามีการกำหนด system admin ID ไว้ใน config
	// systemAdminID = config.GetSystemAdminID()

	// ถ้าไม่มี ให้ใช้ business owner แทน
	if systemAdminID == uuid.Nil {
		business, err := s.businessAccountRepo.GetByID(businessID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("error fetching business: %w", err)
		}

		if business.OwnerID == nil {
			return uuid.Nil, errors.New("business has no owner")
		}

		return *business.OwnerID, nil
	}

	return systemAdminID, nil
}

// convertContentToJSON แปลง content เป็น types.JSONB
func (s *messageService) convertContentToJSON(contentStr string) (types.JSONB, error) {
	var content types.JSONB
	err := json.Unmarshal([]byte(contentStr), &content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// sendWelcomeMessage ฟังก์ชันรวมสำหรับส่งข้อความต้อนรับทุกประเภท
func (s *messageService) SendWelcomeMessage(conversationID, businessID uuid.UUID, messageType string, params map[string]interface{}) error {
	switch messageType {
	case "text":
		content, _ := params["content"].(string)
		return s.SendWelcomeTextMessage(conversationID, businessID, content)

	case "image":
		imageURL, _ := params["image_url"].(string)
		thumbnailURL, _ := params["thumbnail_url"].(string)
		return s.SendWelcomeImageMessage(conversationID, businessID, imageURL, thumbnailURL)

	case "sticker":
		stickerIDStr, _ := params["sticker_id"].(string)
		stickerSetIDStr, _ := params["sticker_set_id"].(string)
		mediaURL, _ := params["media_url"].(string)
		thumbnailURL, _ := params["thumbnail_url"].(string)

		var stickerID, stickerSetID uuid.UUID
		var err error

		if stickerIDStr != "" {
			stickerID, err = uuid.Parse(stickerIDStr)
			if err != nil {
				return fmt.Errorf("invalid sticker ID: %w", err)
			}
		}

		if stickerSetIDStr != "" {
			stickerSetID, err = uuid.Parse(stickerSetIDStr)
			if err != nil {
				return fmt.Errorf("invalid sticker set ID: %w", err)
			}
		}

		return s.SendWelcomeStickerMessage(conversationID, businessID, stickerID, stickerSetID, mediaURL, thumbnailURL)

	case "file":
		mediaURL, _ := params["media_url"].(string)
		fileName, _ := params["file_name"].(string)
		fileSize, _ := params["file_size"].(float64)
		fileType, _ := params["file_type"].(string)

		return s.SendWelcomeFileMessage(conversationID, businessID, mediaURL, fileName, int64(fileSize), fileType)

	case "card", "carousel", "flex":
		contentJSON, err := json.Marshal(params["content"])
		if err != nil {
			return fmt.Errorf("error marshalling custom content: %w", err)
		}

		return s.SendWelcomeCustomMessage(conversationID, businessID, messageType, string(contentJSON))

	default:
		return fmt.Errorf("unsupported welcome message type: %s", messageType)
	}
}

// ต้องเพิ่มฟังก์ชันใหม่ใน MessageService interface ใน domain/service/message_service.go:

/*
// SendWelcomeTextMessage ส่งข้อความข้อความสำหรับ welcome message
SendWelcomeTextMessage(conversationID, businessID uuid.UUID, content string) error

// SendWelcomeImageMessage ส่งข้อความรูปภาพสำหรับ welcome message
SendWelcomeImageMessage(conversationID, businessID uuid.UUID, imageURL, thumbnailURL string) error

// SendWelcomeCustomMessage ส่งข้อความแบบกำหนดเองสำหรับ welcome message
SendWelcomeCustomMessage(conversationID, businessID uuid.UUID, messageType, contentStr string) error

// SendWelcomeStickerMessage ส่งข้อความสติกเกอร์สำหรับ welcome message
SendWelcomeStickerMessage(conversationID, businessID uuid.UUID, stickerID, stickerSetID uuid.UUID, mediaURL, thumbnailURL string) error

// SendWelcomeFileMessage ส่งข้อความไฟล์สำหรับ welcome message
SendWelcomeFileMessage(conversationID, businessID uuid.UUID, mediaURL, fileName string, fileSize int64, fileType string) error

// SendWelcomeMessage ฟังก์ชันรวมสำหรับส่งข้อความต้อนรับทุกประเภท
SendWelcomeMessage(conversationID, businessID uuid.UUID, messageType string, params map[string]interface{}) error
*/
