// application/serviceimpl/message_send_business.go
package serviceimpl

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// SendBusinessTextMessage ส่งข้อความประเภทข้อความในนามธุรกิจ
func (s *messageService) SendBusinessTextMessage(businessID, conversationID, adminID uuid.UUID, content string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error) {

	// ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่
	isAdmin, _, err := s.CheckBusinessAdmin(adminID, businessID)
	if err != nil {
		return nil, fmt.Errorf("error checking business admin: %w", err)
	}

	if !isAdmin {
		return nil, fmt.Errorf("user is not an admin of this business")
	}

	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return nil, fmt.Errorf("this conversation does not belong to your business")
	}

	// ตรวจสอบเนื้อหาข้อความ
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	// สร้าง metadata สำหรับข้อความธุรกิจ
	adminMetadata := make(map[string]interface{})
	if metadata != nil {
		for k, v := range metadata {
			adminMetadata[k] = v
		}
	}

	// ดึงข้อมูลบทบาทของแอดมิน
	admin, err := s.businessAdminRepo.GetByUserAndBusinessID(adminID, businessID)
	if err == nil && admin != nil {
		adminMetadata["admin_id"] = adminID
		adminMetadata["admin_role"] = admin.Role
		// 🆕 ดึงข้อมูลผู้ใช้เพื่อเอา display_name
		user, err := s.userRepo.FindByID(adminID)
		if err == nil && user != nil {
			// เพิ่ม display_name ของแอดมิน
			adminMetadata["admin_display_name"] = user.DisplayName
		}
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       &adminID,
		SenderType:     "business",
		MessageType:    "text",
		Content:        content,
		BusinessID:     &businessID,
		Metadata:       s.convertMetadataToJSON(adminMetadata),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsDeleted:      false,
	}

	// ตรวจสอบการตอบกลับข้อความ
	if replyToID != nil { // ตรวจสอบว่าไม่ใช่ nil ก่อน
		// ตรวจสอบว่าข้อความที่ตอบกลับมีอยู่จริงและอยู่ในการสนทนาเดียวกัน
		replyToMsg, err := s.messageRepo.GetByID(*replyToID) // ต้องใช้ *replyToID เพื่อดึงค่าจาก pointer
		if err == nil && replyToMsg != nil && replyToMsg.ConversationID == conversationID {
			message.ReplyToID = replyToID // กำหนดค่า pointer โดยตรง (ไม่ต้อง &replyToID)
		}
	}

	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับแอดมินผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    adminID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), adminID)
	}

	// อัปเดต last_read_at สำหรับแอดมินผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, adminID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, adminID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, content, now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return message, nil
}

// SendBusinessStickerMessage ส่งข้อความประเภทสติกเกอร์ในนามธุรกิจ
func (s *messageService) SendBusinessStickerMessage(businessID, conversationID, adminID, stickerID, stickerSetID uuid.UUID, mediaURL, thumbnailURL string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error) {

	// ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่
	isAdmin, _, err := s.CheckBusinessAdmin(adminID, businessID)
	if err != nil {
		return nil, fmt.Errorf("error checking business admin: %w", err)
	}

	if !isAdmin {
		return nil, fmt.Errorf("user is not an admin of this business")
	}

	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return nil, fmt.Errorf("this conversation does not belong to your business")
	}

	// ตรวจสอบ URL สติกเกอร์
	if mediaURL == "" {
		return nil, fmt.Errorf("sticker URL is required")
	}

	// สร้าง metadata สำหรับข้อความธุรกิจ
	adminMetadata := make(map[string]interface{})
	if metadata != nil {
		for k, v := range metadata {
			adminMetadata[k] = v
		}
	}

	// ดึงข้อมูลบทบาทของแอดมิน
	admin, err := s.businessAdminRepo.GetByUserAndBusinessID(adminID, businessID)
	if err == nil && admin != nil {
		adminMetadata["admin_id"] = adminID
		adminMetadata["admin_role"] = admin.Role
		user, err := s.userRepo.FindByID(adminID)
		if err == nil && user != nil {
			// เพิ่ม display_name ของแอดมิน
			adminMetadata["admin_display_name"] = user.DisplayName
		}
	}

	// เพิ่มข้อมูลสติกเกอร์ลงใน metadata
	if stickerID != uuid.Nil {
		adminMetadata["sticker_id"] = stickerID
	}

	if stickerSetID != uuid.Nil {
		adminMetadata["sticker_set_id"] = stickerSetID
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:                uuid.New(),
		ConversationID:    conversationID,
		SenderID:          &adminID,
		SenderType:        "business",
		MessageType:       "sticker",
		BusinessID:        &businessID,
		MediaURL:          mediaURL,
		MediaThumbnailURL: thumbnailURL,
		Metadata:          s.convertMetadataToJSON(adminMetadata),
		CreatedAt:         now,
		UpdatedAt:         now,
		IsDeleted:         false,
	}

	// ตรวจสอบการตอบกลับข้อความ
	if replyToID != nil {

		replyToMsg, err := s.messageRepo.GetByID(*replyToID)
		if err == nil && replyToMsg != nil && replyToMsg.ConversationID == conversationID {
			message.ReplyToID = replyToID
		}
	}

	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับแอดมินผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    adminID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), adminID)
	}

	// อัปเดต last_read_at สำหรับแอดมินผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, adminID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, adminID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, "[Sticker]", now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return message, nil
}

// SendBusinessImageMessage ส่งข้อความประเภทรูปภาพในนามธุรกิจ
func (s *messageService) SendBusinessImageMessage(businessID, conversationID, adminID uuid.UUID, mediaURL, thumbnailURL, caption string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error) {
	// แปลง string เป็น UUID

	// ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่
	isAdmin, _, err := s.CheckBusinessAdmin(adminID, businessID)
	if err != nil {
		return nil, fmt.Errorf("error checking business admin: %w", err)
	}

	if !isAdmin {
		return nil, fmt.Errorf("user is not an admin of this business")
	}

	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return nil, fmt.Errorf("this conversation does not belong to your business")
	}

	// ตรวจสอบ URL รูปภาพ
	if mediaURL == "" {
		return nil, fmt.Errorf("image URL is required")
	}

	// สร้าง metadata สำหรับข้อความธุรกิจ
	adminMetadata := make(map[string]interface{})
	if metadata != nil {
		for k, v := range metadata {
			adminMetadata[k] = v
		}
	}

	// ดึงข้อมูลบทบาทของแอดมิน
	admin, err := s.businessAdminRepo.GetByUserAndBusinessID(adminID, businessID)
	if err == nil && admin != nil {
		adminMetadata["admin_id"] = adminID
		adminMetadata["admin_role"] = admin.Role
		user, err := s.userRepo.FindByID(adminID)
		if err == nil && user != nil {
			// เพิ่ม display_name ของแอดมิน
			adminMetadata["admin_display_name"] = user.DisplayName
		}
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:                uuid.New(),
		ConversationID:    conversationID,
		SenderID:          &adminID,
		SenderType:        "business",
		MessageType:       "image",
		Content:           caption,
		BusinessID:        &businessID,
		MediaURL:          mediaURL,
		MediaThumbnailURL: thumbnailURL,
		Metadata:          s.convertMetadataToJSON(adminMetadata),
		CreatedAt:         now,
		UpdatedAt:         now,
		IsDeleted:         false,
	}

	// ตรวจสอบการตอบกลับข้อความ
	if replyToID != nil {

		replyToMsg, err := s.messageRepo.GetByID(*replyToID)
		if err == nil && replyToMsg != nil && replyToMsg.ConversationID == conversationID {
			message.ReplyToID = replyToID
		}
	}

	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับแอดมินผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    adminID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), adminID)
	}

	// อัปเดต last_read_at สำหรับแอดมินผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, adminID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, adminID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	lastMsgText := "[Image]"
	if caption != "" {
		lastMsgText = "[Image] " + caption
	}

	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, lastMsgText, now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return message, nil
}

// SendBusinessFileMessage ส่งข้อความประเภทไฟล์ในนามธุรกิจ
func (s *messageService) SendBusinessFileMessage(businessID, conversationID, adminID uuid.UUID, mediaURL, fileName string, fileSize int64, fileType string, metadata map[string]interface{}, replyToID *uuid.UUID) (*models.Message, error) {

	// ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่
	isAdmin, _, err := s.CheckBusinessAdmin(adminID, businessID)
	if err != nil {
		return nil, fmt.Errorf("error checking business admin: %w", err)
	}

	if !isAdmin {
		return nil, fmt.Errorf("user is not an admin of this business")
	}

	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจนี้หรือไม่
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	if conversation.Type != "business" || conversation.BusinessID == nil || *conversation.BusinessID != businessID {
		return nil, fmt.Errorf("this conversation does not belong to your business")
	}

	// ตรวจสอบ URL ไฟล์
	if mediaURL == "" {
		return nil, fmt.Errorf("file URL is required")
	}

	// สร้าง metadata สำหรับข้อความธุรกิจ
	adminMetadata := make(map[string]interface{})
	if metadata != nil {
		for k, v := range metadata {
			adminMetadata[k] = v
		}
	}

	// ดึงข้อมูลบทบาทของแอดมิน
	admin, err := s.businessAdminRepo.GetByUserAndBusinessID(adminID, businessID)
	if err == nil && admin != nil {
		adminMetadata["admin_id"] = adminID
		adminMetadata["admin_role"] = admin.Role
		user, err := s.userRepo.FindByID(adminID)
		if err == nil && user != nil {
			// เพิ่ม display_name ของแอดมิน
			adminMetadata["admin_display_name"] = user.DisplayName
		}
	}

	// เพิ่มข้อมูลไฟล์ลงใน metadata
	if fileName != "" {
		adminMetadata["file_name"] = fileName
	}

	if fileSize > 0 {
		adminMetadata["file_size"] = fileSize
	}

	if fileType != "" {
		adminMetadata["file_type"] = fileType
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       &adminID,
		SenderType:     "business",
		MessageType:    "file",
		Content:        fileName,
		BusinessID:     &businessID,
		MediaURL:       mediaURL,
		Metadata:       s.convertMetadataToJSON(adminMetadata),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsDeleted:      false,
	}

	// ตรวจสอบการตอบกลับข้อความ
	if replyToID != nil {

		replyToMsg, err := s.messageRepo.GetByID(*replyToID)
		if err == nil && replyToMsg != nil && replyToMsg.ConversationID == conversationID {
			message.ReplyToID = replyToID
		}
	}

	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับแอดมินผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    adminID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), adminID)
	}

	// อัปเดต last_read_at สำหรับแอดมินผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, adminID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, adminID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	lastMsgText := "[File]"
	if fileName != "" {
		lastMsgText = "[File] " + fileName
	}

	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, lastMsgText, now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return message, nil
}
