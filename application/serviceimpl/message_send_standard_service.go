// application/serviceimpl/message_send_standard.go
package serviceimpl

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// SendTextMessage ส่งข้อความประเภทข้อความ (text)
func (s *messageService) SendTextMessage(conversationID, userID uuid.UUID, content string, metadata map[string]interface{}) (*models.Message, error) {

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("error checking conversation membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	// ตรวจสอบเนื้อหาข้อความ
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	// ดึงข้อมูลการสนทนา (เพื่อตรวจสอบประเภทการสนทนา)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	// Extract links จากข้อความและเพิ่มลงใน metadata
	links := s.extractLinks(content)
	if len(links) > 0 {
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["links"] = links
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       &userID,
		SenderType:     "user",
		MessageType:    "text",
		Content:        content,
		Metadata:       s.convertMetadataToJSON(metadata),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsDeleted:      false,
	}


	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    userID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), userID)
	}

	// อัปเดต last_read_at สำหรับผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, userID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, userID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, content, now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return message, nil
}

// SendStickerMessage ส่งข้อความประเภทสติกเกอร์
func (s *messageService) SendStickerMessage(conversationID, userID, stickerID, stickerSetID uuid.UUID, mediaURL, thumbnailURL string, metadata map[string]interface{}) (*models.Message, error) {

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("error checking conversation membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	// ตรวจสอบ URL สติกเกอร์
	if mediaURL == "" {
		return nil, fmt.Errorf("sticker URL is required")
	}

	// สร้าง metadata สำหรับสติกเกอร์
	stickerMetadata := make(map[string]interface{})
	if metadata != nil {
		for k, v := range metadata {
			stickerMetadata[k] = v
		}
	}

	// เพิ่มข้อมูลสติกเกอร์ลงใน metadata
	if stickerID != uuid.Nil {
		stickerMetadata["sticker_id"] = stickerID
	}

	if stickerSetID != uuid.Nil {
		stickerMetadata["sticker_set_id"] = stickerSetID
	}

	// ดึงข้อมูลการสนทนา (เพื่อตรวจสอบประเภทการสนทนา)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:                uuid.New(),
		ConversationID:    conversationID,
		SenderID:          &userID,
		SenderType:        "user",
		MessageType:       "sticker",
		MediaURL:          mediaURL,
		MediaThumbnailURL: thumbnailURL,
		Metadata:          s.convertMetadataToJSON(stickerMetadata),
		CreatedAt:         now,
		UpdatedAt:         now,
		IsDeleted:         false,
	}


	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    userID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), userID)
	}

	// อัปเดต last_read_at สำหรับผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, userID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, userID)
	}

	// อัปเดตข้อความล่าสุดของการสนทนา
	if err := s.messageRepo.UpdateConversationLastMessage(conversationID, "[Sticker]", now); err != nil {
		fmt.Printf("Error updating conversation last message: %v, conversationID: %s", err, conversationID)
	}

	return message, nil
}

// SendImageMessage ส่งข้อความประเภทรูปภาพ
func (s *messageService) SendImageMessage(conversationID, userID uuid.UUID, mediaURL, thumbnailURL, caption string, metadata map[string]interface{}) (*models.Message, error) {

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("error checking conversation membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	// ตรวจสอบ URL รูปภาพ
	if mediaURL == "" {
		return nil, fmt.Errorf("image URL is required")
	}

	// ดึงข้อมูลการสนทนา (เพื่อตรวจสอบประเภทการสนทนา)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:                uuid.New(),
		ConversationID:    conversationID,
		SenderID:          &userID,
		SenderType:        "user",
		MessageType:       "image",
		Content:           caption,
		MediaURL:          mediaURL,
		MediaThumbnailURL: thumbnailURL,
		Metadata:          s.convertMetadataToJSON(metadata),
		CreatedAt:         now,
		UpdatedAt:         now,
		IsDeleted:         false,
	}


	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    userID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), userID)
	}

	// อัปเดต last_read_at สำหรับผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, userID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, userID)
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

// SendFileMessage ส่งข้อความประเภทไฟล์
func (s *messageService) SendFileMessage(conversationID, userID uuid.UUID, mediaURL, fileName string, fileSize int64, fileType string, metadata map[string]interface{}) (*models.Message, error) {

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("error checking conversation membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	// ตรวจสอบ URL ไฟล์
	if mediaURL == "" {
		return nil, fmt.Errorf("file URL is required")
	}

	// สร้าง metadata สำหรับไฟล์
	fileMetadata := make(map[string]interface{})
	if metadata != nil {
		for k, v := range metadata {
			fileMetadata[k] = v
		}
	}

	// เพิ่มข้อมูลไฟล์ลงใน metadata
	if fileName != "" {
		fileMetadata["file_name"] = fileName
	}

	if fileSize > 0 {
		fileMetadata["file_size"] = fileSize
	}

	if fileType != "" {
		fileMetadata["file_type"] = fileType
	}

	// ดึงข้อมูลการสนทนา (เพื่อตรวจสอบประเภทการสนทนา)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation: %w", err)
	}

	// สร้าง message
	now := time.Now()
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       &userID,
		SenderType:     "user",
		MessageType:    "file",
		Content:        fileName,
		MediaURL:       mediaURL,
		Metadata:       s.convertMetadataToJSON(fileMetadata),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsDeleted:      false,
	}


	// บันทึกข้อความลงในฐานข้อมูล
	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("error creating message: %w", err)
	}

	// สร้างบันทึกการอ่านสำหรับผู้ส่ง
	messageRead := &models.MessageRead{
		ID:        uuid.New(),
		MessageID: message.ID,
		UserID:    userID,
		ReadAt:    now,
	}

	if err := s.messageReadRepo.CreateRead(messageRead); err != nil {
		fmt.Printf("Error creating read record: %v, messageID: %s, userID: %s", err, message.ID.String(), userID)
	}

	// อัปเดต last_read_at สำหรับผู้ส่ง
	if err := s.conversationRepo.UpdateMemberLastRead(conversationID, userID, now); err != nil {
		fmt.Printf("Error updating last read time: %v, conversationID: %s, userID: %s", err, conversationID, userID)
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
