// application/serviceimpl/message_broadcast_service.go

package serviceimpl

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// SendBroadcastTextMessage ส่งข้อความข้อความสำหรับ broadcast message
func (s *messageService) SendBroadcastTextMessage(conversationID, businessID, userID uuid.UUID, content string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกในการสนทนานี้หรือไม่
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return fmt.Errorf("error checking membership: %w", err)
	}

	if !isMember {
		return errors.New("user is not a member of this conversation")
	}

	// ดึงข้อมูลแอดมินที่ส่ง broadcast
	adminID, err := s.getBroadcastSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความ broadcast
	metadata := map[string]interface{}{
		"is_broadcast_message": true,
		"broadcast_type":       "text",
		"sent_at":              time.Now().Format(time.RFC3339),
		"target_user_id":       userID.String(),
	}

	// ส่งข้อความโดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessTextMessage(
		businessID,
		conversationID,
		adminID,
		content,
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// SendBroadcastImageMessage ส่งข้อความรูปภาพสำหรับ broadcast message
func (s *messageService) SendBroadcastImageMessage(conversationID, businessID, userID uuid.UUID, mediaURL, thumbnailURL string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกในการสนทนานี้หรือไม่
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return fmt.Errorf("error checking membership: %w", err)
	}

	if !isMember {
		return errors.New("user is not a member of this conversation")
	}

	// ตรวจสอบ URL รูปภาพ
	if mediaURL == "" {
		return errors.New("image URL is required for broadcast image message")
	}

	// ดึงข้อมูลแอดมินที่ส่ง broadcast
	adminID, err := s.getBroadcastSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความ broadcast
	metadata := map[string]interface{}{
		"is_broadcast_message": true,
		"broadcast_type":       "image",
		"sent_at":              time.Now().Format(time.RFC3339),
		"target_user_id":       userID.String(),
	}

	// ส่งรูปภาพโดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessImageMessage(
		businessID,
		conversationID,
		adminID,
		mediaURL,
		thumbnailURL,
		"", // ไม่มีคำอธิบายภาพ
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// SendBroadcastCustomMessage ส่งข้อความแบบกำหนดเองสำหรับ broadcast message
func (s *messageService) SendBroadcastCustomMessage(conversationID, businessID, userID uuid.UUID, messageType, contentStr string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกในการสนทนานี้หรือไม่
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return fmt.Errorf("error checking membership: %w", err)
	}

	if !isMember {
		return errors.New("user is not a member of this conversation")
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

	// ดึงข้อมูลแอดมินที่ส่ง broadcast
	adminID, err := s.getBroadcastSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความ broadcast
	metadata := map[string]interface{}{
		"is_broadcast_message": true,
		"broadcast_type":       messageType,
		"sent_at":              time.Now().Format(time.RFC3339),
		"target_user_id":       userID.String(),
		"content":              content,
	}

	// สร้างข้อความ
	now := time.Now()
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       &adminID,
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

	// สร้างบันทึกการอ่านสำหรับ admin
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    adminID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		// เพียงบันทึกข้อผิดพลาด ไม่ต้องหยุดการทำงาน
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), adminID)
	}

	// อัปเดต last_read_at สำหรับ admin
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, adminID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, adminID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	lastMsgText := fmt.Sprintf("[%s]", messageType)
	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, lastMsgText, now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return nil
}

// application/serviceimpl/message_broadcast_service.go

// SendBroadcastStickerMessage ส่งข้อความสติกเกอร์สำหรับ broadcast message
func (s *messageService) SendBroadcastStickerMessage(conversationID, businessID, userID uuid.UUID, stickerID, stickerSetID uuid.UUID, mediaURL, thumbnailURL string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกในการสนทนานี้หรือไม่
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return fmt.Errorf("error checking membership: %w", err)
	}

	if !isMember {
		return errors.New("user is not a member of this conversation")
	}

	// ตรวจสอบ URL สติกเกอร์
	if mediaURL == "" {
		return errors.New("sticker URL is required for broadcast sticker message")
	}

	// ดึงข้อมูลแอดมินที่ส่ง broadcast
	adminID, err := s.getBroadcastSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความ broadcast
	metadata := map[string]interface{}{
		"is_broadcast_message": true,
		"broadcast_type":       "sticker",
		"sent_at":              time.Now().Format(time.RFC3339),
		"target_user_id":       userID.String(),
	}

	// ส่งสติกเกอร์โดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessStickerMessage(
		businessID,
		conversationID,
		adminID,
		stickerID,
		stickerSetID,
		mediaURL,
		thumbnailURL,
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// SendBroadcastFileMessage ส่งข้อความไฟล์สำหรับ broadcast message
func (s *messageService) SendBroadcastFileMessage(conversationID, businessID, userID uuid.UUID, mediaURL, fileName string, fileSize int64, fileType string) error {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return errors.New("this conversation does not belong to the specified business")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกในการสนทนานี้หรือไม่
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return fmt.Errorf("error checking membership: %w", err)
	}

	if !isMember {
		return errors.New("user is not a member of this conversation")
	}

	// ตรวจสอบ URL ไฟล์
	if mediaURL == "" {
		return errors.New("file URL is required for broadcast file message")
	}

	// ดึงข้อมูลแอดมินที่ส่ง broadcast
	adminID, err := s.getBroadcastSenderID(businessID)
	if err != nil {
		return err
	}

	// เตรียม metadata สำหรับข้อความ broadcast
	metadata := map[string]interface{}{
		"is_broadcast_message": true,
		"broadcast_type":       "file",
		"sent_at":              time.Now().Format(time.RFC3339),
		"target_user_id":       userID.String(),
	}

	// ส่งไฟล์โดยใช้ฟังก์ชันที่มีอยู่
	_, err = s.SendBusinessFileMessage(
		businessID,
		conversationID,
		adminID,
		mediaURL,
		fileName,
		fileSize,
		fileType,
		metadata,
		nil, // ไม่มีการตอบกลับ
	)

	return err
}

// getBroadcastSenderID ดึง ID ของผู้ส่ง broadcast (admin หรือ business owner)
func (s *messageService) getBroadcastSenderID(businessID uuid.UUID) (uuid.UUID, error) {
	// ตรวจสอบว่ามี admin ที่ระบุหรือไม่
	// ตรวจสอบว่ามีการระบุ admin ID ใน Broadcast หรือไม่
	// ถ้าไม่มี ให้ใช้ business owner แทน
	business, err := s.businessAccountRepo.GetByID(businessID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error fetching business: %w", err)
	}

	if business.OwnerID == nil {
		return uuid.Nil, errors.New("business has no owner")
	}

	return *business.OwnerID, nil
}
