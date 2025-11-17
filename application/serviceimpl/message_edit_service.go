// application/serviceimpl/message_edit_service.go
package serviceimpl

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// EditMessage แก้ไขข้อความ
func (s *messageService) EditMessage(messageID, userID uuid.UUID, newContent string) (*models.Message, error) {

	// ดึงข้อมูลข้อความ
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("error fetching message: %w", err)
	}

	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// ตรวจสอบว่าข้อความถูกลบไปแล้วหรือไม่
	if message.IsDeleted {
		return nil, fmt.Errorf("cannot edit deleted message")
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของข้อความหรือไม่
	if message.SenderID == nil || *message.SenderID != userID {
		return nil, fmt.Errorf("only message owner can edit messages")
	}

	// ตรวจสอบประเภทข้อความ (เฉพาะข้อความประเภท "text" เท่านั้นที่แก้ไขได้)
	if message.MessageType != "text" {
		return nil, fmt.Errorf("only text messages can be edited")
	}

	// ถ้าเนื้อหาใหม่เหมือนเนื้อหาเดิม ไม่ต้องอัพเดต
	if message.Content == newContent {
		return message, nil
	}

	// เก็บประวัติการแก้ไข
	editHistory := &models.MessageEditHistory{
		ID:              uuid.New(),
		MessageID:       messageID,
		PreviousContent: message.Content,
		EditedAt:        time.Now(),
		EditedBy:        userID,
		Metadata: s.convertMetadataToJSON(map[string]interface{}{
			"edit_number": message.EditCount + 1,
		}),
	}

	if err := s.messageRepo.CreateEditHistory(editHistory); err != nil {
		fmt.Printf("Failed to save edit history: %v\n", err)
	}

	// Extract links จากเนื้อหาใหม่และอัพเดท metadata
	links := s.extractLinks(newContent)
	if len(links) > 0 {
		// เพิ่ม links ใน metadata
		if message.Metadata == nil {
			message.Metadata = make(types.JSONB)
		}
		message.Metadata["links"] = links
	} else {
		// ถ้าไม่มี links ให้ลบ key "links" ออกจาก metadata
		if message.Metadata != nil {
			delete(message.Metadata, "links")
		}
	}

	// อัพเดทข้อความ
	now := time.Now()
	message.Content = newContent
	message.UpdatedAt = now
	message.IsEdited = true
	message.EditCount++

	if err := s.messageRepo.Update(message); err != nil {
		return nil, fmt.Errorf("error updating message: %w", err)
	}

	// ตรวจสอบว่าเป็นข้อความล่าสุดของการสนทนาหรือไม่ และอัพเดทหากจำเป็น
	lastMessage, err := s.messageRepo.GetLastMessageByConversation(message.ConversationID)
	if err == nil && lastMessage != nil && lastMessage.ID == message.ID {
		if err := s.messageRepo.UpdateConversationLastMessage(message.ConversationID, newContent, now); err != nil {
			fmt.Printf("Error updating conversation last message: %v\n", err)
		}
	}

	return message, nil
}

// GetMessageEditHistory ดึงประวัติการแก้ไขข้อความ
func (s *messageService) GetMessageEditHistory(messageID, userID uuid.UUID) ([]*models.MessageEditHistory, error) {

	// ดึงข้อมูลข้อความ
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("error fetching message: %w", err)
	}

	if message == nil {
		return nil, fmt.Errorf("message not found")
	}

	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนา
	isMember, err := s.conversationRepo.IsMember(message.ConversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("error checking membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("you are not a member of this conversation")
	}

	// ดึงประวัติการแก้ไข
	history, err := s.messageRepo.GetEditHistory(messageID)
	if err != nil {
		return nil, fmt.Errorf("error fetching edit history: %w", err)
	}

	// เพิ่มข้อมูลเพิ่มเติมให้แต่ละรายการ
	for _, edit := range history {
		// ดึงข้อมูลผู้แก้ไข
		editor, err := s.userRepo.FindByID(edit.EditedBy)
		if err == nil && editor != nil {
			// สร้าง metadata ใหม่ที่มีข้อมูลเพิ่มเติม
			metadataMap := types.JSONB{}

			// ถ้ามี Metadata เดิม ให้คัดลอกค่าเดิมมาก่อน
			for k, v := range edit.Metadata {
				metadataMap[k] = v
			}

			// เพิ่มข้อมูลผู้แก้ไข
			metadataMap["editor_name"] = editor.DisplayName
			if metadataMap["editor_name"] == "" {
				metadataMap["editor_name"] = editor.Username
			}
			metadataMap["editor_avatar"] = editor.ProfileImageURL

			// บันทึกกลับไปที่ metadata
			edit.Metadata = metadataMap
		}
	}

	return history, nil
}
